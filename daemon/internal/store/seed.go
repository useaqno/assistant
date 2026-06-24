package store

// seed populates first-run defaults: configuration, the user's contexts, and a
// small set of demo events/tasks/graph nodes so the app feels alive before the
// user adds their own. Every step is idempotent (guarded by COUNT checks).

// DefaultConfig holds the out-of-the-box configuration keys (docs/context.md §6.2, §8).
var DefaultConfig = map[string]string{
	"onboarding.completed": "false",
	"llm.provider":         "anthropic",
	"llm.model":            "claude-sonnet-4-6",
	"llm.base_url":         "",
	"llm.max_tokens":       "2000",
	"llm.temperature":      "0.4",
	"llm.local_provider":   "ollama",
	"llm.local_model":      "llama3.1",
	"llm.embeddings_model": "nomic-embed-text",
	"voice.wake_word":      "aqno",
	"voice.activation":     "ambos",
	"voice.ptt_hotkey":     "Alt+Space",
	"voice.stt_lang":       "pt",
	"voice.stt_engine":     "whisper",
	"voice.model_tier":     "small",
	"voice.tts_voice":      "Luciana",
	"voice.speed":          "1.0",
	"voice.tone":           "amigavel",
	"voice.confirm":        "destrutivas",
	"voice.stt_privacy":    "local",
	"voice.feedback_sound": "true",
}

type seedContext struct {
	nome, cor, aiMode string
}

var seedContexts = []seedContext{
	{"Cogna", "violet", "cloud"},
	{"Bayer", "teal", "local_only"},
	{"Visa", "amber", "local_only"},
	{"Devlith", "rose", "cloud"},
	{"Pitrace", "blue", "cloud"},
	{"Pessoal", "violet", "cloud"},
}

func (s *Store) seed() error {
	if err := s.seedConfig(); err != nil {
		return err
	}
	if err := s.seedContexts(); err != nil {
		return err
	}
	if err := s.seedDemo(); err != nil {
		return err
	}
	return s.seedGraph()
}

func (s *Store) seedConfig() error {
	for k, v := range DefaultConfig {
		if _, err := s.db.Exec(
			`INSERT INTO config (chave, valor) VALUES (?, ?) ON CONFLICT(chave) DO NOTHING`,
			k, v); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) seedContexts() error {
	var n int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM contextos`).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	for _, c := range seedContexts {
		if _, err := s.db.Exec(
			`INSERT INTO contextos (nome, cor, ai_mode) VALUES (?, ?, ?)`,
			c.nome, c.cor, c.aiMode); err != nil {
			return err
		}
	}
	return nil
}

// ctxID resolves a context name to its id (0 if absent).
func (s *Store) ctxID(name string) int64 {
	var id int64
	s.db.QueryRow(`SELECT id FROM contextos WHERE nome = ?`, name).Scan(&id)
	return id
}

func (s *Store) seedDemo() error {
	var n int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM eventos`).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil
	}

	// Recurring + single events. inicio/fim are 'HH:MM'; recurring carry an
	// rrule, singles carry data_unica = today.
	type ev struct {
		ctx, titulo, tipo, inicio, fim, rrule string
		single                                bool
	}
	demo := []ev{
		{"Cogna", "Daily da Cogna", "reuniao", "09:30", "10:00", "FREQ=WEEKLY;BYDAY=MO,TU,WE,TH,FR", false},
		{"Visa", "Bloco de foco · Proposta Visa", "bloco_foco", "11:00", "12:30", "", true},
		{"Visa", "Call · proposta Visa", "reuniao", "14:00", "15:00", "", true},
		{"Bayer", "Bayer · revisão dossiê", "reuniao", "14:00", "14:30", "", true},
		{"Pessoal", "Ligar pro contador", "pessoal", "16:30", "17:00", "", true},
	}
	for _, e := range demo {
		var dataUnica, rrule any
		if e.single {
			dataUnica = todayDate()
		}
		if e.rrule != "" {
			rrule = e.rrule
		}
		if _, err := s.db.Exec(
			`INSERT INTO eventos (contexto_id, titulo, tipo, inicio, fim, rrule, data_unica)
			 VALUES (?, ?, ?, ?, ?, ?, ?)`,
			s.ctxID(e.ctx), e.titulo, e.tipo, e.inicio, e.fim, rrule, dataUnica); err != nil {
			return err
		}
	}

	type tk struct {
		ctx, titulo, status string
	}
	tasks := []tk{
		{"Cogna", "Enviar proposta Q3", "concluida"},
		{"Visa", "Revisar proposta Visa", "concluida"},
		{"Bayer", "Revisar dossiê Bayer", "aberta"},
		{"Pessoal", "Ligar pro contador", "aberta"},
		{"Devlith", "Corrigir bug de pagamento", "aberta"},
	}
	for _, t := range tasks {
		concl := any(nil)
		if t.status == "concluida" {
			concl = todayDate()
		}
		if _, err := s.db.Exec(
			`INSERT INTO tarefas (contexto_id, titulo, status, concluido_em) VALUES (?, ?, ?, ?)`,
			s.ctxID(t.ctx), t.titulo, t.status, concl); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) seedGraph() error {
	var n int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM entidades`).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	// Entities: each context becomes a node, plus a few projects/people/tasks.
	type ent struct {
		tipo, rotulo, ctx string
	}
	ents := []ent{
		{"projeto", "Proposta Q3", "Cogna"},
		{"pessoa", "Marina · PM", "Cogna"},
		{"evento", "Daily 9:30", "Cogna"},
		{"projeto", "Dossiê regulatório", "Bayer"},
		{"tarefa", "Auditoria", "Bayer"},
		{"pessoa", "Dr. Klein", "Bayer"},
		{"projeto", "Integração API", "Visa"},
		{"decisao", "Decisão: tokenização", "Visa"},
		{"projeto", "Sprint 14", "Devlith"},
		{"tarefa", "Bug pagamento", "Devlith"},
		{"evento", "Deploy v2", "Pitrace"},
	}
	// Context nodes first.
	ctxEnt := map[string]int64{}
	for _, c := range seedContexts {
		res, err := s.db.Exec(
			`INSERT INTO entidades (tipo, rotulo, contexto_id) VALUES ('context', ?, ?)`,
			c.nome, s.ctxID(c.nome))
		if err != nil {
			return err
		}
		id, _ := res.LastInsertId()
		ctxEnt[c.nome] = id
	}
	for _, e := range ents {
		res, err := s.db.Exec(
			`INSERT INTO entidades (tipo, rotulo, contexto_id) VALUES (?, ?, ?)`,
			e.tipo, e.rotulo, s.ctxID(e.ctx))
		if err != nil {
			return err
		}
		id, _ := res.LastInsertId()
		if parent, ok := ctxEnt[e.ctx]; ok {
			if _, err := s.db.Exec(
				`INSERT INTO relacoes (origem_id, destino_id, tipo) VALUES (?, ?, 'pertence_a')`,
				id, parent); err != nil {
				return err
			}
		}
	}
	return nil
}
