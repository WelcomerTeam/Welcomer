package backend

import (
	"testing"
	"time"
)

func TestRetentionPercentageFromCohort(t *testing.T) {
	tests := []struct {
		name          string
		cohortSize    int32
		membersLeft   int32
		wantAvailable bool
		want          float64
	}{
		{name: "empty cohort", cohortSize: 0, membersLeft: 0, wantAvailable: false},
		{name: "all members retained", cohortSize: 10, membersLeft: 0, wantAvailable: true, want: 100},
		{name: "half of members retained", cohortSize: 10, membersLeft: 5, wantAvailable: true, want: 50},
		{name: "fractional percentage is truncated to two decimals", cohortSize: 3, membersLeft: 1, wantAvailable: true, want: 66.66},
		{name: "no members retained", cohortSize: 10, membersLeft: 10, wantAvailable: true, want: 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := retentionPercentageFromCohort(test.cohortSize, test.membersLeft)
			if (got != nil) != test.wantAvailable {
				t.Fatalf("retentionPercentageFromCohort(%d, %d) availability = %t, want %t", test.cohortSize, test.membersLeft, got != nil, test.wantAvailable)
			}
			if got != nil && *got != test.want {
				t.Errorf("retentionPercentageFromCohort(%d, %d) = %v, want %v", test.cohortSize, test.membersLeft, *got, test.want)
			}
		})
	}
}

func TestGetPreviousDurationDoesNotOverlap(t *testing.T) {
	from := time.Date(2026, time.November, 3, 0, 0, 0, 0, time.UTC)
	to := from.Add(7 * 24 * time.Hour)

	previousFrom, previousTo := getPreviousDuration(from, to)
	if !previousTo.Before(from) {
		t.Fatalf("previous period ends at %v, want it before current period start %v", previousTo, from)
	}
	if got, want := previousTo.Sub(previousFrom), to.Sub(from); got != want {
		t.Errorf("previous period duration = %v, want %v", got, want)
	}
	if want := from.Add(-time.Microsecond); !previousTo.Equal(want) {
		t.Errorf("previous period end = %v, want %v", previousTo, want)
	}
}

func TestValidateAnalyticsDateRange(t *testing.T) {
	from := time.Date(2026, time.November, 3, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		from    time.Time
		to      time.Time
		wantErr bool
	}{
		{name: "valid range", from: from, to: from.Add(time.Hour)},
		{name: "empty range", from: from, to: from, wantErr: true},
		{name: "reversed range", from: from.Add(time.Hour), to: from, wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateAnalyticsDateRange(test.from, test.to)
			if (err != nil) != test.wantErr {
				t.Fatalf("validateAnalyticsDateRange() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}

func TestIsAllowedFreeAnalyticsDateRange(t *testing.T) {
	location, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}

	now := time.Date(2026, time.March, 18, 12, 0, 0, 0, location)
	today := time.Date(2026, time.March, 18, 0, 0, 0, 0, location)
	endOfDay := func(date time.Time) time.Time {
		return date.AddDate(0, 0, 1).Add(-time.Millisecond)
	}

	tests := []struct {
		name string
		from time.Time
		to   time.Time
		want bool
	}{
		{
			name: "last seven days",
			from: today.AddDate(0, 0, -6),
			to:   now,
			want: true,
		},
		{
			name: "last thirty days",
			from: today.AddDate(0, 0, -29),
			to:   now,
			want: true,
		},
		{
			name: "this week",
			from: today.AddDate(0, 0, -2),
			to:   now,
			want: true,
		},
		{
			name: "last week",
			from: today.AddDate(0, 0, -9),
			to:   endOfDay(today.AddDate(0, 0, -3)),
			want: true,
		},
		{
			name: "this month",
			from: time.Date(2026, time.March, 1, 0, 0, 0, 0, location),
			to:   now,
			want: true,
		},
		{
			name: "last month is not allowed",
			from: time.Date(2026, time.February, 1, 0, 0, 0, 0, location),
			to:   endOfDay(time.Date(2026, time.February, 28, 0, 0, 0, 0, location)),
			want: false,
		},
		{
			name: "custom range is not allowed",
			from: time.Date(2026, time.March, 3, 0, 0, 0, 0, location),
			to:   endOfDay(time.Date(2026, time.March, 7, 0, 0, 0, 0, location)),
			want: false,
		},
		{
			name: "free preset with non-midnight start is not allowed",
			from: today.AddDate(0, 0, -6).Add(time.Hour),
			to:   now,
			want: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := isAllowedFreeAnalyticsDateRange(test.from, test.to, now, location)
			if got != test.want {
				t.Errorf("isAllowedFreeAnalyticsDateRange() = %t, want %t", got, test.want)
			}
		})
	}
}
