<script lang="ts">
  import { page } from '$app/stores';
  import Icon from './Icon.svelte';
  import Presence from './Presence.svelte';
  import ContextChip from './ContextChip.svelte';
  import { voice } from '$stores/voice';
  import type { Context } from '$lib/types';

  let { contexts = [], activeContext = 'cogna' }: { contexts?: Context[]; activeContext?: string } =
    $props();

  const nav = [
    { href: '/', label: 'Início', icon: 'home' },
    { href: '/persona', label: 'Persona', icon: 'persona' },
    { href: '/agenda', label: 'Agenda', icon: 'agenda' },
    { href: '/analise', label: 'Análise', icon: 'chart' },
    { href: '/vps', label: 'VPS', icon: 'server' },
    { href: '/chat', label: 'Chat', icon: 'chat' },
    { href: '/rede', label: 'Rede neural', icon: 'graph' }
  ];

  let path = $derived($page.url.pathname);
  const stateLabel: Record<string, string> = {
    idle: 'Pronto',
    listening: 'Ouvindo',
    transcribing: 'Transcrevendo',
    thinking: 'Pensando',
    speaking: 'Respondendo',
    confirming: 'Confirmado'
  };
</script>

<aside class="sidebar">
  <div class="brand">
    <span class="mark"></span>
    <span class="word">Aqno</span>
  </div>

  <div class="companion">
    <Presence state={$voice.state} size={30} level={$voice.level} />
    <div class="comp-meta">
      <span class="comp-name">Íris</span>
      <span class="comp-state">{stateLabel[$voice.state] ?? 'Pronto'}</span>
    </div>
  </div>

  <nav>
    {#each nav as item}
      <a class="navitem" class:active={path === item.href} href={item.href}>
        <Icon
          name={item.icon}
          size={18}
          stroke={path === item.href ? 'var(--purple-glow)' : 'var(--text-3)'}
        />
        {item.label}
      </a>
    {/each}
  </nav>

  <div class="ctx-label">Contextos</div>
  <div class="contexts">
    {#each contexts as c}
      <ContextChip label={c.label} color={c.color} active={c.id === activeContext} size="sm" />
    {/each}
  </div>

  <a class="settings" href="/ajustes">
    <Icon name="settings" size={18} stroke="var(--text-3)" /> Ajustes
  </a>
</aside>

<style>
  .sidebar {
    width: var(--sidebar-w);
    flex: none;
    display: flex;
    flex-direction: column;
    background: var(--surface-1);
    border-right: 1px solid var(--border-subtle);
    padding: 18px 14px;
  }
  .brand { display: flex; align-items: center; gap: 10px; padding: 4px 8px 12px; }
  .mark {
    width: 26px;
    height: 26px;
    border-radius: 50%;
    background: var(--grad-presence);
    box-shadow: var(--glow-sm);
  }
  .word {
    font-family: var(--font-display);
    font-size: 19px;
    font-weight: 600;
    letter-spacing: -0.02em;
    color: var(--text-1);
  }
  .companion {
    display: flex;
    align-items: center;
    gap: 11px;
    padding: 10px;
    margin-bottom: 8px;
    background: var(--surface-2);
    border: 1px solid var(--border-violet);
    border-radius: var(--radius-md);
    box-shadow: inset 0 1px 0 var(--highlight-top);
  }
  .comp-meta { display: flex; flex-direction: column; gap: 2px; min-width: 0; }
  .comp-name { font-family: var(--font-display); font-size: 14px; font-weight: 600; color: var(--text-1); }
  .comp-state {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--purple-glow);
  }
  nav { display: flex; flex-direction: column; gap: 2px; }
  .navitem {
    display: flex;
    align-items: center;
    gap: 11px;
    padding: 9px 10px;
    border-radius: var(--radius-md);
    border: 1px solid transparent;
    color: var(--text-2);
    font-family: var(--font-ui);
    font-size: 14.5px;
    font-weight: 500;
    cursor: pointer;
    transition: var(--transition-base);
  }
  .navitem:hover { color: var(--text-1); background: var(--surface-2); }
  .navitem.active {
    border-color: var(--purple-024);
    background: var(--purple-012);
    color: var(--text-1);
  }
  .ctx-label {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--text-3);
    padding: 20px 10px 9px;
  }
  .contexts { display: flex; flex-direction: column; gap: 6px; padding: 0 4px; }
  .settings {
    margin-top: auto;
    display: flex;
    align-items: center;
    gap: 11px;
    padding: 9px 10px;
    color: var(--text-2);
    font-family: var(--font-ui);
    font-size: 14px;
    cursor: pointer;
  }
</style>
