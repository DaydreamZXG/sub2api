/**
 * Growth levels: cumulative top-ups make the app line cheaper.
 */

import { apiClient } from './client'

export interface GrowthLevel {
  name: string
  /** Cumulative top-ups needed to reach the level, in balance units. */
  threshold: number
  /** Share taken off the app line's price, 0-0.9. */
  discount: number
}

export interface GrowthLevels {
  enabled: boolean
  levels: GrowthLevel[]
}

export interface MyGrowth extends GrowthLevels {
  /** Index into `levels`, or -1 when growth levels are off. */
  level_index: number
  total_recharged: number
  /** An admin set this user's own price on an app line; it replaces the level discount there. */
  manual_price?: boolean
}

export async function getMyGrowth(): Promise<MyGrowth> {
  const { data } = await apiClient.get<MyGrowth>('/user/growth')
  return data
}

export async function getGrowthLevels(): Promise<GrowthLevels> {
  const { data } = await apiClient.get<GrowthLevels>('/admin/growth-levels')
  return data
}

export async function updateGrowthLevels(levels: GrowthLevels): Promise<GrowthLevels> {
  const { data } = await apiClient.put<GrowthLevels>('/admin/growth-levels', levels)
  return data
}
