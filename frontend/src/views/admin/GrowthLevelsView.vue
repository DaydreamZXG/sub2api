<template>
  <AppLayout>
    <div class="card mx-auto max-w-3xl space-y-6 p-6">
      <div class="flex items-start justify-between gap-6">
        <div>
          <h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('admin.growth.enable') }}</h2>
          <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('admin.growth.enableHint') }}</p>
        </div>
        <Toggle v-model="form.enabled" />
      </div>

      <div class="space-y-3 border-t border-gray-200 pt-5 dark:border-dark-700">
        <div class="grid grid-cols-[1fr_1fr_1fr_2.5rem] gap-3 text-xs font-medium text-gray-500 dark:text-dark-400">
          <span>{{ t('admin.growth.name') }}</span>
          <span>{{ t('admin.growth.threshold') }}</span>
          <span>{{ t('admin.growth.discount') }}</span>
          <span></span>
        </div>
        <div v-for="(level, index) in form.levels" :key="index" class="grid grid-cols-[1fr_1fr_1fr_2.5rem] items-center gap-3">
          <input v-model.trim="level.name" class="input" maxlength="20" :aria-label="t('admin.growth.name')" />
          <input
            v-model.number="level.threshold"
            class="input"
            type="number"
            min="0"
            step="1"
            :disabled="index === 0"
            :aria-label="t('admin.growth.threshold')"
          />
          <input
            v-model.number="level.percent"
            class="input"
            type="number"
            min="0"
            max="90"
            step="0.1"
            :aria-label="t('admin.growth.discount')"
          />
          <button
            class="btn btn-secondary btn-sm"
            :disabled="index === 0"
            :title="t('common.delete')"
            :aria-label="t('common.delete')"
            @click="form.levels.splice(index, 1)"
          >
            <Icon name="trash" size="sm" />
          </button>
        </div>
        <button class="btn btn-secondary btn-sm" :disabled="form.levels.length >= MAX_LEVELS" @click="addLevel">
          {{ t('admin.growth.add') }}
        </button>
        <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('admin.growth.rules') }}</p>
        <p v-if="problem" class="text-sm text-red-600 dark:text-red-400">{{ t(problem) }}</p>
      </div>

      <div class="flex justify-end border-t border-gray-200 pt-5 dark:border-dark-700">
        <button class="btn btn-primary" :disabled="loading || saving || problem !== ''" @click="save">
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
import { getGrowthLevels, updateGrowthLevels, type GrowthLevels } from '@/api/growth'
import { extractApiErrorMessage } from '@/utils/apiError'
import AppLayout from '@/components/layout/AppLayout.vue'
import Toggle from '@/components/common/Toggle.vue'
import Icon from '@/components/icons/Icon.vue'

const MAX_LEVELS = 10

const { t } = useI18n()
const appStore = useAppStore()

// The form edits the discount as a percentage; the API stores a share between 0 and 0.9.
const form = reactive({ enabled: false, levels: [] as { name: string; threshold: number; percent: number }[] })
const loading = ref(true)
const saving = ref(false)

const problem = computed(() => {
  for (const [index, level] of form.levels.entries()) {
    if (!level.name) return 'admin.growth.errorName'
    if (!Number.isFinite(level.threshold) || (index > 0 && level.threshold <= form.levels[index - 1].threshold)) {
      return 'admin.growth.errorThreshold'
    }
    if (!Number.isFinite(level.percent) || level.percent < 0 || level.percent > 90) return 'admin.growth.errorDiscount'
  }
  return ''
})

function fill(saved: GrowthLevels) {
  form.enabled = saved.enabled
  form.levels = saved.levels.map((level) => ({
    name: level.name,
    threshold: level.threshold,
    percent: Math.round(level.discount * 1000) / 10
  }))
}

function addLevel() {
  const last = form.levels[form.levels.length - 1]
  form.levels.push({ name: '', threshold: last ? Math.max(100, last.threshold * 2) : 0, percent: last?.percent ?? 0 })
}

async function save() {
  saving.value = true
  try {
    fill(
      await updateGrowthLevels({
        enabled: form.enabled,
        levels: form.levels.map((level) => ({
          name: level.name,
          threshold: level.threshold,
          discount: Math.round(level.percent * 10) / 1000
        }))
      })
    )
    appStore.showSuccess(t('admin.growth.saved'))
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.growth.failedToSave')))
  } finally {
    saving.value = false
  }
}

onMounted(async () => {
  try {
    fill(await getGrowthLevels())
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.growth.failedToLoad')))
  } finally {
    loading.value = false
  }
})
</script>
