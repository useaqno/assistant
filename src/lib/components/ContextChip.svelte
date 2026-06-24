<script lang="ts">
  import type { ContextColor } from '$lib/types';

  let {
    label,
    color = 'violet',
    active = false,
    size = 'md',
    onclick
  }: {
    label: string;
    color?: ContextColor;
    active?: boolean;
    size?: 'sm' | 'md';
    onclick?: () => void;
  } = $props();

  const palette: Record<string, string> = {
    violet: '#A78BFA',
    teal: '#5EEAD4',
    amber: '#FBBF24',
    rose: '#FB7185',
    blue: '#60A5FA'
  };
  let c = $derived(palette[color] ?? color);
  const dims = { sm: { h: 22, fs: 12, px: 8 }, md: { h: 28, fs: 13, px: 11 } }[size];
</script>

<span
  class="chip"
  class:active
  class:click={!!onclick}
  {onclick}
  role={onclick ? 'button' : undefined}
  tabindex={onclick ? 0 : undefined}
  style="
    height:{dims.h}px;padding:0 {dims.px}px;font-size:{dims.fs}px;
    color:{active ? '#0C0A14' : 'var(--text-1)'};
    background:{active ? c : `color-mix(in srgb, ${c} 10%, transparent)`};
    border-color:{active ? 'transparent' : `color-mix(in srgb, ${c} 32%, transparent)`};"
>
  <span
    class="dot"
    style="background:{active ? '#0C0A14' : c};box-shadow:{active ? 'none' : `0 0 8px ${c}`}"
  ></span>
  {label}
</span>

<style>
  .chip {
    display: inline-flex;
    align-items: center;
    gap: 7px;
    width: fit-content;
    border-radius: var(--radius-pill);
    border: 1px solid transparent;
    font-family: var(--font-ui);
    font-weight: var(--weight-medium);
    letter-spacing: var(--tracking-tight);
    white-space: nowrap;
    transition: var(--transition-base);
  }
  .chip.click { cursor: pointer; }
  .dot { width: 8px; height: 8px; border-radius: 50%; }
</style>
