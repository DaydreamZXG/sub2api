<template>
  <div class="space-y-5" data-test="agents-tab">
    <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('admin.profit.agents.intro') }}</p>

    <section class="card p-5" data-test="agents-pending">
      <h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('admin.profit.agents.pendingTitle') }}</h2>
      <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('admin.profit.agents.pendingHint') }}</p>
      <p v-if="pending.length === 0" class="mt-4 text-sm text-gray-400 dark:text-dark-500">{{ t('admin.profit.agents.pendingEmpty') }}</p>
      <ul v-else class="mt-4 divide-y divide-gray-100 dark:divide-dark-700">
        <li v-for="item in pending" :key="item.id" class="flex flex-wrap items-center justify-between gap-3 py-3" data-test="agents-pending-row">
          <div class="min-w-0">
            <p class="font-medium text-gray-900 dark:text-white">
              <span class="tabular-nums">{{ yuan(item.amount) }}</span>
              <span class="ml-2 text-sm font-normal text-gray-500 dark:text-dark-400">{{ item.email }}</span>
            </p>
            <p class="mt-0.5 break-words text-sm text-gray-500 dark:text-dark-400">
              {{ item.payee_note || t('admin.profit.agents.noPayee') }} · {{ formatDateTime(item.created_at) }}
            </p>
          </div>
          <div class="flex gap-2">
            <button type="button" class="btn btn-primary btn-sm" data-test="agents-pay" @click="openDecision(item, 'paid')">
              {{ t('admin.profit.agents.markPaid') }}
            </button>
            <button type="button" class="btn btn-secondary btn-sm" data-test="agents-reject" @click="openDecision(item, 'rejected')">
              {{ t('admin.profit.agents.reject') }}
            </button>
          </div>
        </li>
      </ul>
    </section>

    <section class="card p-5" data-test="agents-list">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('admin.profit.agents.listTitle') }}</h2>
        <button type="button" class="btn btn-primary btn-sm" data-test="agents-add" @click="openForm(null)">
          {{ t('admin.profit.agents.add') }}
        </button>
      </div>
      <div v-if="loading && agents.length === 0" class="mt-4 text-sm text-gray-500 dark:text-dark-400">{{ t('common.loading') }}</div>
      <p v-else-if="agents.length === 0" class="mt-4 text-sm text-gray-400 dark:text-dark-500">{{ t('admin.profit.agents.listEmpty') }}</p>
      <div v-else class="mt-4 overflow-x-auto">
        <table class="w-full min-w-[640px] text-sm">
          <thead>
            <tr class="text-left text-xs text-gray-500 dark:text-dark-400">
              <th class="pb-2 font-medium">{{ t('admin.profit.agents.columns.agent') }}</th>
              <th class="pb-2 text-right font-medium">{{ t('admin.profit.agents.columns.share') }}</th>
              <th class="pb-2 text-right font-medium">{{ t('admin.profit.agents.columns.invitees') }}</th>
              <th class="pb-2 text-right font-medium">{{ t('admin.profit.agents.columns.shared') }}</th>
              <th class="pb-2 text-right font-medium">{{ t('admin.profit.agents.columns.available') }}</th>
              <th class="pb-2"></th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
            <tr v-for="agent in agents" :key="agent.user_id" data-test="agents-row">
              <td class="py-2.5">
                <p class="font-medium text-gray-900 dark:text-white">{{ agent.email || `#${agent.user_id}` }}</p>
                <p v-if="agent.since" class="text-xs text-gray-400 dark:text-dark-500">
                  {{ t('admin.profit.agents.since', { date: formatDate(agent.since) }) }}
                </p>
              </td>
              <td class="py-2.5 text-right tabular-nums text-gray-900 dark:text-white">{{ agent.share_percent }}%</td>
              <td class="py-2.5 text-right tabular-nums text-gray-600 dark:text-dark-300">{{ agent.invitees }}</td>
              <td class="py-2.5 text-right tabular-nums text-gray-900 dark:text-white">{{ yuan(agent.shared) }}</td>
              <td class="py-2.5 text-right tabular-nums text-gray-900 dark:text-white">{{ yuan(agent.available) }}</td>
              <td class="whitespace-nowrap py-2.5 text-right">
                <button type="button" class="text-xs text-primary-600 dark:text-primary-400" @click="openForm(agent)">
                  {{ t('admin.profit.agents.changeShare') }}
                </button>
                <button type="button" class="ml-3 text-xs text-red-600 dark:text-red-400" data-test="agents-remove" @click="removeAgent(agent)">
                  {{ t('admin.profit.agents.remove') }}
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <section v-if="handled.length" class="card p-5" data-test="agents-handled">
      <h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('admin.profit.agents.handledTitle') }}</h2>
      <ul class="mt-3 divide-y divide-gray-100 text-sm dark:divide-dark-700">
        <li v-for="item in handled" :key="item.id" class="flex flex-wrap items-center justify-between gap-2 py-2.5">
          <span class="min-w-0 text-gray-600 dark:text-dark-300">
            <span class="tabular-nums text-gray-900 dark:text-white">{{ yuan(item.amount) }}</span>
            · {{ item.email }}<template v-if="item.admin_note"> · {{ item.admin_note }}</template>
          </span>
          <span class="text-xs" :class="item.status === 'paid' ? 'text-emerald-600 dark:text-emerald-400' : 'text-gray-400 dark:text-dark-500'">
            {{ t(`admin.profit.agents.status.${item.status}`) }} · {{ formatDateTime(item.handled_at ?? item.created_at) }}
          </span>
        </li>
      </ul>
    </section>

    <BaseDialog :show="formOpen" :title="editing ? t('admin.profit.agents.changeShare') : t('admin.profit.agents.add')" @close="formOpen = false">
      <form class="space-y-4" data-test="agents-form" @submit.prevent="saveAgent">
        <div v-if="editing" class="text-sm text-gray-600 dark:text-dark-300">{{ editing.email || `#${editing.user_id}` }}</div>
        <div v-else>
          <label class="block text-sm font-medium text-gray-900 dark:text-white">
            {{ t('admin.profit.agents.form.user') }}
            <input
              v-model="query"
              class="input mt-1"
              :placeholder="t('admin.profit.agents.form.userPlaceholder')"
              data-test="agents-user-search"
              @input="search"
            />
          </label>
          <ul v-if="found.length" class="mt-2 max-h-40 divide-y divide-gray-100 overflow-auto rounded-lg border border-gray-200 text-sm dark:divide-dark-700 dark:border-dark-600">
            <li v-for="user in found" :key="user.id">
              <button
                type="button"
                class="flex w-full items-center justify-between px-3 py-2 text-left hover:bg-gray-50 dark:hover:bg-dark-700"
                :class="picked?.id === user.id ? 'bg-primary-50 dark:bg-primary-900/30' : ''"
                data-test="agents-user-option"
                @click="picked = user"
              >
                <span class="text-gray-900 dark:text-white">{{ user.email }}</span>
                <span class="text-xs text-gray-400 dark:text-dark-500">{{ user.username }}</span>
              </button>
            </li>
          </ul>
        </div>
        <label class="block text-sm font-medium text-gray-900 dark:text-white">
          {{ t('admin.profit.agents.form.share') }}
          <input v-model.number="sharePercent" type="number" min="0.01" max="100" step="0.01" class="input mt-1" data-test="agents-share" />
          <span class="mt-1 block text-xs font-normal text-gray-500 dark:text-dark-400">{{ t('admin.profit.agents.form.shareHint') }}</span>
        </label>
        <div class="flex justify-end gap-2">
          <button type="button" class="btn btn-secondary" @click="formOpen = false">{{ t('common.cancel') }}</button>
          <button type="submit" class="btn btn-primary" :disabled="saving" data-test="agents-save">{{ t('common.save') }}</button>
        </div>
      </form>
    </BaseDialog>

    <BaseDialog
      :show="decision !== null"
      :title="decision?.status === 'paid' ? t('admin.profit.agents.markPaid') : t('admin.profit.agents.reject')"
      @close="decision = null"
    >
      <form v-if="decision" class="space-y-4" @submit.prevent="decide">
        <p class="text-sm text-gray-600 dark:text-dark-300">
          {{
            decision.status === 'paid'
              ? t('admin.profit.agents.paidConfirm', { amount: yuan(decision.payout.amount), email: decision.payout.email })
              : t('admin.profit.agents.rejectConfirm', { amount: yuan(decision.payout.amount), email: decision.payout.email })
          }}
        </p>
        <label class="block text-sm font-medium text-gray-900 dark:text-white">
          {{ t('admin.profit.agents.note') }}
          <input v-model="decision.note" class="input mt-1" maxlength="200" data-test="agents-note" />
        </label>
        <div class="flex justify-end gap-2">
          <button type="button" class="btn btn-secondary" @click="decision = null">{{ t('common.cancel') }}</button>
          <button type="submit" class="btn btn-primary" :disabled="saving" data-test="agents-decide">{{ t('common.confirm') }}</button>
        </div>
      </form>
    </BaseDialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import * as agentsAPI from '@/api/admin/agents'
