import { http } from './http'
import type { ApiResp, OptionsGroup, RespBase } from '../types/api'
import type { GetSystemCoreReq, GetSystemCoreResp } from '../types/core'

export async function apiGetSystemCore(_: GetSystemCoreReq = {}): Promise<ApiResp<GetSystemCoreResp>> {
  const { data } = await http.get<ApiResp<GetSystemCoreResp>>('/system/core')
  return data
}

type CoreOptionsResp = RespBase & { data: OptionsGroup[] }

let coreOptionsPromise: Promise<CoreOptionsResp> | null = null

export function apiGetCoreOptions(): Promise<CoreOptionsResp> {
  if (!coreOptionsPromise) {
    coreOptionsPromise = apiGetSystemCore().then((res) => ({
      ...res,
      data: res.data?.options || [],
    }))
  }

  return coreOptionsPromise
}
