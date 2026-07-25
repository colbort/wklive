import type { OptionGroup, RespBase } from '@/services'
import {
  apiTenantPayAccountCreate,
  apiTenantPayAccountDetail,
  apiTenantPayAccountList,
  apiTenantPayAccountUpdate,
  apiTenantPayChannelCreate,
  apiTenantPayChannelDetail,
  apiTenantPayChannelList,
  apiTenantPayChannelRuleCreate,
  apiTenantPayChannelRuleDetail,
  apiTenantPayChannelRuleList,
  apiTenantPayChannelRuleUpdate,
  apiTenantPayChannelUpdate,
} from '@/api/payment/tenant'
import { getCoreOptions } from '@/stores/core'

export type TenantPayAccount = {
  id: number // 账号ID
  tenantId: number // 租户ID
  platformId: number // 平台ID
  accountCode: string // 账号编码
  accountName: string // 账号名称
  appId: string // 应用ID
  merchantId: string // 商户号
  merchantName: string // 商户名称
  apiKeyCipher: string // API Key密文
  apiSecretCipher: string // API Secret密文
  privateKeyCipher: string // 私钥密文
  publicKey: string // 公钥
  certCipher: string // 证书密文
  extConfig: string // 扩展配置(JSON)
  enabled: number // 启用状态：1启用 2禁用
  isDefault: number // 是否默认账号：1是 2否
  remark: string // 备注
  createTimes: number // 创建时间
  updateTimes: number // 更新时间
}

export type TenantPayChannel = {
  id: number // 通道ID
  tenantId: number // 租户ID
  platformId: number // 平台ID
  productId: number // 产品ID
  accountId: number // 租户支付账号ID
  channelCode: string // 通道编码
  channelName: string // 通道名称
  displayName: string // 前端展示名称
  icon: string // 图标
  currency: string // 币种
  sort: number // 排序
  visible: number // 是否显示：1显示 2隐藏
  enabled: number // 启用状态：1启用 2禁用
  singleMinAmount: number // 单笔最小金额，单位分
  singleMaxAmount: number // 单笔最大金额，单位分
  dailyMaxAmount: number // 单日最大金额，单位分
  dailyMaxCount: number // 单日最大次数
  feeType: number // 手续费类型：1比例 2固定
  feeRate: string // 手续费比例
  feeFixedAmount: number // 固定手续费，单位分
  extConfig: string // 扩展配置(JSON)
  remark: string // 备注
  createTimes: number // 创建时间
  updateTimes: number // 更新时间
}

export type TenantPayChannelRule = {
  id: number // 规则ID
  tenantId: number // 租户ID
  channelId: number // 通道ID
  ruleName: string // 规则名称
  priority: number // 优先级，越小越优先
  enabled: number // 启用状态：1启用 2禁用
  singleAmountMin: number // 单笔充值最小金额，单位分
  singleAmountMax: number // 单笔充值最大金额，单位分
  userTotalRechargeMin: number // 用户累计充值最小金额，单位分
  userTotalRechargeMax: number // 用户累计充值最大金额，单位分
  memberLevelMin: number // 会员等级最小值
  memberLevelMax: number // 会员等级最大值
  kycLevelMin: number // KYC等级最小值
  kycLevelMax: number // KYC等级最大值
  allowNewUser: number // 是否允许新用户：1是 2否
  allowOldUser: number // 是否允许老用户：1是 2否
  allowTags: string // 允许的用户标签(JSON数组)
  denyTags: string // 禁止的用户标签(JSON数组)
  remark: string // 备注
  createTimes: number // 创建时间
  updateTimes: number // 更新时间
}

export type ListTenantPayAccountsReq = {
  tenantId?: number // 租户ID
  platformId?: number // 平台ID
  keyword?: string // 关键字
  enabled?: number // 启用状态：1启用 2禁用
  cursor?: number // 分页游标
  limit?: number // 分页大小
}

export type CreateTenantPayAccountReq = {
  tenantId: number // 租户ID
  platformId: number // 平台ID
  accountCode: string // 账号编码
  accountName: string // 账号名称
  appId?: string // 应用ID
  merchantId: string // 商户号
  merchantName: string // 商户名称
  apiKeyCipher?: string // API Key密文
  apiSecretCipher?: string // API Secret密文
  privateKeyCipher?: string // 私钥密文
  publicKey?: string // 公钥
  certCipher?: string // 证书密文
  extConfig?: string // 扩展配置(JSON)
  enabled: number // 启用状态：1启用 2禁用
  isDefault: number // 是否默认账号：1是 2否
  remark?: string // 备注
}

export type UpdateTenantPayAccountReq = {
  id: number // 账号ID
  tenantId: number // 租户ID
  accountName: string // 账号名称
  appId?: string // 应用ID
  merchantId?: string // 商户号
  merchantName?: string // 商户名称
  apiKeyCipher?: string // API Key密文
  apiSecretCipher?: string // API Secret密文
  privateKeyCipher?: string // 私钥密文
  publicKey?: string // 公钥
  certCipher?: string // 证书密文
  extConfig?: string // 扩展配置(JSON)
  enabled: number // 启用状态：1启用 2禁用
  isDefault: number // 是否默认账号：1是 2否
  remark?: string // 备注
}

export type ListTenantPayChannelsReq = {
  tenantId?: number // 租户ID
  platformId?: number // 平台ID
  productId?: number // 产品ID
  accountId?: number // 账号ID
  keyword?: string // 关键字
  enabled?: number // 启用状态：1启用 2禁用
  visible?: number // 是否显示：1显示 2隐藏
  cursor?: number // 分页游标
  limit?: number // 分页大小
}

