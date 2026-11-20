<template>
  <section class="yag" data-test="agent-panel">
    <div class="yag-head">
      <div>
        <h3>{{ t('affiliate.agent.title') }}</h3>
        <p class="yag-sub">{{ t('affiliate.agent.rule', { rate: `${agent.tiers?.length ? agent.current_percent : agent.share_percent}%` }) }}</p>
        <p v-if="agent.tiers?.length" class="yag-sub" data-test="agent-tier">
          {{
            agent.next_tier
              ? t('affiliate.agent.nextTier', {
                  left: money(Math.max(agent.next_tier.from - agent.month_spent, 0)),
                  rate: `${agent.next_tier.percent}%`
                })
              : t('affiliate.agent.topTier')
          }}
        </p>
      </div>
      <button
        v-if="!pending && !formOpen"
        type="button"
        class="yag-btn yag-btn-lime"
        data-test="agent-payout-open"
        :disabled="agent.available <= 0"
        @click="openForm"
      >
        {{ t('affiliate.agent.payout.open') }}
      </button>
    </div>

    <div class="yag-facts">
      <div>
        <p class="yag-label">{{ t('affiliate.agent.monthShared') }}</p>
        <p class="yag-value" data-test="agent-month">{{ money(agent.month_shared) }}</p>
        <p class="yag-note">{{ t('affiliate.agent.fromSpent', { amount: money(agent.month_spent) }) }}</p>
      </div>
      <div>
        <p class="yag-label">{{ t('affiliate.agent.shared') }}</p>
        <p class="yag-value">{{ money(agent.shared) }}</p>
        <p class="yag-note">{{ t('affiliate.agent.fromSpent', { amount: money(agent.spent) }) }}</p>
      </div>
      <div>
        <p class="yag-label">{{ t('affiliate.agent.available') }}</p>
        <p class="yag-value">{{ money(agent.available) }}</p>
        <p class="yag-note">{{ t('affiliate.agent.availableNote') }}</p>
      </div>
    </div>

    <p v-if="pending" class="yag-pending" data-test="agent-pending">
      {{ t('affiliate.agent.payout.pending', { amount: money(pending.amount) }) }}
    </p>

    <form v-if="formOpen" class="yag-form" data-test="agent-payout-form" @submit.prevent="submit">
      <label>
        <span>{{ t('affiliate.agent.payout.amount') }}</span>
        <input v-model.number="amount" type="number" min="0.01" step="0.01" :max="agent.available" data-test="agent-payout-amount" />
      </label>
      <label>
        <span>{{ t('affiliate.agent.payout.payee') }}</span>
        <input v-model="payee" type="text" maxlength="500" :placeholder="t('affiliate.agent.payout.payeePlaceholder')" data-test="agent-payout-payee" />
      </label>
      <div class="yag-form-actions">
        <button type="button" class="yag-btn yag-btn-ghost" @click="formOpen = false">{{ t('common.cancel') }}</button>
        <button type="submit" class="yag-btn yag-btn-lime" :disabled="sending" data-test="agent-payout-submit">
          {{ t('affiliate.agent.payout.submit') }}
        </button>
      </div>
      <p class="yag-note">{{ t('affiliate.agent.payout.hint') }}</p>
    </form>

    <template v-if="agent.payouts.length">
      <h4>{{ t('affiliate.agent.payout.history') }}</h4>
      <ul class="yag-list">
        <li v-for="item in agent.payouts" :key="item.id" data-test="agent-payout-row">
          <div>
            <b>{{ money(item.amount) }}</b>
            <span>{{ formatDateTime(item.created_at) }}<template v-if="item.admin_note"> · {{ item.admin_note }}</template></span>
          </div>
          <em :class="`yag-status-${item.status}`">{{ t(`affiliate.agent.payout.status.${item.status}`) }}</em>
        </li>
      </ul>
    </template>

    <h4>{{ t('affiliate.agent.users.title') }}</h4>
    <p v-if="agent.invitees.length === 0" class="yag-empty">{{ t('affiliate.agent.users.empty') }}</p>
    <ul v-else class="yag-list">
      <li v-for="item in agent.invitees" :key="item.user_id" data-test="agent-user-row">
        <div>
          <b>{{ item.email || '-' }}</b>
          <span>{{ t('affiliate.agent.users.spent', { amount: money(item.spent) }) }}</span>
        </div>
        <em>{{ t('affiliate.agent.users.share', { amount: money(item.share) }) }}</em>
      </li>
    </ul>
  </section>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import userAPI from '@/api/user'
import type { AgentOverview } from '@/api/user'
import { useAppStore } from '@/stores/app'
import { formatDateTime } from '@/utils/format'
import { extractApiErrorMessage } from '@/utils/apiError'

