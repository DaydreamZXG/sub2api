import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount, type VueWrapper } from '@vue/test-utils'
import { nextTick } from 'vue'

import type { ApiKey } from '@/types'
import { keysAPI } from '@/api'
import KeysView from '../KeysView.vue'

const {
  issueConnectCode,
  openDesktopApp,
  listKeys,
  updateKey,
  getPublicSettings,
  getDashboardApiKeysUsage,
  getAvailableGroups,
  getUserGroupRates,
  showError,
  showSuccess,
  copyToClipboard,
  isCurrentStep,
  nextStep,
} = vi.hoisted(() => ({
  issueConnectCode: vi.fn(),
  openDesktopApp: vi.fn(),
  listKeys: vi.fn(),
  updateKey: vi.fn(),
  getPublicSettings: vi.fn(),
  getDashboardApiKeysUsage: vi.fn(),
  getAvailableGroups: vi.fn(),
  getUserGroupRates: vi.fn(),
  showError: vi.fn(),
  showSuccess: vi.fn(),
  copyToClipboard: vi.fn(),
  isCurrentStep: vi.fn(),
  nextStep: vi.fn(),
}))

const messages: Record<string, string> = {
  'common.actions': 'Actions',
  'common.name': 'Name',
  'common.refresh': 'Refresh',
  'common.status': 'Status',
  'keys.apiKey': 'API Key',
  'keys.allGroups': 'All Groups',
  'keys.allStatus': 'All Status',
  'keys.createKey': 'Create API Key',
  'keys.created': 'Created',
  'keys.expiresAt': 'Expires',
  'keys.group': 'Group',
  'keys.id': 'ID',
  'keys.currentConcurrency': 'Current Concurrency',
  'keys.lastUsedAt': 'Last Used',
  'keys.lastUsedIP': 'Last Used IP',
  'keys.rateLimitColumn': 'Rate Limit',
  'keys.searchPlaceholder': 'Search name or key...',
  'keys.status.active': 'Active',
  'keys.status.expired': 'Expired',
  'keys.status.inactive': 'Inactive',
  'keys.status.quota_exhausted': 'Quota exhausted',
  'keys.usage': 'Usage',
}

vi.mock('@/api', () => ({
  keysAPI: {
    list: listKeys,
    create: vi.fn(),
    update: updateKey,
    delete: vi.fn(),
    toggleStatus: vi.fn(),
  },
  authAPI: {
    getPublicSettings,
  },
  usageAPI: {
    getDashboardApiKeysUsage,
  },
  userGroupsAPI: {
    getAvailable: getAvailableGroups,
    getUserGroupRates,
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError,
    showSuccess,
    showInfo: vi.fn(),
  }),
}))

vi.mock('@/api/desktop', () => ({
  issueConnectCode,
  openDesktopApp,
  stationOrigin: () => 'https://yeschoy.test',
}))

vi.mock('@/stores/onboarding', () => ({
  useOnboardingStore: () => ({
    isCurrentStep,
    nextStep,
  }),
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({
    copyToClipboard,
  }),
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => messages[key] ?? key,
    }),
  }
})

const createApiKey = (): ApiKey => ({
  id: 1,
  user_id: 1,
  key: 'sk-test-key',
  name: 'test-key',
  group_id: null,
  status: 'active',
  ip_whitelist: [],
  ip_blacklist: [],
  last_used_at: null,
  last_used_ip: null,
  quota: 0,
  quota_used: 0,
  expires_at: null,
  created_at: '2026-06-27T00:00:00Z',
  updated_at: '2026-06-27T00:00:00Z',
  current_concurrency: 3,
  rate_limit_5h: 0,
  rate_limit_1d: 0,
  rate_limit_7d: 0,
  usage_5h: 0,
  usage_1d: 0,
  usage_7d: 0,
  window_5h_start: null,
  window_1d_start: null,
  window_7d_start: null,
  reset_5h_at: null,
  reset_1d_at: null,
  reset_7d_at: null,
})

const AppLayoutStub = {
  template: '<div><slot /></div>',
}

