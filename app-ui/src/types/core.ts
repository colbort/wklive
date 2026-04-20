export interface GetSystemCoreReq {}

export interface Interval {
  name: string
  kType: number
}

export interface SystemCore {
  isCaptchaEnabled: boolean
  isRegisterEnabled: boolean
  isGuestEnabled: boolean
  isCryptoEnabled: boolean
  intervals: Interval[]
}

export type GetSystemCoreResp = SystemCore
