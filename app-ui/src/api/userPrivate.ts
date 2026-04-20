import { clearAccessToken, clearRefreshToken, setTenantId, http } from '@/api/http'
import { buildPath, compactParams } from '@/api/utils'
import type { RespBase } from '@/types/api'
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
  UpdateBankReq,
  UpdateBankResp,
  UpdateProfileReq,
  UpdateProfileResp,
  VerifyGoogle2FAReq,
} from '@/types/auth'
import type { UserBank } from '@/types/user'

export function apiLogout(): Promise<RespBase> {
  return http.post('/user/logout').then((res) => {
    clearAccessToken()
    clearRefreshToken()
    return res.data
  })
}

export function apiGetProfile(): Promise<RespBase & GetProfileResp> {
  return http.get('/user/profile').then((res) => {
    const data = res.data
    if (data.data?.base?.tenantId) setTenantId(data.data.base.tenantId)
    return data
  })
}

export function apiUpdateProfile(params: UpdateProfileReq): Promise<RespBase & UpdateProfileResp> {
  return http.put('/user/profile', params).then((res) => {
    const data = res.data
    if (data.profile?.base?.tenantId) setTenantId(data.profile.base.tenantId)
    return data
  })
}

export function apiChangeLoginPassword(params: ChangeLoginPasswordReq): Promise<RespBase> {
  return http.put('/user/change-login-password', params).then((res) => res.data)
}

export function apiGetIdentity(): Promise<RespBase & GetIdentityResp> {
  return http.get('/user/identity').then((res) => res.data)
}

export function apiSubmitIdentity(params: SubmitIdentityReq): Promise<RespBase & SubmitIdentityResp> {
  return http.post('/user/identity', params).then((res) => res.data)
}

export function apiGetSecurity(): Promise<RespBase & GetSecurityResp> {
  return http.get('/user/security').then((res) => res.data)
}

export function apiSetPayPassword(params: SetPayPasswordReq): Promise<RespBase> {
  return http.post('/user/pay-password', params).then((res) => res.data)
}

export function apiChangePayPassword(params: ChangePayPasswordReq): Promise<RespBase> {
  return http.put('/user/pay-password', params).then((res) => res.data)
}

export function apiInitGoogle2FA(): Promise<RespBase & InitGoogle2FAResp> {
  return http.get('/user/google2fa/init').then((res) => res.data)
}

export function apiEnableGoogle2FA(params: VerifyGoogle2FAReq): Promise<RespBase> {
  return http.post('/user/google2fa/enable', params).then((res) => res.data)
}

export function apiDisableGoogle2FA(params: VerifyGoogle2FAReq): Promise<RespBase> {
  return http.post('/user/google2fa/disable', params).then((res) => res.data)
}

export function apiListBanks(params: ListBanksReq): Promise<RespBase & { data: UserBank[] }> {
  return http.get('/user/banks', { params: compactParams(params) }).then((res) => res.data)
}

export function apiAddBank(params: AddBankReq): Promise<RespBase & AddBankResp> {
  return http.post('/user/banks', params).then((res) => res.data)
}

export function apiUpdateBank(params: UpdateBankReq): Promise<RespBase & UpdateBankResp> {
  return http
    .put(buildPath('/user/banks/:id', { id: params.id }), {
      bankName: params.bankName,
      bankCode: params.bankCode,
      accountName: params.accountName,
      accountNo: params.accountNo,
      branchName: params.branchName,
      countryCode: params.countryCode,
      isDefault: params.isDefault,
    })
    .then((res) => res.data)
}

export function apiDeleteBank(params: DeleteBankReq): Promise<RespBase> {
  return http.delete(buildPath('/user/banks/:id', { id: params.id })).then((res) => res.data)
}

export function apiSetDefaultBank(params: SetDefaultBankReq): Promise<RespBase> {
  return http.put(buildPath('/user/banks/:id/default', { id: params.id })).then((res) => res.data)
}
