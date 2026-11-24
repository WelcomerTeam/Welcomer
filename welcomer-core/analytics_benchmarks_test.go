package welcomer

import (
	"fmt"
	"testing"
	"time"
)

func TestGuildAnalyticsBenchmarkBucketForMembers(t *testing.T) {
	tests := []struct {
		memberCount int64
		bucketID    int16
		wantOK      bool
	}{
		{memberCount: -1, wantOK: false},
		{memberCount: 0, bucketID: 0, wantOK: true},
		{memberCount: 99, bucketID: 0, wantOK: true},
		{memberCount: 100, bucketID: 1, wantOK: true},
		{memberCount: 499, bucketID: 1, wantOK: true},
		{memberCount: 500, bucketID: 2, wantOK: true},
		{memberCount: 1499, bucketID: 2, wantOK: true},
		{memberCount: 1500, bucketID: 3, wantOK: true},
		{memberCount: 3499, bucketID: 3, wantOK: true},
		{memberCount: 3500, bucketID: 4, wantOK: true},
		{memberCount: 9999, bucketID: 4, wantOK: true},
		{memberCount: 10000, bucketID: 5, wantOK: true},
		{memberCount: 99999, bucketID: 5, wantOK: true},
		{memberCount: 100000, bucketID: 6, wantOK: true},
		{memberCount: 1000000, bucketID: 6, wantOK: true},
	}

	for _, test := range tests {
		t.Run(fmt.Sprint(test.memberCount), func(t *testing.T) {
			bucket, ok := GuildAnalyticsBenchmarkBucketForMembers(test.memberCount)
			if ok != test.wantOK {
				t.Fatalf("got ok=%v, want %v", ok, test.wantOK)
			}
			if ok && bucket.ID != test.bucketID {
				t.Fatalf("got bucket ID %d, want %d", bucket.ID, test.bucketID)
			}
		})
	}
}

func TestGuildAnalyticsBenchmarkWindowFor(t *testing.T) {
	now := time.Date(2026, 6, 15, 17, 45, 0, 0, time.FixedZone("local", 3*60*60))
	window := GuildAnalyticsBenchmarkWindowFor(now)

	if window.WeekFrom.Format(time.RFC3339) != "2026-06-08T00:00:00Z" {
		t.Fatalf("got week start %s", window.WeekFrom.Format(time.RFC3339))
	}
	if window.WeekTo.Format(time.RFC3339) != "2026-06-15T00:00:00Z" {
		t.Fatalf("got week end %s", window.WeekTo.Format(time.RFC3339))
	}
	if window.CohortFrom.Format(time.RFC3339) != "2026-04-16T00:00:00Z" {
		t.Fatalf("got cohort start %s", window.CohortFrom.Format(time.RFC3339))
	}
	if window.CohortTo.Format(time.RFC3339) != "2026-05-16T00:00:00Z" {
		t.Fatalf("got cohort end %s", window.CohortTo.Format(time.RFC3339))
	}
}

func TestAverageDiscordAccountAgeDays(t *testing.T) {
	now := time.UnixMilli(1420070400000 + 48*time.Hour.Milliseconds())
	oneDayOldID := int64(uint64(24*time.Hour.Milliseconds()) << 22)
	twoDaysOldID := int64(1)

	average, ok := AverageDiscordAccountAgeDays([]int64{oneDayOldID, twoDaysOldID, 0}, now)
	if !ok {
		t.Fatal("expected at least one usable account ID")
	}
	if average != 1.5 {
		t.Fatalf("got average age %v days, want 1.5", average)
	}

	if _, ok := AverageDiscordAccountAgeDays([]int64{0, -1}, now); ok {
		t.Fatal("expected no usable account IDs")
	}
}
