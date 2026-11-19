/**
 * Garden: the bonus of a top-up is planted as a crop and harvested into the balance once ripe.
 */

import { apiClient } from './client'

export interface GardenSettings {
  enabled: boolean
  ripen_hours: number
}

export interface GardenCrop {
  id: number
  /** What was planted, in balance units. */
  amount: number
  stolen_amount: number
  status: 'growing' | 'harvested'
  planted_at: string
  ripe_at: string
  harvested_at?: string
  harvested_amount: number
}

export interface MyGarden extends GardenSettings {
  /** Server time, so countdowns do not depend on the visitor's clock. */
  now: string
  crops: GardenCrop[]
}

export async function getMyGarden(): Promise<MyGarden> {
  const { data } = await apiClient.get<MyGarden>('/user/garden')
  return data
}

export async function harvestCrop(id: number): Promise<{ amount: number }> {
  const { data } = await apiClient.post<{ amount: number }>(`/user/garden/crops/${id}/harvest`)
  return data
}

export async function getGardenSettings(): Promise<GardenSettings> {
  const { data } = await apiClient.get<GardenSettings>('/admin/garden-settings')
  return data
}

export async function updateGardenSettings(settings: GardenSettings): Promise<GardenSettings> {
  const { data } = await apiClient.put<GardenSettings>('/admin/garden-settings', settings)
  return data
}
