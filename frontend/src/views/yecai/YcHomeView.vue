<template>
  <UpstreamHome v-if="useUpstreamHome" />
  <div v-else class="yc yc-home">
    <header class="yc-rail">
      <router-link to="/home" class="yc-brand">
        <img :src="siteLogo || '/logo.svg'" alt="" class="yc-brand-logo" />
        <span>{{ siteName }}</span>
      </router-link>
      <div class="yc-rail-lang"><LocaleSwitcher /></div>
      <nav class="yc-index" :aria-label="t('yc.nav.aria')">
        <a
          v-for="s in sections"
          :key="s.id"
          :href="`#${s.id}`"
          :class="{ 'is-on': s.id === active }"
          :aria-current="s.id === active ? 'location' : undefined"
        >
          <span class="yc-index-n">{{ num(s.id) }}</span>{{ t(s.label) }}
        </a>
      </nav>
      <div class="yc-rail-foot">
        <router-link v-if="isAuthenticated" :to="consolePath" class="yc-btn yc-btn-lime yc-btn-sm">{{ t('yc.cta.console') }}</router-link>
        <template v-else>
          <router-link to="/login" class="yc-btn yc-btn-ghost yc-btn-sm yc-hide-sm">{{ t('yc.cta.login') }}</router-link>
          <router-link :to="startPath" class="yc-btn yc-btn-lime yc-btn-sm">{{ t('yc.cta.start') }}</router-link>
        </template>
      </div>
    </header>

    <main class="yc-main">
      <section class="yc-hero">
        <p class="yc-label">{{ t('yc.hero.eyebrow') }}</p>
        <h1 class="yc-h1">
          {{ t('yc.hero.titleA') }}<br />
          <span class="yc-mark">{{ t('yc.hero.titleB') }}</span>
        </h1>
        <div class="yc-hero-row">
          <p class="yc-lead">{{ t('yc.hero.lead') }}</p>
          <div class="yc-hero-cta">
            <router-link :to="isAuthenticated ? consolePath : startPath" class="yc-cta-block">
              <span>{{ isAuthenticated ? t('yc.cta.console') : t('yc.cta.start') }}</span>
              <span class="yc-cta-arrow" aria-hidden="true">→</span>
            </router-link>
            <a :href="clientUrl || '#pain-setup'" class="yc-cta-alt">{{ t(clientLabel) }} ↓</a>
          </div>
        </div>
        <dl v-if="site" class="yc-live">
          <div class="yc-live-main">
            <dt><i aria-hidden="true"></i>{{ t('yc.live.label') }}</dt>
            <dd class="yc-live-num yc-num">{{ Math.round(liveTotal).toLocaleString() }}</dd>
          </div>
          <div>
            <dt>{{ t('yc.live.today') }}</dt>
            <dd class="yc-num">{{ Math.round(liveToday).toLocaleString() }}</dd>
          </div>
          <div>
            <dt>{{ t('yc.live.requests') }}</dt>
            <dd class="yc-num">{{ site.total_requests.toLocaleString() }}</dd>
          </div>
        </dl>
        <YcFlowDiagram class="yc-hero-flow" />
      </section>

      <section id="pains">
        <article v-for="p in pains" :id="`pain-${p.id}`" :key="p.id" v-reveal class="yc-sec yc-pain yc-reveal">
          <p class="yc-label"><b>{{ num(`pain-${p.id}`) }}</b>{{ t(`yc.pain.${p.id}.tag`) }}</p>
          <h2 class="yc-pain-h">
            <span class="yc-pain-old">{{ t(`yc.pain.${p.id}.titleA`) }}</span><br />
            <span class="yc-mark">{{ t(`yc.pain.${p.id}.titleB`) }}</span>
          </h2>
          <div class="yc-vs">
            <div class="yc-vs-col yc-vs-before">
              <p class="yc-vs-lab">{{ t('yc.pain.before') }}</p>
              <pre v-if="p.id === 'setup'" class="yc-vs-code" aria-hidden="true">{{ CONFIG_SAMPLE }}</pre>
              <p v-if="p.id === 'price'" class="yc-vs-formula">{{ t('yc.pain.price.formula') }}</p>
              <ul>
                <li v-for="k in ITEMS" :key="k">{{ t(`yc.pain.${p.id}.b${k}`) }}</li>
              </ul>
            </div>
            <div class="yc-vs-col yc-vs-after">
              <p class="yc-vs-lab">{{ t('yc.pain.after') }}</p>
              <p v-if="p.id === 'price' && top" class="yc-vs-price">
                <span>{{ top.name }}</span>
                <b class="yc-num">¥{{ price(top.input) }}</b>
                <small>{{ t('yc.price.perMillionIn') }}</small>
              </p>
              <ul>
                <li v-for="k in ITEMS" :key="k">{{ t(`yc.pain.${p.id}.a${k}`) }}</li>
              </ul>
              <template v-if="p.id === 'setup'">
                <div class="yc-client-get">
                  <a v-if="clientUrl" :href="clientUrl" class="yc-btn yc-btn-lime yc-btn-sm">{{ t(clientLabel) }} ↓</a>
                  <span class="yc-client-alt">
                    <a :href="YC_CLIENT_DOWNLOADS.windows">Windows</a> · <a :href="YC_CLIENT_DOWNLOADS.macos">macOS</a>
                  </span>
                </div>
                <p class="yc-client-note">{{ t('yc.client.manual') }}</p>
                <figure class="yc-shot">
                  <figcaption aria-hidden="true"><i></i><i></i><i></i><span>{{ t('yc.client.window') }}</span></figcaption>
                  <img class="yc-shot-light" :src="shotLight" :alt="t('yc.client.shotAlt')" width="1820" height="880" loading="lazy" />
                  <img class="yc-shot-dark" :src="shotDark" :alt="t('yc.client.shotAlt')" width="1826" height="873" loading="lazy" />
                </figure>
              </template>
              <p v-if="p.id === 'help' && contactInfo" class="yc-vs-contact">{{ contactInfo }}</p>
            </div>
          </div>
        </article>
      </section>

      <section v-if="top" id="price" v-reveal class="yc-flood yc-price yc-reveal">
        <p class="yc-label"><b>{{ num('price') }}</b>{{ t('yc.price.eyebrow') }}</p>
        <h2 class="yc-flood-h">{{ t('yc.price.title') }}</h2>
        <div class="yc-price-top" aria-live="polite">
          <p :key="top.name" class="yc-price-now yc-num">¥{{ price(top.input) }}</p>
          <div class="yc-price-meta">
            <p class="yc-price-name">{{ top.name }}</p>
            <p class="yc-price-unit">{{ t('yc.price.perMillionIn') }}</p>
            <p v-if="top.officialInput && top.share" class="yc-price-was">{{ t('yc.price.official') }} ¥{{ price(top.officialInput) }}</p>
            <p v-if="top.share"><span class="yc-chip">{{ off(top.share) }}</span></p>
            <p class="yc-price-out">{{ t('yc.price.output', { price: price(top.output) }) }}</p>
          </div>
        </div>
        <div v-if="choices.length > 1" class="yc-price-list" role="group" :aria-label="t('yc.price.pick')">
          <p class="yc-price-pick">{{ t('yc.price.pick') }}</p>
          <div class="yc-price-grid">
            <button
              v-for="m in choices"
              :key="m.name"
              type="button"
              class="yc-price-row"
              :class="{ 'is-on': m.name === top.name }"
              :aria-pressed="m.name === top.name"
              @click="picked = m.name"
            >
              <span class="yc-price-row-name">{{ m.name }}</span>
              <b class="yc-num">¥{{ price(m.input) }}</b>
              <span v-if="m.share" class="yc-price-row-off">
                <s v-if="m.officialInput">¥{{ price(m.officialInput) }}</s> {{ off(m.share) }}
              </span>
            </button>
          </div>
        </div>
        <p class="yc-price-note">{{ t('yc.price.note', { rate: cnyPerUsd }) }}</p>
        <router-link v-if="showModelPlaza" to="/model-plaza" class="yc-price-all">{{ t('yc.price.all') }} →</router-link>
      </section>

      <section id="faq" class="yc-sec yc-faq">
        <div>
          <p class="yc-label"><b>{{ num('faq') }}</b>{{ t('yc.faq.eyebrow') }}</p>
          <h2 class="yc-pain-h">{{ t('yc.faq.title') }}</h2>
        </div>
        <div class="yc-faq-list">
          <details v-for="k in FAQS" :key="k">
            <summary>{{ t(`yc.faq.q${k}`) }}<span aria-hidden="true">+</span></summary>
            <p>{{ t(`yc.faq.a${k}`) }}</p>
          </details>
        </div>
      </section>

      <section class="yc-final">
        <h2>{{ t('yc.final.title') }}</h2>
        <router-link :to="isAuthenticated ? consolePath : startPath" class="yc-cta-block">
          <span>{{ isAuthenticated ? t('yc.cta.console') : t('yc.cta.start') }}</span>
          <span class="yc-cta-arrow" aria-hidden="true">→</span>
        </router-link>
      </section>

      <footer class="yc-foot">
        <span>© {{ year }} {{ siteName }}</span>
        <a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer">{{ t('yc.nav.docs') }}</a>
      </footer>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useIntervalFn, useTransition } from '@vueuse/core'
