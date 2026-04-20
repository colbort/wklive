export interface TokenInfo {
  accessToken: string
  refreshToken: string
  expireTime: number
}

export interface TenantInfo {
  tenantId: number
  tenantCode: string
  tenantName: string
  tenantStatus: number
  expireTime: number
}

export interface UserBase {
  id: number
  tenantId: number
  userNo: string
  username: string
  nickname: string
  avatar: string
  language: string
  timezone: string
  inviteCode: string
  signature: string
  registerType: number
  status: number
  memberLevel: number
  source: string
  referrerUserId: number
  lastLoginIp: string
  lastLoginTime: number
  registerIp: string
  registerTime: number
  remark: string
  deleted: boolean
  createTimes: number
  updateTimes: number
}

export interface UserIdentity {
  id: number
  tenantId: number
  userId: number
  phone: string
  email: string
  realName: string
  gender: number
  birthday: number
  countryCode: string
  province: string
  city: string
  address: string
  idType: number
  idNo: string
  frontImage: string
  backImage: string
  handheldImage: string
  kycLevel: number
  verifyStatus: number
  rejectReason: string
  submitTime: number
  verifyTime: number
  verifyBy: number
  createTimes: number
  updateTimes: number
}

export interface UserSecurity {
  id: number
  tenantId: number
  userId: number
  hasPayPassword: boolean
  googleEnabled: number
  loginErrorCount: number
  payErrorCount: number
  lockUntil: number
  riskLevel: number
  createTimes: number
  updateTimes: number
}

export interface UserBank {
  id: number
  tenantId: number
  userId: number
  bankName: string
  bankCode: string
  accountName: string
  accountNo: string
  branchName: string
  countryCode: string
  isDefault: number
  status: number
  createTimes: number
  updateTimes: number
}

export interface UserProfile {
  base: UserBase
  identity?: UserIdentity
  security?: UserSecurity
}

export interface UserDetail {
  base: UserBase
  identity?: UserIdentity
  security?: UserSecurity
  banks?: UserBank[]
}
