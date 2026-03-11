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