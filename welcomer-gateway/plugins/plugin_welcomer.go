package plugins

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"image/color"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	sandwich_daemon "github.com/WelcomerTeam/Sandwich-Daemon"
	sandwich_protobuf "github.com/WelcomerTeam/Sandwich-Daemon/proto"
	sandwich "github.com/WelcomerTeam/Sandwich/sandwich"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	core "github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/jackc/pgx/v4"
	"github.com/savsgio/gotils/strconv"
)

const (
	DefaultFont               = "fredokaone-regular"
	DefaultImageBorderWidth   = 16
	DefaultProfileBorderWidth = 8
)

var (
	white = &color.RGBA{255, 255, 255, 255}
	black = &color.RGBA{0, 0, 0, 255}
)

type WelcomerCog struct {
	EventHandler *sandwich.Handlers
	Client       http.Client
}

// Assert types.

var (
	_ sandwich.Cog           = (*WelcomerCog)(nil)
	_ sandwich.CogWithEvents = (*WelcomerCog)(nil)
)

func NewWelcomerCog() *WelcomerCog {
	return &WelcomerCog{
		EventHandler: sandwich.SetupHandler(nil),
		Client:       http.Client{},
	}
}

func (p *WelcomerCog) CogInfo() *sandwich.CogInfo {
	return &sandwich.CogInfo{
		Name:        "Welcomer",
		Description: "Provides the functionality for the 'Welcomer' feature",
	}
}

func (p *WelcomerCog) GetEventHandlers() *sandwich.Handlers {
	return p.EventHandler
}

func (p *WelcomerCog) RegisterCog(bot *sandwich.Bot) error {
	// Register CustomEventInvokeWelcomer event.
	p.EventHandler.RegisterEventHandler(core.CustomEventInvokeWelcomer, func(eventCtx *sandwich.EventContext, payload sandwich_daemon.ProducedPayload) error {
		var invokeWelcomerPayload core.CustomEventInvokeWelcomerStructure
		if err := eventCtx.DecodeContent(payload, &invokeWelcomerPayload); err != nil {
			return fmt.Errorf("failed to unmarshal payload: %w", err)
		}

		if invokeWelcomerPayload.Member.GuildID != nil {
			eventCtx.Guild = sandwich.NewGuild(*invokeWelcomerPayload.Member.GuildID)
		}

		eventCtx.EventHandler.EventsMu.RLock()
		defer eventCtx.EventHandler.EventsMu.RUnlock()

		for _, event := range eventCtx.EventHandler.Events {
			if f, ok := event.(welcomer.OnInvokeWelcomerFuncType); ok {
				return eventCtx.Handlers.WrapFuncType(eventCtx, f(eventCtx, invokeWelcomerPayload))
			}
		}

		return nil
	})

	// Trigger CustomEventInvokeWelcomer when ON_GUILD_MEMBER_ADD event is received.
	p.EventHandler.RegisterOnGuildMemberAddEvent(func(eventCtx *sandwich.EventContext, member discord.GuildMember) error {
		// Query state cache for guild.
		guild, err := welcomer.FetchGuild(eventCtx.Context, eventCtx.Guild.ID)
		if err != nil {
			welcomer.Logger.Error().Err(err).
				Int64("guild_id", int64(eventCtx.Guild.ID)).
				Msg("Failed to fetch guild from state cache")

			return err
		}

		var usedInvite *discord.Invite
		var hasInviteVariable bool

		guildSettingsWelcomerText, guildSettingsWelcomerImages, guildSettingsWelcomerDMs, err := GetWelcomerSettings(eventCtx)
		if err == nil {
			hasInviteVariable = HasInviteVariable(guildSettingsWelcomerText, guildSettingsWelcomerImages, guildSettingsWelcomerDMs)

			if hasInviteVariable {
				usedInvite, err = p.trackInvites(eventCtx, eventCtx.Guild.ID)
				if err != nil {
					welcomer.Logger.Warn().Err(err).
						Int64("guild_id", int64(eventCtx.Guild.ID)).
						Msg("Failed to track invites")
				}
			}
		}

		welcomer.PushGuildScience.Push(
			eventCtx.Context,
			eventCtx.Guild.ID,
			member.User.ID,
			database.ScienceGuildEventTypeUserJoin,
			core.GuildScienceUserJoined{
				HasInviteTracking: hasInviteVariable,
				IsInviteTracked:   usedInvite != nil,
				InviteCode: welcomer.IfFunc(
					usedInvite != nil,
					func() string { return usedInvite.Code },
					func() string { return "" },
				),
				MemberCount: guild.MemberCount,
				IsPending:   member.Pending,
			},
		)

		if !member.Pending {
			return p.OnInvokeWelcomerEvent(eventCtx, core.CustomEventInvokeWelcomerStructure{
				Interaction: nil,
				Member:      member,
			})
		}

		return nil
	})

	// Trigger CustomEventInvokeWelcomer if user has moved from pending to non-pending.
	p.EventHandler.RegisterOnGuildMemberUpdateEvent(func(eventCtx *sandwich.EventContext, before, after discord.GuildMember) error {
		if before.Pending && !after.Pending {
			return p.OnInvokeWelcomerEvent(eventCtx, core.CustomEventInvokeWelcomerStructure{
				Interaction: nil,
				Member:      after,
			})
		}

		return nil
	})

	// Call OnInvokeWelcomerEvent when CustomEventInvokeWelcomer is triggered.
	p.EventHandler.RegisterEvent(core.CustomEventInvokeWelcomer, nil, (welcomer.OnInvokeWelcomerFuncType)(p.OnInvokeWelcomerEvent))

	return nil
}

