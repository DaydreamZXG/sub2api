<template>
  <YcShell>
    <router-link to="/help" class="yc-guide-back">← {{ t('yc.help.title') }}</router-link>
    <h1 class="yc-page-h1">{{ t('yc.agentGuide.title') }}</h1>
    <p class="yc-page-lead">{{ t('yc.agentGuide.lead') }}</p>

    <section class="yc-surface yc-ag" data-test="agent-guide-terms">
      <h2 class="yc-sec-h2">{{ t('yc.agentGuide.termsTitle') }}</h2>
      <dl class="yc-ag-terms">
        <div v-for="k in TERMS" :key="k">
          <dt>{{ t(`yc.agentGuide.terms.${k}.name`) }}</dt>
          <dd>{{ t(`yc.agentGuide.terms.${k}.text`) }}</dd>
        </div>
      </dl>
      <p class="yc-ag-example">{{ t('yc.agentGuide.example') }}</p>
    </section>

    <section v-for="s in SECTIONS" :key="s.key" class="yc-surface yc-ag" :data-test="`agent-guide-${s.key}`">
      <h2 class="yc-sec-h2">{{ t(`yc.agentGuide.${s.key}.title`) }}</h2>
      <ul class="yc-ag-list">
        <li v-for="n in s.items" :key="n">{{ t(`yc.agentGuide.${s.key}.i${n}`) }}</li>
      </ul>
    </section>

    <section class="yc-surface yc-ag" data-test="agent-guide-apply">
      <h2 class="yc-sec-h2">{{ t('yc.agentGuide.apply.title') }}</h2>
      <p class="yc-ag-text">{{ t('yc.agentGuide.apply.text') }}</p>
      <p v-if="contact" class="yc-ag-contact" data-test="agent-guide-contact">{{ contact }}</p>
      <router-link v-else to="/help" class="yc-btn yc-btn-ghost yc-btn-sm">{{ t('yc.agentGuide.apply.toHelp') }} →</router-link>
    </section>
  </YcShell>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import YcShell from '@/components/yecai/YcShell.vue'

const TERMS = ['rate', 'basis', 'start', 'duration'] as const
// How many points each section has in the locale files.
const SECTIONS = [
  { key: 'join', items: [1, 2] },
  { key: 'rules', items: [1, 2, 3, 4, 5, 6] },
  { key: 'payout', items: [1, 2, 3] }
] as const

const { t } = useI18n()
const appStore = useAppStore()
const contact = computed(() => appStore.cachedPublicSettings?.contact_info || '')
</script>

<style scoped>
.yc-guide-back { display: inline-block; margin-bottom: 14px; font-size: 14px; color: var(--yc-dim); text-decoration: none; }
.yc-guide-back:hover { color: var(--yc-ink); }
.yc-ag { margin-top: 20px; padding: 26px 30px; }
.yc-ag .yc-btn { margin-top: 14px; }
.yc-ag-terms { margin: 14px 0 0; }
.yc-ag-terms > div { display: grid; grid-template-columns: 120px 1fr; gap: 16px; padding: 13px 0; border-top: 1px solid var(--yc-line); }
.yc-ag-terms dt { font-size: 14px; color: var(--yc-dim); }
.yc-ag-terms dd { margin: 0; font-size: 15.5px; line-height: 1.65; }
.yc-ag-example { margin: 8px 0 0; padding: 14px 16px; border-radius: 12px; background: color-mix(in srgb, var(--yc-lime) 22%, transparent); font-size: 15px; line-height: 1.7; }
.yc-ag-list { margin: 12px 0 0; padding-left: 20px; list-style: disc; font-size: 15.5px; line-height: 1.75; }
.yc-ag-list li + li { margin-top: 6px; }
.yc-ag-contact { margin: 10px 0 0; font-size: 16px; line-height: 1.7; white-space: pre-wrap; overflow-wrap: anywhere; }
.yc-ag-text { margin: 10px 0 0; font-size: 15.5px; line-height: 1.7; }
@media (max-width: 640px) {
  .yc-ag { padding: 22px 20px; }
  .yc-ag-terms > div { grid-template-columns: 1fr; gap: 2px; }
}
</style>
