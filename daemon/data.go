package main

// Mock domain data for the Aqno companion. In a real build this layer would be
// backed by SQLite + the on-device models; here it returns plausible, stable
// fixtures so the whole UI is wired end-to-end.

type Context struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Color string `json:"color"` // data-palette key: violet|teal|amber|rose|blue
}

type Interaction struct {
	Title   string `json:"title"`
	Context string `json:"context"`
	Color   string `json:"color"`
	Tag     string `json:"tag"`
	Tone    string `json:"tone"` // badge tone
	When    string `json:"when"`
}

type Task struct {
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

type TodayBrief struct {
	Greeting  string        `json:"greeting"`
	Date      string        `json:"date"`
	Companion string        `json:"companion"`
	State     string        `json:"state"`
	Headline  string        `json:"headline"`
	Meetings  int           `json:"meetings"`
	Tasks     int           `json:"tasks"`
	FocusFree string        `json:"focusFree"`
	NextEvent Event         `json:"nextEvent"`
	TaskList  []Task        `json:"taskList"`
	Mentor    string        `json:"mentor"`
	Recent    []Interaction `json:"recent"`
}

type Event struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Context  string `json:"context"`
	Color    string `json:"color"`
	Start    string `json:"start"`
	End      string `json:"end"`
	StartMin int    `json:"startMin"` // minutes from 00:00, for the timeline
	EndMin   int    `json:"endMin"`
	Kind     string `json:"kind"` // event|focus|personal
	Conflict bool   `json:"conflict"`
}

type Agenda struct {
	Day       string  `json:"day"`
	Conflicts int     `json:"conflicts"`
	Focus     int     `json:"focus"`
	Events    []Event `json:"events"`
}

type AppHealth struct {
	Name    string `json:"name"`
	Status  string `json:"status"` // ok|warn|down
	Latency string `json:"latency"`
	Spark   []int  `json:"spark"`
}

type Analysis struct {
	Summary     string      `json:"summary"`
	Meetings    int         `json:"meetings"`
	Tasks       int         `json:"tasks"`
	FocusFree   string      `json:"focusFree"`
	Contexts    int         `json:"contexts"`
	FocusShare  float64     `json:"focusShare"`
	TasksDone   string      `json:"tasksDone"`
	TasksRatio  float64     `json:"tasksRatio"`
	Apps        []AppHealth `json:"apps"`
	CashMonth   string      `json:"cashMonth"`
	CashDelta   string      `json:"cashDelta"`
	CashBars    []int       `json:"cashBars"`
	Personal    []Metric    `json:"personal"`
	MentorTitle string      `json:"mentorTitle"`
	MentorBody  string      `json:"mentorBody"`
}

type Metric struct {
	Label string  `json:"label"`
	Value float64 `json:"value"`
	Big   string  `json:"big"`
	Color string  `json:"color"`
}

type Container struct {
	Name   string `json:"name"`
	Image  string `json:"image"`
	Status string `json:"status"` // running|restarting|down
	CPU    string `json:"cpu"`
	Mem    string `json:"mem"`
}

type LogLine struct {
	Time  string `json:"time"`
	Level string `json:"level"` // INFO|WARN|OK|CMD
	Body  string `json:"body"`
}

type Vps struct {
	Host       string      `json:"host"`
	Uptime     string      `json:"uptime"`
	Online     bool        `json:"online"`
	Warnings   int         `json:"warnings"`
	CPU        float64     `json:"cpu"`
	RAM        float64     `json:"ram"`
	Disk       float64     `json:"disk"`
	CPUDetail  string      `json:"cpuDetail"`
	RAMDetail  string      `json:"ramDetail"`
	DiskDetail string      `json:"diskDetail"`
	Containers []Container `json:"containers"`
	Logs       []LogLine   `json:"logs"`
}

type ChatRef struct {
	Kind  string `json:"kind"` // memory|action
	Label string `json:"label"`
	Tone  string `json:"tone"`
}

type ChatMessage struct {
	ID        string  `json:"id"`
	From      string  `json:"from"` // user|aqno
	Text      string  `json:"text"`
	Time      string  `json:"time"`
	Streaming bool    `json:"streaming"`
	Ref       *ChatRef `json:"ref,omitempty"`
}

type GraphNode struct {
	ID    string  `json:"id"`
	Label string  `json:"label"`
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	Color string  `json:"color"`
	Kind  string  `json:"kind"`
	Size  int     `json:"size,omitempty"`
}

type GraphEdge struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type Graph struct {
	Nodes []GraphNode `json:"nodes"`
	Edges []GraphEdge `json:"edges"`
}

