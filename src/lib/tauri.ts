// Bridge to the Tauri shell. Degrades gracefully when running in a plain
// browser (`pnpm dev` without `tauri dev`): we fall back to the default daemon
// port so the UI is fully workable in the browser too.

import type { PresenceState } from './types'

export const isTauri = typeof window !== 'undefined' && '__TAURI_INTERNALS__' in window

const DEFAULT_DAEMON = 'http://127.0.0.1:8787'

let cachedUrl: string | null = null

/** Resolve the daemon base URL (asks the Rust side when inside Tauri). */
export async function getDaemonUrl(): Promise<string> {
  if (cachedUrl) return cachedUrl
  if (isTauri) {
    try {
      const { invoke } = await import('@tauri-apps/api/core')
      cachedUrl = await invoke<string>('daemon_url')
      return cachedUrl
    } catch {
      /* fall through to default */
    }
  }
  cachedUrl = DEFAULT_DAEMON
  return cachedUrl
}

/** Fires once the sidecar has announced its port (Tauri only). */
export async function onDaemonReady(cb: (url: string) => void): Promise<() => void> {
  if (!isTauri) return () => {}
  const { listen } = await import('@tauri-apps/api/event')
  return listen<{ url: string; port: number }>('daemon-ready', (e) => {
    cachedUrl = e.payload.url
    cb(e.payload.url)
  })
}

/** Whether this is a packaged build (mic/speech only work in the .app bundle). */
export async function isBundled(): Promise<boolean> {
  if (!isTauri) return false
  try {
    const { invoke } = await import('@tauri-apps/api/core')
    return await invoke<boolean>('is_bundled')
  } catch {
    return false
  }
}

/** Fires when the global push-to-talk hotkey is pressed (Tauri only). */
export async function onVoiceHotkey(cb: () => void): Promise<() => void> {
  if (!isTauri) return () => {}
  const { listen } = await import('@tauri-apps/api/event')
  return listen('voice-hotkey', () => cb())
}

/** Subscribe to the daemon's SSE stream of presence-state ticks. */
export async function subscribePresence(
  cb: (state: PresenceState, level: number) => void
): Promise<() => void> {
  const base = await getDaemonUrl()
  const es = new EventSource(`${base}/v1/events`)
  es.addEventListener('presence', (ev) => {
    try {
      const data = JSON.parse((ev as MessageEvent).data)
      cb(data.state, data.level ?? 0.6)
    } catch {
      /* ignore malformed frames */
    }
  })
  return () => es.close()
}
