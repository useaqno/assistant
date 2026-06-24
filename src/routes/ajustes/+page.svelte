<script lang="ts">
  import { onMount } from 'svelte'
  import Button from '$components/Button.svelte'
  import SegmentedControl from '$components/SegmentedControl.svelte'
  import ContextChip from '$components/ContextChip.svelte'
  import { api } from '$lib/api'
  import { app } from '$stores/app'
  import type { Config, Context, Persona, Server } from '$lib/types'

  let cfg = $state<Config>({})
  let persona = $state<Persona>({
    name: '',
    owner: '',
    avatar: 'orbe',
    auraColor: '#8B5CF6',
    tone: 'amigavel',
    wakeWord: 'aqno'
  })
  let contexts = $state<Context[]>([])
  let servers = $state<Server[]>([])
  let llmKey = $state('')
  let keyConfigured = $state(false)
  let engine = $state<{ active: string; available: boolean; apple: boolean }>({
    active: 'none',
    available: false,
    apple: false
  })
  let toast = $state('')

  // new server form
  let nsName = $state('')
  let nsHost = $state('')
  let nsPort = $state(22)
  let nsUser = $state('')
  let nsAuth = $state('senha')
  let nsSecret = $state('')

  const providers = ['anthropic', 'openai', 'openai_compatible', 'ollama']
  const providerLabels: Record<string, string> = {
    anthropic: 'Anthropic',
    openai: 'OpenAI',
    openai_compatible: 'Compatível',
    ollama: 'Ollama'
  }
  const tiers = ['small', 'medium', 'large-v3-turbo']
  const sttEngines = ['whisper', 'apple', 'auto']
  const engineLabels: Record<string, string> = {
    'whisper.cpp': 'whisper.cpp (local)',
    'whisper-server': 'whisper server',
    'apple-speechanalyzer': 'Apple SpeechAnalyzer',
    none: 'nenhum'
  }
  const provider = $derived(cfg['llm.provider'] ?? 'anthropic')
  const showBaseUrl = $derived(provider === 'openai_compatible' || provider === 'ollama')

  function flash(msg: string) {
    toast = msg
    setTimeout(() => (toast = ''), 2500)
  }

  async function load() {
    try {
      const b = await api.bootstrap()
      cfg = { ...b.config }
      if (b.persona?.name) persona = { ...persona, ...b.persona }
      contexts = [...b.contexts]
    } catch {
      cfg = { ...$app.config }
      contexts = [...$app.contexts]
    }
    try {
      servers = await api.servers()
    } catch {
      servers = []
    }
    refreshKeyStatus()
    refreshEngine()
  }

  async function refreshEngine() {
    try {
      engine = await api.voiceEngine()
    } catch {
      engine = { active: 'none', available: false, apple: false }
    }
  }

  async function setEngine(v: string) {
    saveConfig0('voice.stt_engine', v)
    await refreshEngine()
  }

  async function refreshKeyStatus() {
    try {
      const s = await api.llmKeyStatus(provider)
      keyConfigured = s.configured
    } catch {
      keyConfigured = false
    }
  }

  async function savePersona() {
    try {
      await api.onboarding(persona)
      app.setPersona(persona)
      flash('Persona salva')
    } catch {
      flash('Falha ao salvar persona')
    }
  }

  async function saveConfig(keys: string[]) {
    const patch: Config = {}
    for (const k of keys) patch[k] = cfg[k] ?? ''
    app.setConfig(patch)
    flash('Configuração salva')
  }

  function saveConfig0(key: string, val: string) {
    cfg[key] = val
    app.setConfig({ [key]: val })
    flash('Configuração salva')
  }

  function setProvider(p: string) {
    cfg['llm.provider'] = p
    app.setConfig({ 'llm.provider': p })
    refreshKeyStatus()
  }

  async function saveKey() {
    try {
      await api.setLLMKey(provider, llmKey)
      llmKey = ''
      flash('Chave salva no Keychain')
      refreshKeyStatus()
    } catch {
      flash('Falha ao salvar chave')
    }
  }

  async function toggleAIMode(c: Context) {
    const mode = c.aiMode === 'local_only' ? 'cloud' : 'local_only'
    c.aiMode = mode
    try {
      await api.setContextAIMode(c.label, mode)
      flash(`${c.label}: ${mode === 'local_only' ? 'somente local' : 'nuvem'}`)
    } catch {
      c.aiMode = mode === 'local_only' ? 'cloud' : 'local_only'
    }
  }

  async function addServer() {
    if (!nsHost.trim() || !nsUser.trim()) return
    try {
      await api.createServer({
        name: nsName || nsHost,
        host: nsHost,
        port: nsPort,
        user: nsUser,
        authType: nsAuth as 'senha' | 'chave',
        secret: nsSecret
      })
      nsName = nsHost = nsUser = nsSecret = ''
      nsPort = 22
      servers = await api.servers()
      flash('Servidor adicionado')
    } catch {
      flash('Falha ao adicionar servidor')
    }
  }

  async function removeServer(id?: string) {
    if (!id) return
    try {
      await api.deleteServer(id)
      servers = await api.servers()
      flash('Servidor removido')
    } catch {
      /* ignore */
    }
  }

  onMount(load)
