import {
  authHttp,
  getAccessToken,
  http,
  setAccessToken,
  setRefreshToken,
  setTenantCode,
} from './http'
import type { RespBase } from '../types/api'
import type {
  CreateGuestTransferData,
  CreateGuestTransferReq,
  ExchangeGuestTransferData,
  ExchangeGuestTransferReq,
  GuestLoginData,
  GuestLoginReq,
  LoginReq,
  LoginResp,
  RefreshTokenReq,
  RefreshTokenResp,
  RegisterReq,
  RegisterResp,
  SendVerificationCodeReq,
} from '../types/auth'
import {
  collectGuestFingerprint,
  createGuestFingerprintHash,
  getGuestDeviceId,
  getGuestToken,
  setGuestDeviceId,
  setGuestId,
  setGuestToken,
} from '../utils/guestFingerprint'

export async function apiRegister(params: RegisterReq): Promise<RespBase & RegisterResp> {
  const guestFingerprint = await collectGuestFingerprint()
  const fingerprint = params.fingerprint || JSON.stringify(guestFingerprint)
  const deviceId =
    params.deviceId || getGuestDeviceId() || `web_${createGuestFingerprintHash(guestFingerprint)}`

  if (!params.deviceId && deviceId) setGuestDeviceId(deviceId)

  const payload: RegisterReq = {
    ...params,
    deviceId,
    fingerprint,
  }

  return http.post('/user/register', payload).then((res: { data: any }) => {
    const data = res.data
    if (data.data?.token?.accessToken) setAccessToken(data.data.token.accessToken)
    if (data.data?.token?.refreshToken) setRefreshToken(data.data.token.refreshToken)
    if (params.tenantCode) setTenantCode(params.tenantCode)
    return data
  })
}

export function apiLogin(params: LoginReq): Promise<RespBase & LoginResp> {
  return http.post('/user/login', params).then((res: { data: any }) => {
    const data = res.data
    if (data.data?.token?.accessToken) setAccessToken(data.data.token.accessToken)
    if (data.data?.token?.refreshToken) setRefreshToken(data.data.token.refreshToken)
    if (params.tenantCode) setTenantCode(params.tenantCode)
    return data
  })
}

export async function apiGuestLogin(
  params: Partial<GuestLoginReq> & Pick<GuestLoginReq, 'tenantCode'>,
): Promise<RespBase & { data: GuestLoginData }> {
  const guestFingerprint = await collectGuestFingerprint()
  const payload: GuestLoginReq = {
    ...params,
    deviceId: params.deviceId || getGuestDeviceId(),
    fingerprint: params.fingerprint || JSON.stringify(guestFingerprint),
    tenantCode: params.tenantCode,
  }

  return http.post('/user/guest-login', payload).then(async (res: { data: RespBase & { data: GuestLoginData } }) => {
    const data = res.data as RespBase & { data: GuestLoginData }
    if (data.data?.token) {
      setGuestToken(data.data.token)
      setAccessToken(data.data.token)
    }
    if (data.data?.deviceId) setGuestDeviceId(data.data.deviceId)
    if (data.data?.userId) setGuestId(data.data.userId)
    if (params.tenantCode) setTenantCode(params.tenantCode)
    await tryAutoRedirectGuestTransfer()
    return data
  })
}

export function apiCreateGuestTransfer(): Promise<RespBase & { data: CreateGuestTransferData }> {
  return authHttp.post('/user/guest-transfer/create', {}).then((res) => res.data)
}

export async function tryAutoRedirectGuestTransfer() {
  if (typeof window === 'undefined' || (!getGuestToken() && !getAccessToken())) return false

  const result = await apiCreateGuestTransfer()
  if (result.code !== 200 || !result.data?.redirectUrl) return false

  window.location.replace(result.data.redirectUrl)
  return true
}

export function apiExchangeGuestTransfer(
  params: ExchangeGuestTransferReq,
): Promise<RespBase & { data: ExchangeGuestTransferData }> {
  return http.post('/user/guest-transfer/exchange', params).then((res) => {
    const data = res.data as RespBase & { data: ExchangeGuestTransferData }
    if (data.data?.token) {
      setGuestToken(data.data.token)
      setAccessToken(data.data.token)
    }
    if (data.data?.deviceId) setGuestDeviceId(data.data.deviceId)
    if (data.data?.userId) setGuestId(data.data.userId)
    return data
  })
}

export function apiRefreshToken(params: RefreshTokenReq): Promise<RespBase & RefreshTokenResp> {
  return http.post('/user/refresh-token', params).then((res: { data: any }) => {
    const data = res.data
    if (data.data?.accessToken) setAccessToken(data.data.accessToken)
    if (data.data?.refreshToken) setRefreshToken(data.data.refreshToken)
    if (params.tenantCode) setTenantCode(params.tenantCode)
    return data
  })
}

export function apiSendVerificationCode(params: SendVerificationCodeReq): Promise<RespBase> {
  return http.post('/user/verification-code/send', params).then((res: { data: any }) => res.data)
}
