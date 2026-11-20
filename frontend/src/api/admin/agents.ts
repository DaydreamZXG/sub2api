/**
 * Admin agents API: inviters who are paid a share of what their users spend, and their payout requests.
 * Amounts are in balance units, like the invite quota they are credited to.
 */

import { apiClient } from '../client'

/** From `from` of the month's spending (balance units) on, the agent earns `percent` instead. */
export interface AgentTier {
  from: number
  percent: number
}

export interface AgentEntry {
  user_id: number
  email: string
  /** The share below the first tier, or the only one. */
  share_percent: number
  tiers: AgentTier[]
  /** Their users' spending this month that counted, and the share in force for what comes next. */
  month_spent: number
  current_percent: number
  since: string | null
  invitees: number
  /** Everything credited to this agent so far. */
  shared: number
  /** What can be paid out or moved to balance now. */
  available: number
}

export type AgentPayoutStatus = 'pending' | 'paid' | 'rejected'

export interface AgentPayout {
  id: number
  user_id: number
  email: string
  amount: number
  payee_note: string
  status: AgentPayoutStatus
  admin_note: string
  created_at: string
  handled_at: string | null
}

export async function list(): Promise<AgentEntry[]> {
  const { data } = await apiClient.get<AgentEntry[]>('/admin/agents')
  return data
}

/** `tiers` replaces the agent's ladder; an empty list means one flat share. */
export async function set(userId: number, sharePercent: number, tiers: AgentTier[] = []): Promise<void> {
  await apiClient.put(`/admin/agents/${userId}`, { share_percent: sharePercent, tiers })
}

export async function remove(userId: number): Promise<void> {
  await apiClient.delete(`/admin/agents/${userId}`)
}

export async function listPayouts(status: AgentPayoutStatus | '' = ''): Promise<AgentPayout[]> {
  const { data } = await apiClient.get<AgentPayout[]>('/admin/agents/payouts', { params: { status } })
  return data
}

export async function markPaid(id: number, adminNote = ''): Promise<void> {
  await apiClient.post(`/admin/agents/payouts/${id}/paid`, { admin_note: adminNote })
}

export async function reject(id: number, adminNote = ''): Promise<void> {
  await apiClient.post(`/admin/agents/payouts/${id}/reject`, { admin_note: adminNote })
}

/** A user under an agent. `attached` is true when an admin put them there; `joined_at` is then when. */
export interface AgentUser {
  user_id: number
  email: string
  joined_at: string
  attached: boolean
  /** Their spending that counted for the agent's share, and the share it earned. */
  spent: number
  share: number
}

export async function listUsers(agentId: number): Promise<AgentUser[]> {
  const { data } = await apiClient.get<AgentUser[]>(`/admin/agents/${agentId}/users`)
  return data
}

/** Puts an existing user under the agent; the agent earns on their spending from now on. */
export async function attachUser(agentId: number, userId: number): Promise<void> {
  await apiClient.put(`/admin/agents/${agentId}/users/${userId}`)
}

export async function detachUser(agentId: number, userId: number): Promise<void> {
  await apiClient.delete(`/admin/agents/${agentId}/users/${userId}`)
}
