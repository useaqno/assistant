<script lang="ts">
  import Presence from './Presence.svelte';
  import { voice } from '$stores/voice';

  const stateLabel: Record<string, string> = {
    idle: 'Pronto',
    listening: 'Ouvindo',
    transcribing: 'Transcrevendo',
    thinking: 'Pensando',
    speaking: 'Respondendo',
    confirming: 'Confirmado'
  };

  let showWave = $derived($voice.state === 'listening' || $voice.state === 'speaking');
  const bars = [0.5, 0.9, 0.6, 1, 0.7, 0.4, 0.85, 0.55, 0.7];
</script>

<div class="voicebar" class:active={$voice.state !== 'idle'}>
  <Presence state={$voice.state} size={44} level={$voice.level} />

  <div class="body">
    <span class="state" class:confirm={$voice.state === 'confirming'}>
      {stateLabel[$voice.state] ?? 'Pronto'}
    </span>
    <span class="text" class:dim={!$voice.transcript}>
      {$voice.transcript || $voice.hint}
    </span>
  </div>

  {#if showWave}
    <div class="wave">
      {#each bars as b, i}
        <span style="--h:{Math.max(0.25, b * (0.4 + $voice.level * 0.6))};animation-delay:{i * 50}ms"></span>
      {/each}
    </div>
  {/if}
</div>

<style>
  .voicebar {
    display: flex;
    align-items: center;
    gap: 16px;
    height: var(--voicebar-h);
    padding: 0 16px 0 14px;
    background: color-mix(in srgb, var(--surface-2) 86%, transparent);
    backdrop-filter: var(--blur-panel);
    -webkit-backdrop-filter: var(--blur-panel);
    border-radius: var(--radius-pill);
    border: 1px solid var(--border-violet);
    box-shadow: var(--shadow-3), inset 0 1px 0 var(--highlight-top);
    transition: var(--transition-base);
  }
  .voicebar.active {
    box-shadow: var(--shadow-3), inset 0 1px 0 var(--highlight-top), var(--glow-md);
  }
  .body { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 2px; }
  .state {
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--purple-glow);
  }
  .state.confirm { color: var(--success); }
  .text {
    font-family: var(--font-body);
    font-size: 15px;
    color: var(--text-1);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .text.dim { color: var(--text-3); }
  .wave { display: flex; align-items: center; gap: 3px; height: 28px; }
  .wave span {
    width: 3px;
    height: 100%;
    border-radius: 999px;
    background: var(--purple-glow);
    transform-origin: center;
    transform: scaleY(var(--h));
    opacity: 0.85;
    animation: aqno-wave 660ms var(--ease-in-out) infinite;
  }
</style>
