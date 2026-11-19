<template>
  <div class="yc yc-shell">
    <aside class="yc-side">
      <router-link to="/dashboard" class="yc-side-brand">
        <img :src="siteLogo || '/logo.svg'" alt="" />
        <span>{{ siteName }}</span>
      </router-link>
      <nav class="yc-side-nav" :aria-label="t('yc.nav.aria')">
        <router-link v-for="l in links" :key="l.to" :to="l.to" active-class="is-on">{{ t(l.label) }}</router-link>
      </nav>
      <div class="yc-side-foot">
        <router-link v-if="authStore.isAdmin" to="/admin/dashboard">{{ t('yc.shell.admin') }}</router-link>
        <router-link to="/profile" class="yc-side-me"><i>{{ initial }}</i><span>{{ t('yc.shell.account') }}</span></router-link>
      </div>
    </aside>
    <main class="yc-shell-main"><slot /></main>
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
.yc-shell { min-height: 100vh; display: grid; grid-template-columns: 232px minmax(0, 1fr); }
.yc-side { position: sticky; top: 0; height: 100vh; display: flex; flex-direction: column; padding: 22px 14px 18px; border-right: 1px solid var(--yc-line); background: color-mix(in srgb, var(--yc-paper) 70%, transparent); }
.yc-side-brand { display: flex; align-items: center; gap: 10px; padding: 0 10px 22px; font-weight: 750; font-size: 17px; white-space: nowrap; }
.yc-side-brand img { width: 30px; height: 30px; border-radius: 9px; object-fit: contain; }
.yc-side-nav { display: flex; flex-direction: column; gap: 2px; }
.yc-side-nav a { position: relative; padding: 10px 14px; border-radius: 12px; font-size: 15px; color: var(--yc-mut); white-space: nowrap; transition: background-color 0.15s ease, color 0.15s ease; }
.yc-side-nav a:hover { background: var(--yc-fill); color: var(--yc-ink); }
.yc-side-nav a.is-on { background: var(--yc-lime); color: var(--yc-on-lime); font-weight: 650; box-shadow: 0 8px 20px -10px rgb(140 180 0 / 0.9); }
.yc-side-foot { margin-top: auto; display: flex; flex-direction: column; gap: 4px; font-size: 14px; color: var(--yc-mut); }
.yc-side-foot a { display: flex; align-items: center; gap: 10px; padding: 9px 10px; border-radius: 12px; }
.yc-side-foot a:hover { background: var(--yc-fill); color: var(--yc-ink); }
.yc-side-me i { flex: none; width: 30px; height: 30px; display: grid; place-items: center; border-radius: 50%; background: var(--yc-ink); color: var(--yc-bg); font-style: normal; font-weight: 700; font-size: 13px; }
.yc-shell-main { width: 100%; max-width: 1080px; margin: 0 auto; padding: 44px 36px 80px; }
@media (max-width: 860px) {
  .yc-shell { grid-template-columns: 1fr; }
  .yc-side { position: sticky; z-index: 20; height: auto; flex-direction: row; align-items: center; gap: 8px; padding: 10px 14px; border-right: 0; border-bottom: 1px solid var(--yc-line); background: color-mix(in srgb, var(--yc-bg) 86%, transparent); backdrop-filter: blur(16px); -webkit-backdrop-filter: blur(16px); }
  .yc-side-brand { padding: 0; }
  .yc-side-brand span, .yc-side-me span { display: none; }
  .yc-side-nav { flex: 1; flex-direction: row; overflow-x: auto; scrollbar-width: none; }
  .yc-side-nav a { padding: 7px 12px; font-size: 14.5px; }
  .yc-side-foot { margin: 0; flex-direction: row; }
  .yc-side-foot a:not(.yc-side-me) { display: none; }
  .yc-side-foot a { padding: 0; }
  .yc-shell-main { padding: 24px 18px 56px; }
}
</style>