const SelectStub = {
  name: 'Select',
  props: ['modelValue', 'options'],
  emits: ['update:modelValue'],
  template: '<select :value="modelValue" @change="$emit(\'update:modelValue\', $event.target.value)"></select>',
}

const SearchInputStub = {
  name: 'SearchInput',
  props: ['modelValue'],
  emits: ['update:modelValue', 'search'],
  template: '<input :value="modelValue" @input="$emit(\'update:modelValue\', $event.target.value)" />',
}

const PaginationStub = {
  name: 'Pagination',
  props: ['page', 'total', 'pageSize'],
  emits: ['update:page', 'update:pageSize'],
  template: `
    <div>
      <button data-test="page-size-50" @click="$emit('update:pageSize', 50)">50</button>
      <button data-test="page-2" @click="$emit('update:page', 2)">Page 2</button>
    </div>
  `,
}

const IconStub = {
  props: ['name'],
  template: '<span data-test="icon">{{ name }}</span>',
}

const mountView = async () => {
  const wrapper = mount(KeysView, {
    global: {
      stubs: {
        teleport: true,
        RouterLink: { props: ['to'], template: '<a :data-to="JSON.stringify(to)"><slot /></a>' },
        AppLayout: AppLayoutStub,
        Pagination: PaginationStub,
        BaseDialog: {
          props: ['show', 'title'],
          emits: ['close'],
          template: '<div v-if="show" role="dialog"><button data-test="close-dialog" @click="$emit(\'close\')">Close</button><slot /><slot name="footer" /></div>',
        },
        ConfirmDialog: true,
        EmptyState: true,
        Select: SelectStub,
        SearchInput: SearchInputStub,
        Icon: IconStub,
        UseKeyModal: true,
        BulkEditKeysModal: true,
        EndpointPopover: true,
        GroupBadge: true,
        GroupOptionItem: true,
        Teleport: true,
      },
    },
  })
  await flushPromises()
  await nextTick()
  return wrapper
}

const picked = (wrapper: VueWrapper) =>
  wrapper.findAll('[data-test="key-select"]').map((box) => (box.element as HTMLInputElement).checked)

const pick = async (wrapper: VueWrapper, index = 0) => {
  await wrapper.findAll('[data-test="key-select"]')[index].trigger('change')
}

const sortSelect = (wrapper: VueWrapper) =>
  wrapper.findAllComponents({ name: 'Select' }).find((select) => select.attributes('data-test') === 'key-sort')!

const getButtonByText = (wrapper: VueWrapper, text: string) => {
  const button = wrapper.findAll('button').find((item) => item.text().includes(text))
  if (!button) {
    throw new Error(`Button not found: ${text}`)
  }
  return button
}

