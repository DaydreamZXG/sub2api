<template>
  <YcShell legacy>
    <header v-if="title" class="yc-legacy-head">
      <h1>{{ title }}</h1>
      <p v-if="description">{{ description }}</p>
    </header>
    <slot />
  </YcShell>
</template>

<script setup lang="ts">
// Frames a page from the original UI in the redesigned sidebar; the page itself is left as it is.
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import { resolveRouteMetaKeys } from '@/router/title'
import { resolveSiteBillingMode } from '@/utils/siteBillingMode'
import YcShell from './YcShell.vue'

const { t } = useI18n()
const route = useRoute()
const appStore = useAppStore()

const keys = computed(() =>
  resolveRouteMetaKeys(route, { billingMode: resolveSiteBillingMode(appStore.cachedPublicSettings) })
)
const title = computed(() => {
  if (route.name === 'CustomPage') {
    const item = appStore.cachedPublicSettings?.custom_menu_items?.find((i) => i.id === route.params.id)
    if (item?.label) return item.label
  }
  return keys.value.titleKey ? t(keys.value.titleKey) : (route.meta.title as string) || ''
})
const description = computed(() =>
  keys.value.descriptionKey ? t(keys.value.descriptionKey) : (route.meta.description as string) || ''
)
</script>

<style scoped>
.yc-legacy-head { display: flex; flex-wrap: wrap; align-items: baseline; gap: 4px 14px; margin-bottom: 20px; }
.yc-legacy-head h1 { font-size: 28px; line-height: 1.25; letter-spacing: -0.03em; font-weight: 800; }
.yc-legacy-head p { font-size: 13.5px; color: var(--yc-mut); }
</style>
