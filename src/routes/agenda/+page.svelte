<script lang="ts">
  import SegmentedControl from '$components/SegmentedControl.svelte'
  import ContextChip from '$components/ContextChip.svelte'
  import Button from '$components/Button.svelte'
  import Icon from '$components/Icon.svelte'
  import { api } from '$lib/api'
  import type { EventInput } from '$lib/api'
  import { app } from '$stores/app'
  import { addDays, fmtLong, monthGrid, todayISO, weekDays } from '$lib/dates'
  import type { Event } from '$lib/types'

  let view = $state('Dia')
  let anchor = $state(todayISO())
  let dayEvents = $state<Event[]>([])
  let rangeEvents = $state<Event[]>([])
  let selected = $state<Event | null>(null)
  let showCreate = $state(false)

  // Create form
  let fTitle = $state('')
  let fContext = $state('')
  let fStart = $state('09:00')
  let fEnd = $state('10:00')
  let fKindLabel = $state('Reunião')
  const kindMap: Record<string, 'event' | 'focus' | 'personal'> = {
    Reunião: 'event',
    Foco: 'focus',
    Pessoal: 'personal'
  }
  let fRecur = $state(false)
  let fDays = $state<string[]>(['MO', 'TU', 'WE', 'TH', 'FR'])
  let quick = $state('')
  let busy = $state(false)

  const DAY_START = 480
  const DAY_END = 1080
  const H = 640
  const hours = Array.from({ length: 11 }, (_, i) => 8 + i)
  const weekdayCodes = [
    { c: 'MO', l: 'S' },
    { c: 'TU', l: 'T' },
    { c: 'WE', l: 'Q' },
    { c: 'TH', l: 'Q' },
    { c: 'FR', l: 'S' },
    { c: 'SA', l: 'S' },
    { c: 'SU', l: 'D' }
  ]
  const palette: Record<string, string> = {
    violet: '#A78BFA',
    teal: '#5EEAD4',
    amber: '#FBBF24',
    rose: '#FB7185',
    blue: '#60A5FA'
  }

  function topOf(m: number) {
    return ((m - DAY_START) / (DAY_END - DAY_START)) * H
  }
  function color(ev: Event) {
    return palette[ev.color] ?? 'var(--text-3)'
  }

  const lanes = $derived.by(() => {
    const m: Record<string, number> = {}
    let n = 0
    for (const e of dayEvents) if (e.conflict) m[e.id ?? ''] = n++
    return m
  })
  const conflicts = $derived(new Set(dayEvents.filter((e) => e.conflict).map((e) => e.id)).size)
  const focusCount = $derived(dayEvents.filter((e) => e.kind === 'focus').length)

  async function loadDay() {
    try {
      const a = await api.agenda(anchor)
      dayEvents = a.events ?? []
    } catch {
      dayEvents = []
    }
  }
  async function loadRange(days: number) {
    try {
      rangeEvents = await api.eventsRange(anchor, addDays(anchor, days))
    } catch {
      rangeEvents = []
    }
  }

  async function reload() {
    if (view === 'Dia') await loadDay()
    else if (view === 'Semana') await loadRange(6)
    else await loadRange(35)
  }

  function shift(dir: number) {
    const step = view === 'Mês' ? 30 : view === 'Semana' ? 7 : 1
    anchor = addDays(anchor, dir * step)
  }

  async function create() {
    if (!fTitle.trim()) return
    busy = true
    const input: EventInput = {
      title: fTitle.trim(),
      context: fContext,
      kind: kindMap[fKindLabel] ?? 'event',
      start: fStart,
      end: fEnd
    }
    if (fRecur) input.rrule = `FREQ=WEEKLY;BYDAY=${fDays.join(',')}`
    else input.date = anchor
    try {
      await api.createEvent(input)
      showCreate = false
      fTitle = ''
      await reload()
    } catch {
      /* ignore */
    }
    busy = false
  }

  async function quickCreate() {
    if (!quick.trim()) return
    busy = true
    try {
      await api.voiceIntent(quick.trim())
      quick = ''
      await reload()
    } catch {
      /* ignore */
    }
    busy = false
  }

  async function remove(ev: Event) {
    if (!ev.id) return
    busy = true
    try {
      await api.deleteEvent(ev.id)
      selected = null
      await reload()
    } catch {
      /* ignore */
    }
    busy = false
  }

  async function cancelOccurrence(ev: Event) {
    if (!ev.id || !ev.date) return
    busy = true
    try {
      await api.cancelOccurrence(ev.id, ev.date)
      selected = null
      await reload()
    } catch {
      /* ignore */
    }
    busy = false
  }

  function toggleDay(c: string) {
    fDays = fDays.includes(c) ? fDays.filter((d) => d !== c) : [...fDays, c]
  }

  // group range events by date for week/month
  const byDate = $derived.by(() => {
    const m: Record<string, Event[]> = {}
    for (const e of rangeEvents) {
      const k = e.date ?? ''
      ;(m[k] ??= []).push(e)
    }
    return m
  })
  const week = $derived(weekDays(anchor))
  const month = $derived(monthGrid(anchor))

  $effect(() => {
    // reload() reads `view` and `anchor`, so the effect tracks both
    reload()
  })
