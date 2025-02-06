package welcomer

import (
	"github.com/WelcomerTeam/Discord/discord"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestFormatString(t *testing.T) {
	funcs := GatherFunctions()
	vars := GatherVariables(nil, discord.GuildMember{
		JoinedAt: time.Time{},
		User: &discord.User{
			ID:            1234567890,
			Username:      "john.doe",
			Discriminator: "1234",
			GlobalName:    "John Doe",
			Bot:           false,
			Avatar:        "1234567890",
		},
		Pending: false,
	}, discord.Guild{
		ID:          1234567890,
		Name:        "Test Server",
		Icon:        "1234567890",
		Splash:      "",
		MemberCount: 100,
		Banner:      "",
	}, nil, nil)

	testCases := map[string]string{
		"{{User.CreatedAt}}":     "<t:1420070400:R>",
		"{{User.JoinedAt}}":      "<t:-62135596800:R>",
		"{{User.Name}}":          "John Doe",
		"{{User.Username}}":      "john.doe",
		"{{User.Discriminator}}": "1234",
		"{{User.GlobalName}}":    "John Doe",
		"{{User.Mention}}":       "<@1234567890>",
		"{{User.Avatar}}":        "https://cdn.discordapp.com/avatars/1234567890/1234567890.png",
		"{{User.ID}}":            "1234567890",
		"{{User.Bot}}":           "false",
		"{{User.Pending}}":       "false",

		"{{Guild.Name}}":    "Test Server",
		"{{Guild.Icon}}":    "https://cdn.discordapp.com/icons/1234567890/1234567890.png",
		"{{Guild.Splash}}":  "",
		"{{Guild.Banner}}":  "",
		"{{Guild.ID}}":      "1234567890",
		"{{Guild.Members}}": "100",

		"{{Ordinal(Guild.Members)}}": "100th",

		"":                        "",
		"Hello, world!":           "Hello, world!",
		"Welcome, {{User.Name}}!": "Welcome, John Doe!",

		"{{#User.Bot}}Bot{{/User.Bot}}{{^User.Bot}}Not Bot{{/User.Bot}}": "Not Bot",
	}

	for testCaseMessage, testCaseExpected := range testCases {
		result, err := FormatString(funcs, vars, testCaseMessage)
		assert.NoError(t, err)
		assert.Equal(t, testCaseExpected, result)
	}
}
