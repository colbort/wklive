import { http } from '@/api/http'
import { buildPath, compactParams } from '@/api/utils'
import type { RespBase } from '@/types/api'
import type {
  AvailableRechargeChannel,
  CancelMyRechargeOrderReq,
  CreateRechargeOrderReq,
  CreateWithdrawOrderReq,
  GetMyRechargeOrderReq,
  GetMyRechargeStatReq,
  GetMyWithdrawOrderReq,
  ListAvailableRechargeChannelsReq,
  ListMyRechargeOrdersReq,
  ListMyWithdrawOrdersReq,
  QueryMyRechargeOrderStatusReq,
  RechargeOrder,
  UserRechargeStat,
  WithdrawOrder,
} from '@/types/payment'

export function apiGetMyRechargeStat(
  params: GetMyRechargeStatReq,
): Promise<RespBase & { data: UserRechargeStat }> {
  return http.get('/payment/recharge/stat', { params: compactParams(params) }).then((res) => res.data)
}

export function apiListAvailableRechargeChannels(
  params: ListAvailableRechargeChannelsReq,
): Promise<RespBase & { data: AvailableRechargeChannel[] }> {
  return http.post('/payment/channels/available', params).then((res) => res.data)
}

export function apiCreateRechargeOrder(
  params: CreateRechargeOrderReq,
): Promise<RespBase & { data: RechargeOrder }> {
  return http.post('/payment/create/recharge/orders', params).then((res) => res.data)
}

export function apiGetMyRechargeOrder(
  params: GetMyRechargeOrderReq,
): Promise<RespBase & { data: RechargeOrder }> {
  return http
    .get(buildPath('/payment/recharge/orders/:orderNo', { orderNo: params.orderNo }), {
      params: compactParams({ tenantId: params.tenantId }),
    })
    .then((res) => res.data)
}

export function apiListMyRechargeOrders(
  params: ListMyRechargeOrdersReq,
): Promise<RespBase & { data: RechargeOrder[] }> {
  return http.get('/payment/recharge/orders', { params: compactParams(params) }).then((res) => res.data)
}

export function apiCancelMyRechargeOrder(params: CancelMyRechargeOrderReq): Promise<RespBase> {
  return http
    .post(buildPath('/payment/recharge/orders/:orderNo/cancel', { orderNo: params.orderNo }), {
      tenantId: params.tenantId,
    })
    .then((res) => res.data)
}

export function apiQueryMyRechargeOrderStatus(
  params: QueryMyRechargeOrderStatusReq,
): Promise<RespBase & { data: RechargeOrder }> {
  return http
    .get(buildPath('/payment/recharge/orders/:orderNo/status', { orderNo: params.orderNo }), {
      params: compactParams({ tenantId: params.tenantId }),
    })
    .then((res) => res.data)
}

export function apiCreateWithdrawOrder(
  params: CreateWithdrawOrderReq,
): Promise<RespBase & { id: number }> {
  return http.post('/payment/withdraw/orders', params).then((res) => res.data)
}

export function apiListMyWithdrawOrders(
  params: ListMyWithdrawOrdersReq,
): Promise<RespBase & { data: WithdrawOrder[] }> {
  return http.get('/payment/withdraw/orders', { params: compactParams(params) }).then((res) => res.data)
}

export function apiGetMyWithdrawOrder(
  params: GetMyWithdrawOrderReq,
): Promise<RespBase & { data: WithdrawOrder }> {
  return http
    .get(buildPath('/payment/withdraw/orders/:orderNo', { orderNo: params.orderNo }))
    .then((res) => res.data)
}
