<template>
  <UpstreamHome v-if="useUpstreamHome" />
  <div v-else class="yc yc-home" :class="{ 'is-pin': pin }">
    <header class="yc-top">
      <i class="yc-progress" aria-hidden="true" :style="{ transform: `scaleX(${progress})` }"></i>
      <div class="yc-wrap yc-top-in">
        <router-link to="/home" class="yc-brand">
          <YcBrand />
        </router-link>
        <nav class="yc-top-nav" :aria-label="t('yc.nav.aria')">
          <a href="#pains">{{ t('yc.nav.pains') }}</a>
          <a v-if="top" href="#price">{{ t('yc.nav.prices') }}</a>
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
      <section class="yc-hero yc-wrap" @pointermove="lightDots" @pointerleave="dotsLive = false">
        <div ref="dotsEl" class="yc-hero-dots" :class="{ 'is-live': dotsLive }" aria-hidden="true"><i></i><i></i></div>
        <p class="yc-eyebrow">{{ t('yc.hero.eyebrow') }}</p>
        <h1 v-reveal class="yc-h1">
          <span class="yc-h1-line"><span>{{ t('yc.hero.titleA') }}</span></span>
          <span class="yc-h1-line"><span><span class="yc-mark">{{ t('yc.hero.titleB') }}</span></span></span>
        </h1>
        <p class="yc-lead">{{ t('yc.hero.lead') }}</p>
        <div class="yc-hero-cta">
          <router-link :to="isAuthenticated ? consolePath : startPath" class="yc-btn yc-btn-lime">
            {{ isAuthenticated ? t('yc.cta.console') : t('yc.cta.start') }}
            <span aria-hidden="true">→</span>
          </router-link>
          <a :href="clientUrl || '#pains'" class="yc-btn yc-btn-ghost">{{ t(clientLabel) }}</a>
        </div>
        <div v-if="site" class="yc-live" aria-live="off">
          <p class="yc-live-lab"><i aria-hidden="true"></i>{{ t('yc.live.label') }}</p>
          <p class="yc-live-num yc-num"><YcOdometer :value="site.total_tokens" /></p>
          <p class="yc-live-sub">{{ t('yc.live.today', { tokens: Math.round(liveToday).toLocaleString(), requests: site.total_requests.toLocaleString() }) }}</p>
        </div>
        <YcFlowDiagram class="yc-hero-flow" />
      </section>

      <section id="pains">
        <article v-for="(p, i) in pains" :key="p.id" v-reveal class="yc-pain yc-wrap yc-reveal">
          <div class="yc-pain-stage">
          <header class="yc-pain-head">
            <span class="yc-pain-n yc-num" aria-hidden="true">{{ String(i + 1).padStart(2, '0') }}</span>
            <p class="yc-eyebrow">{{ t(`yc.pain.${p.id}.tag`) }}</p>
            <h2 class="yc-h2">
              <span class="yc-pain-old">{{ t(`yc.pain.${p.id}.titleA`) }}</span><br />
              <span class="yc-mark">{{ t(`yc.pain.${p.id}.titleB`) }}</span>
            </h2>
          </header>
          <div class="yc-vs">
            <div v-reveal="0.3" class="yc-vs-col yc-vs-before">
              <p class="yc-vs-lab">{{ t('yc.pain.before') }}</p>
              <pre v-if="p.id === 'setup'" class="yc-vs-code" aria-hidden="true">{{ CONFIG_SAMPLE }}</pre>
              <p v-if="p.id === 'price'" class="yc-vs-formula">{{ t('yc.pain.price.formula') }}</p>
              <ul>
                <li v-for="k in ITEMS" :key="k" :style="`--i:${k}`"><span>{{ t(`yc.pain.${p.id}.b${k}`) }}</span></li>
              </ul>
            </div>
            <div v-reveal="0.3" v-tilt class="yc-vs-col yc-vs-after yc-surface">
              <p class="yc-vs-lab">{{ t('yc.pain.after') }}</p>
              <p v-if="p.id === 'price' && top" class="yc-vs-price">
                <span>{{ top.name }}</span>
                <b class="yc-num">¥{{ price(top.input) }}</b>
                <small>{{ t('yc.price.perMillionIn') }}</small>
              </p>
              <ul>
                <li v-for="k in ITEMS" :key="k" :style="`--i:${k}`">{{ t(`yc.pain.${p.id}.a${k}`) }}</li>
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
          </div>
        </article>
      </section>

      <section v-if="top" id="price" v-reveal class="yc-sec yc-wrap yc-price yc-reveal">
        <p class="yc-eyebrow">{{ t('yc.price.eyebrow') }}</p>
        <h2 class="yc-h2">{{ t('yc.price.title') }}</h2>
        <div :key="top.name" class="yc-price-top" :class="{ 'is-quick': picked }" aria-live="polite">
          <p class="yc-price-name">{{ top.name }}</p>
          <p v-if="top.officialInput && top.share" class="yc-price-was"><span>{{ t('yc.price.official') }} ¥{{ price(top.officialInput) }}</span></p>
          <p class="yc-price-now yc-num"><span class="yc-mark">¥{{ price(top.input) }}</span></p>
          <p class="yc-price-unit">
            {{ t('yc.price.perMillionIn') }}
            <span v-if="top.share" class="yc-chip">{{ off(top.share) }}</span>
          </p>
          <p class="yc-price-out">{{ t('yc.price.output', { price: price(top.output) }) }}</p>
        </div>
        <div v-if="choices.length > 1" class="yc-price-list yc-surface" role="group" :aria-label="t('yc.price.pick')">
          <p class="yc-price-pick">{{ t('yc.price.pick') }}</p>
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
            <s v-if="m.officialInput && m.share">¥{{ price(m.officialInput) }}</s><s v-else></s>
            <b class="yc-num">¥{{ price(m.input) }}</b>
            <span v-if="m.share" class="yc-chip">{{ off(m.share) }}</span><span v-else></span>
          </button>
        </div>
        <p class="yc-price-note">{{ t('yc.price.note', { rate: cnyPerUsd }) }}</p>
        <router-link v-if="showModelPlaza" to="/model-plaza" class="yc-price-all">{{ t('yc.price.all') }} →</router-link>
      </section>

      <section id="faq" class="yc-sec yc-wrap yc-faq">
        <div>
          <p class="yc-eyebrow">{{ t('yc.faq.eyebrow') }}</p>
          <h2 class="yc-h2">{{ t('yc.faq.title') }}</h2>
        </div>
        <div class="yc-faq-list">
          <details v-for="k in FAQS" :key="k">
            <summary>{{ t(`yc.faq.q${k}`) }}<span aria-hidden="true">+</span></summary>
            <p>{{ t(`yc.faq.a${k}`) }}</p>
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
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useIntervalFn, useTransition, useWindowScroll } from '@vueuse/core'
import { useI18n } from 'vue-i18n'
import { useAuthStore, useAppStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import UpstreamHome from '@/views/HomeView.vue'
import YcBrand from '@/components/yecai/YcBrand.vue'
import YcFlowDiagram from '@/components/yecai/YcFlowDiagram.vue'
import YcOdometer from '@/components/yecai/YcOdometer.vue'
import { useYcPrices } from '@/composables/useYcPrices'
import { getSiteStats, type SiteStats } from '@/api/siteStats'
import { sanitizeUrl } from '@/utils/url'
import { ycSiteName } from '@/utils/ycBrand'
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
const siteName = computed(() => ycSiteName(settings.value?.site_name || appStore.siteName))
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

// Reading progress under the top bar.
const { y: scrollY } = useWindowScroll()
const progress = computed(() => {
  const range = document.documentElement.scrollHeight - window.innerHeight
  return range > 0 ? Math.min(1, scrollY.value / range) : 0
})

// The hero dot grid lights up around the mouse.
const dotsEl = ref<HTMLElement | null>(null)
const dotsLive = ref(false)
function lightDots(event: PointerEvent) {
  const el = dotsEl.value
  if (!el || event.pointerType !== 'mouse') return
  const box = el.getBoundingClientRect()
  el.style.setProperty('--mx', `${event.clientX - box.left}px`)
  el.style.setProperty('--my', `${event.clientY - box.top}px`)
  dotsLive.value = true
}

// Wide screens pin each pain screen and turn the page with the scroll: --p is the progress through
// the screen, --a the eased turn from the old card to the new one.
const pin = ref(false)
let pinQuery: MediaQueryList | null = null
let turnFrame = 0
function turnPages() {
  turnFrame = 0
  if (!pin.value) return
  const stage = window.innerHeight - 64
  for (const el of document.querySelectorAll<HTMLElement>('.yc-pain')) {
    const box = el.getBoundingClientRect()
    const p = Math.min(1, Math.max(0, (64 - box.top) / (box.height - stage)))
    const t = Math.min(1, Math.max(0, (p - 0.1) / 0.55))
    el.style.setProperty('--p', p.toFixed(4))
    el.style.setProperty('--a', (t * t * (3 - 2 * t)).toFixed(4))
    // --b overshoots past 1 and settles, so the new card lands with a bounce.
    el.style.setProperty('--b', (1 + 2.6 * Math.pow(t - 1, 3) + 1.6 * Math.pow(t - 1, 2)).toFixed(4))
  }
}
function queueTurn() {
  if (!turnFrame) turnFrame = requestAnimationFrame(turnPages)
}
function syncPin() {
  pin.value = pinQuery?.matches === true
  queueTurn()
}
onBeforeUnmount(() => {
  pinQuery?.removeEventListener('change', syncPin)
  window.removeEventListener('scroll', queueTurn)
  window.removeEventListener('resize', queueTurn)
  cancelAnimationFrame(turnFrame)
})

const picked = ref('')
const choices = computed(() => prices.value.slice(0, 6))
const top = computed(() => choices.value.find((m) => m.name === picked.value) ?? choices.value[0] ?? null)

function price(value: number): string {
  if (value >= 100) return value.toFixed(0)
  return value >= 10 ? value.toFixed(1) : value.toFixed(2)
}
function off(share: number): string {
  return t('yc.price.off', { zhe: (share * 10).toFixed(1), pct: Math.round((1 - share) * 100) })
}

const vReveal = {
  mounted(el: HTMLElement, binding: { value?: number }) {
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
      { threshold: binding.value ?? 0.15 }
    )
    observer.observe(el)
  }
}

