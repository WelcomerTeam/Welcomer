package plugins

import (
	"bytes"
	"fmt"
	"image/color"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	pb "github.com/WelcomerTeam/Sandwich-Daemon/protobuf"
	"github.com/WelcomerTeam/Sandwich-Daemon/structs"
	sandwich "github.com/WelcomerTeam/Sandwich/sandwich"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	images "github.com/WelcomerTeam/Welcomer/welcomer-images/service"
	jsoniter "github.com/json-iterator/go"
	"github.com/savsgio/gotils/strconv"
)

const (
	DefaultFont        = "fredokaone-regular"
	DefaultBorderWidth = 16
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
	bot.RegisterEventHandler(welcomer.CustomEventInvokeWelcomer, func(eventCtx *sandwich.EventContext, payload structs.SandwichPayload) error {
		var invokeWelcomerPayload welcomer.CustomEventInvokeWelcomerStructure
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

	// Trigger CustomEventInvokeWelcomer when ON_GUILD_MEMBER_ADD is triggered.
	bot.RegisterOnGuildMemberAddEvent(func(eventCtx *sandwich.EventContext, member discord.GuildMember) error {
		return p.OnInvokeWelcomerEvent(eventCtx, welcomer.CustomEventInvokeWelcomerStructure{
			Interaction: nil,
			Member:      &member,
		})
	})

	// Call OnInvokeWelcomerEvent when CustomEventInvokeWelcomer is triggered.
	RegisterOnInvokeWelcomerEvent(bot.Handlers, p.OnInvokeWelcomerEvent)

	return nil
}

// RegisterOnInvokeWelcomerEvent adds a new event handler for the WELCOMER_INVOKE_WELCOMER event.
// It does not override a handler and instead will add another handler.
func RegisterOnInvokeWelcomerEvent(h *sandwich.Handlers, event welcomer.OnInvokeWelcomerFuncType) {
	eventName := welcomer.CustomEventInvokeWelcomer

	h.RegisterEvent(eventName, nil, event)
}

func (p *WelcomerCog) FetchWelcomerImage(options images.GenerateImageOptionsRaw) (io.ReadCloser, string, error) {
	optionsJSON, _ := jsoniter.Marshal(options)

	res, err := p.Client.Post("http://"+os.Getenv("IMAGE_HOST")+"/generate", "application/json", bytes.NewBuffer(optionsJSON))
	if err != nil {
		return nil, "", fmt.Errorf("fetch welcomer image request failed: %w", err)
	}

	return res.Body, res.Header.Get("Content-Type"), nil
}

// OnInvokeWelcomerEvent is called when CustomEventInvokeWelcomer is triggered.
// This can be from when a user joins or a user uses /welcomer test.
func (p *WelcomerCog) OnInvokeWelcomerEvent(eventCtx *sandwich.EventContext, event welcomer.CustomEventInvokeWelcomerStructure) (err error) {
	defer func() {
		// Send follow-up if present.
		if event.Interaction != nil && event.Interaction.Token != "" {
			var message discord.WebhookMessageParams

			if err == nil {
				message = discord.WebhookMessageParams{
					Content: "✔️ Executed successfully!",
				}
			} else {
				message = discord.WebhookMessageParams{
					Content: fmt.Sprintf("❌ Failed to execute: `%s`", err.Error()),
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

	guildSettingsWelcomerText, err := queries.GetWelcomerTextGuildSettings(eventCtx.Context, int64(eventCtx.Guild.ID))
	if err != nil {
		eventCtx.Logger.Error().Err(err).
			Int64("guild_id", int64(eventCtx.Guild.ID)).
			Int64("user_id", int64(event.Member.User.ID)).
			Msg("Failed to fetch welcomer text guild settings")

		return err
	}

	guildSettingsWelcomerImages, err := queries.GetWelcomerImagesGuildSettings(eventCtx.Context, int64(eventCtx.Guild.ID))
	if err != nil {
		eventCtx.Logger.Error().Err(err).
			Int64("guild_id", int64(eventCtx.Guild.ID)).
			Int64("user_id", int64(event.Member.User.ID)).
			Msg("Failed to fetch welcomer image guild settings")

		return err
	}

	guildSettingsWelcomerDMs, err := queries.GetWelcomerDMsGuildSettings(eventCtx.Context, int64(eventCtx.Guild.ID))
	if err != nil {
		eventCtx.Logger.Error().Err(err).
			Int64("guild_id", int64(eventCtx.Guild.ID)).
			Int64("user_id", int64(event.Member.User.ID)).
			Msg("Failed to fetch welcomer dm guild settings")

		return err
	}

	// Quit if nothing is enabled.
	if !guildSettingsWelcomerText.ToggleEnabled && !guildSettingsWelcomerImages.ToggleEnabled && !guildSettingsWelcomerDMs.ToggleEnabled {
		return nil
	}

	// Query state cache for guild.
	guilds, err := eventCtx.Sandwich.SandwichClient.FetchGuild(eventCtx, &pb.FetchGuildRequest{
		GuildIDs: []int64{int64(eventCtx.Guild.ID)},
	})
	if err != nil {
		return err
	}

	var guild *discord.Guild
	guildPb, ok := guilds.Guilds[int64(eventCtx.Guild.ID)]
	if ok {
		guild, err = pb.GRPCToGuild(guildPb)
		if err != nil {
			return err
		}
	}

	functions := welcomer.GatherFunctions()
	variables := welcomer.GatherVariables(eventCtx, *event.Member, *guild)

	var message discord.MessageParams
	var directMessage discord.MessageParams

	var imageReaderCloser io.ReadCloser
	var contentType string
	var fileName string

	if guildSettingsWelcomerImages.ToggleEnabled {
		messageFormat, err := welcomer.FormatString(functions, variables, guildSettingsWelcomerImages.ImageMessage)
		if err != nil {
			eventCtx.Logger.Error().Err(err).
				Int64("guild_id", int64(eventCtx.Guild.ID)).
				Int64("user_id", int64(event.Member.User.ID)).
				Msg("Failed to format welcomer text payload")

			return err
		}

		memberships, err := queries.GetValidUserMembershipsByGuildID(eventCtx.Context, eventCtx.Guild.ID, time.Now())
		if err != nil {
			eventCtx.Logger.Warn().Err(err).
				Int64("guild_id", int64(eventCtx.Guild.ID)).
				Msg("Failed to get welcomer memberships")
		}

		hasWelcomerPro, _ := welcomer.CheckGuildMemberships(memberships)

		var borderWidth int32

		if guildSettingsWelcomerImages.ToggleImageBorder {
			borderWidth = DefaultBorderWidth
		} else {
			borderWidth = 0
		}

		imageReaderCloser, contentType, err = p.FetchWelcomerImage(images.GenerateImageOptionsRaw{
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
			ProfileFloat:       int32(welcomer.ImageAlignmentCenter),
			ProfileBorderColor: tryParseColourAsInt64(guildSettingsWelcomerImages.ColourProfileBorder, white),
			ProfileBorderWidth: DefaultBorderWidth,
			ProfileBorderCurve: guildSettingsWelcomerImages.ImageProfileBorderType,
		})
		if err != nil {
			eventCtx.Logger.Warn().Err(err).
				Int64("guild_id", int64(eventCtx.Guild.ID)).
				Int64("user_id", int64(event.Member.User.ID)).
				Msg("Failed to fetch welcomer image")
		}

		defer imageReaderCloser.Close()

		fileName = "image.png"
	}

	if guildSettingsWelcomerText.ToggleEnabled || guildSettingsWelcomerImages.ToggleEnabled {
		if guildSettingsWelcomerText.Channel == 0 {
			return nil
			// TODO: Return an error for no channel set
		}

		if guildSettingsWelcomerText.ToggleEnabled {
			messageFormat, err := welcomer.FormatString(functions, variables, strconv.B2S(guildSettingsWelcomerText.MessageFormat.Bytes))
			if err != nil {
				eventCtx.Logger.Error().Err(err).
					Int64("guild_id", int64(eventCtx.Guild.ID)).
					Int64("user_id", int64(event.Member.User.ID)).
					Msg("Failed to format welcomer text payload")

				return err
			}

			// Convert MessageFormat to MessageParams so we can send it.
			err = jsoniter.UnmarshalFromString(messageFormat, &message)
			if err != nil {
				eventCtx.Logger.Error().Err(err).
					Int64("guild_id", int64(eventCtx.Guild.ID)).
					Int64("user_id", int64(event.Member.User.ID)).
					Msg("Failed to unmarshal messageFormat")

				return err
			}
		}

		if guildSettingsWelcomerImages.ToggleEnabled {
			// TODO: This is not working, possibly due to the ReaderCloser already being read.
			message.AddFile(discord.File{
				Name:        fileName,
				ContentType: contentType,
				Reader:      imageReaderCloser,
			})

			if len(message.Embeds) == 0 {
				message.AddEmbed(discord.Embed{})
			}

			message.Embeds[0].SetImage(discord.NewEmbedImage("attachment://" + fileName))
		}
	}

	if guildSettingsWelcomerDMs.ToggleEnabled {
		if guildSettingsWelcomerText.ToggleEnabled && guildSettingsWelcomerDMs.ToggleUseTextFormat {
			directMessage = message
		} else {
			messageFormat, err := welcomer.FormatString(functions, variables, strconv.B2S(guildSettingsWelcomerDMs.MessageFormat.Bytes))
			if err != nil {
				eventCtx.Logger.Error().Err(err).
					Int64("guild_id", int64(eventCtx.Guild.ID)).
					Int64("user_id", int64(event.Member.User.ID)).
					Msg("Failed to format welcomer DMs payload")

				return err
			}

			// Convert MessageFormat to MessageParams so we can send it.
			err = jsoniter.UnmarshalFromString(messageFormat, &directMessage)
			if err != nil {
				eventCtx.Logger.Error().Err(err).
					Int64("guild_id", int64(eventCtx.Guild.ID)).
					Int64("user_id", int64(event.Member.User.ID)).
					Msg("Failed to unmarshal messageFormat")

				return err
			}

			if guildSettingsWelcomerImages.ToggleEnabled {
				directMessage.AddFile(discord.File{
					Name:        fileName,
					ContentType: contentType,
					Reader:      imageReaderCloser,
				})

				if len(directMessage.Embeds) == 0 {
					directMessage.AddEmbed(discord.Embed{})
				}

				directMessage.Embeds[0].SetImage(discord.NewEmbedImage("attachment://" + fileName))
			}
		}
	}

	channel := discord.Channel{ID: discord.Snowflake(guildSettingsWelcomerText.Channel)}

	_, err = channel.Send(eventCtx.Session, message)
	if err != nil {
		eventCtx.Logger.Warn().Err(err).
			Int64("guild_id", int64(eventCtx.Guild.ID)).
			Int64("channel_id", guildSettingsWelcomerText.Channel).
			Msg("Failed to send message to channel")
	}

	_, err = event.Member.Send(eventCtx.Session, directMessage)
	if err != nil {
		eventCtx.Logger.Warn().Err(err).
			Int64("guild_id", int64(eventCtx.Guild.ID)).
			Int64("user_id", int64(event.Member.User.ID)).
			Msg("Failed to send message to user")
	}

	return nil
}

func tryParseColourAsInt64(str string, defaultValue *color.RGBA) int64 {
	c, err := welcomer.ParseColour(str, "")
	if err != nil {
		c = defaultValue
	}

	return (int64(c.A) << 24) + (int64(c.R) << 16) + (int64(c.G) << 8) + int64(c.B)
}
