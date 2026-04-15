import { del, get, post, put } from '@/utils/request'

import {
  OptionGroup,
  RespBase,
  AddUserBankReq,
  CheckUserReferrerResp,
  CreateMemberUserReq,
  ListMemberUserBanksReq,
  ListMemberUserIdentitiesReq,
  ListMemberUsersReq,
  UserItem,
  UserBankItem,
  UserDetail,
  UserIdentity,
  UserIdentityItem,
  UserSecurity,
  MemberRespBase,
  ReviewUserIdentityReq,
  SetDefaultUserBankReq,
  UpdateMemberUserBaseReq,
  UpdateMemberUserBankReq,
  UpdateMemberUserBankStatusReq,
  UpdateMemberUserLevelReq,
  UpdateMemberUserRiskLevelReq,
  UpdateMemberUserStatusReq,
} from '@/services'

export function apiMemberUserList(params: ListMemberUsersReq): Promise<MemberRespBase<UserItem[]>> {
  return get<UserItem[]>('/admin/member/users', params)
}

export function apiMemberUserDetail(
  userId: number,
  tenantId: number,
): Promise<MemberRespBase<UserDetail>> {
  return get<UserDetail>(`/admin/member/users/${userId}`, { tenantId })
}

export function apiMemberUserCreate(
  data: CreateMemberUserReq,
): Promise<MemberRespBase<{ userId: number }>> {
  return post<{ userId: number }>('/admin/member/users', data)
}

export function apiMemberUserCheckReferrer(inviteCode: string): Promise<CheckUserReferrerResp> {
  return get<CheckUserReferrerResp>('/admin/member/users/referrer/check', {
    inviteCode,
  }) as Promise<CheckUserReferrerResp>
}

export function apiMemberUserUpdateBase(
  userId: number,
  data: UpdateMemberUserBaseReq,
): Promise<MemberRespBase<UserDetail>> {
  return put<UserDetail>(`/admin/member/users/${userId}/base`, data)
}

export function apiMemberUserUpdateStatus(
  userId: number,
  data: UpdateMemberUserStatusReq,
): Promise<MemberRespBase> {
  return put(`/admin/member/users/${userId}/status`, data)
}

export function apiMemberUserUpdateLevel(
  userId: number,
  data: UpdateMemberUserLevelReq,
): Promise<MemberRespBase> {
  return put(`/admin/member/users/${userId}/level`, data)
}

export function apiMemberUserResetLoginPassword(
  userId: number,
  data: { tenantId: number; newPassword: string },
): Promise<MemberRespBase> {
  return put(`/admin/member/users/${userId}/reset-login-password`, data)
}

export function apiMemberUserResetPayPassword(
  userId: number,
  data: { tenantId: number; newPayPassword: string },
): Promise<MemberRespBase> {
  return put(`/admin/member/users/${userId}/reset-pay-password`, data)
}

export function apiMemberUserUnlock(
  userId: number,
  data: { tenantId: number },
): Promise<MemberRespBase> {
  return put(`/admin/member/users/${userId}/unlock`, data)
}

export function apiMemberUserUpdateRiskLevel(
  userId: number,
  data: UpdateMemberUserRiskLevelReq,
): Promise<MemberRespBase> {
  return put(`/admin/member/users/${userId}/risk-level`, data)
}

export function apiMemberUserDelete(userId: number, tenantId: number): Promise<MemberRespBase> {
  return del(`/admin/member/users/${userId}`, { tenantId })
}

export function apiMemberUserSecurity(
  userId: number,
  tenantId: number,
): Promise<MemberRespBase<UserSecurity>> {
  return get<UserSecurity>(`/admin/member/users/${userId}/security`, { tenantId })
}

export function apiMemberUserReset2fa(
  userId: number,
  data: { tenantId: number },
): Promise<MemberRespBase> {
  return put(`/admin/member/users/${userId}/reset2fa`, data)
}

export function apiMemberUserIdentityList(
  params: ListMemberUserIdentitiesReq,
): Promise<MemberRespBase<UserIdentityItem[]>> {
  return get<UserIdentityItem[]>('/admin/member/user-identities', params)
}

export function apiMemberUserIdentityReview(
  userId: number,
  data: ReviewUserIdentityReq,
): Promise<MemberRespBase<UserIdentity>> {
  return put<UserIdentity>(`/admin/member/user-identities/${userId}/review`, data)
}

export function apiMemberUserBankList(
  params: ListMemberUserBanksReq,
): Promise<MemberRespBase<UserBankItem[]>> {
  return get<UserBankItem[]>('/admin/member/user-banks', params)
}

export function apiMemberUserBankDetail(
  id: number,
  tenantId: number,
): Promise<MemberRespBase<UserBankItem>> {
  return get<UserBankItem>(`/admin/member/user-banks/${id}`, { tenantId })
}

export function apiMemberUserBankAdd(data: AddUserBankReq): Promise<MemberRespBase<UserBankItem>> {
  return post<UserBankItem>('/admin/member/user-banks', data)
}

export function apiMemberUserBankUpdate(
  id: number,
  data: UpdateMemberUserBankReq,
): Promise<MemberRespBase<UserBankItem>> {
  return put<UserBankItem>(`/admin/member/user-banks/${id}`, data)
}

export function apiMemberUserBankDelete(id: number, tenantId: number): Promise<MemberRespBase> {
  return del(`/admin/member/user-banks/${id}`, { tenantId })
}

export function apiMemberUserBankUpdateStatus(
  id: number,
  data: UpdateMemberUserBankStatusReq,
): Promise<MemberRespBase> {
  return put(`/admin/member/user-banks/${id}/status`, data)
}

export function apiMemberUserBankSetDefault(
  id: number,
  data: SetDefaultUserBankReq,
): Promise<MemberRespBase> {
  return put(`/admin/member/user-banks/${id}/default`, data)
}

export function apiOptions(): Promise<RespBase<OptionGroup[]>> {
  return get('/admin/member/options')
}
