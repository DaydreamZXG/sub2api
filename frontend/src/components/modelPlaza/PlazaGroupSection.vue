<template>
  <section class="yc-surface pzg" data-test="plaza-line">
    <header class="pzg-head">
      <div class="pzg-title">
        <PlatformIcon :platform="group.platform as GroupPlatform" size="sm" />
        <h2 class="pzg-name">{{ group.name }}</h2>
        <span v-if="group.app_only" class="yc-badge">{{ t('yc.prices.appOnly') }}</span>
        <span v-if="group.subscription_type === 'subscription'" class="yc-badge">{{ t('modelPlaza.badges.subscription') }}</span>
        <span v-if="group.is_exclusive" class="yc-badge">{{ t('modelPlaza.badges.exclusive') }}</span>
      </div>
      <p v-if="group.description" class="pzg-desc">{{ group.description }}</p>
      <p v-if="peakNote" class="pzg-note"><Icon name="clock" size="xs" />{{ peakNote }}</p>
      <p v-if="longContextNote" class="pzg-note"><Icon name="infoCircle" size="xs" />{{ longContextNote }}</p>
    </header>

    <PlazaModelPricingTable
      v-if="group.models.length > 0"
      :models="group.models"
      :platform="group.platform"
      :rate-multiplier="group.rate_multiplier"
      :user-rate-multiplier="group.user_rate_multiplier ?? null"
      :image-rate-independent="group.image_rate_independent"
      :image-rate-multiplier="group.image_rate_multiplier"
      :video-rate-independent="group.video_rate_independent"
      :video-rate-multiplier="group.video_rate_multiplier"
      :peak-window="peakWindow"
      :peak-rate-multiplier="group.peak_rate_multiplier"
    />
    <p v-else class="pzg-empty">{{ t('modelPlaza.detail.noModels') }}</p>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import PlatformIcon from '@/components/common/PlatformIcon.vue'
import PlazaModelPricingTable from './PlazaModelPricingTable.vue'
import type { ModelPlazaGroup } from '@/api/modelPlaza'
import type { GroupPlatform } from '@/types'
import { priceChangeLabel } from '@/utils/ycPriceWords'
import { hasPeakRate, serverTimezoneLabel } from '@/utils/peak-rate'
import { useAppStore } from '@/stores/app'

const props = defineProps<{
  group: ModelPlazaGroup
}>()

const { t } = useI18n()
const appStore = useAppStore()

/** 高峰窗口描述(含倍率与服务器时区标注);分组未启用高峰为空串。 */
const peakWindow = computed(() => {
  if (!hasPeakRate(props.group)) return ''
  const zone = serverTimezoneLabel(appStore.cachedPublicSettings?.server_utc_offset)
  const hours = `${props.group.peak_start}-${props.group.peak_end}`
  return zone ? `${hours} (${zone})` : hours
})

const peakNote = computed(() => {
  if (!peakWindow.value) return ''
  return t('modelPlaza.detail.peakNote', {
    window: peakWindow.value,
    change: priceChangeLabel(t, props.group.peak_rate_multiplier)
  })
})

/**
 * 分组关闭了长上下文阶梯、但组内有模型官方带阶梯时提示:实付列只展示基础档,
 * 官方阶梯仅供参考。字段缺失(旧后端)不提示。
 */
const longContextNote = computed(() => {
  if (props.group.long_context_pricing_enabled !== false) return ''
  const hasOfficialLadder = props.group.models.some(
    (m) => (m.official_pricing?.intervals?.length ?? 0) > 1
  )
  return hasOfficialLadder ? t('modelPlaza.detail.longContextDisabledNote') : ''
})
</script>

<style scoped>
.pzg { overflow: hidden; }
.pzg-head { padding: 20px 24px 16px; }
.pzg-title { display: flex; flex-wrap: wrap; align-items: center; gap: 8px; }
.pzg-name { margin: 0 4px 0 0; font-size: 17px; font-weight: 700; color: var(--yc-ink); }
.pzg-desc { margin: 8px 0 0; font-size: 14px; color: var(--yc-mut); }
.pzg-note { display: flex; align-items: center; gap: 6px; margin: 6px 0 0; font-size: 12.5px; color: var(--yc-mut); }
.pzg-empty { margin: 0; padding: 20px 24px; border-top: 1px solid var(--yc-line); text-align: center; font-size: 14px; color: var(--yc-dim); }
@media (max-width: 640px) { .pzg-head { padding: 16px 16px 12px; } }
</style>
