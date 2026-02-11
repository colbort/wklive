import { http } from '@/utils/request'


export function apiUserList(params: { keyword?: string; status?: number; page?: number; size?: number }) {
  return http.get('/admin/users', { params })
}

export function apiUserDetail(id: number) {
  return http.get(`/admin/users/${id}`)
}

export function apiUserCreate(data: { username: string; password: string; nickname?: string; status?: number; roleIds?: number[] }) {
  return http.post('/admin/users', data)
}

export function apiUserUpdate(data: { id: number; nickname?: string; status?: number; roleIds?: number[] }) {
  return http.put('/admin/users', data)
}

export function apiUserDelete(id: number) {
  return http.delete(`/admin/users/${id}`)
}

export function apiChangeUserStatus(data: { id: number; status: number }) {
  return http.post('/admin/users/status', data)
}

export function apiResetUserPwd(data: { id: number; password: string }) {
  return http.post('/admin/users/resetPwd', data)
}

export function apiAssignUserRoles(data: { userId: number; roleIds: number[] }) {
  return http.post('/admin/users/assignRoles', data)
}

// ---- Google 2FA ----
export function apiGoogle2faInit(data: { userId: number }) {
  return http.post('/admin/users/google2fa/init', data)
}

export function apiGoogle2faEnable(data: { userId: number; code: string }) {
  return http.post('/admin/users/google2fa/enable', data)
}

export function apiGoogle2faDisable(data: { userId: number; code?: string }) {
  return http.post('/admin/users/google2fa/disable', data)
}

export function apiGoogle2faReset(data: { userId: number }) {
  return http.post('/admin/users/google2fa/reset', data)
}
