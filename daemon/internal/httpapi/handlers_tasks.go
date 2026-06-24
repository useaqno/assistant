package httpapi

import (
	"net/http"

	"aqnod/internal/domain"
)

func (s *Server) handleTasks(w http.ResponseWriter, _ *http.Request) {
	tasks, err := s.deps.Store.Tasks()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, tasks)
}

func (s *Server) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Context     string `json:"context"`
		Title       string `json:"title"`
		OriginVoice string `json:"originVoice"`
	}
	if err := readJSON(r, &b); err != nil || b.Title == "" {
		writeErr(w, http.StatusBadRequest, "title required")
		return
	}
	id, err := s.deps.Store.CreateTask(b.Context, b.Title, b.OriginVoice)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "id": id})
}

func (s *Server) handlePatchTask(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Done *bool `json:"done"`
	}
	if err := readJSON(r, &b); err != nil || b.Done == nil {
		writeErr(w, http.StatusBadRequest, "done flag required")
		return
	}
	if err := s.deps.Store.SetTaskDone(r.PathValue("id"), *b.Done); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	if err := s.deps.Store.DeleteTask(r.PathValue("id")); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleGraph(w http.ResponseWriter, _ *http.Request) {
	persona, _ := s.deps.Store.Persona()
	g, err := domain.BuildGraph(s.deps.Store, persona.Name)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, g)
}