import { useI18n } from 'vue-i18n'
import { useAuthStore, useAppStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import UpstreamHome from '@/views/HomeView.vue'
import YcFlowDiagram from '@/components/yecai/YcFlowDiagram.vue'
import { useYcPrices } from '@/composables/useYcPrices'
import { getSiteStats, type SiteStats } from '@/api/siteStats'
import { sanitizeUrl } from '@/utils/url'
import { YC_CLIENT_DOWNLOADS, detectYcClientPlatform } from '@/utils/ycClient'
import { FeatureFlags, isFeatureFlagEnabled } from '@/utils/featureFlags'
import shotLight from '@/assets/yecai/yecai-client-apps-light-showcase.webp'
import shotDark from '@/assets/yecai/yecai-client-apps-dark-showcase.webp'
import '@/styles/yecai.css'

const { t } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()
const { prices, load: loadPrices, cnyPerUsd } = useYcPrices()

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
const contactInfo = computed(() => settings.value?.contact_info || '')
const isAuthenticated = computed(() => authStore.isAuthenticated)
const consolePath = computed(() => (authStore.isAdmin ? '/admin/dashboard' : '/dashboard'))
const startPath = computed(() => (settings.value?.registration_enabled ? '/register' : '/login'))
const showModelPlaza = computed(
  () =>
    isFeatureFlagEnabled(FeatureFlags.modelPlaza) &&
    (isAuthenticated.value || settings.value?.model_plaza_require_auth !== true)
)
const year = new Date().getFullYear()

const pains = [{ id: 'setup' }, { id: 'price' }, { id: 'bill' }, { id: 'help' }]
const ITEMS = [1, 2, 3]
const FAQS = [1, 2, 3, 4, 5]
const CONFIG_SAMPLE = `~/.claude/settings.json
{
  "env": {
    "ANTHROPIC_BASE_URL": "https://???",
    "ANTHROPIC_AUTH_TOKEN": "sk-???"
  }
}`

const clientPlatform = detectYcClientPlatform()
const clientUrl = clientPlatform ? YC_CLIENT_DOWNLOADS[clientPlatform] : ''
const clientLabel =
  clientPlatform === 'windows' ? 'yc.client.windows' : clientPlatform === 'macos' ? 'yc.client.macos' : 'yc.client.download'

// Site-wide token count, refreshed while the tab is visible. Hidden until there is something to show.
const site = ref<SiteStats | null>(null)
const tick = { duration: 1600, transition: [0.2, 0.7, 0.2, 1] as [number, number, number, number] }
const liveTotal = useTransition(() => site.value?.total_tokens ?? 0, tick)
const liveToday = useTransition(() => site.value?.today_tokens ?? 0, tick)
async function loadSiteStats() {
  if (document.hidden) return
  try {
    const stats = await getSiteStats()
    site.value = stats.total_tokens > 0 ? stats : null
  } catch {
    site.value = null
  }
}
useIntervalFn(loadSiteStats, 15000)

const picked = ref('')
const choices = computed(() => prices.value.slice(0, 6))
const top = computed(() => choices.value.find((m) => m.name === picked.value) ?? choices.value[0] ?? null)

// Left index: one entry per section, numbered in page order; the one in view is highlighted.
const sections = computed(() => [
  ...pains.map((p) => ({ id: `pain-${p.id}`, label: `yc.pain.${p.id}.tag` })),
  ...(top.value ? [{ id: 'price', label: 'yc.price.eyebrow' }] : []),
  { id: 'faq', label: 'yc.faq.eyebrow' }
])
function num(id: string): string {
  return String(sections.value.findIndex((s) => s.id === id) + 1).padStart(2, '0')
}
const active = ref('')
let spy: IntersectionObserver | null = null
function watchSections() {
  spy?.disconnect()
  if (!('IntersectionObserver' in window)) return
  spy = new IntersectionObserver(
    (entries) => {
      for (const entry of entries) if (entry.isIntersecting) active.value = entry.target.id
    },
    { rootMargin: '-35% 0px -60% 0px' }
  )
  for (const s of sections.value) {
    const el = document.getElementById(s.id)
    if (el) spy.observe(el)
  }
}
watch(
  () => sections.value.length,
  () => nextTick(watchSections)
)
onBeforeUnmount(() => spy?.disconnect())

function price(value: number): string {
  if (value >= 100) return value.toFixed(0)
  return value >= 10 ? value.toFixed(1) : value.toFixed(2)
}
function off(share: number): string {
  return t('yc.price.off', { zhe: (share * 10).toFixed(1), pct: Math.round((1 - share) * 100) })
}

const vReveal = {
  mounted(el: HTMLElement) {
    if (!('IntersectionObserver' in window)) {
      el.classList.add('is-in')
      return
    }
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (!entry.isIntersecting) return
        el.classList.add('is-in')
        observer.disconnect()
      },
      { threshold: 0.15 }
    )
    observer.observe(el)
  }
}

