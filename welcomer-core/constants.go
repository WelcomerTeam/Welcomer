package welcomer

import "github.com/WelcomerTeam/Discord/discord"

const (
	EmojiCheck   = "<:check:1196902761627914402>"
	EmojiCross   = "<:cross:1196902764048031744>"
	EmojiNeutral = "<:neutral:1196903241959620730>"

	EmojiRock = "<:rock:732274836038221855>"
)

var (
	EmojiMessageBadge = discord.Emoji{ID: 987044175943970867, Name: "messagebadge"}
	EmojiShieldAlert  = discord.Emoji{ID: 987044177160331322, Name: "shieldalert"}
	EmojiCheckMark    = discord.Emoji{ID: 586907765662941185, Name: "checkboxmarkedoutline"}

	SupportInvite = "https://discord.gg/kQJz33ExK2"
	WebsiteURL    = "https://beta-dev.welcomer.gg"
)
