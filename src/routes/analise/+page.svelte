<script lang="ts">
  import { onMount } from 'svelte'
  import Card from '$components/Card.svelte'
  import Presence from '$components/Presence.svelte'
  import MetricRing from '$components/MetricRing.svelte'
  import SegmentedControl from '$components/SegmentedControl.svelte'
  import Button from '$components/Button.svelte'
  import Icon from '$components/Icon.svelte'
  import { api } from '$lib/api'
  import { fmtLong, todayISO } from '$lib/dates'
  import type { Analysis, AppHealth } from '$lib/types'

  let a = $state<Analysis | null>(null)
  let view = $state('Hoje')

  const today = fmtLong(todayISO())

  const dataColor: Record<string, string> = {
    violet: 'var(--data-violet)',
    teal: 'var(--data-teal)',
    amber: 'var(--data-amber)',
    rose: 'var(--data-rose)',
    blue: 'var(--data-blue)'
  }

  function spark(s: number[]) {
    const w = 84
    const max = Math.max(...s, 1)
    return s.map((v, i) => `${(i / (s.length - 1)) * w},${22 - (v / max) * 16 - 2}`).join(' ')
  }
  function appColor(app: AppHealth) {
    return app.status === 'warn' ? 'var(--data-amber)' : 'var(--data-teal)'
  }
  function appDot(app: AppHealth) {
    return app.status === 'warn' ? 'var(--warning)' : 'var(--success)'
  }

  onMount(async () => {
    try {
      a = await api.analysis()
    } catch {
      /* offline */
    }
  })
</script>

