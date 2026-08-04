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
		JoinedAt: time.Unix(1609459200, 0),
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
	}, nil, nil, true)

	testCases := map[string]string{
		"{{User.CreatedAt}}":     "<t:1420070400:R>",
		"{{User.JoinedAt}}":      "<t:1609459200:R>",
		"{{User.Name}}":          "John Doe",
		"{{User.Username}}":      "john.doe",
		"{{User.Discriminator}}": "1234",
		"{{User.GlobalName}}":    "John Doe",
		"{{User.Mention}}":       "<@1234567890>",
		"{{User.Avatar}}":        "https://cdn.discordapp.com/avatars/1234567890/1234567890.png?size=256",
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

		"{{FormatTime(User.JoinedAt, \"dd/MM/yyyy\")}}":           "01/01/2021",
		"{{FormatTime(User.JoinedAt, \"MM-dd-yyyy\")}}":           "01-01-2021",
		"{{FormatTime(User.JoinedAt, \"yyyy/MM/dd\")}}":           "2021/01/01",
		"{{FormatTime(User.JoinedAt, \"HH:mm:ss\")}}":             "00:00:00",
		"{{FormatTime(User.JoinedAt, \"yyyy-MM-ddTHH:mm:ssZ\")}}": "2021-01-01T00:00:00Z",
		"{{FormatTime(User.JoinedAt, \"MMMM dd, yyyy\")}}":        "January 01, 2021",
		"{{FormatTime(User.JoinedAt, \"MMM dd, yyyy\")}}":         "Jan 01, 2021",
		"{{FormatTime(User.JoinedAt, \"dd MMM yyyy\")}}":          "01 Jan 2021",
		"{{FormatTime(User.JoinedAt, \"yyyyMMdd\")}}":             "20210101",
		"{{FormatTime(User.JoinedAt, \"yyyy-MM-dd\")}}":           "2021-01-01",
		"{{FormatTime(User.JoinedAt, \"MM/dd/yyyy\")}}":           "01/01/2021",
		"{{FormatTime(User.JoinedAt, \"dd/MM/yyyy HH:mm:ss\")}}":  "01/01/2021 00:00:00",

		"{{SinceTime(User.JoinedAt)}}":  "5 years",
		"{{SinceTime(User.CreatedAt)}}": "11 years",

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

func TestFormatStubTime(t *testing.T) {
	testCases := map[string]string{
		"dd/MM/yyyy":           "01/01/2021",
		"MM-dd-yyyy":           "01-01-2021",
		"yyyy/MM/dd":           "2021/01/01",
		"HH:mm:ss":             "00:00:00",
		"yyyy-MM-ddTHH:mm:ssZ": "2021-01-01T00:00:00Z",
		"MMMM dd, yyyy":        "January 01, 2021",
		"MMM dd, yyyy":         "Jan 01, 2021",
		"dd MMM yyyy":          "01 Jan 2021",
		"yyyyMMdd":             "20210101",
		"yyyy-MM-dd":           "2021-01-01",
		"MM/dd/yyyy":           "01/01/2021",
		"dd/MM/yyyy HH:mm:ss":  "01/01/2021 00:00:00",

		"yyyy": "2021",
		"yy":   "21",
		"MMMM": "January",
		"MMM":  "Jan",
		"MM":   "01",
		"M":    "1",
		"dddd": "Friday",
		"ddd":  "Fri",
		"dd":   "01",
		"d":    "1",
		"HH":   "00",
		"h":    "12",
		"hh":   "12",
		"m":    "0",
		"mm":   "00",
		"s":    "0",
		"ss":   "00",
	}

	unixTime := time.Unix(1609459200, 0)

	for format, expected := range testCases {
		result := FormatTime(unixTime, format)
		assert.Equal(t, expected, result)
	}
}

func TestSinceTime(t *testing.T) {
	testCases := map[int]string{
		0:        "now",
		1:        "1 second",
		2:        "2 seconds",
		59:       "59 seconds",
		60:       "1 minute",
		61:       "1 minute",
		120:      "2 minutes",
		3599:     "59 minutes",
		3600:     "1 hour",
		3601:     "1 hour",
		3660:     "1 hour",
		3661:     "1 hour",
		7200:     "2 hours",
		7261:     "2 hours",
		86399:    "23 hours",
		86400:    "1 day",
		86401:    "1 day",
		90061:    "1 day",
		172800:   "2 days",
		172801:   "2 days",
		2592000:  "30 days",
		2592001:  "30 days",
		31536000: "1 year",
		31536001: "1 year",
	}

	for seconds, expected := range testCases {
		result := SinceTime(time.Now().Add(-time.Duration(seconds) * time.Second))
		assert.Equal(t, expected, result)
	}
}
