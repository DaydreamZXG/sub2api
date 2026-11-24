<template>
  <AppLayout>
    <div class="yaf">
      <div v-if="loading" class="yaf-loading">
        <span class="yaf-spin"></span>
      </div>

      <template v-else-if="detail">
        <section class="yaf-hero">
          <div class="yaf-hero-text">
            <h2 data-test="aff-headline">{{ headline }}</h2>
          </div>
          <div class="yaf-share">
            <div class="yaf-copy">
              <span class="yaf-copy-label">{{ t('affiliate.inviteLink') }}</span>
              <code class="yaf-copy-value">{{ inviteLink }}</code>
              <button type="button" class="yaf-btn yaf-btn-lime" @click="copyInviteLink">
                <Icon name="copy" size="sm" />
                <span>{{ t('affiliate.copyLink') }}</span>
              </button>
            </div>
            <div class="yaf-copy">
              <span class="yaf-copy-label">{{ t('affiliate.yourCode') }}</span>
              <code class="yaf-copy-value yaf-copy-code">{{ detail.aff_code }}</code>
              <button type="button" class="yaf-btn yaf-btn-ghost" @click="copyCode">
                <Icon name="copy" size="sm" />
                <span>{{ t('affiliate.copyCode') }}</span>
              </button>
            </div>
          </div>
        </section>

        <div class="yaf-facts">
          <div class="yaf-fact">
            <p class="yaf-fact-label">{{ t('affiliate.stats.invitedUsers') }}</p>
            <p class="yaf-fact-value">{{ formatCount(detail.aff_count) }}</p>
          </div>
          <div class="yaf-fact yaf-fact-main">
            <p class="yaf-fact-label">{{ t('affiliate.stats.availableQuota') }}</p>
            <p class="yaf-fact-value">{{ money(detail.aff_quota) }}</p>
            <button
              type="button"
              data-test="aff-transfer"
              class="yaf-btn yaf-btn-lime"
              :disabled="transferring || detail.aff_quota <= 0"
              @click="transferQuota"
            >
              <Icon v-if="transferring" name="refresh" size="sm" class="animate-spin" />
              <span>{{ transferring ? t('affiliate.transfer.transferring') : t('affiliate.transfer.button') }}</span>
            </button>
            <p v-if="detail.aff_quota <= 0" class="yaf-fact-note">{{ t('affiliate.transfer.empty') }}</p>
          </div>
          <div class="yaf-fact">
            <p class="yaf-fact-label">{{ t('affiliate.stats.totalQuota') }}</p>
            <p class="yaf-fact-value">{{ money(detail.aff_history_quota) }}</p>
            <p v-if="detail.aff_frozen_quota > 0" data-test="aff-frozen" class="yaf-fact-note yaf-warn">
              {{ t('affiliate.stats.frozenQuota') }}: {{ money(detail.aff_frozen_quota) }}
            </p>
          </div>
        </div>

        <AgentPanel v-if="agent" :agent="agent" :money="money" @changed="reloadAgent" />

        <section class="yaf-panel">
          <h3>{{ t('affiliate.tips.title') }}</h3>
          <ol class="yaf-steps">
            <li><span>1</span>{{ t('affiliate.tips.line1') }}</li>
            <li><span>2</span>{{ t('affiliate.tips.line2', { rate: `${formattedRebateRate}%` }) }}</li>
            <li><span>3</span>{{ t('affiliate.tips.line3') }}</li>
          </ol>
          <p v-if="detail.aff_frozen_quota > 0" class="yaf-steps-note">{{ t('affiliate.tips.line4') }}</p>
          <router-link to="/help/agents" class="yaf-agent-link" data-test="aff-agent-guide">
            {{ agent ? t('affiliate.agent.guideForAgent') : t('affiliate.agent.guide') }} →
          </router-link>
        </section>

        <section class="yaf-panel">
          <h3>{{ t('affiliate.invitees.title') }}</h3>
          <p v-if="detail.invitees.length === 0" class="yaf-empty">{{ t('affiliate.invitees.empty') }}</p>
          <ul v-else class="yaf-people">
            <li v-for="item in detail.invitees" :key="item.user_id" data-test="aff-invitee">
              <div class="yaf-person">
                <b>{{ item.email || '-' }}</b>
                <span>{{ item.username || '-' }}</span>
              </div>
              <div class="yaf-person yaf-person-right">
                <b>{{ t('affiliate.invitees.columns.rebate') }} {{ money(item.total_rebate) }}</b>
                <span>{{ formatDateTime(item.created_at) || '-' }}</span>
              </div>
            </li>
          </ul>
        </section>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import AgentPanel from '@/components/user/AgentPanel.vue'
