// ===== 日志相关类型定义 =====

// 登录日志项
export type LoginLogItem = {
  id: number
  userId: number
  username: string
  ip: string
  ua: string
  success: number
  msg?: string
  loginAt: number
}

// 登录日志列表请求
export interface LoginLogListReq {
  page?: number
  size?: number
  username?: string
  success?: number
}

// 登录日志列表响应
export interface LoginLogListResp {
  code: number
  msg: string
  data: LoginLogItem[]
  total?: number
}

// 操作日志项
export type OpLogItem = {
  id: number
  userId: number
  username: string
  method: string
  path: string
  req?: string
  resp?: string
  ip: string
  costMs: number
  createdAt: number
  updatedAt: number
}

// 操作日志列表请求
export interface OpLogListReq {
  page?: number
  size?: number
  username?: string
  method?: string
  path?: string
}

// 操作日志列表响应
export interface OpLogListResp {
  code: number
  msg: string
  data: OpLogItem[]
  total?: number
}