import { http } from '@/api/http'
import type { ApiResp } from '@/types/api'
import type { GetSystemCoreReq, GetSystemCoreResp } from '@/types/core'

export async function apiGetSystemCore(_: GetSystemCoreReq = {}): Promise<ApiResp<GetSystemCoreResp>> {
  const { data } = await http.get<ApiResp<GetSystemCoreResp>>('/system/core')
  return data
}
