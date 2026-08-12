import type { RespBase, BaseService, OptionGroup } from '@/services'
import { getCoreOptions } from '@/stores/core'

import {
  apiMarketProductList,
  apiMarketProductCreate,
  apiMarketProductUpdate,
  apiMarketProductDetail,
  apiMarketProductKline,
  apiSyncMarketProductKlineHistory,
} from '@/api/market/products'

export type MarketProduct = {
  id: number
  categoryType: number
  categoryName: string
  categoryCode: string
  market: string
  symbol: string
  code: string
  name: string
  displayName: string
  baseCoin: string
  quoteCoin: string
  enabled: number
  appVisible: number
  syncPriority: number
  sort: number
  icon: string
  remark: string
  createTimes: number
  updateTimes: number
  exchange: string
  tradingCalendar?: MarketTradingCalendar
}

export type MarketTradingCalendar = {
  id: number
  categoryCode: string
  market: string
  exchange: string
  timezone: string
  tradingDayOffset: number
  weekStart: number
  productSpecific: boolean
  remark: string
  sessions: MarketTradingSession[]
}

export type MarketTradingSession = {
  id: number
  sessionType: string
  startTime: string
  endTime: string
  crossDay: boolean
  weekdayMask: number
  sort: number
}

export type CreateProductReq = {
  categoryType: number
  categoryName: string
  categoryCode: string
  market: string
  symbol: string
  code: string
  name: string
  displayName: string
  baseCoin: string
  quoteCoin: string
  enabled: number
  appVisible: number
  syncPriority: number
  sort: number
  icon: string
  remark: string
}

export type UpdateProductReq = {
  id: number
  name?: string
  displayName?: string
  baseCoin?: string
  quoteCoin?: string
  enabled?: number
  appVisible?: number
  syncPriority?: number
  sort?: number
  icon?: string
  remark?: string
}

export type ListProductsReq = {
  categoryType?: number
  market?: string
  symbol?: string
  keyword?: string
  enabled?: number // 0全部 1启用 2禁用
  appVisible?: number // 0全部 1显示 2隐藏
  cursor?: number
  limit?: number
}

export type GetProductKlineReq = {
  categoryCode: string
  market: string
  symbol: string
  kType: number
  endTs: number
  limit: number
}

export type SyncProductKlineHistoryReq = Omit<GetProductKlineReq, 'limit'>

export type SyncProductKlineHistoryResp = RespBase & {
  syncedCount: number
}

export type Kline = {
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

// ===== ITICK服务 =====

export class ProductsService implements BaseService {
  async getOptions(): Promise<RespBase<OptionGroup[]>> {
    return getCoreOptions()
  }

  async getList(params: ListProductsReq): Promise<RespBase<MarketProduct[]>> {
    return apiMarketProductList(params)
  }

  async create(params: CreateProductReq): Promise<RespBase> {
    return apiMarketProductCreate(params)
  }

  async update(id: string | number, params: Partial<UpdateProductReq>): Promise<RespBase> {
    return apiMarketProductUpdate({ id: Number(id), ...params })
  }

  async detail(id: number): Promise<RespBase<MarketProduct>> {
    return apiMarketProductDetail(id)
  }

  async kline(params: GetProductKlineReq): Promise<RespBase<Kline[]>> {
    return apiMarketProductKline(params)
  }

  async syncKlineHistory(params: SyncProductKlineHistoryReq): Promise<SyncProductKlineHistoryResp> {
    return apiSyncMarketProductKlineHistory(params)
  }
}

export const productsService = new ProductsService()
