import { get, post } from '@/utils/request'
import type {
  AdminManualRedeemReq,
  AdminManualRewardReq,
  AdminOrderDetailReq,
  AdminOrderListReq,
  AdminProductChangeStatusReq,
  AdminProductCreateReq,
  AdminProductDetailReq,
  AdminProductListReq,
  AdminProductUpdateReq,
  AdminRedeemLogListReq,
  AdminRewardLogListReq,
  OptionGroup,
  RespBase,
  StakeOrder,
  StakeProduct,
  StakeRedeemLog,
  StakeRewardLog,
} from '@/services'

export function apiStakingListProducts(
  params: AdminProductListReq,
): Promise<RespBase<StakeProduct[]>> {
  return get<StakeProduct[]>('/admin/staking/products', params)
}

export function apiStakingGetProduct(params: AdminProductDetailReq): Promise<RespBase<StakeProduct>> {
  return get<StakeProduct>('/admin/staking/products/detail', params)
}

export function apiStakingCreateProduct(params: AdminProductCreateReq): Promise<RespBase<number>> {
  return post<number>('/admin/staking/products', params)
}

export function apiStakingUpdateProduct(params: AdminProductUpdateReq): Promise<RespBase<boolean>> {
  return post<boolean>('/admin/staking/products/update', params)
}

export function apiStakingChangeProductStatus(
  params: AdminProductChangeStatusReq,
): Promise<RespBase<boolean>> {
  return post<boolean>('/admin/staking/products/status', params)
}

export function apiStakingListOrders(params: AdminOrderListReq): Promise<RespBase<StakeOrder[]>> {
  return get<StakeOrder[]>('/admin/staking/orders', params)
}

export function apiStakingGetOrder(params: AdminOrderDetailReq): Promise<RespBase<StakeOrder>> {
  return get<StakeOrder>('/admin/staking/orders/detail', params)
}

export function apiStakingListRewardLogs(
  params: AdminRewardLogListReq,
): Promise<RespBase<StakeRewardLog[]>> {
  return get<StakeRewardLog[]>('/admin/staking/reward-logs', params)
}

export function apiStakingListRedeemLogs(
  params: AdminRedeemLogListReq,
): Promise<RespBase<StakeRedeemLog[]>> {
  return get<StakeRedeemLog[]>('/admin/staking/redeem-logs', params)
}

export function apiStakingManualReward(params: AdminManualRewardReq): Promise<RespBase<boolean>> {
  return post<boolean>('/admin/staking/manual-reward', params)
}

export function apiStakingManualRedeem(params: AdminManualRedeemReq): Promise<RespBase<boolean>> {
  return post<boolean>('/admin/staking/manual-redeem', params)
}

export function apiOptions(): Promise<RespBase<OptionGroup[]>> {
  return get('/admin/staking/options')
}
