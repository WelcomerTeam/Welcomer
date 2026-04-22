package welcomer

import (
	"strings"
	"testing"

	"github.com/WelcomerTeam/Discord/discord"
)

func TestURLContainsPathEscape_DetectionPatterns(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		endpoint string
		want     bool
	}{
		{
			name:     "no escape in normal path",
			endpoint: "/channels/123/messages",
			want:     false,
		},
		{
			name:     "contains parent segment in middle",
			endpoint: "/a/../b",
			want:     true,
		},
		{
			name:     "contains current segment in middle",
			endpoint: "/a/./b",
			want:     true,
		},
		{
			name:     "ends with parent segment",
			endpoint: "/a/b/..",
			want:     true,
		},
		{
			name:     "ends with current segment",
			endpoint: "/a/b/.",
			want:     true,
		},
		{
			name:     "double slash but no dot segments",
			endpoint: "/a//b/c",
			want:     false,
		},
		{
			name:     "dot in name only",
			endpoint: "/a/file.txt",
			want:     false,
		},
		{
			name:     "query contains traversal sequence (current behavior)",
			endpoint: "/a/b?next=/../x",
			want:     true,
		},
		{
			name:     "example escape using reactions",
			endpoint: discord.EndpointMessageReaction("1", "2", "reaction_name", "4"),
			want:     false,
		},
		{
			name:     "example escape using reactions",
			endpoint: discord.EndpointMessageReaction("1", "2", "a/../../../../../../guilds/898560324993708032/members/662099680746012686/roles/944746337579192381?", "4"),
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := urlContainsPathEscape(tt.endpoint)
			if err != nil {
				t.Fatalf("urlContainsPathEscape(%q) unexpected error: %v", tt.endpoint, err)
			}

			if got != tt.want {
				t.Fatalf("urlContainsPathEscape(%q) = %v, want %v", tt.endpoint, got, tt.want)
			}
		})
	}
}

func TestURLContainsPathEscape_EncodedPatterns(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		endpoint string
		want     bool
	}{
		{
			name:     "encoded parent segment in middle",
			endpoint: "/a/%2E%2E/b",
			want:     true, // decodes to /a/../b
		},
		{
			name:     "encoded current segment in middle",
			endpoint: "/a/%2E/b",
			want:     true, // decodes to /a/./b
		},
		{
			name:     "encoded trailing parent segment",
			endpoint: "/a/b/%2E%2E",
			want:     true, // decodes to /a/b/..
		},
		{
			name:     "encoded trailing current segment",
			endpoint: "/a/b/%2E",
			want:     true, // decodes to /a/b/.
		},
		{
			name:     "encoded slash only",
			endpoint: "/a%2Fb/c",
			want:     false, // decodes to /a/b/c (no dot traversal segment)
		},
		{
			name:     "example escape using reactions",
			endpoint: discord.EndpointMessageReaction("1", "2", "a%2F..%2F..%2F..%2F..%2F..%2F..%2Fguilds%2F898560324993708032%2Fmembers%2F662099680746012686%2Froles%2F944746337579192381%3F", "4"),
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := urlContainsPathEscape(tt.endpoint)
			if err != nil {
				t.Fatalf("urlContainsPathEscape(%q) unexpected error: %v", tt.endpoint, err)
			}
			if got != tt.want {
				t.Fatalf("urlContainsPathEscape(%q) = %v, want %v", tt.endpoint, got, tt.want)
			}
		})
	}
}

func TestURLContainsPathEscape_InvalidEscapeReturnsError(t *testing.T) {
	t.Parallel()

	endpoint := "/a/%ZZ/b"

	got, err := urlContainsPathEscape(endpoint)
	if err == nil {
		t.Fatalf("urlContainsPathEscape(%q) expected error, got nil", endpoint)
	}
	if got {
		t.Fatalf("urlContainsPathEscape(%q) = %v, want false when error occurs", endpoint, got)
	}
	if !strings.Contains(err.Error(), "failed to unescape url") {
		t.Fatalf("urlContainsPathEscape(%q) error = %q, want message to contain %q", endpoint, err.Error(), "failed to unescape url")
	}
}
