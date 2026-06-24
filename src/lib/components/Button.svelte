<script lang="ts">
  import type { Snippet } from 'svelte'
  import Icon from './Icon.svelte'

  const {
    variant = 'primary',
    size = 'md',
    icon = '',
    iconRight = '',
    full = false,
    disabled = false,
    onclick,
    children
  }: {
    variant?: 'primary' | 'ghost' | 'subtle' | 'danger'
    size?: 'sm' | 'md' | 'lg'
    icon?: string
    iconRight?: string
    full?: boolean
    disabled?: boolean
    onclick?: () => void
    children?: Snippet
  } = $props()

  const dims = {
    sm: { h: 32, px: 12, fs: 13, gap: 6 },
    md: { h: 40, px: 16, fs: 15, gap: 8 },
    lg: { h: 48, px: 22, fs: 16, gap: 10 }
  }[size]
</script>

<button
  class="btn {variant}"
  class:full
  {disabled}
  {onclick}
  style="height:{dims.h}px;padding:0 {dims.px}px;font-size:{dims.fs}px;gap:{dims.gap}px"
>
  {#if icon}<Icon name={icon} size={dims.fs + 2} />{/if}
  {@render children?.()}
  {#if iconRight}<Icon name={iconRight} size={dims.fs + 2} />{/if}
</button>

<style>
  .btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    font-family: var(--font-ui);
    font-weight: var(--weight-medium);
    letter-spacing: var(--tracking-tight);
    border-radius: var(--radius-input);
    border: 1px solid transparent;
    cursor: pointer;
    white-space: nowrap;
    user-select: none;
    transition: var(--transition-base);
  }
  .btn.full {
    width: 100%;
  }
  .btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
  .btn:active:not(:disabled) {
    transform: scale(0.97);
  }

  .primary {
    background: var(--grad-active);
    color: var(--text-on-purple);
    box-shadow:
      var(--glow-sm),
      inset 0 1px 0 rgba(255, 255, 255, 0.22);
  }
  .primary:hover:not(:disabled) {
    filter: brightness(1.08);
    box-shadow:
      var(--glow-md),
      inset 0 1px 0 rgba(255, 255, 255, 0.25);
  }
  .ghost {
    background: transparent;
    color: var(--text-1);
    border-color: var(--border-strong);
  }
  .ghost:hover:not(:disabled) {
    border-color: var(--purple-glow);
  }
  .subtle {
    background: var(--surface-3);
    color: var(--text-1);
    border-color: var(--border-subtle);
    box-shadow: inset 0 1px 0 var(--highlight-top);
  }
  .subtle:hover:not(:disabled) {
    background: #2c2640;
  }
  .danger {
    background: var(--danger-bg);
    color: var(--danger);
    border-color: rgba(248, 113, 113, 0.3);
  }
  .danger:hover:not(:disabled) {
    background: rgba(248, 113, 113, 0.18);
  }
</style>
