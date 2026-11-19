<template>
  <div class="yc yc-shell">
    <header class="yc-shell-top">
      <div class="yc-wrap yc-shell-in">
        <router-link to="/dashboard" class="yc-shell-brand">
          <img :src="siteLogo || '/logo.svg'" alt="" />
          <span>{{ siteName }}</span>
        </router-link>
        <nav class="yc-shell-nav" :aria-label="t('yc.nav.aria')">
          <router-link v-for="l in links" :key="l.to" :to="l.to" active-class="is-on">{{ t(l.label) }}</router-link>
        </nav>
        <router-link v-if="authStore.isAdmin" to="/admin/dashboard" class="yc-shell-side">{{ t('yc.shell.admin') }}</router-link>
        <router-link to="/profile" class="yc-shell-me" :title="t('yc.shell.account')">{{ initial }}</router-link>
      </div>
    </header>
    <main class="yc-wrap yc-shell-main"><slot /></main>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore, useAppStore } from '@/stores'
import { sanitizeUrl } from '@/utils/url'
import '@/styles/yecai.css'

const { t } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()

const settings = computed(() => appStore.cachedPublicSettings)
const siteName = computed(() => settings.value?.site_name || appStore.siteName || 'Sub2API')
const siteLogo = computed(() =>
  sanitizeUrl(settings.value?.site_logo || appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true })
)
const initial = computed(() => (authStore.user?.username || authStore.user?.email || '?').charAt(0).toUpperCase())
const links = computed(() => [
  { to: '/dashboard', label: 'yc.shell.start' },
  { to: '/keys', label: 'yc.shell.keys' },
  { to: '/usage', label: 'yc.shell.bills' },
  { to: settings.value?.payment_enabled ? '/purchase' : '/redeem', label: 'yc.shell.topUp' }
])
</script>

<style scoped>
.yc-shell { min-height: 100vh; }
.yc-shell-top { position: sticky; top: 0; z-index: 20; background: color-mix(in srgb, var(--yc-bg) 84%, transparent); backdrop-filter: saturate(160%) blur(18px); -webkit-backdrop-filter: saturate(160%) blur(18px); border-bottom: 1px solid var(--yc-line); }
.yc-shell-in { height: 60px; display: flex; align-items: center; gap: 24px; }
.yc-shell-brand { display: flex; align-items: center; gap: 10px; font-weight: 750; font-size: 16.5px; white-space: nowrap; }
.yc-shell-brand img { width: 28px; height: 28px; border-radius: 8px; object-fit: contain; }
.yc-shell-nav { display: flex; gap: 4px; margin-right: auto; overflow-x: auto; scrollbar-width: none; }
.yc-shell-nav a { padding: 7px 14px; border-radius: 999px; font-size: 14.5px; color: var(--yc-mut); white-space: nowrap; }
.yc-shell-nav a:hover { color: var(--yc-ink); }
.yc-shell-nav a.is-on { background: var(--yc-lime); color: var(--yc-on-lime); font-weight: 650; }
.yc-shell-side { font-size: 13.5px; color: var(--yc-mut); white-space: nowrap; }
.yc-shell-me { flex: none; width: 34px; height: 34px; display: grid; place-items: center; border-radius: 50%; background: var(--yc-fill); box-shadow: inset 0 0 0 1px var(--yc-line); font-weight: 700; font-size: 14px; }
.yc-shell-main { padding-top: 36px; padding-bottom: 72px; }
@media (max-width: 640px) {
  .yc-shell-in { gap: 12px; }
  .yc-shell-brand span, .yc-shell-side { display: none; }
}
</style>
