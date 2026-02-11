import { http } from '@/utils/request'
import { ApiResp } from '../types'

// ===== types =====
export type SysRole = {
  id: number
  name: string
  code: string
  remark?: string
  status?: number
  isSuper?: boolean // 可选：如果后端有的话
}

export type RoleListResp = {
  list: SysRole[]
  total: number
}

export type SimpleResp = { code: number; msg: string }

export type RoleItem = { id: number; name: string; code: string; status: number; remark?: string }

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