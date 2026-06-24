// aqnod — the Aqno companion daemon.
//
// Runs as a Tauri "sidecar": the desktop shell spawns this binary on launch,
// reads the AQNOD_LISTENING line from stdout to learn the port, and the
// SvelteKit webview talks to it over HTTP (REST + Server-Sent Events).
//
// This is the always-on core: a local SQLite store, the provider-agnostic LLM
// layer, the native voice pipeline and the VPS/SSH connector all hang off it.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"aqnod/internal/httpapi"
	"aqnod/internal/llm"
	"aqnod/internal/store"
)

func main() {
	port := flag.Int("port", 8787, "TCP port to listen on (127.0.0.1)")
	dbPath := flag.String("db", "", "SQLite path (defaults to the app data dir)")
	flag.Parse()

	path := *dbPath
	if path == "" {
		p, err := store.DefaultPath()
		if err != nil {
			log.Fatalf("aqnod: data dir: %v", err)
		}
		path = p
	}

	st, err := store.Open(path)
	if err != nil {
		log.Fatalf("aqnod: open db: %v", err)
	}
	defer st.Close()
	log.Printf("aqnod: database ready at %s (fts=%v)", path, st.HasFTS())

	hub := httpapi.NewHub()
	brain := llm.NewBrain(st)
	api := httpapi.New(httpapi.Deps{Store: st, Hub: hub, Brain: brain})

	addr := fmt.Sprintf("127.0.0.1:%d", *port)
	srv := &http.Server{Addr: addr, Handler: api.Handler()}

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
