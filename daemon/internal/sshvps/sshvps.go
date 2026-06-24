// Package sshvps is the VPS connector: it dials the registered server over SSH
// (credentials from the Keychain), gathers read-only metrics, and performs
// confirmed actions (container restart). See docs/context.md §14, §17.
package sshvps

import (
	"fmt"
	"net"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"

	"aqnod/internal/keychain"
	"aqnod/internal/model"
	"aqnod/internal/store"
)

// KeychainService namespaces SSH credentials in the OS keychain.
const KeychainService = "io.aqno.ssh"

// Provider implements httpapi.VpsProvider against a real SSH host.
type Provider struct {
	st *store.Store
}

// New builds the connector over the store.
func New(st *store.Store) *Provider { return &Provider{st: st} }

// dial opens an SSH session to the primary registered server.
func (p *Provider) dial() (*ssh.Client, model.Server, error) {
	srv, ok := p.st.FirstServer()
	if !ok {
		return nil, model.Server{}, errNoServer
	}
	secret, err := keychain.Get(KeychainService, srv.KeychainRef)
	if err != nil {
		return nil, srv, fmt.Errorf("credencial ausente no Keychain (%s)", srv.KeychainRef)
	}

	var auth ssh.AuthMethod
	if srv.AuthType == "chave" {
		signer, perr := ssh.ParsePrivateKey([]byte(secret))
		if perr != nil {
			return nil, srv, fmt.Errorf("chave privada inválida: %w", perr)
		}
		auth = ssh.PublicKeys(signer)
	} else {
		auth = ssh.Password(secret)
	}

	cfg := &ssh.ClientConfig{
		User:            srv.User,
		Auth:            []ssh.AuthMethod{auth},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // local-first: user's own servers (v1)
		Timeout:         8 * time.Second,
	}
	addr := net.JoinHostPort(srv.Host, fmt.Sprintf("%d", srv.Port))
	client, err := ssh.Dial("tcp", addr, cfg)
	if err != nil {
		return nil, srv, fmt.Errorf("falha ao conectar em %s: %w", addr, err)
	}
	return client, srv, nil
}

// run executes a command and returns combined stdout.
func run(client *ssh.Client, cmd string) (string, error) {
	sess, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer sess.Close()
	out, err := sess.CombinedOutput(cmd)
	return string(out), err
}

// Snapshot connects, gathers metrics and returns the infra payload. When no
// server is registered it returns an offline placeholder (handled by the UI).
func (p *Provider) Snapshot() (model.Vps, error) {
	client, srv, err := p.dial()
	if err != nil {
		if err == errNoServer {
			return offline(), nil
		}
		return offlineErr(srv, err), nil
	}
	defer client.Close()

	raw, _ := run(client, metricsScript)
	v := parseMetrics(raw)
	v.Host = srv.User + "@" + srv.Host
	v.Online = true
	return v, nil
}

// Restart restarts a docker container (confirmation is enforced by the handler).
func (p *Provider) Restart(container string) (string, error) {
	client, _, err := p.dial()
	if err != nil {
		return "", err
	}
	defer client.Close()
	out, err := run(client, "docker restart "+shellQuote(container))
	if err != nil {
		return "", fmt.Errorf("docker restart falhou: %s", strings.TrimSpace(out))
	}
	return fmt.Sprintf("Container reiniciado · %s", container), nil
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func offline() model.Vps {
	return model.Vps{
		Host: "—", Uptime: "sem servidor", Online: false,
		CPUDetail: "cadastre um servidor em Ajustes",
		Containers: []model.Container{},
		Logs:       []model.LogLine{{Time: "--:--:--", Level: "INFO", Body: "nenhum servidor SSH configurado"}},
	}
}

func offlineErr(srv model.Server, err error) model.Vps {
	return model.Vps{
		Host: srv.User + "@" + srv.Host, Uptime: "offline", Online: false, Warnings: 1,
		CPUDetail: "conexão indisponível",
		Containers: []model.Container{},
		Logs:       []model.LogLine{{Time: time.Now().Format("15:04:05"), Level: "WARN", Body: err.Error()}},
	}
}

var errNoServer = fmt.Errorf("no server registered")