import type { AgentEntry, AgentPayout } from '@/api/admin/agents'
import { lookupUsers } from '@/api/admin/affiliates'
import type { SimpleUser } from '@/api/admin/affiliates'
import { useYcMoney } from '@/composables/useYcMoney'
import { extractApiErrorMessage } from '@/utils/apiError'
import { formatDate, formatDateTime } from '@/utils/format'
import BaseDialog from '@/components/common/BaseDialog.vue'

const { t } = useI18n()
const appStore = useAppStore()
// Shares are kept in balance units; what has to be paid is that in yuan.
const { yuan } = useYcMoney()

const loading = ref(false)
const saving = ref(false)
const agents = ref<AgentEntry[]>([])
const payouts = ref<AgentPayout[]>([])
const pending = computed(() => payouts.value.filter((item) => item.status === 'pending'))
const handled = computed(() => payouts.value.filter((item) => item.status !== 'pending'))

async function load() {
  loading.value = true
  try {
    const [agentList, payoutList] = await Promise.all([agentsAPI.list(), agentsAPI.listPayouts()])
    agents.value = agentList
    payouts.value = payoutList
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.profit.failedToLoad')))
  } finally {
    loading.value = false
  }
}

const formOpen = ref(false)
const editing = ref<AgentEntry | null>(null)
const sharePercent = ref(10)
const query = ref('')
const found = ref<SimpleUser[]>([])
const picked = ref<SimpleUser | null>(null)
let searchSeq = 0

