package plugins

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/WelcomerTeam/Discord/discord"
	pb "github.com/WelcomerTeam/Sandwich-Daemon/protobuf"
	"github.com/WelcomerTeam/Sandwich-Daemon/structs"
	sandwich "github.com/WelcomerTeam/Sandwich/sandwich"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	core "github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	utils "github.com/WelcomerTeam/Welcomer/welcomer-utils"
	"github.com/jackc/pgx/v4"
	"github.com/savsgio/gotils/strconv"
	"image/color"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
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
	p.EventHandler.RegisterEventHandler(core.CustomEventInvokeWelcomer, func(eventCtx *sandwich.EventContext, payload structs.SandwichPayload) error {

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
		if !member.Pending {
			return p.OnInvokeWelcomerEvent(eventCtx, core.CustomEventInvokeWelcomerStructure{
				Interaction: nil,
				Member:      member,
			})
		}

		return nil
	})

	// Trigger CustomEventInvokeutils.if user has moved from pending to non-pending.
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

func (p *WelcomerCog) FetchWelcomerImage(options utils.GenerateImageOptionsRaw) (io.ReadCloser, string, error) {
	optionsJSON, _ := json.Marshal(options)

	resp, err := p.Client.Post(os.Getenv("IMAGE_ADDRESS")+"/generate", "application/json", bytes.NewBuffer(optionsJSON))
	if err != nil || resp == nil {
		return nil, "", fmt.Errorf("fetch utils.image request failed: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return nil, "", fmt.Errorf("failed to get utils.image with status %s", resp.Status)
	}

	return resp.Body, resp.Header.Get("Content-Type"), nil
}

func (p *WelcomerCog) trackInvites(eventCtx *sandwich.EventContext, guildID discord.Snowflake) (*discord.Invite, error) {
	var potentialInvite *discord.Invite

	invites, err := discord.GetGuildInvites(eventCtx.Session, guildID)
	if err != nil {
		return nil, err
	}

	queries := welcomer.GetQueriesFromContext(eventCtx.Context)

	databaseInvites, err := queries.GetGuildInvites(eventCtx.Context, int64(guildID))
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
		_, err := queries.CreateOrUpdateGuildInvites(eventCtx.Context, database.CreateOrUpdateGuildInvitesParams(updatedInvite))
		if err != nil {
			eventCtx.Logger.Warn().Err(err).
				Int64("guild_id", int64(guildID)).
				Str("invite_code", updatedInvite.InviteCode).
				Msg("Failed to create or update guild invite")
		}
	}

	return potentialInvite, nil
}

