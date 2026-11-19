<template>
  <div class="locale-toggle" role="group">
    <button
      v-for="item in availableLocales"
      :key="item.code"
      type="button"
      :lang="item.code"
      :title="item.name"
      :disabled="switching"
      :aria-pressed="item.code === locale"
      :class="{ 'is-on': item.code === locale }"
      @click="selectLocale(item.code)"
    >
      {{ SHORT_NAMES[item.code] ?? item.code.toUpperCase() }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { setLocale, availableLocales } from '@/i18n'

const SHORT_NAMES: Record<string, string> = { zh: '中', en: 'EN' }

const { locale } = useI18n()
const switching = ref(false)

async function selectLocale(code: string) {
  if (switching.value || code === locale.value) return
  switching.value = true
  try {
    await setLocale(code)
    // 分组名称等内容由后端按请求语言返回，整页刷新让已加载的数据换成新语言
    window.location.reload()
  } finally {
    switching.value = false
  }
}
</script>

<style scoped>
.locale-toggle {
  @apply inline-flex items-center rounded-full bg-gray-100 p-0.5 dark:bg-dark-800;
}
.locale-toggle button {
  @apply h-7 min-w-[30px] rounded-full px-2 text-xs font-semibold text-gray-500 transition-colors dark:text-dark-400;
}
.locale-toggle button:hover {
  @apply text-gray-900 dark:text-white;
}
.locale-toggle button.is-on {
  @apply bg-white text-gray-900 shadow-sm dark:bg-dark-600 dark:text-white;
}
</style>
