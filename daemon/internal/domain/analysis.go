package domain

import (
	"fmt"
	"time"

	"aqnod/internal/model"
	"aqnod/internal/store"
)

// BuildAnalysis assembles the daily briefing from live data. Financial and
// personal-life cards are illustrative seeds (no source in the schema yet).
func BuildAnalysis(s *store.Store, now time.Time) (model.Analysis, error) {
	persona, _ := s.Persona()
	events, err := DayEvents(s, now)
	if err != nil {
		return model.Analysis{}, err
	}
	tasks, err := s.Tasks()
	if err != nil {
		return model.Analysis{}, err
	}
	ctxs, _ := s.Contexts()

	meetings, focusMin, meetingMin := 0, 0, 0
	for _, e := range events {
		dur := e.EndMin - e.StartMin
		switch e.Kind {
		case "event":
			meetings++
			meetingMin += dur
		case "focus":
			focusMin += dur
		}
	}
	focusShare := 0.5
	if focusMin+meetingMin > 0 {
		focusShare = float64(focusMin) / float64(focusMin+meetingMin)
	}

	total := len(tasks)
	done := 0
	for _, t := range tasks {
		if t.Done {
			done++
		}
	}
	ratio := 0.0
	if total > 0 {
		ratio = float64(done) / float64(total)
	}

	free := freeFocusHours(events)
	summary := fmt.Sprintf("%s Hoje você tem %d reuniões e %d tarefas. %s",
		greeting(now, persona.Owner), meetings, total, focusNote(free))

	return model.Analysis{
		Summary:    summary,
		Meetings:   meetings,
		Tasks:      total,
		FocusFree:  fmt.Sprintf("%.0fh", free),
		Contexts:   len(ctxs),
		FocusShare: round2(focusShare),
		TasksDone:  fmt.Sprintf("%d/%d", done, total),
		TasksRatio: round2(ratio),
		Apps:       seedApps(),
		CashMonth:  "R$ 48,2k",
		CashDelta:  "▲ 12%",
		CashBars:   []int{46, 62, 54, 78, 70, 100},
		Personal: []model.Metric{
			{Label: "sono", Value: 0.85, Big: "7h", Color: "violet"},
			{Label: "água", Value: 0.55, Big: "1.4L", Color: "blue"},
			{Label: "passos", Value: 0.62, Big: "6.2k", Color: "teal"},
		},
		MentorTitle: "Conselho do mentor",
		MentorBody:  mentorAdvice(events, openTasks(tasks)),
	}, nil
}

func focusNote(free float64) string {
	if free >= 2 {
		return "A manhã está mais leve — protegi seus blocos de foco."
	}
	return "O dia está cheio — vou ajudar a priorizar o que importa."
}

func seedApps() []model.AppHealth {
	return []model.AppHealth{
		{Name: "aqno-api", Status: "ok", Latency: "128 ms", Spark: []int{15, 12, 16, 9, 13, 7, 10, 6}},
		{Name: "Postgres", Status: "ok", Latency: "6 ms", Spark: []int{12, 13, 11, 12, 10, 11, 9, 10}},
		{Name: "Worker da fila", Status: "warn", Latency: "1 retry", Spark: []int{8, 12, 7, 14, 9, 16, 11, 18}},
	}
}

func round2(f float64) float64 { return float64(int(f*100+0.5)) / 100 }
