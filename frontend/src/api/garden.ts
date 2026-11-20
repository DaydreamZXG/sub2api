/**
 * Garden: the bonus of a top-up is planted as a crop and harvested into the balance once ripe.
 * Neighbours can water each other's crops and take a share of a ripe crop its owner left unharvested.
 */

import { apiClient } from './client'

export interface GardenSettings {
  enabled: boolean
  ripen_hours: number
  /** Hours after ripening during which only the owner can touch a crop. */
  protect_hours: number
  /** Share of a crop one neighbour takes, in percent. */
  steal_percent: number
  /** The most a crop can lose to neighbours, in percent. */
  max_stolen_percent: number
  daily_steal_limit: number
  /** Minutes one watering brings ripening forward. */
  water_minutes: number
  max_waters_per_crop: number
  neighbours_per_day: number
  /** Chance that a fed dog catches a thief, in percent: lowest and highest growth level. */
  dog_min_percent: number
  dog_max_percent: number
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
  dog_fed_today: boolean
  /** Chance that the user's own fed dog catches a thief, in percent. */
  dog_catch_percent: number
  steals_today: number
  /** Things neighbours did in the user's garden since the feed was last opened. */
  unread: number
}

export interface GardenNeighbour {
  id: number
  name: string
  /** The user's inviter or invitee: stays every day. */
  pinned: boolean
  crops: number
}

export interface GardenNeighbourCrop {
  id: number
  size: 'small' | 'medium' | 'large'
  planted_at: string
  ripe_at: string
  /** When neighbours may start taking from it. */
  open_at: string
  can_water: boolean
  can_steal: boolean
  /** Set once the user tried to take from this crop. */
  tried: '' | 'steal' | 'caught'
}

export interface GardenNeighbourGarden {
  neighbour: GardenNeighbour
  crops: GardenNeighbourCrop[]
  now: string
}

export interface GardenFeedItem {
  kind: 'water' | 'steal' | 'caught'
  amount: number
  /** True when the user did it in someone else's garden. */
  mine: boolean
  name: string
  created_at: string
}

export async function getGardenNeighbours(): Promise<GardenNeighbour[]> {
  const { data } = await apiClient.get<GardenNeighbour[]>('/user/garden/neighbours')
  return data
}

export async function getNeighbourGarden(id: number): Promise<GardenNeighbourGarden> {
  const { data } = await apiClient.get<GardenNeighbourGarden>(`/user/garden/neighbours/${id}`)
  return data
}

export async function waterCrop(id: number): Promise<{ ripe_at: string }> {
  const { data } = await apiClient.post<{ ripe_at: string }>(`/user/garden/crops/${id}/water`)
  return data
}

export async function stealCrop(id: number): Promise<{ caught: boolean; amount: number }> {
  const { data } = await apiClient.post<{ caught: boolean; amount: number }>(`/user/garden/crops/${id}/steal`)
  return data
}

export async function feedDog(): Promise<void> {
  await apiClient.post('/user/garden/dog/feed')
}

/** Reading the feed marks it as seen. */
export async function getGardenFeed(): Promise<GardenFeedItem[]> {
  const { data } = await apiClient.get<GardenFeedItem[]>('/user/garden/feed')
  return data
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
