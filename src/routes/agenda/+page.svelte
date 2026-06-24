<script lang="ts">
  import { onMount } from 'svelte';
  import SegmentedControl from '$components/SegmentedControl.svelte';
  import ContextChip from '$components/ContextChip.svelte';
  import Button from '$components/Button.svelte';
  import Icon from '$components/Icon.svelte';
  import { api } from '$lib/api';
  import { setVoice } from '$stores/voice';
  import type { Agenda, Event } from '$lib/types';

  let data = $state<Agenda | null>(null);
  let view = $state('Dia');

  const DAY_START = 480;
  const DAY_END = 1080;
  const H = 640;
  const hours = Array.from({ length: 11 }, (_, i) => 8 + i);
  const palette: Record<string, string> = { violet: '#A78BFA', teal: '#5EEAD4', amber: '#FBBF24', rose: '#FB7185' };

  function topOf(m: number) {
    return ((m - DAY_START) / (DAY_END - DAY_START)) * H;
  }
  function color(ev: Event) {
    return palette[ev.color] ?? 'var(--text-3)';
  }
  // assign a lane to conflicting events so they sit side-by-side
  let lanes = $derived.by(() => {
    const m: Record<string, number> = {};
    let n = 0;
    for (const e of data?.events ?? []) if (e.conflict) m[e.id ?? ''] = n++;
    return m;
  });

  onMount(async () => {
    setVoice({ state: 'transcribing', transcript: 'na Cogna tem daily de 9:30 às 10:00, recorrente', level: 0.6 });
    try {
      data = await api.agenda();
    } catch {
      /* offline */
    }
  });
</script>

