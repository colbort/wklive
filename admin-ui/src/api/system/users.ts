import { get, post, put, del } from '@/utils/request'
import type { RespBase, SysUserItem, Google2FABindInitResp } from '@/services';


export function apiUserList(params: { keyword?: string; status?: number; page?: number; size?: number }): Promise<RespBase<SysUserItem[]>> {
  return get<SysUserItem[]>('/admin/users', { params })
}

export function apiUserDetail(id: number) : Promise<RespBase<SysUserItem>> {
  return get<SysUserItem>(`/admin/users/${id}`)
}

export function apiUserCreate(data: { username: string; password: string; nickname?: string; status?: number; roleIds?: number[] }): Promise<RespBase> {
  return post<RespBase>('/admin/users', data)
}

export function apiUserUpdate(data: { id: number; nickname?: string; status?: number; roleIds?: number[] }): Promise<RespBase> {
  return put<RespBase>('/admin/users', data)
}

export function apiUserDelete(id: number): Promise<RespBase> {
  return del<RespBase>(`/admin/users/${id}`)
}

export function apiChangeUserStatus(data: { id: number; status: number }): Promise<RespBase> {
  return post<RespBase>('/admin/users/status', data)
}

export function apiResetUserPwd(data: { id: number; password: string }): Promise<RespBase> {
  return post<RespBase>('/admin/users/resetPwd', data)
}

export function apiAssignUserRoles(data: { userId: number; roleIds: number[] }): Promise<RespBase> {
  return post<RespBase>('/admin/users/assignRoles', data)
}

// ---- Google 2FA ----
export function apiGoogle2faInit(data: { userId: number }): Promise<RespBase<Google2FABindInitResp>> {
  return post<Google2FABindInitResp>('/admin/users/google2fa/init', data)
}

export function apiGoogle2faEnable(data: { userId: number; code: string }): Promise<RespBase> {
  return post<RespBase>('/admin/users/google2fa/enable', data)
}

export function apiGoogle2faDisable(data: { userId: number; code?: string }): Promise<RespBase> {
  return post<RespBase>('/admin/users/google2fa/disable', data)
}

export function apiGoogle2faReset(data: { userId: number }): Promise<RespBase> {
  return post<RespBase>('/admin/users/google2fa/reset', data)
}
