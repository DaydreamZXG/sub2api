/**
 * Station side of the desktop assistant's connect protocol.
 * These paths and their { success, data } envelope are fixed by that protocol, so they sit outside /api/v1.
 */

import { apiClient } from './client'
import { buildGatewayUrl } from './url'

export interface ConnectCode {
  code: string
  origin: string
  expiresAt: number
  ttlSeconds: number
}

const CONNECT_CODES_URL = buildGatewayUrl('/api/desktop/v2/connect-codes')

/**
 * The origin the desktop assistant will talk to: the backend's, which is not the page's
 * when the site and its API are served from different hosts.
 */
export function stationOrigin(): string {
  return new URL(CONNECT_CODES_URL, window.location.origin).origin
}

export async function issueConnectCode(): Promise<ConnectCode> {
  const { data } = await apiClient.post<{ success: boolean; data: ConnectCode }>(CONNECT_CODES_URL, {
    origin: stationOrigin()
  })
  return data.data
}

/** The link that hands a one-time code to the installed desktop assistant. */
export function connectDeepLink(connect: Pick<ConnectCode, 'code' | 'origin'>): string {
  return `yeschoy://connect?origin=${encodeURIComponent(connect.origin)}&code=${encodeURIComponent(connect.code)}`
}

/** Asks the operating system to open the desktop assistant with this code. */
export function openDesktopApp(connect: Pick<ConnectCode, 'code' | 'origin'>): void {
  window.location.href = connectDeepLink(connect)
}

export default { stationOrigin, issueConnectCode, connectDeepLink, openDesktopApp }
