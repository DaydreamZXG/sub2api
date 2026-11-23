<template>
  <YcShell>
    <h1 class="yc-page-h1">{{ t('yc.connect.title') }}</h1>
    <p class="yc-page-lead">{{ t('yc.connect.lead') }}</p>

    <section class="yc-surface yc-way yc-way-client">
      <div class="yc-way-text">
        <span class="yc-badge yc-badge-lime">{{ t('yc.connect.best') }}</span>
        <h2 class="yc-sec-h2">{{ t('yc.connect.clientTitle') }}</h2>
        <p class="yc-note">{{ t('yc.connect.clientText') }}</p>
        <p v-if="hasAppLine" class="yc-note yc-way-app">
          {{ t('yc.connect.appCheaper') }}
          <router-link v-if="showPrices" to="/prices" class="yc-link">{{ t('yc.connect.appCheaperLink') }}</router-link>
        </p>
        <div class="yc-way-cta">
          <a v-if="clientUrl" :href="clientUrl" class="yc-btn yc-btn-lime">{{ t(clientLabel) }} ↓</a>
          <button type="button" class="yc-btn yc-btn-ghost" :disabled="handingOver" @click="connectApp">
            {{ t('yc.connect.appConnect') }}
          </button>
          <span class="yc-note">
            <a :href="YC_CLIENT_DOWNLOADS.windows" class="yc-link">Windows</a> ·
            <a :href="YC_CLIENT_DOWNLOADS.macos" class="yc-link">macOS</a>
          </span>
        </div>
      </div>
      <picture class="yc-way-shot" aria-hidden="true">
        <img :src="shotLight" class="yc-only-light" alt="" loading="lazy" />
        <img :src="shotDark" class="yc-only-dark" alt="" loading="lazy" />
      </picture>
    </section>

    <section class="yc-surface yc-way">
      <h2 class="yc-sec-h2">{{ t('yc.connect.fillTitle') }}</h2>
      <p class="yc-note yc-fill-lead">{{ t('yc.connect.fillLead') }}</p>

      <div class="yc-step">
        <p class="yc-step-lab"><i>1</i>{{ t('yc.connect.pickKey') }}</p>
        <p v-if="loading" class="yc-note">{{ t('yc.connect.loading') }}</p>
        <div v-else class="yc-opts">
          <button v-for="k in keys" :key="k.id" type="button" class="yc-opt" :class="{ 'is-on': k.id === key?.id }" @click="keyId = k.id">
            {{ k.name }}<small v-if="k.group">{{ k.group.name }}<template v-if="k.group.app_only"> · {{ t('yc.prices.appOnly') }}</template></small>
          </button>
          <button
            v-for="g in manualGroups"
            :key="`new-${g.id}`"
            type="button"
            class="yc-opt yc-opt-new"
            :disabled="creating"
            @click="createKey(g)"
          >
            + {{ manualGroups.length > 1 ? t('yc.connect.newKeyFor', { line: g.name }) : t('yc.connect.newKey') }}
          </button>
        </div>
        <p v-if="!loading && keys.length === 0 && manualGroups.length === 0" class="yc-note">
          {{ t(hasAppLine ? 'yc.connect.appLinesOnly' : 'yc.connect.noLine') }}
        </p>
        <p v-if="key?.group?.app_only" class="yc-note">{{ t('yc.connect.appKeyNote') }}</p>
      </div>


      <div v-if="key" class="yc-step">
        <p class="yc-step-lab"><i>2</i>{{ t('yc.connect.fillStep') }}</p>
        <dl class="yc-fill">
          <div>
            <dt>{{ t('yc.connect.baseUrl') }}</dt>
            <dd class="yc-mono">{{ rootUrl }}</dd>
            <button type="button" class="yc-opt" @click="copy(rootUrl)">{{ t('yc.connect.copyShort') }}</button>
          </div>
          <div>
            <dt>{{ t('yc.connect.addrV1') }}</dt>
            <dd class="yc-mono">{{ rootUrl }}/v1</dd>
            <button type="button" class="yc-opt" @click="copy(`${rootUrl}/v1`)">{{ t('yc.connect.copyShort') }}</button>
          </div>
          <div>
            <dt>{{ t('yc.connect.keyOf', { name: key.name }) }}</dt>
            <dd class="yc-mono">{{ mask(key.key) }}</dd>
            <button type="button" class="yc-opt" @click="copy(key.key)">{{ t('yc.connect.copyShort') }}</button>
          </div>
          <div>
            <dt>{{ t('yc.connect.modelId') }}</dt>
            <dd>{{ t('yc.connect.modelIdHint') }}</dd>
            <router-link to="/prices" class="yc-opt">{{ t('yc.connect.toPrices') }}</router-link>
          </div>
        </dl>
        <div class="yc-way-cta">
          <router-link to="/help/guide" class="yc-btn yc-btn-lime">{{ t('yc.connect.toolGuide') }} →</router-link>
          <span class="yc-note">{{ t('yc.connect.toolGuideHint') }}</span>
        </div>
      </div>
      <p class="yc-note yc-step-more"><router-link to="/keys" class="yc-link">{{ t('yc.connect.manageKeys') }} →</router-link></p>
    </section>

    <details v-if="key" class="yc-surface yc-way yc-manual">
      <summary>{{ t('yc.connect.cliTitle') }}</summary>
      <div v-if="snippet" class="yc-step">
        <p class="yc-step-lab"><i>1</i>{{ t('yc.connect.pickShell', { tool: snippet.tool }) }}</p>
        <div class="yc-opts">
          <button v-for="s in SHELLS" :key="s" type="button" class="yc-opt" :class="{ 'is-on': s === shell }" @click="shell = s">
            {{ t(`yc.connect.shell.${s}`) }}
          </button>
        </div>
      </div>

      <div class="yc-step">
        <p class="yc-step-lab"><i>{{ snippet ? 2 : 1 }}</i>{{ snippet ? t('yc.connect.paste') : t('yc.connect.files') }}</p>
        <template v-if="snippet">
          <pre class="yc-code">{{ snippet.masked }}</pre>
          <div class="yc-way-cta">
            <button type="button" class="yc-btn yc-btn-lime" @click="copy(snippet.text)">{{ t('yc.connect.copy') }}</button>
            <span class="yc-note">{{ t('yc.connect.after', { tool: snippet.tool }) }}</span>
          </div>
        </template>
        <p class="yc-note yc-step-more">
          <button type="button" class="yc-link" @click="guide = true">{{ t(snippet ? 'yc.connect.fullGuide' : 'yc.connect.openGuide') }}</button>
        </p>
      </div>
    </details>

    <UseKeyModal
      :show="guide"
      :api-key="key?.key || ''"
      :base-url="baseUrl"
      :platform="key?.group?.platform || null"
      :claude-code-only="key?.group?.claude_code_only || false"
      :allow-messages-dispatch="key?.group?.allow_messages_dispatch || false"
      @close="guide = false"
    />
  </YcShell>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { useAppStore } from '@/stores'