describe('user KeysView', () => {
  beforeEach(() => {
    localStorage.clear()

    listKeys.mockReset()
    updateKey.mockReset()
    vi.mocked(keysAPI.create).mockReset()
    getPublicSettings.mockReset()
    getDashboardApiKeysUsage.mockReset()
    getAvailableGroups.mockReset()
    getUserGroupRates.mockReset()
    showError.mockReset()
    showSuccess.mockReset()
    copyToClipboard.mockReset()
    isCurrentStep.mockReset()
    nextStep.mockReset()

    listKeys.mockResolvedValue({
      items: [createApiKey()],
      total: 1,
      page: 1,
      page_size: 20,
      pages: 1,
    })
    getPublicSettings.mockResolvedValue({})
    getDashboardApiKeysUsage.mockResolvedValue({ stats: {} })
    getAvailableGroups.mockResolvedValue([])
    getUserGroupRates.mockResolvedValue({})
    isCurrentStep.mockReturnValue(false)
  })

  it.each([
    { initialStatus: 'quota_exhausted', status: 'active', formStatus: 'active' },
    { initialStatus: 'inactive', status: 'inactive', formStatus: 'inactive' },
    { initialStatus: 'active', status: 'active', formStatus: 'inactive' },
  ] as const)('syncs quota reset from $initialStatus to $status with form status $formStatus', async ({ initialStatus, status, formStatus }) => {
    const key: ApiKey = {
      ...createApiKey(), group_id: 1, quota: 10, quota_used: 10,
      status: initialStatus,
    }
    listKeys.mockResolvedValueOnce({ items: [key], total: 1, page: 1, page_size: 20, pages: 1 })
    updateKey.mockResolvedValue({ ...key, status, quota_used: 0 })
    const wrapper = await mountView()
    await wrapper.get('[data-test="key-more"]').trigger('click')
    await getButtonByText(wrapper, 'common.edit').trigger('click')
    await wrapper.get('[data-tour="key-form-name"]').setValue('Unsaved name')
    const statusSelect = wrapper.findAllComponents({ name: 'Select' })
      .find((select) => select.props('options').length === 2 &&
        select.props('options')[0].value === 'active')!
    statusSelect.vm.$emit('update:modelValue', 'inactive')
    await wrapper.get('button[title="keys.resetQuotaUsed"]').trigger('click')
    const confirmation = wrapper.findAllComponents({ name: 'ConfirmDialog' })
      .find((dialog) => dialog.props('title') === 'keys.resetQuotaTitle')!
    confirmation.vm.$emit('confirm')
    await flushPromises()

    expect(updateKey).toHaveBeenNthCalledWith(1, key.id, { reset_quota: true })
    expect(wrapper.get('[data-test="key-card"] .yk-status').classes()).toContain(`yk-status-${status}`)
    expect(statusSelect.props('modelValue')).toBe(formStatus)
    expect((wrapper.get('[data-tour="key-form-name"]').element as HTMLInputElement).value)
      .toBe('Unsaved name')

    await wrapper.get('#key-form').trigger('submit')
    await flushPromises()
    expect(updateKey).toHaveBeenNthCalledWith(2, key.id, expect.objectContaining({ name: 'Unsaved name', status: formStatus }))
    wrapper.unmount()
  })

  it('copies the key from its card and has no desktop-app button', async () => {
    copyToClipboard.mockResolvedValue(true)
    const wrapper = await mountView()
    expect(wrapper.find('[data-test="key-app"]').exists()).toBe(false)
    await wrapper.get('[data-test="key-copy"]').trigger('click')
    await flushPromises()
    expect(copyToClipboard).toHaveBeenCalledWith(createApiKey().key, 'keys.copied')
    wrapper.unmount()
  })

  it('shows the API address without a trailing slash and copies it', async () => {
    getPublicSettings.mockResolvedValue({ api_base_url: 'https://api.yeschoy.test/' })
    const wrapper = await mountView()
    const bar = wrapper.get('[data-test="api-address"]')

    expect(bar.get('code').text()).toBe('https://api.yeschoy.test')
    await bar.get('button').trigger('click')
    expect(copyToClipboard).toHaveBeenCalledWith('https://api.yeschoy.test', 'keys.copied')
    wrapper.unmount()
  })

  it('keeps the guide, disable, edit, delete and command-line setup in the more menu and closes it once one is chosen', async () => {
    const wrapper = await mountView()
    expect(wrapper.find('[data-test="key-delete"]').exists()).toBe(false)

    await wrapper.get('[data-test="key-more"]').trigger('click')
    expect(wrapper.get('[role="menu"]').findAll('[role="menuitem"]').map((item) => item.attributes('data-test')))
      .toEqual(['key-guide', 'key-toggle', 'key-edit', 'key-delete', 'key-manual'])

    await wrapper.get('[data-test="key-delete"]').trigger('click')
    expect(wrapper.find('[role="menu"]').exists()).toBe(false)
    expect(wrapper.findAllComponents({ name: 'ConfirmDialog' }).some((dialog) => dialog.props('show'))).toBe(true)
    wrapper.unmount()
  })

  it('closes the more menu on Escape and on a click elsewhere', async () => {
    const wrapper = await mountView()
    const more = wrapper.get('[data-test="key-more"]')

    await more.trigger('click')
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' }))
    await flushPromises()
    expect(wrapper.find('[role="menu"]').exists()).toBe(false)

    await more.trigger('click')
    document.body.click()
    await flushPromises()
    expect(wrapper.find('[role="menu"]').exists()).toBe(false)
    wrapper.unmount()
  })

  it('keeps the low-frequency facts folded inside the card', async () => {
    const wrapper = await mountView()
    const details = wrapper.get('[data-test="key-details"]')
    expect(details.attributes('open')).toBeUndefined()
    expect(details.get('[data-test="key-id"]').text()).toBe('#1')
    expect(details.find('[data-test="last-used-ip"]').exists()).toBe(true)
    wrapper.unmount()
  })

  it('opens bulk editing with the ticked keys', async () => {
    const wrapper = await mountView()
    expect(wrapper.find('[data-test="bulk-edit-keys"]').exists()).toBe(false)
    await pick(wrapper)
    await wrapper.get('[data-test="bulk-edit-keys"]').trigger('click')
    const modal = wrapper.findComponent({ name: 'BulkEditKeysModal' })
    expect(modal.props('show')).toBe(true)
    expect(modal.props('selectedKeys').map((key: ApiKey) => key.id)).toEqual([1])
    wrapper.unmount()
  })

  it.each(['filter', 'page size', 'sort'])('clears selection on %s changes', async (change) => {
    const wrapper = await mountView()
    await pick(wrapper)
    expect(picked(wrapper)).toEqual([true])
    if (change === 'filter') {
      wrapper.findComponent({ name: 'SearchInput' }).vm.$emit('search')
    } else if (change === 'page size') {
      await wrapper.get('[data-test="page-size-50"]').trigger('click')
    } else {
      sortSelect(wrapper).vm.$emit('update:modelValue', 'created_at:asc')
    }
    await flushPromises()
    expect(picked(wrapper)).toEqual([false])
    expect(wrapper.find('[data-test="bulk-edit-keys"]').exists()).toBe(false)
    wrapper.unmount()
  })

  it('removes successful keys from the selection and refreshes the table', async () => {
    listKeys.mockResolvedValue({
      items: [createApiKey(), { ...createApiKey(), id: 2, name: 'Second' }],
      total: 2, pages: 1
    })
    const wrapper = await mountView()
    await wrapper.get('[data-test="key-select"]').trigger('change')
    await wrapper.get('[data-test="select-all-keys"]').trigger('click')
    await wrapper.get('[data-test="bulk-edit-keys"]').trigger('click')
    wrapper.findComponent({ name: 'BulkEditKeysModal' }).vm.$emit('updated', [1])
    await flushPromises()
    expect(listKeys).toHaveBeenCalledTimes(2)
    expect(picked(wrapper)).toEqual([false, true])
    wrapper.unmount()
  })

  it('drops keys that are no longer visible after a refresh', async () => {
    const wrapper = await mountView()
    await pick(wrapper)
    listKeys.mockResolvedValue({ items: [], total: 0, pages: 0 })
    await wrapper.get('button[title="Refresh"]').trigger('click')
    await flushPromises()
    expect(picked(wrapper)).toEqual([])
    expect(wrapper.find('[data-test="bulk-edit-keys"]').exists()).toBe(false)
    wrapper.unmount()
  })

  it('renders the current concurrency value', async () => {
    const wrapper = await mountView()

    expect(wrapper.get('[data-test="current-concurrency"]').text()).toBe('3')
  })

  it.each([
    [0.8, 'keys.lineCheaper'],
    [1.5, 'keys.lineDearer'],
    [1, null],
  ])('says a line at %s times the standard price in words, not as a multiplier', async (rate, label) => {
    listKeys.mockResolvedValue({
      items: [{ ...createApiKey(), group_id: 7, group: { id: 7, name: 'Line', platform: 'anthropic', rate_multiplier: rate } }],
      total: 1, page: 1, page_size: 20, pages: 1,
    })
    const wrapper = await mountView()
    const tag = wrapper.find('[data-test="key-line-price"]')
    expect(tag.exists()).toBe(label !== null)
    if (label) expect(tag.text()).toBe(label)
    wrapper.unmount()
  })

  it('reloads the keys in the chosen order', async () => {
    const wrapper = await mountView()
    expect(sortSelect(wrapper).props('modelValue')).toBe('created_at:desc')
    sortSelect(wrapper).vm.$emit('update:modelValue', 'current_concurrency:desc')
    await flushPromises()
    expect(listKeys.mock.calls.at(-1)).toEqual(
      expect.arrayContaining([expect.objectContaining({ sort_by: 'current_concurrency', sort_order: 'desc' })])
    )
    wrapper.unmount()
  })

  it.each([
    { key: 'current_concurrency', order: 'asc' },
    { key: 'group', order: 'asc' },
    { key: 'group', order: 'desc' },
  ] as const)('keeps filters and resets pagination and selection when sorting $key $order', async ({ key, order }) => {
    getAvailableGroups.mockResolvedValue([{ id: 42, name: 'OpenAI' }])
    const wrapper = await mountView()

    await wrapper.get('[data-test="page-size-50"]').trigger('click')
    await flushPromises()

    await wrapper.findComponent({ name: 'SearchInput' }).vm.$emit('update:modelValue', 'target')
    await wrapper.findComponent({ name: 'SearchInput' }).vm.$emit('search')
    await flushPromises()

    const selects = wrapper.findAllComponents({ name: 'Select' })
    await selects[0].vm.$emit('update:modelValue', 42)
    await flushPromises()
    await selects[1].vm.$emit('update:modelValue', 'active')
    await flushPromises()

    await wrapper.get('[data-test="page-2"]').trigger('click')
    await flushPromises()
    await pick(wrapper)
    expect(picked(wrapper)).toEqual([true])
    listKeys.mockClear()

    sortSelect(wrapper).vm.$emit('update:modelValue', `${key}:${order}`)
    await flushPromises()

    expect(picked(wrapper)).toEqual([false])
    expect(listKeys).toHaveBeenLastCalledWith(
      1,
      50,
      {
        search: 'target',
        status: 'active',
        group_id: 42,
        sort_by: key,
        sort_order: order,
      },
      expect.objectContaining({ signal: expect.any(AbortSignal) })
    )
  })

  describe('create provider selection', () => {
    const platforms = ['anthropic', 'openai', 'kimi', 'zhipu', 'deepseek', 'minimax', 'gemini', 'grok', 'antigravity', 'composite', 'opencode_go', 'typesafe']
    const availableGroups = platforms.map((platform, index) => ({
      id: index + 1,
      // Deliberately ambiguous names: classification must follow the platform.
      name: `Shared group ${index + 1}`,
      platform,
      rate_multiplier: 1,
      subscription_type: 'standard',
    }))
    const groupSelect = (wrapper: VueWrapper) => wrapper.findComponent('[data-tour="key-form-group"]')
    const optionIds = (wrapper: VueWrapper) => groupSelect(wrapper).props('options').map((option: { value: number }) => option.value)
    const chooseProvider = (wrapper: VueWrapper, value: string) => wrapper.get(`input[name="key-provider"][value="${value}"]`).setValue()
    const openCreate = async () => {
      const wrapper = await mountView()
      await wrapper.get('[data-tour="keys-create-btn"]').trigger('click')
      return wrapper
    }

    beforeEach(() => {
      getAvailableGroups.mockResolvedValue(availableGroups)
    })

    it('classifies all configured platforms and retains the complete table filter', async () => {
      const wrapper = await openCreate()
      expect(wrapper.findAll('input[name="key-provider"]')).toHaveLength(4)
      expect(optionIds(wrapper)).toEqual([1])
      await chooseProvider(wrapper, 'openai')
      expect(optionIds(wrapper)).toEqual([2])
      await chooseProvider(wrapper, 'domestic')
      expect(optionIds(wrapper)).toEqual([3, 4, 5, 6])
      await chooseProvider(wrapper, 'other')
      expect(optionIds(wrapper)).toEqual([7, 8, 9, 10, 11, 12])
      expect(wrapper.findAllComponents({ name: 'Select' })[0].props('options')).toHaveLength(14)
    })

    it('clears the previous group on provider change and submits only the newly selected group', async () => {
      const wrapper = await openCreate()
      await wrapper.get('[data-tour="key-form-name"]').setValue('My key')
      await groupSelect(wrapper).vm.$emit('update:modelValue', 1)
      await chooseProvider(wrapper, 'domestic')
      expect(groupSelect(wrapper).props('modelValue')).toBeNull()
      await wrapper.get('#key-form').trigger('submit')
      expect(keysAPI.create).not.toHaveBeenCalled()
      expect(showError).toHaveBeenCalledWith('keys.groupRequired')

      await groupSelect(wrapper).vm.$emit('update:modelValue', 5)
      vi.mocked(keysAPI.create).mockResolvedValue({ ...createApiKey(), group_id: 5 })
      await wrapper.get('#key-form').trigger('submit')
      await flushPromises()
      expect(keysAPI.create).toHaveBeenCalledOnce()
      expect(vi.mocked(keysAPI.create).mock.calls[0].slice(0, 2)).toEqual(['My key', 5])
    })

    it('defaults to a provider with available groups and disables empty categories', async () => {
      getAvailableGroups.mockResolvedValue([availableGroups[5]])
      const wrapper = await openCreate()
      expect(wrapper.get<HTMLInputElement>('input[value="domestic"]').element.checked).toBe(true)
      expect(wrapper.get<HTMLInputElement>('input[value="anthropic"]').element.disabled).toBe(true)
      expect(optionIds(wrapper)).toEqual([6])
    })

    it('shows the empty state when no groups are available', async () => {
      getAvailableGroups.mockResolvedValue([])
      const wrapper = await openCreate()
      expect(wrapper.get('[data-tour="key-form-provider"]').text()).toContain('common.noGroupsAvailable')
      expect(optionIds(wrapper)).toEqual([])
      expect(wrapper.findAll<HTMLInputElement>('input[name="key-provider"]').every((input) => input.element.disabled)).toBe(true)
    })

    it('selects an available provider when groups arrive after opening', async () => {
      let resolveGroups!: (value: typeof availableGroups) => void
      getAvailableGroups.mockReturnValue(new Promise((resolve) => { resolveGroups = resolve }))
      const wrapper = await openCreate()
      resolveGroups([availableGroups[1]])
      await flushPromises()
      expect(wrapper.get<HTMLInputElement>('input[value="openai"]').element.checked).toBe(true)
      expect(optionIds(wrapper)).toEqual([2])
    })

    it('resets provider and group when reopening create, and preserves edit options', async () => {
      const wrapper = await openCreate()
      await chooseProvider(wrapper, 'domestic')
      await groupSelect(wrapper).vm.$emit('update:modelValue', 5)
      await wrapper.get('[data-test="close-dialog"]').trigger('click')
      await wrapper.get('[data-tour="keys-create-btn"]').trigger('click')
      expect(optionIds(wrapper)).toEqual([1])
      expect(groupSelect(wrapper).props('modelValue')).toBeNull()
      await wrapper.get('[data-test="close-dialog"]').trigger('click')
      await wrapper.get('[data-test="key-more"]').trigger('click')
      await getButtonByText(wrapper, 'common.edit').trigger('click')
      expect(wrapper.find('[data-tour="key-form-provider"]').exists()).toBe(false)
      expect(optionIds(wrapper)).toHaveLength(12)
    })
  })

  describe('usage stats loading', () => {
    it('renders the key list without waiting for slow usage stats', async () => {
      let resolveUsage!: (value: { stats: Record<string, unknown> }) => void
      getDashboardApiKeysUsage.mockReturnValue(new Promise((resolve) => { resolveUsage = resolve }))
      const wrapper = await mountView()

      expect(wrapper.findAll('[data-test="key-card"]')).toHaveLength(1)
      expect(wrapper.find('[data-test="usage-pending"]').exists()).toBe(true)

      resolveUsage({
        stats: { 1: { api_key_id: 1, today_actual_cost: 0.5, total_actual_cost: 12.25 } },
      })
      await flushPromises()
      expect(wrapper.find('[data-test="usage-pending"]').exists()).toBe(false)
      expect(wrapper.get('[data-test="usage"]').text()).toContain('$12.2500')
    })

    it('stops showing usage placeholders when the usage request fails', async () => {
      const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {})
      getDashboardApiKeysUsage.mockRejectedValue(new Error('timeout'))
      const wrapper = await mountView()

      expect(wrapper.find('[data-test="usage-pending"]').exists()).toBe(false)
      expect(wrapper.get('[data-test="usage"]').text()).toContain('$0.0000')
      consoleError.mockRestore()
    })
  })
})
