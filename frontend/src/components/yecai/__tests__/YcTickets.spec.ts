import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import YcTickets from '../YcTickets.vue'

const { api, showError } = vi.hoisted(() => ({
  api: { list: vi.fn(), create: vi.fn(), getById: vi.fn(), reply: vi.fn(), close: vi.fn() },
  showError: vi.fn()
}))

vi.mock('@/api/tickets', () => ({
  default: api,
  TICKET_CATEGORIES: ['usage', 'billing', 'account', 'other'],
  TICKET_SUBJECT_MAX: 200,
  TICKET_CONTENT_MAX: 5000
}))
vi.mock('@/stores', () => ({ useAppStore: () => ({ showError }) }))
vi.mock('vue-i18n', async (importOriginal) => ({
  ...(await importOriginal<typeof import('vue-i18n')>()),
  useI18n: () => ({ t: (key: string) => key })
}))

const ticket = (overrides = {}) => ({
  id: 1,
  subject: 'Cannot connect',
  category: 'usage',
  status: 'closed',
  last_message_at: '2026-01-01T00:00:00Z',
  created_at: '2026-01-01T00:00:00Z',
  ...overrides
})
const message = (id: number, author_type: string, content: string) => ({ id, author_type, content, created_at: '2026-01-01T00:00:00Z' })

const mountTickets = async () => {
  const wrapper = mount(YcTickets)
  await flushPromises()
  return wrapper
}

describe('YcTickets', () => {
  beforeEach(() => {
    Object.values(api).forEach((fn) => fn.mockReset())
    showError.mockReset()
    api.list.mockResolvedValue({ items: [], total: 0, page: 1, page_size: 20, pages: 0 })
  })

  it('submits a new ticket and opens its conversation', async () => {
    api.create.mockResolvedValue(ticket({ id: 7, status: 'open' }))
    api.getById.mockResolvedValue({ ...ticket({ id: 7, status: 'open' }), messages: [message(1, 'user', 'It says 401')] })
    const wrapper = await mountTickets()
    api.list.mockResolvedValue({ items: [ticket({ id: 7, status: 'open' })], total: 1, page: 1, page_size: 20, pages: 1 })

    await wrapper.get('.yc-tk-head button').trigger('click')
    await wrapper.get('.yc-tk-form input').setValue('Cannot connect')
    await wrapper.get('.yc-tk-form textarea').setValue('It says 401')
    await wrapper.get('.yc-tk-form').trigger('submit')
    await flushPromises()

    expect(api.create).toHaveBeenCalledWith({ subject: 'Cannot connect', category: 'usage', content: 'It says 401' })
    expect(wrapper.find('.yc-tk-form').exists()).toBe(false)
    expect(wrapper.findAll('.yc-tk-msg').map((m) => m.text())).toEqual([expect.stringContaining('It says 401')])
  })

  it('shows a closed ticket as waiting for a reply again once the user writes back', async () => {
    api.list.mockResolvedValue({ items: [ticket()], total: 1, page: 1, page_size: 20, pages: 1 })
    api.getById.mockResolvedValue({ ...ticket(), messages: [message(1, 'user', 'help'), message(2, 'staff', 'fixed')] })
    api.reply.mockResolvedValue(message(3, 'user', 'still broken'))
    const wrapper = await mountTickets()

    await wrapper.get('.yc-tk-row').trigger('click')
    await flushPromises()
    expect(wrapper.get('.yc-tk-row i').text()).toBe('yc.tickets.status.closed')

    await wrapper.get('.yc-tk-reply textarea').setValue('still broken')
    await wrapper.get('.yc-tk-reply').trigger('submit')
    await flushPromises()

    expect(api.reply).toHaveBeenCalledWith(1, 'still broken')
    expect(wrapper.get('.yc-tk-row i').text()).toBe('yc.tickets.status.open')
    expect(wrapper.findAll('.yc-tk-msg')).toHaveLength(3)
  })

  it('reloads the first page after a new ticket so later pages do not repeat or skip tickets', async () => {
    api.list.mockResolvedValue({ items: [ticket({ id: 2 }), ticket({ id: 1 })], total: 3, page: 1, page_size: 2, pages: 2 })
    const wrapper = await mountTickets()
    expect(wrapper.find('.yc-tk-more').exists()).toBe(true)

    api.create.mockResolvedValue(ticket({ id: 4, status: 'open' }))
    api.list.mockResolvedValue({ items: [ticket({ id: 4, status: 'open' }), ticket({ id: 2 })], total: 4, page: 1, page_size: 2, pages: 2 })
    api.getById.mockResolvedValue({ ...ticket({ id: 4, status: 'open' }), messages: [message(1, 'user', 'new')] })
    await wrapper.get('.yc-tk-head button').trigger('click')
    await wrapper.get('.yc-tk-form input').setValue('New one')
    await wrapper.get('.yc-tk-form textarea').setValue('new')
    await wrapper.get('.yc-tk-form').trigger('submit')
    await flushPromises()
    expect(wrapper.findAll('.yc-tk-item')).toHaveLength(2)

    api.list.mockResolvedValue({ items: [ticket({ id: 1 }), ticket({ id: 0 })], total: 4, page: 2, page_size: 2, pages: 2 })
    await wrapper.get('.yc-tk-more').trigger('click')
    await flushPromises()

    expect(api.list).toHaveBeenLastCalledWith(2)
    expect(wrapper.findAll('.yc-tk-item')).toHaveLength(4)
    expect(wrapper.find('.yc-tk-more').exists()).toBe(false)
  })

  it('keeps the draft and reports the error when submitting fails', async () => {
    api.create.mockRejectedValue(new Error('boom'))
    const wrapper = await mountTickets()

    await wrapper.get('.yc-tk-head button').trigger('click')
    await wrapper.get('.yc-tk-form input').setValue('Cannot connect')
    await wrapper.get('.yc-tk-form textarea').setValue('It says 401')
    await wrapper.get('.yc-tk-form').trigger('submit')
    await flushPromises()

    expect(showError).toHaveBeenCalledTimes(1)
    expect((wrapper.get('.yc-tk-form input').element as HTMLInputElement).value).toBe('Cannot connect')
  })
})
