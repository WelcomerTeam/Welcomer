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
		welcomer.GetPushGuildScienceFromContext(eventCtx.Context).Push(
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
		}

		return nil
	})

	p.EventHandler.RegisterOnGuildLeaveEvent(func(eventCtx *sandwich.EventContext, guild discord.Guild) error {
		welcomer.GetPushGuildScienceFromContext(eventCtx.Context).Push(
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

	return nil
}
