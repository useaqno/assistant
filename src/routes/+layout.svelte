<script lang="ts">
  import '@fontsource-variable/inter';
  import '@fontsource-variable/jetbrains-mono';
  import '../app.css';
  import { onMount } from 'svelte';
  import Sidebar from '$components/Sidebar.svelte';
  import VoiceBar from '$components/VoiceBar.svelte';
  import { api } from '$lib/api';
  import { presence } from '$stores/presence';
  import { onDaemonReady } from '$lib/tauri';
  import type { Context } from '$lib/types';

  let { children } = $props();

  // Static fallback so the shell paints instantly even before the daemon answers.
  let contexts = $state<Context[]>([
    { id: 'cogna', label: 'Cogna', color: 'violet' },
    { id: 'bayer', label: 'Bayer', color: 'teal' },
    { id: 'visa', label: 'Visa', color: 'amber' },
    { id: 'devlith', label: 'Devlith', color: 'rose' },
    { id: 'pitrace', label: 'Pitrace', color: 'blue' }
  ]);

  async function loadContexts() {
    try {
      contexts = await api.contexts();
    } catch {
      /* keep fallback */
    }
  }

  onMount(() => {
    presence.connect();
    loadContexts();
    // Re-fetch once the sidecar announces it is live (Tauri).
    const off = onDaemonReady(() => loadContexts());
    return () => {
      presence.disconnect();
      off.then((fn) => fn());
    };
  });
</script>

<div class="shell">
  <Sidebar {contexts} />
  <main class="content">
    {@render children?.()}
    <div class="voice-dock">
      <VoiceBar />
    </div>
  </main>
</div>

<style>
  .shell {
    display: flex;
    height: 100vh;
    width: 100vw;
    overflow: hidden;
  }
  .content {
    flex: 1;
    min-width: 0;
    position: relative;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    background: var(--bg-base);
  }
  .voice-dock {
    position: absolute;
    left: 50%;
    bottom: 24px;
    transform: translateX(-50%);
    width: min(680px, calc(100% - 64px));
    z-index: 50;
  }
</style>
