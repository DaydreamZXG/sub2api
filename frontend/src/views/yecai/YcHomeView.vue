<template>
  <UpstreamHome v-if="useUpstreamHome" />
  <div v-else class="yc yc-home">
    <header class="yc-top">
      <div class="yc-wrap yc-top-in">
        <router-link to="/home" class="yc-brand">
          <img :src="siteLogo || '/logo.svg'" alt="" class="yc-brand-logo" />
          <span>{{ siteName }}</span>
        </router-link>
        <nav class="yc-top-nav" :aria-label="t('yc.nav.aria')">
          <a href="#how">{{ t('yc.nav.how') }}</a>
          <router-link v-if="showModelPlaza" to="/model-plaza">{{ t('yc.nav.prices') }}</router-link>
          <a href="#service">{{ t('yc.nav.service') }}</a>
          <a href="#faq">{{ t('yc.nav.faq') }}</a>
        </nav>
        <div class="yc-top-actions">
          <LocaleSwitcher />
          <router-link v-if="isAuthenticated" :to="consolePath" class="yc-btn yc-btn-lime yc-btn-sm">{{ t('yc.cta.console') }}</router-link>
          <template v-else>
            <router-link to="/login" class="yc-btn yc-btn-ghost yc-btn-sm yc-hide-sm">{{ t('yc.cta.login') }}</router-link>
            <router-link :to="startPath" class="yc-btn yc-btn-lime yc-btn-sm">{{ t('yc.cta.start') }}</router-link>
          </template>
        </div>
      </div>
    </header>

    <main>
      <section class="yc-hero yc-wrap">
        <p class="yc-eyebrow">{{ t('yc.hero.eyebrow') }}</p>
        <h1 class="yc-h1">
          {{ t('yc.hero.titleA') }}<br />
          <span class="yc-mark">{{ t('yc.hero.titleB') }}</span>
        </h1>
        <p class="yc-lead">{{ t('yc.hero.lead') }}</p>
        <div class="yc-hero-cta">
          <router-link :to="isAuthenticated ? consolePath : startPath" class="yc-btn yc-btn-lime">
            {{ isAuthenticated ? t('yc.cta.console') : t('yc.cta.start') }}
            <span aria-hidden="true">→</span>
          </router-link>
          <a href="#how" class="yc-btn yc-btn-ghost">{{ t('yc.cta.how') }}</a>
        </div>
        <ul class="yc-hero-points">
          <li v-for="p in heroPoints" :key="p">{{ t(p) }}</li>
        </ul>
        <YcFlowDiagram class="yc-hero-flow" />
      </section>

      <section id="how" class="yc-sec yc-wrap">
        <p class="yc-eyebrow">{{ t('yc.how.eyebrow') }}</p>
        <h2 class="yc-h2">{{ t('yc.how.title') }}</h2>
        <ol class="yc-steps">
          <li v-for="(s, i) in steps" :key="s.title" class="yc-step">
            <span class="yc-step-n">{{ String(i + 1).padStart(2, '0') }}</span>
            <h3>{{ t(s.title) }}</h3>
            <p>{{ t(s.body) }}</p>
          </li>
        </ol>
      </section>

      <section id="service" class="yc-sec yc-wrap">
        <p class="yc-eyebrow">{{ t('yc.service.eyebrow') }}</p>
        <h2 class="yc-h2">{{ t('yc.service.title') }}</h2>
        <div class="yc-promises">
          <article v-for="s in promises" :key="s.title" class="yc-card yc-promise">
            <h3>{{ t(s.title) }}</h3>
            <p>{{ t(s.body) }}</p>
          </article>
        </div>
      </section>

      <section id="faq" class="yc-sec yc-wrap yc-faq">
        <div>
          <p class="yc-eyebrow">{{ t('yc.faq.eyebrow') }}</p>
          <h2 class="yc-h2">{{ t('yc.faq.title') }}</h2>
        </div>
        <div class="yc-faq-list">
          <details v-for="q in faqs" :key="q.q">
            <summary>{{ t(q.q) }}<span aria-hidden="true">+</span></summary>
            <p>{{ t(q.a) }}</p>
          </details>
        </div>
      </section>

      <section class="yc-wrap">
        <div class="yc-final">
          <h2 class="yc-h2">{{ t('yc.final.title') }}</h2>
          <router-link :to="isAuthenticated ? consolePath : startPath" class="yc-btn yc-btn-lime">
            {{ isAuthenticated ? t('yc.cta.console') : t('yc.cta.start') }}
            <span aria-hidden="true">→</span>
          </router-link>
        </div>
      </section>
    </main>

    <footer class="yc-wrap yc-foot">
      <span>© {{ year }} {{ siteName }}</span>
      <a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer">{{ t('yc.nav.docs') }}</a>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore, useAppStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import UpstreamHome from '@/views/HomeView.vue'
