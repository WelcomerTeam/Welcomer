package welcomer

import (
	"fmt"

	"github.com/WelcomerTeam/Discord/discord"
)

func includeActionRow(messageParams *discord.MessageParams) *discord.MessageParams {
	if len(messageParams.Components) == 0 {
		messageParams.AddComponent(discord.InteractionComponent{
			Type:       discord.InteractionComponentTypeActionRow,
			Components: make([]*discord.InteractionComponent, 0),
		})
	}

	return messageParams
}

func IncludeSentByButton(messageParams *discord.MessageParams, guildName string) *discord.MessageParams {
	messageParams = includeActionRow(messageParams)

	messageParams.Components[0].Components = append(
		messageParams.Components[0].Components,
		&discord.InteractionComponent{
			Type:     discord.InteractionComponentTypeButton,
			Style:    discord.InteractionComponentStylePrimary,
			Label:    fmt.Sprintf("Sent by %s", guildName),
			CustomID: "server",
			Emoji:    &EmojiMessageBadge,
			Disabled: true,
		},
	)

	return messageParams
}

func IncludeScamsButton(messageParams *discord.MessageParams) *discord.MessageParams {
	messageParams = includeActionRow(messageParams)

	messageParams.Components[0].Components = append(
		messageParams.Components[0].Components,
		&discord.InteractionComponent{
			Type:  discord.InteractionComponentTypeButton,
			Style: discord.InteractionComponentStyleLink,
			Label: "Watch out for scams",
			URL:   "https://beta-dev.welcomer.gg/phishing",
			Emoji: &EmojiShieldAlert,
		},
	)

	return messageParams
}

func IncludeBorderwallVerifyButton(messageParams *discord.MessageParams, borderwallLink string) *discord.MessageParams {
	messageParams = includeActionRow(messageParams)

	messageParams.Components[0].Components = append(
		messageParams.Components[0].Components,
		&discord.InteractionComponent{
			Type:  discord.InteractionComponentTypeButton,
			Style: discord.InteractionComponentStyleLink,
			Label: "Verify",
			URL:   borderwallLink,
			Emoji: &EmojiCheckMark,
		},
	)

	return messageParams
}
