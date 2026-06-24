package domain

import (
	"fmt"
	"sort"
	"time"

	"aqnod/internal/model"
	"aqnod/internal/store"
)

const (
	workStart = 8 * 60  // 08:00
	workEnd   = 18 * 60 // 18:00
)

// RangeEvents expands every stored event into concrete occurrences in [from,to].
func RangeEvents(s *store.Store, from, to time.Time) ([]model.Event, error) {
	raw, err := s.RawEvents()
	if err != nil {
		return nil, err
	}
	return ExpandRange(raw, loadExceptions(s, raw), from, to), nil
}

// DayEvents expands the agenda for a single calendar day.
func DayEvents(s *store.Store, day time.Time) ([]model.Event, error) {
	return RangeEvents(s, day, day)
}

// loadExceptions fetches the per-occurrence overrides for the given events.
func loadExceptions(s *store.Store, raw []store.RawEvent) map[string][]store.Exception {
	excs := map[string][]store.Exception{}
	for _, e := range raw {
		if ex, _ := s.ExceptionsFor(e.ID); len(ex) > 0 {
			excs[e.ID] = ex
		}
	}
	return excs
}

// freeFocusHours returns free hours inside work hours after merging events.
func freeFocusHours(evs []model.Event) float64 {
	covered := mergeIntervals(evs)
	free := (workEnd - workStart) - covered
	if free < 0 {
		free = 0
	}
	return float64(free) / 60
}

func mergeIntervals(evs []model.Event) int {
	type iv struct{ s, e int }
	var ivs []iv
	for _, e := range evs {
		s, en := clamp(e.StartMin), clamp(e.EndMin)
		if en > s {
			ivs = append(ivs, iv{s, en})
		}
	}
	sort.Slice(ivs, func(i, j int) bool { return ivs[i].s < ivs[j].s })
	total, curS, curE := 0, -1, -1
	for _, v := range ivs {
		if v.s > curE {
			if curE > curS {
				total += curE - curS
			}
			curS, curE = v.s, v.e
		} else if v.e > curE {
			curE = v.e
		}
	}
	if curE > curS {
		total += curE - curS
	}
	return total
}

func clamp(m int) int {
	if m < workStart {
		return workStart
	}
	if m > workEnd {
		return workEnd
	}
	return m
}

// BuildToday assembles the Home dashboard payload from live data.
func BuildToday(s *store.Store, now time.Time) (model.TodayBrief, error) {
	persona, _ := s.Persona()
	events, err := DayEvents(s, now)
	if err != nil {
		return model.TodayBrief{}, err
	}
	tasks, err := s.Tasks()
	if err != nil {
		return model.TodayBrief{}, err
	}

	meetings := 0
	for _, e := range events {
		if e.Kind == "event" {
			meetings++
		}
	}
	open := openTasks(tasks)
	focus := freeFocusHours(events)

	next := nextEvent(events, now)
	mentor := mentorAdvice(events, open)

	recent, _ := s.RecentInteractions(3)

	companion := persona.Name
	if companion == "" {
		companion = "Aqno"
	}

	return model.TodayBrief{
		Greeting:  greeting(now, persona.Owner),
		Date:      headerDate(now),
		Companion: companion,
		State:     "idle",
		Headline:  fmt.Sprintf("Você tem %d reuniões e %d tarefas hoje. Quer que eu prepare um resumo?", meetings, len(open)),
		Meetings:  meetings,
		Tasks:     len(open),
		FocusFree: fmt.Sprintf("%.0fh", focus),
		NextEvent: next,
		TaskList:  topTasks(tasks, 4),
		Mentor:    mentor,
		Recent:    recent,
	}, nil
}

func openTasks(tasks []model.Task) []model.Task {
	var out []model.Task
	for _, t := range tasks {
		if !t.Done {
			out = append(out, t)
		}
	}
	return out
}

func topTasks(tasks []model.Task, n int) []model.Task {
	if len(tasks) > n {
		return tasks[:n]
	}
	return tasks
}

func nextEvent(events []model.Event, now time.Time) model.Event {
	mins := now.Hour()*60 + now.Minute()
	for _, e := range events {
		if e.EndMin >= mins {
			return e
		}
	}
	if len(events) > 0 {
		return events[len(events)-1]
	}
	return model.Event{Title: "Sem eventos hoje", Start: "--:--", End: "--:--"}
}

func mentorAdvice(events []model.Event, open []model.Task) string {
	for _, e := range events {
		if e.Conflict {
			return fmt.Sprintf("Há um conflito de horário às %s. Quer que eu remarque um dos eventos para liberar a agenda?", e.Start)
		}
	}
	if freeFocusHours(events) >= 2 {
		return "Você tem boas janelas de foco hoje. Proteja uma delas para avançar as entregas mais críticas antes das reuniões."
	}
	if len(open) > 0 {
		return fmt.Sprintf("Você tem %d tarefas em aberto. Posso priorizá-las por contexto e reservar um bloco de foco.", len(open))
	}
	return "Tudo sob controle hoje. Me chame por voz se algo novo aparecer."
}
