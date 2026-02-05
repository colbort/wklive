export interface ApiResp<T = any> {
  code: number
  msg: string
  data?: T
  // 你现在列表接口是把 list/total 放在外层：
  list?: T
  total?: number
}
