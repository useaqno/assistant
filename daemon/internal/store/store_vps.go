package store

import (
	"strconv"

	"aqnod/internal/model"
)

// Servers lists registered VPS hosts (without secrets).
func (s *Store) Servers() ([]model.Server, error) {
	rows, err := s.db.Query(
		`SELECT id, nome, host, porta, usuario, auth_tipo, keychain_ref FROM servidores ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Server
	for rows.Next() {
		var id int64
		var sv model.Server
		if err := rows.Scan(&id, &sv.Name, &sv.Host, &sv.Port, &sv.User, &sv.AuthType, &sv.KeychainRef); err != nil {
			return nil, err
		}
		sv.ID = strconv.FormatInt(id, 10)
		out = append(out, sv)
	}
	return out, rows.Err()
}

// ServerByID returns a single server.
func (s *Store) ServerByID(id string) (model.Server, bool) {
	var sv model.Server
	var nid int64
	err := s.db.QueryRow(
		`SELECT id, nome, host, porta, usuario, auth_tipo, keychain_ref FROM servidores WHERE id = ?`, id).
		Scan(&nid, &sv.Name, &sv.Host, &sv.Port, &sv.User, &sv.AuthType, &sv.KeychainRef)
	if err != nil {
		return model.Server{}, false
	}
	sv.ID = strconv.FormatInt(nid, 10)
	return sv, true
}

// FirstServer returns the primary registered server, if any.
func (s *Store) FirstServer() (model.Server, bool) {
	var sv model.Server
	var nid int64
	err := s.db.QueryRow(
		`SELECT id, nome, host, porta, usuario, auth_tipo, keychain_ref FROM servidores ORDER BY id LIMIT 1`).
		Scan(&nid, &sv.Name, &sv.Host, &sv.Port, &sv.User, &sv.AuthType, &sv.KeychainRef)
	if err != nil {
		return model.Server{}, false
	}
	sv.ID = strconv.FormatInt(nid, 10)
	return sv, true
}

// CreateServer registers a VPS; the secret is stored in the Keychain by the caller.
func (s *Store) CreateServer(name, host string, port int, user, authType, keychainRef string) (string, error) {
	if port == 0 {
		port = 22
	}
	res, err := s.db.Exec(
		`INSERT INTO servidores (nome, host, porta, usuario, auth_tipo, keychain_ref) VALUES (?, ?, ?, ?, ?, ?)`,
		name, host, port, user, def(authType, "senha"), keychainRef)
	if err != nil {
		return "", err
	}
	id, _ := res.LastInsertId()
	return strconv.FormatInt(id, 10), nil
}

// DeleteServer removes a server registration.
func (s *Store) DeleteServer(id string) error {
	_, err := s.db.Exec(`DELETE FROM servidores WHERE id = ?`, id)
	return err
}

// AddAudit records an auditable action (VPS / LLM boundary / config).
func (s *Store) AddAudit(category, action, detail string) error {
	_, err := s.db.Exec(
		`INSERT INTO auditoria (categoria, acao, detalhe) VALUES (?, ?, ?)`,
		category, action, nullStr(detail))
	return err
}

// AddInteraction logs a voice/chat interaction for history.
func (s *Store) AddInteraction(contextName, transcript, intent, result string) error {
	ctxID, _ := s.contextMeta(contextName)
	_, err := s.db.Exec(
		`INSERT INTO interacoes (contexto_id, transcricao, intencao, resultado) VALUES (?, ?, ?, ?)`,
		nullID(ctxID), transcript, nullStr(intent), nullStr(result))
	return err
}

// RecentInteractions returns the latest interactions for the Home screen.
func (s *Store) RecentInteractions(limit int) ([]model.Interaction, error) {
	rows, err := s.db.Query(`
		SELECT i.transcricao, COALESCE(i.resultado,''), COALESCE(c.nome,''), COALESCE(c.cor,''),
		       CAST((julianday('now') - julianday(i.criado_em)) * 24 * 60 AS INTEGER)
		FROM interacoes i LEFT JOIN contextos c ON c.id = i.contexto_id
		ORDER BY i.id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Interaction
	for rows.Next() {
		var transcript, result, ctx, color string
		var mins int
		if err := rows.Scan(&transcript, &result, &ctx, &color, &mins); err != nil {
			return nil, err
		}
		title := transcript
		if result != "" {
			title = "\"" + transcript + "\" — " + result
		}
		out = append(out, model.Interaction{
			Title:   title,
			Context: ctx,
			Color:   color,
			When:    humanAgo(mins),
		})
	}
	return out, rows.Err()
}

func humanAgo(mins int) string {
	switch {
	case mins < 1:
		return "agora"
	case mins < 60:
		return "há " + strconv.Itoa(mins) + " min"
	case mins < 60*24:
		return "há " + strconv.Itoa(mins/60) + "h"
	default:
		return "há " + strconv.Itoa(mins/(60*24)) + "d"
	}
}
