package domain

import (
	"testing"
	"time"

	"aqnod/internal/model"
	"aqnod/internal/store"
)

func date(s string) time.Time {
	t, _ := time.Parse("2006-01-02", s)
	return t
}

func TestRRuleWeeklyByDay(t *testing.T) {
	r := ParseRRule("FREQ=WEEKLY;BYDAY=MO,WE,FR")
	// 2026-06-22 is a Monday.
	cases := map[string]bool{
		"2026-06-22": true,  // Mon
		"2026-06-23": false, // Tue
		"2026-06-24": true,  // Wed
		"2026-06-25": false, // Thu
		"2026-06-26": true,  // Fri
		"2026-06-27": false, // Sat
	}
	for d, want := range cases {
		if got := r.Occurs(date(d)); got != want {
			t.Errorf("Occurs(%s) = %v, want %v", d, got, want)
		}
	}
}

func TestRRuleDailyInterval(t *testing.T) {
	r := ParseRRule("FREQ=DAILY;INTERVAL=2")
	// Two consecutive days cannot both occur with interval 2.
	d0 := date("2026-06-22")
	if r.Occurs(d0) == r.Occurs(d0.AddDate(0, 0, 1)) {
		t.Error("interval=2 should alternate days")
	}
}

func TestRRuleUntil(t *testing.T) {
	r := ParseRRule("FREQ=DAILY;UNTIL=20260623")
	if !r.Occurs(date("2026-06-23")) {
		t.Error("should occur on UNTIL date")
	}
	if r.Occurs(date("2026-06-24")) {
		t.Error("should not occur after UNTIL")
	}
}

func TestExpandRangeWithConflict(t *testing.T) {
	raw := []store.RawEvent{
		{Event: model.Event{ID: "1", Title: "A", Start: "14:00", End: "15:00", Kind: "event"}, DataUnica: "2026-06-23"},
		{Event: model.Event{ID: "2", Title: "B", Start: "14:00", End: "14:30", Kind: "event"}, DataUnica: "2026-06-23"},
		{Event: model.Event{ID: "3", Title: "C", Start: "16:00", End: "16:30", Kind: "event"}, DataUnica: "2026-06-23"},
	}
	out := ExpandRange(raw, nil, date("2026-06-23"), date("2026-06-23"))
	if len(out) != 3 {
		t.Fatalf("got %d events, want 3", len(out))
	}
	conflicts := 0
	for _, e := range out {
		if e.Conflict {
			conflicts++
		}
	}
	if conflicts != 2 {
		t.Errorf("got %d conflicting events, want 2", conflicts)
	}
}

func TestExpandRangeException(t *testing.T) {
	raw := []store.RawEvent{
		{Event: model.Event{ID: "1", Title: "Daily", Start: "09:30", End: "10:00", Kind: "event", RRule: "FREQ=WEEKLY;BYDAY=MO"}},
	}
	excs := map[string][]store.Exception{
		"1": {{Date: "2026-06-22", Type: "cancelado"}},
	}
	// Range covering two Mondays: 2026-06-22 (cancelled) and 2026-06-29.
	out := ExpandRange(raw, excs, date("2026-06-22"), date("2026-06-29"))
	if len(out) != 1 || out[0].Date != "2026-06-29" {
		t.Fatalf("expected only 2026-06-29 occurrence, got %+v", out)
	}
}
