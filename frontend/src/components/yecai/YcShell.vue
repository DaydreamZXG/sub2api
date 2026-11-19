<template>
  <div class="yc yc-shell">
    <aside class="yc-side">
      <router-link to="/dashboard" class="yc-side-brand">
        <YcBrand />
      </router-link>
      <nav ref="nav" class="yc-side-nav" :aria-label="t('yc.nav.aria')">
        <router-link v-for="l in links" :key="l.to" :to="l.to" :class="{ 'is-on': isOn(l) }">{{ l.text }}</router-link>
        <p class="yc-side-more">{{ t('yc.shell.more') }}</p>
        <router-link v-for="l in more" :key="l.to" :to="l.to" class="is-minor" :class="{ 'is-on': isOn(l) }">{{ l.text }}</router-link>
        <YcThemeSwitch class="yc-side-narrow" />
        <button type="button" class="is-minor yc-side-narrow" @click="switchLocale">{{ otherLocale.name }}</button>
        <button type="button" class="is-minor yc-side-narrow" @click="logout">{{ t('nav.logout') }}</button>
      </nav>
      <div class="yc-side-foot">
        <router-link to="/topup" class="yc-side-bal"><span>{{ t('yc.dash.balance') }}</span><b>{{ yuan(authStore.user?.balance) }}</b></router-link>
        <YcThemeSwitch class="yc-side-wide yc-side-theme" />
        <div class="yc-side-tools">
          <AnnouncementBell />
          <button type="button" class="yc-side-wide" @click="switchLocale">{{ otherLocale.name }}</button>
          <button type="button" class="yc-side-wide" @click="logout">{{ t('nav.logout') }}</button>
        </div>
        <router-link v-if="authStore.isAdmin" to="/admin/dashboard" class="yc-side-wide">{{ t('yc.shell.admin') }}</router-link>
        <router-link to="/profile" class="yc-side-me" :class="{ 'is-on': route.path === '/profile' }"><i>{{ initial }}</i><span>{{ t('yc.shell.account') }}</span></router-link>
      </div>
    </aside>
    <main class="yc-shell-main" :class="{ 'yc-legacy': legacy }"><slot /></main>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAdminSettingsStore, useAuthStore, useAppStore } from '@/stores'
import YcBrand from '@/components/yecai/YcBrand.vue'
import YcThemeSwitch from '@/components/yecai/YcThemeSwitch.vue'
import { FeatureFlags, isFeatureFlagEnabled } from '@/utils/featureFlags'
import { availableLocales, setLocale } from '@/i18n'
import { useBatchImageAccess } from '@/composables/useBatchImageAccess'
import { useYcMoney } from '@/composables/useYcMoney'
import { getMyGarden } from '@/api/garden'
import AnnouncementBell from '@/components/common/AnnouncementBell.vue'
import '@/styles/yecai.css'

// `legacy` hosts an older page that brings its own styling and needs the full width.
defineProps<{ legacy?: boolean }>()

interface NavLink {
  to: string
  text: string
  /** Other path prefixes that belong to this entry. */
  also?: string[]
}

const { t, locale } = useI18n()
const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const appStore = useAppStore()
const adminSettingsStore = useAdminSettingsStore()
const { yuan } = useYcMoney()
const { canUseBatchImage, refreshBatchImageAccess } = useBatchImageAccess()

