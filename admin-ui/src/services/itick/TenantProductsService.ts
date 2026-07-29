import type { RespBase, BaseService, OptionGroup } from '@/services'
import { getCoreOptions } from '@/stores/core'

import {
  apiInitTenantMarketDisplay,
  apiMarketTenantProductBatchUpsert,
  apiMarketTenantProductCreate,
  apiMarketTenantProductDetail,
  apiMarketTenantProductList,
  apiMarketTenantProductUpdate,
} from '@/api/market/tenant-products'

export type MarketTenantProduct = {
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

export type CreateTenantProductReq = {
  tenantId: number
  productId: number
  enabled: number
  appVisible: number
  sort: number
  remark: string
}

export type UpdateTenantProductReq = {
  id: number
  tenantId: number
  enabled?: number
  appVisible?: number
  displayName?: string
  sort?: number
  remark?: string
}

export type TenantProductItem = {
  id?: number
  productId: number
  enabled: number
  appVisible: number
  sort: number
  remark: string
}

export type BatchUpsertTenantProductsReq = {
  tenantId: number
  data: TenantProductItem[]
}

export type ListTenantProductsReq = {
  tenantId?: number
  categoryType?: number
  market?: string
  keyword?: string
  enabled?: number
  appVisible?: number
  cursor?: number
  limit?: number
}

export type InitTenantMarketDisplayReq = {
  tenantId: number
  overwrite: number
}

export type InitTenantMarketDisplayResp = {
  categoryCount: number
  productCount: number
}

// ===== ITICK服务 =====

export class TenantProductsService implements BaseService {
  async getOptions(): Promise<RespBase<OptionGroup[]>> {
    return getCoreOptions()
  }

  async getList(params: ListTenantProductsReq): Promise<RespBase<MarketTenantProduct[]>> {
    return apiMarketTenantProductList(params)
  }

  async create(params: CreateTenantProductReq): Promise<RespBase> {
    return apiMarketTenantProductCreate(params)
  }

  async update(id: string | number, params: Partial<UpdateTenantProductReq>): Promise<RespBase> {
    return apiMarketTenantProductUpdate({ id: Number(id), ...params } as UpdateTenantProductReq)
  }

  async detail(id: number, tenantId: number): Promise<RespBase<MarketTenantProduct>> {
    return apiMarketTenantProductDetail(id, tenantId)
  }

  async batchUpsert(params: BatchUpsertTenantProductsReq): Promise<RespBase> {
    return apiMarketTenantProductBatchUpsert(params)
  }

  async initDisplay(
    params: InitTenantMarketDisplayReq,
  ): Promise<RespBase<InitTenantMarketDisplayResp>> {
    return apiInitTenantMarketDisplay(params)
  }
}

export const tenantProductsService = new TenantProductsService()
