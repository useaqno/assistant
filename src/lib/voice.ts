// Voice client. Two capture paths sit behind one API:
//   1. Web Speech (SpeechRecognition) — works in the dev browser and Safari.
//   2. Native streaming to the daemon's whisper.cpp over WebSocket — used in the
//      packaged Tauri app where WKWebView lacks SpeechRecognition (wired in WS7b
//      once the daemon's /v1/voice/stream endpoint exists).
// On a final transcript we POST to /v1/voice/intent and speak the reply.

import { get } from 'svelte/store'
import { api } from './api'
import { getDaemonUrl } from './tauri'
import { setVoice } from '$stores/voice'
import { presence } from '$stores/presence'
import { app } from '$stores/app'

type SpeechRecognitionLike = {
  lang: string
  continuous: boolean
  interimResults: boolean
  start: () => void
  stop: () => void
  onresult: ((e: SpeechRecognitionEventLike) => void) | null
  onerror: ((e: unknown) => void) | null
  onend: (() => void) | null
}
interface SpeechRecognitionEventLike {
  results: ArrayLike<ArrayLike<{ transcript: string }> & { isFinal: boolean }>
}

function recognizer(): SpeechRecognitionLike | null {
  const w = window as unknown as {
    SpeechRecognition?: new () => SpeechRecognitionLike
    webkitSpeechRecognition?: new () => SpeechRecognitionLike
  }
  const Ctor = w.SpeechRecognition ?? w.webkitSpeechRecognition
  return Ctor ? new Ctor() : null
}

let rec: SpeechRecognitionLike | null = null
let active = false

function drive(
  state: Parameters<typeof setVoice>[0]['state'],
  extra: Record<string, unknown> = {}
) {
  setVoice({ state, ...extra })
  presence.set({ state: state ?? 'idle', level: 0.7 })
}

/** Begin a push-to-talk capture. Idempotent. */
export function startListening(): void {
  if (active) return
  const lang = get(app).config['voice.stt_lang'] || 'pt'
  rec = recognizer()
  if (!rec) {
    // Native path lands in WS7b; until then guide the user to type.
    drive('listening', {
      transcript: '',
      hint: 'Reconhecimento nativo em breve — use o chat por enquanto'
    })
    setTimeout(() => active && stopListening(), 1500)
    return
  }
  active = true
  rec.lang = lang.startsWith('pt') ? 'pt-BR' : lang === 'en' ? 'en-US' : lang
  rec.continuous = false
  rec.interimResults = true
  rec.onresult = (e) => {
    let text = ''
    let final = false
    for (let i = 0; i < e.results.length; i++) {
      text += e.results[i][0].transcript
      if (e.results[i].isFinal) final = true
    }
    drive('transcribing', { transcript: text, level: 0.6 })
    if (final) handleTranscript(text)
  }
  rec.onerror = () => stopListening()
  rec.onend = () => {
    active = false
  }
  drive('listening', { transcript: '', level: 0.85 })
  rec.start()
}

/** Stop the current capture. */
export function stopListening(): void {
  active = false
  try {
    rec?.stop()
  } catch {
    /* already stopped */
  }
  drive('idle', { transcript: '' })
}

async function handleTranscript(text: string) {
  if (!text.trim()) {
    drive('idle')
    return
  }
  drive('thinking', { transcript: text })
  try {
    const reply = await api.voiceIntent(text)
    drive('speaking', { transcript: reply.text, level: 0.8 })
    await speak(reply.text)
  } catch {
    drive('idle', { transcript: '' })
    return
  }
  drive('idle', { transcript: '' })
}

/** Speak text via the daemon TTS, falling back to the browser synthesizer. */
export async function speak(text: string): Promise<void> {
  try {
    const base = await getDaemonUrl()
    const res = await fetch(`${base}/v1/voice/speak`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ text })
    })
    if (res.ok) return
  } catch {
    /* fall through to browser TTS */
  }
  if (typeof speechSynthesis !== 'undefined') {
    await new Promise<void>((resolve) => {
      const u = new SpeechSynthesisUtterance(text)
      u.lang = 'pt-BR'
      const speed = parseFloat(get(app).config['voice.speed'] || '1') || 1
      u.rate = speed
      u.onend = () => resolve()
      u.onerror = () => resolve()
      speechSynthesis.speak(u)
    })
  }
}

/** True while a capture is in progress. */
export function isListening(): boolean {
  return active
}
