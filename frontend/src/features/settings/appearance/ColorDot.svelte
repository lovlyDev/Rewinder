<script lang="ts">
  type Props = { value: string; label: string; oninput?: (v: string) => void; onchange?: (v: string) => void };
  let { value = $bindable(), label, oninput, onchange }: Props = $props();
</script>

<!-- A round swatch that opens the system colour picker. -->
<label class="dot" style="--c:{value}">
  <input
    type="color"
    {value}
    aria-label={label}
    oninput={(e) => {
      value = e.currentTarget.value;
      oninput?.(value);
    }}
    onchange={(e) => onchange?.(e.currentTarget.value)}
  />
  <span class="tabular">{value.toUpperCase()}</span>
</label>

<style>
  .dot {
    position: relative;
    display: inline-flex;
    align-items: center;
    gap: 10px;
    height: 38px;
    padding: 0 14px 0 6px;
    border-radius: 999px;
    background: var(--overlay);
    cursor: pointer;
    transition: transform var(--quick) var(--ease);
  }
  .dot:active {
    transform: scale(0.97);
  }
  .dot::before {
    content: '';
    width: 26px;
    height: 26px;
    border-radius: 999px;
    background: var(--c);
    box-shadow: inset 0 0 0 1px #ffffff1f;
  }
  span {
    color: var(--muted);
    font-size: 12px;
    font-weight: 550;
  }
  input {
    position: absolute;
    inset: 0;
    opacity: 0;
    cursor: pointer;
  }
</style>
