<template>
  <YcShell>
    <router-link to="/help" class="yc-guide-back">← {{ t('yc.help.title') }}</router-link>
    <p v-if="locale !== 'zh'" class="yc-note yc-guide-lang">{{ t('yc.guide.zhOnly') }}</p>
    <div ref="root" class="yc-guide">
      <p v-if="failed" class="yc-note">{{ t('yc.guide.loadFailed') }}</p>
    </div>
  </YcShell>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import YcShell from '@/components/yecai/YcShell.vue'
import { mountGuide, type GuideData } from '@/components/yecai/guide/guideApp'
import '@/styles/yc-guide.css'

const { t, locale } = useI18n()
const root = ref<HTMLElement | null>(null)
const failed = ref(false)
let unmount: (() => void) | undefined

onMounted(async () => {
  try {
    const data = (await import('@/data/yc-guide.json')).default as GuideData
    if (root.value) unmount = mountGuide(root.value, data)
  } catch (error) {
    console.error('Failed to load the guide:', error)
    failed.value = true
  }
})

onBeforeUnmount(() => unmount?.())
</script>

<style scoped>
.yc-guide-back { display: inline-block; margin-bottom: 14px; font-size: 14px; color: var(--yc-mut); text-decoration: none; }
.yc-guide-back:hover { color: var(--yc-ink); }
.yc-guide-lang { margin-bottom: 14px; }
</style>
