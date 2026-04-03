import { get, post, put, del } from '@/utils/request'
import type { RespBase, SysRole } from '@/services'

// ===== types =====

export function apiRoleList(params: {
  keyword?: string
  status?: number
  cursor?: string | null
  limit?: number
}): Promise<RespBase<SysRole[]>> {
  return get('/admin/roles', params)
}

export async function apiRoleCreate(data: any): Promise<RespBase> {
  // POST /roles
  return await post('/admin/system/roles', data)
}
export async function apiRoleUpdate(data: any): Promise<RespBase> {
  // PUT /roles
  return await put('/admin/system/roles', data)
}
export async function apiRoleDelete(id: number): Promise<RespBase> {
  // DELETE /roles/:id
  return await del(`/admin/system/roles/${id}`)
}
export async function apiRoleGrant(data: any): Promise<RespBase> {
  // POST /roles/grant
  return await post('/admin/system/roles/grant', data)
}
export async function apiRoleGrantDetail(roleId: number): Promise<RespBase> {
  return await get(`/admin/system/roles/${roleId}/grant`)
}
