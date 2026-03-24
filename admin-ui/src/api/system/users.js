import { get, post, put, del } from '@/utils/request'
export function apiUserList(params) {
  return get('/admin/users', params)
}
export function apiUserDetail(id) {
  return get(`/admin/users/${id}`)
}
export function apiUserCreate(data) {
  return post('/admin/users', data)
}
export function apiUserUpdate(data) {
  return put('/admin/users', data)
}
export function apiUserDelete(id) {
  return del(`/admin/users/${id}`)
}
export function apiChangeUserStatus(data) {
  return post('/admin/users/status', data)
}
export function apiResetUserPwd(data) {
  return post('/admin/users/resetPwd', data)
}
export function apiAssignUserRoles(data) {
  return post('/admin/users/assignRoles', data)
}
// ---- Google 2FA ----
export function apiGoogle2faInit(data) {
  return post('/admin/users/google2fa/init', data)
}
export function apiGoogle2faBind(data) {
  return post('/admin/users/google2fa/bind', data)
}
export function apiGoogle2faEnable(data) {
  return post('/admin/users/google2fa/enable', data)
}
export function apiGoogle2faDisable(data) {
  return post('/admin/users/google2fa/disable', data)
}
export function apiGoogle2faReset(data) {
  return post('/admin/users/google2fa/reset', data)
}
