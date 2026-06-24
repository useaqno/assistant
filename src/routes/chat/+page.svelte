<script lang="ts">
  import { onMount, tick } from 'svelte'
  import ChatBubble from '$components/ChatBubble.svelte'
  import Presence from '$components/Presence.svelte'
  import Icon from '$components/Icon.svelte'
  import { api, streamChat } from '$lib/api'
  import { startListening } from '$lib/voice'
  import { app, companionName } from '$stores/app'
  import type { ChatMessage } from '$lib/types'

  let thread = $state<ChatMessage[]>([])
  let draft = $state('')
  let busy = $state(false)
  let listEl: HTMLDivElement

  const companion = $derived(companionName($app.persona))

  async function send() {
    const text = draft.trim()
    if (!text || busy) return
    busy = true
    thread = [...thread, { id: `u${Date.now()}`, from: 'user', text, time: 'agora' }]
    draft = ''
    await scrollDown()

    const replyId = `a${Date.now()}`
    thread = [...thread, { id: replyId, from: 'aqno', text: '', time: 'agora', streaming: true }]
    const idx = thread.length - 1

    try {
      await streamChat(
        text,
        undefined,
        (chunk) => {
          thread[idx] = { ...thread[idx], text: thread[idx].text + chunk }
          scrollDown()
        },
        (final) => {
          thread[idx] = { ...final, streaming: false }
          busy = false
          scrollDown()
        },
        () => {
          busy = false
        }
      )
    } catch {
      // Fallback to non-streaming send if SSE is unavailable.
      try {
        const reply = await api.sendChat(text)
        thread[idx] = reply
      } catch {
        thread = thread.filter((_, i) => i !== idx)
      }
      busy = false
      await scrollDown()
    }
  }

  async function scrollDown() {
    await tick()
    listEl?.scrollTo({ top: listEl.scrollHeight, behavior: 'smooth' })
  }

  onMount(async () => {
    try {
      thread = await api.chat()
    } catch {
      /* offline */
    }
    await scrollDown()
  })
</script>

<div class="page">
  <header class="head">
    <div class="h-left">
      <h1>Conversa com {companion}</h1>
      <div class="ctx">
        <span class="cchip purple"
          ><Icon name="clock" size={12} stroke="var(--purple-glow)" />Sabe do seu dia</span
        >
        <span class="cchip mut">{thread.length} mensagens</span>
      </div>
    </div>
    <div class="corner">
      <Presence state="idle" size={46} level={0.6} />
      <span class="corner-lbl">{companion}</span>
    </div>
  </header>

  <div class="thread" bind:this={listEl}>
    {#each thread as m (m.id)}
      <ChatBubble
        from={m.from}
        name={m.from === 'aqno' ? companion : ''}
        time={m.time}
        streaming={m.streaming}
      >
        <div>{m.text}</div>
        {#if m.ref?.kind === 'memory'}
          <div class="ref">
            <Icon name="sparkles" size={14} stroke="var(--text-3)" /><span>{m.ref.label}</span>
          </div>
        {:else if m.ref?.kind === 'action'}
          <div class="ref action">
            <Icon name="check" size={15} stroke="var(--success)" strokeWidth={2.4} /><span
              >{m.ref.label}</span
            >
          </div>
        {/if}
      </ChatBubble>
    {/each}
  </div>

  <div class="composer-wrap">
    <div class="composer">
      <input
        bind:value={draft}
        placeholder="Pergunte ou comande {companion}…"
        onkeydown={(e) => e.key === 'Enter' && send()}
      />
      <button class="mic" aria-label="voz" onclick={() => startListening()}
        ><Icon name="mic" size={18} stroke="var(--purple-glow)" /></button
      >
      <button class="send" onclick={send} aria-label="enviar"
        ><Icon name="arrowUp" size={18} stroke="var(--text-on-purple)" /></button
      >
    </div>
    <div class="composer-hint">
      {companion} usa seu contexto e a memória conectada · comandos por texto ou voz
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
    align-items: center;
    justify-content: space-between;
    padding: 22px 36px 16px;
    border-bottom: 1px solid var(--border-subtle);
  }
  h1 {
    font-family: var(--font-display);
    font-size: 20px;
    font-weight: 600;
    letter-spacing: -0.02em;
  }
  .ctx {
    display: flex;
    gap: 7px;
    margin-top: 7px;
  }
  .cchip {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    height: 24px;
    padding: 0 9px;
    border-radius: 999px;
    font-family: var(--font-ui);
    font-size: 11.5px;
  }
  .cchip.purple {
    background: var(--purple-012);
    border: 1px solid var(--purple-024);
    color: var(--purple-glow);
  }
  .cchip.mut {
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid var(--border-strong);
    color: var(--text-2);
  }
  .corner {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 4px;
  }
  .corner-lbl {
    font-family: var(--font-mono);
    font-size: 9.5px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--purple-glow);
  }
  .thread {
    flex: 1;
    min-height: 0;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 20px;
    padding: 26px 36px 0;
    max-width: 820px;
    width: 100%;
    margin: 0 auto;
  }
  .ref {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-top: 10px;
    padding: 8px 11px;
    border-radius: 10px;
    background: var(--surface-1);
    border: 1px solid var(--border-violet);
    font-family: var(--font-mono);
    font-size: 11.5px;
    color: var(--text-2);
  }
  .ref.action {
    background: var(--success-bg);
    border-color: rgba(74, 222, 128, 0.3);
    font-family: var(--font-ui);
    font-size: 12.5px;
    color: var(--text-1);
  }
  .composer-wrap {
    padding: 16px 36px 22px;
    max-width: 820px;
    width: 100%;
    margin: 0 auto;
  }
  .composer {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 8px 8px 8px 18px;
    border-radius: 999px;
    background: var(--surface-2);
    border: 1px solid var(--border-violet);
    box-shadow:
      var(--shadow-2),
      inset 0 1px 0 var(--highlight-top);
  }
  .composer input {
    flex: 1;
    background: none;
    border: none;
    outline: none;
    color: var(--text-1);
    font-family: var(--font-body);
    font-size: 15px;
  }
  .composer input::placeholder {
    color: var(--text-3);
  }
  .mic {
    width: 40px;
    height: 40px;
    border-radius: 999px;
    border: 1px solid var(--border-strong);
    background: var(--surface-3);
    display: grid;
    place-items: center;
    cursor: pointer;
  }
  .send {
    width: 40px;
    height: 40px;
    border-radius: 999px;
    border: none;
    background: var(--grad-active);
    display: grid;
    place-items: center;
    cursor: pointer;
    box-shadow: var(--glow-sm);
  }
  .composer-hint {
    text-align: center;
    font-size: 11.5px;
    color: var(--text-3);
    margin-top: 9px;
  }
</style>
