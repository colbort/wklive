import type { RespBase, BaseService, OptionGroup } from '@/services'

import {
  apiInitTenantItickDisplay,
  apiItickTenantProductBatchUpsert,
  apiItickTenantProductCreate,
  apiItickTenantProductDetail,
  apiItickTenantProductList,
  apiItickTenantProductUpdate,
} from '@/api/itick/tenant-products'
import { apiOptions } from '@/api/itick/categories'

export type ItickTenantProduct = {
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
  tenantId: number
  categoryType?: number
  market?: string
  keyword?: string
  status?: number
  visibleStatus?: number
  cursor?: string | null
  limit?: number
}

export type InitTenantItickDisplayReq = {
  tenantId: number
  overwrite: number
}

export type InitTenantItickDisplayResp = {
  categoryCount: number
  productCount: number
}

// ===== ITICK服务 =====

export class TenantProductsService implements BaseService {
  async getOptions(): Promise<RespBase<OptionGroup[]>> {
    return apiOptions()
  }

  async getList(params: ListTenantProductsReq): Promise<RespBase<ItickTenantProduct[]>> {
    return apiItickTenantProductList(params)
  }

  async create(params: CreateTenantProductReq): Promise<RespBase> {
    return apiItickTenantProductCreate(params)
  }

  async update(id: string | number, params: Partial<UpdateTenantProductReq>): Promise<RespBase> {
    return apiItickTenantProductUpdate({ id: Number(id), ...params } as UpdateTenantProductReq)
  }

  async detail(id: number, tenantId: number): Promise<RespBase<ItickTenantProduct>> {
    return apiItickTenantProductDetail(id, tenantId)
  }

  async batchUpsert(params: BatchUpsertTenantProductsReq): Promise<RespBase> {
    return apiItickTenantProductBatchUpsert(params)
  }

  async initDisplay(
    params: InitTenantItickDisplayReq,
  ): Promise<RespBase<InitTenantItickDisplayResp>> {
    return apiInitTenantItickDisplay(params)
  }
}

export const tenantProductsService = new TenantProductsService()
