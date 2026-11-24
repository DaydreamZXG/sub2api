<template>
  <figure class="yc-flow" :aria-label="t('yc.flow.aria')">
    <svg :viewBox="`0 0 ${box.w} ${box.h}`" role="img" :aria-label="t('yc.flow.aria')">
      <g fill="none" stroke-linecap="round">
        <path v-for="l in links" :key="l.id" :d="l.d" class="yc-flow-line" />
      </g>
      <g class="yc-flow-dots" aria-hidden="true">
        <circle v-for="l in links" :key="`d${l.id}`" r="5" class="yc-flow-dot">
          <animateMotion :dur="`${l.dur}s`" :begin="`-${l.delay}s`" repeatCount="indefinite" :path="l.d" />
        </circle>
      </g>
      <g v-for="n in nodes" :key="n.id" :transform="`translate(${n.x - n.w / 2} ${n.y - nodeH / 2})`">
        <rect :width="n.w" :height="nodeH" :rx="nodeH / 2" class="yc-flow-node" />
        <text :x="n.w / 2" :y="nodeH / 2" class="yc-flow-label" text-anchor="middle" dominant-baseline="central">{{ n.label }}</text>
      </g>
      <ellipse :cx="hub.x" :cy="hub.y" :rx="hub.w * 0.62" :ry="hub.h * 0.95" class="yc-flow-halo" aria-hidden="true" />
      <g :transform="`translate(${hub.x - hub.w / 2} ${hub.y - hub.h / 2})`">
        <rect :width="hub.w" :height="hub.h" :rx="hub.h / 2" class="yc-flow-hub" />
        <text :x="hub.w / 2" :y="hub.h / 2 - 9" class="yc-flow-hub-title" text-anchor="middle" dominant-baseline="central">{{ t('yc.flow.key') }}</text>
        <text :x="hub.w / 2" :y="hub.h / 2 + 12" class="yc-flow-hub-key" text-anchor="middle" dominant-baseline="central">sk-••••••••</text>
      </g>
    </svg>
  </figure>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useMediaQuery } from '@vueuse/core'

const { t } = useI18n()
const vertical = useMediaQuery('(max-width: 720px)')

const TOOLS = ['Claude Code', 'Codex', 'Gemini CLI', 'OpenCode']
const MODELS = ['Claude', 'GPT', 'Gemini', 'Grok', 'DeepSeek']
const nodeH = 44

interface FlowNode { id: string; label: string; x: number; y: number; w: number }

const box = computed(() => (vertical.value ? { w: 360, h: 560 } : { w: 1000, h: 380 }))
const hub = computed(() =>
  vertical.value ? { x: 180, y: 280, w: 190, h: 72 } : { x: 500, y: 190, w: 210, h: 76 }
)

function spread(count: number, from: number, to: number): number[] {
  if (count === 1) return [(from + to) / 2]
  return Array.from({ length: count }, (_, i) => from + ((to - from) * i) / (count - 1))
}

// On phones the two sides stack above and below the key, two per row.
function stack(labels: string[], top: number, prefix: string): FlowNode[] {
  return labels.map((label, i) => ({
    id: `${prefix}${i}`,
    label,
    x: i % 2 === 0 ? 96 : 264,
    y: top + Math.floor(i / 2) * 58,
    w: 150
  }))
}

const tools = computed<FlowNode[]>(() => {
  if (vertical.value) return stack(TOOLS, 40, 't')
  const ys = spread(TOOLS.length, 50, 330)
  return TOOLS.map((label, i) => ({ id: `t${i}`, label, x: 110, y: ys[i], w: 170 }))
})
const models = computed<FlowNode[]>(() => {
  if (vertical.value) return stack(MODELS, 404, 'm')
  const ys = spread(MODELS.length, 36, 344)
  return MODELS.map((label, i) => ({ id: `m${i}`, label, x: 890, y: ys[i], w: 150 }))
})
const nodes = computed(() => [...tools.value, ...models.value])

function curve(ax: number, ay: number, bx: number, by: number): string {
  if (vertical.value) {
    const my = (ay + by) / 2
    return `M${ax} ${ay} C${ax} ${my} ${bx} ${my} ${bx} ${by}`
  }
  const mx = (ax + bx) / 2
  return `M${ax} ${ay} C${mx} ${ay} ${mx} ${by} ${bx} ${by}`
}

const links = computed(() => {
  const h = hub.value
  const v = vertical.value
  const inbound = tools.value.map((n, i) => ({
    id: n.id,
    d: v ? curve(n.x, n.y + nodeH / 2, h.x, h.y - h.h / 2) : curve(n.x + n.w / 2, n.y, h.x - h.w / 2, h.y),
    dur: 2.6 + (i % 3) * 0.5,
    delay: i * 0.55
  }))
  const outbound = models.value.map((n, i) => ({
    id: n.id,
    d: v ? curve(h.x, h.y + h.h / 2, n.x, n.y - nodeH / 2) : curve(h.x + h.w / 2, h.y, n.x - n.w / 2, n.y),
    dur: 2.4 + (i % 3) * 0.45,
    delay: 0.3 + i * 0.5
  }))
  return [...inbound, ...outbound]
})
</script>

<style scoped>
.yc-flow { margin: 0; }
.yc-flow svg { display: block; width: 100%; height: auto; overflow: visible; }
.yc-flow-line { stroke: var(--yc-line); stroke-width: 1.5; }
.yc-flow-dot { fill: var(--yc-lime); stroke: var(--yc-ink); stroke-width: 1.5; }
.yc-flow-halo { fill: var(--yc-lime); opacity: 0.35; filter: blur(22px); transform-box: fill-box; transform-origin: center; animation: yc-halo 3.6s ease-in-out infinite; }
@keyframes yc-halo { 50% { opacity: 0.6; transform: scale(1.12); } }
.yc-flow-node { fill: var(--yc-paper); stroke: var(--yc-line); stroke-width: 1; filter: drop-shadow(0 6px 10px rgb(11 15 12 / 0.07)); }
.yc-flow-label { fill: var(--yc-ink); font-size: 15px; font-weight: 600; }
.yc-flow-hub { fill: var(--yc-lime); stroke: var(--yc-on-lime); stroke-width: 1.5; }
.yc-flow-hub-title { fill: var(--yc-on-lime); font-size: 16px; font-weight: 750; }
.yc-flow-hub-key { fill: var(--yc-on-lime); font-size: 13px; opacity: 0.72; font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; }
:global(.dark .yc .yc-flow-dot) { stroke: none; filter: drop-shadow(0 0 6px rgb(0 113 227 / 0.9)); }
:global(.dark .yc .yc-flow-halo) { opacity: 0.22; }
:global(.dark .yc .yc-flow-hub) { stroke: none; filter: drop-shadow(0 0 28px rgb(0 113 227 / 0.35)); }
@media (prefers-reduced-motion: reduce) {
  .yc-flow-dots { display: none; }
  .yc-flow-halo { animation: none; }
}
</style>
