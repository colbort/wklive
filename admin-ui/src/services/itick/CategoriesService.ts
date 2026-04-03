import type { RespBase, BaseService } from '@/services'

import {
  apiItickCategoryList,
  apiItickCategoryCreate,
  apiItickCategoryUpdate,
  apiItickCategoryDetail,
  apiSyncCategoryProducts,
} from '@/api/itick/categories'

export type ItickCategory = {
  id: number
  categoryType: number
  categoryCode: string
  categoryName: string
  enabled: number
  appVisible: number
  sort: number
  icon: string
  remark: string
  createTime: number
  updateTime: number
}

export type ListCategoriesReq = {
  categoryType?: number
  enabled?: number // 0全部 1启用 2禁用
  appVisible?: number // 0全部 1启用 2禁用
  cursor?: string | null
  limit?: number
}

export type CreateCategoryReq = {
  categoryType: number
  categoryName: string
  enabled: number
  appVisible: number
  sort: number
  icon: string
  remark: string
}

export type UpdateCategoryReq = {
  id: number
  categoryName?: string
  enabled?: number
  appVisible?: number
  sort?: number
  icon?: string
  remark?: string
}

export type SyncCategoryProductsReq = {
  id: number
}

export type SyncCategoryProductsResp = {
  taskNo: string
}

// ===== ITICK服务 =====

export class CategoriesService implements BaseService {
  async getList(params: ListCategoriesReq): Promise<RespBase<ItickCategory[]>> {
    return apiItickCategoryList(params)
  }

  async create(params: CreateCategoryReq): Promise<RespBase> {
    return apiItickCategoryCreate(params)
  }

  async update(id: string | number, params: Partial<UpdateCategoryReq>): Promise<RespBase> {
    return apiItickCategoryUpdate({ id: Number(id), ...params })
  }

  async detail(id: number): Promise<RespBase<ItickCategory>> {
    return apiItickCategoryDetail(id)
  }

  async syncProducts(params: SyncCategoryProductsReq): Promise<RespBase<SyncCategoryProductsResp>> {
    return apiSyncCategoryProducts(params)
  }
}

export const categoriesService = new CategoriesService()
