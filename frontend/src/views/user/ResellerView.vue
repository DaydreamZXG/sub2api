<template>
  <AppLayout>
    <div class="yrs">
      <div v-if="loading" class="yrs-loading"><span class="yrs-spin"></span></div>

      <section v-else-if="!reseller" class="yrs-panel" data-test="reseller-none">
        <h3>{{ t('reseller.none.title') }}</h3>
        <p class="yrs-text">{{ t('reseller.none.text') }}</p>
        <router-link to="/help" class="yrs-btn yrs-btn-line">{{ t('reseller.none.toHelp') }} →</router-link>
      </section>

      <template v-else>
        <section class="yrs-hero">
          <div>
            <h2>{{ t('reseller.headline') }}</h2>
            <p>{{ t('reseller.lead') }}</p>
          </div>
          <div class="yrs-hero-side">
            <span>{{ t('reseller.balance') }}</span>
            <b data-test="reseller-balance">{{ money(reseller.balance) }}</b>
            <router-link to="/topup" class="yrs-btn yrs-btn-lime">{{ t('reseller.topUp') }}</router-link>
          </div>
        </section>

        <p v-if="short" class="yrs-alert" data-test="reseller-short">
          {{ t('reseller.short', { balance: money(reseller.balance), need: money(reseller.reserved) }) }}
        </p>

        <div class="yrs-facts">
          <div class="yrs-fact">
            <p class="yrs-fact-label">{{ t('reseller.facts.clientUsed') }}</p>
            <p class="yrs-fact-value">{{ money(reseller.client_used) }}</p>
            <p class="yrs-fact-note">{{ t('reseller.facts.clientUsedNote') }}</p>
          </div>
          <div class="yrs-fact">
            <p class="yrs-fact-label">{{ t('reseller.facts.cost') }}</p>
            <p class="yrs-fact-value">{{ money(reseller.cost) }}</p>
            <p class="yrs-fact-note">{{ t('reseller.facts.costNote') }}</p>
          </div>
          <div class="yrs-fact">
            <p class="yrs-fact-label">{{ t('reseller.facts.margin') }}</p>
            <p class="yrs-fact-value" data-test="reseller-margin">{{ money(reseller.margin) }}</p>
            <p class="yrs-fact-note">{{ t('reseller.facts.marginNote') }}</p>
          </div>
        </div>

        <section class="yrs-panel">
          <h3>{{ t('reseller.form.title') }}</h3>
          <form class="yrs-form" data-test="reseller-form" @submit.prevent="create">
            <label>
              <span>{{ t('reseller.form.name') }}</span>
              <input v-model.trim="form.name" type="text" maxlength="100" :placeholder="t('reseller.form.namePlaceholder')" data-test="reseller-name" />
            </label>
            <label>
              <span>{{ t('reseller.form.line') }}</span>
              <select v-model="form.groupId" data-test="reseller-line">
                <option :value="null">{{ t('reseller.form.lineDefault') }}</option>
                <option v-for="group in groups" :key="group.id" :value="group.id">{{ group.name }}</option>
              </select>
            </label>
            <label>
              <span>{{ t('reseller.form.budget', { sign }) }}</span>
              <input v-model.number="form.budget" type="number" min="0.01" step="0.01" data-test="reseller-budget" />
            </label>
            <label>
              <span>{{ t('reseller.form.markup') }}</span>
              <input v-model.number="form.markup" type="number" min="1" max="100" step="0.01" data-test="reseller-markup" />
            </label>
            <label>
              <span>{{ t('reseller.form.days') }}</span>
              <input v-model.number="form.days" type="number" min="1" step="1" :placeholder="t('reseller.form.daysPlaceholder')" data-test="reseller-days" />
            </label>
            <button type="submit" class="yrs-btn yrs-btn-lime" :disabled="creating" data-test="reseller-create">{{ t('reseller.form.create') }}</button>
          </form>
          <p class="yrs-hint" data-test="reseller-explain">
            {{ markupOk ? t('reseller.form.explain', { client: money(form.markup), margin: money(form.markup - 1), budget: money(form.budget || 0), cost: money((form.budget || 0) / form.markup) }) : t('reseller.form.markupInvalid') }}
          </p>
          <p class="yrs-hint">{{ t('reseller.form.locked') }}</p>
        </section>

        <section class="yrs-panel">
          <h3>{{ t('reseller.keys.title') }}</h3>
          <p v-if="reseller.keys.length === 0" class="yrs-empty">{{ t('reseller.keys.empty') }}</p>
          <ul v-else class="yrs-keys">
            <li v-for="key in reseller.keys" :key="key.id" data-test="reseller-key">
              <div class="yrs-key-head">
                <b>{{ key.name }}</b>
                <span class="yrs-chip" :class="`is-${key.status}`">{{ t(`reseller.status.${statusName(key.status)}`) }}</span>
                <span class="yrs-key-line">{{ key.group_name || t('reseller.form.lineDefault') }} · {{ t('reseller.keys.markup', { markup: key.markup }) }}</span>
              </div>
              <div class="yrs-key-code">
                <code>{{ key.key }}</code>
                <button type="button" class="yrs-link" data-test="reseller-copy" @click="copy(key.key)">{{ t('reseller.keys.copy') }}</button>
              </div>
              <div class="yrs-bar" :title="`${percent(key)}%`"><i :style="{ width: `${percent(key)}%` }"></i></div>
              <div class="yrs-key-foot">
                <span>{{ t('reseller.keys.used', { used: money(key.client_used), budget: money(key.budget) }) }}</span>
                <span>{{ t('reseller.keys.split', { cost: money(key.cost), margin: money(key.margin) }) }}</span>
                <span v-if="key.expires_at">{{ t('reseller.keys.expires', { date: formatDateOnly(key.expires_at) }) }}</span>
                <span class="yrs-key-actions">
                  <button
                    v-if="key.status === 'active' || key.status === 'inactive'"
                    type="button"
                    class="yrs-link"
                    data-test="reseller-toggle"
                    @click="toggle(key)"
                  >
                    {{ key.status === 'active' ? t('reseller.keys.stop') : t('reseller.keys.start') }}
                  </button>
                  <button type="button" class="yrs-link is-danger" data-test="reseller-delete" @click="remove(key)">{{ t('common.delete') }}</button>
                </span>
              </div>
            </li>
          </ul>
        </section>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import resellerAPI, { type ResellerClientKey, type ResellerOverview } from '@/api/reseller'
