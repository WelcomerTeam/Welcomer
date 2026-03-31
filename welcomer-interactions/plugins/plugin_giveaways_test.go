package plugins

import (
	"testing"

	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/stretchr/testify/assert"
)

func TestParsePrizesFromString(t *testing.T) {
	testCases := map[string]struct {
		input    string
		expected []welcomer.GiveawayPrize
	}{
		"single prize without count": {
			input: "Gaming Laptop",
			expected: []welcomer.GiveawayPrize{
				{Count: 1, Title: "Gaming Laptop"},
			},
		},
		"single prize with count": {
			input: "5 Discord Nitro",
			expected: []welcomer.GiveawayPrize{
				{Count: 5, Title: "Discord Nitro"},
			},
		},
		"multiple prizes mixed": {
			input: "Gaming Laptop\n3 Steam Gift Cards\n10 Discord Nitro",
			expected: []welcomer.GiveawayPrize{
				{Count: 1, Title: "Gaming Laptop"},
				{Count: 3, Title: "Steam Gift Cards"},
				{Count: 10, Title: "Discord Nitro"},
			},
		},
		"prizes with extra whitespace": {
			input: "  5 Gaming Mouse  \n  2 Mechanical Keyboard  ",
			expected: []welcomer.GiveawayPrize{
				{Count: 5, Title: "Gaming Mouse"},
				{Count: 2, Title: "Mechanical Keyboard"},
			},
		},
		"empty lines should be skipped": {
			input: "5 Prize A\n\n\n3 Prize B",
			expected: []welcomer.GiveawayPrize{
				{Count: 5, Title: "Prize A"},
				{Count: 3, Title: "Prize B"},
			},
		},
		"empty string": {
			input:    "",
			expected: []welcomer.GiveawayPrize{},
		},
		"only whitespace": {
			input:    "   \n  \n  ",
			expected: []welcomer.GiveawayPrize{},
		},
		"invalid count falls back to prize name": {
			input: "abc Gaming Laptop",
			expected: []welcomer.GiveawayPrize{
				{Count: 1, Title: "abc Gaming Laptop"},
			},
		},
		"zero count falls back to prize name": {
			input: "0 Gaming Laptop",
			expected: []welcomer.GiveawayPrize{
				{Count: 1, Title: "0 Gaming Laptop"},
			},
		},
		"negative count falls back to prize name": {
			input: "-5 Gaming Laptop",
			expected: []welcomer.GiveawayPrize{
				{Count: 1, Title: "-5 Gaming Laptop"},
			},
		},
		"large count": {
			input: "999 Mega Prize",
			expected: []welcomer.GiveawayPrize{
				{Count: 999, Title: "Mega Prize"},
			},
		},
		"prize with multiple words": {
			input: "2 Limited Edition Gaming PC",
			expected: []welcomer.GiveawayPrize{
				{Count: 2, Title: "Limited Edition Gaming PC"},
			},
		},
		"prize with x in count": {
			input: "2x Limited Edition Gaming PC",
			expected: []welcomer.GiveawayPrize{
				{Count: 2, Title: "Limited Edition Gaming PC"},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			result := parsePrizesFromString(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
