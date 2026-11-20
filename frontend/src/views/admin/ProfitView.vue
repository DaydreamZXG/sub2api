<template>
  <AppLayout>
    <div class="mx-auto max-w-6xl space-y-5">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <div class="inline-flex rounded-lg bg-gray-100 p-1 dark:bg-dark-800" role="tablist">
          <button
            v-for="name in TABS"
            :key="name"
            type="button"
            role="tab"
            :aria-selected="tab === name"
            :data-test="`profit-tab-${name}`"
            class="rounded-md px-4 py-1.5 text-sm font-medium transition-colors"
            :class="tab === name ? 'bg-white text-gray-900 shadow-sm dark:bg-dark-600 dark:text-white' : 'text-gray-500 hover:text-gray-900 dark:text-dark-300 dark:hover:text-white'"
            @click="tab = name"
          >
            {{ t(`admin.profit.tabs.${name}`) }}
          </button>
        </div>
        <button class="btn btn-secondary btn-sm" :disabled="rebuilding" :title="t('admin.profit.rebuildHint')" @click="runRebuild">
          {{ rebuilding ? t('admin.profit.rebuilding') : t('admin.profit.rebuild') }}
        </button>
      </div>

      <!-- Report -->
      <template v-if="tab === 'report'">
        <div class="card flex flex-wrap items-center gap-2 p-4">
          <button
            v-for="name in PRESETS"
            :key="name"
            type="button"
            class="rounded-full border px-3 py-1 text-sm transition-colors"
            :class="preset === name ? 'border-primary-500 bg-primary-50 text-primary-700 dark:bg-primary-900/30 dark:text-primary-300' : 'border-gray-200 text-gray-600 hover:border-gray-300 dark:border-dark-600 dark:text-dark-300'"
            @click="applyPreset(name)"
          >
            {{ t(`admin.profit.presets.${name}`) }}
          </button>
          <span class="mx-1 hidden h-5 w-px bg-gray-200 dark:bg-dark-600 sm:block"></span>
          <input v-model="range.from" type="date" class="input w-40" :max="range.to" :aria-label="t('admin.profit.from')" @change="rangeEdited" />
          <span class="text-sm text-gray-400">–</span>
          <input v-model="range.to" type="date" class="input w-40" :min="range.from" :aria-label="t('admin.profit.to')" @change="rangeEdited" />
        </div>

        <div v-if="reportLoading && !report" class="card p-10 text-center text-sm text-gray-500 dark:text-dark-400">{{ t('common.loading') }}</div>
        <template v-else-if="report">
          <div class="grid grid-cols-2 gap-3 lg:grid-cols-5" data-test="profit-totals">
            <div v-for="tile in tiles" :key="tile.key" class="card p-4" :data-test="`profit-tile-${tile.key}`">
              <p class="text-xs text-gray-500 dark:text-dark-400">{{ t(`admin.profit.tiles.${tile.key}`) }}</p>
              <p class="mt-1 text-xl font-semibold tabular-nums" :class="tile.tone">{{ tile.text }}</p>
              <p class="mt-1 text-xs text-gray-400 dark:text-dark-500">{{ t(`admin.profit.tileHints.${tile.key}`) }}</p>
            </div>
          </div>
          <p v-if="report.totals.plan_usage > 0" class="text-sm text-gray-500 dark:text-dark-400" data-test="profit-plan-note">
            {{ t('admin.profit.planUsageNote', { amount: money(report.totals.plan_usage) }) }}
          </p>
          <p v-if="isEmpty" class="card p-8 text-center text-sm text-gray-500 dark:text-dark-400" data-test="profit-empty">
            {{ t('admin.profit.empty') }}
          </p>

          <section v-if="report.packs.length" class="card overflow-hidden" data-test="profit-by-pack">
            <header class="px-4 pt-4">
              <h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('admin.profit.byPack') }}</h2>
              <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('admin.profit.byPackHint') }}</p>
            </header>
            <div class="mt-3 overflow-x-auto">
              <table class="profit-table">
                <thead>
                  <tr>
                    <th>{{ t('admin.profit.cols.pack') }}</th>
                    <th class="num">{{ t('admin.profit.cols.packCost') }}</th>
                    <th class="num">{{ t('admin.profit.cols.income') }}</th>
                    <th class="num">{{ t('admin.profit.cols.planUsage') }}</th>
                    <th class="num">{{ t('admin.profit.cols.profit') }}</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="row in report.packs" :key="row.pack_id">
                    <td>{{ row.name || `#${row.pack_id}` }}</td>
                    <td class="num">{{ money(row.pack_cost) }}</td>
                    <td class="num">{{ money(row.income) }}</td>
                    <td class="num muted">{{ money(row.plan_usage) }}</td>
                    <td class="num" :class="tone(row.profit)">{{ money(row.profit) }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </section>

          <section v-if="report.groups.length" class="card overflow-hidden" data-test="profit-by-group">
            <header class="px-4 pt-4">
              <h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('admin.profit.byGroup') }}</h2>
              <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('admin.profit.byGroupHint') }}</p>
            </header>
            <div class="mt-3 overflow-x-auto">
              <table class="profit-table">
                <thead>
                  <tr>
                    <th>{{ t('admin.profit.cols.group') }}</th>
                    <th class="num">{{ t('admin.profit.cols.income') }}</th>
                    <th class="num">{{ t('admin.profit.cols.planUsage') }}</th>
                    <th class="num">{{ t('admin.profit.cols.packCost') }}</th>
                    <th class="num">{{ t('admin.profit.cols.rateCost') }}</th>
                    <th class="num">{{ t('admin.profit.cols.profit') }}</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="row in report.groups" :key="row.group_id">
                    <td>{{ groupName(row) }}</td>
                    <td class="num">{{ money(row.income) }}</td>
                    <td class="num muted">{{ money(row.plan_usage) }}</td>
                    <td class="num">{{ money(row.pack_cost) }}</td>
                    <td class="num">{{ money(row.rate_cost) }}</td>
                    <td class="num" :class="tone(row.profit)">{{ money(row.profit) }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </section>

          <section v-if="report.months.length > 1" class="card overflow-hidden" data-test="profit-by-month">
            <header class="px-4 pt-4">
              <h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('admin.profit.byMonth') }}</h2>
            </header>
            <div class="mt-3 overflow-x-auto">
              <table class="profit-table">
                <thead>
                  <tr>
                    <th>{{ t('admin.profit.cols.month') }}</th>
                    <th class="num">{{ t('admin.profit.cols.income') }}</th>
                    <th class="num">{{ t('admin.profit.cols.packCost') }}</th>
                    <th class="num">{{ t('admin.profit.cols.rateCost') }}</th>
                    <th class="num">{{ t('admin.profit.cols.profit') }}</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="row in report.months" :key="row.month">
                    <td>{{ row.month }}</td>
                    <td class="num">{{ money(row.income) }}</td>
                    <td class="num">{{ money(row.pack_cost) }}</td>
                    <td class="num">{{ money(row.rate_cost) }}</td>
                    <td class="num" :class="tone(row.profit)">{{ money(row.profit) }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </section>

          <section v-if="report.accounts.length" class="card overflow-hidden" data-test="profit-by-account">
            <header class="px-4 pt-4">
              <h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('admin.profit.byAccount') }}</h2>
              <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('admin.profit.byAccountHint') }}</p>
            </header>
            <div class="mt-3 overflow-x-auto">
              <table class="profit-table">
                <thead>
                  <tr>
                    <th>{{ t('admin.profit.cols.account') }}</th>
                    <th>{{ t('admin.profit.cols.costBy') }}</th>
                    <th class="num">{{ t('admin.profit.cols.requests') }}</th>
                    <th class="num">{{ t('admin.profit.cols.income') }}</th>
                    <th class="num">{{ t('admin.profit.cols.planUsage') }}</th>
                    <th class="num">{{ t('admin.profit.cols.rateCost') }}</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="row in report.accounts" :key="row.account_id">
                    <td>
                      {{ row.name || `#${row.account_id}` }}
                      <span v-if="row.deleted" class="ml-1 text-xs text-gray-400">{{ t('admin.profit.deleted') }}</span>
                    </td>
                    <td class="muted">
                      {{ row.pack_id ? t('admin.profit.costByPack', { name: row.pack_name || `#${row.pack_id}` }) : t('admin.profit.costByRate', { rate: row.rate_multiplier }) }}
                    </td>
                    <td class="num">{{ row.requests.toLocaleString() }}</td>
                    <td class="num">{{ money(row.income) }}</td>
                    <td class="num muted">{{ money(row.plan_usage) }}</td>
                    <td class="num">{{ row.pack_id && !(row.rate_cost > 0) ? '–' : money(row.rate_cost) }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </section>
        </template>
      </template>

      <!-- Cost packs -->
      <template v-else>
        <div class="card flex flex-wrap items-center justify-between gap-3 p-4">
          <p class="max-w-3xl text-sm text-gray-500 dark:text-dark-400">{{ t('admin.profit.packsIntro') }}</p>
          <button class="btn btn-primary" data-test="pack-new" @click="openCreate">{{ t('admin.profit.newPack') }}</button>
        </div>
        <div v-if="packsLoading && !packs.length" class="card p-10 text-center text-sm text-gray-500 dark:text-dark-400">{{ t('common.loading') }}</div>
        <p v-else-if="!packs.length" class="card p-10 text-center text-sm text-gray-500 dark:text-dark-400" data-test="packs-empty">
          {{ t('admin.profit.noPacks') }}
        </p>
        <div v-else class="grid gap-4 lg:grid-cols-2">
          <article v-for="pack in packs" :key="pack.id" class="card space-y-3 p-5" :class="{ 'opacity-70': pack.ended }" data-test="pack-card">
            <div class="flex items-start justify-between gap-3">
              <div class="min-w-0">
                <h3 class="truncate text-base font-semibold text-gray-900 dark:text-white">{{ pack.name }}</h3>
                <p class="mt-0.5 text-sm text-gray-500 dark:text-dark-400">{{ packKind(pack) }}</p>
              </div>
              <span class="shrink-0 rounded-full px-2 py-0.5 text-xs" :class="pack.ended ? 'bg-gray-100 text-gray-500 dark:bg-dark-700 dark:text-dark-300' : 'bg-emerald-50 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300'">
                {{ pack.ended ? t('admin.profit.ended') : t('admin.profit.running') }}
              </span>
            </div>
            <div class="grid grid-cols-3 gap-2 text-sm">
              <div>
                <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('admin.profit.paid') }}</p>
                <p class="font-semibold tabular-nums text-gray-900 dark:text-white">{{ money(pack.paid) }}</p>
              </div>
              <div>
                <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('admin.profit.earned') }}</p>
                <p class="font-semibold tabular-nums text-gray-900 dark:text-white">{{ money(pack.income) }}</p>
              </div>
              <div>
                <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('admin.profit.cols.profit') }}</p>
                <p class="font-semibold tabular-nums" :class="tone(pack.profit)" data-test="pack-profit">{{ money(pack.profit) }}</p>
              </div>
            </div>
            <div>
              <div class="h-2 overflow-hidden rounded-full bg-gray-100 dark:bg-dark-700">
                <div class="h-full rounded-full" :class="pack.profit >= 0 ? 'bg-emerald-500' : 'bg-amber-500'" :style="{ width: `${Math.min(100, payback(pack))}%` }"></div>
              </div>
              <p class="mt-1 text-xs text-gray-500 dark:text-dark-400" data-test="pack-payback">{{ paybackText(pack) }}</p>
            </div>
            <p class="text-sm text-gray-600 dark:text-dark-300">
              <span class="text-gray-500 dark:text-dark-400">{{ t('admin.profit.accounts') }}</span>
              {{ pack.accounts.length ? pack.accounts.map((a) => a.name || `#${a.id}`).join('、') : t('admin.profit.noAccounts') }}
            </p>
            <p class="text-xs text-gray-500 dark:text-dark-400">
              {{ t('admin.profit.since', { date: pack.started_at }) }}
              <template v-if="pack.ended_at"> · {{ t('admin.profit.until', { date: pack.ended_at }) }}</template>
              <template v-else-if="pack.next_cycle_at"> · {{ t('admin.profit.nextCycle', { date: pack.next_cycle_at }) }}</template>
              <template v-if="pack.plan_usage > 0"> · {{ t('admin.profit.packPlanUsage', { amount: money(pack.plan_usage) }) }}</template>
            </p>
            <p v-if="pack.note" class="text-xs text-gray-500 dark:text-dark-400">{{ pack.note }}</p>
            <div class="flex flex-wrap gap-2 border-t border-gray-100 pt-3 dark:border-dark-700">
              <button class="btn btn-secondary btn-sm" @click="openEdit(pack)">{{ t('common.edit') }}</button>
              <button class="btn btn-secondary btn-sm" @click="openPayment(pack)">{{ t('admin.profit.addPayment') }}</button>
              <button v-if="!pack.ended_at" class="btn btn-secondary btn-sm" @click="endPack(pack)">{{ t('admin.profit.endPack') }}</button>
              <button class="btn btn-secondary btn-sm text-red-600 dark:text-red-400" @click="removePack(pack)">{{ t('common.delete') }}</button>
            </div>
          </article>
        </div>
      </template>
    </div>

    <BaseDialog :show="formOpen" :title="editingId ? t('admin.profit.editPack') : t('admin.profit.newPack')" @close="formOpen = false">
      <form class="space-y-4" data-test="pack-form" @submit.prevent="savePack">
        <label class="block text-sm font-medium text-gray-900 dark:text-white">
          {{ t('admin.profit.form.name') }}
          <input v-model="form.name" class="input mt-1" maxlength="100" :placeholder="t('admin.profit.form.namePlaceholder')" data-test="pack-name" />
        </label>
        <div>
          <p class="text-sm font-medium text-gray-900 dark:text-white">{{ t('admin.profit.form.kind') }}</p>
          <div class="mt-1 grid grid-cols-2 gap-2">
            <button
              v-for="kind in KINDS"
              :key="kind"
              type="button"
              class="rounded-lg border p-3 text-left text-sm transition-colors"
              :class="form.kind === kind ? 'border-primary-500 bg-primary-50 dark:bg-primary-900/20' : 'border-gray-200 dark:border-dark-600'"
              :disabled="editingId !== null"
              :data-test="`pack-kind-${kind}`"
              @click="form.kind = kind"
            >
              <span class="font-medium text-gray-900 dark:text-white">{{ t(`admin.profit.form.kinds.${kind}`) }}</span>
              <span class="mt-0.5 block text-xs text-gray-500 dark:text-dark-400">{{ t(`admin.profit.form.kindHints.${kind}`) }}</span>
            </button>
          </div>
        </div>
        <div class="grid gap-3 sm:grid-cols-2">
          <label v-if="form.kind === 'cycle'" class="block text-sm font-medium text-gray-900 dark:text-white">
            {{ t('admin.profit.form.cycleAmount') }}
            <input v-model.number="form.cycle_amount" class="input mt-1" type="number" min="0" step="0.01" data-test="pack-cycle-amount" />
          </label>
          <label v-if="form.kind === 'cycle'" class="block text-sm font-medium text-gray-900 dark:text-white">
            {{ t('admin.profit.form.cycleDays') }}
            <input v-model.number="form.cycle_days" class="input mt-1" type="number" min="1" max="3660" step="1" />
          </label>
          <label v-if="form.kind === 'once' && editingId === null" class="block text-sm font-medium text-gray-900 dark:text-white">
            {{ t('admin.profit.form.amount') }}
            <input v-model.number="form.amount" class="input mt-1" type="number" min="0" step="0.01" data-test="pack-amount" />
          </label>
          <label class="block text-sm font-medium text-gray-900 dark:text-white">
            {{ t('admin.profit.form.startedAt') }}
            <input v-model="form.started_at" class="input mt-1" type="date" />
          </label>
          <label class="block text-sm font-medium text-gray-900 dark:text-white">
            {{ t('admin.profit.form.endedAt') }}
            <input v-model="form.ended_at" class="input mt-1" type="date" :min="form.started_at" />
          </label>
        </div>
        <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('admin.profit.form.datesHint') }}</p>

        <div>
          <p class="text-sm font-medium text-gray-900 dark:text-white">{{ t('admin.profit.form.accounts') }}</p>
          <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('admin.profit.form.accountsHint') }}</p>
          <div v-if="form.accounts.length" class="mt-2 flex flex-wrap gap-1.5">
            <span v-for="acc in form.accounts" :key="acc.id" class="inline-flex items-center gap-1 rounded-full bg-gray-100 px-2.5 py-1 text-xs text-gray-700 dark:bg-dark-700 dark:text-dark-200">
              {{ acc.name || `#${acc.id}` }}
              <button type="button" class="text-gray-400 hover:text-gray-700 dark:hover:text-white" :aria-label="t('common.delete')" @click="dropAccount(acc.id)">×</button>
            </span>
          </div>
          <input v-model="accountSearch" class="input mt-2" :placeholder="t('admin.profit.form.accountSearch')" data-test="pack-account-search" @input="searchAccounts" />
          <ul v-if="accountResults.length" class="mt-1 max-h-44 overflow-y-auto rounded-lg border border-gray-200 text-sm dark:border-dark-600">
            <li v-for="acc in accountResults" :key="acc.id">
              <button
                type="button"
                class="flex w-full items-center justify-between gap-2 px-3 py-2 text-left hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-50 dark:hover:bg-dark-700"
                :disabled="takenBy(acc.id) !== ''"
                @click="pickAccount(acc)"
              >
                <span class="truncate text-gray-900 dark:text-white">{{ acc.name }} <span class="text-xs text-gray-400">{{ acc.platform }}</span></span>
                <span v-if="takenBy(acc.id)" class="shrink-0 text-xs text-gray-400">{{ t('admin.profit.form.taken', { name: takenBy(acc.id) }) }}</span>
              </button>
            </li>
          </ul>
        </div>
        <label class="block text-sm font-medium text-gray-900 dark:text-white">
          {{ t('admin.profit.form.note') }}
          <input v-model="form.note" class="input mt-1" maxlength="500" />
        </label>
        <p v-if="formError" class="text-sm text-red-600 dark:text-red-400" data-test="pack-form-error">{{ formError }}</p>
        <div class="flex justify-end gap-2">
          <button type="button" class="btn btn-secondary" @click="formOpen = false">{{ t('common.cancel') }}</button>
          <button type="submit" class="btn btn-primary" :disabled="saving" data-test="pack-save">{{ t('common.save') }}</button>
        </div>
      </form>
    </BaseDialog>

    <BaseDialog :show="paymentPack !== null" :title="t('admin.profit.addPaymentTitle', { name: paymentPack?.name ?? '' })" @close="paymentPack = null">
      <form v-if="paymentPack" class="space-y-4" @submit.prevent="savePayment">
        <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('admin.profit.addPaymentHint') }}</p>
        <ul v-if="paymentPack.payments.length" class="divide-y divide-gray-100 rounded-lg border border-gray-200 text-sm dark:divide-dark-700 dark:border-dark-600">
          <li v-for="pay in paymentPack.payments" :key="pay.id" class="flex items-center justify-between gap-3 px-3 py-2">
            <span class="text-gray-600 dark:text-dark-300">{{ pay.paid_at }}<template v-if="pay.note"> · {{ pay.note }}</template></span>
            <span class="flex items-center gap-3">
              <span class="tabular-nums text-gray-900 dark:text-white">{{ money(pay.amount) }}</span>
              <button type="button" class="text-xs text-red-600 dark:text-red-400" @click="removePayment(pay.id)">{{ t('common.delete') }}</button>
            </span>
          </li>
        </ul>
        <div class="grid gap-3 sm:grid-cols-2">
          <label class="block text-sm font-medium text-gray-900 dark:text-white">
            {{ t('admin.profit.form.amount') }}
            <input v-model.number="payment.amount" class="input mt-1" type="number" min="0" step="0.01" />
          </label>
          <label class="block text-sm font-medium text-gray-900 dark:text-white">
            {{ t('admin.profit.form.paidAt') }}
            <input v-model="payment.paid_at" class="input mt-1" type="date" />
          </label>
        </div>
        <label class="block text-sm font-medium text-gray-900 dark:text-white">
          {{ t('admin.profit.form.note') }}
          <input v-model="payment.note" class="input mt-1" maxlength="200" />
        </label>
        <div class="flex justify-end gap-2">
          <button type="button" class="btn btn-secondary" @click="paymentPack = null">{{ t('common.cancel') }}</button>
          <button type="submit" class="btn btn-primary" :disabled="saving || !(payment.amount > 0) || !payment.paid_at">{{ t('admin.profit.addPayment') }}</button>
        </div>
      </form>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import * as profitAPI from '@/api/admin/profit'
