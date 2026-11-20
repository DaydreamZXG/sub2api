<template>
  <div class="pzf">
    <div class="pzf-row">
      <span class="pzf-label">{{ t('modelPlaza.filters.platformLabel') }}</span>
      <div class="pzf-chips">
        <button
          v-for="p in ['all', ...platforms]"
          :key="`platform-${p}`"
          type="button"
          class="pzf-chip"
          :class="{ 'is-on': platform === p }"
          :aria-pressed="platform === p"
          :disabled="p !== 'all' && !platformEnabled(p)"
          @click="$emit('update:platform', p)"
        >
          <PlatformIcon v-if="p !== 'all'" :platform="p as GroupPlatform" size="xs" />
          {{ p === 'all' ? t('modelPlaza.filters.all') : platformLabel(p) }}
        </button>
      </div>
    </div>

    <div class="pzf-row">
      <span class="pzf-label">{{ t('modelPlaza.filters.groupLabel') }}</span>
      <div class="pzf-chips">
        <button
          type="button"
          class="pzf-chip"
          :class="{ 'is-on': groupId === 'all' }"
          :aria-pressed="groupId === 'all'"
          @click="$emit('update:groupId', 'all')"
        >
          {{ t('modelPlaza.filters.all') }}
        </button>
        <button
          v-for="g in groups"
          :key="`group-${g.id}`"
          type="button"
          class="pzf-chip"
          :class="{ 'is-on': groupId === g.id }"
          :aria-pressed="groupId === g.id"
          :disabled="!groupEnabled(g)"
          @click="$emit('update:groupId', g.id)"
        >
          {{ g.name }}
        </button>
      </div>
    </div>

    <div class="pzf-search">
      <Icon name="search" size="sm" class="pzf-search-icon" />
      <input
        :value="search"
        type="search"
        class="yc-input"
        :placeholder="t('modelPlaza.filters.searchPlaceholder')"
        :aria-label="t('modelPlaza.filters.searchPlaceholder')"
        @input="$emit('update:search', ($event.target as HTMLInputElement).value)"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import PlatformIcon from '@/components/common/PlatformIcon.vue'
import { platformLabel } from '@/utils/platformColors'
import type { GroupPlatform } from '@/types'

const props = defineProps<{
  /** Platforms present in the data, deduplicated and sorted. */
  platforms: string[]
  /** Every line; a choice that would leave nothing to show is greyed out. */
  groups: Array<{ id: number; name: string; platform: string }>
  platform: string
  groupId: number | 'all'
  search: string
}>()

defineEmits<{
  'update:platform': [value: string]
  'update:groupId': [value: number | 'all']
  'update:search': [value: string]
}>()

const { t } = useI18n()

function platformEnabled(p: string): boolean {
  return props.groups.some((g) => g.platform === p && (props.groupId === 'all' || g.id === props.groupId))
}

function groupEnabled(g: { platform: string }): boolean {
  return props.platform === 'all' || g.platform === props.platform
}
</script>

<style scoped>
.pzf { display: grid; gap: 12px; }
.pzf-row { display: flex; align-items: flex-start; gap: 14px; }
.pzf-label { flex: none; width: 36px; padding-top: 8px; font-size: 12.5px; font-weight: 650; color: var(--yc-mut); }
.pzf-chips { display: flex; flex-wrap: wrap; gap: 8px; min-width: 0; }
.pzf-chip {
  display: inline-flex; align-items: center; gap: 6px; height: 34px; padding: 0 14px;
  border: 1px solid var(--yc-line); border-radius: 999px; background: var(--yc-paper);
  font-size: 13.5px; font-weight: 600; color: var(--yc-ink);
  transition: background-color 0.15s ease, border-color 0.15s ease, color 0.15s ease;
}
.pzf-chip:hover:not(:disabled):not(.is-on) { border-color: var(--yc-ink); }
.pzf-chip.is-on { border-color: var(--yc-ink); background: var(--yc-ink); color: var(--yc-paper); }
.pzf-chip:disabled { opacity: 0.4; cursor: not-allowed; }
.pzf-chip:focus-visible { outline: 2px solid var(--yc-lime); outline-offset: 2px; }
.pzf-search { position: relative; max-width: 360px; }
.pzf-search-icon { position: absolute; left: 14px; top: 50%; transform: translateY(-50%); color: var(--yc-dim); pointer-events: none; }
.pzf-search .yc-input { padding-left: 40px; }
@media (max-width: 640px) {
  .pzf-row { flex-direction: column; gap: 6px; }
  .pzf-label { width: auto; padding-top: 0; }
  .pzf-search { max-width: none; }
}
@media (prefers-reduced-motion: reduce) { .pzf-chip { transition: none; } }
</style>
