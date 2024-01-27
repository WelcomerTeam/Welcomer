package welcomer

const (
	EmojiCheck   = "<:check:1196902761627914402>"
	EmojiCross   = "<:cross:1196902764048031744>"
	EmojiNeutral = "<:neutral:1196903241959620730>"
)

var (
	WebsiteURL = "https://beta-dev.welcomer.gg"

	WebsiteGuildURL = func(guildID string) string {
		return WebsiteURL + "/dashboard/" + guildID
	}
)