import type { CostPack, ProfitReport } from '@/api/admin/profit'
import { list as listAccounts } from '@/api/admin/accounts'
import { extractApiErrorMessage } from '@/utils/apiError'
import AppLayout from '@/components/layout/AppLayout.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'

const TABS = ['report', 'packs'] as const
const PRESETS = ['thisMonth', 'lastMonth', 'last30', 'last90'] as const
const KINDS = ['once', 'cycle'] as const
type Preset = (typeof PRESETS)[number]
type PickedAccount = { id: number; name: string; platform: string }

const { t } = useI18n()
const appStore = useAppStore()

const tab = ref<(typeof TABS)[number]>('report')

function day(date: Date): string {
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}`
}

function presetRange(name: Preset): { from: string; to: string } {
  const now = new Date()
  if (name === 'thisMonth') return { from: day(new Date(now.getFullYear(), now.getMonth(), 1)), to: day(now) }
  if (name === 'lastMonth') {
    return { from: day(new Date(now.getFullYear(), now.getMonth() - 1, 1)), to: day(new Date(now.getFullYear(), now.getMonth(), 0)) }
  }
  const days = name === 'last30' ? 29 : 89
  return { from: day(new Date(now.getFullYear(), now.getMonth(), now.getDate() - days)), to: day(now) }
}

/** Yuan with two decimals; the minus sign goes before the currency sign. */
function money(amount: number): string {
  const abs = Math.abs(amount).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
  return `${amount < -0.005 ? '-' : ''}¥${abs}`
}

function tone(amount: number): string {
  if (amount < -0.005) return 'text-red-600 dark:text-red-400'
  return amount > 0.005 ? 'text-emerald-600 dark:text-emerald-400' : 'text-gray-900 dark:text-white'
}

// ---- report ----
const preset = ref<Preset | null>('thisMonth')
const range = reactive(presetRange('thisMonth'))
const report = ref<ProfitReport | null>(null)
const reportLoading = ref(false)
let reportSeq = 0

async function loadReport() {
  if (!range.from || !range.to || range.from > range.to) return
  const seq = ++reportSeq
  reportLoading.value = true
  try {
    const data = await profitAPI.getReport(range.from, range.to)
    if (seq === reportSeq) report.value = data
  } catch (error) {
    if (seq === reportSeq) appStore.showError(extractApiErrorMessage(error, t('admin.profit.failedToLoad')))
  } finally {
    if (seq === reportSeq) reportLoading.value = false
  }
}

function applyPreset(name: Preset) {
  preset.value = name
  Object.assign(range, presetRange(name))
  void loadReport()
}

function rangeEdited() {
  preset.value = null
  void loadReport()
}

const tiles = computed(() => {
  const totals = report.value?.totals
  if (!totals) return []
  const cost = totals.pack_cost + totals.rate_cost
  const margin = totals.income > 0 ? `${((totals.profit / totals.income) * 100).toFixed(1)}%` : '–'
  const plain = 'text-gray-900 dark:text-white'
  return [
    { key: 'income', text: money(totals.income), tone: plain },
    { key: 'packCost', text: money(totals.pack_cost), tone: plain },
    { key: 'rateCost', text: money(totals.rate_cost), tone: plain },
    { key: 'profit', text: money(totals.profit), tone: tone(totals.profit) },
    { key: 'margin', text: cost === 0 && totals.income === 0 ? '–' : margin, tone: tone(totals.profit) }
  ]
})

const isEmpty = computed(() => {
  const data = report.value
  return !!data && !data.accounts.length && !data.packs.length && !data.groups.length
})

function groupName(row: { group_id: number; name: string }): string {
  if (row.group_id === -1) return t('admin.profit.unallocated')
  if (row.group_id === 0) return t('admin.profit.noGroup')
  return row.name || `#${row.group_id}`
}