import YcFlowDiagram from '@/components/yecai/YcFlowDiagram.vue'
import { sanitizeUrl } from '@/utils/url'
import { FeatureFlags, isFeatureFlagEnabled } from '@/utils/featureFlags'
import '@/styles/yecai.css'

const { t } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()

const settings = computed(() => appStore.cachedPublicSettings)
// Operators who set custom home content or the compact home keep the upstream page.
const useUpstreamHome = computed(
  () => (settings.value?.home_content || '').trim().length > 0 || settings.value?.compact_home_enabled === true
)
const siteName = computed(() => settings.value?.site_name || appStore.siteName || 'Sub2API')
const siteLogo = computed(() =>
  sanitizeUrl(settings.value?.site_logo || appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true })
)
const docUrl = computed(() => sanitizeUrl(settings.value?.doc_url || appStore.docUrl || ''))
const isAuthenticated = computed(() => authStore.isAuthenticated)
const consolePath = computed(() => (authStore.isAdmin ? '/admin/dashboard' : '/dashboard'))
const startPath = computed(() => (settings.value?.registration_enabled ? '/register' : '/login'))
const showModelPlaza = computed(
  () =>
    isFeatureFlagEnabled(FeatureFlags.modelPlaza) &&
    (isAuthenticated.value || settings.value?.model_plaza_require_auth !== true)
)
const year = new Date().getFullYear()

const heroPoints = ['yc.hero.point1', 'yc.hero.point2', 'yc.hero.point3']
const steps = [
  { title: 'yc.how.s1Title', body: 'yc.how.s1Body' },
  { title: 'yc.how.s2Title', body: 'yc.how.s2Body' },
  { title: 'yc.how.s3Title', body: 'yc.how.s3Body' }
]
const promises = [
  { title: 'yc.service.p1Title', body: 'yc.service.p1Body' },
  { title: 'yc.service.p2Title', body: 'yc.service.p2Body' },
  { title: 'yc.service.p3Title', body: 'yc.service.p3Body' },
  { title: 'yc.service.p4Title', body: 'yc.service.p4Body' }
]
const faqs = [
  { q: 'yc.faq.q1', a: 'yc.faq.a1' },
  { q: 'yc.faq.q2', a: 'yc.faq.a2' },
  { q: 'yc.faq.q3', a: 'yc.faq.a3' },
  { q: 'yc.faq.q4', a: 'yc.faq.a4' }
]

onMounted(() => {
  // Without a saved choice the page follows the system colour scheme.
  const saved = localStorage.getItem('theme')
  const dark = saved ? saved === 'dark' : window.matchMedia('(prefers-color-scheme: dark)').matches
  document.documentElement.classList.toggle('dark', dark)
  authStore.checkAuth()
  if (!appStore.publicSettingsLoaded) appStore.fetchPublicSettings()
})
</script>

