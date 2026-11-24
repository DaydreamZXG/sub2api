<template>
  <div class="yav-list">
    <div v-if="loading" class="yav-state" data-test="channels-loading">
      <Icon name="refresh" size="lg" class="animate-spin" />
    </div>
    <div v-else-if="rows.length === 0" class="yav-state" data-test="channels-empty">
      <Icon name="inbox" size="xl" />
      <p>{{ emptyLabel }}</p>
    </div>
    <template v-else>
      <article
        v-for="(channel, chIdx) in rows"
        :key="`${channel.name}-${chIdx}`"
        class="yav-card"
        data-test="channel-card"
      >
        <header class="yav-head">
          <h3>{{ channel.name }}</h3>
          <p v-if="channel.description">{{ channel.description }}</p>
        </header>

        <section
          v-for="section in channel.platforms"
          :key="`${channel.name}-${section.platform}`"
          class="yav-platform"
          data-test="channel-platform"
        >
          <p class="yav-platform-name">
            <template v-if="sectionFamilies(section).length > 0">
              <span v-for="f in sectionFamilies(section)" :key="f.key" class="yav-family" data-test="channel-family">
                <ModelIcon :model="f.sample" size="16px" />
                {{ f.label }}
              </span>
              <span class="yav-tag">{{ t('availableChannels.protocol', { platform: platformLabel(section.platform) }) }}</span>
            </template>
            <template v-else>
              <PlatformIcon :platform="section.platform as GroupPlatform" size="sm" />
              {{ platformLabel(section.platform) }}
            </template>
          </p>

          <div class="yav-cols">
            <div class="yav-col">
              <p class="yav-label">{{ t('availableChannels.columns.groups') }}</p>
              <ul v-if="section.groups.length > 0" class="yav-lines">
                <li v-for="g in orderedGroups(section)" :key="g.id" data-test="channel-line">
                  <span class="yav-line-name">{{ g.name }}</span>
                  <span
                    class="yav-tag"
                    :class="{ 'yav-tag-own': g.is_exclusive }"
                    :title="t(g.is_exclusive ? 'availableChannels.exclusiveTooltip' : 'availableChannels.publicTooltip')"
                  >
                    {{ t(g.is_exclusive ? 'availableChannels.exclusive' : 'availableChannels.public') }}
                  </span>
                  <span v-if="g.subscription_type === 'subscription'" class="yav-tag">
                    {{ t('availableChannels.subscription') }}
                  </span>
                  <span class="yav-price-k">{{ t('availableChannels.vsChannelPrice') }}</span>
                  <span
                    class="yav-price"
                    :class="{ 'yav-price-good': effectiveRate(g) < 1 }"
                    :title="rawRateTitle(g)"
                    data-test="channel-line-price"
                  >
                    {{ priceLabel(g) }}
                  </span>
                  <span v-if="hasPeakRate(g)" class="yav-peak" :title="peakTitle(g)" data-test="channel-line-peak">
                    {{ peakLabel(g) }}
                  </span>
                </li>
              </ul>
              <p v-else class="yav-none" data-test="channel-no-lines">-</p>
            </div>

            <div class="yav-col">
              <p class="yav-label">{{ t('availableChannels.columns.supportedModels') }}</p>
              <div v-if="section.supported_models.length > 0" class="yav-models">
                <SupportedModelChip
                  v-for="m in section.supported_models"
                  :key="`${section.platform}-${m.name}`"
                  :model="m"
                  :pricing-key-prefix="pricingKeyPrefix"
                  :no-pricing-label="noPricingLabel"
                  :show-platform="false"
                  :platform-hint="section.platform"
                />
              </div>
              <p v-else class="yav-none">{{ noModelsLabel }}</p>
            </div>
          </div>
        </section>
      </article>
    </template>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import PlatformIcon from '@/components/common/PlatformIcon.vue'
import ModelIcon from '@/components/common/ModelIcon.vue'
import SupportedModelChip from './SupportedModelChip.vue'
import type { UserAvailableChannel, UserAvailableGroup, UserChannelPlatformSection } from '@/api/channels'
import type { GroupPlatform } from '@/types'
import { platformLabel } from '@/utils/platformColors'
import { useAppStore } from '@/stores/app'
import { hasPeakRate, serverTimezoneLabel } from '@/utils/peak-rate'
import { priceChangeLabel } from '@/utils/ycPriceWords'
import { familiesOf } from '@/utils/modelFamily'

const props = defineProps<{
  rows: UserAvailableChannel[]
  loading: boolean
  pricingKeyPrefix: string
  noPricingLabel: string
  noModelsLabel: string
  emptyLabel: string
  /** The user's own multiplier per line (group id → multiplier); it replaces the line's default. */
  userGroupRates: Record<number, number>
}>()

