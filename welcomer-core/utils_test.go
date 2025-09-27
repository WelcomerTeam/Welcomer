package welcomer

import (
	"testing"

	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
)

func TestParseDurationAsSeconds(t *testing.T) {
	tests := []struct {
		input          string
		expectedOutput int
		expectedError  bool
	}{
		// Valid cases
		{"10", 10, false},
		{"1h", 3600, false},
		{"1hour", 3600, false},
		{"2h30m", 9000, false},
		{"1d", 86400, false},
		{"1y", 31536000, false},
		{"1y2d3h4m5s", 31719845, false},
		{"  1h 30m  ", 5400, false},
		{"2d 3h", 183600, false},
		{"1 hour 30 minutes", 5400, false},
		{"1 day 2 hours", 93600, false},
		{"1y 2d 3h 4m 5s", 31719845, false},

		{"", 0, false},

		// // Invalid cases
		{"abc", 0, true},
		{"1x", 0, true},
		{"1h2x", 0, true},
		{"1.5h", 0, true},
		{"1h 30", 0, true},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			t.Parallel()

			output, err := ParseDurationAsSeconds(test.input)
			if (err != nil) != test.expectedError {
				t.Errorf("expected error: %v, got: %v", test.expectedError, err)
			}
			if output != test.expectedOutput {
				t.Errorf("expected output: %d, got: %d", test.expectedOutput, output)
			}
		})
	}
}
func TestFormatNumber(t *testing.T) {
	type args struct {
		value  int64
		locale database.NumberLocale
	}
	tests := []struct {
		name     string
		args     args
		expected string
	}{
		{
			name:     "English locale (commas)",
			args:     args{value: 1234567, locale: database.NumberLocaleCommas},
			expected: "1,234,567",
		},
		{
			name:     "German locale (dots)",
			args:     args{value: 1234567, locale: database.NumberLocaleDots},
			expected: "1.234.567",
		},
		{
			name:     "Arabic locale",
			args:     args{value: 1234567, locale: database.NumberLocaleArabic},
			expected: "١٬٢٣٤٬٥٦٧",
		},
		{
			name:     "Default locale (fallback to English)",
			args:     args{value: 1234567},
			expected: "1234567",
		},
		{
			name:     "Zero value",
			args:     args{value: 0, locale: database.NumberLocaleCommas},
			expected: "0",
		},
		{
			name:     "Negative value",
			args:     args{value: -1234567, locale: database.NumberLocaleCommas},
			expected: "-1,234,567",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got := FormatNumber(test.args.value, test.args.locale)
			if got != test.expected {
				t.Errorf("FormatNumber(%d, %v) = %q, want %q", test.args.value, test.args.locale, got, test.expected)
			}
		})
	}
}