<style scoped>
.yc-home { min-height: 100vh; background-image: radial-gradient(var(--yc-line) 1px, transparent 1px); background-size: 28px 28px; background-position: -1px -1px; }
.yc-top { position: sticky; top: 0; z-index: 20; background: color-mix(in srgb, var(--yc-bg) 82%, transparent); backdrop-filter: saturate(160%) blur(18px); -webkit-backdrop-filter: saturate(160%) blur(18px); border-bottom: 1px solid var(--yc-line); }
.yc-top-in { height: 64px; display: flex; align-items: center; gap: 28px; }
.yc-brand { display: flex; align-items: center; gap: 10px; font-weight: 750; font-size: 17px; letter-spacing: -0.01em; white-space: nowrap; }
.yc-brand-logo { width: 30px; height: 30px; border-radius: 8px; object-fit: contain; }
.yc-top-nav { display: flex; gap: 26px; margin-right: auto; font-size: 14.5px; color: var(--yc-mut); }
.yc-top-nav a:hover { color: var(--yc-ink); }
.yc-top-actions { display: flex; align-items: center; gap: 10px; margin-left: auto; }
.yc-hero { padding-top: clamp(56px, 9vw, 112px); text-align: center; }
.yc-h1 { margin-top: 20px; font-size: clamp(40px, 8.2vw, 96px); line-height: 1.04; letter-spacing: -0.045em; font-weight: 800; }
.yc-lead { max-width: 620px; margin: 28px auto 0; font-size: clamp(16px, 1.6vw, 19px); line-height: 1.7; color: var(--yc-mut); text-wrap: pretty; }
.yc-hero-cta { display: flex; flex-wrap: wrap; justify-content: center; gap: 12px; margin-top: 36px; }
.yc-hero-points { display: flex; flex-wrap: wrap; justify-content: center; gap: 8px 28px; margin-top: 28px; font-size: 14px; color: var(--yc-mut); }
.yc-hero-points li::before { content: ''; display: inline-block; width: 7px; height: 7px; margin-right: 9px; border-radius: 50%; background: var(--yc-lime); box-shadow: 0 0 0 1px var(--yc-ink); vertical-align: 1px; }
.yc-hero-flow { margin-top: clamp(48px, 7vw, 88px); }
.yc-sec { padding-top: clamp(88px, 12vw, 152px); }
.yc-sec .yc-h2 { margin-top: 14px; max-width: 760px; }
.yc-steps { display: grid; grid-template-columns: repeat(3, 1fr); margin-top: 56px; border-top: 1.5px solid var(--yc-ink); }
.yc-step { padding: 28px 28px 8px 0; }
.yc-step + .yc-step { padding-left: 28px; border-left: 1px solid var(--yc-line); }
.yc-step-n { font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; font-size: 13px; color: var(--yc-dim); }
.yc-step h3 { margin-top: 44px; font-size: 24px; letter-spacing: -0.02em; font-weight: 700; }
.yc-step p, .yc-promise p { margin-top: 10px; color: var(--yc-mut); line-height: 1.7; }
.yc-promises { display: grid; grid-template-columns: repeat(2, 1fr); gap: 16px; margin-top: 48px; }
.yc-promise { padding: 30px 32px 32px; background: color-mix(in srgb, var(--yc-paper) 88%, transparent); }
.yc-promise h3 { font-size: 21px; letter-spacing: -0.015em; font-weight: 700; }
.yc-faq { display: grid; grid-template-columns: 1fr 1.4fr; gap: 48px; }
.yc-faq-list { border-top: 1.5px solid var(--yc-ink); }
.yc-faq-list details { border-bottom: 1px solid var(--yc-line); }
.yc-faq-list summary { display: flex; justify-content: space-between; gap: 16px; padding: 22px 0; font-size: 17px; font-weight: 650; cursor: pointer; list-style: none; }
.yc-faq-list summary::-webkit-details-marker { display: none; }
.yc-faq-list summary span { color: var(--yc-dim); font-weight: 400; transition: transform 0.2s ease; }
.yc-faq-list details[open] summary span { transform: rotate(45deg); }
.yc-faq-list details p { padding: 0 40px 24px 0; color: var(--yc-mut); line-height: 1.75; }
.yc-final { margin-top: clamp(88px, 12vw, 152px); padding: clamp(40px, 7vw, 80px) 24px; border-radius: 28px; background: var(--yc-lime); color: var(--yc-on-lime); text-align: center; }
.yc-final .yc-h2 { margin: 0 auto 32px; max-width: 640px; }
.yc-final .yc-btn-lime { background: var(--yc-on-lime); color: var(--yc-lime); box-shadow: none; }
.yc-final .yc-btn-lime:hover { box-shadow: 0 10px 24px -8px rgb(11 15 12 / 0.6); }
.yc-foot { display: flex; justify-content: space-between; gap: 16px; padding-top: 48px; padding-bottom: 48px; font-size: 13.5px; color: var(--yc-dim); }
@media (max-width: 860px) {
  .yc-top-nav { display: none; }
  .yc-steps, .yc-promises, .yc-faq { grid-template-columns: 1fr; }
  .yc-step, .yc-step + .yc-step { padding: 24px 0 20px; border-left: 0; }
  .yc-step + .yc-step { border-top: 1px solid var(--yc-line); }
  .yc-step h3 { margin-top: 14px; }
  .yc-faq { gap: 32px; }
}
@media (max-width: 480px) {
  .yc-hide-sm { display: none; }
  .yc-top-in { gap: 12px; }
}
@media (prefers-reduced-motion: no-preference) {
  .yc-h1, .yc-lead, .yc-hero-cta, .yc-hero-points, .yc-hero-flow { animation: yc-rise 0.7s cubic-bezier(0.2, 0.7, 0.2, 1) both; }
  .yc-lead { animation-delay: 0.08s; }
  .yc-hero-cta { animation-delay: 0.16s; }
  .yc-hero-points { animation-delay: 0.22s; }
  .yc-hero-flow { animation-delay: 0.3s; }
}
@keyframes yc-rise { from { opacity: 0; transform: translateY(18px); } to { opacity: 1; transform: none; } }
</style>