<div class="page">
  <header class="head">
    <div>
      <div class="overline">{data?.day ?? 'Segunda · 23 jun · 2026'}</div>
      <h1>Agenda</h1>
    </div>
    <div class="head-right">
      <div class="navbtns">
        <button><Icon name="chevronLeft" size={16} stroke="var(--text-2)" /></button>
        <button><Icon name="chevronRight" size={16} stroke="var(--text-2)" /></button>
      </div>
      <SegmentedControl options={['Dia', 'Semana', 'Mês']} bind:value={view} size="sm" />
    </div>
  </header>

  <div class="body">
    <div class="timeline-card">
      <div class="tl-head">
        <span class="tl-title">Hoje</span>
        <span class="chip danger"><span class="d"></span>{data?.conflicts ?? 1} conflito às 14:00</span>
        <span class="chip purple">{data?.focus ?? 1} bloco de foco</span>
      </div>
      <div class="grid">
        <div class="rail" style="height:{H}px">
          {#each hours as h}
            <div class="hr" style="top:{topOf(h * 60)}px">
              <span class="hr-label mono">{String(h).padStart(2, '0')}:00</span>
              <span class="hr-line"></span>
            </div>
          {/each}

          {#each data?.events ?? [] as ev}
            {@const top = topOf(ev.startMin ?? 0)}
            {@const height = topOf(ev.endMin ?? 0) - top}
            {@const lane = lanes[ev.id ?? '']}
            <div
              class="ev"
              class:focus={ev.kind === 'focus'}
              style="
                top:{top}px;height:{height}px;border-left-color:{color(ev)};
                background:{ev.kind === 'focus'
                ? `repeating-linear-gradient(135deg,var(--purple-012) 0 10px,transparent 10px 20px)`
                : ev.kind === 'personal'
                ? 'rgba(255,255,255,0.04)'
                : `color-mix(in srgb, ${color(ev)} 13%, transparent)`};
                left:{lane === undefined ? '8px' : lane === 0 ? '8px' : 'calc(50% + 4px)'};
                right:{lane === undefined ? '8px' : lane === 0 ? 'calc(50% + 4px)' : '8px'};"
            >
              {#if ev.conflict && lane === 0}<span class="conflict-tag">conflito</span>{/if}
              <span class="ev-title">{ev.title}</span>
              <span class="ev-time mono">{ev.start}–{ev.end}</span>
            </div>
          {/each}
        </div>
      </div>
    </div>

    <div class="side">
      <div class="voice-card">
        <div class="vc-head"><span class="overline" style="color:var(--purple-glow)">Criar por voz</span></div>
        <p class="vc-text">"na Cogna tem <b>daily</b> de <b>9:30 às 10:00</b>"</p>
        <div class="struct">
          <div class="struct-head"><span class="struct-title">Daily</span><ContextChip label="Cogna" color="violet" size="sm" /></div>
          <div class="struct-time mono"><Icon name="clock" size={14} stroke="var(--text-3)" /> 09:30–10:00 · recorrente, seg–sex</div>
        </div>
        <div class="vc-actions">
          <Button variant="primary" size="sm" full>Confirmar</Button>
          <Button variant="subtle" size="sm">Editar</Button>
        </div>
      </div>

      <div class="detail-card">
        <div class="overline">Detalhe do evento</div>
        <div class="d-title">Daily da Cogna</div>
        <div class="d-sub mono">Hoje · 09:30–10:00 · Google Meet</div>
        <div class="d-row"><Icon name="users" size={15} stroke="var(--text-3)" /><span>Marina, Téo, Renato · +2</span></div>
        <div class="d-row"><Icon name="file" size={15} stroke="var(--text-3)" /><span>Íris vai gerar o resumo e os itens de ação ao fim.</span></div>
      </div>
    </div>
  </div>
</div>

<style>
  .page { display: flex; flex-direction: column; height: 100%; }
  .head { display: flex; align-items: flex-end; justify-content: space-between; padding: 26px 36px 0; }
  .overline { font-family: var(--font-mono); font-size: 12px; letter-spacing: 0.08em; text-transform: uppercase; color: var(--purple-glow); }
  h1 { font-family: var(--font-display); font-size: 28px; font-weight: 600; letter-spacing: -0.02em; margin-top: 6px; }
  .head-right { display: flex; align-items: center; gap: 12px; }
  .navbtns { display: flex; gap: 4px; }
  .navbtns button { width: 32px; height: 32px; border-radius: 10px; border: 1px solid var(--border-strong); background: var(--surface-2); display: grid; place-items: center; cursor: pointer; }
  .body { flex: 1; min-height: 0; display: flex; gap: 22px; padding: 18px 36px 100px; }
  .timeline-card { flex: 1; min-width: 0; background: var(--surface-1); border: 1px solid var(--border-violet); border-radius: var(--radius-card); box-shadow: inset 0 1px 0 var(--highlight-top); padding: 16px 18px; display: flex; flex-direction: column; }
  .tl-head { display: flex; align-items: center; gap: 10px; margin-bottom: 12px; }
  .tl-title { font-family: var(--font-display); font-size: 15px; font-weight: 600; }
  .chip { display: inline-flex; align-items: center; gap: 6px; height: 22px; padding: 0 9px; border-radius: 999px; font-family: var(--font-ui); font-size: 11.5px; font-weight: 500; }
  .chip .d { width: 6px; height: 6px; border-radius: 50%; background: var(--danger); }
  .chip.danger { background: var(--danger-bg); border: 1px solid rgba(248, 113, 113, 0.3); color: var(--danger); }
  .chip.purple { background: var(--purple-012); border: 1px solid var(--purple-024); color: var(--purple-glow); }
  .grid { flex: 1; }
  .rail { position: relative; padding-left: 56px; }
  .hr { position: absolute; left: 0; right: 0; }
  .hr-label { position: absolute; left: 0; top: -7px; width: 48px; text-align: right; font-size: 11px; color: var(--text-3); }
  .hr-line { position: absolute; left: 56px; right: 0; top: 0; border-top: 1px solid var(--border-subtle); }
  .ev { position: absolute; border-left: 3px solid; border-radius: 9px; padding: 6px 12px; display: flex; flex-direction: column; justify-content: center; gap: 2px; overflow: hidden; }
  .ev-title { font-family: var(--font-ui); font-size: 13px; font-weight: 500; color: var(--text-1); }
  .ev-time { font-size: 11px; color: var(--text-2); }
  .conflict-tag { position: absolute; top: -9px; right: 8px; display: inline-flex; align-items: center; height: 18px; padding: 0 7px; border-radius: 999px; background: var(--danger); font-family: var(--font-ui); font-size: 10px; font-weight: 600; color: #1a0606; }
  .side { width: 344px; flex: none; display: flex; flex-direction: column; gap: 16px; }
  .voice-card { background: var(--surface-2); border: 1px solid var(--purple-024); border-radius: var(--radius-card); box-shadow: var(--shadow-2), inset 0 1px 0 var(--highlight-top), var(--glow-md); padding: 18px; }
  .vc-head { margin-bottom: 12px; }
  .overline { font-family: var(--font-mono); font-size: 10px; letter-spacing: 0.08em; text-transform: uppercase; color: var(--text-3); }
  .vc-text { font-size: 14.5px; color: var(--text-2); line-height: 1.5; margin-bottom: 14px; }
  .vc-text b { color: var(--text-1); }
  .struct { padding: 13px; border-radius: var(--radius-md); background: var(--surface-1); border: 1px solid var(--border-violet); }
  .struct-head { display: flex; align-items: center; justify-content: space-between; margin-bottom: 8px; }
  .struct-title { font-family: var(--font-display); font-size: 15px; font-weight: 600; }
  .struct-time { display: flex; align-items: center; gap: 8px; font-size: 12px; color: var(--text-2); }
  .vc-actions { display: flex; gap: 8px; margin-top: 13px; }
  .detail-card { background: var(--surface-2); border: 1px solid var(--border-violet); border-radius: var(--radius-card); box-shadow: var(--shadow-2), inset 0 1px 0 var(--highlight-top); padding: 18px; }
  .d-title { font-family: var(--font-display); font-size: 17px; font-weight: 600; margin: 10px 0 3px; }
  .d-sub { font-size: 12px; color: var(--text-2); margin-bottom: 14px; }
  .d-row { display: flex; align-items: flex-start; gap: 9px; margin-bottom: 10px; font-size: 13.5px; color: var(--text-2); }
</style>
