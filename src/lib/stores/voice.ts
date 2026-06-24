import { writable } from 'svelte/store'
import type { PresenceState } from '$lib/types'

export interface Voice {
  state: PresenceState
  transcript: string
  hint: string
  level: number
}

const initial: Voice = {
  state: 'idle',
  transcript: '',
  hint: 'Segure espaço para falar, ou diga "Íris"',
  level: 0.5
}

export const voice = writable<Voice>(initial)

/** Pages call this on mount to put the persistent voice bar into the state that
 *  best illustrates the screen. */
export function setVoice(v: Partial<Voice>) {
  voice.update((c) => ({ ...c, transcript: '', ...v }))
}
