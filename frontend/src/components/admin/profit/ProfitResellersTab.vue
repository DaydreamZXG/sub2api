<template>
  <div class="space-y-5" data-test="resellers-tab">
    <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('admin.profit.resellers.intro') }}</p>

    <section class="card p-5">
      <h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('admin.profit.resellers.addTitle') }}</h2>
      <div class="mt-3 flex flex-wrap items-start gap-3">
        <div class="min-w-0 flex-1">
          <input
            v-model="query"
            class="input"
            :placeholder="t('admin.profit.resellers.userPlaceholder')"
            data-test="resellers-search"
            @input="search"
          />
          <ul v-if="found.length" class="mt-2 max-h-40 divide-y divide-gray-100 overflow-auto rounded-lg border border-gray-200 text-sm dark:divide-dark-700 dark:border-dark-600">
            <li v-for="user in found" :key="user.id">
              <button
                type="button"
                class="flex w-full items-center justify-between px-3 py-2 text-left hover:bg-gray-50 dark:hover:bg-dark-700"
                :class="picked?.id === user.id ? 'bg-primary-50 dark:bg-primary-900/30' : ''"
                data-test="resellers-option"
                @click="picked = user"
              >
                <span class="text-gray-900 dark:text-white">{{ user.email }}</span>
                <span class="text-xs text-gray-400 dark:text-dark-500">{{ user.username }}</span>
              </button>
            </li>
          </ul>
        </div>
        <button type="button" class="btn btn-primary" :disabled="!picked || saving" data-test="resellers-add" @click="add">
          {{ t('admin.profit.resellers.add') }}
        </button>
      </div>
    </section>

    <section class="card p-5" data-test="resellers-portal">
      <h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('admin.profit.resellers.portal.title') }}</h2>
      <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('admin.profit.resellers.portal.intro') }}</p>
      <form class="mt-4 grid gap-4 sm:grid-cols-2" @submit.prevent="savePortal">
        <label class="block text-sm font-medium text-gray-900 dark:text-white">
          {{ t('admin.profit.resellers.portal.domain') }}
          <input v-model.trim="portal.domain" class="input mt-1" placeholder="keys.example.com" data-test="portal-domain" />
          <span class="mt-1 block text-xs font-normal text-gray-500 dark:text-dark-400">{{ t('admin.profit.resellers.portal.domainHint') }}</span>
        </label>
        <label class="block text-sm font-medium text-gray-900 dark:text-white">
          {{ t('admin.profit.resellers.portal.name') }}
          <input v-model.trim="portal.name" class="input mt-1" maxlength="60" data-test="portal-name" />
          <span class="mt-1 block text-xs font-normal text-gray-500 dark:text-dark-400">{{ t('admin.profit.resellers.portal.nameHint') }}</span>
        </label>
        <label class="block text-sm font-medium text-gray-900 dark:text-white">
          {{ t('admin.profit.resellers.portal.contact') }}
          <input v-model.trim="portal.contact" class="input mt-1" maxlength="300" data-test="portal-contact" />
        </label>
        <label class="block text-sm font-medium text-gray-900 dark:text-white">
          {{ t('admin.profit.resellers.portal.logo') }}
          <input v-model.trim="portal.logo" class="input mt-1" placeholder="https://…" maxlength="500" data-test="portal-logo" />
        </label>
        <div class="sm:col-span-2">
          <button type="submit" class="btn btn-primary" :disabled="savingPortal" data-test="portal-save">{{ t('common.save') }}</button>
        </div>
      </form>
      <ol class="mt-4 list-decimal space-y-1 pl-5 text-xs text-gray-500 dark:text-dark-400">
        <li>{{ t('admin.profit.resellers.portal.step1') }}</li>
        <li>{{ t('admin.profit.resellers.portal.step2') }}</li>
        <li>{{ t('admin.profit.resellers.portal.step3') }}</li>
      </ol>
    </section>

    <section class="card p-5" data-test="resellers-list">
      <h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('admin.profit.resellers.listTitle') }}</h2>
      <div v-if="loading && resellers.length === 0" class="mt-4 text-sm text-gray-500 dark:text-dark-400">{{ t('common.loading') }}</div>
      <p v-else-if="resellers.length === 0" class="mt-4 text-sm text-gray-400 dark:text-dark-500">{{ t('admin.profit.resellers.listEmpty') }}</p>
      <div v-else class="mt-4 overflow-x-auto">
        <table class="w-full min-w-[640px] text-sm">
          <thead>
            <tr class="text-left text-xs text-gray-500 dark:text-dark-400">
              <th class="pb-2 font-medium">{{ t('admin.profit.resellers.columns.reseller') }}</th>
              <th class="pb-2 text-right font-medium">{{ t('admin.profit.resellers.columns.balance') }}</th>
              <th class="pb-2 text-right font-medium">{{ t('admin.profit.resellers.columns.keys') }}</th>
              <th class="pb-2 text-right font-medium">{{ t('admin.profit.resellers.columns.clientUsed') }}</th>
              <th class="pb-2 text-right font-medium">{{ t('admin.profit.resellers.columns.cost') }}</th>
              <th class="pb-2"></th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
            <tr v-for="reseller in resellers" :key="reseller.user_id" data-test="resellers-row">
              <td class="py-2.5">
                <p class="font-medium text-gray-900 dark:text-white">{{ reseller.email || `#${reseller.user_id}` }}</p>
                <p class="text-xs text-gray-400 dark:text-dark-500">{{ t('admin.profit.resellers.since', { date: formatDateOnly(reseller.since) }) }}</p>
              </td>
              <td class="py-2.5 text-right tabular-nums text-gray-900 dark:text-white">{{ dollars(reseller.balance) }}</td>
              <td class="py-2.5 text-right tabular-nums text-gray-600 dark:text-dark-300">{{ reseller.keys }}</td>
              <td class="py-2.5 text-right tabular-nums text-gray-600 dark:text-dark-300">{{ dollars(reseller.client_used) }}</td>
              <td class="py-2.5 text-right tabular-nums text-gray-900 dark:text-white">{{ dollars(reseller.cost) }}</td>
              <td class="py-2.5 text-right">
                <button type="button" class="text-xs text-red-600 dark:text-red-400" data-test="resellers-remove" @click="remove(reseller)">
                  {{ t('admin.profit.resellers.remove') }}
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <p class="mt-4 text-xs text-gray-400 dark:text-dark-500">{{ t('admin.profit.resellers.note') }}</p>
    </section>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import * as resellersAPI from '@/api/admin/resellers'