var contexts = []Context{
	{"cogna", "Cogna", "violet"},
	{"bayer", "Bayer", "teal"},
	{"visa", "Visa", "amber"},
	{"devlith", "Devlith", "rose"},
	{"pitrace", "Pitrace", "blue"},
}

func todayBrief() TodayBrief {
	return TodayBrief{
		Greeting:  "Bom dia, Renato.",
		Date:      "Segunda · 23 jun · 09:12",
		Companion: "Íris",
		State:     "listening",
		Headline:  "Você tem 4 reuniões e 3 tarefas hoje. Quer que eu prepare um resumo?",
		Meetings:  4,
		Tasks:     3,
		FocusFree: "3h",
		NextEvent: Event{Title: "Daily da Cogna", Context: "Cogna", Color: "violet", Start: "09:30", End: "10:00"},
		TaskList: []Task{
			{"Enviar proposta Q3", true},
			{"Revisar dossiê Bayer", false},
			{"Ligar pro contador", false},
		},
		Mentor: "Seu bloco de foco das 11h está livre. Proteja-o para avançar a proposta da Visa antes da call das 14h.",
		Recent: []Interaction{
			{"\"Resuma a call da Cogna\" — 3 itens de ação criados", "Cogna", "violet", "", "", "há 12 min"},
			{"\"Lembra de renovar o domínio aqno.io\" — lembrete às 13:30", "", "", "9 dias", "warning", "há 1h"},
			{"\"Como está o VPS?\" — 4 containers ok, RAM em 72%", "", "", "saudável", "success", "há 2h"},
		},
	}
}

func agenda() Agenda {
	return Agenda{
		Day:       "Segunda · 23 jun · 2026",
		Conflicts: 1,
		Focus:     1,
		Events: []Event{
			{ID: "e1", Title: "Daily da Cogna", Context: "Cogna", Color: "violet", Start: "09:30", End: "10:00", StartMin: 570, EndMin: 600, Kind: "event"},
			{ID: "e2", Title: "Bloco de foco · Proposta Visa", Context: "Visa", Color: "amber", Start: "11:00", End: "12:30", StartMin: 660, EndMin: 750, Kind: "focus"},
			{ID: "e3", Title: "Call · proposta Visa", Context: "Visa", Color: "amber", Start: "14:00", End: "15:00", StartMin: 840, EndMin: 900, Kind: "event", Conflict: true},
			{ID: "e4", Title: "Bayer · revisão dossiê", Context: "Bayer", Color: "teal", Start: "14:00", End: "14:30", StartMin: 840, EndMin: 870, Kind: "event", Conflict: true},
			{ID: "e5", Title: "Ligar pro contador · Pessoal", Context: "Pessoal", Color: "", Start: "16:30", End: "17:00", StartMin: 990, EndMin: 1020, Kind: "personal"},
		},
	}
}

func analysis() Analysis {
	return Analysis{
		Summary:    "Bom dia, Renato. Hoje você tem 4 reuniões e 5 tarefas. A manhã está mais leve — protegi seu bloco de foco das 11h. Resolvi 1 conflito às 14h e priorizei a proposta da Visa.",
		Meetings:   4,
		Tasks:      5,
		FocusFree:  "3h",
		Contexts:   6,
		FocusShare: 0.6,
		TasksDone:  "2/5",
		TasksRatio: 0.4,
		Apps: []AppHealth{
			{"aqno-api", "ok", "128 ms", []int{15, 12, 16, 9, 13, 7, 10, 6}},
			{"Postgres", "ok", "6 ms", []int{12, 13, 11, 12, 10, 11, 9, 10}},
			{"Worker da fila", "warn", "1 retry", []int{8, 12, 7, 14, 9, 16, 11, 18}},
		},
		CashMonth: "R$ 48,2k",
		CashDelta: "▲ 12%",
		CashBars:  []int{46, 62, 54, 78, 70, 100},
		Personal: []Metric{
			{"sono", 0.85, "7h", "violet"},
			{"água", 0.55, "1.4L", "blue"},
			{"passos", 0.62, "6.2k", "teal"},
		},
		MentorTitle: "Conselho do mentor",
		MentorBody:  "Reuniões somam 40% do dia. Considere mover a sync da Devlith para quinta e blindar a manhã — você avança 2 entregas críticas.",
	}
}

