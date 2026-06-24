// Live presence state, fed by the daemon's SSE stream. Components subscribe to
// this to make the whole shell feel "alive" (the orb, the voice bar).

import { writable } from 'svelte/store';
import type { PresenceState } from '$lib/types';
import { subscribePresence } from '$lib/tauri';

export interface PresenceSnapshot {
  state: PresenceState;
  level: number;
}

function createPresence() {
  const { subscribe, set } = writable<PresenceSnapshot>({ state: 'idle', level: 0.6 });
  let unsub: (() => void) | null = null;

  return {
    subscribe,
    /** Begin listening to the daemon stream. Returns a stop fn. */
    async connect() {
      if (unsub) return;
      try {
        unsub = await subscribePresence((state, level) => set({ state, level }));
      } catch {
        /* daemon offline — keep the static idle state */
      }
    },
    /** Manually override (e.g. while the user holds space to talk). */
    set: (snap: PresenceSnapshot) => set(snap),
    disconnect() {
      unsub?.();
      unsub = null;
    }
  };
}

export const presence = createPresence();
