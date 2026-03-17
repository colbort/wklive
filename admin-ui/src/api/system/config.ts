import { get, post, put, del } from '@/utils/request'
import type { RespBase, SysConfigListReq, SysConfigListResp, SysConfigCreateReq, SysConfigUpdateReq } from '@/services'


// ===== API 函数 =====

export function apiSysConfigList(params: SysConfigListReq): Promise<SysConfigListResp> {
  return get<SysConfigItem[]>('/admin/configs', { params })
}

export function apiSysConfigCreate(data: SysConfigCreateReq): Promise<RespBase> {
  return post<RespBase>('/admin/configs', data)
}

export function apiSysConfigUpdate(data: SysConfigUpdateReq): Promise<RespBase> {
  return put<RespBase>('/admin/configs', data)
}

export function apiSysConfigDelete(id: number): Promise<RespBase> {
  return del<RespBase>(`/admin/configs/${id}`)
}