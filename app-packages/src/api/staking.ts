import { authHttp, http } from './http'
import { compactParams } from './utils'
import type { OptionsGroup, RespBase } from '../types/api'
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
} from '../types/staking'

export function apiGetStakingOptions(): Promise<
  RespBase & { data: OptionsGroup[] }
> {
  return http.get('/staking/options').then((res: { data: any }) => res.data)
}

export function apiStakingListProducts(
  params: AppProductListReq,
): Promise<RespBase & { data: StakeProduct[] }> {
  return http
    .get('/staking/products', { params: compactParams(params) })
    .then((res: { data: any }) => res.data)
}

export function apiStakingGetProduct(
  params: AppProductDetailReq,
): Promise<RespBase & { data: StakeProduct }> {
  return http
    .get('/staking/products/detail', { params: compactParams(params) })
    .then((res: { data: any }) => res.data)
}

export function apiStakingCreateOrder(
  params: AppCreateOrderReq,
): Promise<RespBase & { data: { id: number; orderNo: string } }> {
  return authHttp
    .post('/staking/orders', params)
    .then((res: { data: any }) => res.data)
}

export function apiStakingListMyOrders(
  params: AppMyOrderListReq,
): Promise<RespBase & { data: StakeOrder[] }> {
  return authHttp
    .get('/staking/my/orders', { params: compactParams(params) })
    .then((res: { data: any }) => res.data)
}

export function apiStakingGetMyOrder(
  params: AppMyOrderDetailReq,
): Promise<RespBase & { data: StakeOrder }> {
  return authHttp
    .get('/staking/my/orders/detail', { params: compactParams(params) })
    .then((res: { data: any }) => res.data)
}

export function apiStakingListMyRewardLogs(
  params: AppMyRewardLogListReq,
): Promise<RespBase & { data: StakeRewardLog[] }> {
  return authHttp
    .get('/staking/my/reward-logs', { params: compactParams(params) })
    .then((res: { data: any }) => res.data)
}

export function apiStakingRedeem(
  params: AppRedeemReq,
): Promise<RespBase & { data: { success: number; redeemNo: string } }> {
  return authHttp
    .post('/staking/redeem', params)
    .then((res: { data: any }) => res.data)
}

export function apiStakingListMyRedeemLogs(
  params: AppMyRedeemLogListReq,
): Promise<RespBase & { data: StakeRedeemLog[] }> {
  return authHttp
    .get('/staking/my/redeem-logs', { params: compactParams(params) })
    .then((res: { data: any }) => res.data)
}
