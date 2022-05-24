package backend

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	discord "github.com/WelcomerTeam/Discord/discord"
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
func (b *Backend) PublishWebhook(s *discord.Session, message discord.WebhookMessageParams) {
	for _, webhook := range b.Configuration.Webhooks {
		_, err := b.SendWebhook(s, webhook, message)
		if err != nil && !errors.Is(err, context.Canceled) {
			b.Logger.Warn().Err(err).Str("url", webhook).Msg("Failed to send webhook")
		}
	}
}

func (b *Backend) SendWebhook(s *discord.Session, webhookURL string, messageParams discord.WebhookMessageParams) (message *discord.WebhookMessage, err error) {
	webhookURL = strings.TrimSpace(webhookURL)

	_, err = url.Parse(webhookURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse webhook URL: %w", err)
	}

	err = s.Interface.FetchJJ(s, http.MethodPost, webhookURL, messageParams, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to execute webhook: %w", err)
	}

	return
}
