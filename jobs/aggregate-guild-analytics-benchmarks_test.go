package main

import (
	"context"
	"errors"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
)

type fakeGuildBenchmarkQueries struct {
	started chan string
	release chan struct{}

	inputParams  database.GetGuildAnalyticsBenchmarkInputsParams
	activeParams database.GetActiveUserCountsParams
	joinParams   database.GetGuildAnalyticsBenchmarkJoinMetricsParams
	inputRows    []*database.GetGuildAnalyticsBenchmarkInputsRow
	activeUsers  *database.GetActiveUserCountsRow
	joinMetrics  *database.GetGuildAnalyticsBenchmarkJoinMetricsRow
	inputErr     error
	activeErr    error
	joinErr      error
}

func (q *fakeGuildBenchmarkQueries) wait(name string) {
	if q.started != nil {
		q.started <- name
		<-q.release
	}
}

func (q *fakeGuildBenchmarkQueries) GetGuildAnalyticsBenchmarkInputs(_ context.Context, params database.GetGuildAnalyticsBenchmarkInputsParams) ([]*database.GetGuildAnalyticsBenchmarkInputsRow, error) {
	q.inputParams = params
	q.wait("input")
	return q.inputRows, q.inputErr
}

func (q *fakeGuildBenchmarkQueries) GetActiveUserCounts(_ context.Context, params database.GetActiveUserCountsParams) (*database.GetActiveUserCountsRow, error) {
	q.activeParams = params
	q.wait("active")
	return q.activeUsers, q.activeErr
}

func (q *fakeGuildBenchmarkQueries) GetGuildAnalyticsBenchmarkJoinMetrics(_ context.Context, params database.GetGuildAnalyticsBenchmarkJoinMetricsParams) (*database.GetGuildAnalyticsBenchmarkJoinMetricsRow, error) {
	q.joinParams = params
	q.wait("join")
	return q.joinMetrics, q.joinErr
}

func TestFetchGuildBenchmarkQueriesRunsConcurrently(t *testing.T) {
	window := welcomer.GuildAnalyticsBenchmarkWindowFor(time.Date(2026, 11, 5, 0, 0, 0, 0, time.UTC))
	input := &database.GetGuildAnalyticsBenchmarkInputsRow{GuildID: 42}
	active := &database.GetActiveUserCountsRow{AverageDailyActiveUsers: 12}
	join := &database.GetGuildAnalyticsBenchmarkJoinMetricsRow{CohortSize: 10}
	queries := &fakeGuildBenchmarkQueries{
		started: make(chan string, 3), release: make(chan struct{}),
		inputRows: []*database.GetGuildAnalyticsBenchmarkInputsRow{input}, activeUsers: active, joinMetrics: join,
	}
	type outcome struct {
		results guildBenchmarkQueryResults
		err     error
	}
	done := make(chan outcome, 1)
	go func() {
		results, err := fetchGuildBenchmarkQueries(context.Background(), queries, 42, window)
		done <- outcome{results, err}
	}()

	seen := map[string]bool{}
	for len(seen) < 3 {
		select {
		case name := <-queries.started:
			seen[name] = true
		case <-time.After(5 * time.Second):
			close(queries.release)
			<-done
			t.Fatalf("queries did not all start before release: %v", seen)
		}
	}
	close(queries.release)
	got := <-done
	if got.err != nil {
		t.Fatal(got.err)
	}
	if got.results.input != input || got.results.activeUsers != active || got.results.joinMetrics != join {
		t.Fatalf("unexpected query results: %+v", got.results)
	}
	if queries.inputParams.GuildID != 42 || queries.inputParams.WeekFrom != window.WeekFrom || queries.inputParams.WeekTo != window.WeekTo {
		t.Errorf("incorrect input parameters: %+v", queries.inputParams)
	}
	if queries.activeParams.GuildID != 42 || queries.activeParams.DateFrom != window.WeekFrom || queries.activeParams.DateTo != window.WeekTo.Add(-time.Microsecond) {
		t.Errorf("incorrect active-user parameters: %+v", queries.activeParams)
	}
	if queries.joinParams.GuildID != 42 || queries.joinParams.CohortFrom != window.CohortFrom || queries.joinParams.CohortTo != window.CohortTo ||
		queries.joinParams.JoinEventType != int32(database.ScienceGuildEventTypeUserJoin) ||
		queries.joinParams.LeaveEventType != int32(database.ScienceGuildEventTypeUserLeave) ||
		queries.joinParams.CohortTo30DayspgCataloginterval != window.CohortTo.AddDate(0, 0, 30) {
		t.Errorf("incorrect join parameters: %+v", queries.joinParams)
	}
}