import groupsAPI from '@/api/groups'
import keysAPI from '@/api/keys'
import type { Group } from '@/types'
import { useAppStore } from '@/stores/app'
import { useClipboard } from '@/composables/useClipboard'
import { formatDateOnly } from '@/utils/format'
import { extractApiErrorMessage } from '@/utils/apiError'
import { ycSign as sign } from '@/composables/ycMoneyState'

// Like the rest of the user pages, amounts are balance units with the site's sign.
const money = (value: number | null | undefined) => `${sign.value}${(value ?? 0).toFixed(2)}`

const { t } = useI18n()
const appStore = useAppStore()
const { copyToClipboard } = useClipboard()

const loading = ref(true)
const creating = ref(false)
const reseller = ref<ResellerOverview | null>(null)
const groups = ref<Group[]>([])
const form = reactive<{ name: string; groupId: number | null; budget: number; markup: number; days: number | '' }>({
  name: '',
  groupId: null,
  budget: 100,
  markup: 1.5,
  days: ''
})

// Calls are paid as they happen: below this, clients can run dry before their budget does.
const short = computed(() => !!reseller.value && reseller.value.reserved > 0 && reseller.value.balance < reseller.value.reserved)
const markupOk = computed(() => form.markup >= 1 && form.markup <= 100)

function percent(key: ResellerClientKey): number {
  return key.budget > 0 ? Math.min(Math.round((key.client_used / key.budget) * 100), 100) : 0
}

