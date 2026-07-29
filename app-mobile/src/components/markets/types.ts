import type { MarketTenantProduct, QuotePayload } from '@/types/market'

export type MarketTopTab = 'watchlist' | 'markets' | 'chart'

export interface MarketTopTabItem {
  key: MarketTopTab
  label: string
}

export interface MarketRow {
  key: string
  product: MarketTenantProduct
  quote: QuotePayload | null
  changeRate: number
  direction: 'up' | 'down' | 'flat'
}
