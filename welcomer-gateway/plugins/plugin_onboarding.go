package plugins

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	sandwich "github.com/WelcomerTeam/Sandwich/sandwich"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
)

var (
	LargeGuildsWebhookURL = os.Getenv("DISCORD_LARGE_GUILDS_WEBHOOK_URL")
	GuildsWebhookURL      = os.Getenv("DISCORD_GUILDS_WEBHOOK_URL")
)

type OnboardingCog struct {
	EventHandler *sandwich.Handlers
}

// Assert types.

var (
	_ sandwich.Cog           = (*OnboardingCog)(nil)
	_ sandwich.CogWithEvents = (*OnboardingCog)(nil)
)

func NewOnboardingCog() *OnboardingCog {
	return &OnboardingCog{
		EventHandler: sandwich.SetupHandler(nil),
	}
}

func (p *OnboardingCog) CogInfo() *sandwich.CogInfo {
	return &sandwich.CogInfo{
		Name:        "Onboarding",
		Description: "Provides the functionality for the 'Onboarding' feature",
	}
}

func (p *OnboardingCog) GetEventHandlers() *sandwich.Handlers {
	return p.EventHandler
}

func (p *OnboardingCog) RegisterCog(bot *sandwich.Bot) error {
	// Register
	p.EventHandler.RegisterOnGuildJoinEvent(func(eventCtx *sandwich.EventContext, guild discord.Guild) error {
		welcomer.PushGuildScience.Push(
			eventCtx.Context,
			eventCtx.Guild.ID,
			0,
			database.ScienceGuildEventTypeGuildJoin,
			nil,
		)

		if guild.MemberCount > 1000 && LargeGuildsWebhookURL != "" {
			err := welcomer.SendWebhookMessage(eventCtx.Context, LargeGuildsWebhookURL, discord.WebhookMessageParams{
				Embeds: []discord.Embed{
					{
						Title:     "New Large Guild",
						Color:     welcomer.EmbedColourSuccess,
						Timestamp: welcomer.ToPointer(time.Now()),
						Fields: []discord.EmbedField{
							{
								Name:   "Name",
								Value:  guild.Name,
								Inline: true,
							},
							{
								Name:   "ID",
								Value:  guild.ID.String(),
								Inline: true,
							},
							{
								Name:   "Members",
								Value:  strconv.Itoa(int(guild.MemberCount)),
								Inline: true,
							},
							{
								Name:   "Owner",
								Value:  fmt.Sprintf("<@%s> %s", guild.OwnerID.String(), guild.OwnerID.String()),
								Inline: true,
							},
						},
					},
				},
			})
			if err != nil {
				welcomer.Logger.Error().Err(err).Msg("Failed to send webhook message")
			}
		} else if GuildsWebhookURL != "" {
			err := welcomer.SendWebhookMessage(eventCtx.Context, GuildsWebhookURL, discord.WebhookMessageParams{
				Embeds: []discord.Embed{
					{
						Title:     "New Guild",
						Color:     welcomer.EmbedColourSuccess,
						Timestamp: welcomer.ToPointer(time.Now()),
						Fields: []discord.EmbedField{
							{
								Name:   "Name",
								Value:  guild.Name,
								Inline: true,
							},
							{
								Name:   "ID",
								Value:  guild.ID.String(),
								Inline: true,
							},
							{
								Name:   "Members",
								Value:  strconv.Itoa(int(guild.MemberCount)),
								Inline: true,
							},
							{
								Name: "Owner",
								Value: welcomer.IfFunc(
									guild.OwnerID != nil,
									func() string {
										return fmt.Sprintf("<@%s> %s", guild.OwnerID.String(), guild.OwnerID.String())
									},
									func() string { return "" },
								),
								Inline: true,
							},
						},
					},
				},
			})
			if err != nil {
				welcomer.Logger.Error().Err(err).Msg("Failed to send webhook message")
			}
		}

		return nil
	})

	p.EventHandler.RegisterOnGuildLeaveEvent(func(eventCtx *sandwich.EventContext, guild discord.Guild) error {
		welcomer.PushGuildScience.Push(
			eventCtx.Context,
			eventCtx.Guild.ID,
			0,
			database.ScienceGuildEventTypeGuildLeave,
			int(time.Since(guild.JoinedAt).Seconds()),
		)

		if guild.MemberCount > 1000 && LargeGuildsWebhookURL != "" {
			err := welcomer.SendWebhookMessage(eventCtx.Context, LargeGuildsWebhookURL, discord.WebhookMessageParams{
				Embeds: []discord.Embed{
					{
						Title:     "Left Large Guild",
						Color:     welcomer.EmbedColourError,
						Timestamp: welcomer.ToPointer(time.Now()),
						Fields: []discord.EmbedField{
							{
								Name:   "Name",
								Value:  guild.Name,
								Inline: true,
							},
							{
								Name:   "ID",
								Value:  guild.ID.String(),
								Inline: true,
							},
							{
								Name:   "Members",
								Value:  strconv.Itoa(int(guild.MemberCount)),
								Inline: true,
							},
							{
								Name:   "Owner",
								Value:  fmt.Sprintf("<@%s> %s", guild.OwnerID.String(), guild.OwnerID.String()),
								Inline: true,
							},
							{
								Name:   "Retention",
								Value:  welcomer.HumanizeDuration(int(time.Since(guild.JoinedAt).Seconds()), true),
								Inline: true,
							},
						},
					},
				},
			})
			if err != nil {
				welcomer.Logger.Error().Err(err).Msg("Failed to send webhook message")
			}
		} else if GuildsWebhookURL != "" {
			err := welcomer.SendWebhookMessage(eventCtx.Context, GuildsWebhookURL, discord.WebhookMessageParams{
				Embeds: []discord.Embed{
					{
						Title:     "Left Guild",
						Color:     welcomer.EmbedColourError,
						Timestamp: welcomer.ToPointer(time.Now()),
						Fields: []discord.EmbedField{
							{
								Name:   "Name",
								Value:  guild.Name,
								Inline: true,
							},
							{
								Name:   "ID",
								Value:  guild.ID.String(),
								Inline: true,
							},
							{
								Name:   "Members",
								Value:  strconv.Itoa(int(guild.MemberCount)),
								Inline: true,
							},
							{
								Name:   "Owner",
								Value:  fmt.Sprintf("<@%s> %s", guild.OwnerID.String(), guild.OwnerID.String()),
								Inline: true,
							},
							{
								Name:   "Retention",
								Value:  welcomer.HumanizeDuration(int(time.Since(guild.JoinedAt).Seconds()), true),
								Inline: true,
							},
						},
					},
				},
			})
			if err != nil {
				welcomer.Logger.Error().Err(err).Msg("Failed to send webhook message")
			}
		}

		return nil
	})

	p.EventHandler.RegisterOnAuditGuildAuditLogEntryCreateEvent(func(eventCtx *sandwich.EventContext, guildID discord.Snowflake, entry discord.AuditLogEntry) error {
		if entry.ActionType != discord.AuditLogActionBotAdd || *entry.TargetID != eventCtx.Identifier.ID {
			return nil
		}

		guild, err := welcomer.FetchGuild(eventCtx.Context, eventCtx.Guild.ID)
		if err != nil {
			welcomer.Logger.Error().Err(err).
				Int64("guild_id", int64(eventCtx.Guild.ID)).
				Msg("Failed to fetch guild from state cache")

			return err
		}

		user, err := welcomer.FetchUser(eventCtx.Context, *entry.UserID, true)
		if err != nil {
			welcomer.Logger.Error().Err(err).
				Int64("guild_id", int64(guildID)).
				Int64("user_id", int64(user.ID)).
				Msg("Failed to fetch user from state cache")

			return err
		}

		_, err = user.Send(eventCtx.Context, eventCtx.Session, welcomer.GetOnboardingMessage(guild.ID))
		if err != nil {
			welcomer.Logger.Error().Err(err).
				Int64("guild_id", int64(guildID)).
				Int64("user_id", int64(user.ID)).
				Msg("Failed to send DM to user")
		}

		welcomer.PushGuildScience.Push(
			eventCtx.Context,
			guild.ID,
			user.ID,
			database.ScienceGuildEventTypeGuildUserOnboarded,
			err == nil,
		)

		return err
	})

	p.EventHandler.RegisterOnGuildJoinEvent(func(eventCtx *sandwich.EventContext, guild discord.Guild) error {
		guild, err := welcomer.FetchGuild(eventCtx.Context, eventCtx.Guild.ID)
		if err != nil {
			welcomer.Logger.Error().Err(err).
				Int64("guild_id", int64(eventCtx.Guild.ID)).
				Msg("Failed to fetch guild from state cache")

			return err
		}

		var eligibleChannel *discord.Channel

		if guild.SystemChannelID != nil {
			eligibleChannel = &discord.Channel{ID: *guild.SystemChannelID}
		} else {
			channels, err := welcomer.FetchGuildChannels(eventCtx.Context, eventCtx.Guild.ID)
			if err != nil {
				welcomer.Logger.Error().Err(err).
					Int64("guild_id", int64(eventCtx.Guild.ID)).
					Msg("Failed to fetch channels from state cache")

				return nil
			}

			for _, channel := range channels {
				if channel.Type == discord.ChannelTypeGuildText &&
					welcomer.CompareStrings(channel.Name, "welcome", "general") {
					eligibleChannel = &channel

					break
				}
			}
		}

		if eligibleChannel == nil {
			welcomer.PushGuildScience.Push(
				eventCtx.Context,
				guild.ID,
				0,
				database.ScienceGuildEventTypeGuildOnboarded,
				false,
			)

			return nil
		}

		_, err = eligibleChannel.Send(eventCtx.Context, eventCtx.Session, welcomer.GetOnboardingMessage(guild.ID))
		if err != nil {
			welcomer.Logger.Error().Err(err).
				Int64("guild_id", int64(eventCtx.Guild.ID)).
				Int64("channel_id", int64(eligibleChannel.ID)).
				Msg("Failed to send onboarding message")

			return nil
		}

		welcomer.PushGuildScience.Push(
			eventCtx.Context,
			guild.ID,
			0,
			database.ScienceGuildEventTypeGuildOnboarded,
			err == nil,
		)

		return err
	})

	return nil
}
