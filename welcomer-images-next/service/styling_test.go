package service

import (
	"context"
	"testing"

	"github.com/WelcomerTeam/Discord/discord"
	welcomer "github.com/WelcomerTeam/Welcomer/welcomer-core"
)

func TestGetFillAsCSS(t *testing.T) {
	t.Parallel()

	is := &ImageService{}
	ctx := &ImageGenerationContext{
		Context: context.Background(),
		CustomWelcomerImageGenerateRequest: welcomer.CustomWelcomerImageGenerateRequest{
			Guild: discord.Guild{ID: 12345},
		},
	}

	tests := []struct {
		name         string
		value        string
		defaultValue string
		want         string
	}{
		{
			name:         "empty value returns default",
			value:        "",
			defaultValue: "transparent",
			want:         "transparent",
		},
		{
			name:         "hex value is truncated to 8 digits",
			value:        "#a1B2c3D4zz",
			defaultValue: "fallback",
			want:         "#a1B2c3D4",
		},
		{
			name:         "hex value stops at first non-hex character",
			value:        "#123gh",
			defaultValue: "fallback",
			want:         "#123",
		},
		{
			name:         "invalid ref value returns default",
			value:        "ref:artifact name?.png",
			defaultValue: "fallback",
			want:         "fallback",
		},
		{
			name:         "ref value returns asset url",
			value:        "ref:123e4567-e89b-12d3-a456-426614174000",
			defaultValue: "fallback",
			want:         "url(https://www.welcomer.gg/api/guild/12345/welcomer/artifact/123e4567-e89b-12d3-a456-426614174000)",
		},
		{
			name:         "known background value builds asset url",
			value:        "default",
			defaultValue: "fallback",
			want:         "url(https://www.welcomer.gg/assets/backgrounds/default.webp)",
		},
		{
			name:         "unknown value returns default",
			value:        "not-a-background",
			defaultValue: "fallback",
			want:         "fallback",
		},
		{
			name:         "regular hex value is returned as-is",
			value:        "#abc123",
			defaultValue: "fallback",
			want:         "#abc123",
		},
		{
			name:         "rgba hex value is returned as-is",
			value:        "#abc12345",
			defaultValue: "fallback",
			want:         "#abc12345",
		},
		{
			name:         "short hex value is returned as-is",
			value:        "#abc",
			defaultValue: "fallback",
			want:         "#abc",
		},
		{
			name:         "just hash returns default",
			value:        "#",
			defaultValue: "fallback",
			want:         "fallback",
		},
		{
			name:         "invalid 2 length hex value returns default",
			value:        "#12",
			defaultValue: "fallback",
			want:         "fallback",
		},
		{
			name:         "invalid 4 length hex value returns default",
			value:        "#1234",
			defaultValue: "fallback",
			want:         "fallback",
		},
		{
			name:         "invalid 5 length hex value returns default",
			value:        "#12345",
			defaultValue: "fallback",
			want:         "fallback",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := is.getFillAsCSS(ctx, tt.value, tt.defaultValue)
			if got != tt.want {
				t.Fatalf("getFillAsCSS(%q, %q) = %q, want %q", tt.value, tt.defaultValue, got, tt.want)
			}
		})
	}
}
