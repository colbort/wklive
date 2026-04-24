import type { RespBase, BaseService, OptionGroup } from '@/services'

import {
  apiItickTenantCategoryBatchUpsert,
  apiItickTenantCategoryCreate,
  apiItickTenantCategoryDetail,
  apiItickTenantCategoryList,
  apiItickTenantCategoryUpdate,
} from '@/api/itick/tenant-categories'
import { apiOptions } from '@/api/itick/categories'

export type ItickTenantCategory = {
  id: number
  tenantId: number
  categoryId: number
  enabled: number
  appVisible: number
  sort: number
  remark: string
  createTimes: number
  updateTimes: number
  categoryType: number
  categoryName: string
  icon: string
}

export type CreateTenantCategoryReq = {
  tenantId: number
  categoryId: number
  enabled: number
  appVisible: number
  sort: number
  remark: string
}

export type UpdateTenantCategoryReq = {
  id: number
  tenantId: number
  enabled?: number
  appVisible?: number
  sort?: number
  remark?: string
}

export type TenantCategoryItem = {
  id?: number
  categoryId: number
  enabled: number
  appVisible: number
  sort: number
  remark: string
}

export type BatchUpsertTenantCategoriesReq = {
  tenantId: number
  data: TenantCategoryItem[]
}

export type ListTenantCategoriesReq = {
  tenantId: number
  categoryType?: number
  status?: number
  visibleStatus?: number
  cursor?: string | null
  limit?: number
}

// ===== ITICK服务 =====

export class TenantCategoriesService implements BaseService {
  async getOptions(): Promise<RespBase<OptionGroup[]>> {
    return apiOptions()
  }

  async getList(params: ListTenantCategoriesReq): Promise<RespBase<ItickTenantCategory[]>> {
    return apiItickTenantCategoryList(params)
  }

  async create(params: CreateTenantCategoryReq): Promise<RespBase> {
    return apiItickTenantCategoryCreate(params)
  }

  async update(id: string | number, params: Partial<UpdateTenantCategoryReq>): Promise<RespBase> {
    return apiItickTenantCategoryUpdate({ id: Number(id), ...params } as UpdateTenantCategoryReq)
  }

  async detail(id: number, tenantId: number): Promise<RespBase<ItickTenantCategory>> {
    return apiItickTenantCategoryDetail(id, tenantId)
  }

  async batchUpsert(params: BatchUpsertTenantCategoriesReq): Promise<RespBase> {
    return apiItickTenantCategoryBatchUpsert(params)
  }
}

export const tenantCategoriesService = new TenantCategoriesService()