import type { ResellerEntry, ResellerPortal } from '@/api/admin/resellers'
import { lookupUsers } from '@/api/admin/affiliates'
import type { SimpleUser } from '@/api/admin/affiliates'
import { useYcMoney } from '@/composables/useYcMoney'
import { extractApiErrorMessage } from '@/utils/apiError'
import { formatDateOnly } from '@/utils/format'

const { t } = useI18n()
const appStore = useAppStore()
const { dollars } = useYcMoney()

const loading = ref(false)
const saving = ref(false)
const resellers = ref<ResellerEntry[]>([])
const query = ref('')
const found = ref<SimpleUser[]>([])
const picked = ref<SimpleUser | null>(null)
let searchSeq = 0
const portal = ref<ResellerPortal>({ domain: '', name: '', contact: '', logo: '' })
const savingPortal = ref(false)

async function load() {
  loading.value = true
  try {
    resellers.value = await resellersAPI.list()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.profit.failedToLoad')))
  } finally {
    loading.value = false
  }
}

async function loadPortal() {
  try {
    portal.value = await resellersAPI.getPortal()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.profit.failedToLoad')))
  }
}

async function savePortal() {
  if (savingPortal.value) return
  // The server reduces an address to its host; a domain needs a name to show to clients.
  if (portal.value.domain && !portal.value.name) {
    appStore.showError(t('admin.profit.resellers.portal.needName'))
    return
  }
  savingPortal.value = true
  try {
    portal.value = await resellersAPI.setPortal(portal.value)
    appStore.showSuccess(t('admin.profit.resellers.portal.saved'))
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.profit.resellers.portal.invalid')))
  } finally {
    savingPortal.value = false
  }
}

async function search() {
  const seq = ++searchSeq
  picked.value = null
  const keyword = query.value.trim()
  if (!keyword) {
    found.value = []
    return
  }
  try {
    const users = await lookupUsers(keyword)
    if (seq === searchSeq) found.value = users
  } catch {
    if (seq === searchSeq) found.value = []
  }
}

async function add() {
  if (!picked.value || saving.value) return
  saving.value = true
  try {
    await resellersAPI.add(picked.value.id)
    query.value = ''
    found.value = []
    picked.value = null
    await load()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.profit.resellers.failed')))
  } finally {
    saving.value = false
  }
}

async function remove(reseller: ResellerEntry) {
  if (!window.confirm(t('admin.profit.resellers.removeConfirm', { email: reseller.email || `#${reseller.user_id}` }))) return
  try {
    await resellersAPI.remove(reseller.user_id)
    await load()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.profit.resellers.failed')))
  }
}

onMounted(() => {
  void load()
  void loadPortal()
})
</script>
