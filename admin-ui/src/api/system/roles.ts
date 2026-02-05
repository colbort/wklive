import { http } from '@/utils/request'

export type RoleItem = { id: number; name: string; code: string; status: number; remark?: string }

export function apiRoleList(params: { keyword?: string; status?: number; page?: number; pageSize?: number }) {
  return http.get('/admin/roles', { params })
}
