import type { TokenInfo, UserBank, UserIdentity, UserProfile, UserSecurity } from './user'

export const VERIFICATION_CODE_CHANNEL_EMAIL = 1
export const VERIFICATION_CODE_CHANNEL_PHONE = 2
export const VERIFICATION_CODE_SCENE_UNKNOWN = 0
export const VERIFICATION_CODE_SCENE_REGISTER = 1
export const VERIFICATION_CODE_SCENE_LOGIN = 2
export const VERIFICATION_CODE_SCENE_RESET_PASSWORD = 3
export const VERIFICATION_CODE_SCENE_BIND_EMAIL = 4
export const VERIFICATION_CODE_SCENE_BIND_PHONE = 5
export const VERIFICATION_CODE_SCENE_CHANGE_PASSWORD = 6
export const VERIFICATION_CODE_SCENE_WITHDRAW = 7
export const VERIFICATION_CODE_SCENE_TEST = 100

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
  deviceId?: string
  fingerprint?: string
}

export interface RegisterData {
  userId: number
  token: TokenInfo
  profile: UserProfile
}

export interface RegisterResp {
  data: RegisterData
}

export interface LoginReq {
  tenantCode: string
  loginType: number
  account: string
  password: string
  googleCode?: string
  loginIp?: string
}

export interface LoginData {
  userId: number
  token: TokenInfo
  profile: UserProfile
}

export interface LoginResp {
  data: LoginData
}

export interface GuestLoginReq {
  deviceId: string
  fingerprint: string
  registerIp?: string
  tenantCode: string
}

export interface GuestLoginData {
  token: string
  userId: string
  deviceId: string
  isNew: boolean
  username: string
}

export interface RefreshTokenReq {
  tenantCode?: string
  refreshToken: string
}

export interface RefreshTokenResp {
  data: TokenInfo
}

export interface SendVerificationCodeReq {
  channel: number
  email?: string
  phone?: string
  scene: number
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

export interface UpdateIdentityReq {
  phone?: string
  email?: string
  realName?: string
  gender?: number
  birthday?: number
  countryCode?: string
  province?: string
  city?: string
  address?: string
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
  data: {
    secret: string
    qrCodeUrl: string
  }
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
  isDefault?: number // 是否默认：1是 2否
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
  data: UserProfile
}

export interface SubmitIdentityResp {
  data: UserIdentity
}

export interface UpdateIdentityResp {
  data: UserIdentity
}

export interface AddBankResp {
  data: UserBank
}

export interface UpdateBankResp {
  data: UserBank
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
