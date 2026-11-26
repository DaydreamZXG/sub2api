<template>
  <BaseDialog :show="show" :title="t('admin.channels.officialBook.title')" width="extra-wide" @close="emit('close')">
    <p class="mb-4 text-sm text-gray-500 dark:text-dark-400">{{ t('admin.channels.officialBook.hint') }}</p>
    <div v-if="loading" class="py-8 text-center text-sm text-gray-400">…</div>
    <div v-else class="overflow-x-auto">
      <table class="w-full min-w-[860px] text-sm">
        <thead>
          <tr class="text-left text-xs text-gray-500 dark:text-dark-400">
            <th class="pb-1 pr-2" rowspan="2">
              {{ t('admin.channels.officialBook.models') }}
              <div class="font-normal text-gray-400">{{ t('admin.channels.officialBook.modelsHint') }}</div>
            </th>
            <th class="pb-1 text-center" colspan="3">{{ t('admin.channels.officialBook.usd') }}</th>
            <th rowspan="2"></th>
          </tr>
          <tr class="text-xs text-gray-400">
            <th v-for="f in fieldLabels" :key="f" class="px-1 pb-2 font-normal">{{ t(f) }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(row, i) in rows" :key="row.key" class="align-top">
            <td class="py-1 pr-2"><input v-model="row.models" type="text" class="input text-sm" placeholder="glm-5.3, GLM-5.3" /></td>
            <td v-for="field in priceFields" :key="field" class="px-1 py-1">
              <input v-model="row[field]" type="number" step="any" min="0" class="input w-24 text-sm" />
            </td>
            <td class="py-1 pl-2">
              <button type="button" class="btn btn-secondary btn-sm" @click="rows.splice(i, 1)">{{ t('admin.channels.officialBook.removeRow') }}</button>
            </td>
          </tr>
          <tr v-if="rows.length === 0">
            <td colspan="5" class="py-6 text-center text-gray-400">{{ t('admin.channels.officialBook.empty') }}</td>
          </tr>
        </tbody>
      </table>
      <button type="button" class="btn btn-secondary btn-sm mt-3" @click="addRow">{{ t('admin.channels.officialBook.addRow') }}</button>
    </div>
    <template #footer>
      <button type="button" class="btn btn-secondary" @click="emit('close')">{{ t('common.cancel') }}</button>
      <button type="button" class="btn btn-primary" :disabled="loading || saving || loadFailed" @click="save">{{ t('common.save') }}</button>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'
import { getOfficialPriceBook, updateOfficialPriceBook, type OfficialPriceEntry } from '@/api/admin/settings'

const props = defineProps<{ show: boolean }>()
const emit = defineEmits<{ close: [] }>()
const { t } = useI18n()
const appStore = useAppStore()

const priceFields = ['usd_input', 'usd_output', 'usd_cache_read'] as const
type PriceField = (typeof priceFields)[number]
const fieldLabels = ['admin.channels.officialBook.input', 'admin.channels.officialBook.output', 'admin.channels.officialBook.cacheRead']

// Yuan prices are no longer edited here but are kept, since saving replaces the whole book.
const keptFields = ['cny_input', 'cny_output', 'cny_cache_read'] as const
type Row = { key: number; models: string } & Record<PriceField | (typeof keptFields)[number], number | string | null>

const rows = ref<Row[]>([])
const loading = ref(false)
const saving = ref(false)
// Saving replaces the whole book, so never save rows that were not loaded.
const loadFailed = ref(false)
let nextKey = 0

function toRow(e: OfficialPriceEntry): Row {
  return { key: nextKey++, models: e.models.join(', '), ...Object.fromEntries([...priceFields, ...keptFields].map((f) => [f, e[f] ?? null])) } as Row
}

function addRow() {
  rows.value.push(toRow({ models: [], cny_input: null, cny_output: null, cny_cache_read: null, usd_input: null, usd_output: null, usd_cache_read: null }))
}

/** An emptied number input gives ''; that is "unknown", not zero. */
function price(v: number | string | null): number | null {
  if (v === null || v === '') return null
  const n = Number(v)
  return Number.isFinite(n) ? n : null
}

function toEntry(r: Row): OfficialPriceEntry {
  const entry = { models: r.models.split(/[,，]/).map((m) => m.trim()).filter(Boolean) } as OfficialPriceEntry
  for (const f of [...priceFields, ...keptFields]) entry[f] = price(r[f])
  return entry
}

async function load() {
  loading.value = true
  loadFailed.value = false
  try {
    rows.value = (await getOfficialPriceBook()).map(toRow)
  } catch (err) {
    loadFailed.value = true
    appStore.showError(extractApiErrorMessage(err, t('admin.channels.officialBook.loadFailed')))
  } finally {
    loading.value = false
  }
}

async function save() {
  if (loadFailed.value) return
  saving.value = true
  try {
    rows.value = (await updateOfficialPriceBook(rows.value.map(toEntry))).map(toRow)
    appStore.showSuccess(t('admin.channels.officialBook.saved'))
    emit('close')
  } catch (err) {
    appStore.showError(extractApiErrorMessage(err, t('common.error')))
  } finally {
    saving.value = false
  }
}

watch(() => props.show, (open) => { if (open) void load() }, { immediate: true })
</script>
