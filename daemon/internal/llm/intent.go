package llm

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Action is the structured result of parsing natural-language input. It mirrors
// the v1 tool set so the same shape drives both the heuristic and the LLM path.
type Action struct {
	Tool    string // criar_evento | criar_tarefa | consultar_agenda | consultar_vps | registrar_nota | conversa
	Title   string
	Context string
	Start   string // HH:MM
	End     string // HH:MM
	RRule   string
	Date    string // YYYY-MM-DD
	Body    string // free text (note)
}

var (
	reTimeColon = regexp.MustCompile(`(\d{1,2})[:h](\d{2})`)
	reTimeHour  = regexp.MustCompile(`(\d{1,2})\s*h\b`)
	reAsHour    = regexp.MustCompile(`(?:às|as|ás)\s*(\d{1,2})(?:\s*(?:horas|h))?\b`)
)

var weekdayWords = map[string]string{
	"segunda": "MO", "segundas": "MO",
	"terca": "TU", "tercas": "TU",
	"quarta": "WE", "quartas": "WE",
	"quinta": "TH", "quintas": "TH",
	"sexta": "FR", "sextas": "FR",
	"sabado": "SA", "sabados": "SA",
	"domingo": "SU", "domingos": "SU",
}

// fold lowercases and strips common Portuguese accents for matching.
func fold(s string) string {
	s = strings.ToLower(s)
	repl := strings.NewReplacer(
		"á", "a", "à", "a", "ã", "a", "â", "a",
		"é", "e", "ê", "e", "í", "i", "ó", "o", "ô", "o", "õ", "o",
		"ú", "u", "ç", "c")
	return repl.Replace(s)
}

// ParseIntent turns free text into a structured Action. `contexts` are the known
// context labels (for matching "na Cogna"). `now` anchors relative dates.
func ParseIntent(text string, contexts []string, now time.Time) Action {
	f := fold(text)

	// Pure queries first.
	switch {
	case containsAny(f, "como esta meu dia", "como ta meu dia", "meu dia", "minha agenda", "o que tenho hoje", "tenho hoje", "resumo do dia"):
		return Action{Tool: "consultar_agenda"}
	case containsAny(f, "vps", "servidor", "containers", "container", "infra"):
		return Action{Tool: "consultar_vps"}
	}

	ctx := matchContext(f, contexts)
	start, end := parseTimes(f)
	rrule := parseRecurrence(f)
	date := parseRelativeDate(f, now)
	title := extractTitle(text, ctx)

	isTask := containsAny(f, "tarefa", "lembrete", "lembra", "lembrar", "preciso", "anota", "anotar")
	isEvent := containsAny(f, "reuniao", "daily", "call", "evento", "bloco", "reuniao", "encontro", "compromisso") || start != "" || rrule != ""

	if isEvent && !isTask {
		if start == "" {
			start = "09:00"
		}
		if end == "" {
			end = addHour(start)
		}
		kind := ""
		if containsAny(f, "foco", "bloco") {
			kind = "bloco_foco"
		}
		_ = kind
		return Action{Tool: "criar_evento", Title: title, Context: ctx, Start: start, End: end, RRule: rrule, Date: date}
	}
	if isTask {
		return Action{Tool: "criar_tarefa", Title: title, Context: ctx}
	}
	// Fall back to conversation.
	return Action{Tool: "conversa", Body: text}
}

func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

func matchContext(folded string, contexts []string) string {
	for _, c := range contexts {
		if c == "" {
			continue
		}
		if strings.Contains(folded, fold(c)) {
			return c
		}
	}
	return ""
}

// parseTimes returns up to two HH:MM times found in the text (start, end).
func parseTimes(f string) (string, string) {
	var times []string
	for _, m := range reTimeColon.FindAllStringSubmatch(f, -1) {
		times = append(times, norm(m[1], m[2]))
	}
	if len(times) == 0 {
		for _, m := range reTimeHour.FindAllStringSubmatch(f, -1) {
			times = append(times, norm(m[1], "00"))
		}
	}
	if len(times) == 0 {
		for _, m := range reAsHour.FindAllStringSubmatch(f, -1) {
			times = append(times, norm(m[1], "00"))
		}
	}
	switch len(times) {
	case 0:
		return "", ""
	case 1:
		return times[0], ""
	default:
		return times[0], times[1]
	}
}

