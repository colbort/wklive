import { get, post, put } from '@/utils/request'

import type {
  RespBase,
  CreateTenantProductReq,
  UpdateTenantProductReq,
  BatchUpsertTenantProductsReq,
  ListTenantProductsReq,
  ItickTenantProduct,
  InitTenantItickDisplayReq,
  InitTenantItickDisplayResp,
} from '@/services'

export function apiItickTenantProductList(
  params: ListTenantProductsReq,
): Promise<RespBase<ItickTenantProduct[]>> {
  return get<ItickTenantProduct[]>('/admin/itick/tenant-products', params)
}

export function apiItickTenantProductCreate(params: CreateTenantProductReq): Promise<RespBase> {
  return post('/admin/itick/tenant-products', params)
}

export function apiItickTenantProductUpdate(params: UpdateTenantProductReq): Promise<RespBase> {
  return put('/admin/itick/tenant-products', params)
}

export function apiItickTenantProductBatchUpsert(
  params: BatchUpsertTenantProductsReq,
): Promise<RespBase> {
  return post('/admin/itick/tenant-products/batch', params)
}

export function apiItickTenantProductDetail(
  id: number,
  tenantId: number,
): Promise<RespBase<ItickTenantProduct>> {
  return get<ItickTenantProduct>(`/admin/itick/tenant-products/${id}`, { tenantId })
}

export function apiInitTenantItickDisplay(
  params: InitTenantItickDisplayReq,
): Promise<RespBase<InitTenantItickDisplayResp>> {
  return post<InitTenantItickDisplayResp>('/admin/itick/tenant-display/init', params)
}
