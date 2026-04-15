import type { RespBase } from '@/services/BaseService'
import {
  apiMemberUserBankAdd,
  apiMemberUserBankDelete,
  apiMemberUserBankDetail,
  apiMemberUserBankList,
  apiMemberUserBankSetDefault,
  apiMemberUserBankUpdate,
  apiMemberUserBankUpdateStatus,
  apiMemberUserCheckReferrer,
  apiMemberUserCreate,
  apiMemberUserCreateOptions,
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
  detail?: T // 详情数据
  security?: T // 安全信息数据
  identity?: T // 实名信息数据
  bank?: T // 银行卡数据
}

export type UserBase = {
  id: number // 用户ID
  tenantId: number // 租户ID
  userNo: string // 用户编号
  username: string // 用户名
  nickname: string // 昵称
  avatar: string // 头像
  passwordHash: string // 登录密码哈希
  language: string // 语言
  timezone: string // 时区
  inviteCode: string // 邀请码
  signature: string // 个性签名
  registerType: number // 注册方式
  status: number // 用户状态
  memberLevel: number // 会员等级
  source: string // 注册来源
  referrerUserId: number // 邀请人ID
  lastLoginIp: string // 最后登录IP
  lastLoginTime: number // 最后登录时间
  registerIp: string // 注册IP
  registerTime: number // 注册时间
  isGuest: number // 是否游客
  isRecharge: number // 是否已充值
  deviceId: string // 设备ID
  fingerprint: string // 浏览器指纹
  remark: string // 备注
  deleted: number // 是否删除
  createTimes: number // 创建时间
  updateTimes: number // 更新时间
}

export type UserIdentity = {
  id: number // 主键ID
  tenantId: number // 租户ID
  userId: number // 用户ID
  phone: string // 手机号
  email: string // 邮箱
  realName: string // 真实姓名
  gender: number // 性别
  birthday: number // 生日
  countryCode: string // 国家/地区代码
  province: string // 省/州
  city: string // 城市
  address: string // 地址
  idType: number // 证件类型
  idNo: string // 证件号码
  frontImage: string // 证件正面
  backImage: string // 证件反面
  handheldImage: string // 手持证件照
  kycLevel: number // KYC等级
  verifyStatus: number // 实名状态
  rejectReason: string // 驳回原因
  submitTime: number // 提交时间
  verifyTime: number // 审核时间
  verifyBy: number // 审核人
  createTimes: number // 创建时间
  updateTimes: number // 更新时间
}

export type UserSecurity = {
  id: number // 主键ID
  tenantId: number // 租户ID
  userId: number // 用户ID
  payPasswordHash: string // 支付密码哈希
  googleSecret: string // Google 密钥
  googleEnabled: number // Google 2FA 是否启用
  loginErrorCount: number // 登录错误次数
  payErrorCount: number // 支付密码错误次数
  lockUntil: number // 锁定到期时间
  riskLevel: number // 风控等级
  createTimes: number // 创建时间
  updateTimes: number // 更新时间
}

export type UserBankItem = {
  id: number // 主键ID
  tenantId: number // 租户ID
  userId: number // 用户ID
  bankName: string // 银行名称
  bankCode: string // 银行编码
  accountName: string // 开户名
  accountNo: string // 银行卡号
  branchName: string // 支行名称
  countryCode: string // 国家/地区
  isDefault: number // 是否默认：0否 1是
  status: number // 状态：1正常 2禁用
  createTimes: number // 创建时间
  updateTimes: number // 更新时间
}

export type UserDetail = {
  base: UserBase // 用户基础信息
  identity: UserIdentity // 实名信息
  security: UserSecurity // 安全信息
  banks: UserBankItem[] // 银行卡列表
}

export type UserItem = {
  id: number // 用户ID
  tenantId: number // 租户ID
  userNo: string // 用户编号
  username: string // 用户名
  nickname: string // 昵称
  avatar: string // 头像
  passwordHash: string // 登录密码哈希
  registerType: number // 注册方式
  status: number // 用户状态
  memberLevel: number // 会员等级
  language: string // 语言
  timezone: string // 时区
  inviteCode: string // 邀请码
  signature: string // 个性签名
  source: string // 注册来源
  referrerUserId: number // 邀请人ID
  lastLoginIp: string // 最后登录IP
  lastLoginTime: number // 最后登录时间
  registerIp: string // 注册IP
  registerTime: number // 注册时间
  isGuest: number // 是否游客
  isRecharge: number // 是否已充值
  deviceId: string // 设备ID
  fingerprint: string // 浏览器指纹
  remark: string // 备注
  deleted: number // 是否删除
  createTimes: number // 创建时间
  updateTimes: number // 更新时间
}

export type UserIdentityItem = {
  userId: number // 用户ID
  userNo: string // 用户编号
  username: string // 用户名
  phone: string // 手机号
  email: string // 邮箱
  realName: string // 真实姓名
  idType: number // 证件类型
  idNo: string // 证件号码
  kycLevel: number // KYC等级
  verifyStatus: number // 实名状态
  rejectReason: string // 驳回原因
  submitTime: number // 提交时间
  verifyTime: number // 审核时间
  verifyBy: number // 审核人
}