<div class="page">
  <header class="head">
    <div>
      <div class="overline">{today}</div>
      <h1>Briefing diário</h1>
    </div>
    <SegmentedControl options={['Hoje', 'Semana', 'Mês']} bind:value={view} size="sm" />
  </header>

  <div class="grid">
    <div class="span2">
      <Card padding={18} glow>
        <div class="c-head">
          <Presence state="speaking" size={30} level={0.7} /><span class="overline pp"
            >Íris · resumo do dia</span
          >
        </div>
        <p class="summary">
          {a?.summary ??
            'Bom dia, Renato. Hoje você tem 4 reuniões e 5 tarefas. A manhã está mais leve — protegi seu bloco de foco das 11h.'}
        </p>
        <div class="stats">
          <div>
            <div class="big">{a?.meetings ?? 4}</div>
            <div class="lbl">reuniões</div>
          </div>
          <div>
            <div class="big">{a?.tasks ?? 5}</div>
            <div class="lbl">tarefas</div>
          </div>
          <div>
            <div class="big" style="color:var(--success)">{a?.focusFree ?? '3h'}</div>
            <div class="lbl">foco livre</div>
          </div>
          <div>
            <div class="big">{a?.contexts ?? 6}</div>
            <div class="lbl">contextos</div>
          </div>
        </div>
      </Card>
    </div>

    <Card padding={18}>
      <div class="overline mb">Foco vs reuniões</div>
      <div class="ring-center">
        <MetricRing
          value={a?.focusShare ?? 0.6}
          size={96}
          label="{Math.round((a?.focusShare ?? 0.6) * 100)}%"
          caption="foco"
          color="var(--data-teal)"
        />
      </div>
      <div class="legend">
        <span><span class="sq" style="background:var(--data-teal)"></span>Foco</span><span
          class="mono">3h 00</span
        >
      </div>
      <div class="legend">
        <span><span class="sq" style="background:var(--surface-3)"></span>Reuniões</span><span
          class="mono">2h 00</span
        >
      </div>
    </Card>

    <Card padding={18}>
      <div class="overline mb">Tarefas / afazeres</div>
      <div class="ring-center">
        <MetricRing
          value={a?.tasksRatio ?? 0.4}
          size={96}
          label={a?.tasksDone ?? '2/5'}
          caption="concluídas"
          color="var(--data-amber)"
        />
      </div>
      <div class="legend">
        <span>Vencendo hoje</span><span class="mono" style="color:var(--warning)">2</span>
      </div>
      <div class="legend">
        <span>Atrasadas</span><span class="mono" style="color:var(--danger)">1</span>
      </div>
    </Card>

    <div class="span2">
      <Card padding={18}>
        <div class="c-head between">
          <span class="overline">Saúde das aplicações</span><span class="chip ok"
            ><span class="d"></span>operacional</span
          >
        </div>
        {#each a?.apps ?? [] as app, i (app.name)}
          <div class="app-row" class:divider={i < (a?.apps.length ?? 0) - 1}>
            <span class="adot" style="background:{appDot(app)};box-shadow:0 0 8px {appDot(app)}"
            ></span>
            <span class="aname">{app.name}</span>
            <svg
              width="84"
              height="22"
              viewBox="0 0 84 22"
              fill="none"
              stroke={appColor(app)}
              stroke-width="1.6"><polyline points={spark(app.spark)} /></svg
            >
            <span
              class="alat mono"
              style="color:{app.status === 'warn' ? 'var(--warning)' : 'var(--text-2)'}"
              >{app.latency}</span
            >
          </div>
        {/each}
      </Card>
    </div>

    <Card padding={18}>
      <div class="overline mb">Finanças e gestão</div>
      <div class="cash mono">{a?.cashMonth ?? 'R$ 48,2k'}</div>
      <div class="delta">
        <span class="mono" style="color:var(--success)">{a?.cashDelta ?? '▲ 12%'}</span><span
          >vs. mês passado</span
        >
      </div>
      <div class="bars">
        {#each a?.cashBars ?? [46, 62, 54, 78, 70, 100] as b, i (i)}
          <div
            class="bar"
            class:last={i === (a?.cashBars.length ?? 6) - 1}
            style="height:{b}%"
          ></div>
        {/each}
      </div>
    </Card>

    <Card padding={18}>
      <div class="overline mb">Vida pessoal</div>
      <div class="personal">
        {#each a?.personal ?? [{ label: 'sono', value: 0.85, big: '7h', color: 'violet' }, { label: 'água', value: 0.55, big: '1.4L', color: 'blue' }, { label: 'passos', value: 0.62, big: '6.2k', color: 'teal' }] as m (m.label)}
          <div class="p-item">
            <MetricRing
              value={m.value}
              size={52}
              label={m.big}
              color={dataColor[m.color] ?? 'var(--data-violet)'}
              glow={false}
            /><span>{m.label}</span>
          </div>
        {/each}
      </div>
    </Card>

    <div class="span4 mentor">
      <span class="m-icon"><Icon name="sparkles" size={20} stroke="var(--purple-glow)" /></span>
      <div class="m-body">
        <div class="m-title">{a?.mentorTitle ?? 'Conselho do mentor'}</div>
        <p>
          {a?.mentorBody ??
            'Reuniões somam 40% do dia. Considere mover a sync da Devlith para quinta e blindar a manhã.'}
        </p>
      </div>
      <Button variant="primary" size="sm">Reagendar com a Íris</Button>
    </div>
  </div>
</div>

<style>
  .page {
    display: flex;
    flex-direction: column;
    height: 100%;
  }
  .head {
    display: flex;
    align-items: flex-end;
    justify-content: space-between;
    padding: 26px 36px 0;
  }
  .overline {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--text-3);
    display: inline-flex;
    align-items: center;
    gap: 6px;
  }
  .head .overline {
    font-size: 12px;
    color: var(--purple-glow);
  }
  h1 {
    font-family: var(--font-display);
    font-size: 28px;
    font-weight: 600;
    letter-spacing: -0.02em;
    margin-top: 6px;
  }
  .grid {
    flex: 1;
    min-height: 0;
    overflow-y: auto;
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    grid-auto-rows: min-content;
    gap: 16px;
    padding: 18px 36px 100px;
  }
  .span2 {
    grid-column: span 2;
  }
  .span4 {
    grid-column: span 4;
  }
  .mb {
    margin-bottom: 14px;
  }
  .pp {
    color: var(--purple-glow);
  }
  .c-head {
    display: flex;
    align-items: center;
    gap: 10px;
    margin-bottom: 12px;
  }
  .c-head.between {
    justify-content: space-between;
  }
  .summary {
    font-size: 15px;
    color: var(--text-1);
    line-height: 1.55;
    margin-bottom: 14px;
  }
  .stats {
    display: flex;
    gap: 22px;
  }
  .big {
    font-family: var(--font-mono);
    font-size: 22px;
    font-weight: 600;
    letter-spacing: -0.02em;
  }
  .lbl {
    font-size: 12px;
    color: var(--text-3);
  }
  .ring-center {
    display: flex;
    justify-content: center;
  }
  .legend {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-top: 8px;
    font-size: 13px;
    color: var(--text-2);
  }
  .legend span {
    display: inline-flex;
    align-items: center;
    gap: 7px;
  }
  .sq {
    width: 8px;
    height: 8px;
    border-radius: 2px;
  }
  .mono {
    font-family: var(--font-mono);
    font-variant-numeric: tabular-nums;
  }
  .chip {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    height: 20px;
    padding: 0 8px;
    border-radius: 999px;
    font-family: var(--font-ui);
    font-size: 11px;
    font-weight: 500;
  }
  .chip.ok {
    background: var(--success-bg);
    border: 1px solid rgba(74, 222, 128, 0.3);
    color: var(--success);
  }
  .chip .d {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: var(--success);
  }
  .app-row {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 9px 0;
  }
  .app-row.divider {
    border-bottom: 1px solid var(--border-subtle);
  }
  .adot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
  }
  .aname {
    flex: 1;
    font-family: var(--font-ui);
    font-size: 13.5px;
    color: var(--text-1);
  }
  .alat {
    font-size: 12px;
    width: 54px;
    text-align: right;
  }
  .cash {
    font-size: 24px;
    font-weight: 600;
    letter-spacing: -0.02em;
  }
  .delta {
    display: flex;
    align-items: center;
    gap: 6px;
    margin: 4px 0 12px;
    font-size: 12px;
    color: var(--text-3);
  }
  .bars {
    display: flex;
    align-items: flex-end;
    gap: 7px;
    height: 42px;
  }
  .bar {
    flex: 1;
    border-radius: 3px;
    background: var(--surface-3);
  }
  .bar.last {
    background: var(--grad-active);
    box-shadow: var(--glow-sm);
  }
  .personal {
    display: flex;
    justify-content: space-between;
  }
  .p-item {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 6px;
    font-size: 11.5px;
    color: var(--text-3);
  }
  .mentor {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 16px 20px;
    border-radius: var(--radius-card);
    background: linear-gradient(120deg, var(--purple-012), transparent 70%), var(--surface-2);
    border: 1px solid var(--purple-024);
    box-shadow:
      var(--shadow-2),
      inset 0 1px 0 var(--highlight-top);
  }
  .m-icon {
    width: 40px;
    height: 40px;
    flex: none;
    border-radius: 12px;
    background: var(--purple-012);
    border: 1px solid var(--purple-024);
    display: grid;
    place-items: center;
  }
  .m-body {
    flex: 1;
  }
  .m-title {
    font-family: var(--font-display);
    font-size: 14px;
    font-weight: 600;
    margin-bottom: 2px;
  }
  .m-body p {
    font-size: 13.5px;
    color: var(--text-2);
    line-height: 1.5;
  }
</style>
