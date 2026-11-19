import { beforeEach, describe, expect, it, vi } from 'vitest'

const { post } = vi.hoisted(() => ({ post: vi.fn() }))

vi.mock('../client', () => ({ apiClient: { post } }))

const issued = { code: 'c'.repeat(43), origin: 'https://api.yeschoy.test', expiresAt: 0, ttlSeconds: 300 }

describe('desktop connect API', () => {
  beforeEach(() => {
    vi.resetModules()
    vi.unstubAllEnvs()
    post.mockReset()
    post.mockResolvedValue({ data: { success: true, data: issued } })
  })

  it('asks the backend host for a code bound to the backend origin when the API is on another host', async () => {
    vi.stubEnv('VITE_API_BASE_URL', 'https://api.yeschoy.test/api/v1')
    const { issueConnectCode, stationOrigin } = await import('../desktop')

    expect(stationOrigin()).toBe('https://api.yeschoy.test')
    await expect(issueConnectCode()).resolves.toEqual(issued)
    expect(post).toHaveBeenCalledWith('https://api.yeschoy.test/api/desktop/v2/connect-codes', {
      origin: 'https://api.yeschoy.test'
    })
  })

  it('uses the page origin when the API is served by the same host', async () => {
    const { issueConnectCode, stationOrigin } = await import('../desktop')

    expect(stationOrigin()).toBe(window.location.origin)
    await issueConnectCode()
    expect(post).toHaveBeenCalledWith(`${window.location.origin}/api/desktop/v2/connect-codes`, {
      origin: window.location.origin
    })
  })
})
