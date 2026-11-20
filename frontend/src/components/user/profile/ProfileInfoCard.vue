<template>
  <div class="ypf-info">
    <section data-testid="profile-overview-hero" class="ypf-hero">
      <div class="ypf-avatar">
        <img v-if="avatarUrl" :src="avatarUrl" :alt="displayName">
        <span v-else>{{ avatarInitial }}</span>
      </div>

      <div class="ypf-who">
        <div class="ypf-who-line">
          <h2 class="ypf-name">{{ displayName }}</h2>
          <span class="ypf-tag" :class="{ 'ypf-tag-lime': user?.role === 'admin' }">
            {{ user?.role === 'admin' ? t('profile.administrator') : t('profile.user') }}
          </span>
          <span class="ypf-tag" :class="user?.status === 'active' ? 'ypf-tag-ok' : 'ypf-tag-bad'">
            {{ user?.status === 'active' ? t('common.active') : t('common.disabled') }}
          </span>
        </div>
        <p class="ypf-email">{{ primaryEmailDisplay }}</p>
      </div>

      <dl class="ypf-facts">
        <div data-testid="profile-overview-metric-balance" class="ypf-fact">
          <dt>{{ t('profile.accountBalance') }}</dt>
          <dd class="ypf-num">{{ formatCurrency(user?.balance || 0) }}</dd>
        </div>
        <div data-testid="profile-overview-metric-concurrency" class="ypf-fact">
          <dt>{{ t('profile.concurrencyLimit') }}</dt>
          <dd class="ypf-num">{{ user?.concurrency || 0 }}</dd>
        </div>
        <div data-testid="profile-overview-metric-member-since" class="ypf-fact">
          <dt>{{ t('profile.memberSince') }}</dt>
          <dd>{{ memberSinceLabel }}</dd>
        </div>
      </dl>
    </section>

    <div data-testid="profile-main-column" class="ypf-stack">
      <section id="profile-basics" data-testid="profile-basics-panel" class="ypf-panel">
        <header class="ypf-panel-head">
          <h3>{{ t('profile.basicsTitle') }}</h3>
          <p>{{ t('profile.basicsDescription') }}</p>
        </header>
        <div class="ypf-two">
          <div class="ypf-cell">
            <ProfileAvatarCard :user="user" embedded />
          </div>
          <div class="ypf-cell">
            <ProfileEditForm :initial-username="user?.username || ''" embedded />
          </div>
        </div>
        <ul v-if="sourceHints.length" data-testid="profile-side-column" class="ypf-hints">
          <li v-for="hint in sourceHints" :key="hint.key">
            <Icon name="link" size="sm" />
            <span>{{ hint.text }}</span>
          </li>
        </ul>
        <span v-else data-testid="profile-side-column" hidden></span>
      </section>

      <section id="profile-sign-in" data-testid="profile-auth-bindings-panel" class="ypf-panel">
        <ProfileIdentityBindingsSection
          :user="user"
          :linuxdo-enabled="linuxdoEnabled"
          :dingtalk-enabled="dingtalkEnabled"
          :oidc-enabled="oidcEnabled"
          :oidc-provider-name="oidcProviderName"
          :wechat-enabled="wechatEnabled"
          :wechat-open-enabled="wechatOpenEnabled"
          :wechat-mp-enabled="wechatMpEnabled"
          embedded
          compact
        />
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import ProfileAvatarCard from '@/components/user/profile/ProfileAvatarCard.vue'
import ProfileEditForm from '@/components/user/profile/ProfileEditForm.vue'
import ProfileIdentityBindingsSection from '@/components/user/profile/ProfileIdentityBindingsSection.vue'
import type { User, UserAuthBindingStatus, UserAuthProvider, UserProfileSourceContext } from '@/types'
import { ycSign as sign } from '@/composables/ycMoneyState'


const props = withDefaults(defineProps<{
  user: User | null
  linuxdoEnabled?: boolean
  dingtalkEnabled?: boolean
  oidcEnabled?: boolean
  oidcProviderName?: string
  wechatEnabled?: boolean
  wechatOpenEnabled?: boolean
  wechatMpEnabled?: boolean
}>(), {
  linuxdoEnabled: false,
  dingtalkEnabled: false,
  oidcEnabled: false,
  oidcProviderName: 'OIDC',
  wechatEnabled: false,
  wechatOpenEnabled: undefined,
  wechatMpEnabled: undefined,
})

