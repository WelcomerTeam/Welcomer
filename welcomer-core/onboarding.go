package welcomer

import (
	"os"

	"github.com/WelcomerTeam/Discord/discord"
)

func GetOnboardingMessage(guildID discord.Snowflake) discord.MessageParams {
	return discord.MessageParams{
		Flags: discord.MessageFlagIsComponentsV2,
		Components: []discord.InteractionComponent{
			{
				Type:        discord.InteractionComponentTypeContainer,
				AccentColor: ToPointer(uint32(0xfbc01b)),
				Components: []discord.InteractionComponent{
					{
						Type: discord.InteractionComponentTypeSection,
						Components: []discord.InteractionComponent{
							{
								Type: discord.InteractionComponentTypeTextDisplay,
								Content: "# Welcome to Welcomer!\n" +
									"Thank you for adding me to your server! I'm here to help you with onboarding users, improving user engagement, and providing a better experience for your community.\n\n" +
									"### Getting started?\n\n" +
									"To get started with using the welcomer module, set a channel to use with `/welcomer setchannel` and then use `/welcomer enable`\n\n" +
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
						Type: discord.InteractionComponentTypeSection,
						Components: []discord.InteractionComponent{
							{
								Type:    discord.InteractionComponentTypeTextDisplay,
								Content: "To get access to more customization options such as with welcome images, check out your server's dashboard.",
							},
						},
						Accessory: &discord.InteractionComponent{
							Type:  discord.InteractionComponentTypeButton,
							Style: discord.InteractionComponentStyleLink,
							Label: "Dashboard",
							URL:   WebsiteURL + "/dashboard/" + guildID.String(),
						},
					},
					{
						Type: discord.InteractionComponentTypeSeparator,
					},
					{
						Type: discord.InteractionComponentTypeSection,
						Components: []discord.InteractionComponent{
							{
								Type:    discord.InteractionComponentTypeTextDisplay,
								Content: "For help with using the bot, check out our support server.",
							},
						},
						Accessory: &discord.InteractionComponent{
							Type:  discord.InteractionComponentTypeButton,
							Style: discord.InteractionComponentStyleLink,
							Label: "Support Server",
							URL:   WebsiteURL + "/support",
						},
					},
					{
						Type: discord.InteractionComponentTypeSeparator,
					},
					{
						Type: discord.InteractionComponentTypeSection,
						Components: []discord.InteractionComponent{
							{
								Type:    discord.InteractionComponentTypeTextDisplay,
								Content: "### Want to get access to more features?\nCheck out Welcomer Pro! If you just want custom backgrounds, you can also get a one-time purchase which lasts forever.",
							},
						},
						Accessory: &discord.InteractionComponent{
							Type:  discord.InteractionComponentTypeButton,
							Style: discord.InteractionComponentStyleLink,
							Label: "Get Welcomer Pro",
							URL:   WebsiteURL + "/premium",
						},
					},
				},
			},
			{
				Type: discord.InteractionComponentTypeActionRow,
				Components: []discord.InteractionComponent{
					{
						Type:  discord.InteractionComponentTypeButton,
						Style: discord.InteractionComponentStyleLink,
						Label: WebsiteURL,
						URL:   WebsiteURL,
					},
					{
						Type:  discord.InteractionComponentTypeButton,
						Style: discord.InteractionComponentStyleLink,
						Label: "Vote for Welcomer",
						URL:   "https://top.gg/bot/330416853971107840/vote",
					},
					{
						Type:  discord.InteractionComponentTypeButton,
						Style: discord.InteractionComponentStylePremium,
						SKUID: discord.Snowflake(TryParseInt(os.Getenv("WELCOMER_PRO_DISCORD_SKU_ID"))),
					},
				},
			},
		},
	}
}
