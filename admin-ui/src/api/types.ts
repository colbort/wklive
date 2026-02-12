export interface ApiResp<T = any> {
  code: number
  msg: string
  data?: T
  total?: number
}