const rebuilding = ref(false)
async function runRebuild() {
  rebuilding.value = true
  try {
    await profitAPI.rebuild()
    appStore.showSuccess(t('admin.profit.rebuilt'))
    await Promise.all([loadReport(), loadPacks()])
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.profit.failedToRebuild')))
  } finally {
    rebuilding.value = false
  }
}

// ---- cost packs ----
const packs = ref<CostPack[]>([])
const packsLoading = ref(false)
let packsLoaded = false

async function loadPacks() {
  packsLoading.value = true
  try {
    packs.value = await profitAPI.listPacks()
    packsLoaded = true
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.profit.failedToLoad')))
  } finally {
    packsLoading.value = false
  }
}

watch(tab, (name) => {
  if (name === 'packs' && !packsLoaded) void loadPacks()
  if (name === 'report') void loadReport()
})

function packKind(pack: CostPack): string {
  return pack.cycle_days > 0
    ? t('admin.profit.kindCycle', { days: pack.cycle_days, amount: money(pack.cycle_amount) })
    : t('admin.profit.kindOnce')
}

/** Share of what was paid that has been earned back, in percent. */
function payback(pack: CostPack): number {
  if (!(pack.paid > 0)) return pack.income > 0 ? 100 : 0
  return (pack.income / pack.paid) * 100
}

