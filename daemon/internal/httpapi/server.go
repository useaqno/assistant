// Package httpapi exposes the daemon's domain over HTTP (REST + SSE). Handlers
// stay thin: parse -> call store/domain -> respond. It maps the IPC contract in
// docs/context.md §5 onto concrete routes the SvelteKit client consumes.
package httpapi

import (
	"encoding/json"
	"net/http"
	"time"

	"aqnod/internal/store"
)

// Deps are the services the API needs. Voice/LLM/SSH are attached as they land.
type Deps struct {
	Store *store.Store
	Hub   *Hub   // SSE/event broadcaster
	Brain Brain  // LLM assistant (nil -> canned fallback)
	Vps   VpsProvider // SSH/metrics provider (nil -> stub fixture)
}

// Server holds dependencies and builds the router.
type Server struct {
	deps Deps
}

// New creates the API server.
func New(deps Deps) *Server { return &Server{deps: deps} }

// now is overridable in tests; production uses the wall clock.
var now = time.Now

// Handler builds the mux (Go 1.22+ method+wildcard routing).
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", s.handleHealth)

	// Bootstrap / onboarding / config
	mux.HandleFunc("GET /v1/bootstrap", s.handleBootstrap)
	mux.HandleFunc("POST /v1/onboarding", s.handleOnboarding)
	mux.HandleFunc("GET /v1/config", s.handleGetConfig)
	mux.HandleFunc("POST /v1/config", s.handleSetConfig)

	// Contexts
	mux.HandleFunc("GET /v1/contexts", s.handleContexts)
	mux.HandleFunc("POST /v1/contexts", s.handleCreateContext)

	// Dashboards
	mux.HandleFunc("GET /v1/today", s.handleToday)
	mux.HandleFunc("GET /v1/analysis", s.handleAnalysis)
	mux.HandleFunc("GET /v1/agenda", s.handleAgenda)

	// Calendar
	mux.HandleFunc("GET /v1/events/range", s.handleEventsRange)
	mux.HandleFunc("POST /v1/events", s.handleCreateEvent)
	mux.HandleFunc("PUT /v1/events/{id}", s.handleUpdateEvent)
	mux.HandleFunc("DELETE /v1/events/{id}", s.handleDeleteEvent)
	mux.HandleFunc("POST /v1/events/{id}/cancel", s.handleCancelOccurrence)

	// Tasks
	mux.HandleFunc("GET /v1/tasks", s.handleTasks)
	mux.HandleFunc("POST /v1/tasks", s.handleCreateTask)
	mux.HandleFunc("PATCH /v1/tasks/{id}", s.handlePatchTask)
	mux.HandleFunc("DELETE /v1/tasks/{id}", s.handleDeleteTask)

	// Graph
	mux.HandleFunc("GET /v1/graph", s.handleGraph)

	// Chat (LLM wired in WS3)
	mux.HandleFunc("GET /v1/chat", s.handleChatThread)
	mux.HandleFunc("POST /v1/chat", s.handleChatSend)
	mux.HandleFunc("GET /v1/chat/stream", s.handleChatStream)

	// Voice intent (heuristic now; LLM in WS3, audio in WS4)
	mux.HandleFunc("POST /v1/voice/intent", s.handleVoiceIntent)

	// VPS (real SSH in WS5)
	mux.HandleFunc("GET /v1/vps", s.handleVps)
	mux.HandleFunc("POST /v1/vps/restart", s.handleRestart)

	// Live events (SSE)
	mux.HandleFunc("GET /v1/events", s.deps.Hub.ServeSSE)

	return withCORS(mux)
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"ok": true, "service": "aqnod", "version": "0.2.0",
		"time": now().Format(time.RFC3339),
	})
}

// ---- helpers ----

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]any{"ok": false, "error": map[string]string{"message": msg}})
}

func readJSON(r *http.Request, dst any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(dst)
}

// withCORS lets the Tauri webview reach the daemon during development.
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
