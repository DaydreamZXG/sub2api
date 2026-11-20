/**
 * Admin profit API: what upstream capacity cost (cost packs) against what users were charged.
 * Every amount is in yuan; dates are calendar days ("2026-03-01") in the site's timezone.
 */

import { apiClient } from '../client'

export interface CostPackPayment {
  id: number
  amount: number
  paid_at: string
  note: string
}

export interface CostPackAccount {
  id: number
  name: string
  platform: string
  deleted: boolean
}

/** Money actually paid for upstream capacity, tied to the accounts it bought. */
export interface CostPack {
  id: number
  name: string
  note: string
  /** 0 for a one-off pack; above 0, cycle_amount is paid again every cycle_days. */
  cycle_days: number
  cycle_amount: number
  started_at: string
  ended_at: string | null
  payments: CostPackPayment[]
  accounts: CostPackAccount[]
  paid: number
  income: number
  plan_usage: number
  profit: number
  requests: number
  ended: boolean
  next_cycle_at: string | null
}

export interface CostPackInput {
  name: string
  note: string
  cycle_days: number
  cycle_amount: number
  started_at: string
  ended_at: string | null
  account_ids: number[]
  /** Purchase price of a one-off pack; only read when the pack is created. */
  amount: number
}

export interface ProfitFigures {
  requests: number
  /** The real money in what users were charged from their balance. */
  income: number
  /** The rest of what they were charged: balance that was given, not paid for. */
  gifted: number
  /** What monthly-plan users used, at plan value. Not money received. */
  plan_usage: number
  rate_cost: number
  pack_cost: number
  profit: number
}

export interface ProfitReport {
  from: string
  to: string
  units_per_yuan: number
  /** What payment orders brought in and gave back in the period, per currency paid in; never added up across currencies. */
  cash: { currency: string; received: number; refunded: number }[]
  /** What agents were credited in the period; already taken off totals.profit. */
  agent_share: number
  totals: ProfitFigures
  months: (ProfitFigures & { month: string })[]
  /** group_id -1 carries pack cost no usage could be matched with; 0 is usage without a group. */
  groups: (ProfitFigures & { group_id: number; name: string })[]
  packs: (ProfitFigures & { pack_id: number; name: string })[]
  accounts: (ProfitFigures & {
    account_id: number
    name: string
    platform: string
    deleted: boolean
    pack_id: number
    pack_name: string
    rate_multiplier: number
  })[]
}

export async function getReport(from: string, to: string): Promise<ProfitReport> {
  const { data } = await apiClient.get<ProfitReport>('/admin/profit/report', { params: { from, to } })
  return data
}

export async function rebuild(): Promise<void> {
  await apiClient.post('/admin/profit/rebuild')
}

export async function listPacks(): Promise<CostPack[]> {
  const { data } = await apiClient.get<CostPack[]>('/admin/cost-packs')
  return data
}

export async function createPack(input: CostPackInput): Promise<void> {
  await apiClient.post('/admin/cost-packs', input)
}

export async function updatePack(id: number, input: CostPackInput): Promise<void> {
  await apiClient.put(`/admin/cost-packs/${id}`, input)
}

export async function deletePack(id: number): Promise<void> {
  await apiClient.delete(`/admin/cost-packs/${id}`)
}

export async function addPayment(id: number, amount: number, paidAt: string, note: string): Promise<void> {
  await apiClient.post(`/admin/cost-packs/${id}/payments`, { amount, paid_at: paidAt, note })
}

export async function deletePayment(id: number, paymentId: number): Promise<void> {
  await apiClient.delete(`/admin/cost-packs/${id}/payments/${paymentId}`)
}

/** Which credits besides paid orders count as real money. */
export interface ProfitSettings {
  manual_is_paid: boolean
  redeem_is_paid: boolean
}

export async function getSettings(): Promise<ProfitSettings> {
  const { data } = await apiClient.get<ProfitSettings>('/admin/profit/settings')
  return data
}

export async function updateSettings(settings: ProfitSettings): Promise<ProfitSettings> {
  const { data } = await apiClient.put<ProfitSettings>('/admin/profit/settings', settings)
  return data
}