// Cards lean toward the mouse and carry a soft light where it points.
const vTilt = {
  mounted(el: HTMLElement) {
    if (!window.matchMedia('(hover: hover) and (prefers-reduced-motion: no-preference)').matches) return
    el.addEventListener('pointermove', (event) => {
      const box = el.getBoundingClientRect()
      const x = (event.clientX - box.left) / box.width
      const y = (event.clientY - box.top) / box.height
      el.style.setProperty('--ry', `${(x - 0.5) * 5}deg`)
      el.style.setProperty('--rx', `${(0.5 - y) * 5}deg`)
      el.style.setProperty('--sx', `${x * 100}%`)
      el.style.setProperty('--sy', `${y * 100}%`)
    })
    el.addEventListener('pointerleave', () => {
      el.style.setProperty('--ry', '0deg')
      el.style.setProperty('--rx', '0deg')
    })
  }
}

onMounted(async () => {
  // Without a saved choice the page follows the system colour scheme.
  const saved = localStorage.getItem('theme')
  const dark = saved ? saved === 'dark' : window.matchMedia('(prefers-color-scheme: dark)').matches
  document.documentElement.classList.toggle('dark', dark)
  authStore.checkAuth()
  loadSiteStats()
  pinQuery = window.matchMedia('(min-width: 861px) and (min-height: 700px) and (prefers-reduced-motion: no-preference)')
  pinQuery.addEventListener('change', syncPin)
  window.addEventListener('scroll', queueTurn, { passive: true })
  window.addEventListener('resize', queueTurn, { passive: true })
  syncPin()
  if (!appStore.publicSettingsLoaded) await appStore.fetchPublicSettings()
  if (showModelPlaza.value) loadPrices()
})
</script>