export type CreateTenantPayChannelReq = {
  tenantId: number // 租户ID
  platformId: number // 平台ID
  productId: number // 产品ID
  accountId: number // 账号ID
  channelCode: string // 通道编码
  channelName: string // 通道名称
  displayName: string // 前端展示名称
  icon?: string // 图标
  currency: string // 币种
  sort: number // 排序
  visible: number // 是否显示：1显示 2隐藏
  enabled: number // 启用状态：1启用 2禁用
  singleMinAmount: number // 单笔最小金额，单位分
  singleMaxAmount: number // 单笔最大金额，单位分
  dailyMaxAmount: number // 单日最大金额，单位分
  dailyMaxCount: number // 单日最大次数
  feeType: number // 手续费类型：1比例 2固定
  feeRate?: string // 手续费比例
  feeFixedAmount?: number // 固定手续费，单位分
  extConfig?: string // 扩展配置(JSON)
  remark?: string // 备注
}

export type UpdateTenantPayChannelReq = {
  id: number // 通道ID
  tenantId: number // 租户ID
  channelName: string // 通道名称
  displayName: string // 前端展示名称
  icon?: string // 图标
  currency: string // 币种
  sort: number // 排序
  visible: number // 是否显示：1显示 2隐藏
  enabled: number // 启用状态：1启用 2禁用
  singleMinAmount: number // 单笔最小金额，单位分
  singleMaxAmount: number // 单笔最大金额，单位分
  dailyMaxAmount: number // 单日最大金额，单位分
  dailyMaxCount: number // 单日最大次数
  feeType: number // 手续费类型：1比例 2固定
  feeRate?: string // 手续费比例
  feeFixedAmount?: number // 固定手续费，单位分
  extConfig?: string // 扩展配置(JSON)
  remark?: string // 备注
}

export type ListTenantPayChannelRulesReq = {
  tenantId?: number // 租户ID
  channelId?: number // 通道ID
  enabled?: number // 启用状态：1启用 2禁用
  cursor?: number // 分页游标
  limit?: number // 分页大小
}

export type CreateTenantPayChannelRuleReq = {
  tenantId: number // 租户ID
  channelId: number // 通道ID
  ruleName: string // 规则名称
  priority: number // 优先级，越小越优先
  enabled: number // 启用状态：1启用 2禁用
  singleAmountMin: number // 单笔充值最小金额，单位分
  singleAmountMax: number // 单笔充值最大金额，单位分
  userTotalRechargeMin: number // 用户累计充值最小金额，单位分
  userTotalRechargeMax: number // 用户累计充值最大金额，单位分
  memberLevelMin: number // 会员等级最小值
  memberLevelMax: number // 会员等级最大值
  kycLevelMin: number // KYC等级最小值
  kycLevelMax: number // KYC等级最大值
  allowNewUser: number // 是否允许新用户：1是 2否
  allowOldUser: number // 是否允许老用户：1是 2否
  allowTags?: string // 允许的用户标签(JSON数组)
  denyTags?: string // 禁止的用户标签(JSON数组)
  remark?: string // 备注
}

export type UpdateTenantPayChannelRuleReq = CreateTenantPayChannelRuleReq & { id: number }

export class TenantService {
  getOptions(): Promise<RespBase<OptionGroup[]>> {
    return getCoreOptions()
  }
  getTenantAccountList(params: ListTenantPayAccountsReq): Promise<RespBase<TenantPayAccount[]>> {
    return apiTenantPayAccountList(params)
  }
  getTenantAccountDetail(id: number, tenantId: number): Promise<RespBase<TenantPayAccount>> {
    return apiTenantPayAccountDetail(id, tenantId)
  }
  createTenantAccount(params: CreateTenantPayAccountReq): Promise<RespBase> {
    return apiTenantPayAccountCreate(params)
  }
  updateTenantAccount(params: UpdateTenantPayAccountReq): Promise<RespBase> {
    return apiTenantPayAccountUpdate(params)
  }
  getTenantChannelList(params: ListTenantPayChannelsReq): Promise<RespBase<TenantPayChannel[]>> {
    return apiTenantPayChannelList(params)
  }
  getTenantChannelDetail(id: number, tenantId: number): Promise<RespBase<TenantPayChannel>> {
    return apiTenantPayChannelDetail(id, tenantId)
  }
  createTenantChannel(params: CreateTenantPayChannelReq): Promise<RespBase> {
    return apiTenantPayChannelCreate(params)
  }
  updateTenantChannel(params: UpdateTenantPayChannelReq): Promise<RespBase> {
    return apiTenantPayChannelUpdate(params)
  }
  getTenantChannelRuleList(
    params: ListTenantPayChannelRulesReq,
  ): Promise<RespBase<TenantPayChannelRule[]>> {
    return apiTenantPayChannelRuleList(params)
  }
  getTenantChannelRuleDetail(
    id: number,
    tenantId: number,
  ): Promise<RespBase<TenantPayChannelRule>> {
    return apiTenantPayChannelRuleDetail(id, tenantId)
  }
  createTenantChannelRule(params: CreateTenantPayChannelRuleReq): Promise<RespBase> {
    return apiTenantPayChannelRuleCreate(params)
  }
  updateTenantChannelRule(params: UpdateTenantPayChannelRuleReq): Promise<RespBase> {
    return apiTenantPayChannelRuleUpdate(params)
  }
}

export const tenantService = new TenantService()
