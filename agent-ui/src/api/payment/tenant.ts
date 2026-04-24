import { get, post, put } from '@/utils/request'
import type {
  RespBase,
  TenantPayPlatform,
  TenantPayAccount,
  TenantPayChannel,
  TenantPayChannelRule,
  OpenTenantPayPlatformReq,
  UpdateTenantPayPlatformReq,
  ListTenantPayPlatformsReq,
  CreateTenantPayAccountReq,
  UpdateTenantPayAccountReq,
  ListTenantPayAccountsReq,
  CreateTenantPayChannelReq,
  UpdateTenantPayChannelReq,
  ListTenantPayChannelsReq,
  CreateTenantPayChannelRuleReq,
  UpdateTenantPayChannelRuleReq,
  ListTenantPayChannelRulesReq,
  OptionGroup,
} from '@/services'

export function apiTenantPayPlatformList(
  params: ListTenantPayPlatformsReq,
): Promise<RespBase<TenantPayPlatform[]>> {
  return get<TenantPayPlatform[]>('/admin/payment/tenant-platforms', params)
}

export function apiTenantPayPlatformDetail(
  id: number,
  tenantId: number,
): Promise<RespBase<TenantPayPlatform>> {
  return get<TenantPayPlatform>('/admin/payment/tenant-platform', { id, tenantId })
}

export function apiOpenTenantPayPlatform(params: OpenTenantPayPlatformReq): Promise<RespBase> {
  return post('/admin/payment/tenant-platform', params)
}

export function apiUpdateTenantPayPlatform(params: UpdateTenantPayPlatformReq): Promise<RespBase> {
  return put('/admin/payment/tenant-platform', params)
}

export function apiTenantPayAccountList(
  params: ListTenantPayAccountsReq,
): Promise<RespBase<TenantPayAccount[]>> {
  return get<TenantPayAccount[]>('/admin/payment/tenant-accounts', params)
}

export function apiTenantPayAccountDetail(
  id: number,
  tenantId: number,
): Promise<RespBase<TenantPayAccount>> {
  return get<TenantPayAccount>('/admin/payment/tenant-account', { id, tenantId })
}

export function apiTenantPayAccountCreate(params: CreateTenantPayAccountReq): Promise<RespBase> {
  return post('/admin/payment/tenant-account', params)
}

export function apiTenantPayAccountUpdate(params: UpdateTenantPayAccountReq): Promise<RespBase> {
  return put('/admin/payment/tenant-account', params)
}

export function apiTenantPayChannelList(
  params: ListTenantPayChannelsReq,
): Promise<RespBase<TenantPayChannel[]>> {
  return get<TenantPayChannel[]>('/admin/payment/tenant-channels', params)
}

export function apiTenantPayChannelDetail(
  id: number,
  tenantId: number,
): Promise<RespBase<TenantPayChannel>> {
  return get<TenantPayChannel>('/admin/payment/tenant-channel', { id, tenantId })
}

export function apiTenantPayChannelCreate(params: CreateTenantPayChannelReq): Promise<RespBase> {
  return post('/admin/payment/tenant-channel', params)
}

export function apiTenantPayChannelUpdate(params: UpdateTenantPayChannelReq): Promise<RespBase> {
  return put('/admin/payment/tenant-channel', params)
}

export function apiTenantPayChannelRuleList(
  params: ListTenantPayChannelRulesReq,
): Promise<RespBase<TenantPayChannelRule[]>> {
  return get<TenantPayChannelRule[]>('/admin/payment/tenant-channel-rules', params)
}

export function apiTenantPayChannelRuleDetail(
  id: number,
  tenantId: number,
): Promise<RespBase<TenantPayChannelRule>> {
  return get<TenantPayChannelRule>('/admin/payment/tenant-channel-rule', { id, tenantId })
}

export function apiTenantPayChannelRuleCreate(
  params: CreateTenantPayChannelRuleReq,
): Promise<RespBase> {
  return post('/admin/payment/tenant-channel-rule', params)
}

export function apiTenantPayChannelRuleUpdate(
  params: UpdateTenantPayChannelRuleReq,
): Promise<RespBase> {
  return put('/admin/payment/tenant-channel-rule', params)
}

export function apiOptions(): Promise<RespBase<OptionGroup[]>> {
  return get('/admin/payment/options')
}
