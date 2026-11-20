<template>
  <!-- Signed in: inside the site's own sidebar. -->
  <YcShell v-if="authStore.isAuthenticated">
    <ModelPlazaContent :response="data" :loading="loading" :error="loadFailed" />
  </YcShell>

  <!-- Visitors: a public page with its own top bar. -->
  <div v-else class="yc pzv">
    <PlazaNavBar />
    <main class="yc-wrap pzv-main">
      <ModelPlazaContent :response="data" :loading="loading" :error="loadFailed" />
    </main>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import YcShell from '@/components/yecai/YcShell.vue'
import PlazaNavBar from '@/components/modelPlaza/PlazaNavBar.vue'
import ModelPlazaContent from '@/components/modelPlaza/ModelPlazaContent.vue'
import { getModelPlaza, type ModelPlazaResponse } from '@/api/modelPlaza'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import { useYcMoney } from '@/composables/useYcMoney'

const appStore = useAppStore()
const authStore = useAuthStore()

// Loads what ¥1 buys, so the prices below are in yuan.
useYcMoney()

const data = ref<ModelPlazaResponse | null>(null)
const loading = ref(true)
const loadFailed = ref(false)

onMounted(async () => {
  // 独立形态导航条需要站点名/Logo;有 __APP_CONFIG__ 注入时同步命中缓存。
  void appStore.fetchPublicSettings()
  try {
    data.value = await getModelPlaza()
  } catch {
    loadFailed.value = true
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.pzv { min-height: 100vh; background: var(--yc-bg); color: var(--yc-ink); }
.pzv-main { padding-top: 40px; padding-bottom: 72px; }
@media (max-width: 640px) { .pzv-main { padding-top: 24px; } }
</style>
