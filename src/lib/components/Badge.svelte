<script lang="ts">
  import type { Snippet } from 'svelte'

  const {
    tone = 'neutral',
    size = 'md',
    dot = false,
    children
  }: {
    tone?: 'neutral' | 'purple' | 'success' | 'warning' | 'danger' | 'info'
    size?: 'sm' | 'md'
    dot?: boolean
    children?: Snippet
  } = $props()

  const map = {
    neutral: {
      fg: 'var(--text-2)',
      soft: 'rgba(255,255,255,0.06)',
      line: 'var(--border-strong)',
      dot: 'var(--text-3)'
    },
    purple: {
      fg: 'var(--purple-glow)',
      soft: 'var(--purple-012)',
      line: 'var(--purple-024)',
      dot: 'var(--purple-glow)'
    },
    success: {
      fg: 'var(--success)',
      soft: 'var(--success-bg)',
      line: 'rgba(74,222,128,0.3)',
      dot: 'var(--success)'
    },
    warning: {
      fg: 'var(--warning)',
      soft: 'var(--warning-bg)',
      line: 'rgba(251,191,36,0.3)',
      dot: 'var(--warning)'
    },
    danger: {
      fg: 'var(--danger)',
      soft: 'var(--danger-bg)',
      line: 'rgba(248,113,113,0.3)',
      dot: 'var(--danger)'
    },
    info: {
      fg: 'var(--info)',
      soft: 'var(--info-bg)',
      line: 'rgba(96,165,250,0.3)',
      dot: 'var(--info)'
    }
  }[tone]
  const dims = { sm: { h: 18, px: 7, fs: 11 }, md: { h: 22, px: 9, fs: 12 } }[size]
</script>

<span
  class="badge"
  style="height:{dims.h}px;padding:0 {dims.px}px;font-size:{dims.fs}px;color:{map.fg};background:{map.soft};border-color:{map.line}"
>
  {#if dot}<span class="d" style="background:{map.dot}"></span>{/if}
  {@render children?.()}
</span>

<style>
  .badge {
    display: inline-flex;
    align-items: center;
    gap: 5px;
    border-radius: var(--radius-pill);
    border: 1px solid transparent;
    font-family: var(--font-ui);
    font-weight: var(--weight-medium);
    letter-spacing: var(--tracking-tight);
    white-space: nowrap;
  }
  .d {
    width: 6px;
    height: 6px;
    border-radius: 50%;
  }
</style>
