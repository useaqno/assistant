package httpapi

import (
	"fmt"
	"net/http"

	"aqnod/internal/model"
)

// VpsProvider returns infra metrics and performs (confirmed) actions. The SSH
// layer (WS5) implements it; a stub fixture is used when no server is set up.
type VpsProvider interface {
	Snapshot() (model.Vps, error)
	Restart(container string) (string, error)
}

func (s *Server) vps() VpsProvider {
	if s.deps.Vps != nil {
		return s.deps.Vps
	}
	return stubVps{}
}

func (s *Server) handleVps(w http.ResponseWriter, _ *http.Request) {
	v, err := s.vps().Snapshot()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, v)
}

func (s *Server) handleRestart(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Container string `json:"container"`
		Confirm   bool   `json:"confirm"`
	}
	_ = readJSON(r, &b)
	if b.Container == "" {
		writeErr(w, http.StatusBadRequest, "container required")
		return
	}
	// Destructive action: require explicit confirmation (docs/context.md §17).
	if !b.Confirm {
		writeJSON(w, http.StatusAccepted, map[string]any{
			"needsConfirm": true,
			"container":    b.Container,
			"message":      fmt.Sprintf("Reiniciar %s? O container ficará indisponível por alguns segundos.", b.Container),
		})
		return
	}
	msg, err := s.vps().Restart(b.Container)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}
	_ = s.deps.Store.AddAudit("vps", "restart", b.Container)
	s.deps.Hub.Broadcast("vps.alert", map[string]string{"level": "ok", "message": msg})
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "container": b.Container, "message": msg})
}

// stubVps serves plausible fixtures until a real server is registered.
type stubVps struct{}

func (stubVps) Snapshot() (model.Vps, error) {
	return model.Vps{
		Host: "—", Uptime: "sem servidor", Online: false, Warnings: 0,
		CPU: 0, RAM: 0, Disk: 0,
		CPUDetail: "cadastre um servidor em Ajustes", RAMDetail: "", DiskDetail: "",
		Containers: []model.Container{},
		Logs: []model.LogLine{
			{Time: "--:--:--", Level: "INFO", Body: "nenhum servidor SSH configurado"},
		},
	}, nil
}

func (stubVps) Restart(container string) (string, error) {
	return "", fmt.Errorf("nenhum servidor configurado")
}
