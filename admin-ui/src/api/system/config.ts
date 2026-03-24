import { get, post, put, del } from '@/utils/request'
import type {
  RespBase,
  SysConfigListReq,
  SysConfigItem,
  SysConfigCreateReq,
  SysConfigUpdateReq,
} from '@/services'

// ===== API 函数 =====

export function apiSysConfigList(params: SysConfigListReq): Promise<RespBase<SysConfigItem[]>> {
  return get<SysConfigItem[]>('/admin/configs', params)
}

export function apiSysConfigCreate(data: SysConfigCreateReq): Promise<RespBase> {
  return post('/admin/configs', data)
}

export function apiSysConfigUpdate(data: SysConfigUpdateReq): Promise<RespBase> {
  return put('/admin/configs', data)
}

export function apiSysConfigDelete(id: number): Promise<RespBase> {
  return del(`/admin/configs/${id}`)
}

export function apiSysConfigKeys(): Promise<RespBase<string[]>> {
  return get<string[]>('/admin/configs/keys')
}
