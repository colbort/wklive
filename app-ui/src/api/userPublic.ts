import { http, setAccessToken, setRefreshToken, setTenantCode, setTenantId } from '@/api/http'
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

export function apiRegister(params: RegisterReq): Promise<RespBase & RegisterResp> {
  return http.post('/user/register', params).then((res) => {
    const data = res.data
    if (data.token?.accessToken) setAccessToken(data.token.accessToken)
    if (data.token?.refreshToken) setRefreshToken(data.token.refreshToken)
    if (params.tenantCode) setTenantCode(params.tenantCode)
    if (data.profile?.base?.tenantId) setTenantId(data.profile.base.tenantId)
    return data
  })
}

export function apiLogin(params: LoginReq): Promise<RespBase & LoginResp> {
  return http.post('/user/login', params).then((res) => {
    const data = res.data
    if (data.token?.accessToken) setAccessToken(data.token.accessToken)
    if (data.token?.refreshToken) setRefreshToken(data.token.refreshToken)
    if (params.tenantCode) setTenantCode(params.tenantCode)
    if (data.profile?.base?.tenantId) setTenantId(data.profile.base.tenantId)
    return data
  })
}

export function apiGuestLogin(params: GuestLoginReq): Promise<RespBase & { data: GuestLoginData }> {
  return http.post('/user/guest-login', params).then((res) => res.data)
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