</script>

<div class="page">
  <header class="head">
    <div>
      <div class="overline">{fmtLong(anchor)}</div>
      <h1>Agenda</h1>
    </div>
    <div class="head-right">
      <div class="navbtns">
        <button onclick={() => shift(-1)} aria-label="anterior"
          ><Icon name="chevronLeft" size={16} stroke="var(--text-2)" /></button
        >
        <button onclick={() => shift(1)} aria-label="próximo"
          ><Icon name="chevronRight" size={16} stroke="var(--text-2)" /></button
        >
      </div>
      <SegmentedControl options={['Dia', 'Semana', 'Mês']} bind:value={view} size="sm" />
      <Button variant="primary" size="sm" onclick={() => (showCreate = true)}>
        <Icon name="plus" size={15} stroke="var(--text-on-purple)" /> Novo
      </Button>
    </div>
  </header>

  <div class="body">
    <div class="main-col">
      {#if view === 'Dia'}
        <div class="timeline-card">
          <div class="tl-head">
            <span class="tl-title">{fmtLong(anchor)}</span>
            {#if conflicts > 0}
              <span class="chip danger"><span class="d"></span>{conflicts} conflito(s)</span>
            {/if}
            {#if focusCount > 0}
              <span class="chip purple">{focusCount} bloco de foco</span>
            {/if}
          </div>
          <div class="grid">
            <div class="rail" style="height:{H}px">
              {#each hours as h (h)}
                <div class="hr" style="top:{topOf(h * 60)}px">
                  <span class="hr-label mono">{String(h).padStart(2, '0')}:00</span>
                  <span class="hr-line"></span>
                </div>
              {/each}
              {#each dayEvents as ev (ev.id)}
                {@const top = topOf(ev.startMin ?? 0)}
                {@const height = Math.max(24, topOf(ev.endMin ?? 0) - top)}
                {@const lane = lanes[ev.id ?? '']}
                <button
                  class="ev"
                  class:focus={ev.kind === 'focus'}
                  class:sel={selected?.id === ev.id}
                  onclick={() => (selected = ev)}
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
                </button>
              {/each}
            </div>
          </div>
        </div>
      {:else if view === 'Semana'}
        <div class="week">
          {#each week as d (d.iso)}
            <div class="wday">
              <div class="wday-head">
                <span class="wd-name">{d.label}</span>
                <span class="wd-num mono">{d.num}</span>
              </div>
              {#each byDate[d.iso] ?? [] as ev (ev.id + d.iso)}
                <button
                  class="wev"
                  onclick={() => (selected = ev)}
                  style="border-left-color:{color(ev)}"
                >
                  <span class="wev-time mono">{ev.start}</span>
                  <span class="wev-title">{ev.title}</span>
                </button>
              {/each}
            </div>
          {/each}
        </div>
      {:else}
        <div class="month-card">
          <div class="month-grid head-row">
            {#each ['Seg', 'Ter', 'Qua', 'Qui', 'Sex', 'Sáb', 'Dom'] as w (w)}
              <span class="mh">{w}</span>
            {/each}
          </div>
          <div class="month-grid">
            {#each month as d (d.iso)}
              {@const evs = byDate[d.iso] ?? []}
              <div class="mcell" class:dim={d.dim}>
                <span class="mnum mono">{d.num}</span>
                <div class="mdots">
                  {#each evs.slice(0, 4) as ev (ev.id + d.iso)}
                    <span class="mdot" style="background:{color(ev)}" title={ev.title}></span>
                  {/each}
                </div>
              </div>
            {/each}
          </div>
        </div>
      {/if}
    </div>

    <div class="side">
      <div class="voice-card">
        <div class="overline" style="color:var(--purple-glow);margin-bottom:10px">Criar rápido</div>
        <p class="vc-text">
          Descreva em linguagem natural — ex.: "na Cogna tem daily 9:30 às 10:00".
        </p>
        <div class="quick">
          <input
            bind:value={quick}
            placeholder="ex.: reunião amanhã 15h com a Visa"
            onkeydown={(e) => e.key === 'Enter' && quickCreate()}
          />
          <Button variant="primary" size="sm" onclick={quickCreate} disabled={busy}>Criar</Button>
        </div>
      </div>

      {#if selected}
        <div class="detail-card">
          <div class="overline">Detalhe do evento</div>
          <div class="d-title">{selected.title}</div>
          <div class="d-sub mono">{selected.date} · {selected.start}–{selected.end}</div>
          {#if selected.context}
            <div class="d-row">
              <ContextChip label={selected.context} color={selected.color} size="sm" />
            </div>
          {/if}
          {#if selected.rrule}
            <div class="d-row">
              <Icon name="refresh" size={14} stroke="var(--text-3)" /><span>Evento recorrente</span>
            </div>
          {/if}
          <div class="d-actions">
            {#if selected.rrule}
              <Button variant="subtle" size="sm" onclick={() => cancelOccurrence(selected!)}>
                Cancelar este dia
              </Button>
            {/if}
            <Button variant="danger" size="sm" onclick={() => remove(selected!)}>Excluir</Button>
          </div>
        </div>
      {:else}
        <div class="detail-card empty">
          <Icon name="agenda" size={22} stroke="var(--text-3)" />
          <p>Selecione um evento para ver detalhes e ações.</p>
        </div>
      {/if}
    </div>
  </div>
</div>

{#if showCreate}
  <div
    class="modal-backdrop"
    onclick={() => (showCreate = false)}
    onkeydown={(e) => e.key === 'Escape' && (showCreate = false)}
    role="presentation"
  >
    <div
      class="modal"
      onclick={(e) => e.stopPropagation()}
      onkeydown={(e) => e.stopPropagation()}
      role="dialog"
      aria-modal="true"
      tabindex="-1"
    >
      <div class="modal-head">
        <h2>Novo evento</h2>
        <button class="x" onclick={() => (showCreate = false)} aria-label="fechar">✕</button>
      </div>
      <label class="field">
        <span>Título</span>
        <input bind:value={fTitle} placeholder="Ex.: Daily da Cogna" />
      </label>
      <label class="field">
        <span>Contexto</span>
        <select bind:value={fContext}>
          <option value="">— nenhum —</option>
          {#each $app.contexts as c (c.id)}
            <option value={c.label}>{c.label}</option>
          {/each}
        </select>
      </label>
      <div class="row2">
        <label class="field"><span>Início</span><input type="time" bind:value={fStart} /></label>
        <label class="field"><span>Fim</span><input type="time" bind:value={fEnd} /></label>
      </div>
      <label class="field">
        <span>Tipo</span>
        <SegmentedControl options={['Reunião', 'Foco', 'Pessoal']} bind:value={fKindLabel} full />
      </label>
      <label class="check">
        <input type="checkbox" bind:checked={fRecur} /> Recorrente (semanal)
      </label>
      {#if fRecur}
        <div class="days">
          {#each weekdayCodes as wd (wd.c)}
            <button
              class="day"
              class:on={fDays.includes(wd.c)}
              onclick={() => toggleDay(wd.c)}
              type="button">{wd.l}</button
            >
          {/each}
        </div>
      {/if}
      <div class="modal-actions">
        <Button variant="ghost" onclick={() => (showCreate = false)}>Cancelar</Button>
        <Button variant="primary" onclick={create} disabled={busy || !fTitle.trim()}
          >Criar evento</Button
        >
      </div>
    </div>
  </div>
{/if}

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
    font-size: 12px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--purple-glow);
  }
  h1 {
    font-family: var(--font-display);
    font-size: 28px;
    font-weight: 600;
    letter-spacing: -0.02em;
    margin-top: 6px;
  }
  .head-right {
    display: flex;
    align-items: center;
    gap: 12px;
  }
  .navbtns {
    display: flex;
    gap: 4px;
  }
  .navbtns button {
    width: 32px;
    height: 32px;
    border-radius: 10px;
    border: 1px solid var(--border-strong);
    background: var(--surface-2);
    display: grid;
    place-items: center;
    cursor: pointer;
  }
  .body {
    flex: 1;
    min-height: 0;
    display: flex;
    gap: 22px;
    padding: 18px 36px 100px;
  }
  .main-col {
    flex: 1;
    min-width: 0;
    display: flex;
  }
  .timeline-card {
    flex: 1;
    min-width: 0;
    background: var(--surface-1);
    border: 1px solid var(--border-violet);
    border-radius: var(--radius-card);
    box-shadow: inset 0 1px 0 var(--highlight-top);
    padding: 16px 18px;
    display: flex;
    flex-direction: column;
    overflow-y: auto;
  }
  .tl-head {
    display: flex;
    align-items: center;
    gap: 10px;
    margin-bottom: 12px;
  }
  .tl-title {
    font-family: var(--font-display);
    font-size: 15px;
    font-weight: 600;
    text-transform: capitalize;
  }
  .chip {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    height: 22px;
    padding: 0 9px;
    border-radius: 999px;
    font-size: 11.5px;
    font-weight: 500;
  }
  .chip .d {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: var(--danger);
  }
  .chip.danger {
    background: var(--danger-bg);
    border: 1px solid rgba(248, 113, 113, 0.3);
    color: var(--danger);
  }
  .chip.purple {
    background: var(--purple-012);
    border: 1px solid var(--purple-024);
    color: var(--purple-glow);
  }
  .grid {
    flex: 1;
  }
  .rail {
    position: relative;
    padding-left: 56px;
  }
  .hr {
    position: absolute;
    left: 0;
    right: 0;
  }
  .hr-label {
    position: absolute;
    left: 0;
    top: -7px;
    width: 48px;
    text-align: right;
    font-size: 11px;
    color: var(--text-3);
  }
  .hr-line {
    position: absolute;
    left: 56px;
    right: 0;
    top: 0;
    border-top: 1px solid var(--border-subtle);
  }
  .ev {
    position: absolute;
    border-left: 3px solid;
    border-radius: 9px;
    padding: 6px 12px;
    display: flex;
    flex-direction: column;
    justify-content: center;
    gap: 2px;
    overflow: hidden;
    cursor: pointer;
    text-align: left;
  }
  .ev.sel {
    outline: 2px solid var(--purple-glow);
  }
  .ev-title {
    font-size: 13px;
    font-weight: 500;
    color: var(--text-1);
  }
  .ev-time {
    font-size: 11px;
    color: var(--text-2);
  }
  .conflict-tag {
    position: absolute;
    top: -9px;
    right: 8px;
    display: inline-flex;
    align-items: center;
    height: 18px;
    padding: 0 7px;
    border-radius: 999px;
    background: var(--danger);
    font-size: 10px;
    font-weight: 600;
    color: #1a0606;
  }
  /* week */
  .week {
    flex: 1;
    display: grid;
    grid-template-columns: repeat(7, 1fr);
    gap: 8px;
    overflow-y: auto;
  }
  .wday {
    background: var(--surface-1);
    border: 1px solid var(--border-violet);
    border-radius: var(--radius-md);
    padding: 10px 8px;
    display: flex;
    flex-direction: column;
    gap: 6px;
    min-height: 200px;
  }
  .wday-head {
    display: flex;
    justify-content: space-between;
    align-items: baseline;
    margin-bottom: 4px;
  }
  .wd-name {
    font-size: 11px;
    color: var(--text-3);
    text-transform: capitalize;
  }
  .wd-num {
    font-size: 13px;
    color: var(--text-1);
  }
  .wev {
    border: none;
    border-left: 3px solid;
    background: var(--surface-2);
    border-radius: 6px;
    padding: 5px 7px;
    display: flex;
    flex-direction: column;
    gap: 1px;
    cursor: pointer;
    text-align: left;
  }
  .wev-time {
    font-size: 10px;
    color: var(--text-3);
  }
  .wev-title {
    font-size: 12px;
    color: var(--text-1);
  }
  /* month */
  .month-card {
    flex: 1;
    display: flex;
    flex-direction: column;
    background: var(--surface-1);
    border: 1px solid var(--border-violet);
    border-radius: var(--radius-card);
    padding: 14px;
    overflow-y: auto;
  }
  .month-grid {
    display: grid;
    grid-template-columns: repeat(7, 1fr);
    gap: 6px;
  }
  .head-row {
    margin-bottom: 8px;
  }
  .mh {
    font-size: 11px;
    color: var(--text-3);
    text-align: center;
  }
  .mcell {
    min-height: 76px;
    border: 1px solid var(--border-subtle);
    border-radius: 8px;
    padding: 6px;
    display: flex;
    flex-direction: column;
    gap: 5px;
  }
  .mcell.dim {
    opacity: 0.4;
  }
  .mnum {
    font-size: 12px;
    color: var(--text-2);
  }
  .mdots {
    display: flex;
    flex-wrap: wrap;
    gap: 3px;
  }
  .mdot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
  }
  /* side */
  .side {
    width: 320px;
    flex: none;
    display: flex;
    flex-direction: column;
    gap: 16px;
  }
  .voice-card {
    background: var(--surface-2);
    border: 1px solid var(--purple-024);
    border-radius: var(--radius-card);
    box-shadow:
      var(--shadow-2),
      inset 0 1px 0 var(--highlight-top);
    padding: 18px;
  }
  .vc-text {
    font-size: 13.5px;
    color: var(--text-2);
    line-height: 1.5;
    margin-bottom: 12px;
  }
  .quick {
    display: flex;
    gap: 8px;
  }
  .quick input {
    flex: 1;
    height: 38px;
    padding: 0 12px;
    background: var(--surface-1);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-input);
    color: var(--text-1);
    font-size: 13.5px;
    outline: none;
  }
  .quick input:focus {
    border-color: var(--purple);
  }
  .detail-card {
    background: var(--surface-2);
    border: 1px solid var(--border-violet);
    border-radius: var(--radius-card);
    box-shadow:
      var(--shadow-2),
      inset 0 1px 0 var(--highlight-top);
    padding: 18px;
  }
  .detail-card.empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 10px;
    text-align: center;
    color: var(--text-3);
    font-size: 13px;
  }
  .d-title {
    font-family: var(--font-display);
    font-size: 17px;
    font-weight: 600;
    margin: 10px 0 3px;
  }
  .d-sub {
    font-size: 12px;
    color: var(--text-2);
    margin-bottom: 12px;
  }
  .d-row {
    display: flex;
    align-items: center;
    gap: 9px;
    margin-bottom: 10px;
    font-size: 13.5px;
    color: var(--text-2);
  }
  .d-actions {
    display: flex;
    gap: 8px;
    margin-top: 8px;
  }
  /* modal */
  .modal-backdrop {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.5);
    backdrop-filter: blur(4px);
    display: grid;
    place-items: center;
    z-index: 100;
  }
  .modal {
    width: 440px;
    max-width: calc(100vw - 48px);
    background: var(--surface-2);
    border: 1px solid var(--purple-024);
    border-radius: var(--radius-lg);
    box-shadow: var(--shadow-modal);
    padding: 24px;
    display: flex;
    flex-direction: column;
    gap: 14px;
  }
  .modal-head {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  .modal-head h2 {
    font-family: var(--font-display);
    font-size: 19px;
    font-weight: 600;
  }
  .x {
    background: none;
    border: none;
    color: var(--text-3);
    font-size: 16px;
    cursor: pointer;
  }
  .field {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .field span {
    font-size: 12.5px;
    color: var(--text-2);
  }
  .field input,
  .field select {
    height: 40px;
    padding: 0 12px;
    background: var(--surface-1);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-input);
    color: var(--text-1);
    font-size: 14px;
    outline: none;
  }
  .field input:focus,
  .field select:focus {
    border-color: var(--purple);
  }
  .row2 {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 12px;
  }
  .check {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 13.5px;
    color: var(--text-2);
  }
  .days {
    display: flex;
    gap: 6px;
  }
  .day {
    width: 34px;
    height: 34px;
    border-radius: 8px;
    border: 1px solid var(--border-strong);
    background: var(--surface-1);
    color: var(--text-2);
    cursor: pointer;
    font-size: 13px;
  }
  .day.on {
    background: var(--purple-012);
    border-color: var(--purple);
    color: var(--text-1);
  }
  .modal-actions {
    display: flex;
    justify-content: flex-end;
    gap: 10px;
    margin-top: 6px;
  }
</style>
