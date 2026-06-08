import {
  authHttp,
  clearAccessToken,
  clearRefreshToken,
  setTenantId,
} from './http'
import { buildPath, compactParams } from './utils'
import type { RespBase } from '../types/api'
import type {
  AddBankReq,
  AddBankResp,
  ChangeLoginPasswordReq,
  ChangePayPasswordReq,
  DeleteBankReq,
  GetIdentityResp,
  GetProfileResp,
  GetSecurityResp,
  InitGoogle2FAResp,
  ListBanksReq,
  SetDefaultBankReq,
  SetPayPasswordReq,
  SubmitIdentityReq,
  SubmitIdentityResp,
  UpdateIdentityReq,
  UpdateIdentityResp,
  UpdateBankReq,
  UpdateBankResp,
  UpdateProfileReq,
  UpdateProfileResp,
  VerifyGoogle2FAReq,
} from '../types/auth'
import type { UserBank } from '../types/user'

export function apiLogout(): Promise<RespBase> {
  return authHttp.post('/user/logout').then((res: { data: any }) => {
    clearAccessToken()
    clearRefreshToken()
    return res.data
  })
}

export function apiGetProfile(): Promise<RespBase & GetProfileResp> {
  return authHttp
    .get<RespBase & GetProfileResp>('/user/profile')
    .then((res: { data: any }) => {
      const data = res.data
      console.log('User profile:', data.data.user)
      setTenantId(data.data?.user.tenantId ?? '')
      return data
    })
}

export function apiUpdateProfile(
  params: UpdateProfileReq,
): Promise<RespBase & UpdateProfileResp> {
  return authHttp.put('/user/profile', params).then((res: { data: any }) => {
    const data = res.data
    return data
  })
}

export function apiChangeLoginPassword(
  params: ChangeLoginPasswordReq,
): Promise<RespBase> {
  return authHttp
    .put('/user/change-login-password', params)
    .then((res: { data: any }) => res.data)
}

export function apiGetIdentity(): Promise<RespBase & GetIdentityResp> {
  return authHttp.get('/user/identity').then((res: { data: any }) => res.data)
}

export function apiSubmitIdentity(
  params: SubmitIdentityReq,
): Promise<RespBase & SubmitIdentityResp> {
  return authHttp
    .post('/user/identity', params)
    .then((res: { data: any }) => res.data)
}

export function apiUpdateIdentity(
  params: UpdateIdentityReq,
): Promise<RespBase & UpdateIdentityResp> {
  return authHttp
    .put('/user/identity', params)
    .then((res: { data: any }) => res.data)
}

export function apiGetSecurity(): Promise<RespBase & GetSecurityResp> {
  return authHttp.get('/user/security').then((res: { data: any }) => res.data)
}

export function apiSetPayPassword(
  params: SetPayPasswordReq,
): Promise<RespBase> {
  return authHttp
    .post('/user/pay-password', params)
    .then((res: { data: any }) => res.data)
}

export function apiChangePayPassword(
  params: ChangePayPasswordReq,
): Promise<RespBase> {
  return authHttp
    .put('/user/pay-password', params)
    .then((res: { data: any }) => res.data)
}

export function apiInitGoogle2FA(): Promise<RespBase & InitGoogle2FAResp> {
  return authHttp
    .get('/user/google2fa/init')
    .then((res: { data: any }) => res.data)
}

export function apiEnableGoogle2FA(
  params: VerifyGoogle2FAReq,
): Promise<RespBase> {
  return authHttp
    .post('/user/google2fa/enable', params)
    .then((res: { data: any }) => res.data)
}

export function apiDisableGoogle2FA(
  params: VerifyGoogle2FAReq,
): Promise<RespBase> {
  return authHttp
    .post('/user/google2fa/disable', params)
    .then((res: { data: any }) => res.data)
}

export function apiListBanks(
  params: ListBanksReq,
): Promise<RespBase & { data: UserBank[] }> {
  return authHttp
    .get('/user/banks', { params: compactParams(params) })
    .then((res: { data: any }) => res.data)
}

export function apiAddBank(
  params: AddBankReq,
): Promise<RespBase & AddBankResp> {
  return authHttp
    .post('/user/banks', params)
    .then((res: { data: any }) => res.data)
}

export function apiUpdateBank(
  params: UpdateBankReq,
): Promise<RespBase & UpdateBankResp> {
  return authHttp
    .put(buildPath('/user/banks/:id', { id: params.id }), {
      bankName: params.bankName,
      bankCode: params.bankCode,
      accountName: params.accountName,
      accountNo: params.accountNo,
      branchName: params.branchName,
      countryCode: params.countryCode,
      isDefault: params.isDefault,
    })
    .then((res: { data: any }) => res.data)
}

export function apiDeleteBank(params: DeleteBankReq): Promise<RespBase> {
  return authHttp
    .delete(buildPath('/user/banks/:id', { id: params.id }))
    .then((res: { data: any }) => res.data)
}

export function apiSetDefaultBank(
  params: SetDefaultBankReq,
): Promise<RespBase> {
  return authHttp
    .put(buildPath('/user/banks/:id/default', { id: params.id }))
    .then((res: { data: any }) => res.data)
}
