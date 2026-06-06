import { http, setAccessToken, setRefreshToken, setTenantCode } from './http'
import type { OptionsGroup, RespBase } from '../types/api'
import type {
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
  setGuestDeviceId,
  setGuestId,
  setGuestToken,
} from '../utils/guestFingerprint'

export function apiGetUserOptions(): Promise<RespBase & { data: OptionsGroup[] }> {
  return http.get('/user/options').then((res: { data: any }) => res.data)
}

export function apiRegister(params: RegisterReq): Promise<RespBase & RegisterResp> {
  const guestFingerprint = collectGuestFingerprint()
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

export function apiGuestLogin(
  params: Partial<GuestLoginReq> & Pick<GuestLoginReq, 'tenantCode'>,
): Promise<RespBase & { data: GuestLoginData }> {
  const guestFingerprint = collectGuestFingerprint()
  const payload: GuestLoginReq = {
    ...params,
    deviceId: params.deviceId || getGuestDeviceId(),
    fingerprint: params.fingerprint || JSON.stringify(guestFingerprint),
    tenantCode: params.tenantCode,
  }

  return http.post('/user/guest-login', payload).then((res: { data: RespBase & { data: GuestLoginData } }) => {
    const data = res.data as RespBase & { data: GuestLoginData }
    if (data.data?.token) {
      setGuestToken(data.data.token)
      setAccessToken(data.data.token)
    }
    if (data.data?.deviceId) setGuestDeviceId(data.data.deviceId)
    if (data.data?.userId) setGuestId(data.data.userId)
    if (params.tenantCode) setTenantCode(params.tenantCode)
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
