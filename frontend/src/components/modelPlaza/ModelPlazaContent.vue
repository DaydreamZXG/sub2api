<template>
  <div class="pz">
    <header>
      <h1 class="yc-page-h1">{{ t('modelPlaza.title') }}</h1>
      <p class="yc-page-lead">{{ t('modelPlaza.description') }}</p>
    </header>

    <!-- Price notes written by the admin (Markdown). -->
    <div v-if="descriptionHtml" class="yc-surface pz-notes" v-html="descriptionHtml"></div>

    <p v-if="!isAuthenticated" class="yc-note pz-hint">
      <Icon name="infoCircle" size="sm" />
      {{ t('modelPlaza.anonymousHint') }}
    </p>

    <p v-if="loading" class="yc-surface pz-state" role="status">{{ t('yc.prices.loading') }}</p>
    <p v-else-if="error" class="yc-surface pz-state pz-state-bad" role="alert">{{ t('modelPlaza.loadFailed') }}</p>
    <template v-else>
      <PlazaFilterBar
        :platforms="platforms"
        :groups="groupOptions"
        :platform="selectedPlatform"
        :group-id="selectedGroupId"
        :search="searchQuery"
        @update:platform="selectedPlatform = $event"
        @update:group-id="selectedGroupId = $event"
        @update:search="searchQuery = $event"
      />

      <div v-if="filteredGroups.length > 0" class="pz-lines">
        <PlazaGroupSection v-for="g in filteredGroups" :key="g.id" :group="g" />
      </div>
      <p v-else class="yc-surface pz-state">
        {{ searchActive ? t('modelPlaza.noSearchResult') : t('modelPlaza.empty') }}
      </p>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import Icon from '@/components/icons/Icon.vue'
import PlazaFilterBar from './PlazaFilterBar.vue'
import PlazaGroupSection from './PlazaGroupSection.vue'
import type { ModelPlazaGroup, ModelPlazaResponse } from '@/api/modelPlaza'
import { useAuthStore } from '@/stores/auth'

const props = defineProps<{
  response: ModelPlazaResponse | null
  loading: boolean
  error?: boolean
}>()

const { t } = useI18n()
const authStore = useAuthStore()
const isAuthenticated = computed(() => authStore.isAuthenticated)

const selectedPlatform = ref<string>('all')
const selectedGroupId = ref<number | 'all'>('all')
const searchQuery = ref('')

const searchActive = computed(() => searchQuery.value.trim() !== '')

const descriptionHtml = computed(() => {
  const md = props.response?.description?.trim()
  if (!md) return ''
  return DOMPurify.sanitize(marked.parse(md) as string)
})

/** 生效倍率 = 用户专属倍率 ?? 分组默认倍率。 */
function effectiveRate(g: ModelPlazaGroup): number {
  return g.user_rate_multiplier ?? g.rate_multiplier
}

const platforms = computed(() =>
  [...new Set((props.response?.groups ?? []).map((g) => g.platform).filter(Boolean))].sort()
)

const groupOptions = computed(() =>
  (props.response?.groups ?? []).map((g) => ({
    id: g.id,
    name: g.name,
    platform: g.platform
  }))
)

const filteredGroups = computed(() => {
  let groups = props.response?.groups ?? []
  if (selectedPlatform.value !== 'all') {
    groups = groups.filter((g) => g.platform === selectedPlatform.value)
  }
  if (selectedGroupId.value !== 'all') {
    groups = groups.filter((g) => g.id === selectedGroupId.value)
  }
  // 模型名搜索:分组内只留命中的模型,整组无命中则隐藏该分组。
  const q = searchQuery.value.trim().toLowerCase()
  if (q) {
    groups = groups
      .map((g) => ({ ...g, models: g.models.filter((m) => m.name.toLowerCase().includes(q)) }))
      .filter((g) => g.models.length > 0)
  }
  // 专属倍率会改变生效值,不能只依赖后端按默认倍率的排序。
  return [...groups].sort(
    (a, b) => effectiveRate(a) - effectiveRate(b) || a.name.localeCompare(b.name)
  )
})
</script>

<style scoped>
.pz { display: grid; gap: 20px; }
.pz-hint { display: flex; align-items: center; gap: 6px; margin: 0; }
.pz-state { margin: 0; padding: 48px 24px; text-align: center; font-size: 14.5px; color: var(--yc-mut); }
.pz-state-bad { color: #c2410c; }
.pz-lines { display: grid; gap: 16px; }
.pz-notes { padding: 18px 24px; font-size: 14px; line-height: 1.7; overflow-wrap: anywhere; color: var(--yc-ink); }
.pz-notes :deep(h1), .pz-notes :deep(h2), .pz-notes :deep(h3) { margin: 12px 0 8px; font-size: 15px; font-weight: 700; }
.pz-notes :deep(:first-child) { margin-top: 0; }
.pz-notes :deep(p) { margin: 0 0 8px; }
.pz-notes :deep(:last-child) { margin-bottom: 0; }
.pz-notes :deep(a) { text-decoration: underline; text-underline-offset: 4px; }
.pz-notes :deep(ul) { margin: 0 0 8px; padding-left: 20px; list-style: disc; }
.pz-notes :deep(ol) { margin: 0 0 8px; padding-left: 20px; list-style: decimal; }
.pz-notes :deep(code) { padding: 2px 6px; border-radius: 6px; background: var(--yc-fill); font-family: ui-monospace, SFMono-Regular, Menlo, monospace; font-size: 12.5px; }
.pz-notes :deep(blockquote) { margin: 8px 0; padding-left: 12px; border-left: 3px solid var(--yc-line); color: var(--yc-mut); }
</style>
