import { http } from '@/utils/request'

export type MenuNode = { id: number; name: string; children?: MenuNode[] }
export type PermItem = { key: string; name: string; group?: string } // key 就是 sys:xxx:yyy

export async function apiMenuTree(): Promise<MenuNode[]> {
  return await http.get('/menus/tree')
}
export async function apiPermList(): Promise<PermItem[]> {
  return await http.get('/perms/list')
}
export async function apiRoleGrantDetail(roleId: number): Promise<{ menuIds: number[]; permKeys: string[] }> {
  return await http.get(`/roles/${roleId}/grant-detail`)
}