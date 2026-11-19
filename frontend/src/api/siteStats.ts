import { apiClient } from './client'

export interface SiteStats {
  total_tokens: number
  today_tokens: number
  total_requests: number
}

/** Site-wide usage totals. Public, no sign-in required. */
export async function getSiteStats(): Promise<SiteStats> {
  const { data } = await apiClient.get<SiteStats>('/site-stats')
  return data
}
