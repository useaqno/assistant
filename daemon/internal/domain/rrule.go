// Package domain holds calendar/briefing computations that sit above the store:
// RRULE expansion, conflict detection, and the Today/Analysis aggregations.
package domain

import (
	"strconv"
	"strings"
	"time"
)

// anchor is a fixed Monday used to make INTERVAL parity deterministic when an
// event has no stored DTSTART (the schema keeps only 'HH:MM' + RRULE).
var anchor = time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC) // 2001-01-01 was a Monday

// RRule is the supported subset of iCalendar recurrence (docs/context.md §6.4).
type RRule struct {
	Freq     string // DAILY | WEEKLY | MONTHLY
	Interval int
	ByDay    []time.Weekday
	Count    int
	Until    time.Time
}

var weekdayCode = map[string]time.Weekday{
	"SU": time.Sunday, "MO": time.Monday, "TU": time.Tuesday, "WE": time.Wednesday,
	"TH": time.Thursday, "FR": time.Friday, "SA": time.Saturday,
}

// ParseRRule parses a string like "FREQ=WEEKLY;BYDAY=MO,TU;INTERVAL=1".
func ParseRRule(s string) RRule {
	r := RRule{Interval: 1}
	for _, part := range strings.Split(s, ";") {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		key, val := strings.ToUpper(strings.TrimSpace(kv[0])), strings.TrimSpace(kv[1])
		switch key {
		case "FREQ":
			r.Freq = strings.ToUpper(val)
		case "INTERVAL":
			if n, err := strconv.Atoi(val); err == nil && n > 0 {
				r.Interval = n
			}
		case "COUNT":
			if n, err := strconv.Atoi(val); err == nil {
				r.Count = n
			}
		case "UNTIL":
			r.Until = parseUntil(val)
		case "BYDAY":
			for _, code := range strings.Split(val, ",") {
				if wd, ok := weekdayCode[strings.ToUpper(strings.TrimSpace(code))]; ok {
					r.ByDay = append(r.ByDay, wd)
				}
			}
		}
	}
	return r
}

func parseUntil(v string) time.Time {
	for _, layout := range []string{"20060102T150405Z", "20060102T150405", "20060102", "2006-01-02"} {
		if t, err := time.Parse(layout, v); err == nil {
			return t
		}
	}
	return time.Time{}
}

// dayOnly truncates to date in UTC for parity math.
func dayOnly(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

// Occurs reports whether the recurrence fires on calendar day d.
func (r RRule) Occurs(d time.Time) bool {
	d = dayOnly(d)
	if !r.Until.IsZero() && d.After(dayOnly(r.Until)) {
		return false
	}
	days := int(d.Sub(anchor).Hours() / 24)
	switch r.Freq {
	case "DAILY":
		if days < 0 || days%r.Interval != 0 {
			return false
		}
		return r.withinCount(days/r.Interval + 1)
	case "WEEKLY":
		week := floorDiv(days, 7)
		if week < 0 || week%r.Interval != 0 {
			return false
		}
		if !r.matchesByDay(d) {
			return false
		}
		return true // COUNT for weekly is best-effort; UNTIL above is authoritative
	case "MONTHLY":
		months := (d.Year()-anchor.Year())*12 + int(d.Month()-anchor.Month())
		if months < 0 || months%r.Interval != 0 {
			return false
		}
		return r.matchesMonthDay(d)
	default:
		return false
	}
}

func (r RRule) matchesByDay(d time.Time) bool {
	if len(r.ByDay) == 0 {
		return d.Weekday() == time.Monday // weekly without BYDAY: anchor weekday
	}
	for _, wd := range r.ByDay {
		if wd == d.Weekday() {
			return true
		}
	}
	return false
}

func (r RRule) matchesMonthDay(d time.Time) bool {
	// Without BYMONTHDAY the schema can't anchor the day; default to the 1st.
	return d.Day() == 1
}

func (r RRule) withinCount(n int) bool {
	if r.Count <= 0 {
		return true
	}
	return n <= r.Count
}

func floorDiv(a, b int) int {
	q := a / b
	if (a%b != 0) && ((a < 0) != (b < 0)) {
		q--
	}
	return q
}
