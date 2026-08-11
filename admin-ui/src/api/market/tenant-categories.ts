import { get, post, put } from '@/utils/request'

import type {
  RespBase,
  CreateTenantCategoryReq,
  UpdateTenantCategoryReq,
  BatchUpsertTenantCategoriesReq,
  ListTenantCategoriesReq,
  MarketTenantCategory,
} from '@/services'

export function apiMarketTenantCategoryList(
  params: ListTenantCategoriesReq,
): Promise<RespBase<MarketTenantCategory[]>> {
  return get<MarketTenantCategory[]>('/admin/market/tenant-categories', params)
}

export function apiMarketTenantCategoryCreate(params: CreateTenantCategoryReq): Promise<RespBase> {
  return post('/admin/market/tenant-categories', params)
}

export function apiMarketTenantCategoryUpdate(params: UpdateTenantCategoryReq): Promise<RespBase> {
  return put('/admin/market/tenant-categories', params)
}

export function apiMarketTenantCategoryBatchUpsert(
  params: BatchUpsertTenantCategoriesReq,
): Promise<RespBase> {
  return post('/admin/market/tenant-categories/batch', params)
}

export function apiMarketTenantCategoryDetail(
  id: number,
  tenantId: number,
): Promise<RespBase<MarketTenantCategory>> {
  return get<MarketTenantCategory>(`/admin/market/tenant-categories/${id}`, { tenantId })
}
