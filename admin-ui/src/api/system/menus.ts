import { del, get, post, put } from '@/utils/request'
import type { PermItem, MenuNode, SysMenuCreateReq, SysMenuUpdateReq, SysMenuListReq, SysMenuListResp } from '../../types/system/menus'
import { ApiResp } from '../types'
import { RespBase } from '@/types/common'


export async function apiMenuTree(): Promise<ApiResp<MenuNode[]>> {
  return await get('/admin/menus/tree')
}
export async function apiPermList(): Promise<ApiResp<PermItem[]>> {
  return await get('/admin/perms')
}

/** 新增菜单 */
export async function sysMenuCreate(data: SysMenuCreateReq) : Promise<ApiResp<RespBase>> {
  return await post('/admin/menus', data)
}

/** 更新菜单 */
export async function sysMenuUpdate(data: SysMenuUpdateReq) : Promise<ApiResp<RespBase>> {
  return await put('/admin/menus', data)
}

/** 删除菜单 */
export async function sysMenuDelete(id: number) : Promise<ApiResp<RespBase>> {
  return await del(`/admin/menus/${id}`)
}

/** 菜单列表 */
export async function sysMenuList(data: SysMenuListReq) : Promise<ApiResp<SysMenuListResp>> {
  return await get('/admin/menus', data)
}
