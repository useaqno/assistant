<script lang="ts">
  let {
    value = 0,
    size = 96,
    thickness = 8,
    color = 'var(--purple)',
    label = '',
    caption = '',
    glow = true
  }: {
    value?: number;
    size?: number;
    thickness?: number;
    color?: string;
    label?: string;
    caption?: string;
    glow?: boolean;
  } = $props();

  let r = $derived((size - thickness) / 2);
  let circ = $derived(2 * Math.PI * r);
  let pct = $derived(Math.max(0, Math.min(1, value)));
  let dash = $derived(circ * pct);
</script>

<div class="ring" style="width:{size}px;height:{size}px">
  <svg width={size} height={size} style="transform:rotate(-90deg)">
    <circle cx={size / 2} cy={size / 2} {r} fill="none" stroke="var(--surface-3)" stroke-width={thickness} />
    <circle
      cx={size / 2}
      cy={size / 2}
      {r}
      fill="none"
      stroke={color}
      stroke-width={thickness}
      stroke-linecap="round"
      stroke-dasharray="{dash} {circ}"
    />
  </svg>
  {#if glow}
    <svg width={size} height={size} class="glow" style="transform:rotate(-90deg)">
      <circle
        cx={size / 2}
        cy={size / 2}
        {r}
        fill="none"
        stroke={color}
        stroke-width={thickness}
        stroke-linecap="round"
        stroke-dasharray="{dash} {circ}"
      />
    </svg>
  {/if}
  <div class="center">
    {#if label}<div class="val" style="font-size:{size * 0.26}px">{label}</div>{/if}
    {#if caption}<div class="cap" style="font-size:{size * 0.11}px">{caption}</div>{/if}
  </div>
</div>

<style>
  .ring { position: relative; display: inline-grid; place-items: center; }
  .glow { position: absolute; opacity: 0.5; filter: blur(4px); }
  .center { position: absolute; text-align: center; }
  .val {
    font-family: var(--font-mono);
    font-weight: var(--weight-semibold);
    color: var(--text-1);
    letter-spacing: -0.02em;
    line-height: 1;
    font-variant-numeric: tabular-nums;
  }
  .cap { font-family: var(--font-ui); color: var(--text-3); margin-top: 4px; }
</style>
