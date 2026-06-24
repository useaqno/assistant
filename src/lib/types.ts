// Domain types — mirror of the Go daemon's JSON (daemon/data.go).

export type ContextColor = 'violet' | 'teal' | 'amber' | 'rose' | 'blue' | string;
export type PresenceState =
  | 'idle'
  | 'listening'
  | 'transcribing'
  | 'thinking'
  | 'speaking'
  | 'confirming';

export interface Context {
  id: string;
  label: string;
  color: ContextColor;
}

export interface Interaction {
  title: string;
  context: string;
  color: ContextColor;
  tag: string;
  tone: string;
  when: string;
}

export interface Task {
  title: string;
  done: boolean;
}

export interface Event {
  id?: string;
  title: string;
  context: string;
  color: ContextColor;
  start: string;
  end: string;
  startMin?: number;
  endMin?: number;
  kind?: 'event' | 'focus' | 'personal';
  conflict?: boolean;
}

export interface TodayBrief {
  greeting: string;
  date: string;
  companion: string;
  state: PresenceState;
  headline: string;
  meetings: number;
  tasks: number;
  focusFree: string;
  nextEvent: Event;
  taskList: Task[];
  mentor: string;
  recent: Interaction[];
}

export interface Agenda {
  day: string;
  conflicts: number;
  focus: number;
  events: Event[];
}

export interface AppHealth {
  name: string;
  status: 'ok' | 'warn' | 'down';
  latency: string;
  spark: number[];
}

export interface Metric {
  label: string;
  value: number;
  big: string;
  color: ContextColor;
}

export interface Analysis {
  summary: string;
  meetings: number;
  tasks: number;
  focusFree: string;
  contexts: number;
  focusShare: number;
  tasksDone: string;
  tasksRatio: number;
  apps: AppHealth[];
  cashMonth: string;
  cashDelta: string;
  cashBars: number[];
  personal: Metric[];
  mentorTitle: string;
  mentorBody: string;
}

export interface Container {
  name: string;
  image: string;
  status: 'running' | 'restarting' | 'down';
  cpu: string;
  mem: string;
}

export interface LogLine {
  time: string;
  level: 'INFO' | 'WARN' | 'OK' | 'CMD';
  body: string;
}

export interface Vps {
  host: string;
  uptime: string;
  online: boolean;
  warnings: number;
  cpu: number;
  ram: number;
  disk: number;
  cpuDetail: string;
  ramDetail: string;
  diskDetail: string;
  containers: Container[];
  logs: LogLine[];
}

export interface ChatRef {
  kind: 'memory' | 'action';
  label: string;
  tone: string;
}

export interface ChatMessage {
  id: string;
  from: 'user' | 'aqno';
  text: string;
  time: string;
  streaming?: boolean;
  ref?: ChatRef;
}

export interface GraphNode {
  id: string;
  label: string;
  x: number;
  y: number;
  color: ContextColor;
  kind: string;
  size?: number;
}

export interface GraphEdge {
  from: string;
  to: string;
}

export interface Graph {
  nodes: GraphNode[];
  edges: GraphEdge[];
}