const { t } = useI18n()

function normalizeBindingStatus(binding: boolean | UserAuthBindingStatus | undefined): boolean | null {
  if (typeof binding === 'boolean') {
    return binding
  }
  if (!binding) {
    return null
  }
  if (typeof binding.bound === 'boolean') {
    return binding.bound
  }
  return Boolean(binding.provider_subject || binding.issuer || binding.provider_key)
}

function isEmailBound(user: User | null | undefined): boolean {
  if (typeof user?.email_bound === 'boolean') {
    return user.email_bound
  }

  const nested = user?.auth_bindings?.email ?? user?.identity_bindings?.email
  const normalized = normalizeBindingStatus(nested)
  return normalized ?? false
}

const avatarUrl = computed(() => props.user?.avatar_url?.trim() || '')
const displayName = computed(() => props.user?.username?.trim() || props.user?.email?.trim() || t('profile.user'))
const primaryEmailDisplay = computed(() => {
  const email = props.user?.email?.trim() || ''
  if (!email) {
    return ''
  }
  if (email.endsWith('.invalid') && !isEmailBound(props.user)) {
    return ''
  }
  return email
})
const avatarInitial = computed(() => displayName.value.charAt(0).toUpperCase() || 'U')
const memberSinceLabel = computed(() => {
  const raw = props.user?.created_at?.trim()
  if (!raw) {
    return '-'
  }

  const date = new Date(raw)
  if (Number.isNaN(date.getTime())) {
    return '-'
  }

  return new Intl.DateTimeFormat(undefined, {
    year: 'numeric',
    month: 'short',
  }).format(date)
})

const providerLabels = computed<Record<UserAuthProvider, string>>(() => ({
  email: t('profile.authBindings.providers.email'),
  linuxdo: t('profile.authBindings.providers.linuxdo'),
  dingtalk: t('profile.authBindings.providers.dingtalk'),
  oidc: t('profile.authBindings.providers.oidc', { providerName: props.oidcProviderName }),
  wechat: t('profile.authBindings.providers.wechat'),
  github: 'GitHub',
  google: 'Google'
}))

function formatCurrency(value: number): string {
  return `${sign.value}${value.toFixed(2)}`
}

function normalizeProvider(value: string): UserAuthProvider | null {
  const normalized = value.trim().toLowerCase()
  if (
    normalized === 'email' ||
    normalized === 'linuxdo' ||
    normalized === 'dingtalk' ||
    normalized === 'wechat' ||
    normalized === 'github' ||
    normalized === 'google'
  ) {
    return normalized
  }
  if (normalized === 'oidc' || normalized.startsWith('oidc:') || normalized.startsWith('oidc/')) {
    return 'oidc'
  }
  return null
}

function readObjectString(source: Record<string, unknown>, ...keys: string[]): string {
  for (const key of keys) {
    const value = source[key]
    if (typeof value === 'string' && value.trim()) {
      return value.trim()
    }
  }
  return ''
}

function resolveThirdPartySource(
  rawSource: string | UserProfileSourceContext | null | undefined
): { provider: UserAuthProvider; label: string } | null {
  if (!rawSource) {
    return null
  }

  if (typeof rawSource === 'string') {
    const provider = normalizeProvider(rawSource)
    if (!provider || provider === 'email') {
      return null
    }
    return {
      provider,
      label: providerLabels.value[provider]
    }
  }

  const sourceRecord = rawSource as Record<string, unknown>
  const provider = normalizeProvider(
    readObjectString(sourceRecord, 'provider', 'source', 'provider_type', 'auth_provider')
  )
  if (!provider || provider === 'email') {
    return null
  }

  const explicitLabel = readObjectString(
    sourceRecord,
    'provider_label',
    'label',
    'provider_name',
    'providerName'
  )

  return {
    provider,
    label: explicitLabel || providerLabels.value[provider]
  }
}

