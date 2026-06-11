import type { OptionsGroup } from './api'

export interface GetSystemCoreReq {}

export interface Interval {
  name: string
  kType: number
}

export interface SystemCore {
  assetUrl?: string
  isCaptchaEnabled: boolean
  isRegisterEnabled: boolean
  isGuestEnabled: boolean
  isCryptoEnabled: boolean
  intervals: Interval[]
  options?: OptionsGroup[]
}

export type GetSystemCoreResp = SystemCore
