<template>
  <AppLayout>
    <div class="yav">
      <div class="yav-bar">
        <label class="yav-search">
          <Icon name="search" size="md" />
          <input
            v-model="searchQuery"
            type="text"
            :placeholder="t('availableChannels.searchPlaceholder')"
            :aria-label="t('availableChannels.searchPlaceholder')"
          />
        </label>
        <span v-if="!loading && channels.length > 0" class="yav-count" data-test="channels-count">
          {{ t('availableChannels.count', { n: filteredChannels.length }) }}
        </span>
        <button type="button" class="yav-refresh" :disabled="loading" data-test="channels-refresh" @click="loadChannels">
          <Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" />
          {{ t('common.refresh') }}
        </button>
      </div>

      <AvailableChannelsList
        :rows="filteredChannels"
        :loading="loading"
        :user-group-rates="userGroupRates"
        pricing-key-prefix="availableChannels.pricing"
        :no-pricing-label="t('availableChannels.noPricing')"
        :no-models-label="t('availableChannels.noModels')"
        :empty-label="searchQuery.trim() && channels.length > 0 ? t('availableChannels.noMatch') : t('availableChannels.empty')"
      />
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import AvailableChannelsList from '@/components/channels/AvailableChannelsList.vue'
import userChannelsAPI, { type UserAvailableChannel } from '@/api/channels'
import userGroupsAPI from '@/api/groups'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'

const { t } = useI18n()
const appStore = useAppStore()

const channels = ref<UserAvailableChannel[]>([])
const userGroupRates = ref<Record<number, number>>({})
const loading = ref(false)
const searchQuery = ref('')

/**
 * 搜索过滤：
 * - 命中渠道名/描述 → 整个渠道（所有 platforms）都保留
 * - 否则按 platform/group/model 维度在 sections 里过滤，保留有匹配的 section
 * - 所有 sections 都不匹配时，渠道本身被过滤掉
 */
const filteredChannels = computed(() => {
  const q = searchQuery.value.trim().toLowerCase()
  if (!q) return channels.value
  return channels.value
    .map((ch) => {
      const nameHit = ch.name.toLowerCase().includes(q)
      const descHit = (ch.description || '').toLowerCase().includes(q)
      if (nameHit || descHit) return ch
      const matchingSections = ch.platforms.filter(
        (p) =>
          p.platform.toLowerCase().includes(q) ||
          p.groups.some((g) => g.name.toLowerCase().includes(q)) ||
          p.supported_models.some((m) => m.name.toLowerCase().includes(q)),
      )
      if (matchingSections.length === 0) return null
      return { ...ch, platforms: matchingSections }
    })
    .filter((ch): ch is UserAvailableChannel => ch !== null)
})

async function loadChannels() {
  loading.value = true
  try {
    // 渠道列表和用户专属倍率并发拉取。专属倍率失败不阻塞渠道展示——
    // 失败时只是无法渲染专属倍率角标，降级为仅显示默认倍率。
    const [list, rates] = await Promise.all([
      userChannelsAPI.getAvailable(),
      userGroupsAPI.getUserGroupRates().catch((err: unknown) => {
        console.error('Failed to load user group rates:', err)
        return {} as Record<number, number>
      }),
    ])
    channels.value = list
    userGroupRates.value = rates
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, t('common.error')))
  } finally {
    loading.value = false
  }
}

onMounted(loadChannels)
</script>

<style scoped>
.yav { display: grid; gap: 16px; min-width: 0; }
.yav-bar { display: flex; flex-wrap: wrap; align-items: center; gap: 12px; }
.yav-search { display: flex; flex: 1 1 260px; align-items: center; gap: 8px; max-width: 420px; height: 44px; padding: 0 16px; border: 1px solid var(--yc-line); border-radius: 999px; background: var(--yc-paper); color: var(--yc-mut); }
.yav-search:focus-within { border-color: var(--yc-ink); }
.yav-search input { flex: 1; min-width: 0; border: 0; background: transparent; padding: 0; font-size: 14px; color: var(--yc-ink); outline: none; box-shadow: none; }
.yav-search input::placeholder { color: var(--yc-mut); }
.yav-count { font-size: 13px; color: var(--yc-mut); }
.yav-refresh { display: inline-flex; align-items: center; gap: 6px; height: 44px; margin-left: auto; padding: 0 18px; border: 1px solid var(--yc-line); border-radius: 999px; font-size: 13px; font-weight: 600; color: var(--yc-ink); white-space: nowrap; }
.yav-refresh:not(:disabled):hover { background: var(--yc-fill); }
.yav-refresh:disabled { opacity: 0.5; cursor: not-allowed; }
</style>
