import { get } from '@/utils/request'
// ===== 登录日志 =====
export function apiLoginLogList(params) {
  return get('/admin/logs/login', params)
}
// ===== 操作日志 =====
export function apiOpLogList(params) {
  return get('/admin/logs/op', params)
}