onMounted(async () => {
  // Without a saved choice the page follows the system colour scheme.
  const saved = localStorage.getItem('theme')
  const dark = saved ? saved === 'dark' : window.matchMedia('(prefers-color-scheme: dark)').matches
  document.documentElement.classList.toggle('dark', dark)
  authStore.checkAuth()
  loadSiteStats()
  watchSections()
  if (!appStore.publicSettingsLoaded) await appStore.fetchPublicSettings()
  if (showModelPlaza.value) loadPrices()
})
</script>

<style scoped>
/* Layout: a fixed index rail on the left, hairline-ruled sections on the right, and one screen flooded with lime. */
.yc-home { --pad: clamp(20px, 5vw, 72px); --mono: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; min-height: 100vh; display: grid; grid-template-columns: clamp(216px, 18vw, 288px) minmax(0, 1fr); overflow-x: clip; }
.yc-home .yc-btn { border-radius: 2px; }
.yc-rail { position: sticky; top: 0; z-index: 20; height: 100vh; height: 100dvh; display: flex; flex-direction: column; padding: 28px 28px 24px; border-right: 1px solid var(--yc-line); overflow-y: auto; }
.yc-brand { display: flex; align-items: center; gap: 10px; font-weight: 750; font-size: 17px; letter-spacing: -0.01em; white-space: nowrap; }
.yc-brand-logo { width: 30px; height: 30px; border-radius: 6px; object-fit: contain; }
.yc-rail-lang { margin: 14px 0 0 -8px; align-self: flex-start; }
.yc-index { margin-top: auto; padding-top: 32px; display: grid; }
.yc-index a { display: flex; align-items: center; gap: 12px; padding: 9px 0; border-top: 1px solid var(--yc-line); font-size: 13.5px; color: var(--yc-dim); transition: color 0.15s ease; }
.yc-index a:last-child { border-bottom: 1px solid var(--yc-line); }
.yc-index a:hover { color: var(--yc-ink); }
.yc-index a::after { content: ''; margin-left: auto; width: 8px; height: 8px; }
.yc-index a.is-on { color: var(--yc-ink); font-weight: 650; }
.yc-index a.is-on::after { background: var(--yc-lime); box-shadow: 0 0 0 1px var(--yc-on-lime); }
.yc-index-n { width: 20px; font-family: var(--mono); font-size: 11.5px; font-weight: 400; }
.yc-rail-foot { margin-top: 24px; display: grid; gap: 8px; }

