import { http } from '@/utils/request'
import { ApiResp } from '../types'
import type { SimpleResp } from '../../types/system/role'

// ===== types =====

export function apiRoleList(params: { keyword?: string; status?: number; page?: number; size?: number }): Promise<ApiResp> {
  return http.get('/admin/roles', { params })
}

export async function apiRoleCreate(data: any): Promise<SimpleResp> {
  // POST /roles
  return await http.post('/admin/roles', data)
}
export async function apiRoleUpdate(data: any): Promise<SimpleResp> {
  // PUT /roles
  return await http.put('/admin/roles', data)
}
export async function apiRoleDelete(id: number): Promise<SimpleResp> {
  // DELETE /roles/:id
  return await http.delete(`/admin/roles/${id}`)
}
export async function apiRoleGrant(data: any): Promise<SimpleResp> {
  // POST /roles/grant
  return await http.post('/admin/roles/grant', data)
}
export async function apiRoleGrantDetail(roleId: number): Promise<{ menuIds: number[]; permKeys: string[] }> {
  return await http.get(`/admin/roles/${roleId}/grant`)
}