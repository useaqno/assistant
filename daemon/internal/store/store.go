package store

import (
	"database/sql"
	"strconv"
	"time"

	"aqnod/internal/model"
)

func todayDate() string { return time.Now().Format("2006-01-02") }

// kindFromTipo maps a DB event type to the UI's kind enum.
func kindFromTipo(tipo string) string {
	switch tipo {
	case "bloco_foco":
		return "focus"
	case "pessoal":
		return "personal"
	default:
		return "event"
	}
}

// ===== Config =====

// Config returns every configuration key/value.
func (s *Store) Config() (map[string]string, error) {
	rows, err := s.db.Query(`SELECT chave, valor FROM config`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]string{}
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		out[k] = v
	}
	return out, rows.Err()
}

// ConfigVal returns one key (def if missing).
func (s *Store) ConfigVal(key, def string) string {
	var v string
	if err := s.db.QueryRow(`SELECT valor FROM config WHERE chave = ?`, key).Scan(&v); err != nil {
		return def
	}
	return v
}

// SetConfig upserts a configuration value.
func (s *Store) SetConfig(key, val string) error {
	_, err := s.db.Exec(
		`INSERT INTO config (chave, valor, atualizado_em) VALUES (?, ?, datetime('now'))
		 ON CONFLICT(chave) DO UPDATE SET valor = excluded.valor, atualizado_em = datetime('now')`,
		key, val)
	return err
}

// Onboarded reports whether first-run setup is complete.
func (s *Store) Onboarded() bool { return s.ConfigVal("onboarding.completed", "false") == "true" }

// ===== Persona =====

// Persona returns the saved persona and whether one exists.
func (s *Store) Persona() (model.Persona, bool) {
	var p model.Persona
	var avatarRef, voice, usuario sql.NullString
	err := s.db.QueryRow(
		`SELECT nome, tipo_avatar, avatar_ref, cor_aura, voz, tom, wake_word, usuario
		 FROM persona ORDER BY id LIMIT 1`).
		Scan(&p.Name, &p.Avatar, &avatarRef, &p.AuraColor, &voice, &p.Tone, &p.WakeWord, &usuario)
	if err != nil {
		return model.Persona{}, false
	}
	p.AvatarRef = avatarRef.String
	p.Voice = voice.String
	p.Owner = usuario.String
	return p, true
}

