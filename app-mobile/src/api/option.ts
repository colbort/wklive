import { http } from '@/api/http'
import { compactParams } from '@/api/utils'
import type { OptionsGroup, RespBase } from '@/types/api'
import type {
  AppCancelOrderReq,
  AppExerciseReq,
  AppGetContractDetailReq,
  AppGetOrderDetailReq,
  AppGetPositionDetailReq,
  AppListAccountsReq,
  AppListBillsReq,
  AppListContractsReq,
  AppListCurrentOrdersReq,
  AppListExercisesReq,
  AppListHistoryOrdersReq,
  AppListPositionsReq,
  AppListTradesReq,
  AppPlaceOrderReq,
  OptionAccount,
  OptionBill,
  OptionContractDetail,
  OptionExerciseDetail,
  OptionOrderDetail,
  OptionPositionDetail,
  OptionTradeDetail,
} from '@/types/option'

export function apiOptionGetOptions(): Promise<RespBase & { data: OptionsGroup[] }> {
  return http.get('/option/options').then((res) => res.data)
}

export function apiOptionListContracts(
  params: AppListContractsReq,
): Promise<RespBase & { data: OptionContractDetail[] }> {
  return http.get('/option/contracts', { params: compactParams(params) }).then((res) => res.data)
}

export function apiOptionGetContractDetail(
  params: AppGetContractDetailReq,
): Promise<RespBase & { data: OptionContractDetail }> {
  return http
    .get('/option/contracts/detail', { params: compactParams(params) })
    .then((res) => res.data)
}

export function apiOptionPlaceOrder(
  params: AppPlaceOrderReq,
): Promise<RespBase & { data: { orderNo: string; orderId: number } }> {
  return http.post('/option/orders', params).then((res) => res.data)
}

export function apiOptionCancelOrder(params: AppCancelOrderReq): Promise<RespBase> {
  return http.post('/option/orders/cancel', params).then((res) => res.data)
}

export function apiOptionGetOrderDetail(
  params: AppGetOrderDetailReq,
): Promise<RespBase & { data: OptionOrderDetail }> {
  return http
    .get('/option/orders/detail', { params: compactParams(params) })
    .then((res) => res.data)
}

export function apiOptionListCurrentOrders(
  params: AppListCurrentOrdersReq,
): Promise<RespBase & { data: OptionOrderDetail[] }> {
  return http
    .get('/option/orders/current', { params: compactParams(params) })
    .then((res) => res.data)
}

export function apiOptionListHistoryOrders(
  params: AppListHistoryOrdersReq,
): Promise<RespBase & { data: OptionOrderDetail[] }> {
  return http
    .get('/option/orders/history', { params: compactParams(params) })
    .then((res) => res.data)
}

export function apiOptionListTrades(
  params: AppListTradesReq,
): Promise<RespBase & { data: OptionTradeDetail[] }> {
  return http.get('/option/trades', { params: compactParams(params) }).then((res) => res.data)
}

export function apiOptionListPositions(
  params: AppListPositionsReq,
): Promise<RespBase & { data: OptionPositionDetail[] }> {
  return http.get('/option/positions', { params: compactParams(params) }).then((res) => res.data)
}

export function apiOptionGetPositionDetail(
  params: AppGetPositionDetailReq,
): Promise<RespBase & { data: OptionPositionDetail }> {
  return http
    .get('/option/positions/detail', { params: compactParams(params) })
    .then((res) => res.data)
}

export function apiOptionExercise(
  params: AppExerciseReq,
): Promise<RespBase & { data: { exerciseNo: string; exerciseId: number } }> {
  return http.post('/option/exercise', params).then((res) => res.data)
}

export function apiOptionListExercises(
  params: AppListExercisesReq,
): Promise<RespBase & { data: OptionExerciseDetail[] }> {
  return http.get('/option/exercises', { params: compactParams(params) }).then((res) => res.data)
}

export function apiOptionListAccounts(
  params: AppListAccountsReq,
): Promise<RespBase & { data: OptionAccount[] }> {
  return http.get('/option/accounts', { params: compactParams(params) }).then((res) => res.data)
}

export function apiOptionListBills(
  params: AppListBillsReq,
): Promise<RespBase & { data: OptionBill[] }> {
  return http.get('/option/bills', { params: compactParams(params) }).then((res) => res.data)
}
