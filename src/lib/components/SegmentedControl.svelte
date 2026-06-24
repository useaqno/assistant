<script lang="ts">
  let {
    options = [],
    value = $bindable(),
    size = 'md',
    full = false,
    onchange
  }: {
    options: string[];
    value?: string;
    size?: 'sm' | 'md';
    full?: boolean;
    onchange?: (v: string) => void;
  } = $props();

  // default to first option
  $effect(() => {
    if (value === undefined && options.length) value = options[0];
  });

  let idx = $derived(Math.max(0, options.indexOf(value ?? options[0])));
  const dims = { sm: { h: 32, fs: 13 }, md: { h: 38, fs: 14 } }[size];

  function select(v: string) {
    value = v;
    onchange?.(v);
  }
</script>

<div
  class="seg"
  class:full
  style="grid-template-columns:repeat({options.length},1fr);height:{dims.h + 6}px"
>
  <div
    class="thumb"
    style="left:calc({(idx / options.length) * 100}% + 3px);width:calc({100 / options.length}% - 6px)"
  ></div>
  {#each options as o}
    <button
      class="opt"
      class:on={o === value}
      style="height:{dims.h}px;font-size:{dims.fs}px"
      onclick={() => select(o)}>{o}</button
    >
  {/each}
</div>

<style>
  .seg {
    position: relative;
    display: inline-grid;
    padding: 3px;
    background: var(--surface-1);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-pill);
    box-shadow: inset 0 1px 2px rgba(0, 0, 0, 0.35);
  }
  .seg.full { display: grid; width: 100%; }
  .thumb {
    position: absolute;
    top: 3px;
    bottom: 3px;
    background: var(--surface-3);
    border-radius: var(--radius-pill);
    border: 1px solid var(--purple-024);
    box-shadow: inset 0 1px 0 var(--highlight-top), 0 1px 3px rgba(0, 0, 0, 0.3);
    transition: left var(--dur-base) var(--ease-out);
  }
  .opt {
    position: relative;
    z-index: 1;
    border: none;
    background: transparent;
    font-family: var(--font-ui);
    font-weight: var(--weight-medium);
    letter-spacing: var(--tracking-tight);
    cursor: pointer;
    padding: 0 16px;
    color: var(--text-3);
    transition: color var(--dur-base) var(--ease-out);
  }
  .opt.on { color: var(--text-1); }
</style>