func norm(h, m string) string {
	hi, _ := strconv.Atoi(h)
	mi, _ := strconv.Atoi(m)
	if hi > 23 {
		hi = hi % 24
	}
	if mi > 59 {
		mi = 0
	}
	return fmt.Sprintf("%02d:%02d", hi, mi)
}

func addHour(hhmm string) string {
	parts := strings.SplitN(hhmm, ":", 2)
	if len(parts) != 2 {
		return hhmm
	}
	h, _ := strconv.Atoi(parts[0])
	m, _ := strconv.Atoi(parts[1])
	h = (h + 1) % 24
	return fmt.Sprintf("%02d:%02d", h, m)
}

func parseRecurrence(f string) string {
	if containsAny(f, "todo dia", "todos os dias", "diariamente", "daily", "dias uteis", "dia util", "recorrente") {
		return "FREQ=WEEKLY;BYDAY=MO,TU,WE,TH,FR"
	}
	if strings.Contains(f, "toda ") || strings.Contains(f, "todas ") || strings.Contains(f, "toda-") {
		for word, code := range weekdayWords {
			if strings.Contains(f, word) {
				return "FREQ=WEEKLY;BYDAY=" + code
			}
		}
	}
	if strings.Contains(f, "semanal") {
		return "FREQ=WEEKLY"
	}
	return ""
}

func parseRelativeDate(f string, now time.Time) string {
	switch {
	case strings.Contains(f, "depois de amanha"):
		return now.AddDate(0, 0, 2).Format("2006-01-02")
	case strings.Contains(f, "amanha"):
		return now.AddDate(0, 0, 1).Format("2006-01-02")
	case strings.Contains(f, "hoje"):
		return now.Format("2006-01-02")
	}
	return ""
}

// leadingPhrases are command prefixes stripped from the start of a title.
var leadingPhrases = []string{
	"por favor", "me lembra de", "me lembre de", "lembra de", "lembrar de",
	"lembrete de", "lembrete", "preciso de", "preciso",
	"tem uma nova", "tem um novo", "tem uma", "tem um", "tem",
	"cria", "criar", "marca", "marcar", "agenda", "agende", "adiciona", "adicionar",
}

// dropTokens are connective/recurrence words removed anywhere in the title
// (matched accent-folded). Meaningful prepositions like "de"/"pro" are kept.
var dropTokens = map[string]bool{
	"das": true, "as": true, "ate": true, "recorrente": true, "diariamente": true,
	"semanal": true, "semanalmente": true, "todo": true, "todos": true, "toda": true,
	"todas": true, "dia": true, "dias": true, "amanha": true, "hoje": true,
	"segunda": true, "segundas": true, "terca": true, "tercas": true,
	"quarta": true, "quartas": true, "quinta": true, "quintas": true,
	"sexta": true, "sextas": true, "sabado": true, "sabados": true,
	"domingo": true, "domingos": true,
}

// extractTitle derives a concise title by stripping wake words, the context
// mention, time/recurrence/date expressions and command filler.
func extractTitle(original, ctx string) string {
	t := original
	t = reTimeColon.ReplaceAllString(t, " ")
	t = reTimeHour.ReplaceAllString(t, " ")
	t = reAsHour.ReplaceAllString(t, " ")
	t = regexp.MustCompile(`(?i)^\s*(aqno|iris|íris)[,\s]+`).ReplaceAllString(t, "")

	// drop context mention with common prepositions
	if ctx != "" {
		for _, pre := range []string{"na ", "no ", "pra ", "para ", "da ", "do ", "com a ", "com o ", "com ", ""} {
			re := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(pre+ctx) + `\b`)
			t = re.ReplaceAllString(t, " ")
		}
	}

	// strip a leading command phrase
	low := fold(strings.TrimSpace(t))
	for _, p := range leadingPhrases {
		if strings.HasPrefix(low, p+" ") || low == p {
			t = strings.TrimSpace(t)[len(p):]
			break
		}
	}

	// word-level cleanup of connective/recurrence residue
	var kept []string
	for _, w := range strings.Fields(t) {
		clean := strings.Trim(w, ",.;:-")
		if clean == "" || dropTokens[fold(clean)] {
			continue
		}
		kept = append(kept, clean)
	}
	result := strings.Join(kept, " ")
	if result == "" {
		return "Novo item"
	}
	r := []rune(result)
	r[0] = []rune(strings.ToUpper(string(r[0])))[0]
	return string(r)
}
