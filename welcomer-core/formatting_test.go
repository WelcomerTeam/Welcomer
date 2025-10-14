package welcomer

import (
	"testing"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/stretchr/testify/assert"
)

func TestFormatString(t *testing.T) {
	funcs := GatherFunctions(database.NumberLocaleDefault)
	vars := GatherVariables(nil, &discord.GuildMember{
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
	}, GuildVariables{
		Guild: &discord.Guild{
			ID:          1234567890,
			Name:        "Test Server",
			Icon:        "1234567890",
			Splash:      "",
			MemberCount: 1234,
			Banner:      "",
		},
		MembersJoined: 123456,
		NumberLocale:  database.NumberLocaleDefault,
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

		"{{Guild.Name}}":          "Test Server",
		"{{Guild.Icon}}":          "https://cdn.discordapp.com/icons/1234567890/1234567890.png",
		"{{Guild.Splash}}":        "",
		"{{Guild.Banner}}":        "",
		"{{Guild.ID}}":            "1234567890",
		"{{Guild.Members}}":       "1234",
		"{{Guild.MembersJoined}}": "123456",

		"{{Ordinal(Guild.Members)}}":       "1234th",
		"{{Ordinal(Guild.MembersJoined)}}": "123456th",

		"{{FormatNumber(Guild.Members)}}":       "1234",
		"{{FormatNumber(Guild.MembersJoined)}}": "123456",

		"{{FormatNumber(Guild.MembersJoined, \"default\")}}": "123456",
		"{{FormatNumber(Guild.MembersJoined, \"dots\")}}":    "123.456",
		"{{FormatNumber(Guild.MembersJoined, \"commas\")}}":  "123,456",
		"{{FormatNumber(Guild.MembersJoined, \"indian\")}}":  "1,23,456",
		"{{FormatNumber(Guild.MembersJoined, \"arabic\")}}":  "١٢٣٬٤٥٦",

		"{{Upper(User.Username)}}": "JOHN.DOE",
		"{{Lower(User.Username)}}": "john.doe",
		"{{Title(User.Username)}}": "John.Doe",

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

	funcs = GatherFunctions(database.NumberLocaleArabic)

	testCases = map[string]string{
		"{{Ordinal(Guild.Members)}}":            "١٬٢٣٤th",
		"{{Ordinal(Guild.MembersJoined)}}":      "١٢٣٬٤٥٦th",
		"{{FormatNumber(Guild.Members)}}":       "١٬٢٣٤",
		"{{FormatNumber(Guild.MembersJoined)}}": "١٢٣٬٤٥٦",
	}

	for testCaseMessage, testCaseExpected := range testCases {
		result, err := FormatString(funcs, vars, testCaseMessage)
		assert.NoError(t, err)
		assert.Equal(t, testCaseExpected, result)
	}

	funcs = GatherFunctions(database.NumberLocaleCommas)

	testCases = map[string]string{
		"{{Ordinal(Guild.Members)}}":            "1,234th",
		"{{Ordinal(Guild.MembersJoined)}}":      "123,456th",
		"{{FormatNumber(Guild.Members)}}":       "1,234",
		"{{FormatNumber(Guild.MembersJoined)}}": "123,456",
	}

	for testCaseMessage, testCaseExpected := range testCases {
		result, err := FormatString(funcs, vars, testCaseMessage)
		assert.NoError(t, err)
		assert.Equal(t, testCaseExpected, result)
	}

	funcs = GatherFunctions(database.NumberLocaleDots)

	testCases = map[string]string{
		"{{Ordinal(Guild.Members)}}":            "1.234th",
		"{{Ordinal(Guild.MembersJoined)}}":      "123.456th",
		"{{FormatNumber(Guild.Members)}}":       "1.234",
		"{{FormatNumber(Guild.MembersJoined)}}": "123.456",
	}

	for testCaseMessage, testCaseExpected := range testCases {
		result, err := FormatString(funcs, vars, testCaseMessage)
		assert.NoError(t, err)
		assert.Equal(t, testCaseExpected, result)
	}

	funcs = GatherFunctions(database.NumberLocaleIndian)

	testCases = map[string]string{
		"{{Ordinal(Guild.Members)}}":            "1,234th",
		"{{FormatNumber(Guild.Members)}}":       "1,234",
		"{{Ordinal(Guild.MembersJoined)}}":      "1,23,456th",
		"{{FormatNumber(Guild.MembersJoined)}}": "1,23,456",
	}

	for testCaseMessage, testCaseExpected := range testCases {
		result, err := FormatString(funcs, vars, testCaseMessage)
		assert.NoError(t, err)
		assert.Equal(t, testCaseExpected, result)
	}
}
