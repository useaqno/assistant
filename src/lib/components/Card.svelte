<script lang="ts">
  import type { Snippet } from 'svelte';

  let {
    padding = 20,
    interactive = false,
    glow = false,
    children
  }: { padding?: number; interactive?: boolean; glow?: boolean; children?: Snippet } = $props();
</script>

<div
  class="card"
  class:interactive
  class:glow
  style="padding:{padding}px"
>
  {@render children?.()}
</div>

<style>
  .card {
    background: var(--surface-2);
    border-radius: var(--radius-card);
    border: 1px solid var(--border-violet);
    box-shadow: var(--shadow-2), inset 0 1px 0 var(--highlight-top);
    transition: var(--transition-base);
  }
  .card.glow {
    border-color: var(--purple-024);
    box-shadow: var(--shadow-2), inset 0 1px 0 var(--highlight-top), var(--glow-md);
  }
  .card.interactive { cursor: pointer; }
  .card.interactive:hover {
    transform: translateY(-2px);
    border-color: var(--purple-024);
  }
</style>
