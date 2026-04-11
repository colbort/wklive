import type { RespBase } from '@/services/BaseService'
import {
  apiMemberUserBankAdd,
  apiMemberUserBankDelete,
  apiMemberUserBankDetail,
  apiMemberUserBankList,
  apiMemberUserBankSetDefault,
  apiMemberUserBankUpdate,
  apiMemberUserBankUpdateStatus,
  apiMemberUserCreate,
  apiMemberUserDelete,
  apiMemberUserDetail,
  apiMemberUserIdentityList,
  apiMemberUserIdentityReview,
  apiMemberUserList,
  apiMemberUserReset2fa,
  apiMemberUserResetLoginPassword,
  apiMemberUserResetPayPassword,
  apiMemberUserSecurity,
  apiMemberUserUnlock,
  apiMemberUserUpdateBase,
  apiMemberUserUpdateLevel,
  apiMemberUserUpdateRiskLevel,
  apiMemberUserUpdateStatus,
} from '@/api/member/users'

export type MemberRespBase<T = any> = RespBase<T> & {
  detail?: T
  security?: T
  identity?: T
  bank?: T
}

export type MemberUserBase = {
  id: number
  tenantId: number
  userNo: string
  username: string
  nickname: string
  avatar: string
  language: string
  timezone: string
  inviteCode: string
  signature: string
  registerType: number
  status: number
  memberLevel: number
  source: string
  referrerUserId: number
  lastLoginIp: string
  lastLoginTime: number
  registerIp: string
  registerTime: number
  remark: string
  deleted: boolean
  createTimes: number
  updateTimes: number
}

export type MemberUserIdentity = {
  id: number
  tenantId: number
  userId: number
  phone: string
  email: string
  realName: string
  gender: number
  birthday: string
  countryCode: string
  province: string
  city: string
  address: string
  idType: number
  idNo: string
  frontImage: string
  backImage: string
  handheldImage: string
  kycLevel: number
  verifyStatus: number
  rejectReason: string
  submitTime: number
  verifyTime: number
  verifyBy: number
  createTimes: number
  updateTimes: number
}

export type MemberUserSecurity = {
  id: number
  tenantId: number
  userId: number
  hasPayPassword: boolean
  googleEnabled: boolean
  loginErrorCount: number
  payErrorCount: number
  lockUntil: number
  riskLevel: number
  createTimes: number
  updateTimes: number
}

export type MemberUserBank = {
  id: number
  tenantId: number
  userId: number
  bankName: string
  bankCode: string
  accountName: string
  accountNo: string
  maskedAccountNo: string
  branchName: string
  countryCode: string
  isDefault: boolean
  status: number
  createTimes: number
  updateTimes: number
}

export type MemberUserDetail = {
  base: MemberUserBase
  identity: MemberUserIdentity
  security: MemberUserSecurity
  banks: MemberUserBank[]
}

export type MemberUserItem = {
  userId: number
  userNo: string
  username: string
  nickname: string
  avatar: string
  phone: string
  email: string
  realName: string
  status: number
  memberLevel: number
  kycLevel: number
  verifyStatus: number
  inviteCode: string
  lastLoginIp: string
  lastLoginTime: number
  registerTime: number
}

export type MemberUserIdentityItem = {
  userId: number
  userNo: string
  username: string
  phone: string
  email: string
  realName: string
  idType: number
  idNo: string
  kycLevel: number
  verifyStatus: number
  rejectReason: string
  submitTime: number
  verifyTime: number
  verifyBy: number
}

export type MemberUserBankItem = {
  id: number
  userId: number
  userNo: string
  username: string
  bankName: string
  bankCode: string
  accountName: string
  accountNo: string
  maskedAccountNo: string
  branchName: string
  countryCode: string
  isDefault: boolean
  status: number
  createTimes: number
}

export type ListMemberUsersReq = {
  cursor?: number | string | null
  limit?: number
  tenantId?: number
  tenantCode?: string
  keyword?: string
  userId?: number
  userNo?: string
  username?: string
  phone?: string
  email?: string
  status?: number
  memberLevel?: number
  verifyStatus?: number
  kycLevel?: number
  inviteCode?: string
  registerTimeStart?: number
  registerTimeEnd?: number
}

export type CreateMemberUserReq = {
  tenantId: number
  username: string
  nickname?: string
  avatar?: string
  phone?: string
  email?: string
  password: string
  registerType: number
  status: number
  memberLevel?: number
  language?: string
  timezone?: string
  inviteCode?: string
  signature?: string
  source?: string
  referrerUserId?: number
  remark?: string
}

