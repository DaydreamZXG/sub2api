<template>
  <AppLayout>
    <div data-testid="profile-shell" class="ypf">
      <nav class="ypf-nav" :aria-label="t('profile.nav.label')">
        <a
          v-for="item in sections"
          :key="item.id"
          :href="'#' + item.id"
          class="ypf-nav-link"
          @click.prevent="showSection(item.id)"
        >{{ t(item.label) }}</a>
      </nav>

      <div class="ypf-main">
        <ProfileInfoCard
          :user="user"
          :linuxdo-enabled="linuxdoOAuthEnabled"
          :dingtalk-enabled="dingtalkOAuthEnabled"
          :oidc-enabled="oidcOAuthEnabled"
          :oidc-provider-name="oidcOAuthProviderName"
          :wechat-enabled="wechatOAuthEnabled"
          :wechat-open-enabled="wechatOAuthOpenEnabled"
          :wechat-mp-enabled="wechatOAuthMPEnabled"
        />

        <section id="profile-password" class="ypf-sec">
          <ProfilePasswordForm />
        </section>

        <section v-if="user && balanceLowNotifyEnabled" id="profile-notify" class="ypf-sec">
          <ProfileBalanceNotifyCard
            :enabled="user.balance_notify_enabled ?? true"
            :threshold="user.balance_notify_threshold"
            :extra-emails="user.balance_notify_extra_emails ?? []"
            :system-default-threshold="systemDefaultThreshold"
            :user-email="user.email"
          />
        </section>

        <section id="profile-2fa" class="ypf-sec">
          <ProfileTotpCard />
        </section>

        <section id="profile-passkey" class="ypf-sec">
          <ProfilePasskeyCard :enabled="passkeyEnabled" />
        </section>

        <p v-if="contactInfo" class="ypf-contact">
          <Icon name="chat" size="md" />
          <span><b>{{ t('common.contactSupport') }}</b>{{ contactInfo }}</span>
        </p>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { Icon } from '@/components/icons'
import AppLayout from '@/components/layout/AppLayout.vue'
import ProfileBalanceNotifyCard from '@/components/user/profile/ProfileBalanceNotifyCard.vue'
import ProfileInfoCard from '@/components/user/profile/ProfileInfoCard.vue'
import ProfilePasswordForm from '@/components/user/profile/ProfilePasswordForm.vue'
import ProfileTotpCard from '@/components/user/profile/ProfileTotpCard.vue'
import ProfilePasskeyCard from '@/components/user/profile/ProfilePasskeyCard.vue'
import { isWeChatWebOAuthEnabled } from '@/api/auth'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'

const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()
const user = computed(() => authStore.user)

const contactInfo = ref('')
const balanceLowNotifyEnabled = ref(false)
const systemDefaultThreshold = ref(0)
const linuxdoOAuthEnabled = ref(false)
const dingtalkOAuthEnabled = ref(false)
const wechatOAuthEnabled = ref(false)
const wechatOAuthOpenEnabled = ref<boolean | undefined>(undefined)
const wechatOAuthMPEnabled = ref<boolean | undefined>(undefined)
const oidcOAuthEnabled = ref(false)
const oidcOAuthProviderName = ref('OIDC')
const passkeyEnabled = ref(false)

const sections = computed(() => [
  { id: 'profile-basics', label: 'profile.nav.basics' },
  { id: 'profile-sign-in', label: 'profile.nav.signIn' },
  { id: 'profile-password', label: 'profile.nav.password' },
  ...(user.value && balanceLowNotifyEnabled.value ? [{ id: 'profile-notify', label: 'profile.nav.notify' }] : []),
  { id: 'profile-2fa', label: 'profile.nav.twoFactor' },
  { id: 'profile-passkey', label: 'profile.nav.passkey' }
])

