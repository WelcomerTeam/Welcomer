package backend

import (
	"context"
	"errors"
	"time"

	discord "github.com/WelcomerTeam/Discord/discord"
	sandwich "github.com/WelcomerTeam/Sandwich/sandwich"
)

// Embed colours for webhooks.
const (
	EmbedColourSandwich = 16701571
	EmbedColourWarning  = 16760839
	EmbedColourDanger   = 14431557

	WebhookRateLimitDuration = 5 * time.Second
	WebhookRateLimitLimit    = 5
)

// PublishSimpleWebhook is a helper function for creating quicker webhook messages.
func (b *Backend) PublishSimpleWebhook(s *discord.Session, title string, description string, footer string, colour int32) {
	now := time.Now().UTC()

	b.PublishWebhook(s, discord.WebhookMessageParams{
		Embeds: []*discord.Embed{
			{
				Title:       title,
				Description: description,
				Color:       colour,
				Timestamp:   &now,
				Footer: &discord.EmbedFooter{
					Text: footer,
				},
			},
		},
	})
}

// PublishWebhook sends a webhook message to all added webhooks in the configuration.
func (b *Backend) PublishWebhook(session *discord.Session, message discord.WebhookMessageParams) {
	for _, webhookURL := range b.Configuration.Webhooks {
		webhook, err := sandwich.WebhookFromURL(webhookURL)
		if err != nil {
			b.Logger.Warn().Err(err).Str("url", webhookURL).Msg("Failed to parse webhook from URL")

			continue
		}

		_, err = webhook.Send(session, message, false)
		if err != nil && !errors.Is(err, context.Canceled) {
			b.Logger.Warn().Err(err).Str("url", webhookURL).Msg("Failed to send webhook")
		}
	}
}