function paybackText(pack: CostPack): string {
  if (!(pack.paid > 0)) return t('admin.profit.paybackNoCost')
  return pack.profit >= 0
    ? t('admin.profit.paybackDone', { amount: money(pack.profit) })
    : t('admin.profit.paybackSoFar', { percent: Math.floor(payback(pack)), amount: money(-pack.profit) })
}

const formOpen = ref(false)
const editingId = ref<number | null>(null)
const saving = ref(false)
const formError = ref('')
const form = reactive({
  name: '',
  note: '',
  kind: 'once' as (typeof KINDS)[number],
  cycle_days: 30,
  cycle_amount: 0,
  amount: 0,
  started_at: day(new Date()),
  ended_at: '',
  accounts: [] as PickedAccount[]
})
const accountSearch = ref('')
const accountResults = ref<PickedAccount[]>([])
let accountSeq = 0

function resetAccountSearch() {
  accountSearch.value = ''
  accountResults.value = []
}

function openCreate() {
  editingId.value = null
  formError.value = ''
  Object.assign(form, {
    name: '', note: '', kind: 'once', cycle_days: 30, cycle_amount: 0, amount: 0, started_at: day(new Date()), ended_at: '', accounts: []
  })
  resetAccountSearch()
  formOpen.value = true
}