func TestFetchGuildBenchmarkQueriesErrorsAndMissingInput(t *testing.T) {
	inputErr := errors.New("input failed")
	activeErr := errors.New("active failed")
	joinErr := errors.New("join failed")
	tests := []struct {
		name        string
		rows        []*database.GetGuildAnalyticsBenchmarkInputsRow
		inputErr    error
		activeErr   error
		joinErr     error
		wantErr     error
		wantMessage string
		wantInput   bool
	}{
		{name: "input error takes precedence", inputErr: inputErr, activeErr: activeErr, joinErr: joinErr, wantErr: inputErr, wantMessage: "activity inputs"},
		{name: "missing input skips speculative errors", activeErr: activeErr, joinErr: joinErr},
		{name: "multiple input rows are skipped", rows: []*database.GetGuildAnalyticsBenchmarkInputsRow{{}, {}}},
		{name: "active error takes precedence", rows: []*database.GetGuildAnalyticsBenchmarkInputsRow{{}}, activeErr: activeErr, joinErr: joinErr, wantErr: activeErr, wantMessage: "daily activity", wantInput: true},
		{name: "join error", rows: []*database.GetGuildAnalyticsBenchmarkInputsRow{{}}, joinErr: joinErr, wantErr: joinErr, wantMessage: "retention and onboarding", wantInput: true},
		{name: "success", rows: []*database.GetGuildAnalyticsBenchmarkInputsRow{{}}, wantInput: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queries := &fakeGuildBenchmarkQueries{
				inputRows: tt.rows, activeUsers: &database.GetActiveUserCountsRow{},
				joinMetrics: &database.GetGuildAnalyticsBenchmarkJoinMetricsRow{},
				inputErr:    tt.inputErr, activeErr: tt.activeErr, joinErr: tt.joinErr,
			}
			results, err := fetchGuildBenchmarkQueries(context.Background(), queries, 42, welcomer.GuildAnalyticsBenchmarkWindow{})
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("got error %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr != nil && (!strings.Contains(err.Error(), tt.wantMessage) || !strings.Contains(err.Error(), "guild 42")) {
				t.Errorf("error lacks query and guild context: %v", err)
			}
			if (results.input != nil) != tt.wantInput {
				t.Errorf("got input %v, want presence %v", results.input, tt.wantInput)
			}
		})
	}
}

func TestSampleGuildsPerBucket(t *testing.T) {
	guilds := []*database.GetGuildAnalyticsBenchmarkGuildsRow{
		{GuildID: 1, MemberCount: 1},
		{GuildID: 2, MemberCount: 50},
		{GuildID: 3, MemberCount: 99},
		{GuildID: 4, MemberCount: 100},
		{GuildID: 5, MemberCount: 250},
		{GuildID: 6, MemberCount: 499},
	}

	selected, eligibleCounts := sampleGuildsPerBucket(
		guilds,
		2,
		rand.New(rand.NewSource(1)),
	)
	if len(selected) != 4 {
		t.Fatalf("got %d selected guilds, want 4", len(selected))
	}
	if eligibleCounts[0] != 3 || eligibleCounts[1] != 3 {
		t.Fatalf("got eligible counts %v, want 3 guilds in each bucket", eligibleCounts)
	}

	selectedCounts := map[int16]int{}
	seen := map[int64]bool{}
	for _, guild := range selected {
		bucket, ok := welcomer.GuildAnalyticsBenchmarkBucketForMembers(int64(guild.MemberCount))
		if !ok {
			t.Fatalf("guild %d has no bucket", guild.GuildID)
		}
		selectedCounts[bucket.ID]++
		if seen[guild.GuildID] {
			t.Fatalf("guild %d was selected more than once", guild.GuildID)
		}
		seen[guild.GuildID] = true
	}
	if selectedCounts[0] != 2 || selectedCounts[1] != 2 {
		t.Fatalf("got sampled counts %v, want 2 guilds in each bucket", selectedCounts)
	}
}

func TestSuppressedP90(t *testing.T) {
	if _, ok := suppressedP90(make([]float64, 29)); ok {
		t.Fatal("expected 29 samples to be suppressed")
	}

	for _, tt := range []struct {
		count int
		want  float64
	}{
		{30, 27},
		{31, 28},
		{39, 36},
		{40, 36},
	} {
		values := make([]float64, tt.count)
		for i := range values {
			values[i] = float64(tt.count - i)
		}
		p90, ok := suppressedP90(values)
		if !ok || p90 != tt.want {
			t.Errorf("for %d samples got (%v, %v), want (%v, true)", tt.count, p90, ok, tt.want)
		}
		if values[0] != float64(tt.count) || values[tt.count-1] != 1 {
			t.Errorf("p90 modified the input for %d samples", tt.count)
		}
	}
}
