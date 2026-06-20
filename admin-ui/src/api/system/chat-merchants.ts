import { get, post, put, del } from '@/utils/request'
import type {
  RespBase,
  SysChatMerchantCreateReq,
  SysChatMerchantDetailReq,
  SysChatMerchantItem,
  SysChatMerchantListReq,
  SysChatMerchantUpdateReq,
} from '@/services'

export function apiSysChatMerchantList(
  params: SysChatMerchantListReq,
): Promise<RespBase<SysChatMerchantItem[]>> {
  return get<SysChatMerchantItem[]>('/admin/system/chat-merchants', params)
}

export function apiSysChatMerchantCreate(data: SysChatMerchantCreateReq): Promise<RespBase> {
  return post('/admin/system/chat-merchant', data)
}

export function apiSysChatMerchantUpdate(data: SysChatMerchantUpdateReq): Promise<RespBase> {
  return put('/admin/system/chat-merchant', data)
}

export function apiSysChatMerchantDelete(id: number): Promise<RespBase> {
  return del(`/admin/system/chat-merchant/${id}`)
}

export function apiSysChatMerchantDetail(
  params: SysChatMerchantDetailReq,
): Promise<RespBase<SysChatMerchantItem>> {
  return get<SysChatMerchantItem>('/admin/system/chat-merchant/detail', params)
}
