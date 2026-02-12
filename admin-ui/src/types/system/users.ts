export type PageReq = { page?: number; size?: number } // 你的后端如果是 page/size 或 pageNo/pageSize，自行对齐


export type SysUserItem = {
  id               :number
  username         :string
  nickname         :string
  status           :number
  roleIds          :number[]
  createdAt        :number
  google2faEnabled :number
}

export type Google2FABindInitResp = {
  secret     :string
  otpauthUrl :string
  qrCode     :string
}