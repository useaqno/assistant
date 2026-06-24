package httpapi

import (
	"net/http"
	"time"

	"aqnod/internal/domain"
	"aqnod/internal/model"
)

func (s *Server) handleToday(w http.ResponseWriter, _ *http.Request) {
	brief, err := domain.BuildToday(s.deps.Store, now())
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, brief)
}

func (s *Server) handleAnalysis(w http.ResponseWriter, _ *http.Request) {
	a, err := domain.BuildAnalysis(s.deps.Store, now())
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, a)
}

func (s *Server) handleAgenda(w http.ResponseWriter, r *http.Request) {
	day := now()
	if d := r.URL.Query().Get("date"); d != "" {
		if parsed, err := time.Parse("2006-01-02", d); err == nil {
			day = parsed
		}
	}
	events, err := domain.DayEvents(s.deps.Store, day)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	conflicts, focus := 0, 0
	seen := map[string]bool{}
	for _, e := range events {
		if e.Conflict && !seen[e.ID] {
			conflicts++
			seen[e.ID] = true
		}
		if e.Kind == "focus" {
			focus++
		}
	}
	writeJSON(w, http.StatusOK, model.Agenda{
		Day:       domain.DayDate(day),
		Conflicts: conflicts,
		Focus:     focus,
		Events:    events,
	})
}

func (s *Server) handleEventsRange(w http.ResponseWriter, r *http.Request) {
	from, to := now(), now().AddDate(0, 0, 7)
	if f := r.URL.Query().Get("from"); f != "" {
		if p, err := time.Parse("2006-01-02", f); err == nil {
			from = p
		}
	}
	if t := r.URL.Query().Get("to"); t != "" {
		if p, err := time.Parse("2006-01-02", t); err == nil {
			to = p
		}
	}
	events, err := domain.RangeEvents(s.deps.Store, from, to)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, events)
}

type eventBody struct {
	Context     string `json:"context"`
	Title       string `json:"title"`
	Kind        string `json:"kind"`  // event|focus|personal
	Tipo        string `json:"tipo"`  // optional explicit DB type
	Start       string `json:"start"` // HH:MM
	End         string `json:"end"`
	RRule       string `json:"rrule"`
	Date        string `json:"date"` // YYYY-MM-DD (single)
	OriginVoice string `json:"originVoice"`
}

func (b eventBody) tipo() string {
	if b.Tipo != "" {
		return b.Tipo
	}
	switch b.Kind {
	case "focus":
		return "bloco_foco"
	case "personal":
		return "pessoal"
	default:
		return "reuniao"
	}
}

func (s *Server) handleCreateEvent(w http.ResponseWriter, r *http.Request) {
	var b eventBody
	if err := readJSON(r, &b); err != nil || b.Title == "" || b.Start == "" {
		writeErr(w, http.StatusBadRequest, "title and start required")
		return
	}
	if b.RRule == "" && b.Date == "" {
		b.Date = now().Format("2006-01-02")
	}
	id, err := s.deps.Store.CreateEvent(b.Context, b.Title, b.tipo(), b.Start, b.End, b.RRule, b.Date, b.OriginVoice)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "id": id})
}

func (s *Server) handleUpdateEvent(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var b eventBody
	if err := readJSON(r, &b); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid event")
		return
	}
	if err := s.deps.Store.UpdateEvent(id, b.Context, b.Title, b.tipo(), b.Start, b.End, b.RRule, b.Date); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleDeleteEvent(w http.ResponseWriter, r *http.Request) {
	if err := s.deps.Store.DeleteEvent(r.PathValue("id")); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleCancelOccurrence(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Date string `json:"date"`
	}
	_ = readJSON(r, &b)
	if b.Date == "" {
		b.Date = now().Format("2006-01-02")
	}
	if err := s.deps.Store.CancelOccurrence(r.PathValue("id"), b.Date); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}
