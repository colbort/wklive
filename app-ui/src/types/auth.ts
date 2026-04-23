import type { TokenInfo, UserBank, UserIdentity, UserProfile, UserSecurity } from '@/types/user'

export interface RegisterReq {
  tenantCode: string
  registerType: number
  username?: string
  phone?: string
  email?: string
  password: string
  confirmPassword: string
  inviteCode?: string
  source?: string
  registerIp?: string
}

export interface RegisterResp {
  userId: number
  token: TokenInfo
  profile: UserProfile
}

export interface LoginReq {
  tenantCode: string
  loginType: number
  account: string
  password: string
  googleCode?: string
  loginIp?: string
}

export interface LoginResp {
  userId: number
  token: TokenInfo
  profile: UserProfile
}

export interface GuestLoginReq {
  deviceId: string
  fingerprint: string
  registerIp?: string
  tenantCode: string
}

export interface GuestLoginData {
  token: string
  uid: string
  deviceId: string
  isNew: boolean
  username: string
}

export interface RefreshTokenReq {
  tenantCode?: string
  refreshToken: string
}

export interface RefreshTokenResp {
  token: TokenInfo
}

export interface WsTickTopic {
  topic: string
  categoryCode: string
  symbol: string
  market?: string
  interval?: string
}

export interface WsMessage {
  Type: string
  ClientTs?: number
  ServerTs?: number
  Topics?: WsTickTopic[]
  Code?: number
  Message?: string
  Topic?: string
  categoryCode?: string
  Symbol?: string
  Market?: string
  Interval?: string
  Payload?: number[]
}

export interface UpdateProfileReq {
  nickname?: string
  avatar?: string
  language?: string
  timezone?: string
  signature?: string
  gender?: number
  birthday?: number
  countryCode?: string
  province?: string
  city?: string
  address?: string
}

export interface ChangeLoginPasswordReq {
  oldPassword: string
  newPassword: string
  confirmPassword: string
}

export interface SubmitIdentityReq {
  phone?: string
  email?: string
  realName: string
  gender?: number
  birthday?: number
  countryCode?: string
  province?: string
  city?: string
  address?: string
  idType: number
  idNo: string
  frontImage?: string
  backImage?: string
  handheldImage?: string
  kycLevel?: number
}

export interface SetPayPasswordReq {
  password: string
  confirmPassword: string
}

export interface ChangePayPasswordReq {
  oldPassword: string
  newPassword: string
  confirmPassword: string
}

export interface InitGoogle2FAResp {
  secret: string
  qrCodeUrl: string
}

export interface VerifyGoogle2FAReq {
  googleCode: string
}

export interface ListBanksReq {
  cursor?: number
  limit?: number
}

export interface AddBankReq {
  bankName: string
  bankCode?: string
  accountName: string
  accountNo: string
  branchName?: string
  countryCode?: string
  isDefault?: number
}

export interface UpdateBankReq extends AddBankReq {
  id: number
}

export interface DeleteBankReq {
  id: number
}

export interface SetDefaultBankReq {
  id: number
}

export interface UpdateProfileResp {
  profile: UserProfile
}

export interface SubmitIdentityResp {
  data: UserIdentity
}

export interface AddBankResp {
  bank: UserBank
}

export interface UpdateBankResp {
  bank: UserBank
}

export interface GetProfileResp {
  data: UserProfile
}

export interface GetIdentityResp {
  data: UserIdentity
}

export interface GetSecurityResp {
  data: UserSecurity
}
