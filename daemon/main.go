// aqnod — the Aqno companion daemon.
//
// Runs as a Tauri "sidecar": the desktop shell spawns this binary on launch,
// reads the AQNOD_LISTENING line from stdout to learn the port, and the
// SvelteKit webview talks to it over HTTP (REST + Server-Sent Events).
//
// In production this process is where the always-on work lives: audio capture,
// transcription, the local models, the SQLite memory store and the VPS links.
// Here it serves stable fixtures so the whole UI is wired end-to-end.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	port := flag.Int("port", 8787, "TCP port to listen on (127.0.0.1)")
	flag.Parse()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/v1/contexts", jsonHandler(func(*http.Request) any { return contexts }))
	mux.HandleFunc("/v1/today", jsonHandler(func(*http.Request) any { return todayBrief() }))
	mux.HandleFunc("/v1/agenda", jsonHandler(func(*http.Request) any { return agenda() }))
	mux.HandleFunc("/v1/analysis", jsonHandler(func(*http.Request) any { return analysis() }))
	mux.HandleFunc("/v1/vps", jsonHandler(func(*http.Request) any { return vps() }))
	mux.HandleFunc("/v1/chat", handleChat)
	mux.HandleFunc("/v1/graph", jsonHandler(func(*http.Request) any { return graph() }))
	mux.HandleFunc("/v1/vps/restart", handleRestart)
	mux.HandleFunc("/v1/events", handleEvents)

	addr := fmt.Sprintf("127.0.0.1:%d", *port)
	srv := &http.Server{Addr: addr, Handler: withCORS(mux)}

	go func() {
		// Contract with the Tauri sidecar supervisor: announce the live port on
		// stdout so the Rust side can wait for readiness before loading the UI.
		fmt.Printf("AQNOD_LISTENING %d\n", *port)
		os.Stdout.Sync()
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("aqnod: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Println("aqnod: shutting down")
	srv.Close()
}

// withCORS lets the Tauri webview (custom scheme / localhost) reach the daemon
// during development. Tighten the origin allow-list for production builds.
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func jsonHandler(fn func(*http.Request) any) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, fn(r))
	}
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":      true,
		"service": "aqnod",
		"version": "0.1.0",
		"time":    time.Now().Format(time.RFC3339),
	})
}

func handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var body struct {
			Text string `json:"text"`
		}
		json.NewDecoder(r.Body).Decode(&body)
		// Echo a stub reply; a real build would route to the model + memory.
		reply := ChatMessage{
			ID:   fmt.Sprintf("m%d", time.Now().UnixMilli()),
			From: "aqno",
			Text: "Entendi — vou cuidar disso e te confirmo em seguida.",
			Time: time.Now().Format("15:04"),
		}
		writeJSON(w, http.StatusOK, reply)
		return
	}
	writeJSON(w, http.StatusOK, chatThread())
}

func handleRestart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "use POST"})
		return
	}
	var body struct {
		Container string `json:"container"`
		Confirm   bool   `json:"confirm"`
	}
	json.NewDecoder(r.Body).Decode(&body)
	if !body.Confirm {
		// Explicit confirmation is required — destructive action.
		writeJSON(w, http.StatusAccepted, map[string]any{
			"needsConfirm": true,
			"container":    body.Container,
			"message":      fmt.Sprintf("Reiniciar %s? O container ficará indisponível por ~3s.", body.Container),
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":        true,
		"container": body.Container,
		"state":     "restarting",
		"message":   fmt.Sprintf("Container reiniciado · %s", body.Container),
	})
}

// handleEvents streams presence-state and log ticks over SSE, giving the UI its
// "living" feel without polling. Cycles through the six presence states.
func handleEvents(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	states := []string{"idle", "listening", "transcribing", "thinking", "speaking", "confirming"}
	ticker := time.NewTicker(2200 * time.Millisecond)
	defer ticker.Stop()

	i := 0
	send := func(ev string, payload any) {
		b, _ := json.Marshal(payload)
		fmt.Fprintf(w, "event: %s\ndata: %s\n\n", ev, b)
		flusher.Flush()
	}
	send("presence", map[string]string{"state": "idle"})
	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			i++
			send("presence", map[string]any{"state": states[i%len(states)], "level": 0.4 + 0.5*float64(i%3)/2})
		}
	}
}
