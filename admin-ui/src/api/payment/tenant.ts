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
} from '@/services'

export function apiTenantPayPlatformList(
  params: ListTenantPayPlatformsReq,
): Promise<RespBase<TenantPayPlatform[]>> {
  return get<TenantPayPlatform[]>('/admin/payment/tenant-platforms/list', params)
}

export function apiTenantPayPlatformDetail(
  id: number,
  tenantId: number,
): Promise<RespBase<TenantPayPlatform>> {
  return get<TenantPayPlatform[]>('/admin/payment/tenant-platforms', { id, tenantId }) as Promise<
    RespBase<TenantPayPlatform>
  >
}

export function apiOpenTenantPayPlatform(params: OpenTenantPayPlatformReq): Promise<RespBase> {
  return post('/admin/payment/tenant-platforms/open', params)
}

export function apiUpdateTenantPayPlatform(params: UpdateTenantPayPlatformReq): Promise<RespBase> {
  return put('/admin/payment/tenant-platforms', params)
}

export function apiTenantPayAccountList(
  params: ListTenantPayAccountsReq,
): Promise<RespBase<TenantPayAccount[]>> {
  return get<TenantPayAccount[]>('/admin/payment/tenant-accounts/list', params)
}

export function apiTenantPayAccountDetail(
  id: number,
  tenantId: number,
): Promise<RespBase<TenantPayAccount>> {
  return get<TenantPayAccount>('/admin/payment/tenant-accounts', { id, tenantId })
}

export function apiTenantPayAccountCreate(params: CreateTenantPayAccountReq): Promise<RespBase> {
  return post('/admin/payment/tenant-accounts', params)
}

export function apiTenantPayAccountUpdate(params: UpdateTenantPayAccountReq): Promise<RespBase> {
  return put('/admin/payment/tenant-accounts', params)
}

export function apiTenantPayChannelList(
  params: ListTenantPayChannelsReq,
): Promise<RespBase<TenantPayChannel[]>> {
  return get<TenantPayChannel[]>('/admin/payment/tenant-channels/list', params)
}

export function apiTenantPayChannelDetail(
  id: number,
  tenantId: number,
): Promise<RespBase<TenantPayChannel>> {
  return get<TenantPayChannel>('/admin/payment/tenant-channels', { id, tenantId })
}

export function apiTenantPayChannelCreate(params: CreateTenantPayChannelReq): Promise<RespBase> {
  return post('/admin/payment/tenant-channels', params)
}

export function apiTenantPayChannelUpdate(params: UpdateTenantPayChannelReq): Promise<RespBase> {
  return put('/admin/payment/tenant-channels', params)
}

export function apiTenantPayChannelRuleList(
  params: ListTenantPayChannelRulesReq,
): Promise<RespBase<TenantPayChannelRule[]>> {
  return get<TenantPayChannelRule[]>('/admin/payment/tenant-channel-rules/list', params)
}

export function apiTenantPayChannelRuleDetail(
  id: number,
  tenantId: number,
): Promise<RespBase<TenantPayChannelRule>> {
  return get<TenantPayChannelRule>('/admin/payment/tenant-channel-rules', { id, tenantId })
}

export function apiTenantPayChannelRuleCreate(
  params: CreateTenantPayChannelRuleReq,
): Promise<RespBase> {
  return post('/admin/payment/tenant-channel-rules', params)
}

export function apiTenantPayChannelRuleUpdate(
  params: UpdateTenantPayChannelRuleReq,
): Promise<RespBase> {
  return put('/admin/payment/tenant-channel-rules', params)
}