function openForm(agent: AgentEntry | null) {
  editing.value = agent
  sharePercent.value = agent?.share_percent ?? 10
  query.value = ''
  found.value = []
  picked.value = null
  formOpen.value = true
}

async function search() {
  const seq = ++searchSeq
  picked.value = null
  const q = query.value.trim()
  if (!q) {
    found.value = []
    return
  }
  try {
    const users = await lookupUsers(q)
    if (seq === searchSeq) found.value = users
  } catch {
    if (seq === searchSeq) found.value = []
  }
}

async function saveAgent() {
  const userId = editing.value?.user_id ?? picked.value?.id
  if (!userId) {
    appStore.showError(t('admin.profit.agents.form.pickUser'))
    return
  }
  if (!(sharePercent.value > 0 && sharePercent.value <= 100)) {
    appStore.showError(t('admin.profit.agents.form.shareInvalid'))
    return
  }
  saving.value = true
  try {
    await agentsAPI.set(userId, sharePercent.value)
    formOpen.value = false
    appStore.showSuccess(t('admin.profit.agents.saved'))
    await load()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.profit.agents.failed')))
  } finally {
    saving.value = false
  }
}

async function removeAgent(agent: AgentEntry) {
  if (!window.confirm(t('admin.profit.agents.removeConfirm', { email: agent.email || `#${agent.user_id}` }))) return
  try {
    await agentsAPI.remove(agent.user_id)
    appStore.showSuccess(t('admin.profit.agents.saved'))
    await load()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.profit.agents.failed')))
  }
}

const decision = ref<{ payout: AgentPayout; status: 'paid' | 'rejected'; note: string } | null>(null)

function openDecision(payout: AgentPayout, status: 'paid' | 'rejected') {
  decision.value = { payout, status, note: '' }
}

async function decide() {
  const current = decision.value
  if (!current) return
  saving.value = true
  try {
    if (current.status === 'paid') await agentsAPI.markPaid(current.payout.id, current.note)
    else await agentsAPI.reject(current.payout.id, current.note)
    decision.value = null
    appStore.showSuccess(t('admin.profit.agents.saved'))
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.profit.agents.failed')))
  } finally {
    saving.value = false
    await load()
  }
}

onMounted(load)
</script>
