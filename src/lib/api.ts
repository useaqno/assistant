// Typed REST client for the aqnod daemon.

import { getDaemonUrl } from './tauri'
import type {
  Agenda,
  Analysis,
  Bootstrap,
  ChatMessage,
  Config,
  Context,
  Event,
  Graph,
  Persona,
  Server,
  Task,
  TodayBrief,
  Vps
} from './types'

async function req<T>(path: string, init?: RequestInit): Promise<T> {
  const base = await getDaemonUrl()
  const res = await fetch(`${base}${path}`, init)
  if (!res.ok) throw new Error(`aqnod ${path} → ${res.status}`)
  return (await res.json()) as T
}

function get<T>(path: string): Promise<T> {
  return req<T>(path)
}

function send<T>(method: string, path: string, body?: unknown): Promise<T> {
  return req<T>(path, {
    method,
    headers: { 'Content-Type': 'application/json' },
    body: body === undefined ? undefined : JSON.stringify(body)
  })
}

export interface EventInput {
  context?: string
  title: string
  kind?: 'event' | 'focus' | 'personal'
  start: string
  end?: string
  rrule?: string
  date?: string
  originVoice?: string
}

export const api = {
  health: () => get<{ ok: boolean; version: string }>('/health'),

  // bootstrap / onboarding / config
  bootstrap: () => get<Bootstrap>('/v1/bootstrap'),
  onboarding: (p: Persona) => send<{ ok: boolean }>('POST', '/v1/onboarding', p),
  config: () => get<Config>('/v1/config'),
  setConfig: (patch: Config) => send<{ ok: boolean }>('POST', '/v1/config', patch),

  // contexts
  contexts: () => get<Context[]>('/v1/contexts'),
  createContext: (label: string, color: string, aiMode: string) =>
    send<Context>('POST', '/v1/contexts', { label, color, aiMode }),
  setContextAIMode: (label: string, mode: string) =>
    send<{ ok: boolean }>('POST', '/v1/contexts/ai_mode', { label, mode }),

  // dashboards
  today: () => get<TodayBrief>('/v1/today'),
  analysis: () => get<Analysis>('/v1/analysis'),
  agenda: (date?: string) => get<Agenda>(`/v1/agenda${date ? `?date=${date}` : ''}`),

  // calendar
  eventsRange: (from: string, to: string) => get<Event[]>(`/v1/events/range?from=${from}&to=${to}`),
  createEvent: (e: EventInput) => send<{ ok: boolean; id: string }>('POST', '/v1/events', e),
  updateEvent: (id: string, e: EventInput) => send<{ ok: boolean }>('PUT', `/v1/events/${id}`, e),
  deleteEvent: (id: string) => send<{ ok: boolean }>('DELETE', `/v1/events/${id}`),
  cancelOccurrence: (id: string, date: string) =>
    send<{ ok: boolean }>('POST', `/v1/events/${id}/cancel`, { date }),

  // tasks
  tasks: () => get<Task[]>('/v1/tasks'),
  createTask: (title: string, context?: string, originVoice?: string) =>
    send<{ ok: boolean; id: string }>('POST', '/v1/tasks', { title, context, originVoice }),
  setTaskDone: (id: string, done: boolean) =>
    send<{ ok: boolean }>('PATCH', `/v1/tasks/${id}`, { done }),
  deleteTask: (id: string) => send<{ ok: boolean }>('DELETE', `/v1/tasks/${id}`),

  // graph
  graph: () => get<Graph>('/v1/graph'),

  // chat
  chat: (conversation?: string) =>
    get<ChatMessage[]>(`/v1/chat${conversation ? `?conversation=${conversation}` : ''}`),
  sendChat: (text: string, conversation?: string) =>
    send<ChatMessage>('POST', '/v1/chat', { text, conversation }),

  // voice intent (text path; audio handled via WS in voice.ts)
  voiceIntent: (transcript: string, context?: string) =>
    send<ChatMessage>('POST', '/v1/voice/intent', { transcript, context }),
  speak: (text: string) => send<{ ok: boolean }>('POST', '/v1/voice/speak', { text }),

  // vps
  vps: () => get<Vps>('/v1/vps'),
  restart: (container: string, confirm: boolean) =>
    send<{ ok?: boolean; needsConfirm?: boolean; message: string }>('POST', '/v1/vps/restart', {
      container,
      confirm
    }),

  // servers
  servers: () => get<Server[]>('/v1/servers'),
  createServer: (s: Partial<Server> & { secret?: string }) =>
    send<{ ok: boolean; id: string }>('POST', '/v1/servers', s),
  deleteServer: (id: string) => send<{ ok: boolean }>('DELETE', `/v1/servers/${id}`),

  // llm key (stored in the Keychain by the daemon)
  setLLMKey: (provider: string, key: string) =>
    send<{ ok: boolean }>('POST', '/v1/llm/key', { provider, key }),
  llmKeyStatus: (provider: string) =>
    get<{ provider: string; configured: boolean }>(`/v1/llm/key?provider=${provider}`),

  // voice models / engine
  voiceEngine: () =>
    get<{ active: string; available: boolean; apple: boolean }>('/v1/voice/engine'),
  voiceModels: () => get<VoiceModel[]>('/v1/voice/models'),
  downloadVoiceModel: (tier: string) =>
    send<{ ok: boolean; status?: string }>('POST', `/v1/voice/models/${tier}`, {})
}

export interface VoiceModel {
  tier: string
  present: boolean
  verified: boolean
  bytes: number
  sizeBytes: number
}

/** Stream a chat reply token-by-token over SSE. Returns a cancel fn. */
export async function streamChat(
  text: string,
  conversation: string | undefined,
  onDelta: (chunk: string) => void,
  onDone: (final: ChatMessage) => void,
  onError: (msg: string) => void
): Promise<() => void> {
  const base = await getDaemonUrl()
  const url = `${base}/v1/chat/stream?text=${encodeURIComponent(text)}${
    conversation ? `&conversation=${conversation}` : ''
  }`
  const es = new EventSource(url)
  es.addEventListener('delta', (e) => onDelta(JSON.parse((e as MessageEvent).data)))
  es.addEventListener('done', (e) => {
    onDone(JSON.parse((e as MessageEvent).data))
    es.close()
  })
  es.addEventListener('error', (e) => {
    const data = (e as MessageEvent).data
    onError(data ? String(JSON.parse(data)) : 'stream error')
    es.close()
  })
  return () => es.close()
}
