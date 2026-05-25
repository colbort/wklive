import type { RespBase, BaseService } from '@/services'
import {
  apiMenuTree,
  apiPermList,
  sysMenuList,
  sysMenuCreate,
  sysMenuDelete,
  sysMenuUpdate,
} from '@/api/system/menus'
import { apiOptions } from '@/api/system/options'
import type { OptionGroup } from '@/services'

// ===== 菜单相关类型定义 =====

export type MenuNode = { id: number; name: string; children?: MenuNode[] }
export type PermItem = { key: string; name: string; group?: string }

export type SysMenuCreateReq = {
  parentId: number
  name: string
  menuType: number
  path: string
  component: string
  icon: string
  sort: number
  visible: number
  status: number
  perms: string
}

export type SysMenuDeleteReq = {
  id: number
}

export type SysMenuUpdateReq = {
  id: number
  parentId: number
  name: string
  menuType: number
  path: string
  component: string
  icon: string
  sort: number
  visible: number
  status: number
  perms: string
}

export type SysMenuListReq = {
  cursor?: number
  limit?: number
  keyword: string
  menuType: number
  status: number
  visible: number
}

export type SysMenuItem = {
  id: number
  parentId: number
  name: string
  menuType: number // 1目录 2菜单 3按钮
  path: string
  component: string
  icon: string
  sort: number
  visible: number
  status: number
  perms: string // 按钮必须有，例如 sys:user:add
}

export type SysMenuTreeItem = SysMenuItem & {
  children?: SysMenuTreeItem[]
}

/**
 * 菜单服务类
 * 实现 BaseService 接口，使用现有的 API 函数
 */
export class MenuService implements BaseService {
  /**
   * 获取菜单树
   */
  async getMenuTree(tenantId: number): Promise<RespBase<MenuNode[]>> {
    return apiMenuTree(tenantId)
  }

  /**
   * 获取权限列表
   */
  async getPermissionList(): Promise<RespBase<PermItem[]>> {
    return apiPermList()
  }

  /**
   * 获取菜单列表
   */
  async getList(params?: SysMenuListReq): Promise<RespBase<SysMenuItem[]>> {
    return sysMenuList(params || ({} as SysMenuListReq))
  }

  async getOptions(): Promise<RespBase<OptionGroup[]>> {
    return apiOptions()
  }

  /**
   * 创建菜单
   */
  async create(data: SysMenuCreateReq): Promise<RespBase> {
    return sysMenuCreate(data)
  }

  /**
   * 更新菜单
   */
  async update(id: string | number, data: SysMenuUpdateReq): Promise<RespBase> {
    return sysMenuUpdate({ ...data, id: Number(id) })
  }

  /**
   * 删除菜单
   */
  async delete(id: string | number): Promise<RespBase> {
    return sysMenuDelete(Number(id))
  }
}

// 导出单例实例
export const menuService = new MenuService()
