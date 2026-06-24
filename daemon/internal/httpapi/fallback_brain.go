package httpapi

import (
	"encoding/json"
	"strings"
	"time"

	"aqnod/internal/model"
	"aqnod/internal/store"
)

// fallbackBrain is the minimal offline assistant used until the LLM layer (WS3)
// is injected. It persists a helpful canned reply so the chat stays usable with
// no provider configured.
type fallbackBrain struct{ st *store.Store }

func (b fallbackBrain) reply(text string) string {
	t := strings.ToLower(text)
	switch {
	case strings.Contains(t, "dia") || strings.Contains(t, "agenda"):
		return "Posso te dar o resumo do dia assim que um provedor de IA estiver configurado em Ajustes. Por enquanto, confira a Agenda e a Análise."
	case strings.Contains(t, "vps") || strings.Contains(t, "servidor"):
		return "Cadastre um servidor em Ajustes para eu monitorar a VPS por SSH."
	default:
		return "Entendi. Configure um provedor de IA em Ajustes para respostas completas — eu já registro tudo localmente."
	}
}

func (b fallbackBrain) persist(convID, text string) (model.ChatMessage, error) {
	r := b.reply(text)
	id, err := b.st.AddMessage(convID, "assistant", r, nil)
	if err != nil {
		return model.ChatMessage{}, err
	}
	return model.ChatMessage{ID: id, From: "aqno", Text: r, Time: time.Now().Format("15:04")}, nil
}

func (b fallbackBrain) Chat(convID, text string) (model.ChatMessage, error) {
	return b.persist(convID, text)
}

func (b fallbackBrain) Stream(convID, text string, emit func(string)) (model.ChatMessage, error) {
	r := b.reply(text)
	for _, word := range strings.Fields(r) {
		emit(word + " ")
	}
	id, err := b.st.AddMessage(convID, "assistant", r, nil)
	if err != nil {
		return model.ChatMessage{}, err
	}
	return model.ChatMessage{ID: id, From: "aqno", Text: r, Time: time.Now().Format("15:04")}, nil
}

func (b fallbackBrain) Intent(transcript, contextName string) (model.ChatMessage, error) {
	_ = b.st.AddInteraction(contextName, transcript, "", "registrado")
	return model.ChatMessage{
		From: "aqno",
		Text: "Registrei o que você falou. Configure um provedor de IA em Ajustes para eu executar comandos por voz.",
		Time: time.Now().Format("15:04"),
	}, nil
}

// ---- json helpers (shared by streaming handlers) ----

func jsonString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func mustJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
