<template>
  <span class="yc-lockup" :class="{ 'is-stacked': stacked }">
    <img :src="logo || '/logo.svg'" alt="" />
    <span class="yc-lockup-text">
      <b>{{ name }}</b>
    </span>
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useAppStore } from '@/stores'
import { sanitizeUrl } from '@/utils/url'
import { ycSiteName } from '@/utils/ycBrand'

defineProps<{ stacked?: boolean }>()

const appStore = useAppStore()
const rawName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName)
const name = computed(() => ycSiteName(rawName.value))
const logo = computed(() =>
  sanitizeUrl(appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '', {
    allowRelative: true,
    allowDataUrl: true
  })
)
</script>

<style scoped>
.yc-lockup { display: inline-flex; align-items: center; gap: 10px; white-space: nowrap; }
.yc-lockup img { width: 32px; height: 32px; border-radius: 9px; object-fit: contain; flex: none; }
.yc-lockup-text { display: flex; flex-direction: column; line-height: 1.05; }
.yc-lockup b { font-size: 17px; font-weight: 800; letter-spacing: 0.02em; color: var(--yc-ink); }
.yc-lockup i {
  margin-top: 3px;
  font-size: 9.5px;
}

.yc-lockup.is-stacked { flex-direction: column; gap: 14px; }
.yc-lockup.is-stacked img { width: 64px; height: 64px; border-radius: 18px; }
.yc-lockup.is-stacked .yc-lockup-text { align-items: center; }
.yc-lockup.is-stacked b { font-size: 30px; letter-spacing: 0.04em; }
.yc-lockup.is-stacked i { margin-top: 8px; font-size: 12px; letter-spacing: 0.32em; margin-right: -0.32em; }
</style>
