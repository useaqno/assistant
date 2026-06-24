<script lang="ts">
  import { goto } from '$app/navigation'
  import Presence from '$components/Presence.svelte'
  import Button from '$components/Button.svelte'
  import SegmentedControl from '$components/SegmentedControl.svelte'
  import Icon from '$components/Icon.svelte'
  import { api } from '$lib/api'
  import { app } from '$stores/app'
  import type { Persona, PresenceState } from '$lib/types'

  let step = $state(0)
  const total = 5

  let owner = $state('')
  let name = $state('Íris')
  let shape = $state<'orbe' | 'animal' | 'personagem'>('orbe')
  let color = $state('#8B5CF6')
  let chosenVoice = $state('Luciana')
  let tom = $state('Amigável')
  let wakeWord = $state('aqno')
  let micState = $state<'idle' | 'asking' | 'granted' | 'denied'>('idle')
  let saving = $state(false)

  const swatches = ['#8B5CF6', '#5EEAD4', '#FBBF24', '#FB7185', '#60A5FA']
  const voices = ['Luciana', 'Aurora', 'Vega']
  const previewState: PresenceState[] = ['idle', 'listening', 'speaking', 'thinking', 'speaking']

  const canNext = $derived(
    (step === 1 && owner.trim() !== '' && name.trim() !== '') ||
      step === 0 ||
      step === 2 ||
      step === 3 ||
      step === 4
  )

  async function requestMic() {
    micState = 'asking'
    try {
      const s = await navigator.mediaDevices.getUserMedia({ audio: true })
      s.getTracks().forEach((t) => t.stop())
      micState = 'granted'
    } catch {
      micState = 'denied'
    }
  }

  function next() {
    if (step < total - 1) step += 1
  }
  function back() {
    if (step > 0) step -= 1
  }

  async function finish() {
    saving = true
    const persona: Persona = {
      name: name.trim() || 'Íris',
      owner: owner.trim(),
      avatar: shape,
      auraColor: color,
      voice: chosenVoice,
      tone: tom.toLowerCase(),
      wakeWord: wakeWord.trim() || 'aqno'
    }
    try {
      await api.onboarding(persona)
      app.setPersona(persona)
    } catch {
      /* offline: still proceed so the user isn't stuck */
    }
    saving = false
    goto('/')
  }
</script>

