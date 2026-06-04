import { http } from '@/api/http'
import { buildPath, compactParams } from '@/api/utils'
import type { OptionsGroup, RespBase } from '@/types/api'
import type {
  AvailableRechargeChannel,
  CancelMyRechargeOrderReq,
  CreateCryptoRechargeOrderReq,
  CreateRechargeOrderReq,
  CreateWithdrawOrderReq,
  CryptoRechargeAddress,
  CryptoRechargeOrderData,
  CryptoRechargeTx,
  GetMyCryptoRechargeAddressReq,
  GetMyCryptoRechargeTxReq,
  GetMyRechargeOrderReq,
  GetMyRechargeStatReq,
  GetMyWithdrawOrderReq,
  ListMyCryptoRechargeAddressesReq,
  ListMyCryptoRechargeTxsReq,
  ListAvailableRechargeChannelsReq,
  ListMyRechargeOrdersReq,
  ListMyWithdrawOrdersReq,
  QueryMyRechargeOrderStatusReq,
  RechargeOrder,
  UserRechargeStat,
  WithdrawOrder,
} from '@/types/payment'

export function apiGetPaymentOptions(): Promise<RespBase & { data: OptionsGroup[] }> {
  return http.get('/payment/options').then((res) => res.data)
}

export function apiGetMyRechargeStat(
  params: GetMyRechargeStatReq,
): Promise<RespBase & { data: UserRechargeStat }> {
  return http
    .get('/payment/recharge/stat', { params: compactParams(params) })
    .then((res) => res.data)
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

export function apiCreateCryptoRechargeOrder(
  params: CreateCryptoRechargeOrderReq,
): Promise<RespBase & { data: CryptoRechargeOrderData }> {
  return http.post('/payment/crypto/recharge/orders', params).then((res) => res.data)
}

export function apiGetMyRechargeOrder(
  params: GetMyRechargeOrderReq,
): Promise<RespBase & { data: RechargeOrder }> {
  return http
    .get(buildPath('/payment/recharge/orders/:orderNo', { orderNo: params.orderNo }))
    .then((res) => res.data)
}

export function apiListMyRechargeOrders(
  params: ListMyRechargeOrdersReq,
): Promise<RespBase & { data: RechargeOrder[] }> {
  return http
    .get('/payment/recharge/orders', { params: compactParams(params) })
    .then((res) => res.data)
}

export function apiCancelMyRechargeOrder(params: CancelMyRechargeOrderReq): Promise<RespBase> {
  return http
    .post(buildPath('/payment/recharge/orders/:orderNo/cancel', { orderNo: params.orderNo }))
    .then((res) => res.data)
}

export function apiQueryMyRechargeOrderStatus(
  params: QueryMyRechargeOrderStatusReq,
): Promise<RespBase & { data: RechargeOrder }> {
  return http
    .get(buildPath('/payment/recharge/orders/:orderNo/status', { orderNo: params.orderNo }))
    .then((res) => res.data)
}

export function apiGetMyCryptoRechargeAddress(
  params: GetMyCryptoRechargeAddressReq,
): Promise<RespBase & { data: CryptoRechargeAddress }> {
  return http
    .get('/payment/crypto/recharge/address', { params: compactParams(params) })
    .then((res) => res.data)
}

export function apiListMyCryptoRechargeAddresses(
  params: ListMyCryptoRechargeAddressesReq,
): Promise<RespBase & { data: CryptoRechargeAddress[] }> {
  return http
    .get('/payment/crypto/recharge/addresses', { params: compactParams(params) })
    .then((res) => res.data)
}

export function apiListMyCryptoRechargeTxs(
  params: ListMyCryptoRechargeTxsReq,
): Promise<RespBase & { data: CryptoRechargeTx[] }> {
  return http
    .get('/payment/crypto/recharge/txs', { params: compactParams(params) })
    .then((res) => res.data)
}

export function apiGetMyCryptoRechargeTx(
  params: GetMyCryptoRechargeTxReq,
): Promise<RespBase & { data: CryptoRechargeTx }> {
  return http
    .get(buildPath('/payment/crypto/recharge/txs/:id', { id: params.id }), {
      params: compactParams({ txHash: params.txHash }),
    })
    .then((res) => res.data)
}

export function apiCreateWithdrawOrder(
  params: CreateWithdrawOrderReq,
): Promise<RespBase & { data: number }> {
  return http.post('/payment/withdraw/orders', params).then((res) => res.data)
}

export function apiListMyWithdrawOrders(
  params: ListMyWithdrawOrdersReq,
): Promise<RespBase & { data: WithdrawOrder[] }> {
  return http
    .get('/payment/withdraw/orders', { params: compactParams(params) })
    .then((res) => res.data)
}

export function apiGetMyWithdrawOrder(
  params: GetMyWithdrawOrderReq,
): Promise<RespBase & { data: WithdrawOrder }> {
  return http
    .get(buildPath('/payment/withdraw/orders/:orderNo', { orderNo: params.orderNo }))
    .then((res) => res.data)
}
