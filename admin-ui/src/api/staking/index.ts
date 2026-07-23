import { get, post } from '@/utils/request'
import type {
  ManualRedeemReq,
  ManualRewardReq,
  OrderDetailReq,
  OrderListReq,
  ProductChangeStatusReq,
  ProductCreateReq,
  ProductDetailReq,
  ProductListReq,
  ProductUpdateReq,
  RedeemLogListReq,
  RewardLogListReq,
  RespBase,
  StakeOrder,
  StakeProduct,
  StakeRedeemLog,
  StakeRewardLog,
} from '@/services'

export function apiStakingListProducts(
  params: ProductListReq,
): Promise<RespBase<StakeProduct[]>> {
  return get<StakeProduct[]>('/admin/staking/products', params)
}

export function apiStakingGetProduct(
  params: ProductDetailReq,
): Promise<RespBase<StakeProduct>> {
  return get<StakeProduct>('/admin/staking/products/detail', params)
}

export function apiStakingCreateProduct(params: ProductCreateReq): Promise<RespBase<number>> {
  return post<number>('/admin/staking/products', params)
}

export function apiStakingUpdateProduct(params: ProductUpdateReq): Promise<RespBase<boolean>> {
  return post<boolean>('/admin/staking/products/update', params)
}

export function apiStakingChangeProductStatus(
  params: ProductChangeStatusReq,
): Promise<RespBase<boolean>> {
  return post<boolean>('/admin/staking/products/status', params)
}

export function apiStakingListOrders(params: OrderListReq): Promise<RespBase<StakeOrder[]>> {
  return get<StakeOrder[]>('/admin/staking/orders', params)
}

export function apiStakingGetOrder(params: OrderDetailReq): Promise<RespBase<StakeOrder>> {
  return get<StakeOrder>('/admin/staking/orders/detail', params)
}

export function apiStakingListRewardLogs(
  params: RewardLogListReq,
): Promise<RespBase<StakeRewardLog[]>> {
  return get<StakeRewardLog[]>('/admin/staking/reward-logs', params)
}

export function apiStakingListRedeemLogs(
  params: RedeemLogListReq,
): Promise<RespBase<StakeRedeemLog[]>> {
  return get<StakeRedeemLog[]>('/admin/staking/redeem-logs', params)
}

export function apiStakingManualReward(params: ManualRewardReq): Promise<RespBase<boolean>> {
  return post<boolean>('/admin/staking/manual-reward', params)
}

export function apiStakingManualRedeem(params: ManualRedeemReq): Promise<RespBase<boolean>> {
  return post<boolean>('/admin/staking/manual-redeem', params)
}
