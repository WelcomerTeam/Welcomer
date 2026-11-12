package backend

import "testing"

func TestIsValidEmoji(t *testing.T) {
	tests := []struct {
		name  string
		emoji string
		want  bool
	}{
		{name: "purple circle", emoji: "🟣", want: true},
		{name: "symbol emoji with variation selector", emoji: "©️", want: true},
		{name: "keycap sequence", emoji: "1️⃣", want: true},
		{name: "regional indicator flag", emoji: "🇺🇸", want: true},
		{name: "tag sequence flag", emoji: "\U0001F3F4\U000E0067\U000E0062\U000E0065\U000E006E\U000E0067\U000E007F", want: true},
		{name: "emoji added in Unicode 17", emoji: "\U0001FACD", want: true},
		{name: "plain text", emoji: "a", want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := isValidEmoji(test.emoji); got != test.want {
				t.Errorf("isValidEmoji(%q) = %t, want %t", test.emoji, got, test.want)
			}
		})
	}
}
