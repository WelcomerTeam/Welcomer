package welcomer

import "github.com/WelcomerTeam/Discord/discord"

const (
	EmbedColourInfo    = 0x2F80ED
	EmbedColourSuccess = 0x4CD787
	EmbedColourError   = 0xFC6A70
	EmbedColourWarn    = 0xFBC01B
)

func NewEmbed(message string, color int32) []*discord.Embed {
	embeds := []*discord.Embed{
		{
			Description: message,
			Color:       color,
		},
	}

	return embeds
}
