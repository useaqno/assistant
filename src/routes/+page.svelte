<script lang="ts">
  import { onMount } from 'svelte'
  import Card from '$components/Card.svelte'
  import Presence from '$components/Presence.svelte'
  import ContextChip from '$components/ContextChip.svelte'
  import Badge from '$components/Badge.svelte'
  import Icon from '$components/Icon.svelte'
  import { api } from '$lib/api'
  import { presence } from '$stores/presence'
  import { app, companionName } from '$stores/app'
  import { startListening } from '$lib/voice'
  import type { Task, TodayBrief } from '$lib/types'

  let brief = $state<TodayBrief | null>(null)

  const companion = $derived(companionName($app.persona))
  const ownerInitial = $derived(($app.persona.owner || 'A').charAt(0).toUpperCase())
  const doneCount = $derived((brief?.taskList ?? []).filter((t) => t.done).length)

  async function load() {
    try {
      brief = await api.today()
    } catch {
      /* daemon offline */
    }
  }

  async function toggle(t: Task) {
    if (!t.id) return
    const done = !t.done
    t.done = done
    try {
      await api.setTaskDone(t.id, done)
    } catch {
      t.done = !done
    }
  }

  onMount(load)
</script>

<div class="scroll">
  <header class="head">
    <div>
      <div class="overline" style="color:var(--purple-glow)">
        {brief?.date ?? ''}
      </div>
      <h1>{brief?.greeting ?? `Olá, ${$app.persona.owner || ''}`}</h1>
    </div>
    <div class="head-right">
      <div class="search"><Icon name="search" size={15} /> Buscar <kbd>⌘K</kbd></div>
      <div class="avatar">{ownerInitial}</div>
    </div>
  </header>

  <section class="hero">
    <Presence state={$presence.state} size={150} level={$presence.level} />
    <div class="hero-name">
      <span class="n">{companion}</span>
      <span class="tag">Pronto</span>
    </div>
    <p class="hero-line">
      {brief?.headline ?? `Olá! Sou ${companion}, sua companheira. Quer que eu prepare seu dia?`}
    </p>
    <div class="cta-row">
      <button class="cta" onclick={() => startListening()}>
        <Icon name="mic" size={20} /> Falar com {companion}
      </button>
      <span class="hint">ou segure <kbd>espaço</kbd></span>
    </div>
  </section>

  <section class="cards">
    <Card padding={18}>
      <div class="c-head">
        <span class="overline">Próximo evento</span>
      </div>
      <div class="c-title">{brief?.nextEvent.title ?? '—'}</div>
      <div class="c-sub mono">
        {brief?.nextEvent.start ?? '--:--'} — {brief?.nextEvent.end ?? '--:--'}
      </div>
      {#if brief?.nextEvent.context}
        <ContextChip
          label={brief.nextEvent.context}
          color={brief.nextEvent.color || 'violet'}
          size="sm"
        />
      {/if}
    </Card>

    <Card padding={18}>
      <div class="c-head">
        <span class="overline">Tarefas de hoje</span>
        <span class="mono mut">{doneCount} / {brief?.taskList?.length ?? 0}</span>
      </div>
      {#each brief?.taskList ?? [] as t (t.id ?? t.title)}
        <button class="task" onclick={() => toggle(t)}>
          <span class="box" class:done={t.done}>
            {#if t.done}<Icon
                name="check"
                size={9}
                stroke="var(--success)"
                strokeWidth={3.4}
              />{/if}
          </span>
          <span class="t-label" class:done={t.done}>{t.title}</span>
        </button>
      {/each}
      {#if (brief?.taskList?.length ?? 0) === 0}
        <span class="empty">Nenhuma tarefa por enquanto.</span>
      {/if}
    </Card>

    <Card padding={18} glow>
      <div class="c-head">
        <span class="overline" style="color:var(--purple-glow)"
          ><Icon name="sparkles" size={13} stroke="var(--purple-glow)" /> Conselho do mentor</span
        >
      </div>
      <p class="mentor">
        {brief?.mentor ??
          'Seu bloco de foco das 11h está livre. Proteja-o para avançar a proposta da Visa antes da call das 14h.'}
      </p>
    </Card>
  </section>

  <section class="recent">
    <div class="overline" style="margin-bottom:6px">Conversas recentes</div>
    {#each brief?.recent ?? [] as r, i (r.title)}
      <div class="rrow" class:divider={i < (brief?.recent.length ?? 0) - 1}>
        <span
          class="rdot"
          style="background:var(--data-{r.color ||
            'violet'});box-shadow:0 0 8px var(--data-{r.color || 'violet'})"
        ></span>
        <div class="rtitle">{r.title}</div>
        {#if r.context}<ContextChip label={r.context} color={r.color} size="sm" />{/if}
        {#if r.tag}<Badge tone={r.tone || 'neutral'} size="sm">{r.tag}</Badge>{/if}
        <span class="rwhen mono">{r.when}</span>
      </div>
    {/each}
  </section>
</div>

<style>
  .scroll {
    flex: 1;
    overflow-y: auto;
    padding: 26px 36px 120px;
    background:
      radial-gradient(circle at 50% 18%, rgba(139, 92, 246, 0.1), transparent 55%), var(--bg-base);
  }
  .head {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
  }
  h1 {
    font-family: var(--font-display);
    font-size: 28px;
    font-weight: 600;
    letter-spacing: -0.02em;
    margin-top: 6px;
  }
  .overline {
    font-family: var(--font-mono);
    font-size: 12px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--text-3);
    display: inline-flex;
    align-items: center;
    gap: 6px;
  }
  .head-right {
    display: flex;
    align-items: center;
    gap: 10px;
  }
  .search {
    display: flex;
    align-items: center;
    gap: 8px;
    height: 38px;
    padding: 0 12px;
    border-radius: 999px;
    background: var(--surface-2);
    border: 1px solid var(--border-violet);
    color: var(--text-2);
    font-size: 13px;
  }
  kbd {
    font-family: var(--font-mono);
    font-size: 11px;
    padding: 2px 6px;
    border-radius: 6px;
    background: var(--surface-1);
    border: 1px solid var(--border-strong);
    color: var(--text-2);
  }
  .avatar {
    width: 38px;
    height: 38px;
    border-radius: 999px;
    background: linear-gradient(135deg, #3a2f55, #241d36);
    border: 1px solid var(--border-violet);
    display: grid;
    place-items: center;
    font-family: var(--font-display);
    font-weight: 600;
  }
  .hero {
    display: flex;
    flex-direction: column;
    align-items: center;
    text-align: center;
    padding: 14px 0 6px;
  }
  .hero-name {
    display: flex;
    align-items: center;
    gap: 10px;
    margin-top: 14px;
  }
  .hero-name .n {
    font-family: var(--font-display);
    font-size: 22px;
    font-weight: 600;
    letter-spacing: -0.02em;
  }
  .hero-name .tag {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--purple-glow);
    padding: 3px 8px;
    border-radius: 999px;
    background: var(--purple-012);
    border: 1px solid var(--purple-024);
  }
  .hero-line {
    font-size: 16px;
    color: var(--text-2);
    margin-top: 8px;
    max-width: 440px;
  }
  .cta-row {
    display: flex;
    align-items: center;
    gap: 14px;
    margin-top: 16px;
  }
  .cta {
    display: inline-flex;
    align-items: center;
    gap: 10px;
    height: 52px;
    padding: 0 26px;
    border: none;
    border-radius: 999px;
    background: var(--grad-active);
    color: var(--text-on-purple);
    font-family: var(--font-ui);
    font-size: 16px;
    font-weight: 600;
    cursor: pointer;
    box-shadow:
      var(--glow-md),
      inset 0 1px 0 rgba(255, 255, 255, 0.25);
  }
  .hint {
    font-size: 13px;
    color: var(--text-3);
  }
  .cards {
    display: grid;
    grid-template-columns: 1fr 1fr 1fr;
    gap: 16px;
    margin-top: 18px;
  }
  .c-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 10px;
  }
  .mut {
    color: var(--text-2);
  }
  .c-title {
    font-family: var(--font-display);
    font-size: 17px;
    font-weight: 600;
  }
  .c-sub {
    font-size: 13px;
    color: var(--text-2);
    margin: 4px 0 12px;
  }
  .task {
    display: flex;
    align-items: center;
    gap: 9px;
    margin-bottom: 7px;
    width: 100%;
    padding: 2px 0;
    background: none;
    border: none;
    text-align: left;
    cursor: pointer;
  }
  .empty {
    font-size: 13px;
    color: var(--text-3);
  }
  .box {
    width: 15px;
    height: 15px;
    border-radius: 5px;
    border: 1.5px solid var(--border-strong);
    display: grid;
    place-items: center;
  }
  .box.done {
    border-color: var(--success);
  }
  .t-label {
    font-size: 13.5px;
    color: var(--text-1);
  }
  .t-label.done {
    color: var(--text-3);
    text-decoration: line-through;
  }
  .mentor {
    font-size: 14.5px;
    color: var(--text-1);
    line-height: 1.5;
  }
  .recent {
    margin-top: 18px;
  }
  .rrow {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 9px 4px;
  }
  .rrow.divider {
    border-bottom: 1px solid var(--border-subtle);
  }
  .rdot {
    width: 9px;
    height: 9px;
    border-radius: 50%;
    flex: none;
  }
  .rtitle {
    flex: 1;
    min-width: 0;
    font-family: var(--font-ui);
    font-size: 14.5px;
    font-weight: 500;
    color: var(--text-1);
  }
  .rwhen {
    font-size: 11px;
    color: var(--text-3);
  }
</style>
