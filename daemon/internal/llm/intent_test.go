package llm

import (
	"testing"
	"time"
)

func TestParseIntentRecurringEvent(t *testing.T) {
	now := time.Date(2026, 6, 23, 9, 0, 0, 0, time.UTC)
	a := ParseIntent("na Cogna tem uma nova daily todo dia das 9:30 às 10:00",
		[]string{"Cogna", "Visa", "Bayer"}, now)
	if a.Tool != "criar_evento" {
		t.Fatalf("tool = %q, want criar_evento", a.Tool)
	}
	if a.Context != "Cogna" {
		t.Errorf("context = %q, want Cogna", a.Context)
	}
	if a.Start != "09:30" || a.End != "10:00" {
		t.Errorf("times = %s-%s, want 09:30-10:00", a.Start, a.End)
	}
	if a.RRule != "FREQ=WEEKLY;BYDAY=MO,TU,WE,TH,FR" {
		t.Errorf("rrule = %q", a.RRule)
	}
	if a.Title != "Daily" {
		t.Errorf("title = %q, want Daily", a.Title)
	}
}

func TestParseIntentTask(t *testing.T) {
	a := ParseIntent("me lembra de ligar pro contador amanhã", nil, time.Now())
	if a.Tool != "criar_tarefa" {
		t.Fatalf("tool = %q, want criar_tarefa", a.Tool)
	}
	if a.Title != "Ligar pro contador" {
		t.Errorf("title = %q, want 'Ligar pro contador'", a.Title)
	}
}

func TestParseIntentQueries(t *testing.T) {
	if a := ParseIntent("como está meu dia?", nil, time.Now()); a.Tool != "consultar_agenda" {
		t.Errorf("agenda query tool = %q", a.Tool)
	}
	if a := ParseIntent("como está o VPS?", nil, time.Now()); a.Tool != "consultar_vps" {
		t.Errorf("vps query tool = %q", a.Tool)
	}
}

func TestParseIntentSingleEventTomorrow(t *testing.T) {
	now := time.Date(2026, 6, 23, 9, 0, 0, 0, time.UTC)
	a := ParseIntent("reunião amanhã às 15h com a Visa", []string{"Visa"}, now)
	if a.Tool != "criar_evento" {
		t.Fatalf("tool = %q", a.Tool)
	}
	if a.Start != "15:00" {
		t.Errorf("start = %q, want 15:00", a.Start)
	}
	if a.Date != "2026-06-24" {
		t.Errorf("date = %q, want 2026-06-24", a.Date)
	}
	if a.Context != "Visa" {
		t.Errorf("context = %q, want Visa", a.Context)
	}
}