function openEdit(pack: CostPack) {
  editingId.value = pack.id
  formError.value = ''
  Object.assign(form, {
    name: pack.name,
    note: pack.note,
    kind: pack.cycle_days > 0 ? 'cycle' : 'once',
    cycle_days: pack.cycle_days > 0 ? pack.cycle_days : 30,
    cycle_amount: pack.cycle_amount,
    amount: 0,
    started_at: pack.started_at,
    ended_at: pack.ended_at ?? '',
    accounts: pack.accounts.map((a) => ({ id: a.id, name: a.name, platform: a.platform }))
  })
  resetAccountSearch()
  formOpen.value = true
}

async function searchAccounts() {
  const seq = ++accountSeq
  const search = accountSearch.value.trim()
  if (!search) {
    accountResults.value = []
    return
  }
  try {
    const page = await listAccounts(1, 20, { search, lite: '1' })
    if (seq !== accountSeq) return
    accountResults.value = page.items
      .filter((a) => !form.accounts.some((picked) => picked.id === a.id))
      .map((a) => ({ id: a.id, name: a.name, platform: a.platform }))
  } catch {
    if (seq === accountSeq) accountResults.value = []
  }
}

/** Name of another pack that already holds the account, or '' when it is free. */
function takenBy(accountId: number): string {
  const holder = packs.value.find((p) => p.id !== editingId.value && p.accounts.some((a) => a.id === accountId))
  return holder ? holder.name : ''
}

