<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <div class="flex flex-wrap items-center gap-3">
          <Select v-model="status" :options="statusOptions" class="w-40" @change="reload" />
          <div class="flex flex-1 justify-end">
            <button class="btn btn-secondary" :disabled="loading" :title="t('common.refresh')" @click="load">
              <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
            </button>
          </div>
        </div>
      </template>

      <template #table>
        <DataTable :columns="columns" :data="tickets" :loading="loading">
          <template #cell-subject="{ row }">
            <button class="max-w-md truncate text-left text-sm font-medium text-gray-900 hover:underline dark:text-white" @click="open(row)">
              {{ row.subject }}
            </button>
          </template>
          <template #cell-category="{ value }">
            <span class="text-sm text-gray-600 dark:text-gray-300">{{ t(`admin.tickets.category.${value}`) }}</span>
          </template>
          <template #cell-status="{ value }">
            <span :class="['badge', STATUS_BADGES[value as TicketStatus]]">{{ t(`admin.tickets.status.${value}`) }}</span>
          </template>
          <template #cell-last_message_at="{ value }">
            <span class="text-sm text-gray-500 dark:text-dark-400">{{ formatDateTime(value) }}</span>
          </template>
          <template #cell-actions="{ row }">
            <button class="btn btn-secondary btn-sm" @click="open(row)">{{ t('admin.tickets.view') }}</button>
          </template>
        </DataTable>
      </template>

      <template #pagination>
        <Pagination
          v-if="total > 0"
          :page="page"
          :total="total"
          :page-size="pageSize"
          @update:page="changePage"
          @update:pageSize="changePageSize"
        />
      </template>
    </TablePageLayout>

    <BaseDialog :show="current !== null" :title="current?.subject ?? ''" width="wide" @close="current = null">
      <div v-if="current" class="space-y-4">
        <p class="text-sm text-gray-500 dark:text-dark-400">
          {{ current.user_email || `#${current.user_id}` }} · {{ t(`admin.tickets.category.${current.category}`) }} ·
          <span :class="['badge', STATUS_BADGES[current.status]]">{{ t(`admin.tickets.status.${current.status}`) }}</span>
        </p>
        <div class="max-h-[46vh] space-y-3 overflow-y-auto">
          <div
            v-for="message in current.messages"
            :key="message.id"
            class="max-w-[86%] rounded-2xl px-4 py-3 text-sm"
            :class="message.author_type === 'user' ? 'bg-gray-100 dark:bg-dark-700' : 'ml-auto bg-primary-100 dark:bg-primary-500/15'"
          >
            <p class="mb-1 text-xs text-gray-500 dark:text-dark-400">
              {{ t(`admin.tickets.author.${message.author_type}`) }} · {{ formatDateTime(message.created_at) }}
            </p>
            <p class="whitespace-pre-wrap break-words text-gray-900 dark:text-gray-100">{{ message.content }}</p>
          </div>
        </div>
        <textarea
          v-model="replyText"
          class="input"
          rows="4"
          :maxlength="TICKET_CONTENT_MAX"
          :placeholder="t('admin.tickets.replyPlaceholder')"
        ></textarea>
      </div>
      <template #footer>
        <div v-if="current" class="flex flex-wrap justify-end gap-3">
          <button class="btn btn-secondary" :disabled="saving" @click="setStatus(current.status === 'closed' ? 'open' : 'closed')">
            {{ t(current.status === 'closed' ? 'admin.tickets.reopen' : 'admin.tickets.close') }}
          </button>
          <button class="btn btn-primary" :disabled="saving || !replyText.trim()" @click="sendReply">
            {{ t('admin.tickets.reply') }}
          </button>
        </div>
      </template>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { getPersistedPageSize } from '@/composables/usePersistedPageSize'
import ticketsAPI, { type AdminTicket, type AdminTicketDetail } from '@/api/admin/tickets'
import { TICKET_CONTENT_MAX, type TicketStatus } from '@/api/tickets'
import { extractApiErrorMessage } from '@/utils/apiError'
import { formatDateTime } from '@/utils/format'
import type { Column } from '@/components/common/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'

const STATUS_BADGES: Record<TicketStatus, string> = {
  open: 'badge-warning',
  answered: 'badge-success',
  closed: 'badge-gray'
}

const { t } = useI18n()
const appStore = useAppStore()

const tickets = ref<AdminTicket[]>([])
const loading = ref(false)
const status = ref<TicketStatus | ''>('open')
const page = ref(1)
const pageSize = ref(getPersistedPageSize())
const total = ref(0)

const current = ref<AdminTicketDetail | null>(null)
const replyText = ref('')
const saving = ref(false)

const statusOptions = computed(() => [
  { value: '', label: t('admin.tickets.allStatus') },
  ...(['open', 'answered', 'closed'] as const).map((value) => ({ value, label: t(`admin.tickets.status.${value}`) }))
])

const columns = computed<Column[]>(() => [
  { key: 'subject', label: t('admin.tickets.columns.subject') },
  { key: 'user_email', label: t('admin.tickets.columns.user') },
  { key: 'category', label: t('admin.tickets.columns.category') },
  { key: 'status', label: t('admin.tickets.columns.status') },
  { key: 'last_message_at', label: t('admin.tickets.columns.lastMessage') },
  { key: 'actions', label: t('common.actions') }
])

async function load() {
  loading.value = true
  try {
    const result = await ticketsAPI.list(page.value, pageSize.value, status.value)
    tickets.value = result.items
    total.value = result.total
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.tickets.failedToLoad')))
  } finally {
    loading.value = false
  }
}

function reload() {
  page.value = 1
  void load()
}

function changePage(next: number) {
  page.value = next
  void load()
}

function changePageSize(next: number) {
  pageSize.value = next
  reload()
}

async function open(ticket: AdminTicket) {
  try {
    current.value = await ticketsAPI.getById(ticket.id)
    replyText.value = ''
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.tickets.failedToLoad')))
  }
}

// Runs a change on the open ticket, then refreshes both the dialog and the list behind it.
async function save(change: (id: number) => Promise<unknown>) {
  if (!current.value || saving.value) return
  const id = current.value.id
  saving.value = true
  try {
    await change(id)
    replyText.value = ''
    current.value = await ticketsAPI.getById(id)
    await load()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.tickets.failedToSave')))
  } finally {
    saving.value = false
  }
}

function sendReply() {
  return save((id) => ticketsAPI.reply(id, replyText.value))
}

function setStatus(next: TicketStatus) {
  return save((id) => ticketsAPI.updateStatus(id, next))
}

onMounted(load)
</script>
