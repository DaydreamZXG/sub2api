import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import AffiliateView from '../AffiliateView.vue'

const { copyToClipboard, getAffiliateDetail } = vi.hoisted(() => ({
  copyToClipboard: vi.fn(),
  getAffiliateDetail: vi.fn(),
}))

vi.mock('@/api/user', () => ({
  default: {
    getAffiliateDetail,
    transferAffiliateQuota: vi.fn(),
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError: vi.fn(),
    showSuccess: vi.fn(),
  }),
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({
    refreshUser: vi.fn(),
  }),
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({ copyToClipboard }),
}))

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  }
})

describe('AffiliateView', () => {
  const affiliateCode = 'affiliate-code-that-is-long-enough-to-overflow-a-mobile-viewport'

  beforeEach(() => {
    vi.clearAllMocks()
    copyToClipboard.mockResolvedValue(true)
    getAffiliateDetail.mockResolvedValue({
      user_id: 1,
      aff_code: affiliateCode,
      inviter_id: null,
      aff_count: 0,
      aff_quota: 0,
      aff_frozen_quota: 0,
      aff_history_quota: 0,
      effective_rebate_rate_percent: 10,
      invitees: [],
    })
  })

  function mountView() {
    return mount(AffiliateView, {
      global: {
        stubs: {
          AppLayout: { template: '<main><slot /></main>' },
          Icon: true,
        },
      },
    })
  }

  it('keeps the long link and code on one clipped line each and copies them', async () => {
    const wrapper = mountView()
    await flushPromises()

    const values = wrapper.findAll('code')
    expect(values).toHaveLength(2)
    for (const value of values) {
      expect(value.classes()).toContain('yaf-copy-value')
    }

    const button = (text: string) => wrapper.findAll('button').find((b) => b.text() === text)!
    await button('affiliate.copyCode').trigger('click')
    await button('affiliate.copyLink').trigger('click')
    await flushPromises()

    expect(copyToClipboard).toHaveBeenNthCalledWith(1, affiliateCode, 'affiliate.codeCopied')
    expect(copyToClipboard).toHaveBeenNthCalledWith(
      2,
      `${window.location.origin}/register?aff=${encodeURIComponent(affiliateCode)}`,
      'affiliate.linkCopied',
    )
  })

  it('puts the rebate in the headline, and falls back to the page name when there is none', async () => {
    const withRate = mountView()
    await flushPromises()
    expect(withRate.get('[data-test="aff-headline"]').text()).toBe('affiliate.hero.title')

    getAffiliateDetail.mockResolvedValue({
      user_id: 1,
      aff_code: affiliateCode,
      inviter_id: null,
      aff_count: 0,
      aff_quota: 0,
      aff_frozen_quota: 0,
      aff_history_quota: 0,
      effective_rebate_rate_percent: 0,
      invitees: [],
    })
    const withoutRate = mountView()
    await flushPromises()
    expect(withoutRate.get('[data-test="aff-headline"]').text()).toBe('affiliate.title')
  })

  it('disables the transfer when there is nothing to move and lists the invited people', async () => {
    const empty = mountView()
    await flushPromises()
    expect(empty.get('[data-test="aff-transfer"]').attributes('disabled')).toBeDefined()
    expect(empty.text()).toContain('affiliate.transfer.empty')
    expect(empty.findAll('[data-test="aff-invitee"]')).toHaveLength(0)

    getAffiliateDetail.mockResolvedValue({
      user_id: 1,
      aff_code: affiliateCode,
      inviter_id: null,
      aff_count: 1,
      aff_quota: 12.5,
      aff_frozen_quota: 3,
      aff_history_quota: 40,
      effective_rebate_rate_percent: 10,
      invitees: [{ user_id: 7, email: 'friend@example.com', username: 'friend', total_rebate: 15.5, created_at: '2026-11-01T00:00:00Z' }],
    })
    const filled = mountView()
    await flushPromises()
    expect(filled.get('[data-test="aff-transfer"]').attributes('disabled')).toBeUndefined()
    expect(filled.get('[data-test="aff-frozen"]').text()).toContain('3.00')
    const rows = filled.findAll('[data-test="aff-invitee"]')
    expect(rows).toHaveLength(1)
    expect(rows[0].text()).toContain('friend@example.com')
    expect(rows[0].text()).toContain('15.50')
  })
})
