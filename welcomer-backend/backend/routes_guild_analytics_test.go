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
