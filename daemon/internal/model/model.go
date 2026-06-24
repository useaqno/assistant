// Package model holds the JSON-serializable domain types shared across the
// daemon (store, domain logic, HTTP API). Field tags mirror the SvelteKit
// client contract in src/lib/types.ts so the UI consumes them unchanged.
package model

// Context is a workspace (company or personal) that scopes events and tasks.
type Context struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	Color    string `json:"color"`            // palette key: violet|teal|amber|rose|blue
	AIMode   string `json:"aiMode,omitempty"` // cloud | local_only
	Archived bool   `json:"archived,omitempty"`
}

// Persona is the named companion plus the human owner's name.
type Persona struct {
	Name      string `json:"name"`
	Owner     string `json:"owner"`
	Avatar    string `json:"avatar"`  // orbe | animal | personagem | imagem
	AvatarRef string `json:"avatarRef,omitempty"`
	AuraColor string `json:"auraColor"`
	Voice     string `json:"voice,omitempty"`
	Tone      string `json:"tone"`
	WakeWord  string `json:"wakeWord"`
}

// Event is a calendar entry. Recurring events carry an RRULE; concrete
// occurrences returned by a range query also fill Date/StartMin/EndMin.
type Event struct {
	ID         string `json:"id"`
	ContextID  string `json:"contextId,omitempty"`
	Title      string `json:"title"`
	Context    string `json:"context"` // context label for display
	Color      string `json:"color"`
	Start      string `json:"start"` // 'HH:MM'
	End        string `json:"end"`
	StartMin   int    `json:"startMin"`
	EndMin     int    `json:"endMin"`
	Kind       string `json:"kind"`           // event|focus|personal
	Conflict   bool   `json:"conflict"`
	RRule      string `json:"rrule,omitempty"`
	Date       string `json:"date,omitempty"` // 'YYYY-MM-DD' (occurrence date)
	ReminderM  int    `json:"reminderMin,omitempty"`
	OriginVoice string `json:"originVoice,omitempty"`
}

// Task is a to-do, optionally bound to a context.
type Task struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Done      bool   `json:"done"`
	Status    string `json:"status,omitempty"` // aberta | em_andamento | concluida
	ContextID string `json:"contextId,omitempty"`
	Context   string `json:"context,omitempty"`
	Priority  int    `json:"priority,omitempty"`
	Deadline  string `json:"deadline,omitempty"`
}

// Interaction is a recent voice/chat event shown on the Home screen.
type Interaction struct {
	Title   string `json:"title"`
	Context string `json:"context"`
	Color   string `json:"color"`
	Tag     string `json:"tag"`
	Tone    string `json:"tone"`
	When    string `json:"when"`
}

// TodayBrief is the Home dashboard payload.
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

// Agenda is the day payload (events already expanded for the day).
type Agenda struct {
	Day       string  `json:"day"`
	Conflicts int     `json:"conflicts"`
	Focus     int     `json:"focus"`
	Events    []Event `json:"events"`
}

// AppHealth is a monitored service row on the Analysis screen.
type AppHealth struct {
	Name    string `json:"name"`
	Status  string `json:"status"` // ok|warn|down
	Latency string `json:"latency"`
	Spark   []int  `json:"spark"`
}

// Metric is a small labelled ring (personal life cards).
type Metric struct {
	Label string  `json:"label"`
	Value float64 `json:"value"`
	Big   string  `json:"big"`
	Color string  `json:"color"`
}

// Analysis is the daily briefing payload.
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

// ChatRef is an optional memory/action chip attached to a message.
type ChatRef struct {
	Kind  string `json:"kind"` // memory|action
	Label string `json:"label"`
	Tone  string `json:"tone"`
}

// ChatMessage is one turn in a conversation.
type ChatMessage struct {
	ID        string   `json:"id"`
	From      string   `json:"from"` // user|aqno
	Text      string   `json:"text"`
	Time      string   `json:"time"`
	Streaming bool     `json:"streaming,omitempty"`
	Ref       *ChatRef `json:"ref,omitempty"`
}

// GraphNode and GraphEdge feed the knowledge-graph view.
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

// Server is a registered VPS (credentials live in the Keychain).
type Server struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
	User        string `json:"user"`
	AuthType    string `json:"authType"` // senha | chave
	KeychainRef string `json:"keychainRef,omitempty"`
}

// Container, LogLine and Vps describe the infra screen.
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

// Bootstrap is the initial app state (IPC app.bootstrap).
type Bootstrap struct {
	Persona    Persona           `json:"persona"`
	Contexts   []Context         `json:"contexts"`
	Onboarded  bool              `json:"onboarded"`
	Config     map[string]string `json:"config"`
}