// SavePersona replaces the single persona row.
func (s *Store) SavePersona(p model.Persona) error {
	if _, err := s.db.Exec(`DELETE FROM persona`); err != nil {
		return err
	}
	_, err := s.db.Exec(
		`INSERT INTO persona (nome, tipo_avatar, avatar_ref, cor_aura, voz, tom, wake_word, usuario)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		p.Name, p.Avatar, nullStr(p.AvatarRef), p.AuraColor, nullStr(p.Voice),
		def(p.Tone, "amigavel"), def(p.WakeWord, "aqno"), p.Owner)
	return err
}

// ===== Contexts =====

// Contexts lists non-archived contexts.
func (s *Store) Contexts() ([]model.Context, error) {
	rows, err := s.db.Query(
		`SELECT id, nome, cor, ai_mode, arquivado FROM contextos ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Context
	for rows.Next() {
		var id int64
		var c model.Context
		var arch int
		if err := rows.Scan(&id, &c.Label, &c.Color, &c.AIMode, &arch); err != nil {
			return nil, err
		}
		c.ID = slug(c.Label)
		c.Archived = arch == 1
		out = append(out, c)
	}
	return out, rows.Err()
}

// CreateContext inserts a context and returns it.
func (s *Store) CreateContext(label, color, aiMode string) (model.Context, error) {
	res, err := s.db.Exec(
		`INSERT INTO contextos (nome, cor, ai_mode) VALUES (?, ?, ?)`,
		label, def(color, "violet"), def(aiMode, "cloud"))
	if err != nil {
		return model.Context{}, err
	}
	id, _ := res.LastInsertId()
	_ = id
	return model.Context{ID: slug(label), Label: label, Color: def(color, "violet"), AIMode: def(aiMode, "cloud")}, nil
}

// SetContextAIMode updates a context's privacy routing.
func (s *Store) SetContextAIMode(label, mode string) error {
	_, err := s.db.Exec(`UPDATE contextos SET ai_mode = ? WHERE nome = ?`, mode, label)
	return err
}

// contextMeta resolves a context name to its color + id for event/task joins.
func (s *Store) contextMeta(name string) (id int64, color string) {
	s.db.QueryRow(`SELECT id, cor FROM contextos WHERE nome = ?`, name).Scan(&id, &color)
	return
}

// ===== Events =====

// Exception is a per-occurrence override.
type Exception struct {
	Date       string
	Type       string
	NewStart   string
	NewEnd     string
}

// RawEvent is a stored event (recurring or single) before expansion.
type RawEvent struct {
	model.Event
	DataUnica string
}

// RawEvents returns all active events joined with their context label/color.
func (s *Store) RawEvents() ([]RawEvent, error) {
	rows, err := s.db.Query(`
		SELECT e.id, e.titulo, e.tipo, e.inicio, COALESCE(e.fim,''), COALESCE(e.rrule,''),
		       COALESCE(e.data_unica,''), COALESCE(c.nome,''), COALESCE(c.cor,''),
		       COALESCE(e.origem_voz,''), e.contexto_id
		FROM eventos e LEFT JOIN contextos c ON c.id = e.contexto_id
		WHERE e.ativo = 1 ORDER BY e.inicio`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []RawEvent
	for rows.Next() {
		var r RawEvent
		var id int64
		var tipo string
		var ctxID sql.NullInt64
		if err := rows.Scan(&id, &r.Title, &tipo, &r.Start, &r.End, &r.RRule,
			&r.DataUnica, &r.Context, &r.Color, &r.OriginVoice, &ctxID); err != nil {
			return nil, err
		}
		r.ID = strconv.FormatInt(id, 10)
		r.Kind = kindFromTipo(tipo)
		if ctxID.Valid {
			r.ContextID = slug(r.Context)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// ExceptionsFor returns the overrides for an event id.
func (s *Store) ExceptionsFor(eventID string) ([]Exception, error) {
	rows, err := s.db.Query(
		`SELECT data, tipo, COALESCE(novo_inicio,''), COALESCE(novo_fim,'') FROM excecoes WHERE evento_id = ?`,
		eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Exception
	for rows.Next() {
		var e Exception
		if err := rows.Scan(&e.Date, &e.Type, &e.NewStart, &e.NewEnd); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// CreateEvent inserts an event (recurring if rrule set, single if date set).
func (s *Store) CreateEvent(contextName, title, tipo, start, end, rrule, date, originVoice string) (string, error) {
	ctxID, _ := s.contextMeta(contextName)
	res, err := s.db.Exec(
		`INSERT INTO eventos (contexto_id, titulo, tipo, inicio, fim, rrule, data_unica, origem_voz)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		nullID(ctxID), title, def(tipo, "reuniao"), start, nullStr(end),
		nullStr(rrule), nullStr(date), nullStr(originVoice))
	if err != nil {
		return "", err
	}
	id, _ := res.LastInsertId()
	return strconv.FormatInt(id, 10), nil
}

// UpdateEvent edits core fields of an event.
func (s *Store) UpdateEvent(id, contextName, title, tipo, start, end, rrule, date string) error {
	ctxID, _ := s.contextMeta(contextName)
	_, err := s.db.Exec(
		`UPDATE eventos SET contexto_id=?, titulo=?, tipo=?, inicio=?, fim=?, rrule=?, data_unica=? WHERE id=?`,
		nullID(ctxID), title, def(tipo, "reuniao"), start, nullStr(end), nullStr(rrule), nullStr(date), id)
	return err
}

// DeleteEvent removes an event entirely.
func (s *Store) DeleteEvent(id string) error {
	_, err := s.db.Exec(`DELETE FROM eventos WHERE id = ?`, id)
	return err
}

// CancelOccurrence marks a single occurrence of a recurring event cancelled.
func (s *Store) CancelOccurrence(eventID, date string) error {
	_, err := s.db.Exec(
		`INSERT INTO excecoes (evento_id, data, tipo) VALUES (?, ?, 'cancelado')`,
		eventID, date)
	return err
}

// ===== Tasks =====

// Tasks lists tasks (most recent first), newest open before completed.
func (s *Store) Tasks() ([]model.Task, error) {
	rows, err := s.db.Query(`
		SELECT t.id, t.titulo, t.status, t.prioridade, COALESCE(t.prazo,''), COALESCE(c.nome,'')
		FROM tarefas t LEFT JOIN contextos c ON c.id = t.contexto_id
		ORDER BY (t.status='concluida'), t.prioridade DESC, t.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Task
	for rows.Next() {
		var id int64
		var t model.Task
		if err := rows.Scan(&id, &t.Title, &t.Status, &t.Priority, &t.Deadline, &t.Context); err != nil {
			return nil, err
		}
		t.ID = strconv.FormatInt(id, 10)
		t.Done = t.Status == "concluida"
		if t.Context != "" {
			t.ContextID = slug(t.Context)
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// CreateTask inserts a task.
func (s *Store) CreateTask(contextName, title, originVoice string) (string, error) {
	ctxID, _ := s.contextMeta(contextName)
	res, err := s.db.Exec(
		`INSERT INTO tarefas (contexto_id, titulo, origem_voz) VALUES (?, ?, ?)`,
		nullID(ctxID), title, nullStr(originVoice))
	if err != nil {
		return "", err
	}
	id, _ := res.LastInsertId()
	return strconv.FormatInt(id, 10), nil
}

// SetTaskDone toggles completion.
func (s *Store) SetTaskDone(id string, done bool) error {
	if done {
		_, err := s.db.Exec(
			`UPDATE tarefas SET status='concluida', concluido_em=datetime('now') WHERE id=?`, id)
		return err
	}
	_, err := s.db.Exec(`UPDATE tarefas SET status='aberta', concluido_em=NULL WHERE id=?`, id)
	return err
}

// DeleteTask removes a task.
func (s *Store) DeleteTask(id string) error {
	_, err := s.db.Exec(`DELETE FROM tarefas WHERE id = ?`, id)
	return err
}

// ===== small helpers =====

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func nullID(id int64) any {
	if id == 0 {
		return nil
	}
	return id
}

func def(v, d string) string {
	if v == "" {
		return d
	}
	return v
}
