<template>
  <YcShell>
    <h1 class="yc-page-h1">{{ t('yc.help.title') }}</h1>
    <p class="yc-page-lead">{{ t('yc.help.lead') }}</p>

    <section class="yc-surface yc-help">
      <h2 class="yc-sec-h2">{{ t('yc.help.contactTitle') }}</h2>
      <p v-if="contact" class="yc-help-contact">{{ contact }}</p>
      <p v-else class="yc-note">{{ t('yc.help.noContact') }}</p>
      <a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer" class="yc-btn yc-btn-ghost yc-btn-sm">{{ t('yc.help.docs') }} →</a>
    </section>

    <section v-if="news.length" class="yc-surface yc-help">
      <h2 class="yc-sec-h2">{{ t('yc.help.newsTitle') }}</h2>
      <details v-for="n in news" :key="n.id" class="yc-help-item" @toggle="read(n)">
        <summary><i v-if="!n.read_at" aria-hidden="true"></i>{{ n.title }}<time>{{ formatRelativeTime(n.created_at) }}</time></summary>
        <p>{{ n.content }}</p>
      </details>
    </section>

    <section class="yc-surface yc-help">
      <h2 class="yc-sec-h2">{{ t('yc.faq.eyebrow') }}</h2>
      <details v-for="k in FAQS" :key="k" class="yc-help-item">
        <summary>{{ t(`yc.faq.q${k}`) }}</summary>
        <p>{{ t(`yc.faq.a${k}`) }}</p>
      </details>
    </section>
  </YcShell>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import { list, markRead } from '@/api/announcements'
import type { UserAnnouncement } from '@/types'
import { formatRelativeTime } from '@/utils/format'
import { sanitizeUrl } from '@/utils/url'
import YcShell from '@/components/yecai/YcShell.vue'

const FAQS = [1, 2, 3, 4, 5]

const { t } = useI18n()
const appStore = useAppStore()

const contact = computed(() => appStore.cachedPublicSettings?.contact_info || '')
const docUrl = computed(() => sanitizeUrl(appStore.cachedPublicSettings?.doc_url || ''))
const news = ref<UserAnnouncement[]>([])

function read(item: UserAnnouncement) {
  if (item.read_at) return
  item.read_at = new Date().toISOString()
  markRead(item.id).catch(() => undefined)
}

onMounted(async () => {
  try {
    news.value = await list()
  } catch (error) {
    console.error('Failed to load announcements:', error)
  }
})
</script>

<style scoped>
.yc-help { margin-top: 20px; padding: 26px 30px; }
.yc-help-contact { margin: 10px 0 4px; font-size: 16px; line-height: 1.7; white-space: pre-wrap; overflow-wrap: anywhere; }
.yc-help .yc-note { margin-top: 8px; }
.yc-help .yc-btn { margin-top: 14px; }
.yc-help-item { border-top: 1px solid var(--yc-line); }
.yc-help .yc-sec-h2 + .yc-help-item { margin-top: 14px; }
.yc-help-item summary { display: flex; align-items: center; gap: 10px; padding: 15px 0; cursor: pointer; list-style: none; font-size: 15.5px; font-weight: 600; }
.yc-help-item summary::-webkit-details-marker { display: none; }
.yc-help-item summary::after { content: '+'; margin-left: auto; color: var(--yc-dim); font-weight: 400; }
.yc-help-item[open] summary::after { content: '−'; }
.yc-help-item summary i { flex: none; width: 8px; height: 8px; border-radius: 50%; background: var(--yc-lime); box-shadow: 0 0 0 1px var(--yc-ink); }
.yc-help-item summary time { font-size: 12.5px; font-weight: 400; color: var(--yc-dim); }
.yc-help-item p { padding: 0 0 16px; font-size: 14.5px; line-height: 1.7; color: var(--yc-mut); white-space: pre-wrap; overflow-wrap: anywhere; }
@media (max-width: 760px) {
  .yc-help { padding: 22px 20px; }
}
</style>
