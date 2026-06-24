package httpapi

import (
	"net/http"

	"aqnod/internal/model"
)

func (s *Server) handleBootstrap(w http.ResponseWriter, _ *http.Request) {
	persona, _ := s.deps.Store.Persona()
	contexts, err := s.deps.Store.Contexts()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	cfg, _ := s.deps.Store.Config()
	// Never leak secrets via config; keys live in the Keychain, not here.
	writeJSON(w, http.StatusOK, model.Bootstrap{
		Persona:   persona,
		Contexts:  contexts,
		Onboarded: s.deps.Store.Onboarded(),
		Config:    cfg,
	})
}

func (s *Server) handleOnboarding(w http.ResponseWriter, r *http.Request) {
	var p model.Persona
	if err := readJSON(r, &p); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid persona")
		return
	}
	if p.Name == "" {
		p.Name = "Aqno"
	}
	if err := s.deps.Store.SavePersona(p); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if p.WakeWord != "" {
		_ = s.deps.Store.SetConfig("voice.wake_word", p.WakeWord)
	}
	if err := s.deps.Store.SetConfig("onboarding.completed", "true"); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleGetConfig(w http.ResponseWriter, _ *http.Request) {
	cfg, err := s.deps.Store.Config()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, cfg)
}

func (s *Server) handleSetConfig(w http.ResponseWriter, r *http.Request) {
	var body map[string]string
	if err := readJSON(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid config")
		return
	}
	voiceChanged := false
	for k, v := range body {
		if err := s.deps.Store.SetConfig(k, v); err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		if len(k) >= 6 && k[:6] == "voice." {
			voiceChanged = true
		}
	}
	if voiceChanged && s.deps.Voice != nil {
		s.deps.Voice.Reload()
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleContexts(w http.ResponseWriter, _ *http.Request) {
	contexts, err := s.deps.Store.Contexts()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, contexts)
}

func (s *Server) handleContextAIMode(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Label string `json:"label"`
		Mode  string `json:"mode"`
	}
	if err := readJSON(r, &body); err != nil || body.Label == "" {
		writeErr(w, http.StatusBadRequest, "label and mode required")
		return
	}
	if err := s.deps.Store.SetContextAIMode(body.Label, body.Mode); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleCreateContext(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Label  string `json:"label"`
		Color  string `json:"color"`
		AIMode string `json:"aiMode"`
	}
	if err := readJSON(r, &body); err != nil || body.Label == "" {
		writeErr(w, http.StatusBadRequest, "label required")
		return
	}
	c, err := s.deps.Store.CreateContext(body.Label, body.Color, body.AIMode)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, c)
}
