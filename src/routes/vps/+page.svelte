<script lang="ts">
  import { onMount } from 'svelte';
  import Card from '$components/Card.svelte';
  import MetricRing from '$components/MetricRing.svelte';
  import Button from '$components/Button.svelte';
  import Icon from '$components/Icon.svelte';
  import { api } from '$lib/api';
  import { setVoice } from '$stores/voice';
  import type { Vps } from '$lib/types';

  let v = $state<Vps | null>(null);
  // The worker is flagged unstable — show the explicit restart confirmation by
  // default, matching the "voice command awaiting confirmation" moment.
  let confirmTarget = $state<string | null>('worker-fila');

  const levelColor: Record<string, string> = { INFO: 'var(--info)', WARN: 'var(--warning)', OK: 'var(--success)', CMD: 'var(--purple-glow)' };
  const ringColor = ['var(--data-teal)', 'var(--data-amber)', 'var(--data-violet)'];

  function askRestart(name: string) {
    confirmTarget = name;
  }
  async function doRestart() {
    if (!confirmTarget) return;
    const name = confirmTarget;
    try {
      const r = await api.restart(name, true);
      setVoice({ state: 'confirming', transcript: r.message, level: 0.6 });
    } catch {
      setVoice({ state: 'confirming', transcript: `Container reiniciado · ${name}`, level: 0.6 });
    }
    confirmTarget = null;
  }

  onMount(async () => {
    setVoice({ state: 'listening', transcript: 'Íris, reinicia o worker da fila', level: 0.6 });
    try {
      v = await api.vps();
    } catch {
      /* offline */
    }
  });

  let gauges = $derived([
    { label: 'CPU · 8 vCPU', detail: v?.cpuDetail ?? 'load 2.1 / 1.8 / 1.5', value: v?.cpu ?? 0.42, spark: '0,18 15,14 30,19 45,10 60,15 75,8 90,13 105,9 120,12' },
    { label: 'Memória', detail: v?.ramDetail ?? '11.5 / 16 GB', value: v?.ram ?? 0.72, spark: '0,14 15,13 30,12 45,11 60,12 75,9 90,8 105,7 120,6' },
    { label: 'Disco · SSD', detail: v?.diskDetail ?? '232 / 400 GB', value: v?.disk ?? 0.58, spark: '0,16 15,16 30,15 45,15 60,14 75,14 90,13 105,13 120,12' }
  ]);
</script>

