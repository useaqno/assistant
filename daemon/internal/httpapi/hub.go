package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Hub fan-outs server-sent events (presence states, mentor/vps alerts) to all
// connected webviews. It replaces the old demo state-cycle with real pushes.
type Hub struct {
	mu      sync.RWMutex
	clients map[chan sseMsg]struct{}
}

type sseMsg struct {
	event string
	data  []byte
}

// NewHub creates an event hub.
func NewHub() *Hub { return &Hub{clients: map[chan sseMsg]struct{}{}} }

// Broadcast sends an event to every connected client (non-blocking).
func (h *Hub) Broadcast(event string, payload any) {
	b, err := json.Marshal(payload)
	if err != nil {
		return
	}
	msg := sseMsg{event: event, data: b}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range h.clients {
		select {
		case ch <- msg:
		default: // drop if the client is slow; presence is best-effort
		}
	}
}

// ServeSSE streams events to one client until it disconnects.
func (h *Hub) ServeSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	ch := make(chan sseMsg, 16)
	h.mu.Lock()
	h.clients[ch] = struct{}{}
	h.mu.Unlock()
	defer func() {
		h.mu.Lock()
		delete(h.clients, ch)
		h.mu.Unlock()
		close(ch)
	}()

	// Initial state so the UI paints immediately.
	send(w, flusher, "presence", []byte(`{"state":"idle","level":0.6}`))

	ping := time.NewTicker(15 * time.Second)
	defer ping.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case msg := <-ch:
			send(w, flusher, msg.event, msg.data)
		case <-ping.C:
			fmt.Fprint(w, ": ping\n\n")
			flusher.Flush()
		}
	}
}

func send(w http.ResponseWriter, f http.Flusher, event string, data []byte) {
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, data)
	f.Flush()
}
