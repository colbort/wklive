import { http } from '@/api/http'
import { compactParams } from '@/api/utils'
import type { RespBase } from '@/types/api'
import type {
  AppCreateOrderReq,
  AppMyOrderDetailReq,
  AppMyOrderListReq,
  AppMyRedeemLogListReq,
  AppMyRewardLogListReq,
  AppProductDetailReq,
  AppProductListReq,
  AppRedeemReq,
  StakeOrder,
  StakeProduct,
  StakeRedeemLog,
  StakeRewardLog,
} from '@/types/staking'

export function apiStakingListProducts(
  params: AppProductListReq,
): Promise<RespBase & { data: StakeProduct[] }> {
  return http.get('/staking/products', { params: compactParams(params) }).then((res) => res.data)
}

export function apiStakingGetProduct(
  params: AppProductDetailReq,
): Promise<RespBase & { data: StakeProduct }> {
  return http.get('/staking/products/detail', { params: compactParams(params) }).then((res) => res.data)
}

export function apiStakingCreateOrder(
  params: AppCreateOrderReq,
): Promise<RespBase & { id: number; orderNo: string }> {
  return http.post('/staking/orders', params).then((res) => res.data)
}

export function apiStakingListMyOrders(
  params: AppMyOrderListReq,
): Promise<RespBase & { data: StakeOrder[] }> {
  return http.get('/staking/my/orders', { params: compactParams(params) }).then((res) => res.data)
}

export function apiStakingGetMyOrder(
  params: AppMyOrderDetailReq,
): Promise<RespBase & { data: StakeOrder }> {
  return http.get('/staking/my/orders/detail', { params: compactParams(params) }).then((res) => res.data)
}

export function apiStakingListMyRewardLogs(
  params: AppMyRewardLogListReq,
): Promise<RespBase & { data: StakeRewardLog[] }> {
  return http.get('/staking/my/reward-logs', { params: compactParams(params) }).then((res) => res.data)
}

export function apiStakingRedeem(
  params: AppRedeemReq,
): Promise<RespBase & { success: number; redeemNo: string }> {
  return http.post('/staking/redeem', params).then((res) => res.data)
}

export function apiStakingListMyRedeemLogs(
  params: AppMyRedeemLogListReq,
): Promise<RespBase & { data: StakeRedeemLog[] }> {
  return http.get('/staking/my/redeem-logs', { params: compactParams(params) }).then((res) => res.data)
}
