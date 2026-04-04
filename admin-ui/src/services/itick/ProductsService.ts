import type { RespBase, BaseService } from '@/services'

import {
  apiItickProductList,
  apiItickProductCreate,
  apiItickProductUpdate,
  apiItickProductDetail,
  apiItickProductKline,
} from '@/api/itick/products'

export type ItickProduct = {
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
  sort: number
  icon: string
  remark: string
  createTimes: number
  updateTimes: number
}

export type CreateProductReq = {
  categoryType: number
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
}

export type UpdateProductReq = {
  id: number
  name?: string
  displayName?: string
  baseCoin?: string
  quoteCoin?: string
  enabled?: number
  appVisible?: number
  sort?: number
  icon?: string
  remark?: string
}

export type ListProductsReq = {
  categoryType?: number
  market?: string
  keyword?: string
  enabled?: number // 0全部 1启用 2禁用
  appVisible?: number // 0全部 1显示 2隐藏
  cursor?: string | null
  limit?: number
}

export type GetProductKlineReq = {
  market: string
  symbol: string
  kType: number
  endTs: number
  limit: number
}

export type Kline = {
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

// ===== ITICK服务 =====

export class ProductsService implements BaseService {
  async getList(params: ListProductsReq): Promise<RespBase<ItickProduct[]>> {
    return apiItickProductList(params)
  }

  async create(params: CreateProductReq): Promise<RespBase> {
    return apiItickProductCreate(params)
  }

  async update(id: string | number, params: Partial<UpdateProductReq>): Promise<RespBase> {
    return apiItickProductUpdate({ id: Number(id), ...params })
  }

  async detail(id: number): Promise<RespBase<ItickProduct>> {
    return apiItickProductDetail(id)
  }

  async kline(params: GetProductKlineReq): Promise<RespBase<Kline[]>> {
    return apiItickProductKline(params)
  }
}

export const productsService = new ProductsService()
