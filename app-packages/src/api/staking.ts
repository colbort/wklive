import { authHttp, http } from './http'
import { compactParams } from './utils'
import type { RespBase } from '../types/api'
import type {
  CreateOrderReq,
  MyOrderDetailReq,
  MyOrderListReq,
  MyRedeemLogListReq,
  MyRewardLogListReq,
  ProductDetailReq,
  ProductListReq,
  RedeemReq,
  StakeOrder,
  StakeProduct,
  StakeRedeemLog,
  StakeRewardLog,
} from '../types/staking'

export function apiStakingListProducts(
  params: ProductListReq,
): Promise<RespBase & { data: StakeProduct[] }> {
  return http
    .get('/staking/products', { params: compactParams(params) })
    .then((res: { data: any }) => res.data)
}

export function apiStakingGetProduct(
  params: ProductDetailReq,
): Promise<RespBase & { data: StakeProduct }> {
  return http
    .get('/staking/products/detail', { params: compactParams(params) })
    .then((res: { data: any }) => res.data)
}

export function apiStakingCreateOrder(
  params: CreateOrderReq,
): Promise<RespBase & { data: { id: number; orderNo: string } }> {
  return authHttp
    .post('/staking/orders', params)
    .then((res: { data: any }) => res.data)
}

export function apiStakingListMyOrders(
  params: MyOrderListReq,
): Promise<RespBase & { data: StakeOrder[] }> {
  return authHttp
    .get('/staking/my/orders', { params: compactParams(params) })
    .then((res: { data: any }) => res.data)
}

export function apiStakingGetMyOrder(
  params: MyOrderDetailReq,
): Promise<RespBase & { data: StakeOrder }> {
  return authHttp
    .get('/staking/my/orders/detail', { params: compactParams(params) })
    .then((res: { data: any }) => res.data)
}

export function apiStakingListMyRewardLogs(
  params: MyRewardLogListReq,
): Promise<RespBase & { data: StakeRewardLog[] }> {
  return authHttp
    .get('/staking/my/reward-logs', { params: compactParams(params) })
    .then((res: { data: any }) => res.data)
}

export function apiStakingRedeem(
  params: RedeemReq,
): Promise<RespBase & { data: { success: number; redeemNo: string } }> {
  return authHttp
    .post('/staking/redeem', params)
    .then((res: { data: any }) => res.data)
}

export function apiStakingListMyRedeemLogs(
  params: MyRedeemLogListReq,
): Promise<RespBase & { data: StakeRedeemLog[] }> {
  return authHttp
    .get('/staking/my/redeem-logs', { params: compactParams(params) })
    .then((res: { data: any }) => res.data)
}
