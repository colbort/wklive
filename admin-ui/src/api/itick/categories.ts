import { get, post, put } from '@/utils/request'

import type {
  RespBase,
  ListCategoriesReq,
  ItickCategory,
  CreateCategoryReq,
  UpdateCategoryReq,
  SyncCategoryProductsReq,
  SyncCategoryProductsResp,
} from '@/services'

export function apiItickCategoryList(
  params: ListCategoriesReq,
): Promise<RespBase<ItickCategory[]>> {
  return get<ItickCategory[]>('/admin/itick/categories', params)
}

export function apiItickCategoryCreate(params: CreateCategoryReq): Promise<RespBase> {
  return post('/admin/itick/categories', params)
}

export function apiItickCategoryUpdate(params: UpdateCategoryReq): Promise<RespBase> {
  return put('/admin/itick/categories', params)
}

export function apiItickCategoryDetail(id: number): Promise<RespBase<ItickCategory>> {
  return get<ItickCategory>(`/admin/itick/categories/${id}`)
}

export function apiSyncCategoryProducts(
  params: SyncCategoryProductsReq,
): Promise<RespBase<SyncCategoryProductsResp>> {
  return post<SyncCategoryProductsResp>('/admin/itick/categories/sync-products', params)
}
