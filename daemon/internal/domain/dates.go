package domain

import (
	"fmt"
	"time"
)

var ptWeekday = map[time.Weekday]string{
	time.Sunday: "Domingo", time.Monday: "Segunda", time.Tuesday: "Terça",
	time.Wednesday: "Quarta", time.Thursday: "Quinta", time.Friday: "Sexta",
	time.Saturday: "Sábado",
}

var ptMonth = map[time.Month]string{
	time.January: "jan", time.February: "fev", time.March: "mar", time.April: "abr",
	time.May: "mai", time.June: "jun", time.July: "jul", time.August: "ago",
	time.September: "set", time.October: "out", time.November: "nov", time.December: "dez",
}

// headerDate formats "Segunda · 23 jun · 09:12".
func headerDate(t time.Time) string {
	return fmt.Sprintf("%s · %d %s · %02d:%02d",
		ptWeekday[t.Weekday()], t.Day(), ptMonth[t.Month()], t.Hour(), t.Minute())
}

// dayDate formats "Segunda · 23 jun · 2026".
func dayDate(t time.Time) string {
	return fmt.Sprintf("%s · %d %s · %d", ptWeekday[t.Weekday()], t.Day(), ptMonth[t.Month()], t.Year())
}

// DayDate is the exported header label for the agenda screen.
func DayDate(t time.Time) string { return dayDate(t) }

// greeting picks a time-of-day salutation.
func greeting(t time.Time, owner string) string {
	var g string
	switch h := t.Hour(); {
	case h < 12:
		g = "Bom dia"
	case h < 18:
		g = "Boa tarde"
	default:
		g = "Boa noite"
	}
	if owner != "" {
		return g + ", " + owner + "."
	}
	return g + "."
}
