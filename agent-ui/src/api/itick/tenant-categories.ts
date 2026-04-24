import { get, post, put } from '@/utils/request'

import type {
  RespBase,
  CreateTenantCategoryReq,
  UpdateTenantCategoryReq,
  BatchUpsertTenantCategoriesReq,
  ListTenantCategoriesReq,
  ItickTenantCategory,
} from '@/services'

export function apiItickTenantCategoryList(
  params: ListTenantCategoriesReq,
): Promise<RespBase<ItickTenantCategory[]>> {
  return get<ItickTenantCategory[]>('/admin/itick/tenant-categories', params)
}

export function apiItickTenantCategoryCreate(params: CreateTenantCategoryReq): Promise<RespBase> {
  return post('/admin/itick/tenant-categories', params)
}

export function apiItickTenantCategoryUpdate(params: UpdateTenantCategoryReq): Promise<RespBase> {
  return put('/admin/itick/tenant-categories', params)
}

export function apiItickTenantCategoryBatchUpsert(
  params: BatchUpsertTenantCategoriesReq,
): Promise<RespBase> {
  return post('/admin/itick/tenant-categories/batch', params)
}

export function apiItickTenantCategoryDetail(
  id: number,
  tenantId: number,
): Promise<RespBase<ItickTenantCategory>> {
  return get<ItickTenantCategory>('/admin/itick/tenant-categories/${id}', { tenantId })
}