<style scoped>
.yc-home { min-height: 100vh; overflow-x: clip; }
.yc-top { position: sticky; top: 0; z-index: 20; background: color-mix(in srgb, var(--yc-bg) 78%, transparent); backdrop-filter: saturate(180%) blur(20px); -webkit-backdrop-filter: saturate(180%) blur(20px); border-bottom: 1px solid var(--yc-line); }
.yc-progress { position: absolute; left: 0; right: 0; bottom: -2px; height: 2px; background: var(--yc-lime); transform-origin: left; }
.yc-top-in { height: 64px; display: flex; align-items: center; gap: 28px; }
.yc-brand { display: flex; align-items: center; gap: 10px; font-weight: 750; font-size: 17px; letter-spacing: -0.01em; white-space: nowrap; }
.yc-top-nav { display: flex; gap: 26px; margin-right: auto; font-size: 14.5px; color: var(--yc-mut); }
.yc-top-nav a:hover { color: var(--yc-ink); }
.yc-top-actions { display: flex; align-items: center; gap: 10px; margin-left: auto; }

.yc-hero { position: relative; padding-top: clamp(64px, 10vw, 128px); text-align: center; }
.yc-hero::before { content: ''; position: absolute; z-index: 0; left: 50%; top: 4%; width: min(900px, 110%); aspect-ratio: 2 / 1; transform: translateX(-50%); background: radial-gradient(closest-side, rgb(200 255 0 / 0.42), transparent 72%); filter: blur(48px); pointer-events: none; }
:global(.dark .yc .yc-hero::before) { background: radial-gradient(closest-side, rgb(200 255 0 / 0.16), transparent 72%); }
.yc-hero > * { position: relative; z-index: 1; }
.yc-hero > .yc-hero-dots { position: absolute; z-index: 0; top: 0; left: 50%; width: 100vw; height: min(100%, 880px); transform: translateX(-50%); pointer-events: none; }
.yc-hero-dots i { position: absolute; inset: 0; background-size: 26px 26px; background-position: center top; }
.yc-hero-dots i:first-child { background-image: radial-gradient(circle, rgb(11 15 12 / 0.17) 1.1px, transparent 1.6px); -webkit-mask-image: radial-gradient(ellipse 62% 72% at 50% 36%, #000 18%, transparent 78%); mask-image: radial-gradient(ellipse 62% 72% at 50% 36%, #000 18%, transparent 78%); }
.yc-hero-dots i:last-child { background-image: radial-gradient(circle, var(--yc-ink) 1.7px, transparent 2.2px); -webkit-mask-image: radial-gradient(210px circle at var(--mx, 50%) var(--my, 30%), #000, transparent); mask-image: radial-gradient(210px circle at var(--mx, 50%) var(--my, 30%), #000, transparent); opacity: 0; transition: opacity 0.45s ease; }
.yc-hero-dots.is-live i:last-child { opacity: 1; }
:global(.dark .yc .yc-hero-dots i:first-child) { background-image: radial-gradient(circle, rgb(244 246 241 / 0.14) 1.1px, transparent 1.6px); }
:global(.dark .yc .yc-hero-dots i:last-child) { background-image: radial-gradient(circle, #c8ff00 1.7px, transparent 2.2px); }
.yc-h1 { margin-top: 22px; font-size: clamp(40px, 8.4vw, 104px); line-height: 1.03; letter-spacing: -0.05em; font-weight: 800; }
.yc-h1-line { display: block; }
.yc-lead { max-width: 640px; margin: 30px auto 0; font-size: clamp(16px, 1.7vw, 20px); line-height: 1.7; color: var(--yc-mut); text-wrap: pretty; }
.yc-hero-cta { display: flex; flex-wrap: wrap; justify-content: center; gap: 12px; margin-top: 38px; }
.yc-live { margin-top: clamp(44px, 6vw, 72px); }
.yc-live-lab { display: inline-flex; align-items: center; gap: 9px; font-size: 13.5px; font-weight: 650; color: var(--yc-mut); }
.yc-live-lab i { width: 8px; height: 8px; border-radius: 50%; background: var(--yc-lime); box-shadow: 0 0 0 1px var(--yc-ink); animation: yc-blink 1.6s ease-in-out infinite; }
.yc-live-num { margin-top: 8px; font-size: clamp(34px, 7.2vw, 88px); line-height: 1.05; overflow-wrap: anywhere; }
.yc-live-sub { margin-top: 10px; font-size: 14px; color: var(--yc-mut); }
@keyframes yc-blink { 50% { opacity: 0.35; } }
.yc-hero-flow { margin-top: clamp(48px, 7vw, 88px); }

.yc-pain { padding-top: clamp(96px, 14vw, 184px); }
.yc-pain-head { position: relative; }
.yc-pain-n { position: absolute; right: 0; top: -0.32em; font-size: clamp(96px, 17vw, 220px); line-height: 1; color: transparent; -webkit-text-stroke: 1.5px var(--yc-line); pointer-events: none; user-select: none; }
.yc-pain .yc-h2 { position: relative; margin-top: 16px; font-size: clamp(30px, 5.4vw, 64px); line-height: 1.14; letter-spacing: -0.04em; }
.yc-pain-old { color: var(--yc-dim); }
.yc-vs { display: grid; grid-template-columns: 1fr 1.15fr; gap: 20px; margin-top: clamp(36px, 5vw, 60px); align-items: stretch; }
.yc-vs-col { padding: 28px 30px 30px; border-radius: 24px; }
.yc-vs-before { border: 1px dashed var(--yc-line); color: var(--yc-mut); }
.yc-vs-lab { display: inline-block; padding: 4px 12px; border-radius: 999px; font-size: 12.5px; font-weight: 700; letter-spacing: 0.06em; }
.yc-vs-before .yc-vs-lab { background: var(--yc-fill); color: var(--yc-mut); }
.yc-vs-after .yc-vs-lab { background: var(--yc-lime); color: var(--yc-on-lime); }
.yc-vs-col ul { margin-top: 20px; display: grid; gap: 12px; font-size: 17px; line-height: 1.55; }
.yc-vs-col li { position: relative; padding-left: 26px; }
.yc-vs-before li::before { content: '×'; position: absolute; left: 2px; color: var(--yc-dim); }
.yc-vs-before li { text-decoration: line-through; text-decoration-color: var(--yc-line); }
.yc-vs-after li { font-weight: 600; }
.yc-vs-after li::before { content: ''; position: absolute; left: 2px; top: 0.5em; width: 11px; height: 6px; border-left: 2.5px solid var(--yc-ink); border-bottom: 2.5px solid var(--yc-ink); transform: rotate(-45deg); }
.yc-vs-code { margin-top: 18px; padding: 14px 16px; border-radius: 14px; background: var(--yc-fill); font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; font-size: 12.5px; line-height: 1.7; color: var(--yc-dim); overflow-x: auto; }
.yc-vs-formula { margin-top: 18px; padding: 14px 16px; border-radius: 14px; background: var(--yc-fill); font-size: 15px; color: var(--yc-dim); }
.yc-vs-price { margin-top: 18px; display: flex; flex-wrap: wrap; align-items: baseline; gap: 6px 12px; }
.yc-vs-price span { width: 100%; font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; font-size: 13.5px; color: var(--yc-mut); }
.yc-vs-price b { font-size: clamp(44px, 6vw, 68px); line-height: 1; }
.yc-vs-price small { font-size: 14px; color: var(--yc-mut); }
.yc-client-get { margin-top: 22px; display: flex; flex-wrap: wrap; align-items: center; gap: 12px 16px; }
.yc-client-alt { font-size: 13.5px; color: var(--yc-mut); }
.yc-client-alt a { text-decoration: underline; text-underline-offset: 3px; }
.yc-client-alt a:hover { color: var(--yc-ink); }
.yc-client-note { margin-top: 12px; font-size: 13.5px; color: var(--yc-mut); }
.yc-shot { margin: 24px -30px -30px; border-top: 1px solid var(--yc-line); border-radius: 0 0 23px 23px; overflow: hidden; background: var(--yc-fill); }
.yc-shot figcaption { display: flex; align-items: center; gap: 6px; height: 34px; padding: 0 14px; border-bottom: 1px solid var(--yc-line); font-size: 12px; color: var(--yc-mut); }
.yc-shot figcaption i { width: 9px; height: 9px; border-radius: 50%; background: var(--yc-line); }
.yc-shot figcaption span { margin-left: 8px; }
.yc-shot img { display: block; width: 100%; height: auto; }
.yc-shot-dark { display: none !important; }
:global(.dark .yc .yc-shot-light) { display: none; }
:global(.dark .yc .yc-shot-dark) { display: block !important; }
.yc-vs-after { transform: perspective(1100px) rotateX(var(--rx, 0deg)) rotateY(var(--ry, 0deg)); transition: transform 0.25s ease-out; }
.yc-vs-after::after { content: ''; position: absolute; inset: 0; border-radius: inherit; pointer-events: none; background: radial-gradient(380px circle at var(--sx, 50%) var(--sy, 50%), rgb(200 255 0 / 0.22), transparent 65%); opacity: 0; transition: opacity 0.3s ease; }
.yc-vs-after:hover::after { opacity: 1; }
.yc-vs-contact { margin-top: 18px; padding-top: 16px; border-top: 1px solid var(--yc-line); font-size: 15px; color: var(--yc-mut); white-space: pre-line; }

.yc-sec { padding-top: clamp(96px, 14vw, 184px); }
.yc-sec .yc-h2 { margin-top: 14px; max-width: 760px; }
.yc-price { text-align: center; }
.yc-price .yc-h2 { margin-left: auto; margin-right: auto; }
.yc-price-top { margin-top: clamp(36px, 5vw, 64px); }
.yc-price-name { font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; font-size: 15px; color: var(--yc-mut); }
.yc-price-was span { display: inline-block; position: relative; }
.yc-price-was { margin-top: 18px; font-size: clamp(20px, 2.6vw, 30px); color: var(--yc-dim); text-decoration: line-through; }
.yc-price-now { margin-top: 4px; font-size: clamp(88px, 19vw, 240px); line-height: 1; }
.yc-price-unit { margin-top: 20px; display: flex; flex-wrap: wrap; justify-content: center; align-items: center; gap: 12px; font-size: 16px; color: var(--yc-mut); }
.yc-chip { display: inline-block; padding: 4px 12px; border-radius: 999px; background: var(--yc-ink); color: var(--yc-bg); font-size: 13px; font-weight: 700; white-space: nowrap; }
.yc-price-list { margin: clamp(40px, 6vw, 72px) auto 0; max-width: 820px; padding: 6px 14px 14px; text-align: left; }
.yc-price-out { margin-top: 10px; font-size: 14px; color: var(--yc-dim); }
.yc-price-pick { padding: 14px 14px 8px; font-size: 13px; color: var(--yc-mut); }
.yc-price-row { width: 100%; display: grid; grid-template-columns: minmax(0, 1fr) auto auto 96px; gap: 18px; align-items: center; padding: 16px 14px; border-radius: 14px; text-align: left; transition: background-color 0.15s ease; }
.yc-price-row:hover { background: var(--yc-fill); }
.yc-price-row.is-on { background: var(--yc-lime); color: var(--yc-on-lime); }
.yc-price-row.is-on s { color: rgb(11 15 12 / 0.55); }
.yc-price-row.is-on .yc-chip { background: var(--yc-on-lime); color: var(--yc-lime); }
.yc-price-row-name { min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; font-size: 14.5px; }
.yc-price-row s { color: var(--yc-dim); font-size: 15px; }
.yc-price-row b { font-size: 26px; }
.yc-price-row .yc-chip { justify-self: end; }
.yc-price-note { margin: 22px auto 0; max-width: 640px; font-size: 13px; line-height: 1.7; color: var(--yc-dim); }
.yc-price-all { display: inline-block; margin-top: 18px; font-weight: 650; border-bottom: 2px solid var(--yc-lime); }

.yc-faq { display: grid; grid-template-columns: 1fr 1.4fr; gap: 48px; }
.yc-faq-list { border-top: 1.5px solid var(--yc-ink); }
.yc-faq-list details { border-bottom: 1px solid var(--yc-line); }
.yc-faq-list summary { display: flex; justify-content: space-between; gap: 16px; padding: 22px 0; font-size: 17px; font-weight: 650; cursor: pointer; list-style: none; }
.yc-faq-list summary::-webkit-details-marker { display: none; }
.yc-faq-list summary span { color: var(--yc-dim); font-weight: 400; transition: transform 0.2s ease; }
.yc-faq-list details[open] summary span { transform: rotate(45deg); }
.yc-faq-list details p { padding: 0 40px 24px 0; color: var(--yc-mut); line-height: 1.75; }
.yc-final { position: relative; overflow: hidden; margin-top: clamp(96px, 14vw, 184px); padding: clamp(48px, 8vw, 96px) 24px; border-radius: 32px; background: var(--yc-lime); color: var(--yc-on-lime); text-align: center; }
.yc-final::before { content: ''; position: absolute; inset: 0; background: linear-gradient(115deg, rgb(255 255 255 / 0.55), transparent 38%); pointer-events: none; }
.yc-final > * { position: relative; }
.yc-final .yc-h2 { margin: 0 auto 32px; max-width: 640px; }
.yc-final .yc-btn-lime { background: var(--yc-on-lime); color: var(--yc-lime); box-shadow: none; }
.yc-final .yc-btn-lime:hover { box-shadow: 0 10px 24px -8px rgb(11 15 12 / 0.6); }
.yc-foot { display: flex; justify-content: space-between; gap: 16px; padding-top: 48px; padding-bottom: 48px; font-size: 13.5px; color: var(--yc-dim); }
@media (max-width: 860px) {
  .yc-top-nav { display: none; }
  .yc-vs, .yc-faq { grid-template-columns: 1fr; }
  .yc-vs-col { padding: 22px 22px 24px; }
  .yc-shot { margin: 22px -22px -24px; }
  .yc-faq { gap: 32px; }
  .yc-price-row { grid-template-columns: minmax(0, 1fr) auto; }
  .yc-price-row s, .yc-price-row > span:last-child { display: none; }
  .yc-price-list { padding: 4px 8px 8px; }
}
@media (max-width: 480px) {
  .yc-hide-sm { display: none; }
  .yc-top-in { gap: 12px; }
}
@media (prefers-reduced-motion: no-preference) {
  .yc-lead, .yc-hero-cta, .yc-hero-flow { animation: yc-rise 1s cubic-bezier(0.16, 1, 0.3, 1) 0.35s both; }

  /* Opening: both headline lines surge up from below, then the highlighter sweeps. */
  .yc-h1-line { overflow: hidden; padding: 0.3em 0.3em 0.1em; margin: -0.3em -0.3em -0.1em; }
  .yc-h1-line > span { display: inline-block; transform-origin: 0 100%; animation: yc-surge 1.25s cubic-bezier(0.2, 1.35, 0.35, 1) both; }
  .yc-h1-line + .yc-h1-line > span { animation-delay: 0.14s; }
  .yc-home .yc-h1.is-in .yc-mark { animation-delay: 0.95s; animation-duration: 0.45s; }

  /* Narrow screens: the two cards swing in from either side. */
  .yc-home:not(.is-pin) .yc-vs-col { transition: opacity 0.6s ease, transform 0.95s cubic-bezier(0.2, 1.4, 0.35, 1); }
  .yc-home:not(.is-pin) .yc-vs-before:not(.is-in) { opacity: 0; transform: translateX(-60%) rotate(-14deg) scale(0.8); }
  .yc-home:not(.is-pin) .yc-vs-after:not(.is-in) { opacity: 0; transform: translateX(70%) rotate(14deg) scale(0.8); }

  /* Wide screens: each pain screen holds still while the old card is knocked back and the new one takes over. */
  .is-pin .yc-pain { height: 240vh; padding-top: 0; }
  .is-pin .yc-pain-stage { position: sticky; top: 64px; height: calc(100vh - 64px); display: flex; flex-direction: column; justify-content: center; }
  .is-pin .yc-pain .yc-h2 { font-size: clamp(30px, 4.4vw, 54px); }
  .is-pin .yc-pain-n { top: -0.42em; font-size: clamp(280px, 48vw, 720px); transform: translateX(calc((0.5 - var(--p, 0)) * 190vw)); }
  .is-pin .yc-pain-old { opacity: calc(1 - var(--a, 0) * 0.55); }
  .yc-home.is-pin .yc-pain .yc-mark { animation: none; background-size: calc(var(--a, 0) * 100%) 100%; }
  :global(.dark .yc.yc-home.is-pin .yc-pain .yc-mark) { animation: none; background-size: 200% 100%; background-position: calc((1 - var(--a, 0)) * 100%) 0; }
  .is-pin .yc-vs { grid-template-columns: minmax(0, 1fr); margin-top: clamp(20px, 4vh, 44px); }
  .is-pin .yc-vs-col { grid-area: 1 / 1; align-self: center; width: 100%; max-width: 920px; margin: 0 auto; }
  .is-pin .yc-vs-before { background: var(--yc-bg); transform-origin: 0 100%; transform: perspective(1400px) translate(calc(var(--a, 0) * -95vw), calc(var(--a, 0) * 16vh)) rotate(calc(var(--a, 0) * -38deg)) rotateY(calc(var(--a, 0) * 70deg)) scale(calc(1 - var(--a, 0) * 0.35)); opacity: calc(1 - var(--a, 0) * 0.9); }
  .is-pin .yc-vs-after { z-index: 1; transform-origin: 50% 0; transform: translate(calc((1 - var(--b, 0)) * 92vw), calc((1 - var(--a, 0)) * 30vh)) rotate(calc((1 - var(--b, 0)) * 32deg)) scale(calc(0.45 + var(--b, 0) * 0.55)) perspective(1100px) rotateX(var(--rx, 0deg)) rotateY(var(--ry, 0deg)); transition: transform 0.18s ease-out; box-shadow: 0 1px 0 rgb(255 255 255 / 0.9) inset, 0 40px 90px -30px rgb(11 15 12 / 0.4); }
  .is-pin .yc-shot img { max-height: min(22vh, 240px); object-fit: cover; object-position: top; }
  @media (max-height: 800px) {
    .is-pin .yc-shot img { max-height: 13vh; }
    .is-pin .yc-pain .yc-h2 { font-size: clamp(28px, 3.6vw, 44px); }
  }
  .is-pin .yc-vs-after li { transition-delay: calc(0.1s + var(--i) * 0.12s); }
  .is-pin .yc-vs-after li::before { transition-delay: calc(0.25s + var(--i) * 0.12s); }

  /* Price: the official price is struck out and shrinks away, then ours slams in. */
  .yc-price-top { --d: 0.2s; }
  .yc-price-top.is-quick { --d: 0s; }
  .yc-price .yc-price-was { text-decoration: none; opacity: 0; }
  .yc-price.is-in .yc-price-was { animation: yc-was 1.15s cubic-bezier(0.2, 0.7, 0.2, 1) var(--d) both; }
  .yc-price-was span::after { content: ''; position: absolute; left: -4%; right: -4%; top: 52%; height: 0.09em; background: currentColor; transform: scaleX(0); transform-origin: left; }
  .yc-price.is-in .yc-price-was span::after { animation: yc-cross 0.3s ease-out calc(var(--d) + 0.5s) both; }
  .yc-price .yc-price-now { opacity: 0; animation: none; }
  .yc-price.is-in .yc-price-now { animation: yc-slam 0.75s cubic-bezier(0.2, 0.8, 0.2, 1) calc(var(--d) + 1s) both; }
  .yc-home .yc-price.is-in .yc-price-now .yc-mark { animation-delay: calc(var(--d) + 1.5s); animation-duration: 0.6s; }
  .yc-price.is-in .yc-price-top { animation: yc-quake 0.55s linear calc(var(--d) + 1.38s) both; }
  .yc-price.is-in .yc-price-top + .yc-price-list { animation: yc-jolt 0.7s cubic-bezier(0.3, 1.6, 0.5, 1) calc(var(--d) + 1.4s) both; }
  .yc-price .yc-price-unit, .yc-price .yc-price-out { opacity: 0; }
  .yc-price.is-in .yc-price-unit, .yc-price.is-in .yc-price-out { animation: yc-rise 0.6s ease calc(var(--d) + 1.55s) both; }
  .yc-lead { animation-delay: 0.45s; }
  .yc-hero-cta { animation-delay: 0.55s; }
  .yc-hero-flow { animation-delay: 0.75s; }
  .yc-live { animation: yc-rise 1s 0.65s cubic-bezier(0.16, 1, 0.3, 1) both; }

  /* The highlighter is drawn across each headline as it comes into view. */
  .yc-home .yc-mark { background-repeat: no-repeat; background-size: 0% 100%; }
  .yc-home .is-in .yc-mark { animation: yc-draw 0.9s cubic-bezier(0.3, 0.7, 0.2, 1) 0.3s both; }
  :global(.dark .yc.yc-home .yc-mark) { background: linear-gradient(90deg, #c8ff00 50%, var(--yc-ink) 50%) 100% 0 / 200% 100% no-repeat; -webkit-background-clip: text; background-clip: text; color: transparent; }
  :global(.dark .yc.yc-home .is-in .yc-mark) { animation: yc-draw-dark 0.9s cubic-bezier(0.3, 0.7, 0.2, 1) 0.3s both; }

  /* Before: struck out one by one. Now: ticked one by one. */
  .yc-vs-col li { opacity: 0; transform: translateY(10px); transition: opacity 0.5s ease, transform 0.5s cubic-bezier(0.2, 0.7, 0.2, 1); transition-delay: calc(var(--i) * 0.13s); }
  .yc-vs-col.is-in li { opacity: 1; transform: none; }
  .yc-vs-before li { text-decoration: none; }
  .yc-vs-before li span { background: linear-gradient(var(--yc-dim), var(--yc-dim)) no-repeat 0 56% / 0% 1.5px; transition: background-size 0.5s ease; transition-delay: calc(0.4s + var(--i) * 0.13s); }
  .yc-vs-before.is-in li span { background-size: 100% 1.5px; }
  .yc-vs-after li { transition-delay: calc(0.8s + var(--i) * 0.16s); }
  .yc-vs-after li::before { transform: rotate(-45deg) scale(0); transition: transform 0.4s cubic-bezier(0.3, 1.7, 0.5, 1); transition-delay: calc(1s + var(--i) * 0.16s); }
  .yc-vs-after.is-in li::before { transform: rotate(-45deg) scale(1); }

  @media (max-width: 860px) {
    .yc-vs-after li { transition-delay: calc(0.1s + var(--i) * 0.16s); }
    .yc-vs-after li::before { transition-delay: calc(0.3s + var(--i) * 0.16s); }
  }

  .yc-hero-cta .yc-btn-lime { position: relative; overflow: hidden; }
  .yc-hero-cta .yc-btn-lime::after { content: ''; position: absolute; top: 0; bottom: 0; left: -60%; width: 40%; background: linear-gradient(100deg, transparent, rgb(255 255 255 / 0.8), transparent); transform: skewX(-18deg); animation: yc-sheen 3.6s ease-in-out 1.4s infinite; }
}
@keyframes yc-draw { to { background-size: 100% 100%; } }
@keyframes yc-draw-dark { to { background-position: 0 0; } }
@keyframes yc-sheen { 0%, 62% { left: -60%; } 100% { left: 130%; } }
@keyframes yc-surge { from { transform: translateY(190%) rotate(14deg) scale(1.25); } to { transform: none; } }
@keyframes yc-was { 0% { opacity: 0; transform: translateY(40px) scale(1.7); } 22%, 72% { opacity: 1; transform: scale(1.7); } 100% { opacity: 1; transform: none; } }
@keyframes yc-cross { to { transform: scaleX(1); } }
@keyframes yc-slam { 0% { opacity: 0; transform: scale(7) rotate(-8deg); } 50% { opacity: 1; transform: scale(0.86) rotate(1deg); } 72% { transform: scale(1.06); } 100% { opacity: 1; transform: none; } }
@keyframes yc-quake { 0%, 100% { transform: none; } 15% { transform: translate(-14px, 10px) rotate(-0.8deg); } 30% { transform: translate(12px, -8px) rotate(0.7deg); } 45% { transform: translate(-9px, 6px) rotate(-0.4deg); } 60% { transform: translate(6px, -4px); } 80% { transform: translate(-3px, 2px); } }
@keyframes yc-jolt { 0% { transform: none; } 25% { transform: translateY(26px) rotate(1.2deg); } 100% { transform: none; } }
@keyframes yc-rise { from { opacity: 0; transform: translateY(48px); } to { opacity: 1; transform: none; } }
</style>
