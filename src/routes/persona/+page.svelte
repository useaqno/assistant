<script lang="ts">
  import { onMount } from 'svelte';
  import Presence from '$components/Presence.svelte';
  import Button from '$components/Button.svelte';
  import SegmentedControl from '$components/SegmentedControl.svelte';
  import Icon from '$components/Icon.svelte';
  import { setVoice } from '$stores/voice';

  let shape = $state<'orbe' | 'animal' | 'personagem'>('orbe');
  let color = $state('#8B5CF6');
  let name = $state('Íris');
  let chosenVoice = $state('Aurora');
  let tom = $state('Amigável');

  const swatches = ['#8B5CF6', '#5EEAD4', '#FBBF24', '#FB7185', '#60A5FA'];
  const voices = ['Aurora', 'Vega', 'Lúmen'];

  onMount(() => {
    setVoice({ state: 'idle', hint: 'Diga algo — a Íris responde no tom que você escolher', level: 0.5 });
  });
</script>

<div class="page">
  <header class="head">
    <div>
      <div class="overline">Criar companheira · passo 2 de 4</div>
      <h1>Dê vida à sua companheira</h1>
    </div>
    <div class="actions">
      <Button variant="ghost">Pular</Button>
      <Button variant="primary">Concluir</Button>
    </div>
  </header>

  <div class="body">
    <div class="controls">
      <div class="group">
        <div class="overline">Forma</div>
        <div class="tiles">
          <button class="tile" class:on={shape === 'orbe'} onclick={() => (shape = 'orbe')}>
            <span class="orb-mini"></span>Orbe
          </button>
          <button class="tile" class:on={shape === 'animal'} onclick={() => (shape = 'animal')}>
            <Icon name="persona" size={30} stroke="var(--text-2)" />Animal
          </button>
          <button class="tile" class:on={shape === 'personagem'} onclick={() => (shape = 'personagem')}>
            <Icon name="users" size={28} stroke="var(--text-2)" />Personagem
          </button>
        </div>
      </div>

      <div class="group">
        <div class="overline">Cor e aura</div>
        <div class="swatches">
          {#each swatches as s}
            <button
              class="sw"
              class:on={color === s}
              style="background:{s};box-shadow:{color === s
                ? `0 0 0 2px var(--bg-base),0 0 0 4px ${s},0 0 14px ${s}`
                : 'none'}"
              onclick={() => (color = s)}
              aria-label="cor"
            ></button>
          {/each}
        </div>
        <div class="aura">
          <span>Aura · forte</span>
          <div class="track"><div class="fill" style="width:74%"></div><div class="knob" style="left:74%"></div></div>
        </div>
      </div>

      <label class="field">
        <span class="f-label">Nome</span>
        <input bind:value={name} />
      </label>

      <div class="group">
        <div class="overline">Voz</div>
        <div class="pills">
          {#each voices as v}
            <button class="pill" class:on={chosenVoice === v} onclick={() => (chosenVoice = v)}>
              {#if chosenVoice === v}<Icon name="check" size={14} stroke="#0C0A14" strokeWidth={2.4} />{/if}
              {v}
            </button>
          {/each}
        </div>
      </div>

      <div class="group">
        <div class="overline">Tom de personalidade</div>
        <SegmentedControl options={['Amigável', 'Direto', 'Formal']} bind:value={tom} full />
      </div>
    </div>

    <div class="stage">
      <span class="stage-tag overline">Prévia ao vivo</span>
      <Presence state="speaking" size={168} level={0.8} label="respondendo" />
      <div class="stage-name">
        <span class="sn">{name || 'Íris'}</span>
        <span class="st">{tom}</span>
      </div>
      <div class="caption">
        <p>"Oi, Renato. Sou a {name || 'Íris'} — vou te ajudar a manter tudo em ordem, no seu ritmo."</p>
      </div>
      <button class="test"><Icon name="mic" size={17} stroke="var(--purple-glow)" /> Testar voz</button>
    </div>
  </div>
</div>

<style>
  .page { display: flex; flex-direction: column; height: 100%; background: radial-gradient(circle at 72% 36%, rgba(139, 92, 246, 0.12), transparent 55%), var(--bg-base); }
  .head { display: flex; align-items: flex-end; justify-content: space-between; padding: 26px 36px 0; }
  .overline { font-family: var(--font-mono); font-size: 12px; letter-spacing: 0.08em; text-transform: uppercase; color: var(--purple-glow); }
  .group .overline, .stage-tag { font-size: 10px; color: var(--text-3); }
  h1 { font-family: var(--font-display); font-size: 28px; font-weight: 600; letter-spacing: -0.02em; margin-top: 6px; }
  .actions { display: flex; gap: 10px; }
  .body { flex: 1; min-height: 0; display: flex; gap: 26px; padding: 22px 36px 104px; }
  .controls { width: 432px; flex: none; display: flex; flex-direction: column; gap: 18px; overflow-y: auto; }
  .group .overline { margin-bottom: 9px; }
  .tiles { display: grid; grid-template-columns: 1fr 1fr 1fr; gap: 10px; }
  .tile { display: flex; flex-direction: column; align-items: center; gap: 9px; padding: 16px 8px; border-radius: var(--radius-md); background: var(--surface-2); border: 1px solid var(--border-violet); color: var(--text-2); font-family: var(--font-ui); font-size: 13px; font-weight: 500; cursor: pointer; }
  .tile.on { background: var(--purple-012); border-color: var(--purple); box-shadow: var(--glow-sm); color: var(--text-1); }
  .orb-mini { width: 34px; height: 34px; border-radius: 50%; background: var(--grad-presence); box-shadow: var(--glow-sm); }
  .swatches { display: flex; align-items: center; gap: 12px; margin-bottom: 14px; }
  .sw { width: 30px; height: 30px; border-radius: 50%; border: 1px solid rgba(255, 255, 255, 0.14); cursor: pointer; }
  .aura { display: flex; align-items: center; gap: 12px; }
  .aura span { font-size: 13px; color: var(--text-2); width: 74px; }
  .track { flex: 1; height: 6px; border-radius: 999px; background: var(--surface-1); position: relative; }
  .fill { position: absolute; left: 0; top: 0; bottom: 0; border-radius: 999px; background: var(--grad-active); box-shadow: var(--glow-sm); }
  .knob { position: absolute; top: 50%; transform: translate(-50%, -50%); width: 16px; height: 16px; border-radius: 50%; background: #fff; box-shadow: 0 1px 4px rgba(0, 0, 0, 0.5); }
  .field { display: flex; flex-direction: column; gap: 6px; }
  .f-label { font-family: var(--font-ui); font-size: 13px; font-weight: 500; color: var(--text-2); }
  .field input { height: 42px; padding: 0 12px; background: var(--surface-1); border: 1px solid var(--border-subtle); border-radius: var(--radius-input); color: var(--text-1); font-family: var(--font-body); font-size: 15px; outline: none; }
  .field input:focus { border-color: var(--purple); box-shadow: var(--glow-sm); }
  .pills { display: flex; gap: 8px; }
  .pill { display: inline-flex; align-items: center; gap: 7px; height: 36px; padding: 0 14px; border-radius: 999px; font-family: var(--font-ui); font-size: 13.5px; font-weight: 500; color: var(--text-1); background: var(--surface-2); border: 1px solid var(--border-strong); cursor: pointer; }
  .pill.on { color: #0c0a14; background: var(--purple-glow); border-color: transparent; }
  .stage { flex: 1; min-width: 0; position: relative; border-radius: var(--radius-lg); border: 1px solid var(--border-violet); overflow: hidden; background: radial-gradient(circle at 50% 40%, rgba(167, 139, 250, 0.2), transparent 58%), var(--surface-sunken); display: flex; flex-direction: column; align-items: center; justify-content: center; box-shadow: inset 0 1px 0 var(--highlight-top); }
  .stage-tag { position: absolute; top: 18px; left: 20px; }
  .stage-name { display: flex; align-items: center; gap: 10px; margin-top: 18px; }
  .sn { font-family: var(--font-display); font-size: 24px; font-weight: 600; letter-spacing: -0.02em; }
  .st { font-family: var(--font-mono); font-size: 10px; letter-spacing: 0.08em; text-transform: uppercase; color: var(--purple-glow); padding: 3px 8px; border-radius: 999px; background: var(--purple-012); border: 1px solid var(--purple-024); }
  .caption { margin-top: 18px; max-width: 380px; padding: 14px 18px; border-radius: var(--radius-card); background: color-mix(in srgb, var(--surface-2) 80%, transparent); backdrop-filter: var(--blur-subtle); -webkit-backdrop-filter: var(--blur-subtle); border: 1px solid var(--border-violet); box-shadow: inset 0 1px 0 var(--highlight-top); text-align: center; }
  .caption p { font-size: 15.5px; color: var(--text-1); line-height: 1.5; }
  .test { display: inline-flex; align-items: center; gap: 9px; margin-top: 20px; height: 44px; padding: 0 20px; border-radius: 999px; border: 1px solid var(--border-strong); background: var(--surface-2); color: var(--text-1); font-family: var(--font-ui); font-size: 14px; font-weight: 500; cursor: pointer; }
</style>
