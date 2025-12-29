package service_test

import (
	"testing"

	"github.com/WelcomerTeam/Welcomer/welcomer-images-next/service"
)

func TestMarkdownWithDiscordEmoji(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		input          string
		expectedOutput string
	}{
		{
			name:           "Single static emoji",
			input:          "Hello <:smile:123456789012345678> world!",
			expectedOutput: `Hello <img class="emoji" style="object-fit: contain; width: 1.375em; height: 1.375em; vertical-align: bottom; display: inline;" src="https://cdn.discordapp.com/emojis/123456789012345678.png"> world!`,
		},
		{
			name:           "Single animated emoji",
			input:          "Hello <a:dance:987654321098765432> world!",
			expectedOutput: `Hello <img class="emoji emoji-animated" style="object-fit: contain; width: 1.375em; height: 1.375em; vertical-align: bottom; display: inline;" src="https://cdn.discordapp.com/emojis/987654321098765432.gif"> world!`,
		},
		{
			name:           "Multiple emojis",
			input:          "Emojis: <:happy:111111111111111111> <a:party:222222222222222222>",
			expectedOutput: `Emojis: <img class="emoji" style="object-fit: contain; width: 1.375em; height: 1.375em; vertical-align: bottom; display: inline;" src="https://cdn.discordapp.com/emojis/111111111111111111.png"> <img class="emoji emoji-animated" style="object-fit: contain; width: 1.375em; height: 1.375em; vertical-align: bottom; display: inline;" src="https://cdn.discordapp.com/emojis/222222222222222222.gif">`,
		},
		{
			name:           "No emojis",
			input:          "Just some text.",
			expectedOutput: "Just some text.",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			output, err := service.Render(test.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if output != test.expectedOutput {
				t.Errorf("expected: %s, got: %s", test.expectedOutput, output)
			}
		})
	}
}

func TestRegularMarkdown(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		input          string
		expectedOutput string
	}{
		{
			name:           "Bold, Italic and Underline",
			input:          "**This is Bold** and *this is italic* and __this is underlined__.",
			expectedOutput: "<strong>This is Bold</strong> and <em>this is italic</em> and <u>this is underlined</u>.",
		},
		{
			name:           "Mixed content",
			input:          "Hello **world**! This is a test with *markdown* and <:emoji:123456789012345678>.",
			expectedOutput: `Hello <strong>world</strong>! This is a test with <em>markdown</em> and <img class="emoji" style="object-fit: contain; width: 1.375em; height: 1.375em; vertical-align: bottom; display: inline;" src="https://cdn.discordapp.com/emojis/123456789012345678.png">.`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			output, err := service.Render(test.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if output != test.expectedOutput {
				t.Errorf("expected: %s, got: %s", test.expectedOutput, output)
			}
		})
	}
}