.yc-main { min-width: 0; }
.yc-label { display: flex; align-items: baseline; gap: 14px; font-family: var(--mono); font-size: 12.5px; letter-spacing: 0.08em; text-transform: uppercase; color: var(--yc-mut); }
.yc-label b { font-weight: 600; color: var(--yc-ink); }

.yc-hero { padding: clamp(48px, 7vw, 104px) var(--pad) clamp(56px, 7vw, 96px); }
.yc-h1 { margin-top: 28px; font-size: clamp(44px, 8.6vw, 128px); line-height: 1.02; letter-spacing: -0.055em; font-weight: 800; }
.yc-hero-row { margin-top: clamp(40px, 5vw, 72px); padding-top: 28px; border-top: 1.5px solid var(--yc-ink); display: grid; grid-template-columns: minmax(0, 1fr) auto; gap: 32px 56px; align-items: start; }
.yc-lead { max-width: 600px; font-size: clamp(18px, 2vw, 26px); line-height: 1.55; font-weight: 500; letter-spacing: -0.01em; text-wrap: pretty; }
.yc-hero-cta { display: grid; gap: 14px; }
.yc-cta-block { display: flex; flex-direction: column; justify-content: space-between; gap: 28px; width: clamp(220px, 22vw, 300px); min-height: 150px; padding: 20px 22px; background: var(--yc-lime); color: var(--yc-on-lime); font-size: 22px; font-weight: 750; letter-spacing: -0.02em; }
.yc-cta-arrow { align-self: flex-end; font-size: 34px; line-height: 1; transition: transform 0.25s cubic-bezier(0.2, 0.7, 0.2, 1); }
.yc-cta-block:hover .yc-cta-arrow { transform: translateX(8px); }
.yc-cta-alt { font-size: 13.5px; color: var(--yc-mut); text-decoration: underline; text-underline-offset: 4px; }
.yc-cta-alt:hover { color: var(--yc-ink); }

