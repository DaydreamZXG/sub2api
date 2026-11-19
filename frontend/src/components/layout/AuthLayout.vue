<template>
  <div class="yc relative flex min-h-screen items-center justify-center overflow-hidden p-4">
    <!-- A single soft lime glow, the same accent the home page uses -->
    <div class="yc-auth-glow pointer-events-none absolute" aria-hidden="true"></div>

    <!-- Content Container -->
    <div class="relative z-10 w-full max-w-md">
      <!-- Logo/Brand -->
      <div class="mb-8 text-center">
        <!-- Custom Logo or Default Logo -->
        <template v-if="settingsLoaded">
          <h1 class="mb-3">
            <router-link to="/home"><YcBrand stacked /></router-link>
          </h1>
          <p class="text-sm text-gray-500 dark:text-dark-400">
            {{ siteSubtitle }}
          </p>
        </template>
      </div>

      <!-- Card Container -->
      <div class="yc-surface p-8">
        <slot />
      </div>

      <!-- Footer Links -->
      <div class="mt-6 text-center text-sm">
        <slot name="footer" />
      </div>

      <!-- Copyright -->
      <div class="mt-8 text-center text-xs text-gray-400 dark:text-dark-500">
        &copy; {{ currentYear }} {{ siteName }}. All rights reserved.
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import YcBrand from '@/components/yecai/YcBrand.vue'
import { isStockSiteSubtitle, ycSiteName } from '@/utils/ycBrand'
import '@/styles/yecai.css'

const { t } = useI18n()
const appStore = useAppStore()

const siteName = computed(() => ycSiteName(appStore.siteName))
const siteSubtitle = computed(() => {
  const subtitle = appStore.cachedPublicSettings?.site_subtitle
  return isStockSiteSubtitle(subtitle) ? t('yc.hero.eyebrow') : subtitle
})
const settingsLoaded = computed(() => appStore.publicSettingsLoaded)

const currentYear = computed(() => new Date().getFullYear())

onMounted(() => {
  appStore.fetchPublicSettings()
})
</script>

<style scoped>
.yc-auth-glow { top: -22rem; left: 50%; width: 52rem; height: 40rem; transform: translateX(-50%); border-radius: 50%; background: radial-gradient(closest-side, rgb(200 255 0 / 0.28), transparent); }
:global(.dark) .yc-auth-glow { background: radial-gradient(closest-side, rgb(200 255 0 / 0.1), transparent); }
</style>