const settings = computed(() => appStore.cachedPublicSettings)
// The garden is listed once it is open, or while the user still has crops in it.
const hasGarden = ref(false)
const initial = computed(() => (authStore.user?.username || authStore.user?.email || '?').charAt(0).toUpperCase())
const on = (show: unknown, item: NavLink): NavLink[] => (show ? [item] : [])
const flag = isFeatureFlagEnabled
const links = computed<NavLink[]>(() => {
  const full = !authStore.isSimpleMode
  return [
    { to: '/dashboard', text: t('yc.shell.start') },
    { to: '/connect', text: t('yc.shell.connect') },
    ...on(full && flag(FeatureFlags.modelPlaza), { to: '/prices', text: t('yc.shell.prices'), also: ['/model-plaza'] }),
    ...on(full, { to: '/bills', text: t('yc.shell.bills'), also: ['/usage'] }),
    ...on(full, { to: '/topup', text: t('yc.shell.topUp'), also: ['/purchase', '/orders', '/redeem', '/payment'] }),
    ...on(full && hasGarden.value, { to: '/garden', text: t('yc.shell.garden') }),
    { to: '/help', text: t('yc.shell.help') }
  ]
})
// Things most people need rarely; kept one click away.
const more = computed<NavLink[]>(() => {
  const full = !authStore.isSimpleMode
  const custom = (settings.value?.custom_menu_items ?? [])
    .filter((item) => item.visibility === 'user')
    .sort((a, b) => a.sort_order - b.sort_order)
  return [
    ...on(full && flag(FeatureFlags.subscription), { to: '/subscriptions', text: t('yc.shell.plans') }),
    ...on(full && flag(FeatureFlags.affiliate), { to: '/affiliate', text: t('yc.shell.invite') }),
    { to: '/keys', text: t('yc.shell.keys') },
    ...on(full && canUseBatchImage.value, { to: '/batch-image', text: t('nav.batchImage') }),
    ...on(full && flag(FeatureFlags.availableChannels), { to: '/available-channels', text: t('nav.availableChannels') }),
    ...on(flag(FeatureFlags.channelMonitor), { to: '/monitor', text: t('nav.channelStatus') }),
    ...on(flag(FeatureFlags.mcp), { to: '/settings/mcp', text: t('mcp.title') }),
    ...custom.map((item) => ({ to: `/custom/${item.id}`, text: item.label }))
  ]
})

function isOn(link: NavLink): boolean {
  return [link.to, ...(link.also ?? [])].some((path) => route.path === path || route.path.startsWith(`${path}/`))
}

const otherLocale = computed(() => availableLocales.find((l) => l.code !== locale.value) ?? availableLocales[0])
async function switchLocale() {
  await setLocale(otherLocale.value.code)
  // Names that come from the server are localised per request, so reload to refresh them.
  window.location.reload()
}

async function logout() {
  try {
    await authStore.logout()
  } catch (error) {
    console.error('Logout error:', error)
  }
  await router.push('/login')
}

const nav = ref<HTMLElement | null>(null)
onMounted(() => {
  if (!authStore.isSimpleMode) {
    getMyGarden()
      .then((garden) => (hasGarden.value = garden.enabled || garden.crops.some((crop) => crop.status === 'growing')))
      .catch(() => undefined)
  }
  void refreshBatchImageAccess()
  // Admin-only custom pages are resolved from these settings; the old sidebar used to load them.
  if (authStore.isAdmin) void adminSettingsStore.fetch()
  // On narrow screens the menu scrolls sideways; bring the current entry into view.
  const current = nav.value?.querySelector<HTMLElement>('.is-on')
  if (nav.value && current) {
    nav.value.scrollLeft = current.offsetLeft - nav.value.offsetLeft - (nav.value.clientWidth - current.offsetWidth) / 2
  }
})
</script>

