package httpapi

import (
	"fmt"
	"net/http"

	"aqnod/internal/keychain"
	"aqnod/internal/llm"
	"aqnod/internal/sshvps"
)

func (s *Server) handleServers(w http.ResponseWriter, _ *http.Request) {
	servers, err := s.deps.Store.Servers()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, servers)
}

func (s *Server) handleCreateServer(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Name     string `json:"name"`
		Host     string `json:"host"`
		Port     int    `json:"port"`
		User     string `json:"user"`
		AuthType string `json:"authType"` // senha | chave
		Secret   string `json:"secret"`   // password or private-key PEM
	}
	if err := readJSON(r, &b); err != nil || b.Host == "" || b.User == "" {
		writeErr(w, http.StatusBadRequest, "host and user required")
		return
	}
	ref := fmt.Sprintf("ssh-%s-%s", b.Host, b.User)
	if b.Secret != "" {
		if err := keychain.Set(sshvps.KeychainService, ref, b.Secret); err != nil {
			writeErr(w, http.StatusInternalServerError, "keychain: "+err.Error())
			return
		}
	}
	name := b.Name
	if name == "" {
		name = b.Host
	}
	id, err := s.deps.Store.CreateServer(name, b.Host, b.Port, b.User, b.AuthType, ref)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	_ = s.deps.Store.AddAudit("config", "server.create", b.Host)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "id": id})
}

func (s *Server) handleDeleteServer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if srv, ok := s.deps.Store.ServerByID(id); ok {
		_ = keychain.Delete(sshvps.KeychainService, srv.KeychainRef)
	}
	if err := s.deps.Store.DeleteServer(id); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// handleSetLLMKey stores a provider API key in the Keychain (never in SQLite).
func (s *Server) handleSetLLMKey(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Provider string `json:"provider"`
		Key      string `json:"key"`
	}
	if err := readJSON(r, &b); err != nil || b.Provider == "" {
		writeErr(w, http.StatusBadRequest, "provider required")
		return
	}
	if b.Key == "" {
		_ = keychain.Delete(llm.KeychainService, b.Provider)
	} else if err := keychain.Set(llm.KeychainService, b.Provider, b.Key); err != nil {
		writeErr(w, http.StatusInternalServerError, "keychain: "+err.Error())
		return
	}
	_ = s.deps.Store.AddAudit("config", "llm.key.set", b.Provider)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// handleLLMKeyStatus reports whether a key is stored for a provider (no secret).
func (s *Server) handleLLMKeyStatus(w http.ResponseWriter, r *http.Request) {
	provider := r.URL.Query().Get("provider")
	if provider == "" {
		provider = s.deps.Store.ConfigVal("llm.provider", "anthropic")
	}
	_, err := keychain.Get(llm.KeychainService, provider)
	writeJSON(w, http.StatusOK, map[string]any{"provider": provider, "configured": err == nil})
}