function statusName(status: string): string {
  return ['active', 'inactive', 'quota_exhausted', 'expired'].includes(status) ? status : 'inactive'
}

async function load(): Promise<void> {
  try {
    reseller.value = await resellerAPI.getReseller()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('reseller.loadFailed')))
  } finally {
    loading.value = false
  }
}

async function create(): Promise<void> {
  if (creating.value) return
  if (!form.name || !(form.budget > 0) || !markupOk.value) {
    appStore.showError(t('reseller.form.invalid'))
    return
  }
  creating.value = true
  try {
    const made = await resellerAPI.createClientKey({
      name: form.name,
      group_id: form.groupId,
      budget: form.budget,
      markup: form.markup,
      ...(typeof form.days === 'number' && form.days > 0 ? { expires_in_days: Math.floor(form.days) } : {})
    })
    form.name = ''
    await load()
    await copyToClipboard(made.key, t('reseller.form.created'))
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('reseller.form.failed')))
  } finally {
    creating.value = false
  }
}

async function copy(key: string): Promise<void> {
  await copyToClipboard(key, t('reseller.keys.copied'))
}

async function toggle(key: ResellerClientKey): Promise<void> {
  try {
    await keysAPI.toggleStatus(key.id, key.status === 'active' ? 'inactive' : 'active')
    await load()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('reseller.keys.failed')))
  }
}

async function remove(key: ResellerClientKey): Promise<void> {
  if (!window.confirm(t('reseller.keys.deleteConfirm', { name: key.name }))) return
  try {
    await keysAPI.deleteKey(key.id)
    await load()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('reseller.keys.failed')))
  }
}

onMounted(async () => {
  await load()
  if (!reseller.value) return
  try {
    groups.value = await groupsAPI.getAvailable()
  } catch {
    // The key can still be made on the default line.
  }
})
</script>

<style scoped>
.yrs { display: grid; gap: 20px; max-width: 1040px; }
.yrs-loading { display: grid; place-items: center; padding: 64px 0; }
.yrs-spin { width: 28px; height: 28px; border: 2px solid var(--yc-line); border-top-color: var(--yc-ink); border-radius: 50%; animation: yrs-spin 0.8s linear infinite; }
@keyframes yrs-spin { to { transform: rotate(360deg); } }

