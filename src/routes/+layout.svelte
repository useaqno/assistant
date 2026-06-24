<script lang="ts">
  import '@fontsource-variable/inter'
  import '@fontsource-variable/jetbrains-mono'
  import '../app.css'
  import { onMount } from 'svelte'
  import { page } from '$app/stores'
  import { goto } from '$app/navigation'
  import Sidebar from '$components/Sidebar.svelte'
  import VoiceBar from '$components/VoiceBar.svelte'
  import { presence } from '$stores/presence'
  import { app } from '$stores/app'
  import { onDaemonReady } from '$lib/tauri'

  const { children } = $props()

  const path = $derived($page.url.pathname)
  const bare = $derived(path === '/onboarding')

  async function hydrate() {
    const onboarded = await app.load()
    if (!onboarded && path !== '/onboarding') {
      goto('/onboarding')
    }
  }

  onMount(() => {
    presence.connect()
    hydrate()
    const off = onDaemonReady(() => hydrate())
    return () => {
      presence.disconnect()
      off.then((fn) => fn())
    }
  })
</script>

{#if bare}
  {@render children?.()}
{:else}
  <div class="shell">
    <Sidebar contexts={$app.contexts} />
    <main class="content">
      {@render children?.()}
      <div class="voice-dock">
        <VoiceBar />
      </div>
    </main>
  </div>
{/if}

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