.yc-live { margin-top: clamp(48px, 6vw, 88px); display: grid; grid-template-columns: 1fr 1fr; border-top: 1px solid var(--yc-line); border-bottom: 1px solid var(--yc-line); }
.yc-live > div { padding: 20px 0 22px; }
.yc-live > div + div { border-top: 1px solid var(--yc-line); }
.yc-live > div:last-child { padding-left: 24px; border-left: 1px solid var(--yc-line); }
.yc-live-main { grid-column: 1 / -1; }
.yc-live dt { display: flex; align-items: center; gap: 9px; font-family: var(--mono); font-size: 12.5px; letter-spacing: 0.04em; color: var(--yc-mut); }
.yc-live dt i { width: 8px; height: 8px; background: var(--yc-lime); box-shadow: 0 0 0 1px var(--yc-on-lime); animation: yc-blink 1.6s ease-in-out infinite; }
.yc-live dd { margin-top: 8px; font-size: clamp(20px, 2.4vw, 30px); line-height: 1.1; overflow-wrap: anywhere; }
.yc-live .yc-live-num { font-size: clamp(36px, 7vw, 96px); line-height: 1; }
@keyframes yc-blink { 50% { opacity: 0.35; } }
.yc-hero-flow { margin-top: clamp(48px, 6vw, 80px); }