function pickAccount(acc: PickedAccount) {
  form.accounts.push(acc)
  accountResults.value = accountResults.value.filter((a) => a.id !== acc.id)
}

function dropAccount(id: number) {
  form.accounts = form.accounts.filter((a) => a.id !== id)
}

function validateForm(): string {
  if (!form.name.trim()) return t('admin.profit.form.errors.name')
  if (!form.started_at) return t('admin.profit.form.errors.startedAt')
  if (form.ended_at && form.ended_at < form.started_at) return t('admin.profit.form.errors.endedAt')
  if (form.kind === 'cycle') {
    if (!(form.cycle_amount > 0)) return t('admin.profit.form.errors.amount')
    if (!Number.isInteger(form.cycle_days) || form.cycle_days < 1 || form.cycle_days > 3660) return t('admin.profit.form.errors.cycleDays')
  } else if (editingId.value === null && !(form.amount > 0)) {
    return t('admin.profit.form.errors.amount')
  }
  return ''
}

async function savePack() {
  formError.value = validateForm()
  if (formError.value) return
  const input = {
    name: form.name.trim(),
    note: form.note.trim(),
    cycle_days: form.kind === 'cycle' ? form.cycle_days : 0,
    cycle_amount: form.kind === 'cycle' ? form.cycle_amount : 0,
    started_at: form.started_at,
    ended_at: form.ended_at || null,
    account_ids: form.accounts.map((a) => a.id),
    amount: form.kind === 'once' ? form.amount : 0
  }
  saving.value = true
  try {
    if (editingId.value === null) await profitAPI.createPack(input)
    else await profitAPI.updatePack(editingId.value, input)
    formOpen.value = false
    appStore.showSuccess(t('admin.profit.saved'))
    await loadPacks()
  } catch (error) {
    formError.value = extractApiErrorMessage(error, t('admin.profit.failedToSave'))
  } finally {
    saving.value = false
  }
}

