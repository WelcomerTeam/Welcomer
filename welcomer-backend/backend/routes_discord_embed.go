package backend

import (
	"net/http"

	"github.com/WelcomerTeam/Discord/discord"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/gin-gonic/gin"
)

func discordEmbedHandler(ctx *gin.Context) {
	// path := ctx.Request.URL.Query().Get("path")

	// // if path != "/" {
	// // 	return
	// // }

	var component discord.InteractionComponent

	component = discord.InteractionComponent{
		Type:        discord.InteractionComponentTypeContainer,
		AccentColor: new(uint32(0xfbc01b)),
		Components: []discord.InteractionComponent{
			{
				Type: discord.InteractionComponentTypeSection,
				Components: []discord.InteractionComponent{
					{
						Type: discord.InteractionComponentTypeTextDisplay,
						Content: "# Hello, I'm Welcomer!\n" +
							"I'm here to help you with onboarding users, improving user engagement, and providing a better experience for your community.\n\n" +
							"I don't just welcome users though! I can do more for your server, check out our other features [here](https://welcomer.gg/#features).",
					},
				},
				Accessory: &discord.InteractionComponent{
					Type: discord.InteractionComponentTypeThumbnail,
					Media: &discord.MediaItem{
						URL: "https://welcomer.gg/assets/wave.gif",
					},
				},
			},
			{
				Type: discord.InteractionComponentTypeSeparator,
			},
			{
				Type: discord.InteractionComponentTypeActionRow,
				Components: []discord.InteractionComponent{
					{
						Type:  discord.InteractionComponentTypeButton,
						Style: discord.InteractionComponentStyleLink,
						Label: "Vote for Welcomer",
						URL:   "https://top.gg/bot/330416853971107840/vote",
					},
					{
						Type:  discord.InteractionComponentTypeButton,
						Style: discord.InteractionComponentStyleLink,
						Label: "Get Welcomer Pro",
						URL:   welcomer.WebsiteURL + "/premium",
					},
					{
						Type:  discord.InteractionComponentTypeButton,
						Style: discord.InteractionComponentStyleLink,
						Label: "Support Server",
						URL:   welcomer.WebsiteURL + "/support",
					},
				},
			},
		},
	}

	ctx.JSON(http.StatusOK, gin.H{
		"component": component,
	})
}

func registerDiscordEmbedRoute(g *gin.Engine) {
	g.GET("/api/discord-embed.json", discordEmbedHandler)
}
