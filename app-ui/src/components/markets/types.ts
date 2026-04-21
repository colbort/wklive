import type { ItickTenantProduct, QuotePayload } from '@/types/itick'

export type MarketTopTab = 'watchlist' | 'markets' | 'chart'

export interface MarketTopTabItem {
  key: MarketTopTab
  label: string
}

export interface MarketRow {
  key: string
  product: ItickTenantProduct
  quote: QuotePayload | null
  changeRate: number
  direction: 'up' | 'down' | 'flat'
}
