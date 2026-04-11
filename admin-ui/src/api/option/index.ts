import { get, post } from '@/utils/request'
import type { OptionAdminCommonResp, OptionContract, OptionContractDetail, OptionMarket, OptionMarketSnapshot, RespBase } from '@/services'

export function apiOptionListContracts(params: Record<string, any>): Promise<RespBase<OptionContractDetail[]>> {
  return get<OptionContractDetail[]>('/admin/option/contracts', params)
}

export function apiOptionGetContract(params: Record<string, any>): Promise<RespBase<OptionContractDetail>> {
  return get<OptionContractDetail>('/admin/option/contracts/detail', params)
}

export function apiOptionCreateContract(params: Record<string, any>): Promise<RespBase<{ id: number }>> {
  return post<{ id: number }>('/admin/option/contracts', params)
}

export function apiOptionUpdateContract(params: Record<string, any>): Promise<OptionAdminCommonResp> {
  return post('/admin/option/contracts/update', params)
}

export function apiOptionGetMarket(params: Record<string, any>): Promise<RespBase<OptionMarket>> {
  return get<OptionMarket>('/admin/option/market/detail', params)
}

export function apiOptionUpdateMarket(params: Record<string, any>): Promise<OptionAdminCommonResp> {
  return post('/admin/option/market/update', params)
}

export function apiOptionListMarketSnapshots(params: Record<string, any>): Promise<RespBase<OptionMarketSnapshot[]>> {
  return get<OptionMarketSnapshot[]>('/admin/option/market/snapshots', params)
}

export function apiOptionListOrders(params: Record<string, any>): Promise<RespBase<any[]>> {
  return get<any[]>('/admin/option/orders', params)
}

export function apiOptionGetOrder(params: Record<string, any>): Promise<RespBase<any>> {
  return get<any>('/admin/option/orders/detail', params)
}

export function apiOptionListTrades(params: Record<string, any>): Promise<RespBase<any[]>> {
  return get<any[]>('/admin/option/trades', params)
}

export function apiOptionGetTrade(params: Record<string, any>): Promise<RespBase<any>> {
  return get<any>('/admin/option/trades/detail', params)
}

export function apiOptionListPositions(params: Record<string, any>): Promise<RespBase<any[]>> {
  return get<any[]>('/admin/option/positions', params)
}

export function apiOptionGetPosition(params: Record<string, any>): Promise<RespBase<any>> {
  return get<any>('/admin/option/positions/detail', params)
}

export function apiOptionListExercises(params: Record<string, any>): Promise<RespBase<any[]>> {
  return get<any[]>('/admin/option/exercises', params)
}

export function apiOptionGetExercise(params: Record<string, any>): Promise<RespBase<any>> {
  return get<any>('/admin/option/exercises/detail', params)
}

export function apiOptionListSettlements(params: Record<string, any>): Promise<RespBase<any[]>> {
  return get<any[]>('/admin/option/settlements', params)
}

export function apiOptionGetSettlement(params: Record<string, any>): Promise<RespBase<any>> {
  return get<any>('/admin/option/settlements/detail', params)
}

export function apiOptionListAccounts(params: Record<string, any>): Promise<RespBase<any[]>> {
  return get<any[]>('/admin/option/accounts', params)
}

export function apiOptionGetAccount(params: Record<string, any>): Promise<RespBase<any>> {
  return get<any>('/admin/option/accounts/detail', params)
}

export function apiOptionListBills(params: Record<string, any>): Promise<RespBase<any[]>> {
  return get<any[]>('/admin/option/bills', params)
}

export function apiOptionGetBill(params: Record<string, any>): Promise<RespBase<any>> {
  return get<any>('/admin/option/bills/detail', params)
}