</script>

<div class="page">
  <header class="head">
    <div>
      <div class="overline">Configurações</div>
      <h1>Ajustes</h1>
    </div>
  </header>

  {#if toast}<div class="toast">{toast}</div>{/if}

  <div class="scroll">
    <!-- Persona -->
    <section class="card">
      <div class="s-title">Persona</div>
      <div class="grid2">
        <label class="field"><span>Seu nome</span><input bind:value={persona.owner} /></label>
        <label class="field"
          ><span>Nome da companheira</span><input bind:value={persona.name} /></label
        >
      </div>
      <div class="grid2">
        <label class="field"
          ><span>Palavra de ativação</span><input bind:value={persona.wakeWord} /></label
        >
        <div class="field">
          <span>Tom</span>
          <SegmentedControl
            options={['amigavel', 'direto', 'formal']}
            value={persona.tone}
            onchange={(v) => (persona.tone = v)}
            full
          />
        </div>
      </div>
      <div class="row-end">
        <Button variant="primary" size="sm" onclick={savePersona}>Salvar persona</Button>
      </div>
    </section>

    <!-- IA -->
    <section class="card">
      <div class="s-title">Inteligência Artificial</div>
      <div class="field">
        <span>Provedor</span>
        <div class="pills">
          {#each providers as p (p)}
            <button class="pill" class:on={provider === p} onclick={() => setProvider(p)}>
              {providerLabels[p]}
            </button>
          {/each}
        </div>
      </div>
      <div class="grid2">
        <label class="field">
          <span>Modelo</span>
          <input
            value={cfg['llm.model'] ?? ''}
            oninput={(e) => (cfg['llm.model'] = e.currentTarget.value)}
          />
        </label>
        {#if showBaseUrl}
          <label class="field">
            <span>Base URL</span>
            <input
              value={cfg['llm.base_url'] ?? ''}
              oninput={(e) => (cfg['llm.base_url'] = e.currentTarget.value)}
              placeholder="http://localhost:11434"
            />
          </label>
        {/if}
      </div>
      <div class="grid2">
        <label class="field">
          <span>Max tokens</span>
          <input
            value={cfg['llm.max_tokens'] ?? '2000'}
            oninput={(e) => (cfg['llm.max_tokens'] = e.currentTarget.value)}
          />
        </label>
        <label class="field">
          <span>Temperatura</span>
          <input
            value={cfg['llm.temperature'] ?? '0.4'}
            oninput={(e) => (cfg['llm.temperature'] = e.currentTarget.value)}
          />
        </label>
      </div>
      <div class="row-end">
        <Button
          variant="subtle"
          size="sm"
          onclick={() =>
            saveConfig(['llm.model', 'llm.base_url', 'llm.max_tokens', 'llm.temperature'])}
        >
          Salvar modelo
        </Button>
      </div>

      {#if provider !== 'ollama'}
        <div class="key-row">
          <label class="field grow">
            <span>
              Chave da API
              {#if keyConfigured}<span class="ok-badge">✓ configurada</span>{/if}
            </span>
            <input
              type="password"
              bind:value={llmKey}
              placeholder="cole sua chave aqui (vai pro Keychain)"
            />
          </label>
          <Button variant="primary" size="sm" onclick={saveKey} disabled={!llmKey}
            >Salvar chave</Button
          >
        </div>
        <p class="note">A chave é guardada no Keychain do macOS — nunca no banco de dados.</p>
      {/if}
    </section>

    <!-- Voz -->
    <section class="card">
      <div class="s-title">Voz</div>
      <div class="field">
        <span>Motor de reconhecimento (STT)</span>
        <SegmentedControl
          options={sttEngines}
          value={cfg['voice.stt_engine'] ?? 'whisper'}
          onchange={setEngine}
          full
        />
      </div>
      <p class="note">
        Ativo agora: <b>{engineLabels[engine.active] ?? engine.active}</b>
        {#if !engine.available}· nenhum motor pronto — baixe um modelo abaixo{/if}
        {#if cfg['voice.stt_engine'] === 'apple' && !engine.apple}
          · Apple SpeechAnalyzer indisponível (requer macOS 26)
        {/if}
      </p>
      <div class="field">
        <span>Qualidade do modelo (whisper)</span>
        <SegmentedControl
          options={tiers}
          value={cfg['voice.model_tier'] ?? 'small'}
          onchange={(v) => saveConfig0('voice.model_tier', v)}
          full
        />
      </div>
      <div class="grid2">
        <div class="field">
          <span>Idioma do STT</span>
          <SegmentedControl
            options={['pt', 'en', 'auto']}
            value={cfg['voice.stt_lang'] ?? 'pt'}
            onchange={(v) => saveConfig0('voice.stt_lang', v)}
            full
          />
        </div>
        <label class="field">
          <span>Voz do TTS</span>
          <input
            value={cfg['voice.tts_voice'] ?? 'Luciana'}
            oninput={(e) => (cfg['voice.tts_voice'] = e.currentTarget.value)}
            onblur={() => saveConfig0('voice.tts_voice', cfg['voice.tts_voice'] ?? '')}
          />
        </label>
      </div>
      <div class="field">
        <span>Confirmação por voz</span>
        <SegmentedControl
          options={['destrutivas', 'sempre']}
          value={cfg['voice.confirm'] ?? 'destrutivas'}
          onchange={(v) => saveConfig0('voice.confirm', v)}
          full
        />
      </div>
    </section>

    <!-- Contextos -->
    <section class="card">
      <div class="s-title">Contextos e privacidade</div>
      <p class="note">Contextos marcados como “somente local” nunca enviam dados para a nuvem.</p>
      <div class="ctx-list">
        {#each contexts as c (c.id)}
          <div class="ctx-row">
            <ContextChip label={c.label} color={c.color} size="sm" />
            <button
              class="mode"
              class:local={c.aiMode === 'local_only'}
              onclick={() => toggleAIMode(c)}
            >
              {c.aiMode === 'local_only' ? '🔒 somente local' : '☁ nuvem'}
            </button>
          </div>
        {/each}
      </div>
    </section>

    <!-- Servidores -->
    <section class="card">
      <div class="s-title">Servidores (VPS)</div>
      {#each servers as srv (srv.id)}
        <div class="srv-row">
          <div class="srv-meta">
            <span class="srv-name">{srv.name}</span>
            <span class="srv-host mono">{srv.user}@{srv.host}:{srv.port}</span>
          </div>
          <Button variant="subtle" size="sm" onclick={() => removeServer(srv.id)}>Remover</Button>
        </div>
      {/each}
      {#if servers.length === 0}
        <p class="note">Nenhum servidor cadastrado.</p>
      {/if}

      <div class="srv-form">
        <div class="grid2">
          <label class="field"
            ><span>Nome</span><input bind:value={nsName} placeholder="Produção" /></label
          >
          <label class="field"
            ><span>Host</span><input bind:value={nsHost} placeholder="10.0.0.1" /></label
          >
        </div>
        <div class="grid3">
          <label class="field"
            ><span>Usuário</span><input bind:value={nsUser} placeholder="root" /></label
          >
          <label class="field"><span>Porta</span><input type="number" bind:value={nsPort} /></label>
          <div class="field">
            <span>Auth</span>
            <SegmentedControl options={['senha', 'chave']} bind:value={nsAuth} full />
          </div>
        </div>
        <label class="field">
          <span>{nsAuth === 'chave' ? 'Chave privada (PEM)' : 'Senha'}</span>
          {#if nsAuth === 'chave'}
            <textarea
              bind:value={nsSecret}
              rows="3"
              placeholder="-----BEGIN OPENSSH PRIVATE KEY-----"></textarea>
          {:else}
            <input type="password" bind:value={nsSecret} />
          {/if}
        </label>
        <div class="row-end">
          <Button variant="primary" size="sm" onclick={addServer} disabled={!nsHost || !nsUser}>
            Adicionar servidor
          </Button>
        </div>
      </div>
    </section>
  </div>
</div>

<style>
  .page {
    display: flex;
    flex-direction: column;
    height: 100%;
  }
  .head {
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
  .toast {
    margin: 12px 36px 0;
    padding: 10px 14px;
    border-radius: var(--radius-md);
    background: var(--success-bg);
    border: 1px solid rgba(74, 222, 128, 0.3);
    color: var(--success);
    font-size: 13px;
  }
  .scroll {
    flex: 1;
    min-height: 0;
    overflow-y: auto;
    padding: 18px 36px 120px;
    display: flex;
    flex-direction: column;
    gap: 16px;
    max-width: 820px;
  }
  .card {
    background: var(--surface-1);
    border: 1px solid var(--border-violet);
    border-radius: var(--radius-card);
    box-shadow: inset 0 1px 0 var(--highlight-top);
    padding: 20px;
    display: flex;
    flex-direction: column;
    gap: 14px;
  }
  .s-title {
    font-family: var(--font-display);
    font-size: 16px;
    font-weight: 600;
  }
  .grid2 {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 12px;
  }
  .grid3 {
    display: grid;
    grid-template-columns: 1fr 1fr 1fr;
    gap: 12px;
  }
  .field {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .field.grow {
    flex: 1;
  }
  .field span {
    font-size: 12.5px;
    color: var(--text-2);
    display: flex;
    align-items: center;
    gap: 8px;
  }
  .field input,
  .field textarea {
    padding: 10px 12px;
    background: var(--surface-2);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-input);
    color: var(--text-1);
    font-family: var(--font-body);
    font-size: 14px;
    outline: none;
  }
  .field input:focus,
  .field textarea:focus {
    border-color: var(--purple);
  }
  .pills {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
  }
  .pill {
    height: 34px;
    padding: 0 14px;
    border-radius: 999px;
    font-size: 13px;
    font-weight: 500;
    color: var(--text-2);
    background: var(--surface-2);
    border: 1px solid var(--border-strong);
    cursor: pointer;
  }
  .pill.on {
    color: #0c0a14;
    background: var(--purple-glow);
    border-color: transparent;
  }
  .row-end {
    display: flex;
    justify-content: flex-end;
  }
  .key-row {
    display: flex;
    align-items: flex-end;
    gap: 10px;
  }
  .ok-badge {
    color: var(--success);
    font-family: var(--font-mono);
    font-size: 11px;
  }
  .note {
    font-size: 12.5px;
    color: var(--text-3);
  }
  .ctx-list {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  .ctx-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }
  .mode {
    height: 30px;
    padding: 0 12px;
    border-radius: 999px;
    border: 1px solid var(--border-strong);
    background: var(--surface-2);
    color: var(--text-2);
    font-size: 12.5px;
    cursor: pointer;
  }
  .mode.local {
    border-color: var(--purple-024);
    background: var(--purple-012);
    color: var(--purple-glow);
  }
  .srv-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 10px 0;
    border-bottom: 1px solid var(--border-subtle);
  }
  .srv-meta {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
  .srv-name {
    font-size: 14px;
    font-weight: 500;
    color: var(--text-1);
  }
  .srv-host {
    font-size: 12px;
    color: var(--text-3);
  }
  .mono {
    font-family: var(--font-mono);
  }
  .srv-form {
    display: flex;
    flex-direction: column;
    gap: 12px;
    margin-top: 8px;
    padding-top: 14px;
    border-top: 1px dashed var(--border-subtle);
  }
</style>
