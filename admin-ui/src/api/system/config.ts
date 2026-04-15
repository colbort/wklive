import { get, post, put, del } from '@/utils/request'
import type {
  RespBase,
  SysConfigListReq,
  SysConfigItem,
  SysConfigCreateReq,
  SysConfigUpdateReq,
} from '@/services'

type OptionItem = {
  label: string
  value: string
  code?: string
}

type OptionGroup = {
  key: string
  label: string
  options: OptionItem[]
}

// ===== API 函数 =====

export function apiSysConfigList(params: SysConfigListReq): Promise<RespBase<SysConfigItem[]>> {
  return get<SysConfigItem[]>('/admin/system/configs', params)
}

export function apiSysConfigCreate(data: SysConfigCreateReq): Promise<RespBase> {
  return post('/admin/system/configs', data)
}

export function apiSysConfigUpdate(data: SysConfigUpdateReq): Promise<RespBase> {
  return put('/admin/system/configs', data)
}

export function apiSysConfigDelete(id: number): Promise<RespBase> {
  return del(`/admin/system/configs/${id}`)
}

export function apiSystemOptions(): Promise<RespBase<OptionGroup[]>> {
  return get<OptionGroup[]>('/admin/system/options')
}
