package domain

import (
	"sort"
	"strconv"
	"strings"
	"time"

	"aqnod/internal/model"
	"aqnod/internal/store"
)

// minutesOf parses 'HH:MM' to minutes-from-midnight (-1 on failure).
func minutesOf(hhmm string) int {
	parts := strings.SplitN(hhmm, ":", 2)
	if len(parts) != 2 {
		return -1
	}
	h, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
	m, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err1 != nil || err2 != nil {
		return -1
	}
	return h*60 + m
}

// ExpandRange turns stored events into concrete occurrences within [from,to]
// (inclusive, date granularity), applying per-occurrence exceptions and
// flagging time conflicts. Results are sorted by date then start.
func ExpandRange(events []store.RawEvent, excs map[string][]store.Exception, from, to time.Time) []model.Event {
	from, to = dayOnly(from), dayOnly(to)
	var out []model.Event

	for _, e := range events {
		excList := excs[e.ID]
		if e.RRule == "" {
			// Single event: emit if its date falls in range.
			date := e.DataUnica
			if date == "" {
				date = from.Format("2006-01-02")
			}
			d, err := time.Parse("2006-01-02", date)
			if err != nil || d.Before(from) || d.After(to) {
				continue
			}
			if occ, ok := occurrence(e, date, excList); ok {
				out = append(out, occ)
			}
			continue
		}
		// Recurring: walk each day in the window.
		rule := ParseRRule(e.RRule)
		for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
			if !rule.Occurs(d) {
				continue
			}
			date := d.Format("2006-01-02")
			if occ, ok := occurrence(e, date, excList); ok {
				out = append(out, occ)
			}
		}
	}

	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Date != out[j].Date {
			return out[i].Date < out[j].Date
		}
		return out[i].StartMin < out[j].StartMin
	})
	markConflicts(out)
	return out
}

// occurrence builds one concrete event for a date, honouring exceptions.
func occurrence(e store.RawEvent, date string, excs []store.Exception) (model.Event, bool) {
	start, end := e.Start, e.End
	for _, x := range excs {
		if x.Date != date {
			continue
		}
		if x.Type == "cancelado" {
			return model.Event{}, false
		}
		if x.Type == "remarcado" {
			if x.NewStart != "" {
				start = x.NewStart
			}
			if x.NewEnd != "" {
				end = x.NewEnd
			}
		}
	}
	occ := e.Event // copy embedded model.Event
	occ.Start, occ.End = start, end
	occ.Date = date
	occ.StartMin = minutesOf(start)
	if occ.StartMin < 0 {
		occ.StartMin = 0
	}
	occ.EndMin = minutesOf(end)
	if occ.EndMin <= occ.StartMin {
		occ.EndMin = occ.StartMin + 30
	}
	occ.RRule = e.RRule
	return occ, true
}

// markConflicts flags events on the same date whose times overlap.
func markConflicts(evs []model.Event) {
	for i := range evs {
		for j := i + 1; j < len(evs); j++ {
			if evs[i].Date != evs[j].Date {
				continue
			}
			if evs[i].StartMin < evs[j].EndMin && evs[j].StartMin < evs[i].EndMin {
				evs[i].Conflict = true
				evs[j].Conflict = true
			}
		}
	}
}
