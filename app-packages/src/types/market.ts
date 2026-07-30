export interface PageReq {
  cursor?: number
  limit?: number
}

export interface MarketCategory {
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

export interface MarketProduct {
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
  open: string
  high: string
  low: string
  close: string
  volume: string
  turnover: string
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
  lastPrice: string
  openPrice: string
  highPrice: string
  lowPrice: string
  prevClosePrice: string
  changeValue: string
  changeRate: string
  volume: string
  turnover: string
  quoteTs: number
  tradeStatus: number
}

export interface MarketTenantCategory {
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

export interface MarketTenantProduct {
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
  tenantCode: string
}

export type ListVisibleCategoriesResp = MarketTenantCategory[]

export interface ListVisibleProductsReq extends PageReq {
  categoryType?: number
  categoryCode?: string
  market?: string
  keyword?: string
  tenantCode: string
}

export type ListVisibleProductsResp = MarketTenantProduct[]

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

export type MarketWsTopic = 'quote' | 'depth' | 'tick' | 'kline'

export type MarketWsConnectionState = 'connecting' | 'open' | 'closed'

export interface MarketWsTopicConfig {
  topic: MarketWsTopic
  categoryCode: string
  symbol: string
  market: string
  interval?: string
}

export interface MarketWsSubscribeMessage {
  type: 'subscribe'
  topics: MarketWsTopicConfig[]
}

export interface MarketWsPingMessage {
  type: 'ping'
  clientTs: number
}

export interface MarketWsPongMessage {
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

export interface RawDepthPayload {
  asks?: DepthLevel[]
  bids?: DepthLevel[]
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

export interface MarketWsServerMessage<TPayload = unknown> {
  topic: MarketWsTopic
  categoryCode: string
  symbol: string
  market?: string
  interval?: string
  payload: TPayload
}
