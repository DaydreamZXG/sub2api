/**
 * A reseller's client keys: keys handed to their own clients, with a budget in the client's price.
 */

import { apiClient } from './client'

export interface ResellerClientKey {
  id: number
  name: string
  key: string
  group_id: number | null
  group_name: string
  status: string
  /** A call counts against the budget at what it cost the reseller times this. Fixed at creation. */
  markup: number
  /** Budget and client_used are in the client's price; cost is what the same usage cost the reseller. */
  budget: number
  client_used: number
  cost: number
  margin: number
  expires_at: string | null
  last_used_at: string | null
  created_at: string
}

export interface ResellerOverview {
  balance: number
  client_used: number
  cost: number
  margin: number
  /** What the unused budgets of the usable keys would cost the reseller. */
  reserved: number
  keys: ResellerClientKey[]
}

export interface ClientKeyInput {
  name: string
  group_id: number | null
  budget: number
  markup: number
  expires_in_days?: number
}

/** Null for a user who is not a reseller. */
export async function getReseller(): Promise<ResellerOverview | null> {
  const { data } = await apiClient.get<{ reseller: ResellerOverview | null }>('/user/reseller')
  return data.reseller
}

export async function createClientKey(input: ClientKeyInput): Promise<{ id: number; key: string }> {
  const { data } = await apiClient.post<{ id: number; key: string }>('/user/reseller/keys', input)
  return data
}

export const resellerAPI = { getReseller, createClientKey }
export default resellerAPI
