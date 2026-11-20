<template>
  <BaseDialog :show="!!agent" :title="t('admin.profit.agents.users.title', { agent: agent?.email || `#${agent?.user_id}` })" @close="emit('close')">
    <div class="space-y-4" data-test="agent-users">
      <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('admin.profit.agents.users.hint') }}</p>

      <div>
        <div class="flex gap-2">
          <input
            v-model="query"
            class="input flex-1"
            :placeholder="t('admin.profit.agents.form.userPlaceholder')"
            data-test="agent-users-search"
            @input="search"
          />
          <button type="button" class="btn btn-primary" :disabled="!picked || saving" data-test="agent-users-add" @click="attach">
            {{ t('admin.profit.agents.users.add') }}
          </button>
        </div>
        <ul v-if="found.length" class="mt-2 max-h-40 divide-y divide-gray-100 overflow-auto rounded-lg border border-gray-200 text-sm dark:divide-dark-700 dark:border-dark-600">
          <li v-for="user in found" :key="user.id">
            <button
              type="button"
              class="flex w-full items-center justify-between px-3 py-2 text-left hover:bg-gray-50 dark:hover:bg-dark-700"
              :class="picked?.id === user.id ? 'bg-primary-50 dark:bg-primary-900/30' : ''"
              data-test="agent-users-option"
              @click="picked = user"
            >
              <span class="text-gray-900 dark:text-white">{{ user.email }}</span>
              <span class="text-xs text-gray-400 dark:text-dark-500">{{ user.username }}</span>
            </button>
          </li>
        </ul>
      </div>

      <p v-if="!loading && users.length === 0" class="text-sm text-gray-500 dark:text-dark-400">{{ t('admin.profit.agents.users.empty') }}</p>
      <ul v-else class="max-h-72 divide-y divide-gray-100 overflow-auto dark:divide-dark-700">
        <li v-for="user in users" :key="user.user_id" class="flex items-center justify-between gap-3 py-2.5" data-test="agent-users-row">
          <div class="min-w-0">
            <div class="truncate text-sm text-gray-900 dark:text-white">{{ user.email || `#${user.user_id}` }}</div>
            <div class="text-xs text-gray-500 dark:text-dark-400">
              {{ t(user.attached ? 'admin.profit.agents.users.attachedAt' : 'admin.profit.agents.users.joinedAt', { date: formatDate(user.joined_at) }) }}
              · {{ t('admin.profit.agents.users.figures', { spent: yuan(user.spent), share: yuan(user.share) }) }}
            </div>
          </div>
          <button type="button" class="shrink-0 text-xs text-red-600 dark:text-red-400" :disabled="saving" data-test="agent-users-remove" @click="detach(user)">
            {{ t('admin.profit.agents.users.remove') }}
          </button>
        </li>
      </ul>
    </div>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import * as agentsAPI from '@/api/admin/agents'
import type { AgentEntry, AgentUser } from '@/api/admin/agents'
import { lookupUsers } from '@/api/admin/affiliates'
import type { SimpleUser } from '@/api/admin/affiliates'
import { useYcMoney } from '@/composables/useYcMoney'
import { extractApiErrorMessage } from '@/utils/apiError'
import { formatDate } from '@/utils/format'
import BaseDialog from '@/components/common/BaseDialog.vue'

const props = defineProps<{ agent: AgentEntry | null }>()
const emit = defineEmits<{ (event: 'close'): void; (event: 'changed'): void }>()

const { t } = useI18n()
const appStore = useAppStore()
const { yuan } = useYcMoney()

const users = ref<AgentUser[]>([])
const loading = ref(false)
const saving = ref(false)
const query = ref('')
const found = ref<SimpleUser[]>([])
const picked = ref<SimpleUser | null>(null)
let searchSeq = 0

async function load() {
  if (!props.agent) return
  loading.value = true
  try {
    users.value = await agentsAPI.listUsers(props.agent.user_id)
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.profit.agents.failed')))
  } finally {
    loading.value = false
  }
}

watch(
  () => props.agent?.user_id,
  () => {
    users.value = []
    query.value = ''
    found.value = []
    picked.value = null
    void load()
  },
  { immediate: true }
)

async function search() {
  const seq = ++searchSeq
  picked.value = null
  const q = query.value.trim()
  if (!q) {
    found.value = []
    return
  }
  try {
    const result = await lookupUsers(q)
    if (seq === searchSeq) found.value = result
  } catch {
    if (seq === searchSeq) found.value = []
  }
}

async function change(apply: () => Promise<void>) {
  saving.value = true
  try {
    await apply()
    appStore.showSuccess(t('admin.profit.agents.saved'))
    await load()
    emit('changed')
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.profit.agents.failed')))
  } finally {
    saving.value = false
  }
}

async function attach() {
  const agent = props.agent
  const user = picked.value
  if (!agent || !user) return
  if (!window.confirm(t('admin.profit.agents.users.addConfirm', { user: user.email, agent: agent.email }))) return
  await change(async () => {
    await agentsAPI.attachUser(agent.user_id, user.id)
    query.value = ''
    found.value = []
    picked.value = null
  })
}

async function detach(user: AgentUser) {
  const agent = props.agent
  if (!agent) return
  if (!window.confirm(t('admin.profit.agents.users.removeConfirm', { user: user.email || `#${user.user_id}` }))) return
  await change(() => agentsAPI.detachUser(agent.user_id, user.user_id))
}
</script>
