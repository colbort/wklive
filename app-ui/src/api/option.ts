import { http } from '@/api/http'
import { compactParams } from '@/api/utils'
import type { RespBase } from '@/types/api'
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

export function apiOptionListContracts(
  params: AppListContractsReq,
): Promise<RespBase & { list: OptionContractDetail[] }> {
  return http.get('/option/contracts', { params: compactParams(params) }).then((res) => res.data)
}

export function apiOptionGetContractDetail(
  params: AppGetContractDetailReq,
): Promise<RespBase & { data: OptionContractDetail }> {
  return http.get('/option/contracts/detail', { params: compactParams(params) }).then((res) => res.data)
}

export function apiOptionPlaceOrder(
  params: AppPlaceOrderReq,
): Promise<RespBase & { orderNo: string; orderId: number }> {
  return http.post('/option/orders', params).then((res) => res.data)
}

export function apiOptionCancelOrder(params: AppCancelOrderReq): Promise<RespBase> {
  return http.post('/option/orders/cancel', params).then((res) => res.data)
}

export function apiOptionGetOrderDetail(
  params: AppGetOrderDetailReq,
): Promise<RespBase & { data: OptionOrderDetail }> {
  return http.get('/option/orders/detail', { params: compactParams(params) }).then((res) => res.data)
}

export function apiOptionListCurrentOrders(
  params: AppListCurrentOrdersReq,
): Promise<RespBase & { list: OptionOrderDetail[] }> {
  return http.get('/option/orders/current', { params: compactParams(params) }).then((res) => res.data)
}

export function apiOptionListHistoryOrders(
  params: AppListHistoryOrdersReq,
): Promise<RespBase & { list: OptionOrderDetail[] }> {
  return http.get('/option/orders/history', { params: compactParams(params) }).then((res) => res.data)
}

export function apiOptionListTrades(
  params: AppListTradesReq,
): Promise<RespBase & { list: OptionTradeDetail[] }> {
  return http.get('/option/trades', { params: compactParams(params) }).then((res) => res.data)
}

export function apiOptionListPositions(
  params: AppListPositionsReq,
): Promise<RespBase & { list: OptionPositionDetail[] }> {
  return http.get('/option/positions', { params: compactParams(params) }).then((res) => res.data)
}

export function apiOptionGetPositionDetail(
  params: AppGetPositionDetailReq,
): Promise<RespBase & { data: OptionPositionDetail }> {
  return http.get('/option/positions/detail', { params: compactParams(params) }).then((res) => res.data)
}

export function apiOptionExercise(
  params: AppExerciseReq,
): Promise<RespBase & { exerciseNo: string; exerciseId: number }> {
  return http.post('/option/exercise', params).then((res) => res.data)
}

export function apiOptionListExercises(
  params: AppListExercisesReq,
): Promise<RespBase & { list: OptionExerciseDetail[] }> {
  return http.get('/option/exercises', { params: compactParams(params) }).then((res) => res.data)
}

export function apiOptionListAccounts(
  params: AppListAccountsReq,
): Promise<RespBase & { list: OptionAccount[] }> {
  return http.get('/option/accounts', { params: compactParams(params) }).then((res) => res.data)
}

export function apiOptionListBills(params: AppListBillsReq): Promise<RespBase & { list: OptionBill[] }> {
  return http.get('/option/bills', { params: compactParams(params) }).then((res) => res.data)
}
