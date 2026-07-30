import { get, post, put } from '@/utils/request'

import type {
  RespBase,
  CreateTenantProductReq,
  UpdateTenantProductReq,
  BatchUpsertTenantProductsReq,
  ListTenantProductsReq,
  MarketTenantProduct,
  InitTenantMarketDisplayReq,
  InitTenantMarketDisplayResp,
} from '@/services'

export function apiMarketTenantProductList(
  params: ListTenantProductsReq,
): Promise<RespBase<MarketTenantProduct[]>> {
  return get<MarketTenantProduct[]>('/admin/market/tenant-products', params)
}

export function apiMarketTenantProductCreate(params: CreateTenantProductReq): Promise<RespBase> {
  return post('/admin/market/tenant-products', params)
}

export function apiMarketTenantProductUpdate(params: UpdateTenantProductReq): Promise<RespBase> {
  return put('/admin/market/tenant-products', params)
}

export function apiMarketTenantProductBatchUpsert(
  params: BatchUpsertTenantProductsReq,
): Promise<RespBase> {
  return post('/admin/market/tenant-products/batch', params)
}

export function apiMarketTenantProductDetail(
  id: number,
  tenantId: number,
): Promise<RespBase<MarketTenantProduct>> {
  return get<MarketTenantProduct>(`/admin/market/tenant-products/${id}`, { tenantId })
}

export function apiInitTenantMarketDisplay(
  params: InitTenantMarketDisplayReq,
): Promise<RespBase<InitTenantMarketDisplayResp>> {
  return post<InitTenantMarketDisplayResp>('/admin/market/tenant-display/init', params)
}
