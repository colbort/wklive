import { get, post, put, del } from '@/utils/request'
import { ApiResp } from '../types'
import type { SysRole } from '../../types/system/roles'

// ===== types =====

export function apiRoleList(params: { keyword?: string; status?: number; page?: number; size?: number }): Promise<ApiResp<SysRole[]>> {
  return get('/admin/roles', { params })
}

export async function apiRoleCreate(data: any): Promise<ApiResp> {
  // POST /roles
  return await post('/admin/roles', data)
}
export async function apiRoleUpdate(data: any): Promise<ApiResp> {
  // PUT /roles
  return await put('/admin/roles', data)
}
export async function apiRoleDelete(id: number): Promise<ApiResp> {
  // DELETE /roles/:id
  return await del(`/admin/roles/${id}`)
}
export async function apiRoleGrant(data: any): Promise<ApiResp> {
  // POST /roles/grant
  return await post('/admin/roles/grant', data)
}
export async function apiRoleGrantDetail(roleId: number): Promise<ApiResp> {
  return await get(`/admin/roles/${roleId}/grant`)
}