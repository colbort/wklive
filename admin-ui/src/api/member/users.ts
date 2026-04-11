import { del, get, post, put } from '@/utils/request'
import type {
  AddUserBankReq,
  CreateMemberUserReq,
  ListMemberUserBanksReq,
  ListMemberUserIdentitiesReq,
  ListMemberUsersReq,
  MemberUserBank,
  MemberUserBankItem,
  MemberUserDetail,
  MemberUserIdentity,
  MemberUserIdentityItem,
  MemberUserItem,
  MemberUserSecurity,
  MemberRespBase,
  ReviewUserIdentityReq,
  SetDefaultUserBankReq,
  UpdateMemberUserBaseReq,
  UpdateMemberUserBankReq,
  UpdateMemberUserBankStatusReq,
  UpdateMemberUserLevelReq,
  UpdateMemberUserRiskLevelReq,
  UpdateMemberUserStatusReq,
} from '@/services/member/MemberUserService'

export function apiMemberUserList(params: ListMemberUsersReq): Promise<MemberRespBase<MemberUserItem[]>> {
  return get<MemberUserItem[]>('/admin/member/users', params)
}

export function apiMemberUserDetail(userId: number, tenantId: number): Promise<MemberRespBase<MemberUserDetail>> {
  return get<MemberUserDetail>(`/admin/member/users/${userId}`, { tenantId })
}

export function apiMemberUserCreate(data: CreateMemberUserReq): Promise<MemberRespBase<{ userId: number }>> {
  return post<{ userId: number }>('/admin/member/users', data)
}

export function apiMemberUserUpdateBase(userId: number, data: UpdateMemberUserBaseReq): Promise<MemberRespBase<MemberUserDetail>> {
  return put<MemberUserDetail>(`/admin/member/users/${userId}/base`, data)
}

export function apiMemberUserUpdateStatus(userId: number, data: UpdateMemberUserStatusReq): Promise<MemberRespBase> {
  return put(`/admin/member/users/${userId}/status`, data)
}

export function apiMemberUserUpdateLevel(userId: number, data: UpdateMemberUserLevelReq): Promise<MemberRespBase> {
  return put(`/admin/member/users/${userId}/level`, data)
}

export function apiMemberUserResetLoginPassword(userId: number, data: { tenantId: number; newPassword: string }): Promise<MemberRespBase> {
  return put(`/admin/member/users/${userId}/reset-login-password`, data)
}

export function apiMemberUserResetPayPassword(userId: number, data: { tenantId: number; newPayPassword: string }): Promise<MemberRespBase> {
  return put(`/admin/member/users/${userId}/reset-pay-password`, data)
}

export function apiMemberUserUnlock(userId: number, data: { tenantId: number }): Promise<MemberRespBase> {
  return put(`/admin/member/users/${userId}/unlock`, data)
}

export function apiMemberUserUpdateRiskLevel(userId: number, data: UpdateMemberUserRiskLevelReq): Promise<MemberRespBase> {
  return put(`/admin/member/users/${userId}/risk-level`, data)
}

export function apiMemberUserDelete(userId: number, tenantId: number): Promise<MemberRespBase> {
  return del(`/admin/member/users/${userId}`, { tenantId })
}

export function apiMemberUserSecurity(userId: number, tenantId: number): Promise<MemberRespBase<MemberUserSecurity>> {
  return get<MemberUserSecurity>(`/admin/member/users/${userId}/security`, { tenantId })
}

export function apiMemberUserReset2fa(userId: number, data: { tenantId: number }): Promise<MemberRespBase> {
  return put(`/admin/member/users/${userId}/reset2fa`, data)
}

export function apiMemberUserIdentityList(params: ListMemberUserIdentitiesReq): Promise<MemberRespBase<MemberUserIdentityItem[]>> {
  return get<MemberUserIdentityItem[]>('/admin/member/user-identities', params)
}

export function apiMemberUserIdentityReview(userId: number, data: ReviewUserIdentityReq): Promise<MemberRespBase<MemberUserIdentity>> {
  return put<MemberUserIdentity>(`/admin/member/user-identities/${userId}/review`, data)
}

export function apiMemberUserBankList(params: ListMemberUserBanksReq): Promise<MemberRespBase<MemberUserBankItem[]>> {
  return get<MemberUserBankItem[]>('/admin/member/user-banks', params)
}

export function apiMemberUserBankDetail(id: number, tenantId: number): Promise<MemberRespBase<MemberUserBank>> {
  return get<MemberUserBank>(`/admin/member/user-banks/${id}`, { tenantId })
}

export function apiMemberUserBankAdd(data: AddUserBankReq): Promise<MemberRespBase<MemberUserBank>> {
  return post<MemberUserBank>('/admin/member/user-banks', data)
}

export function apiMemberUserBankUpdate(id: number, data: UpdateMemberUserBankReq): Promise<MemberRespBase<MemberUserBank>> {
  return put<MemberUserBank>(`/admin/member/user-banks/${id}`, data)
}

export function apiMemberUserBankDelete(id: number, tenantId: number): Promise<MemberRespBase> {
  return del(`/admin/member/user-banks/${id}`, { tenantId })
}

export function apiMemberUserBankUpdateStatus(id: number, data: UpdateMemberUserBankStatusReq): Promise<MemberRespBase> {
  return put(`/admin/member/user-banks/${id}/status`, data)
}

export function apiMemberUserBankSetDefault(id: number, data: SetDefaultUserBankReq): Promise<MemberRespBase> {
  return put(`/admin/member/user-banks/${id}/default`, data)
}