export type MemberUserBase = UserBase
export type MemberUserIdentity = UserIdentity
export type MemberUserSecurity = UserSecurity
export type MemberUserBank = UserBankItem
export type MemberUserBankItem = UserBankItem
export type MemberUserDetail = UserDetail
export type MemberUserItem = UserItem

export type ListMemberUsersReq = {
  cursor?: number | string | null // 游标
  limit?: number // 条数限制
  tenantId?: number // 租户ID
  tenantCode?: string // 租户编码
  keyword?: string // 关键字
  userId?: number // 用户ID
  userNo?: string // 用户编号
  username?: string // 用户名
  phone?: string // 手机号
  email?: string // 邮箱
  status?: number // 用户状态
  memberLevel?: number // 会员等级
  verifyStatus?: number // 实名状态
  kycLevel?: number // KYC等级
  inviteCode?: string // 邀请码
  registerTimeStart?: number // 注册开始时间
  registerTimeEnd?: number // 注册结束时间
}

export type CreateMemberUserReq = {
  tenantCode: string // 租户编码
  username: string // 用户名
  nickname?: string // 昵称
  avatar?: string // 头像
  phone?: string // 手机号
  email?: string // 邮箱
  password: string // 登录密码
  registerType: number // 注册方式
  status: number // 用户状态
  memberLevel?: number // 会员等级
  language?: string // 语言
  timezone?: string // 时区
  inviteCode?: string // 邀请码
  signature?: string // 个性签名
  source?: string // 注册来源
  referrerUserId?: number // 邀请人ID
  referrerInviteCode?: string // 推荐人邀请码
  remark?: string // 备注
}

export type UpdateMemberUserBaseReq = {
  tenantId: number // 租户ID
  username?: string // 用户名
  nickname?: string // 昵称
  avatar?: string // 头像
  language?: string // 语言
  timezone?: string // 时区
  signature?: string // 个性签名
  source?: string // 注册来源
  referrerUserId?: number // 邀请人ID
  referrerInviteCode?: string // 推荐人邀请码
  remark?: string // 备注
  phone?: string // 手机号
  email?: string // 邮箱
}

export type UpdateMemberUserStatusReq = {
  tenantId: number // 租户ID
  status: number // 用户状态
  remark?: string // 备注
}

export type UpdateMemberUserLevelReq = {
  tenantId: number // 租户ID
  memberLevel: number // 会员等级
}

export type UpdateMemberUserRiskLevelReq = {
  tenantId: number // 租户ID
  riskLevel: number // 风控等级
}

export type UserOptionItem = {
  label: string
  value: number
  code?: string
}

export type CreateUserOptionsResp = MemberRespBase & {
  registerTypes: UserOptionItem[]
  statuses: UserOptionItem[]
}

export type CheckUserReferrerResp = MemberRespBase<{
  userId: number
  username: string
  nickname: string
  inviteCode: string
}> & {
  exists: boolean
}

export type ListMemberUserIdentitiesReq = {
  cursor?: number | string | null // 游标
  limit?: number // 条数限制
  tenantId?: number // 租户ID
  tenantCode?: string // 租户编码
  keyword?: string // 关键字
  userId?: number // 用户ID
  userNo?: string // 用户编号
  username?: string // 用户名
  phone?: string // 手机号
  email?: string // 邮箱
  realName?: string // 真实姓名
  verifyStatus?: number // 实名状态
  kycLevel?: number // KYC等级
  idType?: number // 证件类型
}

export type ReviewUserIdentityReq = {
  tenantId: number // 租户ID
  verifyStatus: number // 审核状态
  rejectReason?: string // 驳回原因
  verifyBy: number // 审核人
}

export type ListMemberUserBanksReq = {
  cursor?: number | string | null // 游标
  limit?: number // 条数限制
  tenantId?: number // 租户ID
  userId?: number // 用户ID
  keyword?: string // 关键字
  status?: number // 银行卡状态
}

export type AddUserBankReq = {
  tenantId: number // 租户ID
  userId: number // 用户ID
  bankName: string // 银行名称
  bankCode?: string // 银行编码
  accountName: string // 开户名
  accountNo: string // 银行卡号
  branchName?: string // 支行名称
  countryCode?: string // 国家/地区
  isDefault: number // 是否默认：0否 1是
  status: number // 状态：1正常 2禁用
}

export type UpdateMemberUserBankReq = AddUserBankReq

export type UpdateMemberUserBankStatusReq = {
  tenantId: number // 租户ID
  status: number // 银行卡状态
}

export type SetDefaultUserBankReq = {
  tenantId: number // 租户ID
  userId: number // 用户ID
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

  createOptions() {
    return apiMemberUserCreateOptions()
  }

  checkReferrer(inviteCode: string) {
    return apiMemberUserCheckReferrer(inviteCode)
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