async function endPack(pack: CostPack) {
  if (!window.confirm(t('admin.profit.endConfirm', { name: pack.name }))) return
  try {
    await profitAPI.updatePack(pack.id, {
      name: pack.name,
      note: pack.note,
      cycle_days: pack.cycle_days,
      cycle_amount: pack.cycle_amount,
      started_at: pack.started_at,
      ended_at: day(new Date()) < pack.started_at ? pack.started_at : day(new Date()),
      account_ids: pack.accounts.map((a) => a.id),
      amount: 0
    })
    await loadPacks()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.profit.failedToSave')))
  }
}

async function removePack(pack: CostPack) {
  if (!window.confirm(t('admin.profit.deleteConfirm', { name: pack.name }))) return
  try {
    await profitAPI.deletePack(pack.id)
    await loadPacks()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.profit.failedToSave')))
  }
}

const paymentPack = ref<CostPack | null>(null)
const payment = reactive({ amount: 0, paid_at: day(new Date()), note: '' })

function openPayment(pack: CostPack) {
  Object.assign(payment, { amount: 0, paid_at: day(new Date()), note: '' })
  paymentPack.value = pack
}

async function refreshPaymentPack() {
  await loadPacks()
  const id = paymentPack.value?.id
  paymentPack.value = packs.value.find((p) => p.id === id) ?? null
}

