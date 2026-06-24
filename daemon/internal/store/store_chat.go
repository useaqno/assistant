package store

import (
	"database/sql"
	"strconv"

	"aqnod/internal/model"
)

// Conversation is a chat thread header.
type Conversation struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Latest string `json:"latest"`
}

// Conversations lists threads newest first.
func (s *Store) Conversations() ([]Conversation, error) {
	rows, err := s.db.Query(
		`SELECT id, COALESCE(titulo,'Conversa') FROM conversas ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Conversation
	for rows.Next() {
		var id int64
		var c Conversation
		if err := rows.Scan(&id, &c.Title); err != nil {
			return nil, err
		}
		c.ID = strconv.FormatInt(id, 10)
		out = append(out, c)
	}
	return out, rows.Err()
}

// EnsureConversation returns the latest conversation id, creating one if none.
func (s *Store) EnsureConversation() (string, error) {
	var id int64
	err := s.db.QueryRow(`SELECT id FROM conversas ORDER BY id DESC LIMIT 1`).Scan(&id)
	if err == sql.ErrNoRows {
		return s.CreateConversation("Conversa com a Íris")
	}
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(id, 10), nil
}

// CreateConversation opens a new thread.
func (s *Store) CreateConversation(title string) (string, error) {
	res, err := s.db.Exec(`INSERT INTO conversas (titulo) VALUES (?)`, title)
	if err != nil {
		return "", err
	}
	id, _ := res.LastInsertId()
	return strconv.FormatInt(id, 10), nil
}

// Messages returns the turns of a conversation oldest-first.
func (s *Store) Messages(convID string) ([]model.ChatMessage, error) {
	rows, err := s.db.Query(`
		SELECT id, papel, conteudo, COALESCE(ref_kind,''), COALESCE(ref_label,''), COALESCE(ref_tone,''),
		       strftime('%H:%M', criado_em, 'localtime')
		FROM mensagens WHERE conversa_id = ? ORDER BY id`, convID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.ChatMessage
	for rows.Next() {
		var id int64
		var papel, conteudo, rk, rl, rt, t string
		if err := rows.Scan(&id, &papel, &conteudo, &rk, &rl, &rt, &t); err != nil {
			return nil, err
		}
		m := model.ChatMessage{
			ID:   strconv.FormatInt(id, 10),
			From: from(papel),
			Text: conteudo,
			Time: t,
		}
		if rk != "" {
			m.Ref = &model.ChatRef{Kind: rk, Label: rl, Tone: rt}
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// AddMessage appends a turn and returns its id.
func (s *Store) AddMessage(convID, role, content string, ref *model.ChatRef) (string, error) {
	var rk, rl, rt any
	if ref != nil {
		rk, rl, rt = ref.Kind, ref.Label, ref.Tone
	}
	res, err := s.db.Exec(
		`INSERT INTO mensagens (conversa_id, papel, conteudo, ref_kind, ref_label, ref_tone)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		convID, role, content, rk, rl, rt)
	if err != nil {
		return "", err
	}
	id, _ := res.LastInsertId()
	return strconv.FormatInt(id, 10), nil
}

func from(papel string) string {
	if papel == "user" {
		return "user"
	}
	return "aqno"
}
