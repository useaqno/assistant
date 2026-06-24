<script lang="ts">
  // Presence — Aqno's living aura. A violet orb that breathes and reacts to
  // voice. Six states: idle · listening · transcribing · thinking · speaking ·
  // confirming. Honours prefers-reduced-motion via the global keyframes.
  import type { PresenceState } from '$lib/types'

  const {
    state = 'idle',
    size = 96,
    level = 0.6,
    label = ''
  }: { state?: PresenceState; size?: number; level?: number; label?: string } = $props()

  const coreAnim: Record<PresenceState, string> = {
    idle: 'aqno-breathe var(--dur-breath) var(--ease-organic) infinite',
    listening: 'aqno-breathe 2600ms var(--ease-organic) infinite',
    transcribing: 'aqno-breathe 3200ms var(--ease-organic) infinite',
    thinking: 'aqno-shimmer var(--dur-shimmer) ease-in-out infinite',
    speaking: 'aqno-breathe 1800ms var(--ease-organic) infinite',
    confirming: 'none'
  }
  const auraOpacity: Record<PresenceState, number> = {
    idle: 0.5,
    listening: 0.85,
    transcribing: 0.6,
    thinking: 0.7,
    speaking: 0.95,
    confirming: 0.8
  }

  const isConfirm = $derived(state === 'confirming')
  const bars = $derived([0.45, 0.8, 1, 0.7, 0.35, 0.6, 0.9])
</script>

<div class="presence" style="--sz:{size}px">
  <div class="orb-wrap" style="width:{size}px;height:{size}px">
    <!-- aura -->
    <div
      class="aura"
      style="
        inset:{-size * 0.5}px;
        opacity:{auraOpacity[state]};
        background:{isConfirm
        ? 'radial-gradient(circle at 50% 50%, rgba(74,222,128,0.5) 0%, rgba(74,222,128,0) 70%)'
        : 'var(--grad-aura)'};"
    ></div>

    <!-- listening ripples -->
    {#if state === 'listening'}
      {#each [0, 1, 2] as i (i)}
        <div class="ripple" style="animation-delay:{i * 800}ms"></div>
      {/each}
    {/if}

    <!-- core -->
    <div
      class="core"
      style="
        animation:{coreAnim[state]};
        background:{isConfirm
        ? 'radial-gradient(circle at 50% 42%, #86EFAC 0%, #4ADE80 45%, #16A34A 100%)'
        : 'var(--grad-presence)'};
        box-shadow: var(--glow-presence), inset 0 {size * 0.06}px {size *
        0.12}px rgba(255,255,255,0.35), inset 0 {-size * 0.08}px {size * 0.16}px rgba(20,8,40,0.5);"
    >
      <div class="spec"></div>

      {#if state === 'transcribing'}
        <div class="ring" style="inset:{size * 0.16}px"></div>
      {/if}

      {#if state === 'speaking'}
        <div class="wave" style="gap:{size * 0.04}px">
          {#each bars as b, i (i)}
            <span
              style="
                width:{Math.max(2, size * 0.035)}px;
                animation-delay:{i * 60}ms;
                --h:{Math.max(0.22, b * (0.5 + level * 0.6))};"
            ></span>
          {/each}
        </div>
      {/if}

      {#if isConfirm}
        <svg
          width={size * 0.42}
          height={size * 0.42}
          viewBox="0 0 24 24"
          fill="none"
          stroke="#08210F"
          stroke-width="3.2"
          stroke-linecap="round"
          stroke-linejoin="round"
          style="position:relative"
        >
          <path d="M4 12.5l5 5L20 6.5" />
        </svg>
      {/if}
    </div>
  </div>

  {#if label}
    <span class="label" style="color:{isConfirm ? 'var(--success)' : 'var(--text-2)'}">{label}</span
    >
  {/if}
</div>

<style>
  .presence {
    display: inline-flex;
    flex-direction: column;
    align-items: center;
    gap: 14px;
  }
  .orb-wrap {
    position: relative;
    display: grid;
    place-items: center;
  }
  .aura {
    position: absolute;
    border-radius: 50%;
    filter: blur(2px);
    transition: opacity var(--dur-slow) var(--ease-out);
  }
  .ripple {
    position: absolute;
    inset: 0;
    border-radius: 50%;
    border: 1.5px solid var(--purple-glow);
    animation: aqno-ripple 2400ms var(--ease-out) infinite;
  }
  .core {
    position: relative;
    width: 100%;
    height: 100%;
    border-radius: 50%;
    display: grid;
    place-items: center;
    overflow: hidden;
  }
  .spec {
    position: absolute;
    top: 14%;
    left: 22%;
    width: 34%;
    height: 26%;
    border-radius: 50%;
    background: radial-gradient(circle, rgba(255, 255, 255, 0.8), rgba(255, 255, 255, 0));
    filter: blur(2px);
  }
  .ring {
    position: absolute;
    border-radius: 50%;
    border: 2px solid rgba(255, 255, 255, 0.25);
    border-top-color: var(--text-on-purple);
    animation: aqno-spin 900ms linear infinite;
  }
  .wave {
    position: absolute;
    inset: 0;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .wave span {
    height: 60%;
    border-radius: 999px;
    background: var(--text-on-purple);
    transform-origin: center;
    transform: scaleY(var(--h));
    opacity: 0.92;
    animation: aqno-wave 680ms var(--ease-in-out) infinite;
  }
  .label {
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: var(--tracking-caps);
    text-transform: uppercase;
  }
</style>
