// App-wide state hydrated once from the daemon's bootstrap: who the companion
// is, the user's contexts, and configuration. Screens read the persona name
// from here instead of hardcoding it.

import { writable } from 'svelte/store'
import type { Bootstrap, Config, Context, Persona } from '$lib/types'
import { api } from '$lib/api'

const emptyPersona: Persona = {
  name: '',
  owner: '',
  avatar: 'orbe',
  auraColor: '#8B5CF6',
  tone: 'amigavel',
  wakeWord: 'aqno'
}

export interface AppState {
  ready: boolean
  onboarded: boolean
  persona: Persona
  contexts: Context[]
  config: Config
}

const initial: AppState = {
  ready: false,
  onboarded: false,
  persona: emptyPersona,
  contexts: [],
  config: {}
}

function createApp() {
  const { subscribe, set, update } = writable<AppState>(initial)

  return {
    subscribe,
    /** Load bootstrap from the daemon. Returns whether onboarding is needed. */
    async load(): Promise<boolean> {
      try {
        const b: Bootstrap = await api.bootstrap()
        set({
          ready: true,
          onboarded: b.onboarded,
          persona: b.persona?.name ? b.persona : emptyPersona,
          contexts: b.contexts ?? [],
          config: b.config ?? {}
        })
        return b.onboarded
      } catch {
        set({ ...initial, ready: true })
        return true // daemon offline: don't trap the user in onboarding
      }
    },
    setConfig(patch: Config) {
      update((s) => ({ ...s, config: { ...s.config, ...patch } }))
      api.setConfig(patch).catch(() => {})
    },
    setPersona(persona: Persona) {
      update((s) => ({ ...s, persona, onboarded: true }))
    }
  }
}

export const app = createApp()

/** Companion display name with a sensible fallback. */
export function companionName(persona: Persona): string {
  return persona.name || 'Aqno'
}
