import type { RespBase, BaseService } from '@/services'

import {
  apiSysTenantList,
  apiSysTenantCreate,
  apiSysTenantUpdate,
  apiSysTenantDelete,
  apiSysTenantDetail,
} from '@/api/system/tenants'

export type SysTenantCreateReq = {
  tenantCode: string
  tenantName: string
  status: number
  expireTime: number
  contactName: string
  contactPhone: string
  remark: string
}

export type SysTenantUpdateReq = {
  id: number
  tenantCode?: string
  tenantName?: string
  status?: number
  expireTime?: number
  contactName?: string
  contactPhone?: string
  remark?: string
}

export type SysTenantItem = {
  id: number
  tenantCode: string
  tenantName: string
  status: number
  expireTime: number
  contactName: string
  contactPhone: string
  remark: string
  createTimes: number
  updateTimes: number
}

export type SysTenantListReq = {
  keyword?: string
  status?: number
  tenantCode?: string
  tenantName?: string
  contactName?: string
  contactPhone?: string
  cursor?: string | null
  limit?: number
}

export type SysTenantDetailReq = {
  tenantId?: number
  tenantCode?: string
}

// ========= 租户服务 =========
export class TenantsService implements BaseService {
  async getList(params: SysTenantListReq): Promise<RespBase<SysTenantItem[]>> {
    return apiSysTenantList(params)
  }

  async create(params: SysTenantCreateReq): Promise<RespBase> {
    return apiSysTenantCreate(params)
  }

  async update(id: string | number, params: Partial<SysTenantUpdateReq>): Promise<RespBase> {
    return apiSysTenantUpdate({ id: Number(id), ...params })
  }

  async delete(id: string | number): Promise<RespBase> {
    return apiSysTenantDelete(Number(id))
  }

  async detail(params: SysTenantDetailReq): Promise<RespBase<SysTenantItem>> {
    return apiSysTenantDetail(params)
  }
}

export const tenantsService = new TenantsService()
