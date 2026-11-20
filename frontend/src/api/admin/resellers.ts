/**
 * Admin API: who is a reseller.
 */

import { apiClient } from '../client'

export interface ResellerEntry {
  user_id: number
  email: string
  since: string
  balance: number
  keys: number
  /** What their clients used at the clients' price, and what that cost the reseller. */
  client_used: number
  cost: number
}

export async function list(): Promise<ResellerEntry[]> {
  const { data } = await apiClient.get<ResellerEntry[]>('/admin/resellers')
  return data
}

export async function add(userId: number): Promise<void> {
  await apiClient.put(`/admin/resellers/${userId}`)
}

/** Keys the user already made keep working. */
export async function remove(userId: number): Promise<void> {
  await apiClient.delete(`/admin/resellers/${userId}`)
}

/** The second domain on which resellers' clients look up their key. An empty domain means none. */
export interface ResellerPortal {
  domain: string
  name: string
  contact: string
  logo: string
}

export async function getPortal(): Promise<ResellerPortal> {
  const { data } = await apiClient.get<ResellerPortal>('/admin/reseller-portal')
  return data
}

export async function setPortal(portal: ResellerPortal): Promise<ResellerPortal> {
  const { data } = await apiClient.put<ResellerPortal>('/admin/reseller-portal', portal)
  return data
}

export const resellersAPI = { list, add, remove, getPortal, setPortal }
export default resellersAPI
