// Typed REST client for the aqnod daemon.

import { getDaemonUrl } from './tauri';
import type { Agenda, Analysis, ChatMessage, Context, Graph, TodayBrief, Vps } from './types';

async function get<T>(path: string): Promise<T> {
  const base = await getDaemonUrl();
  const res = await fetch(`${base}${path}`);
  if (!res.ok) throw new Error(`aqnod ${path} → ${res.status}`);
  return (await res.json()) as T;
}

async function post<T>(path: string, body: unknown): Promise<T> {
  const base = await getDaemonUrl();
  const res = await fetch(`${base}${path}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body)
  });
  if (!res.ok) throw new Error(`aqnod ${path} → ${res.status}`);
  return (await res.json()) as T;
}

export const api = {
  health: () => get<{ ok: boolean; version: string }>('/health'),
  contexts: () => get<Context[]>('/v1/contexts'),
  today: () => get<TodayBrief>('/v1/today'),
  agenda: () => get<Agenda>('/v1/agenda'),
  analysis: () => get<Analysis>('/v1/analysis'),
  vps: () => get<Vps>('/v1/vps'),
  graph: () => get<Graph>('/v1/graph'),
  chat: () => get<ChatMessage[]>('/v1/chat'),
  sendChat: (text: string) => post<ChatMessage>('/v1/chat', { text }),
  restart: (container: string, confirm: boolean) =>
    post<{ ok?: boolean; needsConfirm?: boolean; message: string }>('/v1/vps/restart', {
      container,
      confirm
    })
};