const props = defineProps<{
  agent: AgentOverview
  /** The invite page's own way of printing an amount, so both read the same. */
  money: (value: number | null | undefined) => string
}>()
const emit = defineEmits<{ (event: 'changed'): void }>()

const { t } = useI18n()
const appStore = useAppStore()

const pending = computed(() => props.agent.payouts.find((item) => item.status === 'pending') ?? null)
const formOpen = ref(false)
const sending = ref(false)
const amount = ref(0)
const payee = ref('')

function openForm() {
  amount.value = Math.floor(props.agent.available * 100) / 100
  formOpen.value = true
}

async function submit() {
  if (!(amount.value > 0) || amount.value > props.agent.available) {
    appStore.showError(t('affiliate.agent.payout.amountInvalid', { max: props.money(props.agent.available) }))
    return
  }
  if (!payee.value.trim()) {
    appStore.showError(t('affiliate.agent.payout.payeeRequired'))
    return
  }
  sending.value = true
  try {
    await userAPI.requestAgentPayout(amount.value, payee.value.trim())
    formOpen.value = false
    appStore.showSuccess(t('affiliate.agent.payout.sent'))
    emit('changed')
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('affiliate.agent.payout.failed')))
  } finally {
    sending.value = false
  }
}
</script>

<style scoped>
.yag { padding: 22px 24px; border: 1px solid var(--yc-line); border-radius: 22px; background: var(--yc-paper); }
.yag-head { display: flex; flex-wrap: wrap; align-items: flex-start; justify-content: space-between; gap: 12px; }
.yag h3 { font-size: 18px; font-weight: 720; color: var(--yc-ink); }
.yag h4 { margin-top: 22px; font-size: 14px; font-weight: 680; color: var(--yc-ink); }
.yag-sub { margin-top: 4px; color: var(--yc-mut); font-size: 14px; }
.yag-facts { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 12px; margin-top: 18px; }
.yag-facts > div { padding: 14px 16px; border: 1px solid var(--yc-line); border-radius: 16px; }
.yag-label { color: var(--yc-mut); font-size: 13px; }
.yag-value { margin-top: 4px; font-size: 22px; font-weight: 720; color: var(--yc-ink); font-variant-numeric: tabular-nums; }
.yag-note { margin-top: 4px; color: var(--yc-mut); font-size: 12px; }
.yag-pending { margin-top: 16px; padding: 12px 14px; border-radius: 14px; background: color-mix(in srgb, var(--yc-lime) 22%, transparent); color: var(--yc-ink); font-size: 14px; }
.yag-form { display: grid; gap: 12px; margin-top: 16px; padding: 16px; border: 1px solid var(--yc-line); border-radius: 16px; }
.yag-form label { display: grid; gap: 6px; color: var(--yc-ink); font-size: 14px; }
.yag-form input { height: 42px; padding: 0 12px; border: 1px solid var(--yc-line); border-radius: 12px; background: transparent; color: var(--yc-ink); font-size: 14px; }
.yag-form-actions { display: flex; justify-content: flex-end; gap: 8px; }
.yag-btn { display: inline-flex; align-items: center; justify-content: center; height: 40px; padding: 0 18px; border: 0; border-radius: 999px; font-size: 14px; font-weight: 650; cursor: pointer; }
.yag-btn:disabled { opacity: 0.5; cursor: not-allowed; }
.yag-btn-lime { background: var(--yc-lime); color: var(--yc-on-lime); box-shadow: inset 0 0 0 1px rgba(0, 0, 0, 0.12); }
.yag-btn-ghost { background: transparent; color: var(--yc-ink); box-shadow: inset 0 0 0 1px var(--yc-line); }
.yag-list { margin: 8px 0 0; padding: 0; list-style: none; }
.yag-list li { display: flex; align-items: center; justify-content: space-between; gap: 16px; padding: 12px 0; border-top: 1px solid var(--yc-line); }
.yag-list li:first-child { border-top: 0; }
.yag-list li > div { display: grid; gap: 2px; min-width: 0; }
.yag-list b { color: var(--yc-ink); font-size: 14px; font-weight: 650; overflow-wrap: anywhere; }
.yag-list span { color: var(--yc-mut); font-size: 12px; }
.yag-list em { flex-shrink: 0; color: var(--yc-ink); font-size: 13px; font-style: normal; }
.yag-list em.yag-status-rejected { color: var(--yc-mut); }
.yag-empty { margin-top: 10px; color: var(--yc-mut); font-size: 14px; }
@media (max-width: 640px) {
  .yag-facts { grid-template-columns: 1fr; }
}
</style>
