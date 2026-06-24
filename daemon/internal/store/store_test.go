package store

import (
	"path/filepath"
	"testing"

	"aqnod/internal/model"
)

func openTemp(t *testing.T) *Store {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	s, err := Open(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestSeedAndCore(t *testing.T) {
	s := openTemp(t)

	if s.Onboarded() {
		t.Error("expected not onboarded on first run")
	}

	ctxs, err := s.Contexts()
	if err != nil || len(ctxs) != 6 {
		t.Fatalf("contexts = %d (%v), want 6", len(ctxs), err)
	}
	if ctxs[0].ID != "cogna" {
		t.Errorf("first context id = %q, want cogna", ctxs[0].ID)
	}

	evs, err := s.RawEvents()
	if err != nil || len(evs) != 5 {
		t.Fatalf("events = %d (%v), want 5", len(evs), err)
	}

	tasks, err := s.Tasks()
	if err != nil || len(tasks) != 5 {
		t.Fatalf("tasks = %d (%v), want 5", len(tasks), err)
	}
	done := 0
	for _, tk := range tasks {
		if tk.Done {
			done++
		}
	}
	if done != 2 {
		t.Errorf("done tasks = %d, want 2", done)
	}

	nodes, edges, err := s.GraphData()
	if err != nil || len(nodes) == 0 || len(edges) == 0 {
		t.Fatalf("graph nodes=%d edges=%d err=%v", len(nodes), len(edges), err)
	}
}

func TestPersonaRoundtrip(t *testing.T) {
	s := openTemp(t)
	p := model.Persona{Name: "Íris", Owner: "Renato", Avatar: "orbe", AuraColor: "#8B5CF6", Tone: "direto", WakeWord: "iris"}
	if err := s.SavePersona(p); err != nil {
		t.Fatal(err)
	}
	got, ok := s.Persona()
	if !ok || got.Name != "Íris" || got.Owner != "Renato" || got.Tone != "direto" {
		t.Fatalf("persona roundtrip = %+v ok=%v", got, ok)
	}
}

func TestConfigAndOnboard(t *testing.T) {
	s := openTemp(t)
	if err := s.SetConfig("onboarding.completed", "true"); err != nil {
		t.Fatal(err)
	}
	if !s.Onboarded() {
		t.Error("expected onboarded after set")
	}
	if got := s.ConfigVal("llm.provider", "x"); got != "anthropic" {
		t.Errorf("llm.provider = %q, want anthropic", got)
	}
}

func TestEventLifecycle(t *testing.T) {
	s := openTemp(t)
	id, err := s.CreateEvent("Cogna", "Nova daily", "reuniao", "10:00", "10:30",
		"FREQ=WEEKLY;BYDAY=MO", "", "na cogna tem daily")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.CancelOccurrence(id, "2026-06-22"); err != nil {
		t.Fatal(err)
	}
	exc, err := s.ExceptionsFor(id)
	if err != nil || len(exc) != 1 || exc[0].Type != "cancelado" {
		t.Fatalf("exceptions = %+v err=%v", exc, err)
	}
	if err := s.DeleteEvent(id); err != nil {
		t.Fatal(err)
	}
}

func TestTaskToggle(t *testing.T) {
	s := openTemp(t)
	id, err := s.CreateTask("Bayer", "Ligar", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.SetTaskDone(id, true); err != nil {
		t.Fatal(err)
	}
	tasks, _ := s.Tasks()
	var found bool
	for _, tk := range tasks {
		if tk.ID == id {
			found = tk.Done
		}
	}
	if !found {
		t.Error("task not marked done")
	}
}
