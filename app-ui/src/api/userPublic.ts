import { http, setAccessToken, setRefreshToken, setTenantCode } from '@/api/http'
import type { RespBase } from '@/types/api'
import type {
  GuestLoginData,
  GuestLoginReq,
  LoginReq,
  LoginResp,
  RefreshTokenReq,
  RefreshTokenResp,
  RegisterReq,
  RegisterResp,
} from '@/types/auth'
import {
  collectGuestFingerprint,
  getGuestDeviceId,
  setGuestDeviceId,
  setGuestId,
  setGuestToken,
} from '@/utils/guestFingerprint'

export function apiRegister(params: RegisterReq): Promise<RespBase & RegisterResp> {
  return http.post('/user/register', params).then((res) => {
    const data = res.data
    if (data.token?.accessToken) setAccessToken(data.token.accessToken)
    if (data.token?.refreshToken) setRefreshToken(data.token.refreshToken)
    if (params.tenantCode) setTenantCode(params.tenantCode)
    return data
  })
}

export function apiLogin(params: LoginReq): Promise<RespBase & LoginResp> {
  return http.post('/user/login', params).then((res) => {
    const data = res.data
    if (data.token?.accessToken) setAccessToken(data.token.accessToken)
    if (data.token?.refreshToken) setRefreshToken(data.token.refreshToken)
    if (params.tenantCode) setTenantCode(params.tenantCode)
    return data
  })
}

export function apiGuestLogin(params: Partial<GuestLoginReq> & Pick<GuestLoginReq, 'tenantCode'>): Promise<RespBase & { data: GuestLoginData }> {
  const guestFingerprint = collectGuestFingerprint()
  const payload: GuestLoginReq = {
    ...params,
    deviceId: params.deviceId || getGuestDeviceId(),
    fingerprint: params.fingerprint || JSON.stringify(guestFingerprint),
    tenantCode: params.tenantCode,
  }

  return http.post('/user/guest-login', payload).then((res) => {
    const data = res.data as RespBase & { data: GuestLoginData }
    if (data.data?.token) {
      setGuestToken(data.data.token)
      setAccessToken(data.data.token)
    }
    if (data.data?.deviceId) setGuestDeviceId(data.data.deviceId)
    if (data.data?.uid) setGuestId(data.data.uid)
    if (params.tenantCode) setTenantCode(params.tenantCode)
    return data
  })
}

export function apiRefreshToken(params: RefreshTokenReq): Promise<RespBase & RefreshTokenResp> {
  return http.post('/user/refresh-token', params).then((res) => {
    const data = res.data
    if (data.token?.accessToken) setAccessToken(data.token.accessToken)
    if (data.token?.refreshToken) setRefreshToken(data.token.refreshToken)
    if (params.tenantCode) setTenantCode(params.tenantCode)
    return data
  })
}