import userAPI from '@/api/user'
import type { AgentOverview } from '@/api/user'
import type { UserAffiliateDetail } from '@/types'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import { useClipboard } from '@/composables/useClipboard'
import { formatDateTime } from '@/utils/format'
import { extractApiErrorMessage } from '@/utils/apiError'
import { ycSign as sign } from '@/composables/ycMoneyState'

const money = (value: number | null | undefined) => `${sign.value}${(value ?? 0).toFixed(2)}`

const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()
const { copyToClipboard } = useClipboard()

const loading = ref(true)
const transferring = ref(false)
const detail = ref<UserAffiliateDetail | null>(null)
const agent = ref<AgentOverview | null>(null)

// Most users are not agents; the page works the same when this fails.
async function loadAgent(): Promise<void> {
  try {
    agent.value = await userAPI.getAgent()
  } catch {
    agent.value = null
  }
}

async function reloadAgent(): Promise<void> {
  await Promise.all([loadAgent(), loadAffiliateDetail(true)])
}

const inviteLink = computed(() => {
  if (!detail.value) return ''
  if (typeof window === 'undefined') return `/register?aff=${encodeURIComponent(detail.value.aff_code)}`
  return `${window.location.origin}/register?aff=${encodeURIComponent(detail.value.aff_code)}`
})

// Rebate rate is a percentage in the range [0, 100]; backend already clamps it.
// We trim trailing zeros (e.g. 20.00 → "20", 12.50 → "12.5") for a cleaner UI.
const formattedRebateRate = computed(() => {
  const v = detail.value?.effective_rebate_rate_percent ?? 0
  const rounded = Math.round(v * 100) / 100
  return Number.isInteger(rounded) ? String(rounded) : rounded.toString()
})

// With no rebate set, the banner just carries the page's name.
const headline = computed(() =>
  (detail.value?.effective_rebate_rate_percent ?? 0) > 0
    ? t('affiliate.hero.title', { rate: `${formattedRebateRate.value}%` })
    : t('affiliate.title')
)

function formatCount(value: number): string {
  return value.toLocaleString()
}

async function loadAffiliateDetail(silent = false): Promise<void> {
  if (!silent) {
    loading.value = true
  }
  try {
    detail.value = await userAPI.getAffiliateDetail()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('affiliate.loadFailed')))
  } finally {
    if (!silent) {
      loading.value = false
    }
  }
}

async function copyCode(): Promise<void> {
  if (!detail.value?.aff_code) return
  await copyToClipboard(detail.value.aff_code, t('affiliate.codeCopied'))
}

async function copyInviteLink(): Promise<void> {
  if (!inviteLink.value) return
  await copyToClipboard(inviteLink.value, t('affiliate.linkCopied'))
}

async function transferQuota(): Promise<void> {
  if (!detail.value || detail.value.aff_quota <= 0 || transferring.value) return
  transferring.value = true
  try {
    const resp = await userAPI.transferAffiliateQuota()
    appStore.showSuccess(t('affiliate.transfer.success', { amount: money(resp.transferred_quota) }))
    await Promise.all([
      loadAffiliateDetail(true),
      loadAgent(),
      authStore.refreshUser().catch(() => undefined),
    ])
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('affiliate.transferFailed')))
  } finally {
    transferring.value = false
  }
}

onMounted(() => {
  void loadAffiliateDetail()
  void loadAgent()
})
</script>

<style scoped>
.yaf { display: grid; gap: 20px; max-width: 1040px; }
.yaf-loading { display: grid; place-items: center; padding: 64px 0; }
.yaf-spin { width: 28px; height: 28px; border: 2px solid var(--yc-line); border-top-color: var(--yc-ink); border-radius: 50%; animation: yaf-spin 0.8s linear infinite; }
@keyframes yaf-spin { to { transform: rotate(360deg); } }

