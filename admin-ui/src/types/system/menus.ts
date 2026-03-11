
import {RespBase } from '../common'

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
  page: number
  size: number
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

export type SysMenuListResp = {
  base: RespBase
  data: SysMenuItem[]
}