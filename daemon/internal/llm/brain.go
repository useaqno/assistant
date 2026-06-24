package llm

import (
	"context"
	"fmt"
	"strings"
	"time"

	"aqnod/internal/domain"
	"aqnod/internal/keychain"
	"aqnod/internal/model"
	"aqnod/internal/store"
)

// KeychainService namespaces LLM API keys in the OS keychain.
const KeychainService = "io.aqno.llm"

// Brain implements the assistant: it parses intents, executes the v1 tools
// against the store, and falls back to (or augments with) a configured LLM.
// It satisfies the httpapi.Brain interface structurally.
type Brain struct {
	st *store.Store
}

// NewBrain builds the assistant over the store.
func NewBrain(st *store.Store) *Brain { return &Brain{st: st} }

// Config resolves provider settings from config + keychain.
func (b *Brain) config() Config {
	provider := b.st.ConfigVal("llm.provider", "anthropic")
	cfg := Config{
		Provider:    provider,
		Model:       b.st.ConfigVal("llm.model", "claude-sonnet-4-6"),
		BaseURL:     b.st.ConfigVal("llm.base_url", ""),
		MaxTokens:   atoi(b.st.ConfigVal("llm.max_tokens", "2000"), 2000),
		Temperature: float32(atof(b.st.ConfigVal("llm.temperature", "0.4"), 0.4)),
	}
	if key, err := keychain.Get(KeychainService, provider); err == nil {
		cfg.APIKey = key
	}
	return cfg
}

// client returns a provider client when one is configured/usable.
func (b *Brain) client() (LLM, Config, bool) {
	cfg := b.config()
	c, ok := NewClient(cfg)
	return c, cfg, ok
}

// contextNames lists the user's context labels for intent matching.
func (b *Brain) contextNames() []string {
	ctxs, _ := b.st.Contexts()
	out := make([]string, 0, len(ctxs))
	for _, c := range ctxs {
		out = append(out, c.Label)
	}
	return out
}

// Chat handles a chat turn: act on commands, else answer (LLM or canned).
func (b *Brain) Chat(convID, userText string) (model.ChatMessage, error) {
	reply, ref := b.respond(userText, "")
	return b.persist(convID, reply, ref)
}

// Stream emits the reply token-by-token then persists it.
func (b *Brain) Stream(convID, userText string, emit func(string)) (model.ChatMessage, error) {
	reply, ref := b.respond(userText, "")
	for _, tok := range strings.SplitAfter(reply, " ") {
		emit(tok)
		time.Sleep(12 * time.Millisecond)
	}
	return b.persist(convID, reply, ref)
}

// Intent handles a voice command: parse, execute, log the interaction.
func (b *Brain) Intent(transcript, contextName string) (model.ChatMessage, error) {
	reply, ref := b.respond(transcript, contextName)
	intent := ""
	if ref != nil {
		intent = ref.Kind
	}
	_ = b.st.AddInteraction(contextName, transcript, intent, reply)
	return model.ChatMessage{From: "aqno", Text: reply, Time: time.Now().Format("15:04"), Ref: ref}, nil
}

// respond is the core engine shared by chat and voice.
func (b *Brain) respond(text, contextHint string) (string, *model.ChatRef) {
	action := ParseIntent(text, b.contextNames(), time.Now())
	if contextHint != "" && action.Context == "" {
		action.Context = contextHint
	}

	switch action.Tool {
	case "criar_evento", "criar_tarefa", "consultar_agenda", "consultar_vps", "registrar_nota":
		return b.execAction(action)
	}

	// Open conversation: use the LLM when configured, else a grounded canned reply.
	if c, cfg, ok := b.client(); ok {
		if reply, err := b.converse(c, cfg, text); err == nil && reply != "" {
			return reply, nil
		}
	}
	return b.cannedAnswer(text), nil
}