import { keysAPI } from '@/api/keys'
import { userGroupsAPI } from '@/api/groups'
import type { ApiKey, Group } from '@/types'
import { useClipboard } from '@/composables/useClipboard'
import { useYcAppConnect } from '@/composables/useYcAppConnect'
import { YC_CLIENT_DOWNLOADS, detectYcClientPlatform } from '@/utils/ycClient'
import { FeatureFlags, isFeatureFlagEnabled } from '@/utils/featureFlags'
import YcShell from '@/components/yecai/YcShell.vue'
import UseKeyModal from '@/components/keys/UseKeyModal.vue'
import shotLight from '@/assets/yecai/yecai-client-apps-light-showcase.webp'
import shotDark from '@/assets/yecai/yecai-client-apps-dark-showcase.webp'

type Shell = 'unix' | 'powershell' | 'cmd'
const SHELLS: Shell[] = ['unix', 'powershell', 'cmd']

const { t } = useI18n()
const appStore = useAppStore()
const { copyToClipboard } = useClipboard()

const keys = ref<ApiKey[]>([])
const groups = ref<Group[]>([])
// App-only lines bill their price only through the desktop app, so pasted keys are offered on the other lines.
const hasAppLine = computed(() => groups.value.some((g) => g.app_only))
// The price list is the model plaza's data: with the plaza off there is nothing behind the link.
const showPrices = computed(() => isFeatureFlagEnabled(FeatureFlags.modelPlaza))
const manualGroups = computed(() => groups.value.filter((g) => !g.app_only))
const loading = ref(true)
const creating = ref(false)
const keyId = ref<number | null>(null)
const guide = ref(false)

const clientPlatform = detectYcClientPlatform()
const clientUrl = clientPlatform ? YC_CLIENT_DOWNLOADS[clientPlatform] : ''
const clientLabel =
  clientPlatform === 'windows' ? 'yc.client.windows' : clientPlatform === 'macos' ? 'yc.client.macos' : 'yc.client.download'
const shell = ref<Shell>(clientPlatform === 'windows' ? 'powershell' : 'unix')

const baseUrl = computed(() => appStore.cachedPublicSettings?.api_base_url || window.location.origin)
// Claude Code appends /v1/messages itself, so it needs the address without /v1 (same as the full guide).
const claudeBaseUrl = computed(() => baseUrl.value.replace(/([^/])\/v1\/?$/, '$1').replace(/\/+$/, ''))
// The address without /v1; tools that want it spelled out get the second row.
const rootUrl = claudeBaseUrl
const key = computed(() => keys.value.find((k) => k.id === keyId.value) ?? keys.value[0] ?? null)

function mask(secret: string): string {
  return secret.length > 12 ? `${secret.slice(0, 7)}…${secret.slice(-4)}` : secret
}