.yrs-btn { display: inline-flex; align-items: center; justify-content: center; gap: 6px; padding: 8px 16px; border-radius: 999px; font-size: 14px; font-weight: 650; white-space: nowrap; text-decoration: none; }
.yrs-btn:disabled { opacity: 0.45; cursor: not-allowed; }
.yrs-btn-lime { background: var(--yc-lime); color: var(--yc-on-lime); box-shadow: inset 0 0 0 1px rgba(0, 0, 0, 0.12); }
.yrs-btn-lime:not(:disabled):hover { filter: brightness(0.96); }
.yrs-btn-line { margin-top: 14px; border: 1px solid var(--yc-line); color: var(--yc-ink); }
.yrs-link { font-size: 13px; color: var(--yc-ink); text-decoration: underline; text-underline-offset: 3px; }
.yrs-link.is-danger { color: #dc2626; }

.yrs-hero { display: flex; align-items: center; justify-content: space-between; gap: 28px; padding: 28px; border-radius: 22px; background: var(--yc-ink); color: var(--yc-paper); }
.yrs-hero h2 { font-size: 28px; font-weight: 780; line-height: 1.2; letter-spacing: -0.02em; text-wrap: balance; }
.yrs-hero p { margin-top: 8px; font-size: 14.5px; opacity: 0.75; }
.yrs-hero-side { flex: none; display: grid; justify-items: end; gap: 4px; }
.yrs-hero-side span { font-size: 12px; opacity: 0.6; }
.yrs-hero-side b { font-size: 26px; font-weight: 760; font-variant-numeric: tabular-nums; }
.yrs-hero-side .yrs-btn { margin-top: 6px; }

.yrs-alert { padding: 14px 18px; border-radius: 16px; background: color-mix(in srgb, #f59e0b 16%, transparent); color: var(--yc-ink); font-size: 14.5px; line-height: 1.6; }

.yrs-facts { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 20px; }
.yrs-fact, .yrs-panel { padding: 22px 24px; border: 1px solid var(--yc-line); border-radius: 22px; background: var(--yc-paper); }
.yrs-fact { display: grid; align-content: start; gap: 6px; }
.yrs-fact-label { font-size: 13px; color: var(--yc-mut); }
.yrs-fact-value { font-size: 28px; font-weight: 760; color: var(--yc-ink); font-variant-numeric: tabular-nums; }
.yrs-fact-note { font-size: 12px; color: var(--yc-mut); }

.yrs-panel h3 { font-size: 18px; font-weight: 720; color: var(--yc-ink); }
.yrs-text { margin-top: 8px; font-size: 14.5px; line-height: 1.7; color: var(--yc-mut); }
.yrs-form { display: grid; grid-template-columns: 2fr 1.4fr 1fr 1fr 1fr auto; align-items: end; gap: 12px; margin-top: 16px; }
.yrs-form label { display: grid; gap: 6px; min-width: 0; font-size: 13px; color: var(--yc-mut); }
.yrs-form input, .yrs-form select { width: 100%; min-width: 0; height: 40px; padding: 0 12px; border: 1px solid var(--yc-line); border-radius: 12px; background: var(--yc-paper); color: var(--yc-ink); font-size: 14px; }
.yrs-form .yrs-btn { height: 40px; }
.yrs-hint { margin-top: 12px; font-size: 13px; line-height: 1.6; color: var(--yc-mut); }
.yrs-hint + .yrs-hint { margin-top: 4px; }

.yrs-empty { margin-top: 16px; padding: 28px; border: 1px dashed var(--yc-line); border-radius: 16px; color: var(--yc-mut); font-size: 14px; text-align: center; }
.yrs-keys { margin: 12px 0 0; padding: 0; list-style: none; }
.yrs-keys li { display: grid; gap: 10px; padding: 16px 0; border-top: 1px solid var(--yc-line); }
.yrs-keys li:first-child { border-top: 0; }
.yrs-key-head { display: flex; flex-wrap: wrap; align-items: center; gap: 8px 10px; font-size: 14px; }
.yrs-key-head b { color: var(--yc-ink); font-weight: 650; font-size: 15px; }
.yrs-key-line { color: var(--yc-mut); font-size: 13px; }
.yrs-chip { padding: 2px 10px; border-radius: 999px; background: var(--yc-fill); color: var(--yc-mut); font-size: 12px; font-weight: 600; }
.yrs-chip.is-active { background: color-mix(in srgb, var(--yc-lime) 45%, transparent); color: var(--yc-ink); }
.yrs-key-code { display: flex; align-items: center; gap: 12px; min-width: 0; }
.yrs-key-code code { min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: 13px; color: var(--yc-mut); }
.yrs-bar { height: 6px; border-radius: 999px; background: var(--yc-fill); overflow: hidden; }
.yrs-bar i { display: block; height: 100%; border-radius: 999px; background: var(--yc-ink); }
.yrs-key-foot { display: flex; flex-wrap: wrap; align-items: center; gap: 6px 18px; font-size: 13px; color: var(--yc-mut); font-variant-numeric: tabular-nums; }
.yrs-key-actions { display: flex; gap: 14px; margin-left: auto; }

@media (max-width: 900px) {
  .yrs-hero { flex-direction: column; align-items: flex-start; padding: 22px 20px; }
  .yrs-hero h2 { font-size: 24px; }
  .yrs-hero-side { justify-items: start; }
  .yrs-facts { grid-template-columns: minmax(0, 1fr); gap: 12px; }
  .yrs-form { grid-template-columns: minmax(0, 1fr) minmax(0, 1fr); }
  .yrs-form label:first-child, .yrs-form .yrs-btn { grid-column: 1 / -1; }
}
</style>
