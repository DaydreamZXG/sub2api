<template>
  <AppLayout>
    <div class="card mx-auto max-w-3xl space-y-6 p-6">
      <div class="flex items-start justify-between gap-6">
        <div>
          <h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('admin.garden.enable') }}</h2>
          <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('admin.garden.enableHint') }}</p>
        </div>
        <Toggle v-model="form.enabled" />
      </div>

      <div class="space-y-3 border-t border-gray-200 pt-5 dark:border-dark-700">
        <label class="block text-sm font-medium text-gray-900 dark:text-white" for="garden-ripen-hours">
          {{ t('admin.garden.ripenHours') }}
        </label>
        <input
          id="garden-ripen-hours"
          v-model.number="form.ripen_hours"
          class="input max-w-xs"
          type="number"
          min="0.1"
          :max="MAX_HOURS"
          step="1"
        />
        <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('admin.garden.ripenHoursHint') }}</p>
        <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('admin.garden.rules') }}</p>
        <p v-if="invalid" class="text-sm text-red-600 dark:text-red-400">{{ t('admin.garden.errorHours', { max: MAX_HOURS }) }}</p>
      </div>

      <div class="flex justify-end border-t border-gray-200 pt-5 dark:border-dark-700">
        <button class="btn btn-primary" :disabled="loading || saving || invalid" @click="save">
          {{ t('common.save') }}
        </button>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { getGardenSettings, updateGardenSettings, type GardenSettings } from '@/api/garden'
import { extractApiErrorMessage } from '@/utils/apiError'
import AppLayout from '@/components/layout/AppLayout.vue'
import Toggle from '@/components/common/Toggle.vue'

const MAX_HOURS = 720

const { t } = useI18n()
const appStore = useAppStore()

const form = reactive<GardenSettings>({ enabled: false, ripen_hours: 24 })
const loading = ref(true)
const saving = ref(false)

const invalid = computed(
  () => !Number.isFinite(form.ripen_hours) || form.ripen_hours <= 0 || form.ripen_hours > MAX_HOURS
)

async function save() {
  saving.value = true
  try {
    Object.assign(form, await updateGardenSettings({ ...form }))
    appStore.showSuccess(t('admin.garden.saved'))
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.garden.failedToSave')))
  } finally {
    saving.value = false
  }
}

onMounted(async () => {
  try {
    Object.assign(form, await getGardenSettings())
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.garden.failedToLoad')))
  } finally {
    loading.value = false
  }
})
</script>
