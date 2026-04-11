import type { RespBase } from '@/services'
import {
  apiOpenTenantPayPlatform,
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
  apiTenantPayPlatformDetail,
  apiTenantPayPlatformList,
  apiUpdateTenantPayPlatform,
} from '@/api/payment/tenant'

export type TenantPayPlatform = {
  id: number
  tenantId: number
  platformId: number
  status: number
  openStatus: number
  remark: string
  createTimes: number
  updateTimes: number
}

export type TenantPayAccount = {
  id: number
  tenantId: number
  tenantPayPlatformId: number
  platformId: number
  accountCode: string
  accountName: string
  appId: string
  merchantId: string
  merchantName: string
  publicKey: string
  extConfig: string
  status: number
  isDefault: number
  remark: string
  createTimes: number
  updateTimes: number
}

export type TenantPayChannel = {
  id: number
  tenantId: number
  platformId: number
  productId: number
  accountId: number
  channelCode: string
  channelName: string
  displayName: string
  icon: string
  currency: string
  sort: number
  visible: number
  status: number
  singleMinAmount: number
  singleMaxAmount: number
  dailyMaxAmount: number
  dailyMaxCount: number
  feeType: number
  feeRate: string
  feeFixedAmount: number
  extConfig: string
  remark: string
  createTimes: number
  updateTimes: number
}

export type TenantPayChannelRule = {
  id: number
  tenantId: number
  channelId: number
  ruleName: string
  priority: number
  status: number
  singleAmountMin: number
  singleAmountMax: number
  userTotalRechargeMin: number
  userTotalRechargeMax: number
  memberLevelMin: number
  memberLevelMax: number
  kycLevelMin: number
  kycLevelMax: number
  allowNewUser: number
  allowOldUser: number
  allowTags: string
  denyTags: string
  remark: string
  createTimes: number
  updateTimes: number
}

export type ListTenantPayPlatformsReq = {
  tenantId?: number
  platformId?: number
  status?: number
  openStatus?: number
  cursor?: string | null
  limit?: number
}

export type OpenTenantPayPlatformReq = {
  tenantId: number
  platformId: number
  status: number
  openStatus: number
  remark?: string
}

export type UpdateTenantPayPlatformReq = {
  id: number
  tenantId: number
  status: number
  openStatus: number
  remark?: string
}

export type ListTenantPayAccountsReq = {
  tenantId?: number
  platformId?: number
  tenantPayPlatformId?: number
  keyword?: string
  status?: number
  cursor?: string | null
  limit?: number
}

export type CreateTenantPayAccountReq = {
  tenantId: number
  tenantPayPlatformId: number
  platformId: number
  accountCode: string
  accountName: string
  appId?: string
  merchantId?: string
  merchantName?: string
  apiKeyCipher?: string
  apiSecretCipher?: string
  privateKeyCipher?: string
  publicKey?: string
  certCipher?: string
  extConfig?: string
  status: number
  isDefault: number
  remark?: string
}

export type UpdateTenantPayAccountReq = {
  id: number
  tenantId: number
  accountName: string
  appId?: string
  merchantId?: string
  merchantName?: string
  apiKeyCipher?: string
  apiSecretCipher?: string
  privateKeyCipher?: string
  publicKey?: string
  certCipher?: string
  extConfig?: string
  status: number
  isDefault: number
  remark?: string
}

export type ListTenantPayChannelsReq = {
  tenantId?: number
  platformId?: number
  productId?: number
  accountId?: number
  keyword?: string
  status?: number
  visible?: number
  cursor?: string | null
  limit?: number
}

export type CreateTenantPayChannelReq = {
  tenantId: number
  platformId: number
  productId: number
  accountId: number
  channelCode: string
  channelName: string
  displayName: string
  icon?: string
  currency: string
  sort: number
  visible: number
  status: number
  singleMinAmount: number
  singleMaxAmount: number
  dailyMaxAmount: number
  dailyMaxCount: number
  feeType: number
  feeRate?: string
  feeFixedAmount?: number
  extConfig?: string
  remark?: string
}

export type UpdateTenantPayChannelReq = {
  id: number
  tenantId: number
  channelName: string
  displayName: string
  icon?: string
  currency: string
  sort: number
  visible: number
  status: number
  singleMinAmount: number
  singleMaxAmount: number
  dailyMaxAmount: number
  dailyMaxCount: number
  feeType: number
  feeRate?: string
  feeFixedAmount?: number
  extConfig?: string
  remark?: string
}

export type ListTenantPayChannelRulesReq = {
  tenantId?: number
  channelId?: number
  status?: number
  cursor?: string | null
  limit?: number
}

export type CreateTenantPayChannelRuleReq = {
  tenantId: number
  channelId: number
  ruleName: string
  priority: number
  status: number
  singleAmountMin: number
  singleAmountMax: number
  userTotalRechargeMin: number
  userTotalRechargeMax: number
  memberLevelMin: number
  memberLevelMax: number
  kycLevelMin: number
  kycLevelMax: number
  allowNewUser: number
  allowOldUser: number
  allowTags?: string
  denyTags?: string
  remark?: string
}

export type UpdateTenantPayChannelRuleReq = CreateTenantPayChannelRuleReq & { id: number }

export class TenantService {
  getTenantPlatformList(params: ListTenantPayPlatformsReq): Promise<RespBase<TenantPayPlatform[]>> {
    return apiTenantPayPlatformList(params)
  }
  getTenantPlatformDetail(id: number, tenantId: number): Promise<RespBase<TenantPayPlatform>> {
    return apiTenantPayPlatformDetail(id, tenantId)
  }
  openTenantPlatform(params: OpenTenantPayPlatformReq): Promise<RespBase> {
    return apiOpenTenantPayPlatform(params)
  }
  updateTenantPlatform(params: UpdateTenantPayPlatformReq): Promise<RespBase> {
    return apiUpdateTenantPayPlatform(params)
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
