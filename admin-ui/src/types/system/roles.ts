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

export type RespBase = { code: number; msg: string }

export type RoleItem = { id: number; name: string; code: string; status: number; remark?: string }