// Scrolls here instead of navigating to the hash: a hash navigation goes through the router,
// whose scrollBehavior sends new entries to the top of the page.
function showSection(id: string) {
  document.getElementById(id)?.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

onMounted(async () => {
  const profileRefresh = authStore.refreshUser().catch((error) => {
    console.error('Failed to refresh profile:', error)
  })

  const settingsLoad = appStore.fetchPublicSettings()
    .then((settings) => {
      if (!settings) {
        return
      }
      contactInfo.value = settings.contact_info || ''
      balanceLowNotifyEnabled.value = settings.balance_low_notify_enabled ?? false
      systemDefaultThreshold.value = settings.balance_low_notify_threshold ?? 0
      linuxdoOAuthEnabled.value = settings.linuxdo_oauth_enabled ?? false
      dingtalkOAuthEnabled.value = settings.dingtalk_oauth_enabled ?? false
      wechatOAuthEnabled.value = isWeChatWebOAuthEnabled(settings)
      wechatOAuthOpenEnabled.value = typeof settings.wechat_oauth_open_enabled === 'boolean'
        ? settings.wechat_oauth_open_enabled
        : undefined
      wechatOAuthMPEnabled.value = typeof settings.wechat_oauth_mp_enabled === 'boolean'
        ? settings.wechat_oauth_mp_enabled
        : undefined
      oidcOAuthEnabled.value = settings.oidc_oauth_enabled ?? false
      oidcOAuthProviderName.value = settings.oidc_oauth_provider_name || 'OIDC'
      passkeyEnabled.value = settings.passkey_enabled === true
    })
    .catch((error) => {
      console.error('Failed to load settings:', error)
    })

  await Promise.all([profileRefresh, settingsLoad])
})
</script>

<style scoped>
.ypf {
  display: grid;
  grid-template-columns: 148px minmax(0, 1fr);
  align-items: start;
  gap: 28px;
  max-width: 1040px;
}
.ypf-nav { position: sticky; top: 24px; display: grid; gap: 2px; }
.ypf-nav-link {
  padding: 8px 12px;
  border-radius: 10px;
  color: var(--yc-mut);
  font-size: 14px;
  font-weight: 600;
  white-space: nowrap;
}
.ypf-nav-link:hover { background: var(--yc-fill); color: var(--yc-ink); }
.ypf-main { display: grid; gap: 20px; min-width: 0; }
.ypf-sec { scroll-margin-top: 24px; }

/* The settings cards keep their own markup; here they take the page's panel look. */
.ypf-sec :deep(.card) {
  overflow: hidden;
  border: 1px solid var(--yc-line);
  border-radius: 22px;
  background: var(--yc-paper);
  box-shadow: none;
}
.ypf-sec :deep(.card > div:first-child) { padding: 22px 28px 16px; border-color: var(--yc-line); }
.ypf-sec :deep(.card > div:first-child h2) { font-size: 18px; font-weight: 720; color: var(--yc-ink); }
.ypf-sec :deep(.card > div:first-child p) { color: var(--yc-mut); }
.ypf-sec :deep(.card > div:nth-child(2)) { padding: 22px 28px 26px; }

.ypf-contact {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 16px 20px;
  border-radius: 16px;
  background: var(--yc-fill);
  color: var(--yc-mut);
  font-size: 14px;
}
.ypf-contact b { margin-right: 8px; color: var(--yc-ink); }

@media (max-width: 900px) {
  .ypf { grid-template-columns: minmax(0, 1fr); gap: 16px; }
  .ypf-nav { position: static; display: flex; gap: 6px; overflow-x: auto; padding-bottom: 2px; }
  .ypf-nav-link { border: 1px solid var(--yc-line); border-radius: 999px; padding: 6px 12px; font-size: 13px; }
  .ypf-sec :deep(.card > div:first-child) { padding: 18px 20px 14px; }
  .ypf-sec :deep(.card > div:nth-child(2)) { padding: 18px 20px 22px; }
}
</style>