<style scoped>
.yc-shell { min-height: 100vh; display: grid; grid-template-columns: 232px minmax(0, 1fr); }
.yc-side { position: sticky; top: 0; height: 100vh; display: flex; flex-direction: column; padding: 22px 14px 18px; border-right: 1px solid var(--yc-line); background: color-mix(in srgb, var(--yc-paper) 70%, transparent); }
.yc-side-brand { display: flex; align-items: center; gap: 10px; padding: 0 10px 22px; font-weight: 750; font-size: 17px; white-space: nowrap; }
.yc-side-nav { display: flex; flex-direction: column; gap: 2px; }
.yc-side-nav a, .yc-side-nav button { position: relative; text-align: left; padding: 10px 14px; border-radius: 12px; font-size: 15px; color: var(--yc-mut); white-space: nowrap; transition: background-color 0.15s ease, color 0.15s ease; }
.yc-side-nav a:hover, .yc-side-nav button:hover { background: var(--yc-fill); color: var(--yc-ink); }
.yc-side-more { margin: 18px 14px 6px; font-size: 12px; font-weight: 650; letter-spacing: 0.08em; color: var(--yc-dim); }
.yc-side-nav .is-minor { padding: 8px 14px; font-size: 14px; }
.yc-side-nav .yc-side-narrow { display: none; }
.yc-side-nav a.is-on { background: var(--yc-lime); color: var(--yc-on-lime); font-weight: 650; box-shadow: 0 8px 20px -10px rgb(140 180 0 / 0.9); }
.yc-side-foot { margin-top: auto; display: flex; flex-direction: column; gap: 4px; font-size: 14px; color: var(--yc-mut); }
.yc-side-foot a { display: flex; align-items: center; gap: 10px; padding: 9px 10px; border-radius: 12px; }
.yc-side-foot a:hover, .yc-side-foot a.is-on, .yc-side-tools > button:hover { background: var(--yc-fill); color: var(--yc-ink); }
.yc-side-bal { justify-content: space-between; margin-bottom: 4px; background: var(--yc-fill); }
.yc-side-bal b { font-size: 15px; color: var(--yc-ink); font-variant-numeric: tabular-nums; }
.yc-side-theme { align-self: flex-start; margin: 2px 0 2px 6px; }
.yc-side-tools { display: flex; align-items: center; gap: 2px; }
.yc-side-tools > button { padding: 7px 10px; border-radius: 10px; font-size: 13.5px; white-space: nowrap; }
.yc-side-me i { flex: none; width: 30px; height: 30px; display: grid; place-items: center; border-radius: 50%; background: var(--yc-ink); color: var(--yc-bg); font-style: normal; font-weight: 700; font-size: 13px; }
.yc-shell-main { width: 100%; max-width: 1080px; margin: 0 auto; padding: 44px 36px 80px; }
.yc-shell-main.yc-legacy { max-width: none; padding: 26px 32px 40px; }
@media (max-width: 860px) {
  .yc-shell { grid-template-columns: minmax(0, 1fr); }
  .yc-side { position: sticky; z-index: 20; height: auto; flex-direction: row; align-items: center; gap: 8px; padding: 10px 14px; border-right: 0; border-bottom: 1px solid var(--yc-line); background: color-mix(in srgb, var(--yc-bg) 86%, transparent); backdrop-filter: blur(16px); -webkit-backdrop-filter: blur(16px); }
  .yc-side-brand { padding: 0; }
  .yc-side-brand :deep(.yc-lockup-text), .yc-side-me span, .yc-side-more, .yc-side-foot .yc-side-wide, .yc-side-foot .yc-side-bal { display: none; }
  .yc-side-nav .yc-side-narrow { display: block; }
  .yc-side-nav .yc-theme-switch.yc-side-narrow { display: inline-flex; flex: none; }
  .yc-side-nav { flex: 1; min-width: 0; flex-direction: row; overflow-x: auto; scrollbar-width: none; }
  .yc-side-nav a, .yc-side-nav button, .yc-side-nav .is-minor { flex: none; padding: 7px 12px; font-size: 14.5px; }
  .yc-side-foot { margin: 0; flex-direction: row; align-items: center; }
  .yc-side-foot a { padding: 0; }
  .yc-side-foot a.is-on { background: none; }
  .yc-shell-main, .yc-shell-main.yc-legacy { padding: 24px 18px 56px; }
  .yc-shell-main.yc-legacy { padding: 18px 14px 40px; }
}
</style>