const { t } = useI18n()
const appStore = useAppStore()

/** Lines granted to this user come first, then the ones open to everyone. */
function orderedGroups(section: UserChannelPlatformSection): UserAvailableGroup[] {
  return [...section.groups.filter((g) => g.is_exclusive), ...section.groups.filter((g) => !g.is_exclusive)]
}

/** The makers of the models on this section, which is what users choose by; the API protocol is a side note. */
function sectionFamilies(section: UserChannelPlatformSection) {
  return familiesOf(section.supported_models.map((m) => m.name))
}

function effectiveRate(group: UserAvailableGroup): number {
  return props.userGroupRates[group.id] ?? group.rate_multiplier
}

function priceLabel(group: UserAvailableGroup): string {
  return priceChangeLabel(t, effectiveRate(group)) || t('userSubscriptions.standardPrice')
}

/** The raw multiplier stays available on hover for those who want it. */
function rawRateTitle(group: UserAvailableGroup): string {
  return t('modelPlaza.table.rawRate', { rate: effectiveRate(group) })
}

function peakLabel(group: UserAvailableGroup): string {
  const zone = serverTimezoneLabel(appStore.cachedPublicSettings?.server_utc_offset)
  const hours = `${group.peak_start}-${group.peak_end}`
  return t('modelPlaza.detail.peakNote', {
    window: zone ? `${hours} (${zone})` : hours,
    change: priceChangeLabel(t, group.peak_rate_multiplier)
  })
}

function peakTitle(group: UserAvailableGroup): string {
  return (
    t('modelPlaza.table.rawRate', { rate: group.peak_rate_multiplier }) + t('common.peakRateImageNote')
  )
}
</script>

<style scoped>
.yav-list { display: grid; gap: 16px; min-width: 0; }
.yav-state { display: grid; justify-items: center; gap: 10px; padding: 56px 20px; border: 1px dashed var(--yc-line); border-radius: 22px; color: var(--yc-mut); font-size: 14px; }

.yav-card { min-width: 0; border: 1px solid var(--yc-line); border-radius: 22px; background: var(--yc-paper); overflow: hidden; }
.yav-head { padding: 20px 24px 16px; }
.yav-head h3 { font-size: 18px; font-weight: 720; color: var(--yc-ink); overflow-wrap: anywhere; }
.yav-head p { margin-top: 4px; font-size: 13px; line-height: 1.6; color: var(--yc-mut); overflow-wrap: anywhere; }

.yav-platform { padding: 16px 24px 20px; border-top: 1px solid var(--yc-line); }
.yav-platform-name { display: flex; flex-wrap: wrap; align-items: center; gap: 7px 12px; font-size: 14px; font-weight: 650; color: var(--yc-ink); }
.yav-family { display: inline-flex; align-items: center; gap: 6px; }
.yav-price-k { margin-left: auto; font-size: 12px; color: var(--yc-mut); }
.yav-cols { display: grid; grid-template-columns: minmax(0, 5fr) minmax(0, 7fr); gap: 16px 32px; margin-top: 12px; }
.yav-col { min-width: 0; }
.yav-label { margin-bottom: 8px; font-size: 12px; color: var(--yc-mut); }
.yav-none { font-size: 13px; color: var(--yc-mut); }

.yav-lines { display: grid; gap: 8px; }
.yav-lines li { display: flex; flex-wrap: wrap; align-items: center; gap: 6px 8px; min-width: 0; padding: 10px 12px; border-radius: 14px; background: var(--yc-fill); }
.yav-line-name { font-size: 14px; font-weight: 600; color: var(--yc-ink); overflow-wrap: anywhere; }
.yav-tag { padding: 1px 8px; border: 1px solid var(--yc-line); border-radius: 999px; font-size: 11px; color: var(--yc-mut); white-space: nowrap; }
.yav-tag-own { border-color: transparent; background: var(--yc-on); color: #fff; }
.yav-price { margin-left: auto; padding: 1px 9px; border: 1px solid var(--yc-line); border-radius: 999px; background: var(--yc-paper); color: var(--yc-ink); font-size: 12px; font-weight: 650; white-space: nowrap; cursor: help; }
/* Only a price below the standard one gets the lime. */
.yav-price-good { border-color: transparent; background: var(--yc-lime); color: var(--yc-on-lime); }
.yav-peak { flex-basis: 100%; font-size: 12px; color: var(--yc-mut); cursor: help; }

.yav-models { display: flex; flex-wrap: wrap; gap: 6px; min-width: 0; }

@media (max-width: 860px) {
  .yav-cols { grid-template-columns: minmax(0, 1fr); }
}
@media (max-width: 640px) {
  .yav-head { padding: 16px 16px 12px; }
  .yav-platform { padding: 14px 16px 16px; }
}
</style>