const sourceHints = computed(() => {
  const currentUser = props.user
  if (!currentUser) {
    return []
  }

  const hints: Array<{ key: string; text: string }> = []
  const avatarSource = resolveThirdPartySource(
    currentUser.profile_sources?.avatar ?? currentUser.avatar_source
  )
  const usernameSource = resolveThirdPartySource(
    currentUser.profile_sources?.username ??
      currentUser.profile_sources?.display_name ??
      currentUser.profile_sources?.nickname ??
      currentUser.display_name_source ??
      currentUser.username_source ??
      currentUser.nickname_source
  )

  if (avatarSource) {
    hints.push({
      key: 'avatar',
      text: t('profile.authBindings.source.avatar', { providerName: avatarSource.label })
    })
  }

  if (usernameSource) {
    hints.push({
      key: 'username',
      text: t('profile.authBindings.source.username', { providerName: usernameSource.label })
    })
  }

  return hints
})
</script>

<style scoped>
.ypf-info { display: grid; gap: 20px; }
.ypf-hero {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto;
  align-items: center;
  gap: 20px 24px;
  padding: 24px 28px;
  border-radius: 22px;
  background: var(--yc-ink);
  color: var(--yc-paper);
}
.ypf-avatar {
  display: grid;
  place-items: center;
  width: 72px;
  height: 72px;
  overflow: hidden;
  border-radius: 20px;
  background: var(--yc-lime);
  color: var(--yc-on-lime);
  font-size: 28px;
  font-weight: 750;
}
.ypf-avatar img { width: 100%; height: 100%; object-fit: cover; }
.ypf-who { min-width: 0; }
.ypf-who-line { display: flex; flex-wrap: wrap; align-items: center; gap: 8px; }
.ypf-name { min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: 24px; font-weight: 750; letter-spacing: -0.02em; }
.ypf-email { margin-top: 4px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: 14px; opacity: 0.7; }
.ypf-tag {
  padding: 2px 10px;
  border: 1px solid color-mix(in srgb, var(--yc-paper) 28%, transparent);
  border-radius: 999px;
  font-size: 12px;
  font-weight: 650;
  white-space: nowrap;
}
.ypf-tag-lime { border-color: var(--yc-lime); background: var(--yc-lime); color: var(--yc-on-lime); }
.ypf-tag-ok { border-color: var(--yc-lime); background: color-mix(in srgb, var(--yc-lime) 24%, transparent); }
.ypf-tag-bad { border-color: #ef4444; background: color-mix(in srgb, #ef4444 24%, transparent); }
.ypf-facts { display: flex; gap: 28px; margin: 0; }
.ypf-fact dt { font-size: 12px; opacity: 0.6; white-space: nowrap; }
.ypf-fact dd { margin: 4px 0 0; font-size: 20px; font-weight: 700; white-space: nowrap; }
.ypf-num { font-variant-numeric: tabular-nums; }

.ypf-stack { display: grid; gap: 20px; }
.ypf-panel {
  padding: 24px 28px;
  border: 1px solid var(--yc-line);
  border-radius: 22px;
  background: var(--yc-paper);
  scroll-margin-top: 24px;
}
.ypf-panel-head { margin-bottom: 18px; }
.ypf-panel-head h3 { font-size: 18px; font-weight: 720; color: var(--yc-ink); }
.ypf-panel-head p { margin-top: 4px; font-size: 14px; color: var(--yc-mut); }
.ypf-two { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 16px; }
.ypf-cell { padding: 18px; border-radius: 16px; background: var(--yc-fill); }
.ypf-hints { display: grid; gap: 8px; margin: 16px 0 0; padding: 0; list-style: none; }
.ypf-hints li { display: flex; align-items: flex-start; gap: 8px; font-size: 13px; color: var(--yc-mut); }

@media (max-width: 900px) {
  .ypf-hero { grid-template-columns: auto minmax(0, 1fr); padding: 20px; }
  .ypf-facts { grid-column: 1 / -1; flex-wrap: wrap; gap: 16px 28px; }
  .ypf-two { grid-template-columns: minmax(0, 1fr); }
  .ypf-panel { padding: 20px; }
}
</style>
