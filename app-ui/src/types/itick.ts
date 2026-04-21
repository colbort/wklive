export interface PageReq {
  cursor?: number
  limit?: number
}

export interface ItickCategory {
  id: number
  categoryType: number
  categoryCode: string
  categoryName: string
  enabled: number
  appVisible: number
  sort: number
  icon: string
  remark: string
  createTimes: number
  updateTimes: number
}

export interface ItickProduct {
  id: number
  categoryType: number
  categoryCode: string
  categoryName: string
  market: string
  symbol: string
  code: string
  name: string
  displayName: string
  baseCoin: string
  quoteCoin: string
  enabled: number
  appVisible: number
  sort: number
  icon: string
  remark: string
  createTimes: number
  updateTimes: number
}

export interface Kline {
  categoryCode: string
  market: string
  symbol: string
  kType: number
  ts: number
  open: number
  high: number
  low: number
  close: number
  volume: number
  turnover: number
}

export interface DepthLevel {
  price: number
  volume: number
  position?: number
  originVolume?: number
}

export interface Depth {
  categoryCode: string
  market: string
  symbol: string
  asks: DepthLevel[]
  bids: DepthLevel[]
  ts: number
}

export interface Quote {
  categoryCode: string
  market: string
  symbol: string
  lastPrice: number
  openPrice: number
  highPrice: number
  lowPrice: number
  prevClosePrice: number
  changeValue: number
  changeRate: number
  volume: number
  turnover: number
  quoteTs: number
  tradeStatus: number
}

export interface ItickTenantCategory {
  id: number
  tenantId: number
  categoryId: number
  enabled: number
  appVisible: number
  sort: number
  remark: string
  createTimes: number
  updateTimes: number
  categoryType: number
  categoryCode: string
  categoryName: string
  icon: string
}

export interface ItickTenantProduct {
  id: number
  tenantId: number
  productId: number
  enabled: number
  appVisible: number
  sort: number
  remark: string
  createTimes: number
  updateTimes: number
  categoryType: number
  categoryCode: string
  categoryName: string
  market: string
  symbol: string
  code: string
  name: string
  displayName: string
  baseCoin: string
  quoteCoin: string
  icon: string
}

export interface ListVisibleCategoriesReq extends PageReq {
  tenantId: number
}

export type ListVisibleCategoriesResp = ItickTenantCategory[]

export interface ListVisibleProductsReq extends PageReq {
  categoryType?: number
  categoryCode?: string
  market?: string
  keyword?: string
  tenantId: number
}

export type ListVisibleProductsResp = ItickTenantProduct[]

export interface GetKlineReq {
  categoryCode: string
  market: string
  symbol: string
  kType: number
  endTs?: number
  limit?: number
}

export type GetKlineResp = Kline[]

export interface GetQuoteReq {
  categoryCode: string
  market: string
  symbol: string
}

export type GetQuoteResp = Quote

export interface MarketSymbol {
  categoryCode: string
  market: string
  symbol: string
}

export interface BatchGetQuoteReq {
  categoryCode?: string
  market?: string
  data: MarketSymbol[]
}

export type BatchGetQuoteResp = Quote[]

export type ItickWsTopic = 'quote' | 'depth' | 'tick' | 'kline'

export type ItickWsConnectionState = 'connecting' | 'open' | 'closed'

export interface ItickWsTopicConfig {
  topic: ItickWsTopic
  categoryCode: string
  symbol: string
  market: string
  interval?: string
}

export interface ItickWsSubscribeMessage {
  type: 'subscribe'
  topics: ItickWsTopicConfig[]
}

export interface ItickWsPingMessage {
  type: 'ping'
  clientTs: number
}

export interface ItickWsPongMessage {
  type: 'pong'
  clientTs: number
  serverTs: number
}

export interface QuotePayload {
  lastPrice: number
  open: number
  high: number
  low: number
  volume: number
  turnover: number
  ts: number
}

export interface TickPayload {
  lastPrice: number
  volume: number
  ts: number
}

export interface DepthPayload {
  asks: DepthLevel[]
  bids: DepthLevel[]
}

export interface KlinePayload {
  interval: string
  open: number
  high: number
  low: number
  close: number
  volume: number
  turnover: number
  ts: number
}

export interface ItickWsServerMessage<TPayload = unknown> {
  topic: ItickWsTopic
  categoryCode: string
  symbol: string
  market?: string
  interval?: string
  payload: TPayload
}