func vps() Vps {
	return Vps{
		Host:       "aqno@10.0.4.12",
		Uptime:     "uptime 27d",
		Online:     true,
		Warnings:   1,
		CPU:        0.42,
		RAM:        0.72,
		Disk:       0.58,
		CPUDetail:  "load 2.1 / 1.8 / 1.5",
		RAMDetail:  "11.5 / 16 GB",
		DiskDetail: "232 / 400 GB",
		Containers: []Container{
			{"aqno-api", "node:20", "running", "18%", "240 MB"},
			{"postgres-15", "pg:15", "running", "6%", "512 MB"},
			{"worker-fila", "go:1.22", "restarting", "2%", "88 MB"},
			{"nginx-proxy", "nginx", "running", "1%", "32 MB"},
			{"redis", "redis:7", "running", "3%", "64 MB"},
		},
		Logs: []LogLine{
			{"09:14:02", "INFO", "api · POST /v1/notes 200 128ms"},
			{"09:14:03", "INFO", "pg · checkpoint complete 412ms"},
			{"09:14:05", "WARN", "worker · retry job#4821 (timeout)"},
			{"09:14:05", "WARN", "worker · fila acima de 200 itens"},
			{"09:14:06", "CMD", "íris · restart worker-fila (aguard. confirmação)"},
			{"09:14:09", "INFO", "worker · SIGTERM enviado"},
			{"09:14:11", "OK", "worker · container healthy"},
		},
	}
}

func chatThread() []ChatMessage {
	return []ChatMessage{
		{ID: "m1", From: "user", Text: "Bom dia. Como tá meu dia na Cogna?", Time: "09:11"},
		{ID: "m2", From: "aqno", Text: "Bom dia, Renato. Na Cogna você tem a daily às 9:30 e a proposta Q3 vence sexta. A Marina pediu revisão do orçamento ontem.", Time: "09:11", Ref: &ChatRef{"memory", "memória · call da Cogna, 21 jun", ""}},
		{ID: "m3", From: "user", Text: "Cria um lembrete pra revisar o orçamento às 13:30.", Time: "09:12"},
		{ID: "m4", From: "aqno", Text: "Anotado. Criei um lembrete pra 13:30 e vinculei à Cogna.", Time: "09:12", Ref: &ChatRef{"action", "Lembrete criado · hoje 13:30 · Cogna", "success"}},
		{ID: "m5", From: "user", Text: "E o VPS, tá tudo certo?", Time: "09:13"},
		{ID: "m6", From: "aqno", Text: "Verifiquei agora — 4 containers ok, o worker-fila está reiniciando. RAM em 72%", Time: "agora", Streaming: true},
	}
}

func graph() Graph {
	return Graph{
		Nodes: []GraphNode{
			{"iris", "Íris", 0.5, 0.5, "#C4B5FD", "context", 17},
			{"cogna", "Cogna", 0.24, 0.27, "violet", "context", 13},
			{"c1", "Daily 9:30", 0.11, 0.15, "violet", "event", 0},
			{"c2", "Proposta Q3", 0.35, 0.12, "violet", "project", 0},
			{"c3", "Marina · PM", 0.09, 0.37, "violet", "person", 0},
			{"bayer", "Bayer", 0.76, 0.27, "teal", "context", 13},
			{"b1", "Auditoria", 0.89, 0.15, "teal", "task", 0},
			{"b2", "Dossiê regulatório", 0.66, 0.12, "teal", "project", 0},
			{"b3", "Dr. Klein", 0.91, 0.38, "teal", "person", 0},
			{"visa", "Visa", 0.78, 0.74, "amber", "context", 13},
			{"v1", "Integração API", 0.91, 0.86, "amber", "project", 0},
			{"v2", "Decisão: tokenização", 0.63, 0.88, "amber", "decision", 0},
			{"dev", "Devlith", 0.22, 0.74, "rose", "context", 13},
			{"d1", "Sprint 14", 0.10, 0.87, "rose", "project", 0},
			{"d2", "Bug pagamento", 0.34, 0.88, "rose", "task", 0},
			{"pit", "Pitrace", 0.5, 0.21, "blue", "context", 12},
			{"p1", "Deploy v2", 0.5, 0.07, "blue", "event", 0},
		},
		Edges: []GraphEdge{
			{"iris", "cogna"}, {"iris", "bayer"}, {"iris", "visa"}, {"iris", "dev"}, {"iris", "pit"},
			{"cogna", "c1"}, {"cogna", "c2"}, {"cogna", "c3"},
			{"bayer", "b1"}, {"bayer", "b2"}, {"bayer", "b3"},
			{"visa", "v1"}, {"visa", "v2"},
			{"dev", "d1"}, {"dev", "d2"},
			{"pit", "p1"}, {"pit", "visa"}, {"c2", "p1"},
		},
	}
}