func (p *WelcomerCog) FetchWelcomerImage(options welcomer.GenerateImageOptionsRaw) (io.ReadCloser, string, error) {
	optionsJSON, _ := json.Marshal(options)

	resp, err := p.Client.Post(os.Getenv("IMAGE_ADDRESS")+"/generate", "application/json", bytes.NewBuffer(optionsJSON))
	if err != nil || resp == nil {
		return nil, "", fmt.Errorf("fetch welcomer.image request failed: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return nil, "", fmt.Errorf("failed to get welcomer.image with status %s", resp.Status)
	}

	return resp.Body, resp.Header.Get("Content-Type"), nil
}

func (p *WelcomerCog) trackInvites(eventCtx *sandwich.EventContext, guildID discord.Snowflake) (*discord.Invite, error) {
	var potentialInvite *discord.Invite

	invites, err := discord.GetGuildInvites(eventCtx.Context, eventCtx.Session, guildID)
	if err != nil {
		return nil, err
	}

	databaseInvites, err := welcomer.Queries.GetGuildInvites(eventCtx.Context, int64(guildID))
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	type inviteData struct {
		Uses  int32
		Index int
	}

	beforeInvites := make(map[string]inviteData)
	for i, beforeInvite := range databaseInvites {
		beforeInvites[beforeInvite.InviteCode] = inviteData{
			Uses:  int32(beforeInvite.Uses),
			Index: i,
		}
	}

	updatedInvites := make([]database.GuildInvites, 0)

	for _, invite := range invites {
		beforeInvite, ok := beforeInvites[invite.Code]
		if !ok {
			// New invite.

			// Is it the first time we've seen it?
			if invite.Uses == 1 {
				if potentialInvite == nil {
					potentialInvite = &invite
				} else {
					// Multiple new invites with 1 use, now none are potential.
					potentialInvite = nil
				}
			}

			var createdBy int64
			if invite.Inviter != nil {
				createdBy = int64(invite.Inviter.ID)
			}

			updatedInvites = append(updatedInvites, database.GuildInvites{
				InviteCode: invite.Code,
				GuildID:    int64(guildID),
				CreatedBy:  createdBy,
				CreatedAt:  invite.CreatedAt,
				Uses:       int64(invite.Uses),
			})
		} else {
			if invite.Uses-beforeInvite.Uses == 1 {
				if potentialInvite == nil {
					potentialInvite = &invite
				} else {
					// Multiple invites have changed uses, now none are potential.
					potentialInvite = nil
				}
			}

			databaseInvites[beforeInvite.Index].Uses = int64(invite.Uses)

			updatedInvites = append(updatedInvites, *databaseInvites[beforeInvite.Index])
		}
	}

	for _, updatedInvite := range updatedInvites {
		_, err := welcomer.Queries.CreateOrUpdateGuildInvites(eventCtx.Context, database.CreateOrUpdateGuildInvitesParams(updatedInvite))
		if err != nil {
			welcomer.Logger.Warn().Err(err).
				Int64("guild_id", int64(guildID)).
				Str("invite_code", updatedInvite.InviteCode).
				Msg("Failed to create or update guild invite")
		}
	}

	return potentialInvite, nil
}

func GetWelcomerSettings(eventCtx *sandwich.EventContext) (*database.GuildSettingsWelcomerText, *database.GuildSettingsWelcomerImages, *database.GuildSettingsWelcomerDms, error) {
	var err error

	var guildSettingsWelcomerText *database.GuildSettingsWelcomerText

	guildSettingsWelcomerText, err = welcomer.Queries.GetWelcomerTextGuildSettings(eventCtx.Context, int64(eventCtx.Guild.ID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			guildSettingsWelcomerText = &database.GuildSettingsWelcomerText{
				GuildID:       int64(eventCtx.Guild.ID),
				ToggleEnabled: welcomer.DefaultWelcomerText.ToggleEnabled,
				Channel:       welcomer.DefaultWelcomerText.Channel,
				MessageFormat: welcomer.DefaultWelcomerText.MessageFormat,
			}
		} else {
			welcomer.Logger.Error().Err(err).
				Int64("guild_id", int64(eventCtx.Guild.ID)).
				Msg("Failed to get welcomer text guild settings")

			return nil, nil, nil, err
		}
	}

	var guildSettingsWelcomerImages *database.GuildSettingsWelcomerImages

	guildSettingsWelcomerImages, err = welcomer.Queries.GetWelcomerImagesGuildSettings(eventCtx.Context, int64(eventCtx.Guild.ID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			guildSettingsWelcomerImages = &database.GuildSettingsWelcomerImages{
				GuildID:                int64(eventCtx.Guild.ID),
				ToggleEnabled:          welcomer.DefaultWelcomerImages.ToggleEnabled,
				ToggleImageBorder:      welcomer.DefaultWelcomerImages.ToggleImageBorder,
				ToggleShowAvatar:       welcomer.DefaultWelcomerImages.ToggleShowAvatar,
				BackgroundName:         welcomer.DefaultWelcomerImages.BackgroundName,
				ColourText:             welcomer.DefaultWelcomerImages.ColourText,
				ColourTextBorder:       welcomer.DefaultWelcomerImages.ColourTextBorder,
				ColourImageBorder:      welcomer.DefaultWelcomerImages.ColourImageBorder,
				ColourProfileBorder:    welcomer.DefaultWelcomerImages.ColourProfileBorder,
				ImageAlignment:         welcomer.DefaultWelcomerImages.ImageAlignment,
				ImageTheme:             welcomer.DefaultWelcomerImages.ImageTheme,
				ImageMessage:           welcomer.DefaultWelcomerImages.ImageMessage,
				ImageProfileBorderType: welcomer.DefaultWelcomerImages.ImageProfileBorderType,
			}
		} else {
			welcomer.Logger.Error().Err(err).
				Int64("guild_id", int64(eventCtx.Guild.ID)).
				Msg("Failed to get welcomer.image guild settings")

			return nil, nil, nil, err
		}
	}

	var guildSettingsWelcomerDMs *database.GuildSettingsWelcomerDms

	guildSettingsWelcomerDMs, err = welcomer.Queries.GetWelcomerDMsGuildSettings(eventCtx.Context, int64(eventCtx.Guild.ID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			guildSettingsWelcomerDMs = &database.GuildSettingsWelcomerDms{
				GuildID:             int64(eventCtx.Guild.ID),
				ToggleEnabled:       welcomer.DefaultWelcomerDms.ToggleEnabled,
				ToggleUseTextFormat: welcomer.DefaultWelcomerDms.ToggleUseTextFormat,
				ToggleIncludeImage:  welcomer.DefaultWelcomerDms.ToggleIncludeImage,
				MessageFormat:       welcomer.DefaultWelcomerDms.MessageFormat,
			}
		} else {
			welcomer.Logger.Error().Err(err).
				Int64("guild_id", int64(eventCtx.Guild.ID)).
				Msg("Failed to get welcomer dm guild settings")

			return nil, nil, nil, err
		}
	}

	return guildSettingsWelcomerText, guildSettingsWelcomerImages, guildSettingsWelcomerDMs, nil
}

func ShouldTrackInvites(eventCtx *sandwich.EventContext, event core.CustomEventInvokeWelcomerStructure) (bool, error) {
	// Fetch guild settings.
	guildSettingsWelcomerText, guildSettingsWelcomerImages, guildSettingsWelcomerDMs, err := GetWelcomerSettings(eventCtx)
	if err != nil {
		return false, err
	}

	// Quit if nothing is enabled.
	if !guildSettingsWelcomerText.ToggleEnabled && !guildSettingsWelcomerImages.ToggleEnabled && !guildSettingsWelcomerDMs.ToggleEnabled {
		return false, nil
	}

	return HasInviteVariable(guildSettingsWelcomerText, guildSettingsWelcomerImages, guildSettingsWelcomerDMs), nil
}

func HasInviteVariable(guildSettingsWelcomerText *database.GuildSettingsWelcomerText, guildSettingsWelcomerImages *database.GuildSettingsWelcomerImages, guildSettingsWelcomerDMs *database.GuildSettingsWelcomerDms) bool {
	// Check if the welcomer text, dms or images possibly has an invite variable. This also checks if the module is enabled or not.
	hasInviteVariable := ((guildSettingsWelcomerText.ToggleEnabled || (guildSettingsWelcomerDMs.ToggleEnabled && guildSettingsWelcomerDMs.ToggleUseTextFormat)) && strings.Contains(string(guildSettingsWelcomerText.MessageFormat.Bytes), "{{Invite")) ||
		((guildSettingsWelcomerDMs.ToggleEnabled && !guildSettingsWelcomerDMs.ToggleUseTextFormat) && strings.Contains(string(guildSettingsWelcomerDMs.MessageFormat.Bytes), "{{Invite")) ||
		((guildSettingsWelcomerImages.ToggleEnabled) && strings.Contains(guildSettingsWelcomerImages.ImageMessage, "{{Invite"))

	return hasInviteVariable
}

// OnInvokeWelcomerEvent is called when CustomEventInvokeWelcomer is triggered.
// This can be from when a user joins or a user uses /welcomer test.
func (p *WelcomerCog) OnInvokeWelcomerEvent(eventCtx *sandwich.EventContext, event core.CustomEventInvokeWelcomerStructure) (err error) {
	defer func() {
		if err != nil {
			welcomer.Logger.Error().Err(err).Msg("Failed to execute welcomer event")
		}

		// Send follow-up if present.
		if event.Interaction != nil && event.Interaction.Token != "" {
			var message discord.WebhookMessageParams

			if err == nil {
				message = discord.WebhookMessageParams{
					Embeds: welcomer.NewEmbed("Executed successfully", welcomer.EmbedColourSuccess),
				}
			} else {
				message = discord.WebhookMessageParams{
					Embeds: welcomer.NewEmbed(fmt.Sprintf("Failed to execute: `%s`", err.Error()), welcomer.EmbedColourError),
				}
			}

			_, err = event.Interaction.SendFollowup(eventCtx.Context, eventCtx.Session, message)
			if err != nil {
				welcomer.Logger.Warn().Err(err).
					Int64("guild_id", int64(eventCtx.Guild.ID)).
					Int64("application_id", int64(event.Interaction.ApplicationID)).
					Str("token", event.Interaction.Token).
					Msg("Failed to send interaction follow-up")
			}
		}
	}()

	// Fetch guild settings.

	guildSettingsWelcomerText, guildSettingsWelcomerImages, guildSettingsWelcomerDMs, err := GetWelcomerSettings(eventCtx)
	if err != nil {
		return err
	}

	// Quit if nothing is enabled.
	if !guildSettingsWelcomerText.ToggleEnabled && !guildSettingsWelcomerImages.ToggleEnabled && !guildSettingsWelcomerDMs.ToggleEnabled {
		return nil
	}

	var user *discord.User

	// Query state cache for user if welcomer DMs are enabled.
	// This is for fetching direct message channels for the user.
	if guildSettingsWelcomerDMs.ToggleEnabled {
		// Query state cache for user.
		var usersPb *sandwich_protobuf.FetchUserResponse

		usersPb, err = welcomer.SandwichClient.FetchUser(eventCtx, &sandwich_protobuf.FetchUserRequest{
			UserIds: []int64{int64(event.Member.User.ID)},
		})
		if err != nil {
			return err
		}

		userPb, ok := usersPb.Users[int64(event.Member.User.ID)]
		if ok {
			user = sandwich_protobuf.PBToUser(userPb)
		} else {
			welcomer.Logger.Warn().
				Int64("user_id", int64(event.Member.User.ID)).
				Msg("Failed to get user from state cache, falling back to event.Member.User")
			user = event.Member.User
		}
	} else {
		user = event.Member.User
	}

	var guilds *sandwich_protobuf.FetchGuildResponse

	// Query state cache for guild.
	guilds, err = welcomer.SandwichClient.FetchGuild(eventCtx, &sandwich_protobuf.FetchGuildRequest{
		GuildIds: []int64{int64(eventCtx.Guild.ID)},
	})
	if err != nil {
		return err
	}

	var guild *discord.Guild

	guildPb, ok := guilds.Guilds[int64(eventCtx.Guild.ID)]
	if ok {
		guild = sandwich_protobuf.PBToGuild(guildPb)
	}

	var usedInvite *discord.Invite

	var joinEvent *welcomer.GuildScienceUserJoined

	// Look through the buffer for the event before checking the database.
	welcomer.PushGuildScience.RLock()
	for _, scienceEvent := range welcomer.PushGuildScience.Buffer {
		if discord.Snowflake(scienceEvent.GuildID) == eventCtx.Guild.ID && discord.Snowflake(scienceEvent.UserID.Int64) == event.Member.User.ID {
			if scienceEvent.EventType == int32(database.ScienceGuildEventTypeUserJoin) {
				joinEvent = &welcomer.GuildScienceUserJoined{false, false, "", 0, false}

				err = json.Unmarshal(scienceEvent.Data.Bytes, joinEvent)
				if err != nil {
					welcomer.Logger.Warn().Err(err).
						Msg("Failed to unmarshal guild science user joined event")

					continue
				}
			} else if database.ScienceGuildEventType(scienceEvent.EventType) == database.ScienceGuildEventTypeUserLeave {
				joinEvent = nil
			}
		}
	}
	welcomer.PushGuildScience.RUnlock()

	// Override the member count if we have a join event.
	if joinEvent != nil && joinEvent.MemberCount > 0 {
		guild.MemberCount = joinEvent.MemberCount
	}

	hasInviteVariable := HasInviteVariable(guildSettingsWelcomerText, guildSettingsWelcomerImages, guildSettingsWelcomerDMs)

	// Handle invite tracking.
	if hasInviteVariable {
		// If we found the event in the buffer, get the invite from the database.
		if joinEvent != nil && joinEvent.InviteCode != "" {
			invite, err := welcomer.Queries.GetGuildInvite(eventCtx.Context, database.GetGuildInviteParams{
				InviteCode: joinEvent.InviteCode,
				GuildID:    int64(eventCtx.Guild.ID),
			})
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				welcomer.Logger.Warn().Err(err).
					Int64("guild_id", int64(eventCtx.Guild.ID)).
					Str("invite_code", joinEvent.InviteCode).
					Msg("Failed to get guild invite")
			} else {
				welcomer.Logger.Info().
					Int64("guild_id", int64(eventCtx.Guild.ID)).
					Int64("user_id", int64(event.Member.User.ID)).
					Str("invite_code", invite.InviteCode).
					Msg("Received invite from buffer")

				usedInvite = &discord.Invite{
					CreatedAt: invite.CreatedAt,
					Inviter: &discord.User{
						ID: discord.Snowflake(invite.CreatedBy),
					},
					Code: invite.InviteCode,
					Uses: int32(invite.Uses),
				}
			}
		}

		// If the event was not in the buffer or the invite was not found, check the database for the event.
		if usedInvite == nil {
			welcomer.Logger.Info().
				Int64("guild_id", int64(eventCtx.Guild.ID)).
				Int64("user_id", int64(event.Member.User.ID)).
				Msg("Invite not found in buffer, checking database")

			recentEvent, err := welcomer.Queries.GetScienceGuildJoinLeaveEventForUser(eventCtx.Context, database.GetScienceGuildJoinLeaveEventForUserParams{
				EventType:   int32(database.ScienceGuildEventTypeUserJoin),
				EventType_2: int32(database.ScienceGuildEventTypeUserLeave),
				GuildID:     int64(eventCtx.Guild.ID),
				UserID:      sql.NullInt64{Int64: int64(event.Member.User.ID), Valid: true},
			})
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				welcomer.Logger.Warn().Err(err).
					Int64("guild_id", int64(eventCtx.Guild.ID)).
					Int64("user_id", int64(event.Member.User.ID)).
					Msg("Failed to get guild join leave event for user")
			}

			if recentEvent != nil &&
				database.ScienceGuildEventType(recentEvent.EventType) == database.ScienceGuildEventTypeGuildJoin &&
				recentEvent.InviteCode.Valid {

				welcomer.Logger.Info().
					Int64("guild_id", int64(eventCtx.Guild.ID)).
					Int64("user_id", int64(event.Member.User.ID)).
					Msg("Received invite from database")

				usedInvite = &discord.Invite{
					CreatedAt: recentEvent.CreatedAt_2.Time,
					Inviter: &discord.User{
						ID: discord.Snowflake(recentEvent.CreatedBy.Int64),
					},
					Code: recentEvent.InviteCode.String,
					Uses: int32(recentEvent.Uses.Int64),
				}
				// TODO: store more invite data or fetch from DC
			}
		}

		// If the invite was not found in the database, check the invites from the API.
		if usedInvite == nil {
			welcomer.Logger.Info().
				Int64("guild_id", int64(eventCtx.Guild.ID)).
				Int64("user_id", int64(event.Member.User.ID)).
				Msg("Invite not found in database, checking API")

			usedInvite, err = p.trackInvites(eventCtx, eventCtx.Guild.ID)
			if err != nil {
				welcomer.Logger.Warn().Err(err).
					Int64("guild_id", int64(eventCtx.Guild.ID)).
					Msg("Failed to track invites")
			}
		}
	}

	functions := welcomer.GatherFunctions()
	variables := welcomer.GatherVariables(eventCtx, &event.Member, guild, usedInvite, nil)

	var serverMessage discord.MessageParams
	var directMessage discord.MessageParams

	var file *discord.File

	// If welcomer images are enabled, prepare an image.
	if guildSettingsWelcomerImages.ToggleEnabled {
		var messageFormat string

		messageFormat, err = welcomer.FormatString(functions, variables, guildSettingsWelcomerImages.ImageMessage)
		if err != nil {
			welcomer.Logger.Error().Err(err).
				Int64("guild_id", int64(eventCtx.Guild.ID)).
				Int64("user_id", int64(event.Member.User.ID)).
				Msg("Failed to format welcomer text payload")

			return err
		}

		var memberships []*database.GetUserMembershipsByGuildIDRow

		// Check if the guild has welcomer pro.
		memberships, err = welcomer.Queries.GetValidUserMembershipsByGuildID(eventCtx.Context, eventCtx.Guild.ID, time.Now())
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			welcomer.Logger.Warn().Err(err).
				Int64("guild_id", int64(eventCtx.Guild.ID)).
				Msg("Failed to get welcomer memberships")

			return err
		}

		hasWelcomerPro, _ := welcomer.CheckGuildMemberships(memberships)

		var borderWidth int32
		if guildSettingsWelcomerImages.ToggleImageBorder {
			borderWidth = DefaultImageBorderWidth
		} else {
			borderWidth = 0
		}

		var profileFloat welcomer.ImageAlignment
		switch guildSettingsWelcomerImages.ImageTheme {
		case int32(welcomer.ImageThemeDefault):
			profileFloat = welcomer.ImageAlignmentLeft
		case int32(welcomer.ImageThemeVertical):
			profileFloat = welcomer.ImageAlignmentCenter
		case int32(welcomer.ImageThemeCard):
			profileFloat = welcomer.ImageAlignmentLeft
		}

		var imageReaderCloser io.ReadCloser
		var contentType string

		// Fetch the welcomer.image.
		imageReaderCloser, contentType, err = p.FetchWelcomerImage(welcomer.GenerateImageOptionsRaw{
			ShowAvatar:         guildSettingsWelcomerImages.ToggleShowAvatar,
			GuildID:            int64(eventCtx.Guild.ID),
			UserID:             int64(event.Member.User.ID),
			AllowAnimated:      hasWelcomerPro,
			AvatarURL:          welcomer.GetUserAvatar(event.Member.User),
			Theme:              guildSettingsWelcomerImages.ImageTheme,
			Background:         guildSettingsWelcomerImages.BackgroundName,
			Text:               messageFormat,
			TextFont:           DefaultFont,
			TextStroke:         true,
			TextAlign:          guildSettingsWelcomerImages.ImageAlignment,
			TextColor:          tryParseColourAsInt64(guildSettingsWelcomerImages.ColourText, white),
			TextStrokeColor:    tryParseColourAsInt64(guildSettingsWelcomerImages.ColourTextBorder, black),
			ImageBorderColor:   tryParseColourAsInt64(guildSettingsWelcomerImages.ColourImageBorder, white),
			ImageBorderWidth:   borderWidth,
			ProfileFloat:       int32(profileFloat),
			ProfileBorderColor: tryParseColourAsInt64(guildSettingsWelcomerImages.ColourProfileBorder, white),
			ProfileBorderWidth: DefaultProfileBorderWidth,
			ProfileBorderCurve: guildSettingsWelcomerImages.ImageProfileBorderType,
		})
		if err != nil {
			welcomer.Logger.Warn().Err(err).
				Int64("guild_id", int64(eventCtx.Guild.ID)).
				Int64("user_id", int64(event.Member.User.ID)).
				Msg("Failed to get welcomer.image")
		}

		if imageReaderCloser != nil {
			defer imageReaderCloser.Close()

			var imageFileType welcomer.ImageFileType

			if err := imageFileType.UnmarshalText([]byte(contentType)); err != nil {
				imageFileType = welcomer.ImageFileTypeUnknown
			}

			file = &discord.File{
				Name:        "welcome-" + eventCtx.Guild.ID.String() + "-" + event.Member.User.ID.String() + "." + imageFileType.GetExtension(),
				ContentType: contentType,
				Reader:      imageReaderCloser,
			}
		}
	}

	// If welcomer text or images are enabled, prepare to send a message.
	if guildSettingsWelcomerText.ToggleEnabled || guildSettingsWelcomerImages.ToggleEnabled {
		// If welcomer text is enabled but no channel is set, return an error.
		if guildSettingsWelcomerText.Channel == 0 {
			// If welcomer dms are enabled, then we can continue without an error.
			if !guildSettingsWelcomerDMs.ToggleEnabled {
				return welcomer.ErrMissingChannel
			}
		} else {
			if guildSettingsWelcomerText.ToggleEnabled && !welcomer.IsJSONBEmpty(guildSettingsWelcomerText.MessageFormat.Bytes) {
				var messageFormat string

				messageFormat, err = welcomer.FormatString(functions, variables, strconv.B2S(guildSettingsWelcomerText.MessageFormat.Bytes))
				if err != nil {
					welcomer.Logger.Error().Err(err).
						Int64("guild_id", int64(eventCtx.Guild.ID)).
						Int64("user_id", int64(event.Member.User.ID)).
						Msg("Failed to format welcomer DMs payload")

					return err
				}

				// Convert MessageFormat to MessageParams so we can send it.
				err = json.Unmarshal(strconv.S2B(messageFormat), &serverMessage)
				if err != nil {
					welcomer.Logger.Error().Err(err).
						Int64("guild_id", int64(eventCtx.Guild.ID)).
						Int64("user_id", int64(event.Member.User.ID)).
						Msg("Failed to unmarshal welcomer messageFormat")

					return err
				}
			}

			if file != nil {
				serverMessage.AddFile(*file)

				if len(serverMessage.Embeds) == 0 {
					serverMessage.AddEmbed(discord.Embed{})
				}

				serverMessage.Embeds[0].SetImage(discord.NewEmbedImage("attachment://" + file.Name))
			}
		}
	}

	if guildSettingsWelcomerDMs.ToggleEnabled {
		if guildSettingsWelcomerDMs.ToggleUseTextFormat {
			if !welcomer.IsJSONBEmpty(guildSettingsWelcomerText.MessageFormat.Bytes) {
				var messageFormat string

				messageFormat, err = welcomer.FormatString(functions, variables, strconv.B2S(guildSettingsWelcomerText.MessageFormat.Bytes))
				if err != nil {
					welcomer.Logger.Error().Err(err).
						Int64("guild_id", int64(eventCtx.Guild.ID)).
						Int64("user_id", int64(event.Member.User.ID)).
						Msg("Failed to format welcomer DMs payload")

					return err
				}

				// Convert MessageFormat to MessageParams so we can send it.
				err = json.Unmarshal(strconv.S2B(messageFormat), &directMessage)
				if err != nil {
					welcomer.Logger.Error().Err(err).
						Int64("guild_id", int64(eventCtx.Guild.ID)).
						Int64("user_id", int64(event.Member.User.ID)).
						Msg("Failed to unmarshal welcomer messageFormat")

					return err
				}
			}
		} else {
			if !welcomer.IsJSONBEmpty(guildSettingsWelcomerDMs.MessageFormat.Bytes) {
				var messageFormat string

				messageFormat, err = welcomer.FormatString(functions, variables, strconv.B2S(guildSettingsWelcomerDMs.MessageFormat.Bytes))
				if err != nil {
					welcomer.Logger.Error().Err(err).
						Int64("guild_id", int64(eventCtx.Guild.ID)).
						Int64("user_id", int64(event.Member.User.ID)).
						Msg("Failed to format welcomer DMs payload")

					return err
				}

				// Convert MessageFormat to MessageParams so we can send it.
				err = json.Unmarshal(strconv.S2B(messageFormat), &directMessage)
				if err != nil {
					welcomer.Logger.Error().Err(err).
						Int64("guild_id", int64(eventCtx.Guild.ID)).
						Int64("user_id", int64(event.Member.User.ID)).
						Msg("Failed to unmarshal welcomer dms messageFormat")

					return err
				}
			}
		}
	}

	var serr error
	var dmerr error

	// Send server message if it's not empty.
	if !welcomer.IsMessageParamsEmpty(serverMessage) {
		validGuild, err := core.CheckChannelGuild(eventCtx.Context, welcomer.SandwichClient, eventCtx.Guild.ID, discord.Snowflake(guildSettingsWelcomerText.Channel))
		if err != nil {
			welcomer.Logger.Error().Err(err).
				Int64("guild_id", int64(eventCtx.Guild.ID)).
				Int64("channel_id", guildSettingsWelcomerText.Channel).
				Msg("Failed to check channel guild")
		} else if !validGuild {
			welcomer.Logger.Warn().
				Int64("guild_id", int64(eventCtx.Guild.ID)).
				Int64("channel_id", guildSettingsWelcomerText.Channel).
				Msg("Channel does not belong to guild")
		} else {
			channel := discord.Channel{ID: discord.Snowflake(guildSettingsWelcomerText.Channel)}

			_, serr = channel.Send(eventCtx.Context, eventCtx.Session, serverMessage)

			welcomer.Logger.Info().
				Int64("guild_id", int64(eventCtx.Guild.ID)).
				Int64("channel_id", guildSettingsWelcomerText.Channel).
				Msg("Sent welcomer message to channel")

			if serr != nil {
				welcomer.Logger.Warn().Err(serr).
					Int64("guild_id", int64(eventCtx.Guild.ID)).
					Int64("channel_id", guildSettingsWelcomerText.Channel).
					Msg("Failed to send welcomer message to channel")
			}
		}
	}

	// Send direct message if it's not empty.
	if !welcomer.IsMessageParamsEmpty(directMessage) {
		directMessage = welcomer.IncludeSentByButton(directMessage, guild.Name)
		directMessage = welcomer.IncludeScamsButton(directMessage)

		_, dmerr = user.Send(eventCtx.Context, eventCtx.Session, directMessage)

		welcomer.Logger.Info().
			Int64("guild_id", int64(eventCtx.Guild.ID)).
			Int64("user_id", int64(event.Member.User.ID)).
			Msg("Sent welcomer DMs to user")

		if dmerr != nil {
			welcomer.Logger.Warn().Err(dmerr).
				Int64("guild_id", int64(eventCtx.Guild.ID)).
				Int64("user_id", int64(event.Member.User.ID)).
				Msg("Failed to send message to user")
		}
	}

	if serr != nil {
		err = serr
	} else if dmerr != nil {
		err = dmerr
	}

	welcomer.PushGuildScience.Push(
		eventCtx.Context,
		eventCtx.Guild.ID,
		event.Member.User.ID,
		database.ScienceGuildEventTypeUserWelcomed,
		welcomer.GuildScienceUserWelcomed{
			HasImage:          file != nil,
			HasMessage:        !welcomer.IsMessageParamsEmpty(serverMessage),
			HasDM:             !welcomer.IsMessageParamsEmpty(directMessage),
			HasInviteTracking: hasInviteVariable,
			IsInviteTracked:   usedInvite != nil,
			InviteCode: welcomer.IfFunc(
				usedInvite != nil,
				func() string { return usedInvite.Code },
				func() string { return "" },
			),
		})

	return err
}

func tryParseColourAsInt64(str string, defaultValue *color.RGBA) int64 {
	c, err := welcomer.ParseColour(str, "")
	if err != nil {
		c = defaultValue
	}

	return (int64(c.A) << 24) + (int64(c.R) << 16) + (int64(c.G) << 8) + int64(c.B)
}