.yc-sec { padding: clamp(72px, 9vw, 128px) var(--pad); border-top: 1px solid var(--yc-line); }
.yc-pain-h { margin-top: 22px; max-width: 980px; font-size: clamp(28px, 4.4vw, 60px); line-height: 1.16; letter-spacing: -0.035em; font-weight: 780; text-wrap: balance; }
.yc-pain-old { color: var(--yc-dim); font-weight: 500; }
.yc-vs { margin-top: clamp(36px, 5vw, 64px); display: grid; grid-template-columns: 1fr 1.25fr; border-top: 1.5px solid var(--yc-ink); }
.yc-vs-col { padding: 26px 0 0; }
.yc-vs-before { padding-right: 36px; color: var(--yc-mut); }
.yc-vs-after { padding-left: 36px; border-left: 1px solid var(--yc-line); }
.yc-vs-lab { display: inline-block; font-family: var(--mono); font-size: 12.5px; letter-spacing: 0.08em; }
.yc-vs-before .yc-vs-lab { color: var(--yc-dim); }
.yc-vs-after .yc-vs-lab { padding: 2px 8px; background: var(--yc-lime); color: var(--yc-on-lime); font-weight: 600; }
.yc-vs-col ul { margin-top: 20px; font-size: 17px; line-height: 1.55; }
.yc-vs-col li { position: relative; padding: 12px 0 12px 28px; border-top: 1px solid var(--yc-line); }
.yc-vs-before li::before { content: '×'; position: absolute; left: 2px; color: var(--yc-dim); }
.yc-vs-before li { text-decoration: line-through; text-decoration-color: var(--yc-line); }
.yc-vs-after li { font-weight: 600; }
.yc-vs-after li::before { content: ''; position: absolute; left: 2px; top: 1.25em; width: 11px; height: 6px; border-left: 2.5px solid var(--yc-ink); border-bottom: 2.5px solid var(--yc-ink); transform: rotate(-45deg); }
.yc-vs-code, .yc-vs-formula { margin-top: 18px; padding: 14px 16px; background: var(--yc-fill); border-left: 2px solid var(--yc-line); color: var(--yc-dim); }
.yc-vs-code { font-family: var(--mono); font-size: 12.5px; line-height: 1.7; overflow-x: auto; }
.yc-vs-formula { font-size: 15px; }
.yc-vs-price { margin-top: 18px; display: flex; flex-wrap: wrap; align-items: baseline; gap: 6px 12px; }
.yc-vs-price span { width: 100%; font-family: var(--mono); font-size: 13.5px; color: var(--yc-mut); }
.yc-vs-price b { font-size: clamp(44px, 6vw, 72px); line-height: 1; }
.yc-vs-price small { font-size: 14px; color: var(--yc-mut); }
.yc-client-get { margin-top: 22px; display: flex; flex-wrap: wrap; align-items: center; gap: 12px 16px; }
.yc-client-alt { font-size: 13.5px; color: var(--yc-mut); }
.yc-client-alt a { text-decoration: underline; text-underline-offset: 3px; }
.yc-client-alt a:hover { color: var(--yc-ink); }
.yc-client-note { margin-top: 12px; font-size: 13.5px; color: var(--yc-mut); }
.yc-shot { margin-top: 24px; border: 1px solid var(--yc-line); background: var(--yc-fill); }
.yc-shot figcaption { display: flex; align-items: center; gap: 6px; height: 32px; padding: 0 12px; border-bottom: 1px solid var(--yc-line); font-family: var(--mono); font-size: 11.5px; color: var(--yc-mut); }
.yc-shot figcaption i { width: 8px; height: 8px; border-radius: 50%; background: var(--yc-line); }
.yc-shot figcaption span { margin-left: 8px; }
.yc-shot img { display: block; width: 100%; height: auto; }
.yc-shot-dark { display: none !important; }
:global(.dark .yc .yc-shot-light) { display: none; }
:global(.dark .yc .yc-shot-dark) { display: block !important; }
.yc-vs-contact { margin-top: 18px; padding-top: 16px; border-top: 1px solid var(--yc-line); font-size: 15px; color: var(--yc-mut); white-space: pre-line; }

/* The price screen is lime edge to edge in both themes. */
.yc-flood { padding: clamp(72px, 9vw, 128px) var(--pad); background: var(--yc-lime); color: var(--yc-on-lime); }
.yc-flood :focus-visible { outline-color: var(--yc-on-lime); }
.yc-flood .yc-label, .yc-flood .yc-label b { color: var(--yc-on-lime); }
.yc-flood-h { margin-top: 22px; font-size: clamp(28px, 4.4vw, 60px); line-height: 1.1; letter-spacing: -0.035em; font-weight: 800; }
.yc-price-top { margin-top: clamp(36px, 5vw, 64px); padding-top: 28px; border-top: 2px solid var(--yc-on-lime); display: flex; flex-wrap: wrap; align-items: flex-end; gap: 20px 48px; }
.yc-price-now { font-size: clamp(88px, 17vw, 260px); line-height: 0.86; letter-spacing: -0.06em; font-weight: 900; }
.yc-price-meta { display: grid; gap: 8px; padding-bottom: 6px; font-size: 15px; }
.yc-price-name { font-family: var(--mono); font-size: 17px; font-weight: 700; }
.yc-price-was { text-decoration: line-through; opacity: 0.6; }
.yc-chip { display: inline-block; padding: 4px 10px; background: var(--yc-on-lime); color: var(--yc-lime); font-size: 13px; font-weight: 700; white-space: nowrap; }
.yc-price-out { opacity: 0.7; }
.yc-price-list { margin-top: clamp(40px, 6vw, 72px); }
.yc-price-pick { font-family: var(--mono); font-size: 12.5px; letter-spacing: 0.04em; }
.yc-price-grid { margin-top: 12px; display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); border-top: 1.5px solid var(--yc-on-lime); border-left: 1.5px solid var(--yc-on-lime); }
.yc-price-row { display: grid; gap: 10px; padding: 18px 20px 20px; border-right: 1.5px solid var(--yc-on-lime); border-bottom: 1.5px solid var(--yc-on-lime); text-align: left; transition: background-color 0.15s ease, color 0.15s ease; }
.yc-price-row:hover { background: rgb(11 15 12 / 0.08); }
.yc-price-row.is-on { background: var(--yc-on-lime); color: var(--yc-lime); }
.yc-price-row-name { min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-family: var(--mono); font-size: 13.5px; }
.yc-price-row b { font-size: clamp(24px, 2.8vw, 34px); line-height: 1; }
.yc-price-row-off { font-size: 13px; opacity: 0.75; }
.yc-price-note { margin-top: 24px; max-width: 640px; font-size: 13px; line-height: 1.7; opacity: 0.7; }
.yc-price-all { display: inline-block; margin-top: 18px; font-weight: 700; border-bottom: 2px solid var(--yc-on-lime); }