// OnInvokeWelcomerEvent is called when CustomEventInvokeWelcomer is triggered.
// This can be from when a user joins or a user uses /welcomer test.
func (p *WelcomerCog) OnInvokeWelcomerEvent(eventCtx *sandwich.EventContext, event core.CustomEventInvokeWelcomerStructure) (err error) {
	defer func() {
		if err != nil {
			eventCtx.Logger.Error().Err(err).Msg("Failed to execute welcomer event")
		}

		// Send follow-up if present.
		if event.Interaction != nil && event.Interaction.Token != "" {
			var message discord.WebhookMessageParams

			if err == nil {
				message = discord.WebhookMessageParams{
					Embeds: utils.NewEmbed("Executed successfully", utils.EmbedColourSuccess),
				}
			} else {
				message = discord.WebhookMessageParams{
					Embeds: utils.NewEmbed(fmt.Sprintf("Failed to execute: `%s`", err.Error()), utils.EmbedColourError),
				}
			}

			_, err = event.Interaction.SendFollowup(eventCtx.Session, message)
			if err != nil {
				eventCtx.Logger.Warn().Err(err).
					Int64("guild_id", int64(eventCtx.Guild.ID)).
					Int64("application_id", int64(event.Interaction.ApplicationID)).
					Str("token", event.Interaction.Token).
					Msg("Failed to send interaction follow-up")
			}
		}
	}()

	queries := welcomer.GetQueriesFromContext(eventCtx.Context)

	// Fetch guild settings.

	var guildSettingsWelcomerText *database.GuildSettingsWelcomerText

	guildSettingsWelcomerText, err = queries.GetWelcomerTextGuildSettings(eventCtx.Context, int64(eventCtx.Guild.ID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			guildSettingsWelcomerText = &database.GuildSettingsWelcomerText{
				GuildID:       int64(eventCtx.Guild.ID),
				ToggleEnabled: database.DefaultWelcomerText.ToggleEnabled,
				Channel:       database.DefaultWelcomerText.Channel,
				MessageFormat: database.DefaultWelcomerText.MessageFormat,
			}
		} else {
			eventCtx.Logger.Error().Err(err).
				Int64("guild_id", int64(eventCtx.Guild.ID)).
				Msg("Failed to get welcomer text guild settings")

			return err
		}
	}

	var guildSettingsWelcomerImages *database.GuildSettingsWelcomerImages

	guildSettingsWelcomerImages, err = queries.GetWelcomerImagesGuildSettings(eventCtx.Context, int64(eventCtx.Guild.ID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			guildSettingsWelcomerImages = &database.GuildSettingsWelcomerImages{
				GuildID:                int64(eventCtx.Guild.ID),
				ToggleEnabled:          database.DefaultWelcomerImages.ToggleEnabled,
				ToggleImageBorder:      database.DefaultWelcomerImages.ToggleImageBorder,
				ToggleShowAvatar:       database.DefaultWelcomerImages.ToggleShowAvatar,
				BackgroundName:         database.DefaultWelcomerImages.BackgroundName,
				ColourText:             database.DefaultWelcomerImages.ColourText,
				ColourTextBorder:       database.DefaultWelcomerImages.ColourTextBorder,
				ColourImageBorder:      database.DefaultWelcomerImages.ColourImageBorder,
				ColourProfileBorder:    database.DefaultWelcomerImages.ColourProfileBorder,
				ImageAlignment:         database.DefaultWelcomerImages.ImageAlignment,
				ImageTheme:             database.DefaultWelcomerImages.ImageTheme,
				ImageMessage:           database.DefaultWelcomerImages.ImageMessage,
				ImageProfileBorderType: database.DefaultWelcomerImages.ImageProfileBorderType,
			}
		} else {
			eventCtx.Logger.Error().Err(err).
				Int64("guild_id", int64(eventCtx.Guild.ID)).
				Msg("Failed to get utils.image guild settings")

			return err
		}
	}

	var guildSettingsWelcomerDMs *database.GuildSettingsWelcomerDms

	guildSettingsWelcomerDMs, err = queries.GetWelcomerDMsGuildSettings(eventCtx.Context, int64(eventCtx.Guild.ID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			guildSettingsWelcomerDMs = &database.GuildSettingsWelcomerDms{
				GuildID:             int64(eventCtx.Guild.ID),
				ToggleEnabled:       database.DefaultWelcomerDms.ToggleEnabled,
				ToggleUseTextFormat: database.DefaultWelcomerDms.ToggleUseTextFormat,
				ToggleIncludeImage:  database.DefaultWelcomerDms.ToggleIncludeImage,
				MessageFormat:       database.DefaultWelcomerDms.MessageFormat,
			}
		} else {
			eventCtx.Logger.Error().Err(err).
				Int64("guild_id", int64(eventCtx.Guild.ID)).
				Msg("Failed to get welcomer dm guild settings")

			return err
		}
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
		var users *pb.UsersResponse

		users, err = eventCtx.Sandwich.SandwichClient.FetchUsers(eventCtx, &pb.FetchUsersRequest{
			UserIDs:         []int64{int64(event.Member.User.ID)},
			CreateDMChannel: true,
		})
		if err != nil {
			return err
		}

		userPb, ok := users.Users[int64(event.Member.User.ID)]
		if ok {
			var pUser discord.User

			pUser, err = pb.GRPCToUser(userPb)
			if err != nil {
				return err
			}

			user = &pUser
		} else {
			eventCtx.Logger.Warn().
				Int64("user_id", int64(event.Member.User.ID)).
				Msg("Failed to get user from state cache, falling back to event.Member.User")
			user = event.Member.User
		}
	} else {
		user = event.Member.User
	}

	var guilds *pb.GuildsResponse

	// Query state cache for guild.
	guilds, err = eventCtx.Sandwich.SandwichClient.FetchGuild(eventCtx, &pb.FetchGuildRequest{
		GuildIDs: []int64{int64(eventCtx.Guild.ID)},
	})
	if err != nil {
		return err
	}

	var guild discord.Guild
	guildPb, ok := guilds.Guilds[int64(eventCtx.Guild.ID)]
	if ok {
		guild, err = pb.GRPCToGuild(guildPb)
		if err != nil {
			return err
		}
	}

	// Check if the welcomer text, dms or images possibly has an invite variable. This also checks if the module is enabled or not.
	hasInviteVariable := ((guildSettingsWelcomerText.ToggleEnabled || (guildSettingsWelcomerDMs.ToggleEnabled && guildSettingsWelcomerDMs.ToggleUseTextFormat)) && strings.Contains(string(guildSettingsWelcomerText.MessageFormat.Bytes), "{{Invite")) ||
		((guildSettingsWelcomerDMs.ToggleEnabled && !guildSettingsWelcomerDMs.ToggleUseTextFormat) && strings.Contains(string(guildSettingsWelcomerDMs.MessageFormat.Bytes), "{{Invite")) ||
		((guildSettingsWelcomerImages.ToggleEnabled) && strings.Contains(guildSettingsWelcomerImages.ImageMessage, "{{Invite"))

	var usedInvite *discord.Invite

	if hasInviteVariable {
		usedInvite, err = p.trackInvites(eventCtx, eventCtx.Guild.ID)
		if err != nil {
			eventCtx.Logger.Warn().Err(err).
				Int64("guild_id", int64(eventCtx.Guild.ID)).
				Msg("Failed to track invites")
		}
	}

	functions := welcomer.GatherFunctions()
	variables := welcomer.GatherVariables(eventCtx, event.Member, guild, usedInvite, nil)

	var serverMessage discord.MessageParams
	var directMessage discord.MessageParams

	var file *discord.File

	// If welcomer images are enabled, prepare an image.
	if guildSettingsWelcomerImages.ToggleEnabled {
		var messageFormat string

		messageFormat, err = welcomer.FormatString(functions, variables, guildSettingsWelcomerImages.ImageMessage)
		if err != nil {
			eventCtx.Logger.Error().Err(err).
				Int64("guild_id", int64(eventCtx.Guild.ID)).
				Int64("user_id", int64(event.Member.User.ID)).
				Msg("Failed to format welcomer text payload")

			return err
		}

		var memberships []*database.GetUserMembershipsByGuildIDRow

		// Check if the guild has welcomer pro.
		memberships, err = queries.GetValidUserMembershipsByGuildID(eventCtx.Context, eventCtx.Guild.ID, time.Now())
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			eventCtx.Logger.Warn().Err(err).
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

		var profileFloat utils.ImageAlignment
		switch guildSettingsWelcomerImages.ImageTheme {
		case int32(utils.ImageThemeDefault):
			profileFloat = utils.ImageAlignmentLeft
		case int32(utils.ImageThemeVertical):
			profileFloat = utils.ImageAlignmentCenter
		case int32(utils.ImageThemeCard):
			profileFloat = utils.ImageAlignmentLeft
		}

		var imageReaderCloser io.ReadCloser
		var contentType string

		var avatarURL string

		if guildSettingsWelcomerImages.ToggleShowAvatar {
			avatarURL = welcomer.GetUserAvatar(event.Member.User)
		}

		// Fetch the utils.image.
		imageReaderCloser, contentType, err = p.FetchWelcomerImage(utils.GenerateImageOptionsRaw{
			ShowAvatar:         guildSettingsWelcomerImages.ToggleShowAvatar,
			GuildID:            int64(eventCtx.Guild.ID),
			UserID:             int64(event.Member.User.ID),
			AllowAnimated:      hasWelcomerPro,
			AvatarURL:          avatarURL,
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
			eventCtx.Logger.Warn().Err(err).
				Int64("guild_id", int64(eventCtx.Guild.ID)).
				Int64("user_id", int64(event.Member.User.ID)).
				Msg("Failed to get utils.image")
		}

		if imageReaderCloser != nil {
			defer imageReaderCloser.Close()

			var imageFileType utils.ImageFileType

			if err := imageFileType.UnmarshalText([]byte(contentType)); err != nil {
				imageFileType = utils.ImageFileTypeUnknown
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
				return utils.ErrMissingChannel
			}
		} else {
			if guildSettingsWelcomerText.ToggleEnabled && !utils.IsJSONBEmpty(guildSettingsWelcomerText.MessageFormat.Bytes) {
				var messageFormat string

				messageFormat, err = welcomer.FormatString(functions, variables, strconv.B2S(guildSettingsWelcomerText.MessageFormat.Bytes))
				if err != nil {
					eventCtx.Logger.Error().Err(err).
						Int64("guild_id", int64(eventCtx.Guild.ID)).
						Int64("user_id", int64(event.Member.User.ID)).
						Msg("Failed to format welcomer DMs payload")

					return err
				}

				// Convert MessageFormat to MessageParams so we can send it.
				err = json.Unmarshal(strconv.S2B(messageFormat), &serverMessage)
				if err != nil {
					eventCtx.Logger.Error().Err(err).
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
			if !utils.IsJSONBEmpty(guildSettingsWelcomerText.MessageFormat.Bytes) {
				var messageFormat string

				messageFormat, err = welcomer.FormatString(functions, variables, strconv.B2S(guildSettingsWelcomerText.MessageFormat.Bytes))
				if err != nil {
					eventCtx.Logger.Error().Err(err).
						Int64("guild_id", int64(eventCtx.Guild.ID)).
						Int64("user_id", int64(event.Member.User.ID)).
						Msg("Failed to format welcomer DMs payload")

					return err
				}

				// Convert MessageFormat to MessageParams so we can send it.
				err = json.Unmarshal(strconv.S2B(messageFormat), &directMessage)
				if err != nil {
					eventCtx.Logger.Error().Err(err).
						Int64("guild_id", int64(eventCtx.Guild.ID)).
						Int64("user_id", int64(event.Member.User.ID)).
						Msg("Failed to unmarshal welcomer messageFormat")

					return err
				}
			}
		} else {
			if !utils.IsJSONBEmpty(guildSettingsWelcomerDMs.MessageFormat.Bytes) {
				var messageFormat string

				messageFormat, err = welcomer.FormatString(functions, variables, strconv.B2S(guildSettingsWelcomerDMs.MessageFormat.Bytes))
				if err != nil {
					eventCtx.Logger.Error().Err(err).
						Int64("guild_id", int64(eventCtx.Guild.ID)).
						Int64("user_id", int64(event.Member.User.ID)).
						Msg("Failed to format welcomer DMs payload")

					return err
				}

				// Convert MessageFormat to MessageParams so we can send it.
				err = json.Unmarshal(strconv.S2B(messageFormat), &directMessage)
				if err != nil {
					eventCtx.Logger.Error().Err(err).
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
	if !utils.IsMessageParamsEmpty(serverMessage) {
		validGuild, err := core.CheckChannelGuild(eventCtx.Context, eventCtx.Sandwich.SandwichClient, eventCtx.Guild.ID, discord.Snowflake(guildSettingsWelcomerText.Channel))
		if err != nil {
			eventCtx.Logger.Error().Err(err).
				Int64("guild_id", int64(eventCtx.Guild.ID)).
				Int64("channel_id", guildSettingsWelcomerText.Channel).
				Msg("Failed to check channel guild")
		} else if !validGuild {
			eventCtx.Logger.Warn().
				Int64("guild_id", int64(eventCtx.Guild.ID)).
				Int64("channel_id", guildSettingsWelcomerText.Channel).
				Msg("Channel does not belong to guild")
		} else {
			channel := discord.Channel{ID: discord.Snowflake(guildSettingsWelcomerText.Channel)}

			_, serr = channel.Send(eventCtx.Session, serverMessage)

			eventCtx.Logger.Info().
				Int64("guild_id", int64(eventCtx.Guild.ID)).
				Int64("channel_id", guildSettingsWelcomerText.Channel).
				Msg("Sent welcomer message to channel")

			if serr != nil {
				eventCtx.Logger.Warn().Err(serr).
					Int64("guild_id", int64(eventCtx.Guild.ID)).
					Int64("channel_id", guildSettingsWelcomerText.Channel).
					Msg("Failed to send welcomer message to channel")
			}
		}
	}

	// Send direct message if it's not empty.
	if !utils.IsMessageParamsEmpty(directMessage) {
		directMessage = welcomer.IncludeSentByButton(directMessage, guild.Name)
		directMessage = welcomer.IncludeScamsButton(directMessage)

		_, dmerr = user.Send(eventCtx.Session, directMessage)

		eventCtx.Logger.Info().
			Int64("guild_id", int64(eventCtx.Guild.ID)).
			Int64("user_id", int64(event.Member.User.ID)).
			Msg("Sent welcomer DMs to user")

		if dmerr != nil {
			eventCtx.Logger.Warn().Err(dmerr).
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

	return nil
}

func tryParseColourAsInt64(str string, defaultValue *color.RGBA) int64 {
	c, err := utils.ParseColour(str, "")
	if err != nil {
		c = defaultValue
	}

	return (int64(c.A) << 24) + (int64(c.R) << 16) + (int64(c.G) << 8) + int64(c.B)
}