export type UpdateMemberUserBaseReq = {
  tenantId: number
  username?: string
  nickname?: string
  avatar?: string
  language?: string
  timezone?: string
  signature?: string
  source?: string
  referrerUserId?: number
  remark?: string
  phone?: string
  email?: string
}

export type UpdateMemberUserStatusReq = {
  tenantId: number
  status: number
  remark?: string
}

export type UpdateMemberUserLevelReq = {
  tenantId: number
  memberLevel: number
}

export type UpdateMemberUserRiskLevelReq = {
  tenantId: number
  riskLevel: number
}

export type ListMemberUserIdentitiesReq = {
  cursor?: number | string | null
  limit?: number
  tenantId?: number
  tenantCode?: string
  keyword?: string
  userId?: number
  userNo?: string
  username?: string
  phone?: string
  email?: string
  realName?: string
  verifyStatus?: number
  kycLevel?: number
  idType?: number
}

export type ReviewUserIdentityReq = {
  tenantId: number
  verifyStatus: number
  rejectReason?: string
  verifyBy: number
}

export type ListMemberUserBanksReq = {
  cursor?: number | string | null
  limit?: number
  tenantId?: number
  userId?: number
  keyword?: string
  status?: number
}

export type AddUserBankReq = {
  tenantId: number
  userId: number
  bankName: string
  bankCode?: string
  accountName: string
  accountNo: string
  branchName?: string
  countryCode?: string
  isDefault: boolean
  status: number
}

export type UpdateMemberUserBankReq = AddUserBankReq

export type UpdateMemberUserBankStatusReq = {
  tenantId: number
  status: number
}

export type SetDefaultUserBankReq = {
  tenantId: number
  userId: number
}

export class MemberUserService {
  getList(params: ListMemberUsersReq) {
    return apiMemberUserList(params)
  }

  getDetail(userId: number, tenantId: number) {
    return apiMemberUserDetail(userId, tenantId)
  }

  create(data: CreateMemberUserReq) {
    return apiMemberUserCreate(data)
  }

  updateBase(userId: number, data: UpdateMemberUserBaseReq) {
    return apiMemberUserUpdateBase(userId, data)
  }

  updateStatus(userId: number, data: UpdateMemberUserStatusReq) {
    return apiMemberUserUpdateStatus(userId, data)
  }

  updateLevel(userId: number, data: UpdateMemberUserLevelReq) {
    return apiMemberUserUpdateLevel(userId, data)
  }

  resetLoginPassword(userId: number, tenantId: number, newPassword: string) {
    return apiMemberUserResetLoginPassword(userId, { tenantId, newPassword })
  }

  resetPayPassword(userId: number, tenantId: number, newPayPassword: string) {
    return apiMemberUserResetPayPassword(userId, { tenantId, newPayPassword })
  }

  unlock(userId: number, tenantId: number) {
    return apiMemberUserUnlock(userId, { tenantId })
  }

  updateRiskLevel(userId: number, data: UpdateMemberUserRiskLevelReq) {
    return apiMemberUserUpdateRiskLevel(userId, data)
  }

  delete(userId: number, tenantId: number) {
    return apiMemberUserDelete(userId, tenantId)
  }

  getSecurity(userId: number, tenantId: number) {
    return apiMemberUserSecurity(userId, tenantId)
  }

  reset2fa(userId: number, tenantId: number) {
    return apiMemberUserReset2fa(userId, { tenantId })
  }

  listIdentities(params: ListMemberUserIdentitiesReq) {
    return apiMemberUserIdentityList(params)
  }

  reviewIdentity(userId: number, data: ReviewUserIdentityReq) {
    return apiMemberUserIdentityReview(userId, data)
  }

  listBanks(params: ListMemberUserBanksReq) {
    return apiMemberUserBankList(params)
  }

  getBank(id: number, tenantId: number) {
    return apiMemberUserBankDetail(id, tenantId)
  }

  addBank(data: AddUserBankReq) {
    return apiMemberUserBankAdd(data)
  }

  updateBank(id: number, data: UpdateMemberUserBankReq) {
    return apiMemberUserBankUpdate(id, data)
  }

  deleteBank(id: number, tenantId: number) {
    return apiMemberUserBankDelete(id, tenantId)
  }

  updateBankStatus(id: number, data: UpdateMemberUserBankStatusReq) {
    return apiMemberUserBankUpdateStatus(id, data)
  }

  setDefaultBank(id: number, data: SetDefaultUserBankReq) {
    return apiMemberUserBankSetDefault(id, data)
  }
}

export const memberUserService = new MemberUserService()
