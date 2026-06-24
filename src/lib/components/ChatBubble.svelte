<script lang="ts">
  import type { Snippet } from 'svelte';

  let {
    from = 'aqno',
    name = '',
    time = '',
    streaming = false,
    children
  }: {
    from?: 'user' | 'aqno';
    name?: string;
    time?: string;
    streaming?: boolean;
    children?: Snippet;
  } = $props();

  let isUser = $derived(from === 'user');
</script>

<div class="row" class:user={isUser}>
  {#if !isUser && name}
    <div class="who"><span class="seed"></span>{name}</div>
  {/if}
  <div class="bubble" class:user={isUser}>
    {@render children?.()}
    {#if streaming}
      <span class="typing">
        {#each [0, 1, 2] as i}<span style="animation-delay:{i * 180}ms"></span>{/each}
      </span>
    {/if}
  </div>
  {#if time}<span class="time">{time}</span>{/if}
</div>

<style>
  .row {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 4px;
    max-width: 78%;
    align-self: flex-start;
  }
  .row.user { align-items: flex-end; align-self: flex-end; }
  .who {
    display: flex;
    align-items: center;
    gap: 6px;
    padding-left: 4px;
    font-family: var(--font-ui);
    font-size: 12px;
    font-weight: var(--weight-medium);
    color: var(--text-2);
  }
  .seed {
    width: 14px;
    height: 14px;
    border-radius: 50%;
    background: var(--grad-presence);
    box-shadow: var(--glow-sm);
  }
  .bubble {
    padding: 11px 15px;
    font-family: var(--font-body);
    font-size: 15px;
    line-height: var(--leading-snug);
    border-radius: var(--radius-card);
    border-bottom-left-radius: 6px;
    background: var(--surface-2);
    color: var(--text-1);
    border: 1px solid var(--border-violet);
    box-shadow: inset 0 1px 0 var(--highlight-top);
  }
  .bubble.user {
    border-bottom-left-radius: var(--radius-card);
    border-bottom-right-radius: 6px;
    background: var(--grad-active);
    color: var(--text-on-purple);
    border: 1px solid transparent;
    box-shadow: var(--glow-sm), inset 0 1px 0 rgba(255, 255, 255, 0.18);
  }
  .time {
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--text-3);
    padding: 0 4px;
  }
  .typing { display: inline-flex; gap: 3px; margin-left: 6px; vertical-align: middle; }
  .typing span {
    width: 5px;
    height: 5px;
    border-radius: 50%;
    background: currentColor;
    opacity: 0.7;
    animation: aqno-shimmer 1000ms ease-in-out infinite;
  }
</style>
