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
        <p v-if="hoursInvalid" class="text-sm text-red-600 dark:text-red-400">{{ t('admin.garden.errorHours', { max: MAX_HOURS }) }}</p>
      </div>

      <div class="space-y-4 border-t border-gray-200 pt-5 dark:border-dark-700">
        <div>
          <h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('admin.garden.social') }}</h2>
          <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('admin.garden.socialHint') }}</p>
        </div>
        <div class="grid gap-4 sm:grid-cols-2">
          <label v-for="field in SOCIAL_FIELDS" :key="field.key" class="block text-sm font-medium text-gray-900 dark:text-white">
            {{ t(`admin.garden.fields.${field.key}`) }}
            <input v-model.number="form[field.key]" class="input mt-2" type="number" :min="field.min" :max="field.max" :step="field.step" />
          </label>
        </div>
        <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('admin.garden.dogHint') }}</p>
        <p v-if="socialInvalid" class="text-sm text-red-600 dark:text-red-400">{{ t('admin.garden.errorSocial') }}</p>
      </div>

      <div class="space-y-3 border-t border-gray-200 pt-5 dark:border-dark-700">
        <label class="block text-sm font-medium text-gray-900 dark:text-white" for="garden-invite-crop">
          {{ t('admin.garden.inviteCrop') }}
        </label>
        <input id="garden-invite-crop" v-model.number="form.invite_crop_amount" class="input max-w-xs" type="number" min="0" :max="MAX_GIFT" step="1" />
        <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('admin.garden.inviteCropHint') }}</p>
        <p v-if="inviteInvalid" class="text-sm text-red-600 dark:text-red-400">{{ t('admin.garden.errorGift', { max: MAX_GIFT }) }}</p>
      </div>

      <div class="flex justify-end border-t border-gray-200 pt-5 dark:border-dark-700">
        <button class="btn btn-primary" :disabled="loading || saving || invalid" @click="save">
          {{ t('common.save') }}
        </button>
      </div>
    </div>
    <div class="card mx-auto mt-6 max-w-3xl space-y-4 p-6">
      <div>
        <h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('admin.garden.gift') }}</h2>
        <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('admin.garden.giftHint') }}</p>
      </div>
      <div class="grid gap-4 sm:grid-cols-[minmax(0,1fr)_160px_auto] sm:items-end">
        <label class="block text-sm font-medium text-gray-900 dark:text-white">
          {{ t('admin.garden.giftEmail') }}
          <input v-model="gift.email" class="input mt-2" type="email" autocomplete="off" />
        </label>
        <label class="block text-sm font-medium text-gray-900 dark:text-white">
          {{ t('admin.garden.giftAmount') }}
          <input v-model.number="gift.amount" class="input mt-2" type="number" min="0" :max="MAX_GIFT" step="1" />
        </label>
        <button class="btn btn-primary" :disabled="gifting || giftInvalid" @click="sendGift">{{ t('admin.garden.giftSend') }}</button>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { getGardenSettings, giftCrop, updateGardenSettings, type GardenSettings } from '@/api/garden'
import { extractApiErrorMessage } from '@/utils/apiError'
import AppLayout from '@/components/layout/AppLayout.vue'
import Toggle from '@/components/common/Toggle.vue'

const MAX_HOURS = 720
const MAX_GIFT = 10000

const { t } = useI18n()
const appStore = useAppStore()

type SocialKey = Exclude<keyof GardenSettings, 'enabled' | 'ripen_hours' | 'invite_crop_amount'>
// Each rule neighbours play by, with the range the server accepts.
const SOCIAL_FIELDS: { key: SocialKey; min: number; max: number; step: number; whole?: boolean }[] = [
  { key: 'protect_hours', min: 0, max: MAX_HOURS, step: 1 },
  { key: 'steal_percent', min: 0, max: 100, step: 1 },
  { key: 'max_stolen_percent', min: 0, max: 100, step: 1 },
  { key: 'daily_steal_limit', min: 0, max: 1000, step: 1, whole: true },
  { key: 'water_minutes', min: 0, max: 1440, step: 5 },
  { key: 'max_waters_per_crop', min: 0, max: 100, step: 1, whole: true },
  { key: 'neighbours_per_day', min: 0, max: 50, step: 1, whole: true },
  { key: 'dog_min_percent', min: 0, max: 100, step: 1 },
  { key: 'dog_max_percent', min: 0, max: 100, step: 1 }
]

const form = reactive<GardenSettings>({
  enabled: false, ripen_hours: 24, protect_hours: 2, steal_percent: 3, max_stolen_percent: 50, daily_steal_limit: 10,
  water_minutes: 60, max_waters_per_crop: 5, neighbours_per_day: 10, dog_min_percent: 10, dog_max_percent: 50,
  invite_crop_amount: 0
})
const loading = ref(true)
const saving = ref(false)

const hoursInvalid = computed(
  () => !Number.isFinite(form.ripen_hours) || form.ripen_hours <= 0 || form.ripen_hours > MAX_HOURS
)
const socialInvalid = computed(
  () =>
    SOCIAL_FIELDS.some(
      ({ key, min, max, whole }) =>
        !Number.isFinite(form[key]) || form[key] < min || form[key] > max || (whole && !Number.isInteger(form[key]))
    ) ||
    form.steal_percent > form.max_stolen_percent ||
    form.dog_min_percent > form.dog_max_percent
)
const inviteInvalid = computed(
  () => !Number.isFinite(form.invite_crop_amount) || form.invite_crop_amount < 0 || form.invite_crop_amount > MAX_GIFT
)
const invalid = computed(() => hoursInvalid.value || socialInvalid.value || inviteInvalid.value)

// A crop handed to one user, for a post, a promotion or any other campaign.
const gift = reactive({ email: '', amount: 5 })
const gifting = ref(false)
const giftInvalid = computed(
  () => !gift.email.trim() || !Number.isFinite(gift.amount) || gift.amount <= 0 || gift.amount > MAX_GIFT
)

async function sendGift() {
  gifting.value = true
  try {
    await giftCrop(gift.email.trim(), gift.amount)
    appStore.showSuccess(t('admin.garden.giftDone', { email: gift.email.trim() }))
    gift.email = ''
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.garden.giftFailed')))
  } finally {
    gifting.value = false
  }
}

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
