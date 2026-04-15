import { get, post } from '@/utils/request'
import type {
  CreateContractReq,
  GetAccountReq,
  GetBillReq,
  GetContractReq,
  GetExerciseReq,
  GetMarketReq,
  GetOrderReq,
  GetPositionReq,
  GetSettlementReq,
  GetTradeReq,
  ListAccountsReq,
  ListBillsReq,
  ListContractsReq,
  ListExercisesReq,
  ListMarketSnapshotsReq,
  ListOrdersReq,
  ListPositionsReq,
  ListSettlementsReq,
  ListTradesReq,
  OptionAccount,
  OptionAdminCommonResp,
  OptionBill,
  OptionContractDetail,
  OptionExerciseDetail,
  OptionGroup,
  OptionMarket,
  OptionMarketSnapshot,
  OptionOrder,
  OptionOrderDetail,
  OptionPosition,
  OptionPositionDetail,
  OptionSettlement,
  OptionSettlementDetail,
  OptionTrade,
  OptionTradeDetail,
  RespBase,
  UpdateContractReq,
  UpdateMarketReq,
} from '@/services'

export function apiOptionListContracts(
  params: ListContractsReq,
): Promise<RespBase<OptionContractDetail[]>> {
  return get<OptionContractDetail[]>('/admin/option/contracts', params)
}

export function apiOptionGetContract(
  params: GetContractReq,
): Promise<RespBase<OptionContractDetail>> {
  return get<OptionContractDetail>('/admin/option/contracts/detail', params)
}

export function apiOptionCreateContract(
  params: CreateContractReq,
): Promise<RespBase<{ id: number }>> {
  return post<{ id: number }>('/admin/option/contracts', params)
}

export function apiOptionUpdateContract(
  params: UpdateContractReq,
): Promise<OptionAdminCommonResp> {
  return post('/admin/option/contracts/update', params)
}

export function apiOptionGetMarket(params: GetMarketReq): Promise<RespBase<OptionMarket>> {
  return get<OptionMarket>('/admin/option/market/detail', params)
}

export function apiOptionUpdateMarket(params: UpdateMarketReq): Promise<OptionAdminCommonResp> {
  return post('/admin/option/market/update', params)
}

export function apiOptionListMarketSnapshots(
  params: ListMarketSnapshotsReq,
): Promise<RespBase<OptionMarketSnapshot[]>> {
  return get<OptionMarketSnapshot[]>('/admin/option/market/snapshots', params)
}

export function apiOptionListOrders(params: ListOrdersReq): Promise<RespBase<OptionOrder[]>> {
  return get<OptionOrder[]>('/admin/option/orders', params)
}

export function apiOptionGetOrder(params: GetOrderReq): Promise<RespBase<OptionOrderDetail>> {
  return get<OptionOrderDetail>('/admin/option/orders/detail', params)
}

export function apiOptionListTrades(params: ListTradesReq): Promise<RespBase<OptionTrade[]>> {
  return get<OptionTrade[]>('/admin/option/trades', params)
}

export function apiOptionGetTrade(params: GetTradeReq): Promise<RespBase<OptionTradeDetail>> {
  return get<OptionTradeDetail>('/admin/option/trades/detail', params)
}

export function apiOptionListPositions(params: ListPositionsReq): Promise<RespBase<OptionPosition[]>> {
  return get<OptionPosition[]>('/admin/option/positions', params)
}

export function apiOptionGetPosition(params: GetPositionReq): Promise<RespBase<OptionPositionDetail>> {
  return get<OptionPositionDetail>('/admin/option/positions/detail', params)
}

export function apiOptionListExercises(params: ListExercisesReq): Promise<RespBase<OptionExerciseDetail[]>> {
  return get<OptionExerciseDetail[]>('/admin/option/exercises', params)
}

export function apiOptionGetExercise(params: GetExerciseReq): Promise<RespBase<OptionExerciseDetail>> {
  return get<OptionExerciseDetail>('/admin/option/exercises/detail', params)
}

export function apiOptionListSettlements(params: ListSettlementsReq): Promise<RespBase<OptionSettlement[]>> {
  return get<OptionSettlement[]>('/admin/option/settlements', params)
}

export function apiOptionGetSettlement(params: GetSettlementReq): Promise<RespBase<OptionSettlementDetail>> {
  return get<OptionSettlementDetail>('/admin/option/settlements/detail', params)
}

export function apiOptionListAccounts(params: ListAccountsReq): Promise<RespBase<OptionAccount[]>> {
  return get<OptionAccount[]>('/admin/option/accounts', params)
}

export function apiOptionGetAccount(params: GetAccountReq): Promise<RespBase<OptionAccount>> {
  return get<OptionAccount>('/admin/option/accounts/detail', params)
}

export function apiOptionListBills(params: ListBillsReq): Promise<RespBase<OptionBill[]>> {
  return get<OptionBill[]>('/admin/option/bills', params)
}

export function apiOptionGetBill(params: GetBillReq): Promise<RespBase<OptionBill>> {
  return get<OptionBill>('/admin/option/bills/detail', params)
}

export function apiOptions(): Promise<RespBase<OptionGroup[]>> {
  return get('/admin/option/options')
}
