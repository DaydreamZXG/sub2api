<template>
  <header class="pzn">
    <div class="yc-wrap pzn-in">
      <router-link to="/home" class="pzn-brand">
        <YcBrand />
      </router-link>
      <div class="pzn-actions">
        <YcThemeSwitch />
        <router-link v-if="isAuthenticated" :to="backTarget" class="yc-btn yc-btn-sm yc-btn-lime">
          {{ t('modelPlaza.nav.backToDashboard') }}
        </router-link>
        <router-link
          v-else
          :to="{ path: '/login', query: { redirect: '/model-plaza' } }"
          class="yc-btn yc-btn-sm yc-btn-lime"
        >
          {{ t('modelPlaza.nav.login') }}
        </router-link>
      </div>
    </div>
  </header>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import YcBrand from '@/components/yecai/YcBrand.vue'
import YcThemeSwitch from '@/components/yecai/YcThemeSwitch.vue'
import { useAuthStore } from '@/stores/auth'

const { t } = useI18n()
const authStore = useAuthStore()

const isAuthenticated = computed(() => authStore.isAuthenticated)
const backTarget = computed(() => (authStore.isAdmin ? '/admin/dashboard' : '/dashboard'))
</script>

<style scoped>
.pzn { position: sticky; top: 0; z-index: 20; border-bottom: 1px solid var(--yc-line); background: color-mix(in srgb, var(--yc-bg) 82%, transparent); backdrop-filter: saturate(160%) blur(14px); }
.pzn-in { display: flex; align-items: center; justify-content: space-between; gap: 16px; height: 64px; }
.pzn-brand { display: flex; align-items: center; gap: 10px; font-size: 17px; font-weight: 750; letter-spacing: -0.01em; white-space: nowrap; color: var(--yc-ink); }
.pzn-actions { display: flex; align-items: center; gap: 12px; }
</style>
