import { get } from '@/utils/request'
import type { PermItem, MenuNode } from '../../types/system/menus'
import { ApiResp } from '../types'


export async function apiMenuTree(): Promise<ApiResp<MenuNode[]>> {
  return await get('/admin/menus/tree')
}
export async function apiPermList(): Promise<ApiResp<PermItem[]>> {
  return await get('/admin/perms')
}