<div class="onb">
  <div class="stage">
    <Presence state={previewState[step]} size={150} level={0.8} />
    <div class="stage-name">
      <span class="sn">{name || 'Íris'}</span>
      {#if step >= 3}<span class="st">{tom}</span>{/if}
    </div>
    {#if step >= 1 && owner}
      <p class="caption">
        "Oi, {owner.split(' ')[0]}. Sou a {name} — vou manter tudo em ordem, no seu ritmo."
      </p>
    {/if}
  </div>

  <div class="panel">
    <div class="progress">
      {#each Array(total) as _, i (i)}
        <span class="dot" class:on={i <= step}></span>
      {/each}
    </div>

    {#if step === 0}
      <div class="overline">Bem-vindo ao Aqno</div>
      <h1>Seu companheiro de IA, com a voz como interface</h1>
      <p class="lead">
        O Aqno vive na sua máquina e organiza suas demandas por voz. Vamos criar sua companheira em
        alguns passos.
      </p>
    {:else if step === 1}
      <div class="overline">Passo 1 · Identidade</div>
      <h1>Como vocês se chamam?</h1>
      <label class="field">
        <span class="f-label">Seu nome</span>
        <input bind:value={owner} placeholder="Ex.: Renato" />
      </label>
      <label class="field">
        <span class="f-label">Nome da companheira</span>
        <input bind:value={name} placeholder="Ex.: Íris" />
      </label>
    {:else if step === 2}
      <div class="overline">Passo 2 · Avatar</div>
      <h1>Escolha a forma</h1>
      <div class="tiles">
        <button class="tile" class:on={shape === 'orbe'} onclick={() => (shape = 'orbe')}>
          <span class="orb-mini"></span>Orbe
        </button>
        <button class="tile" class:on={shape === 'animal'} onclick={() => (shape = 'animal')}>
          <Icon name="persona" size={28} stroke="var(--text-2)" />Animal
        </button>
        <button
          class="tile"
          class:on={shape === 'personagem'}
          onclick={() => (shape = 'personagem')}
        >
          <Icon name="users" size={26} stroke="var(--text-2)" />Personagem
        </button>
      </div>
      <div class="overline mt">Cor da aura</div>
      <div class="swatches">
        {#each swatches as s (s)}
          <button
            class="sw"
            style="background:{s};box-shadow:{color === s
              ? `0 0 0 2px var(--bg-base),0 0 0 4px ${s},0 0 14px ${s}`
              : 'none'}"
            onclick={() => (color = s)}
            aria-label="cor"
          ></button>
        {/each}
      </div>
    {:else if step === 3}
      <div class="overline">Passo 3 · Voz e tom</div>
      <h1>Como ela soa?</h1>
      <div class="overline mt">Voz</div>
      <div class="pills">
        {#each voices as v (v)}
          <button class="pill" class:on={chosenVoice === v} onclick={() => (chosenVoice = v)}>
            {#if chosenVoice === v}<Icon
                name="check"
                size={13}
                stroke="#0C0A14"
                strokeWidth={2.4}
              />{/if}
            {v}
          </button>
        {/each}
      </div>
      <div class="overline mt">Tom de personalidade</div>
      <SegmentedControl options={['Amigável', 'Direto', 'Formal']} bind:value={tom} full />
      <label class="field mt">
        <span class="f-label">Palavra de ativação</span>
        <input bind:value={wakeWord} placeholder="aqno" />
      </label>
    {:else if step === 4}
      <div class="overline">Passo 4 · Microfone</div>
      <h1>Permita ouvir você</h1>
      <p class="lead">
        O áudio é processado localmente na sua máquina. A permissão de microfone é o que torna a voz
        possível.
      </p>
      <div class="mic-row">
        <Button variant={micState === 'granted' ? 'subtle' : 'primary'} onclick={requestMic}>
          {micState === 'granted'
            ? '✓ Microfone liberado'
            : micState === 'denied'
              ? 'Tentar novamente'
              : 'Permitir microfone'}
        </Button>
        {#if micState === 'denied'}
          <span class="hint warn"
            >Permissão negada — você pode liberar depois nas configurações do sistema.</span
          >
        {/if}
      </div>
    {/if}

    <div class="nav">
      {#if step > 0}
        <Button variant="ghost" onclick={back}>Voltar</Button>
      {:else}
        <span></span>
      {/if}
      {#if step < total - 1}
        <Button variant="primary" onclick={next} disabled={!canNext}>Continuar</Button>
      {:else}
        <Button variant="primary" onclick={finish} disabled={saving}>
          {saving ? 'Criando…' : 'Criar e começar'}
        </Button>
      {/if}
    </div>
  </div>
</div>

<style>
  .onb {
    display: flex;
    height: 100vh;
    width: 100vw;
    background:
      radial-gradient(circle at 30% 40%, rgba(139, 92, 246, 0.14), transparent 55%), var(--bg-base);
  }
  .stage {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 18px;
    border-right: 1px solid var(--border-subtle);
  }
  .stage-name {
    display: flex;
    align-items: center;
    gap: 10px;
  }
  .sn {
    font-family: var(--font-display);
    font-size: 26px;
    font-weight: 600;
    letter-spacing: -0.02em;
  }
  .st {
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
  .caption {
    max-width: 380px;
    text-align: center;
    font-size: 15px;
    color: var(--text-2);
    line-height: 1.5;
  }
  .panel {
    width: 480px;
    flex: none;
    display: flex;
    flex-direction: column;
    padding: 48px 44px;
    overflow-y: auto;
  }
  .progress {
    display: flex;
    gap: 7px;
    margin-bottom: 26px;
  }
  .dot {
    width: 28px;
    height: 4px;
    border-radius: 999px;
    background: var(--surface-3);
  }
  .dot.on {
    background: var(--grad-active);
  }
  .overline {
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--purple-glow);
  }
  .overline.mt {
    margin-top: 20px;
    margin-bottom: 9px;
    color: var(--text-3);
  }
  h1 {
    font-family: var(--font-display);
    font-size: 26px;
    font-weight: 600;
    letter-spacing: -0.02em;
    margin: 8px 0 14px;
  }
  .lead {
    font-size: 15px;
    color: var(--text-2);
    line-height: 1.55;
  }
  .field {
    display: flex;
    flex-direction: column;
    gap: 6px;
    margin-top: 14px;
  }
  .field.mt {
    margin-top: 20px;
  }
  .f-label {
    font-size: 13px;
    font-weight: 500;
    color: var(--text-2);
  }
  .field input {
    height: 44px;
    padding: 0 14px;
    background: var(--surface-1);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-input);
    color: var(--text-1);
    font-family: var(--font-body);
    font-size: 15px;
    outline: none;
  }
  .field input:focus {
    border-color: var(--purple);
    box-shadow: var(--glow-sm);
  }
  .tiles {
    display: grid;
    grid-template-columns: 1fr 1fr 1fr;
    gap: 10px;
  }
  .tile {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 9px;
    padding: 16px 8px;
    border-radius: var(--radius-md);
    background: var(--surface-2);
    border: 1px solid var(--border-violet);
    color: var(--text-2);
    font-size: 13px;
    font-weight: 500;
    cursor: pointer;
  }
  .tile.on {
    background: var(--purple-012);
    border-color: var(--purple);
    box-shadow: var(--glow-sm);
    color: var(--text-1);
  }
  .orb-mini {
    width: 30px;
    height: 30px;
    border-radius: 50%;
    background: var(--grad-presence);
    box-shadow: var(--glow-sm);
  }
  .swatches {
    display: flex;
    gap: 12px;
  }
  .sw {
    width: 30px;
    height: 30px;
    border-radius: 50%;
    border: 1px solid rgba(255, 255, 255, 0.14);
    cursor: pointer;
  }
  .pills {
    display: flex;
    gap: 8px;
  }
  .pill {
    display: inline-flex;
    align-items: center;
    gap: 7px;
    height: 36px;
    padding: 0 14px;
    border-radius: 999px;
    font-size: 13.5px;
    font-weight: 500;
    color: var(--text-1);
    background: var(--surface-2);
    border: 1px solid var(--border-strong);
    cursor: pointer;
  }
  .pill.on {
    color: #0c0a14;
    background: var(--purple-glow);
    border-color: transparent;
  }
  .mic-row {
    display: flex;
    flex-direction: column;
    gap: 12px;
    margin-top: 18px;
  }
  .hint {
    font-size: 13px;
    color: var(--text-3);
  }
  .hint.warn {
    color: var(--warning);
  }
  .nav {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-top: auto;
    padding-top: 28px;
  }
</style>
