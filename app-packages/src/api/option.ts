import { authHttp, http } from './http'
import { compactParams } from './utils'
import type { RespBase } from '../types/api'
import type {
  OptionCancelOrderReq,
  ExerciseReq,
  GetContractDetailReq,
  OptionGetOrderDetailReq,
  GetPositionDetailReq,
  ListAccountsReq,
  ListBillsReq,
  ListContractsReq,
  ListCurrentOrdersReq,
  ListExercisesReq,
  ListHistoryOrdersReq,
  ListPositionsReq,
  ListTradesReq,
  OptionPlaceOrderReq,
  OptionAccount,
  OptionBill,
  OptionContractDetail,
  OptionExerciseDetail,
  OptionOrderDetail,
  OptionPositionDetail,
  OptionTradeDetail,
} from '../types/option'

export function apiOptionListContracts(
  params: ListContractsReq,
): Promise<RespBase & { data: OptionContractDetail[] }> {
  return http
    .get('/option/contracts', { params: compactParams(params) })
    .then((res: { data: any }) => res.data)
}

export function apiOptionGetContractDetail(
  params: GetContractDetailReq,
): Promise<RespBase & { data: OptionContractDetail }> {
  return http
    .get('/option/contracts/detail', { params: compactParams(params) })
    .then((res: { data: any }) => res.data)
}

export function apiOptionPlaceOrder(
  params: OptionPlaceOrderReq,
): Promise<RespBase & { data: { orderNo: string; orderId: number } }> {
  return authHttp
    .post('/option/orders', params)
    .then((res: { data: any }) => res.data)
}

export function apiOptionCancelOrder(
  params: OptionCancelOrderReq,
): Promise<RespBase> {
  return authHttp
    .post('/option/orders/cancel', params)
    .then((res: { data: any }) => res.data)
}

export function apiOptionGetOrderDetail(
  params: OptionGetOrderDetailReq,
): Promise<RespBase & { data: OptionOrderDetail }> {
  return authHttp
    .get('/option/orders/detail', { params: compactParams(params) })
    .then((res: { data: any }) => res.data)
}

export function apiOptionListCurrentOrders(
  params: ListCurrentOrdersReq,
): Promise<RespBase & { data: OptionOrderDetail[] }> {
  return authHttp
    .get('/option/orders/current', { params: compactParams(params) })
    .then((res: { data: any }) => res.data)
}

export function apiOptionListHistoryOrders(
  params: ListHistoryOrdersReq,
): Promise<RespBase & { data: OptionOrderDetail[] }> {
  return authHttp
    .get('/option/orders/history', { params: compactParams(params) })
    .then((res: { data: any }) => res.data)
}

export function apiOptionListTrades(
  params: ListTradesReq,
): Promise<RespBase & { data: OptionTradeDetail[] }> {
  return authHttp
    .get('/option/trades', { params: compactParams(params) })
    .then((res: { data: any }) => res.data)
}

export function apiOptionListPositions(
  params: ListPositionsReq,
): Promise<RespBase & { data: OptionPositionDetail[] }> {
  return authHttp
    .get('/option/positions', { params: compactParams(params) })
    .then((res: { data: any }) => res.data)
}

export function apiOptionGetPositionDetail(
  params: GetPositionDetailReq,
): Promise<RespBase & { data: OptionPositionDetail }> {
  return authHttp
    .get('/option/positions/detail', { params: compactParams(params) })
    .then((res: { data: any }) => res.data)
}

export function apiOptionExercise(
  params: ExerciseReq,
): Promise<RespBase & { data: { exerciseNo: string; exerciseId: number } }> {
  return authHttp
    .post('/option/exercise', params)
    .then((res: { data: any }) => res.data)
}

export function apiOptionListExercises(
  params: ListExercisesReq,
): Promise<RespBase & { data: OptionExerciseDetail[] }> {
  return authHttp
    .get('/option/exercises', { params: compactParams(params) })
    .then((res: { data: any }) => res.data)
}

export function apiOptionListAccounts(
  params: ListAccountsReq,
): Promise<RespBase & { data: OptionAccount[] }> {
  return authHttp
    .get('/option/accounts', { params: compactParams(params) })
    .then((res: { data: any }) => res.data)
}

export function apiOptionListBills(
  params: ListBillsReq,
): Promise<RespBase & { data: OptionBill[] }> {
  return authHttp
    .get('/option/bills', { params: compactParams(params) })
    .then((res: { data: any }) => res.data)
}