.yaf-btn { display: inline-flex; align-items: center; justify-content: center; gap: 6px; padding: 8px 16px; border-radius: 8px; font-size: 14px; font-weight: 600; white-space: nowrap; }
.yaf-btn:disabled { opacity: 0.45; cursor: not-allowed; }
.yaf-btn-lime { background: var(--yc-lime); color: var(--yc-on-lime); }
.yaf-btn-lime:not(:disabled):hover { filter: brightness(0.96); }
.yaf-btn-ghost { border: 1px solid color-mix(in srgb, var(--yc-paper) 32%, transparent); color: var(--yc-paper); }
.yaf-btn-ghost:hover { background: color-mix(in srgb, var(--yc-paper) 12%, transparent); }

.yaf-hero { --yc-paper: #fff; display: grid; grid-template-columns: minmax(0, 5fr) minmax(0, 6fr); gap: 28px; align-items: center; padding: 28px; border-radius: 22px; background: var(--yc-on); color: #fff; }
.yaf-hero-text h2 { font-size: 28px; font-weight: 780; line-height: 1.2; letter-spacing: -0.02em; text-wrap: balance; }
.yaf-share { display: grid; gap: 12px; min-width: 0; }
.yaf-copy { display: grid; grid-template-columns: minmax(0, 1fr) auto; align-items: center; gap: 4px 12px; padding: 12px 14px; border-radius: 16px; background: color-mix(in srgb, var(--yc-paper) 10%, transparent); }
.yaf-copy-label { grid-column: 1 / -1; font-size: 12px; opacity: 0.6; }
.yaf-copy-value { min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: 14px; }
.yaf-copy-code { font-size: 18px; font-weight: 700; letter-spacing: 0.04em; }

.yaf-facts { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 20px; }
.yaf-fact, .yaf-panel { padding: 22px 24px; border: 1px solid var(--yc-line); border-radius: 22px; background: var(--yc-paper); }
.yaf-fact { display: grid; align-content: start; justify-items: start; gap: 6px; }
.yaf-fact-label { font-size: 13px; color: var(--yc-mut); }
.yaf-fact-value { font-size: 28px; font-weight: 760; color: var(--yc-ink); font-variant-numeric: tabular-nums; }
.yaf-fact .yaf-btn { margin-top: 6px; }
.yaf-fact-note { font-size: 12px; color: var(--yc-mut); }
.yaf-warn { color: #d97706; }

.yaf-panel h3 { font-size: 18px; font-weight: 720; color: var(--yc-ink); }
.yaf-steps { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 12px; margin: 16px 0 0; padding: 0; list-style: none; }
.yaf-steps li { display: grid; gap: 10px; padding: 16px; border-radius: 16px; background: var(--yc-fill); color: var(--yc-ink); font-size: 14px; }
.yaf-steps li span { display: grid; place-items: center; width: 26px; height: 26px; border-radius: 50%; background: var(--yc-on); color: #fff; font-size: 13px; font-weight: 700; }
.yaf-agent-link { display: inline-block; margin-top: 14px; font-size: 14px; color: var(--yc-dim); text-decoration: underline; text-underline-offset: 3px; }
.yaf-agent-link:hover { color: var(--yc-ink); }
.yaf-steps-note { margin-top: 12px; font-size: 13px; color: #d97706; }

.yaf-empty { margin-top: 16px; padding: 28px; border: 1px dashed var(--yc-line); border-radius: 16px; color: var(--yc-mut); font-size: 14px; text-align: center; }
.yaf-people { margin: 12px 0 0; padding: 0; list-style: none; }
.yaf-people li { display: flex; align-items: center; justify-content: space-between; gap: 16px; padding: 14px 0; border-top: 1px solid var(--yc-line); }
.yaf-people li:first-child { border-top: 0; }
.yaf-person { display: grid; gap: 2px; min-width: 0; font-size: 14px; }
.yaf-person b { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; color: var(--yc-ink); font-weight: 650; }
.yaf-person span { color: var(--yc-mut); font-size: 12px; }
.yaf-person-right { flex-shrink: 0; text-align: right; }
.yaf-person-right b { font-variant-numeric: tabular-nums; }

@media (max-width: 900px) {
  .yaf-hero { grid-template-columns: minmax(0, 1fr); padding: 22px 20px; }
  .yaf-hero-text h2 { font-size: 24px; }
  .yaf-facts, .yaf-steps { grid-template-columns: minmax(0, 1fr); }
  .yaf-facts { gap: 12px; }
}
</style>