// Tools that only need environment variables get a ready-to-paste snippet; the rest open the full guide.
const snippet = computed(() => {
  const k = key.value
  const vars =
    k?.group?.platform === 'anthropic'
      ? { tool: 'Claude Code', url: 'ANTHROPIC_BASE_URL', token: 'ANTHROPIC_AUTH_TOKEN' }
      : k?.group?.platform === 'gemini'
        ? { tool: 'Gemini CLI', url: 'GOOGLE_GEMINI_BASE_URL', token: 'GEMINI_API_KEY' }
        : null
  if (!k || !vars) return null
  const lines = (secret: string) =>
    [
      [vars.url, k.group?.platform === 'anthropic' ? claudeBaseUrl.value : baseUrl.value],
      [vars.token, secret]
    ]
      .map(([name, value]) =>
        shell.value === 'unix'
          ? `export ${name}="${value}"`
          : shell.value === 'powershell'
            ? `$env:${name}="${value}"`
            : `set ${name}=${value}`
      )
      .join('\n')
  return { tool: vars.tool, text: lines(k.key), masked: lines(mask(k.key)) }
})

function copy(text: string) {
  copyToClipboard(text, t('yc.connect.copied'))
}

async function createKey(group: Group) {
  creating.value = true
  try {
    const created = await keysAPI.create(t('yc.connect.keyName', { line: group.name }), group.id)
    keys.value = [{ ...created, group: created.group ?? group }, ...keys.value]
    keyId.value = created.id
    appStore.showSuccess(t('yc.connect.created'))
  } catch (error) {
    appStore.showError((error as { message?: string })?.message || t('yc.connect.createFailed'))
  } finally {
    creating.value = false
  }
}

onMounted(async () => {
  try {
    const [list, available] = await Promise.all([keysAPI.list(1, 50, { status: 'active' }), userGroupsAPI.getAvailable()])
    keys.value = list.items
    groups.value = available
  } catch (error) {
    console.error('Failed to load keys:', error)
  } finally {
    loading.value = false
  }
})

const { handingOver, connectApp } = useYcAppConnect()

// The desktop app opens /desktop/connect?auto=1: hand over without another click.
// The flag is dropped so reloading the page does not hand over again.
const route = useRoute()
const router = useRouter()
onMounted(() => {
  if (route.query.auto !== '1') return
  void router.replace({ path: route.path, query: { ...route.query, auto: undefined } })
  void connectApp()
})
</script>

<style scoped>
.yc-way { margin-top: 20px; padding: 28px 30px; }
.yc-way-client { display: grid; grid-template-columns: minmax(0, 1fr) minmax(0, 1.1fr); gap: 28px; align-items: center; overflow: hidden; }
.yc-way-text .yc-sec-h2 { margin-top: 12px; }
.yc-way-text .yc-note { margin-top: 8px; max-width: 420px; }
.yc-way-cta { display: flex; flex-wrap: wrap; align-items: center; gap: 14px; margin-top: 18px; }
.yc-way-shot { margin: -28px -30px -28px 0; align-self: stretch; }
.yc-way-shot img { width: 100%; height: 100%; min-height: 220px; object-fit: cover; object-position: left top; border-left: 1px solid var(--yc-line); }
.yc-only-dark, :global(.dark .yc .yc-only-light) { display: none; }
:global(.dark .yc .yc-only-dark) { display: block; }
.yc-step { margin-top: 24px; }
.yc-step-lab { display: flex; align-items: center; gap: 10px; margin-bottom: 12px; font-size: 15.5px; font-weight: 650; }
.yc-step-lab i { flex: none; width: 24px; height: 24px; display: grid; place-items: center; border-radius: 50%; background: var(--yc-lime); color: var(--yc-on-lime); font-style: normal; font-size: 13px; font-weight: 800; }
.yc-opt small { font-size: 12px; font-weight: 500; opacity: 0.7; }
.yc-opt-new { box-shadow: inset 0 0 0 1px var(--yc-dim); border-style: dashed; }
.yc-step-more { margin-top: 14px; }
.yc-manual summary { cursor: pointer; font-size: 16px; font-weight: 650; list-style: none; }
.yc-manual summary::after { content: '+'; float: right; color: var(--yc-dim); font-weight: 400; }
.yc-manual[open] summary::after { content: '−'; }
.yc-fill-lead { margin-top: 8px; max-width: 640px; }
.yc-fill { display: grid; gap: 10px; }
.yc-fill > div { display: grid; grid-template-columns: 140px minmax(0, 1fr) auto; gap: 14px; align-items: center; padding: 10px 14px; border-radius: 14px; background: var(--yc-fill); }
.yc-fill dt { font-size: 13.5px; color: var(--yc-mut); }
.yc-fill dd { font-size: 13.5px; overflow-wrap: anywhere; }
.yc-fill a.yc-opt { text-decoration: none; }
@media (max-width: 760px) {
  .yc-way { padding: 22px 20px; }
  .yc-way-client { grid-template-columns: 1fr; }
  .yc-way-shot { margin: 0 -20px -22px; }
  .yc-way-shot img { border-left: 0; border-top: 1px solid var(--yc-line); }
  .yc-fill > div { grid-template-columns: minmax(0, 1fr) auto; }
  .yc-fill dt { grid-column: 1 / -1; }
}
</style>
