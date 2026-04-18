import { get } from '@/utils/request'
import type { RespBase } from '@/services'

type SystemCore = {
  siteName: string
  siteLogo: string
  assetUrl: string
}

export function getSystemCore(): Promise<RespBase<SystemCore>> {
  return get<SystemCore>('/admin/system/core')
}
