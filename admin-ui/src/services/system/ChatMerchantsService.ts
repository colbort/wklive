import type { BaseService, OptionGroup, RespBase } from '@/services'
import { getCoreOptions } from '@/stores/core'
import {
  apiSysChatMerchantCreate,
  apiSysChatMerchantDelete,
  apiSysChatMerchantDetail,
  apiSysChatMerchantList,
  apiSysChatMerchantUpdate,
} from '@/api/system/chat-merchants'

export type SysChatMerchantCreateReq = {
  merchantCode: string
  merchantName: string
  password: string
  enabled: number
  expireTime: number
  contactName: string
  contactPhone: string
  contactEmail: string
  remark: string
}

export type SysChatMerchantUpdateReq = {
  id: number
  merchantCode?: string
  merchantName?: string
  password?: string
  enabled?: number
  expireTime?: number
  contactName?: string
  contactPhone?: string
  contactEmail?: string
  remark?: string
}

export type SysChatMerchantItem = {
  id: number
  merchantCode: string
  merchantName: string
  enabled: number
  expireTime: number
  contactName: string
  contactPhone: string
  contactEmail: string
  remark: string
  createBy: string
  createTimes: number
  updateBy: string
  updateTimes: number
}

export type SysChatMerchantListReq = {
  keyword?: string
  enabled?: number
  merchantCode?: string
  merchantName?: string
  contactName?: string
  contactPhone?: string
  contactEmail?: string
  cursor?: number
  limit?: number
}

export type SysChatMerchantDetailReq = {
  id?: number
  merchantCode?: string
}

export class ChatMerchantsService implements BaseService {
  async getOptions(): Promise<RespBase<OptionGroup[]>> {
    return getCoreOptions()
  }

  async getList(params: SysChatMerchantListReq): Promise<RespBase<SysChatMerchantItem[]>> {
    return apiSysChatMerchantList(params)
  }

  async create(params: SysChatMerchantCreateReq): Promise<RespBase> {
    return apiSysChatMerchantCreate(params)
  }

  async update(id: string | number, params: Partial<SysChatMerchantUpdateReq>): Promise<RespBase> {
    return apiSysChatMerchantUpdate({ id: Number(id), ...params })
  }

  async delete(id: string | number): Promise<RespBase> {
    return apiSysChatMerchantDelete(Number(id))
  }

  async detail(params: SysChatMerchantDetailReq): Promise<RespBase<SysChatMerchantItem>> {
    return apiSysChatMerchantDetail(params)
  }
}

export const chatMerchantsService = new ChatMerchantsService()
