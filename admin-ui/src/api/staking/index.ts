import { get, post } from '@/utils/request'
import type { OptionGroup, RespBase, StakeOrder, StakeProduct, StakeRedeemLog, StakeRewardLog } from '@/services'

export function apiStakingListProducts(params: Record<string, any>): Promise<RespBase<StakeProduct[]>> {
  return get<StakeProduct[]>('/admin/staking/products', params)
}

export function apiStakingGetProduct(params: Record<string, any>): Promise<RespBase<StakeProduct>> {
  return get<StakeProduct>('/admin/staking/products/detail', params)
}

export function apiStakingCreateProduct(params: Record<string, any>): Promise<RespBase<number>> {
  return post<number>('/admin/staking/products', params)
}

export function apiStakingUpdateProduct(params: Record<string, any>): Promise<RespBase<boolean>> {
  return post<boolean>('/admin/staking/products/update', params)
}

export function apiStakingChangeProductStatus(params: Record<string, any>): Promise<RespBase<boolean>> {
  return post<boolean>('/admin/staking/products/status', params)
}

export function apiStakingListOrders(params: Record<string, any>): Promise<RespBase<StakeOrder[]>> {
  return get<StakeOrder[]>('/admin/staking/orders', params)
}

export function apiStakingGetOrder(params: Record<string, any>): Promise<RespBase<StakeOrder>> {
  return get<StakeOrder>('/admin/staking/orders/detail', params)
}

export function apiStakingListRewardLogs(params: Record<string, any>): Promise<RespBase<StakeRewardLog[]>> {
  return get<StakeRewardLog[]>('/admin/staking/reward-logs', params)
}

export function apiStakingListRedeemLogs(params: Record<string, any>): Promise<RespBase<StakeRedeemLog[]>> {
  return get<StakeRedeemLog[]>('/admin/staking/redeem-logs', params)
}

export function apiStakingManualReward(params: Record<string, any>): Promise<RespBase<boolean>> {
  return post<boolean>('/admin/staking/manual-reward', params)
}

export function apiStakingManualRedeem(params: Record<string, any>): Promise<RespBase<boolean>> {
  return post<boolean>('/admin/staking/manual-redeem', params)
}

export function apiOptions(): Promise<RespBase<OptionGroup[]>> {
  return get('/admin/staking/options')
}
