import { http } from '@/utils/request'
import type { PermItem, MenuNode } from '../../types/system/menus'


export async function apiMenuTree(): Promise<MenuNode[]> {
  return await http.get('/admin/menus/tree')
}
export async function apiPermList(): Promise<PermItem[]> {
  return await http.get('/admin/perms')
}
