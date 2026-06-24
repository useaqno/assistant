<script lang="ts">
  import { onMount } from 'svelte';
  import GraphView from '$components/GraphView.svelte';
  import Button from '$components/Button.svelte';
  import Icon from '$components/Icon.svelte';
  import { api } from '$lib/api';
  import { setVoice } from '$stores/voice';
  import type { Graph, GraphNode } from '$lib/types';

  let g = $state<Graph | null>(null);
  let active = $state('cogna');

  const palette: Record<string, string> = { violet: '#A78BFA', teal: '#5EEAD4', amber: '#FBBF24', rose: '#FB7185', blue: '#60A5FA' };
  const kindLabel: Record<string, string> = { context: 'contexto', project: 'projeto', event: 'evento', task: 'tarefa', person: 'pessoa', decision: 'decisão' };

  let activeNode = $derived(g?.nodes.find((n) => n.id === active));
  let neighbors = $derived.by<GraphNode[]>(() => {
    if (!g) return [];
    const ids = new Set<string>();
    for (const e of g.edges) {
      if (e.from === active) ids.add(e.to);
      if (e.to === active) ids.add(e.from);
    }
    return g.nodes.filter((n) => ids.has(n.id) && n.id !== 'iris');
  });
  function col(c: string) {
    return palette[c] ?? c;
  }

  onMount(async () => {
    setVoice({ state: 'thinking', transcript: 'Conectando proposta Q3 à decisão de tokenização da Visa…', level: 0.5 });
    try {
      g = await api.graph();
    } catch {
      /* offline */
    }
  });
</script>