.yc-faq { display: grid; grid-template-columns: 1fr 1.6fr; gap: 48px; }
.yc-faq-list { border-top: 1.5px solid var(--yc-ink); }
.yc-faq-list details { border-bottom: 1px solid var(--yc-line); }
.yc-faq-list summary { display: flex; justify-content: space-between; gap: 16px; padding: 22px 0; font-size: 17px; font-weight: 650; cursor: pointer; list-style: none; }
.yc-faq-list summary::-webkit-details-marker { display: none; }
.yc-faq-list summary span { color: var(--yc-dim); font-weight: 400; transition: transform 0.2s ease; }
.yc-faq-list details[open] summary span { transform: rotate(45deg); }
.yc-faq-list details p { padding: 0 40px 24px 0; color: var(--yc-mut); line-height: 1.75; }

.yc-final { padding: clamp(72px, 10vw, 144px) var(--pad); background: #0b0f0c; color: #c8ff00; border-top: 1px solid var(--yc-line); display: flex; flex-wrap: wrap; align-items: flex-end; justify-content: space-between; gap: 40px; }
.yc-final :focus-visible { outline-color: #c8ff00; }
.yc-final h2 { max-width: 11em; font-size: clamp(40px, 7.6vw, 112px); line-height: 1.02; letter-spacing: -0.055em; font-weight: 900; }
.yc-foot { display: flex; justify-content: space-between; gap: 16px; padding: 28px var(--pad); border-top: 1px solid var(--yc-line); font-family: var(--mono); font-size: 12.5px; color: var(--yc-dim); }

@media (max-width: 900px) {
  .yc-home { display: block; }
  .yc-rail { height: 60px; flex-direction: row; align-items: center; gap: 12px; padding: 0 20px; border-right: 0; border-bottom: 1px solid var(--yc-line); background: color-mix(in srgb, var(--yc-bg) 82%, transparent); backdrop-filter: saturate(180%) blur(20px); -webkit-backdrop-filter: saturate(180%) blur(20px); overflow: visible; }
  .yc-index { display: none; }
  .yc-pain, .yc-flood, .yc-faq { scroll-margin-top: 60px; }
  .yc-rail-lang { margin: 0 0 0 auto; align-self: center; }
  .yc-rail-foot { margin-top: 0; display: flex; }
  .yc-hero-row, .yc-vs, .yc-faq { grid-template-columns: 1fr; }
  .yc-cta-block { width: 100%; min-height: 0; flex-direction: row; align-items: center; }
  .yc-cta-arrow { align-self: center; }
  .yc-vs-before { padding-right: 0; padding-bottom: 28px; }
  .yc-vs-after { padding-left: 0; border-left: 0; border-top: 1px solid var(--yc-line); }
  .yc-faq { gap: 32px; }
  .yc-price-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .yc-price-row { padding: 14px 14px 16px; }
}
@media (max-width: 480px) {
  .yc-hide-sm { display: none; }
}
@media (prefers-reduced-motion: no-preference) {
  .yc-h1, .yc-hero-row, .yc-live, .yc-hero-flow { animation: yc-rise 0.8s cubic-bezier(0.2, 0.7, 0.2, 1) both; }
  .yc-hero-row { animation-delay: 0.1s; }
  .yc-live { animation-delay: 0.2s; }
  .yc-hero-flow { animation-delay: 0.28s; }
  .yc-price-now { animation: yc-pop 0.45s cubic-bezier(0.2, 0.9, 0.2, 1.1) both; transform-origin: left bottom; }
}
@keyframes yc-pop { from { opacity: 0; transform: scale(0.92); } to { opacity: 1; transform: none; } }
@keyframes yc-rise { from { opacity: 0; transform: translateY(20px); } to { opacity: 1; transform: none; } }
</style>
