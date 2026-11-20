<template>
  <section class="yc-field" :class="{ 'with-sign': $slots.sign }">
    <div class="yc-field-sky" aria-hidden="true">
      <i class="yc-field-sun" />
      <i class="yc-field-cloud a" /><i class="yc-field-cloud b" />
      <svg viewBox="0 0 1200 90" preserveAspectRatio="none">
        <path class="far" d="M0 90V52c130-44 270-46 410-10s300 38 420 4 250-40 370-4V90z" />
        <path class="near" d="M0 90V70c170-32 340-28 520-6s350 22 680-10V90z" />
      </svg>
    </div>
    <div v-if="$slots.sign" class="yc-field-sign"><slot name="sign" /></div>
    <div class="yc-field-soil"><slot /></div>
  </section>
</template>

<script setup lang="ts">
// The garden drawn as a field: a sky, hills, and soil in which the parent lays out `.yc-plot` beds.
// The field keeps its own dark ground in both themes, so the styles below are not scoped: neighbours' beds reuse them.
</script>

<style>
.yc .yc-field, .yc .yc-field-inset {
  --f-soil: #0b0f0c; --f-mound: #1a221c; --f-edge: rgb(244 246 241 / 0.08); --f-ink: #f4f6f1; --f-mut: rgb(244 246 241 / 0.66);
  --yc-plant-line: var(--f-mound); --yc-plant-stem: #c8ff00; --yc-plant-bloom: #ffffff; --yc-plant-soil: transparent;
  color: var(--f-ink);
}
.yc .yc-field { position: relative; margin-top: 28px; overflow: hidden; border-radius: 28px; background: var(--f-soil); }
.yc .yc-field.has-yard { border-radius: 28px 28px 0 0; }
.yc .yc-field-sky { position: relative; height: 128px; background: var(--yc-lime); }
.yc .yc-field-sky svg { position: absolute; inset: auto 0 -1px; width: 100%; height: 78px; }
.yc .yc-field-sky .far { fill: rgb(11 15 12 / 0.16); }
.yc .yc-field-sky .near { fill: var(--f-soil); }
.yc .yc-field-sun { position: absolute; top: 22px; right: 9%; width: 46px; height: 46px; border: 2px solid #0b0f0c; border-radius: 50%; background: #ffffff; }
.yc .yc-field-cloud { position: absolute; height: 18px; border: 2px solid #0b0f0c; border-radius: 999px; background: #ffffff; animation: yc-field-drift 38s linear infinite alternate; }
.yc .yc-field-cloud.a { top: 26px; left: 12%; width: 74px; }
.yc .yc-field-cloud.b { top: 54px; left: 46%; width: 48px; animation-duration: 52s; }
.dark .yc .yc-field { box-shadow: 0 0 0 1px var(--yc-line); }
.dark .yc .yc-field-sky { background: #141c17; }
.dark .yc .yc-field-sky .far { fill: rgb(200 255 0 / 0.1); }
.dark .yc .yc-field-sun { border-color: transparent; background: var(--yc-lime); box-shadow: 0 0 34px rgb(200 255 0 / 0.35); }
.dark .yc .yc-field-cloud { border-color: transparent; background: rgb(244 246 241 / 0.14); }
.yc .yc-field-sign { position: absolute; top: 20px; left: 26px; max-width: 60%; color: #0b0f0c; }
.dark .yc .yc-field-sign { color: var(--f-ink); }
.yc .yc-field.with-sign .yc-field-sky { height: 206px; }
.yc .yc-field-sign .yc-btn { background: #0b0f0c; color: #ffffff; }
.dark .yc .yc-field-sign .yc-btn { background: var(--yc-lime); color: #0b0f0c; }
.yc .yc-field-soil {
  display: grid; grid-template-columns: repeat(auto-fill, minmax(172px, 1fr)); gap: 14px; padding: 6px 24px 26px;
  background: repeating-linear-gradient(0deg, transparent 0 30px, rgb(244 246 241 / 0.03) 30px 32px);
}
.yc .yc-field-inset { padding: 16px; border-radius: 20px; background-color: var(--f-soil); }

.yc .yc-plot { position: relative; display: flex; flex-direction: column; align-items: center; gap: 4px; min-height: 212px; padding: 12px 12px 16px; border: 1px solid var(--f-edge); border-radius: 20px; background: var(--f-mound); text-align: center; }
.yc .yc-plot.ripe { border-color: var(--yc-lime); box-shadow: 0 0 0 1px var(--yc-lime) inset, 0 10px 34px -14px rgb(200 255 0 / 0.55); }
.yc .yc-plot.empty { justify-content: center; border-style: dashed; background: transparent; color: var(--f-mut); font-size: 13.5px; }
.yc .yc-plot.empty::before { content: ''; width: 46px; height: 12px; margin-bottom: 8px; border-radius: 50%; background: var(--f-mound); }
.yc .yc-plot-bed { position: relative; display: grid; place-items: end center; width: 100%; height: 104px; }
.yc .yc-plot-bed::before { content: ''; position: absolute; inset: auto 16% 2px; height: 16px; border-radius: 50%; background: rgb(0 0 0 / 0.45); }
.yc .yc-plot-plant { position: relative; width: 100px; height: 100px; transform-origin: 50% 88%; }
.yc .yc-plot.ripe .yc-plot-plant { animation: yc-field-sway 2.6s ease-in-out infinite; }
.yc .yc-plot-amount { font-size: 22px; line-height: 1.2; letter-spacing: -0.02em; font-weight: 800; }
.yc .yc-plot-name { font-size: 15px; font-weight: 800; }
.yc .yc-plot-state { min-height: 20px; font-size: 13.5px; line-height: 1.45; color: var(--f-mut); }
.yc .yc-plot.ripe .yc-plot-state { color: var(--yc-lime); font-weight: 650; }
.yc .yc-plot-bar { width: 100%; height: 6px; margin-top: 8px; overflow: hidden; border-radius: 999px; background: rgb(244 246 241 / 0.12); }
.yc .yc-plot-bar i { display: block; height: 100%; border-radius: inherit; background: var(--yc-lime); }
.yc .yc-plot .yc-btn { margin-top: 8px; }
.yc .yc-plot .yc-btn-ghost { box-shadow: inset 0 0 0 1px rgb(244 246 241 / 0.32); background: transparent; color: var(--f-ink); }
.yc .yc-plot .yc-btn-ghost:hover { background: rgb(244 246 241 / 0.08); }
.yc .yc-plot .yc-badge { position: absolute; top: 10px; left: 10px; }

/* What just happened to a bed: watered, taken from, or the dog caught the visitor. */
.yc .yc-plot.fx-water .yc-plot-bed::after {
  content: ''; position: absolute; inset: 0 22% auto; height: 22px;
  background: radial-gradient(circle 3.5px at 12% 50%, #ffffff 98%, transparent), radial-gradient(circle 3.5px at 50% 20%, #ffffff 98%, transparent), radial-gradient(circle 3.5px at 88% 60%, #ffffff 98%, transparent);
  animation: yc-field-rain 0.9s ease-in forwards;
}
.yc .yc-plot.fx-steal .yc-plot-plant { animation: yc-field-pluck 0.9s ease-out; }
.yc .yc-plot.fx-caught { animation: yc-field-shake 0.5s linear; border-color: #ff5a4f; }

@keyframes yc-field-sway { 0%, 100% { transform: rotate(-3deg); } 50% { transform: rotate(3deg); } }
@keyframes yc-field-drift { from { transform: translateX(-30px); } to { transform: translateX(90px); } }
@keyframes yc-field-rain { from { transform: translateY(-6px); opacity: 0; } 30% { opacity: 1; } to { transform: translateY(78px); opacity: 0; } }
@keyframes yc-field-pluck { 35% { transform: translateY(-26px) scale(1.08) rotate(-6deg); } 70% { transform: translateY(-4px) scale(0.94); } }
@keyframes yc-field-shake { 20%, 60% { transform: translateX(-6px); } 40%, 80% { transform: translateX(6px); } }
@media (max-width: 760px) {
  .yc .yc-field-sky { height: 104px; }
  .yc .yc-field.with-sign .yc-field-sky { height: 236px; }
  .yc .yc-field-sign { top: 16px; left: 18px; max-width: 78%; }
  .yc .yc-field-soil { grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 10px; padding: 4px 12px 16px; }
  .yc .yc-plot { min-height: 196px; padding: 10px 8px 14px; }
  .yc .yc-plot-plant { width: 84px; height: 84px; }
  .yc .yc-plot-bed { height: 88px; }
  .yc .yc-field-inset { padding: 10px; }
}
@media (prefers-reduced-motion: reduce) {
  .yc .yc-field-cloud, .yc .yc-plot .yc-plot-plant, .yc .yc-plot.fx-caught, .yc .yc-plot.fx-water .yc-plot-bed::after { animation: none; }
}
</style>
