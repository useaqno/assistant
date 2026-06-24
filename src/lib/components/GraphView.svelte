<script lang="ts">
  import type { GraphNode, GraphEdge } from '$lib/types'

  const {
    nodes = [],
    edges = [],
    activeId = '',
    onselect
  }: {
    nodes?: GraphNode[]
    edges?: GraphEdge[]
    activeId?: string
    onselect?: (id: string) => void
  } = $props()

  const palette: Record<string, string> = {
    violet: '#A78BFA',
    teal: '#5EEAD4',
    amber: '#FBBF24',
    rose: '#FB7185',
    blue: '#60A5FA'
  }
  const byId = $derived(Object.fromEntries(nodes.map((n) => [n.id, n])))
  function col(n: GraphNode) {
    return palette[n.color] ?? n.color ?? '#A78BFA'
  }
  function radius(n: GraphNode) {
    return n.size || (n.kind === 'context' ? 13 : 8)
  }
</script>

<div class="graph">
  <svg class="edges" preserveAspectRatio="none">
    {#each edges as e (e.from + '-' + e.to)}
      {@const a = byId[e.from]}
      {@const b = byId[e.to]}
      {#if a && b}
        {@const on = activeId && (e.from === activeId || e.to === activeId)}
        <line
          x1="{a.x * 100}%"
          y1="{a.y * 100}%"
          x2="{b.x * 100}%"
          y2="{b.y * 100}%"
          stroke={on ? 'var(--purple-glow)' : 'rgba(168,148,255,0.18)'}
          stroke-width={on ? 1.5 : 1}
        />
      {/if}
    {/each}
  </svg>

  {#each nodes as n (n.id)}
    {@const c = col(n)}
    {@const r = radius(n)}
    {@const isActive = n.id === activeId}
    <button
      class="node"
      class:active={isActive}
      style="left:{n.x * 100}%;top:{n.y * 100}%"
      onclick={() => onselect?.(n.id)}
    >
      <span
        class="bead"
        style="
          width:{r * 2}px;height:{r * 2}px;background:{c};
          box-shadow:{isActive
          ? `0 0 0 3px color-mix(in srgb, ${c} 30%, transparent), 0 0 20px ${c}`
          : `0 0 12px color-mix(in srgb, ${c} 60%, transparent)`};"
      ></span>
      {#if n.label}<span class="cap" class:on={isActive}>{n.label}</span>{/if}
    </button>
  {/each}
</div>

<style>
  .graph {
    position: relative;
    width: 100%;
    height: 100%;
    background: radial-gradient(circle at 50% 40%, #161226 0%, var(--bg-base) 70%);
  }
  .edges {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
  }
  .node {
    position: absolute;
    transform: translate(-50%, -50%);
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 6px;
    background: none;
    border: none;
    cursor: pointer;
    padding: 0;
  }
  .bead {
    border-radius: 50%;
    border: 1.5px solid rgba(255, 255, 255, 0.25);
    transition: var(--transition-base);
  }
  .node.active .bead {
    animation: aqno-breathe 2400ms ease-in-out infinite;
  }
  .cap {
    font-family: var(--font-ui);
    font-size: 12px;
    font-weight: var(--weight-medium);
    color: var(--text-2);
    white-space: nowrap;
    background: color-mix(in srgb, var(--bg-base) 70%, transparent);
    padding: 1px 6px;
    border-radius: 6px;
  }
  .cap.on {
    color: var(--text-1);
  }
</style>
