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
          <router-link to="/home" class="mb-4 inline-flex h-16 w-16 items-center justify-center overflow-hidden rounded-2xl">
            <img :src="siteLogo || '/logo.svg'" alt="Logo" class="h-full w-full object-contain" />
          </router-link>
          <h1 class="yc-auth-title mb-2">
            {{ siteName }}
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
import { useAppStore } from '@/stores'
import { sanitizeUrl } from '@/utils/url'
import '@/styles/yecai.css'

const appStore = useAppStore()

const siteName = computed(() => appStore.siteName || 'Sub2API')
const siteLogo = computed(() => sanitizeUrl(appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }))
const siteSubtitle = computed(() => appStore.cachedPublicSettings?.site_subtitle || 'Subscription to API Conversion Platform')
const settingsLoaded = computed(() => appStore.publicSettingsLoaded)

const currentYear = computed(() => new Date().getFullYear())

onMounted(() => {
  appStore.fetchPublicSettings()
})
</script>

<style scoped>
.yc-auth-title { font-size: 30px; line-height: 1.2; letter-spacing: -0.03em; font-weight: 800; color: var(--yc-ink); }
.yc-auth-glow { top: -22rem; left: 50%; width: 52rem; height: 40rem; transform: translateX(-50%); border-radius: 50%; background: radial-gradient(closest-side, rgb(200 255 0 / 0.28), transparent); }
:global(.dark) .yc-auth-glow { background: radial-gradient(closest-side, rgb(200 255 0 / 0.1), transparent); }
</style>