<div class="page">
  <header class="head">
    <div>
      <div class="overline">{v?.host ?? 'aqno@10.0.4.12'} · {v?.uptime ?? 'uptime 27d'}</div>
      <h1>Infra · VPS</h1>
    </div>
    <div class="badges">
      <span class="chip warn"><span class="d"></span>{v?.warnings ?? 1} aviso</span>
      <span class="chip ok"><span class="d glow"></span>{v?.online === false ? 'offline' : 'online'}</span>
    </div>
  </header>

  <div class="body">
    <div class="gauges">
      {#each gauges as g, i}
        <Card padding={18}>
          <div class="g-inner">
            <MetricRing value={g.value} size={88} thickness={7} label="{Math.round(g.value * 100)}%" caption={g.label.split(' ')[0]} color={ringColor[i]} />
            <div class="g-meta">
              <div class="overline">{g.label}</div>
              <div class="g-detail mono">{g.detail}</div>
              <svg width="100%" height="26" viewBox="0 0 120 26" preserveAspectRatio="none" fill="none" stroke={ringColor[i]} stroke-width="1.6"><polyline points={g.spark} /></svg>
            </div>
          </div>
        </Card>
      {/each}
    </div>

    <div class="cols">
      <div class="containers">
        <div class="ct-head"><span class="ct-title">Containers</span><span class="ct-sub mono">{v?.containers.length ?? 5} ativos</span></div>
        {#each v?.containers ?? [] as c, i}
          <div class="crow" class:warn={c.status === 'restarting'} class:divider={i < (v?.containers.length ?? 0) - 1}>
            <span class="cdot" style="background:{c.status === 'restarting' ? 'var(--warning)' : 'var(--success)'};box-shadow:0 0 8px {c.status === 'restarting' ? 'var(--warning)' : 'var(--success)'}"></span>
            <span class="cname">{c.name}</span>
            {#if c.status === 'restarting'}
              <span class="cstatus mono"><span class="spin"><Icon name="refresh" size={11} stroke="var(--warning)" strokeWidth={2.4} /></span>reiniciando</span>
            {:else}
              <span class="cimg mono">{c.image}</span>
            {/if}
            <span class="ccpu mono">{c.cpu}</span>
            <span class="cmem mono">{c.mem}</span>
            <button class="cbtn" onclick={() => askRestart(c.name)} aria-label="reiniciar"><Icon name="refresh" size={15} stroke={c.status === 'restarting' ? 'var(--warning)' : 'var(--text-3)'} /></button>
          </div>
        {/each}

        {#if confirmTarget}
          <div class="confirm">
            <span class="cf-icon"><Icon name="alert" size={18} stroke="var(--warning)" /></span>
            <div class="cf-body">
              <div class="cf-title">Reiniciar {confirmTarget}?</div>
              <div class="cf-sub">O container ficará indisponível por ~3s. Diga "confirmar" ou clique.</div>
            </div>
            <Button variant="subtle" size="sm" onclick={() => (confirmTarget = null)}>Cancelar</Button>
            <Button variant="danger" size="sm" onclick={doRestart}>Reiniciar</Button>
          </div>
        {/if}
      </div>

      <div class="logs">
        <div class="logs-head"><span class="overline">Logs ao vivo</span><span class="stream mono"><span class="sdot"></span>stream</span></div>
        <div class="log-lines">
          {#each v?.logs ?? [] as l}
            <div class="log"><span class="lt">{l.time}</span><span class="ll" style="color:{levelColor[l.level]}">{l.level.padEnd(4)}</span><span class="lb">{l.body}</span></div>
          {/each}
        </div>
      </div>
    </div>
  </div>
</div>

<style>
  .page { display: flex; flex-direction: column; height: 100%; }
  .head { display: flex; align-items: flex-end; justify-content: space-between; padding: 26px 36px 0; }
  .overline { font-family: var(--font-mono); font-size: 10px; letter-spacing: 0.08em; text-transform: uppercase; color: var(--text-3); }
  .head .overline { font-size: 12px; color: var(--purple-glow); }
  h1 { font-family: var(--font-display); font-size: 28px; font-weight: 600; letter-spacing: -0.02em; margin-top: 6px; }
  .badges { display: flex; gap: 10px; }
  .chip { display: inline-flex; align-items: center; gap: 7px; height: 34px; padding: 0 13px; border-radius: 999px; font-family: var(--font-ui); font-size: 13px; font-weight: 500; }
  .chip .d { width: 7px; height: 7px; border-radius: 50%; }
  .chip.warn { background: var(--warning-bg); border: 1px solid rgba(251, 191, 36, 0.3); color: var(--warning); }
  .chip.warn .d { background: var(--warning); }
  .chip.ok { background: var(--success-bg); border: 1px solid rgba(74, 222, 128, 0.3); color: var(--success); }
  .chip.ok .d { background: var(--success); }
  .chip.ok .d.glow { box-shadow: 0 0 8px var(--success); }
  .body { flex: 1; min-height: 0; display: flex; flex-direction: column; gap: 16px; padding: 18px 36px 100px; }
  .gauges { display: grid; grid-template-columns: repeat(3, 1fr); gap: 16px; }
  .g-inner { display: flex; align-items: center; gap: 16px; }
  .g-meta { flex: 1; min-width: 0; }
  .g-detail { font-size: 13px; color: var(--text-2); margin: 6px 0 8px; }
  .cols { flex: 1; min-height: 0; display: flex; gap: 16px; }
  .containers { flex: 1.5; min-width: 0; background: var(--surface-1); border: 1px solid var(--border-violet); border-radius: var(--radius-card); box-shadow: inset 0 1px 0 var(--highlight-top); padding: 16px 18px; position: relative; }
  .ct-head { display: flex; align-items: center; justify-content: space-between; margin-bottom: 8px; }
  .ct-title { font-family: var(--font-display); font-size: 15px; font-weight: 600; }
  .ct-sub { font-size: 11px; color: var(--text-3); }
  .crow { display: flex; align-items: center; gap: 12px; padding: 11px 8px; border-radius: 10px; }
  .crow.divider { border-bottom: 1px solid var(--border-subtle); border-radius: 0; }
  .crow.warn { background: var(--warning-bg); }
  .cdot { width: 8px; height: 8px; border-radius: 50%; }
  .cname { flex: 1; font-family: var(--font-mono); font-size: 13px; color: var(--text-1); }
  .cimg { font-size: 11px; color: var(--text-3); width: 78px; }
  .cstatus { display: inline-flex; align-items: center; gap: 5px; font-size: 11px; color: var(--warning); width: 78px; }
  .spin { display: inline-grid; animation: aqno-spin 900ms linear infinite; }
  .ccpu { font-size: 12px; color: var(--text-2); width: 48px; text-align: right; }
  .cmem { font-size: 12px; color: var(--text-2); width: 70px; text-align: right; }
  .cbtn { width: 30px; height: 30px; border-radius: 8px; display: grid; place-items: center; border: none; background: none; cursor: pointer; }
  .confirm { position: absolute; left: 18px; right: 18px; bottom: 16px; display: flex; align-items: center; gap: 14px; padding: 14px 16px; border-radius: var(--radius-md); background: var(--surface-3); border: 1px solid var(--purple-024); box-shadow: var(--shadow-3), inset 0 1px 0 var(--highlight-top), var(--glow-md); }
  .cf-icon { width: 34px; height: 34px; flex: none; border-radius: 9px; background: var(--warning-bg); border: 1px solid rgba(251, 191, 36, 0.3); display: grid; place-items: center; }
  .cf-body { flex: 1; }
  .cf-title { font-family: var(--font-ui); font-size: 14px; font-weight: 600; }
  .cf-sub { font-size: 12.5px; color: var(--text-2); }
  .logs { flex: 1; min-width: 0; background: var(--surface-sunken); border: 1px solid var(--border-violet); border-radius: var(--radius-card); box-shadow: inset 0 1px 0 var(--highlight-top); padding: 14px 16px; display: flex; flex-direction: column; }
  .logs-head { display: flex; align-items: center; justify-content: space-between; margin-bottom: 12px; }
  .stream { display: flex; align-items: center; gap: 6px; font-size: 11px; color: var(--success); }
  .sdot { width: 6px; height: 6px; border-radius: 50%; background: var(--success); box-shadow: 0 0 6px var(--success); }
  .log-lines { display: flex; flex-direction: column; gap: 9px; font-family: var(--font-mono); font-size: 11.5px; line-height: 1.4; }
  .log { display: flex; gap: 9px; }
  .lt { color: var(--text-3); }
  .ll { white-space: pre; }
  .lb { color: var(--text-2); }
</style>
