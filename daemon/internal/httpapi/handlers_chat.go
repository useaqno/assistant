package httpapi

import (
	"fmt"
	"net/http"

	"aqnod/internal/model"
)

// Brain produces assistant replies and parses voice intents. The LLM layer
// (WS3) implements it; when absent a minimal canned fallback is used so the
// app remains functional offline.
type Brain interface {
	Chat(conversationID, userText string) (model.ChatMessage, error)
	Stream(conversationID, userText string, emit func(delta string)) (model.ChatMessage, error)
	Intent(transcript, contextName string) (model.ChatMessage, error)
}

func (s *Server) brain() Brain {
	if s.deps.Brain != nil {
		return s.deps.Brain
	}
	return fallbackBrain{s.deps.Store}
}

func (s *Server) handleChatThread(w http.ResponseWriter, r *http.Request) {
	conv := r.URL.Query().Get("conversation")
	if conv == "" {
		id, err := s.deps.Store.EnsureConversation()
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		conv = id
	}
	msgs, err := s.deps.Store.Messages(conv)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if msgs == nil {
		msgs = []model.ChatMessage{}
	}
	writeJSON(w, http.StatusOK, msgs)
}

func (s *Server) handleChatSend(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Text         string `json:"text"`
		Conversation string `json:"conversation"`
	}
	if err := readJSON(r, &b); err != nil || b.Text == "" {
		writeErr(w, http.StatusBadRequest, "text required")
		return
	}
	conv := b.Conversation
	if conv == "" {
		conv, _ = s.deps.Store.EnsureConversation()
	}
	if _, err := s.deps.Store.AddMessage(conv, "user", b.Text, nil); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	reply, err := s.brain().Chat(conv, b.Text)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, reply)
}

func (s *Server) handleChatStream(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeErr(w, http.StatusInternalServerError, "streaming unsupported")
		return
	}
	text := r.URL.Query().Get("text")
	conv := r.URL.Query().Get("conversation")
	if text == "" {
		writeErr(w, http.StatusBadRequest, "text required")
		return
	}
	if conv == "" {
		conv, _ = s.deps.Store.EnsureConversation()
	}
	_, _ = s.deps.Store.AddMessage(conv, "user", text, nil)

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	emit := func(delta string) {
		fmt.Fprintf(w, "event: delta\ndata: %s\n\n", jsonString(delta))
		flusher.Flush()
	}
	final, err := s.brain().Stream(conv, text, emit)
	if err != nil {
		fmt.Fprintf(w, "event: error\ndata: %s\n\n", jsonString(err.Error()))
		flusher.Flush()
		return
	}
	fmt.Fprintf(w, "event: done\ndata: %s\n\n", mustJSON(final))
	flusher.Flush()
}

func (s *Server) handleVoiceIntent(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Transcript string `json:"transcript"`
		Context    string `json:"context"`
	}
	if err := readJSON(r, &b); err != nil || b.Transcript == "" {
		writeErr(w, http.StatusBadRequest, "transcript required")
		return
	}
	msg, err := s.brain().Intent(b.Transcript, b.Context)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, msg)
}