// execAction runs a tool and returns a human confirmation + UI ref.
func (b *Brain) execAction(a Action) (string, *model.ChatRef) {
	switch a.Tool {
	case "criar_evento":
		id, err := b.st.CreateEvent(a.Context, a.Title, eventTipo(a), a.Start, a.End, a.RRule, a.Date, a.Title)
		if err != nil {
			return "Não consegui criar o evento agora.", nil
		}
		_ = id
		when := scheduleLabel(a)
		ctxLabel := a.Context
		if ctxLabel == "" {
			ctxLabel = "pessoal"
		}
		return fmt.Sprintf("Criei \"%s\" %s.", a.Title, when),
			&model.ChatRef{Kind: "action", Label: fmt.Sprintf("Evento criado · %s · %s", a.Title, ctxLabel), Tone: "success"}
	case "criar_tarefa":
		if _, err := b.st.CreateTask(a.Context, a.Title, a.Title); err != nil {
			return "Não consegui criar a tarefa agora.", nil
		}
		return fmt.Sprintf("Anotei a tarefa \"%s\".", a.Title),
			&model.ChatRef{Kind: "action", Label: "Tarefa criada · " + a.Title, Tone: "success"}
	case "consultar_agenda":
		brief, err := domain.BuildToday(b.st, time.Now())
		if err != nil {
			return "Não consegui consultar sua agenda.", nil
		}
		return brief.Headline + " " + brief.Mentor,
			&model.ChatRef{Kind: "memory", Label: "agenda de hoje", Tone: ""}
	case "consultar_vps":
		srv, ok := b.st.FirstServer()
		if !ok {
			return "Você ainda não cadastrou um servidor. Adicione um em Ajustes para eu monitorar a VPS.", nil
		}
		return fmt.Sprintf("Servidor %s (%s) está cadastrado. Veja métricas ao vivo na tela de Infra.", srv.Name, srv.Host),
			&model.ChatRef{Kind: "memory", Label: "vps · " + srv.Name, Tone: ""}
	case "registrar_nota":
		_ = b.st.AddInteraction(a.Context, a.Body, "nota", "registrada")
		return "Anotei isso na memória.", &model.ChatRef{Kind: "memory", Label: "nota registrada", Tone: ""}
	}
	return "Não entendi o comando.", nil
}

// converse calls the LLM for open-ended chat, grounded with today's context.
func (b *Brain) converse(c LLM, cfg Config, text string) (string, error) {
	system := b.systemPrompt()
	req := CompletionRequest{
		System:      system,
		Messages:    []Message{{Role: RoleUser, Content: text}},
		Tools:       Tools(),
		MaxTokens:   cfg.MaxTokens,
		Temperature: cfg.Temperature,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	resp, err := c.Complete(ctx, req)
	if err != nil {
		return "", err
	}
	// Execute any tool the model requested.
	for _, tc := range resp.ToolCalls {
		if reply, ref := b.execToolCall(tc); reply != "" {
			_ = ref
			return reply, nil
		}
	}
	return resp.Text, nil
}

// execToolCall maps an LLM tool-call onto execAction.
func (b *Brain) execToolCall(tc ToolCall) (string, *model.ChatRef) {
	a := Action{Tool: tc.Name}
	a.Title = str(tc.Args["titulo"])
	a.Context = str(tc.Args["contexto"])
	a.Start = str(tc.Args["inicio"])
	a.End = str(tc.Args["fim"])
	a.RRule = str(tc.Args["rrule"])
	a.Date = str(tc.Args["data"])
	a.Body = str(tc.Args["texto"])
	return b.execAction(a)
}

func (b *Brain) systemPrompt() string {
	persona, _ := b.st.Persona()
	name := persona.Name
	if name == "" {
		name = "Aqno"
	}
	brief, _ := domain.BuildToday(b.st, time.Now())
	return fmt.Sprintf(
		"Você é %s, uma assistente pessoal de IA voice-first, em português do Brasil. Seja concisa, calorosa e prática. "+
			"Contexto do usuário hoje: %d reuniões, %d tarefas, %s de foco livre. %s",
		name, brief.Meetings, brief.Tasks, brief.FocusFree, brief.Mentor)
}

func (b *Brain) cannedAnswer(text string) string {
	brief, err := domain.BuildToday(b.st, time.Now())
	if err == nil && containsAny(fold(text), "obrigad", "valeu") {
		return "Por nada! Estou por aqui."
	}
	if err == nil {
		return brief.Headline + " Configure um provedor de IA em Ajustes para conversas mais ricas."
	}
	return "Configure um provedor de IA em Ajustes para eu responder melhor — já registro tudo localmente."
}

func (b *Brain) persist(convID, reply string, ref *model.ChatRef) (model.ChatMessage, error) {
	id, err := b.st.AddMessage(convID, "assistant", reply, ref)
	if err != nil {
		return model.ChatMessage{}, err
	}
	return model.ChatMessage{ID: id, From: "aqno", Text: reply, Time: time.Now().Format("15:04"), Ref: ref}, nil
}

// ---- small helpers ----

func eventTipo(a Action) string {
	if strings.Contains(fold(a.Title), "foco") || strings.Contains(fold(a.Title), "bloco") {
		return "bloco_foco"
	}
	return "reuniao"
}

func scheduleLabel(a Action) string {
	if a.RRule != "" {
		return fmt.Sprintf("recorrente às %s", a.Start)
	}
	if a.Date != "" {
		return fmt.Sprintf("em %s às %s", a.Date, a.Start)
	}
	return fmt.Sprintf("hoje às %s", a.Start)
}

func str(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func atoi(s string, def int) int {
	n := def
	fmt.Sscanf(s, "%d", &n)
	return n
}

func atof(s string, def float64) float64 {
	f := def
	fmt.Sscanf(s, "%g", &f)
	return f
}
