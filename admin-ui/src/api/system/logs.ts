import { get } from '@/utils/request'
import { ApiResp } from '../types'
import { LoginLogItem, LoginLogListReq, LoginLogListResp, OpLogItem, OpLogListReq, OpLogListResp } from '@/types/system/logs'

// ===== 登录日志 =====
export function apiLoginLogList(params: LoginLogListReq): Promise<ApiResp<LoginLogItem[]>> {
  return get<LoginLogItem[]>('/admin/logs/login', { params })
}

// ===== 操作日志 =====
export function apiOpLogList(params: OpLogListReq): Promise<ApiResp<OpLogItem[]>> {
  return get<OpLogItem[]>('/admin/logs/op', { params })
}