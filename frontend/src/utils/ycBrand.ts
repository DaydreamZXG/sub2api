// The site's own name and wordmark, used until the operator sets a site name in admin settings.
export const YC_BRAND_NAME = '野菜'
export const YC_WORDMARK = 'YESCHOY'

const STOCK_NAMES = ['', 'Sub2API']
const STOCK_SUBTITLE = 'Subscription to API Conversion Platform'

export function isStockSiteName(name: string | null | undefined): boolean {
  return STOCK_NAMES.includes((name ?? '').trim())
}

export function ycSiteName(name: string | null | undefined): string {
  return isStockSiteName(name) ? YC_BRAND_NAME : (name ?? '').trim()
}

export function isStockSiteSubtitle(subtitle: string | null | undefined): boolean {
  const value = (subtitle ?? '').trim()
  return value === '' || value === STOCK_SUBTITLE
}
