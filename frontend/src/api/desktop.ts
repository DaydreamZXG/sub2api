/**
 * Station side of the desktop assistant's connect protocol.
 * These paths and their { success, data } envelope are fixed by that protocol, so they sit outside /api/v1.
 */

import { apiClient } from './client'

export interface ConnectCode {
  code: string
  origin: string
  expiresAt: number
  ttlSeconds: number
}

export async function issueConnectCode(origin: string): Promise<ConnectCode> {
  const { data } = await apiClient.post<{ success: boolean; data: ConnectCode }>(
    '/api/desktop/v2/connect-codes',
    { origin },
    { baseURL: '' }
  )
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

export default { issueConnectCode, connectDeepLink, openDesktopApp }
