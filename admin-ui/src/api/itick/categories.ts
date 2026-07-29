import { get, post, put } from '@/utils/request'

import type {
  RespBase,
  ListCategoriesReq,
  MarketCategory,
  CreateCategoryReq,
  UpdateCategoryReq,
  SyncCategoryProductsReq,
  SyncCategoryProductsResp,
} from '@/services'

export function apiMarketCategoryList(
  params: ListCategoriesReq,
): Promise<RespBase<MarketCategory[]>> {
  return get<MarketCategory[]>('/admin/market/categories', params)
}

export function apiMarketCategoryCreate(params: CreateCategoryReq): Promise<RespBase> {
  return post('/admin/market/categories', params)
}

export function apiMarketCategoryUpdate(params: UpdateCategoryReq): Promise<RespBase> {
  return put('/admin/market/categories', params)
}

export function apiMarketCategoryDetail(id: number): Promise<RespBase<MarketCategory>> {
  return get<MarketCategory>(`/admin/market/categories/${id}`)
}

export function apiSyncCategoryProducts(
  params: SyncCategoryProductsReq,
): Promise<RespBase<SyncCategoryProductsResp>> {
  return post<SyncCategoryProductsResp>('/admin/market/categories/sync-products', params)
}