<div class="page">
  <header class="head">
    <div>
      <div class="overline">142 nós · 318 conexões</div>
      <h1>Rede neural</h1>
    </div>
    <div class="search"><Icon name="search" size={15} stroke="var(--text-2)" /> Buscar entidade</div>
  </header>

  <div class="canvas-wrap">
    <div class="canvas">
      <GraphView nodes={g?.nodes ?? []} edges={g?.edges ?? []} activeId={active} onselect={(id) => (active = id)} />

      <div class="filters">
        <span class="fchip on"><span class="d" style="background:#0C0A14"></span>Cogna</span>
        <span class="fchip glass">Todos os contextos</span>
      </div>

      {#if activeNode}
        <div class="detail">
          <div class="d-head"><span class="d-bead" style="background:{col(activeNode.color)};box-shadow:0 0 12px {col(activeNode.color)}"></span><span class="d-name">{activeNode.label}</span></div>
          <div class="overline">{activeNode.kind === 'context' ? 'Empresa · contexto' : kindLabel[activeNode.kind] ?? activeNode.kind}</div>
          <p class="d-desc">38 memórias · 12 eventos · 5 pessoas. Última atividade há 12 min.</p>
          <div class="overline">Conexões fortes</div>
          <div class="conns">
            {#each neighbors.slice(0, 4) as n}
              <div class="conn"><span class="cb" style="background:{col(n.color)}"></span><span class="cn">{n.label}</span><span class="ck mono">{kindLabel[n.kind] ?? n.kind}</span></div>
            {/each}
          </div>
          <Button variant="primary" size="sm" full>Abrir contexto</Button>
        </div>
      {/if}

      <div class="zoom">
        <button aria-label="zoom in"><Icon name="plus" size={16} stroke="var(--text-2)" /></button>
        <button aria-label="zoom out"><Icon name="minus" size={16} stroke="var(--text-2)" /></button>
        <button aria-label="ajustar"><Icon name="expand" size={15} stroke="var(--text-2)" /></button>
      </div>

      <div class="legend">
        <span><span class="ld" style="background:#A78BFA"></span>Contexto</span>
        <span><span class="ld" style="background:#5EEAD4"></span>Projeto</span>
        <span><span class="ld" style="background:#FBBF24"></span>Decisão</span>
        <span><span class="ld" style="background:#FB7185"></span>Tarefa</span>
        <span><span class="ld" style="background:#60A5FA"></span>Pessoa</span>
      </div>
    </div>
  </div>
</div>

<style>
  .page { display: flex; flex-direction: column; height: 100%; }
  .head { display: flex; align-items: flex-end; justify-content: space-between; padding: 26px 36px 0; }
  .overline { font-family: var(--font-mono); font-size: 10px; letter-spacing: 0.08em; text-transform: uppercase; color: var(--text-3); }
  .head .overline { font-size: 12px; color: var(--purple-glow); }
  h1 { font-family: var(--font-display); font-size: 28px; font-weight: 600; letter-spacing: -0.02em; margin-top: 6px; }
  .search { display: inline-flex; align-items: center; gap: 7px; height: 34px; padding: 0 13px; border-radius: 999px; background: var(--surface-2); border: 1px solid var(--border-violet); font-size: 13px; color: var(--text-2); }
  .canvas-wrap { flex: 1; min-height: 0; padding: 18px 36px 100px; }
  .canvas { position: relative; width: 100%; height: 100%; border-radius: var(--radius-lg); overflow: hidden; border: 1px solid var(--border-violet); box-shadow: inset 0 1px 0 var(--highlight-top); }
  .filters { position: absolute; top: 16px; left: 16px; display: flex; gap: 7px; }
  .fchip { display: inline-flex; align-items: center; gap: 6px; height: 28px; padding: 0 11px; border-radius: 999px; font-family: var(--font-ui); font-size: 12.5px; font-weight: 500; }
  .fchip.on { background: #a78bfa; color: #0c0a14; }
  .fchip.on .d { width: 7px; height: 7px; border-radius: 50%; }
  .fchip.glass { background: color-mix(in srgb, var(--surface-3) 80%, transparent); backdrop-filter: var(--blur-subtle); -webkit-backdrop-filter: var(--blur-subtle); border: 1px solid var(--border-strong); color: var(--text-2); }
  .detail { position: absolute; top: 16px; right: 16px; width: 284px; padding: 18px; border-radius: var(--radius-card); background: color-mix(in srgb, var(--surface-3) 90%, transparent); backdrop-filter: var(--blur-panel); -webkit-backdrop-filter: var(--blur-panel); border: 1px solid var(--purple-024); box-shadow: var(--shadow-modal), inset 0 1px 0 var(--highlight-top); }
  .d-head { display: flex; align-items: center; gap: 9px; margin-bottom: 12px; }
  .d-bead { width: 14px; height: 14px; border-radius: 50%; border: 1.5px solid rgba(255, 255, 255, 0.3); }
  .d-name { font-family: var(--font-display); font-size: 17px; font-weight: 600; }
  .d-desc { font-size: 13px; color: var(--text-2); line-height: 1.5; margin: 4px 0 14px; }
  .conns { display: flex; flex-direction: column; gap: 8px; margin: 9px 0 14px; }
  .conn { display: flex; align-items: center; gap: 9px; }
  .cb { width: 7px; height: 7px; border-radius: 50%; }
  .cn { font-family: var(--font-ui); font-size: 13px; color: var(--text-1); }
  .ck { margin-left: auto; font-size: 11px; color: var(--text-3); }
  .zoom { position: absolute; bottom: 16px; right: 16px; display: flex; flex-direction: column; gap: 6px; }
  .zoom button { width: 34px; height: 34px; border-radius: 9px; background: color-mix(in srgb, var(--surface-3) 85%, transparent); backdrop-filter: var(--blur-subtle); -webkit-backdrop-filter: var(--blur-subtle); border: 1px solid var(--border-strong); display: grid; place-items: center; cursor: pointer; }
  .legend { position: absolute; bottom: 16px; left: 16px; display: flex; align-items: center; gap: 14px; padding: 9px 14px; border-radius: 999px; background: color-mix(in srgb, var(--surface-3) 80%, transparent); backdrop-filter: var(--blur-subtle); -webkit-backdrop-filter: var(--blur-subtle); border: 1px solid var(--border-violet); font-size: 12px; color: var(--text-3); }
  .legend span { display: flex; align-items: center; gap: 6px; }
  .ld { width: 7px; height: 7px; border-radius: 50%; }
</style>
