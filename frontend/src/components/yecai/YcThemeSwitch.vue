<template>
  <div class="yc-theme-switch" role="group" :aria-label="t('yc.theme.label')">
    <button
      v-for="option in OPTIONS"
      :key="option"
      type="button"
      :title="t(`yc.theme.${option}`)"
      :aria-label="t(`yc.theme.${option}`)"
      :aria-pressed="mode === option"
      :class="{ 'is-on': mode === option }"
      @click="setMode(option)"
    >
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
        <template v-if="option === 'light'">
          <circle cx="12" cy="12" r="4" />
          <path d="M12 2.5v2.5M12 19v2.5M2.5 12H5M19 12h2.5M5.3 5.3l1.8 1.8M16.9 16.9l1.8 1.8M5.3 18.7l1.8-1.8M16.9 7.1l1.8-1.8" />
        </template>
        <path v-else-if="option === 'dark'" d="M20 14.5A8.5 8.5 0 0 1 9.5 4a8.5 8.5 0 1 0 10.5 10.5Z" />
        <template v-else>
          <rect x="3" y="4.5" width="18" height="12" rx="2" />
          <path d="M8.5 20h7M12 16.5V20" />
        </template>
      </svg>
    </button>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { useYcTheme, type YcThemeMode } from '@/composables/useYcTheme'

const OPTIONS: YcThemeMode[] = ['light', 'dark', 'system']

const { t } = useI18n()
const { mode, setMode } = useYcTheme()
</script>

<style scoped>
.yc-theme-switch { display: inline-flex; align-items: center; padding: 2px; border-radius: 999px; background: var(--yc-fill); }
.yc-theme-switch button { display: grid; place-items: center; width: 28px; height: 28px; border-radius: 999px; color: var(--yc-dim); transition: background-color 0.15s ease, color 0.15s ease; }
.yc-theme-switch button:hover { color: var(--yc-ink); }
.yc-theme-switch button.is-on { background: var(--yc-paper); color: var(--yc-ink); box-shadow: 0 1px 2px rgb(11 15 12 / 0.12); }
.yc-theme-switch svg { width: 15px; height: 15px; }
</style>