async function savePayment() {
  if (!paymentPack.value) return
  saving.value = true
  try {
    await profitAPI.addPayment(paymentPack.value.id, payment.amount, payment.paid_at, payment.note.trim())
    paymentPack.value = null
    appStore.showSuccess(t('admin.profit.saved'))
    await loadPacks()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.profit.failedToSave')))
  } finally {
    saving.value = false
  }
}

async function removePayment(paymentId: number) {
  if (!paymentPack.value) return
  try {
    await profitAPI.deletePayment(paymentPack.value.id, paymentId)
    await refreshPaymentPack()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.profit.failedToSave')))
  }
}

onMounted(() => {
  void loadReport()
  // The report names packs and the form needs to know which accounts are taken.
  void loadPacks()
})
</script>

<style scoped>
.profit-table {
  width: 100%;
  min-width: 560px;
  font-size: 0.875rem;
}

.profit-table th {
  padding: 0.5rem 1rem;
  text-align: left;
  font-size: 0.75rem;
  font-weight: 500;
  color: rgb(107 114 128);
  white-space: nowrap;
}

.profit-table td {
  padding: 0.625rem 1rem;
  border-top: 1px solid rgb(243 244 246);
  color: rgb(17 24 39);
}

.profit-table .num {
  text-align: right;
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
}

.profit-table td.muted {
  color: rgb(107 114 128);
}

.dark .profit-table th,
.dark .profit-table td.muted {
  color: rgb(148 163 184);
}

.dark .profit-table td {
  border-top-color: rgb(51 65 85 / 0.6);
  color: rgb(241 245 249);
}
</style>
