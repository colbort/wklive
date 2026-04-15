import { get, post } from '@/utils/request'
import type { OptionGroup, RespBase, TradeOrder, TradeSymbol } from '@/services'

export function apiTradeListSymbols(params: Record<string, any>): Promise<RespBase<TradeSymbol[]>> {
  return get<TradeSymbol[]>('/admin/trade/symbols', params)
}

export function apiTradeGetSymbol(params: Record<string, any>): Promise<RespBase<TradeSymbol>> {
  return get<TradeSymbol>('/admin/trade/symbols/detail', params)
}

export function apiTradeCreateSymbol(params: Record<string, any>): Promise<RespBase> {
  return post('/admin/trade/symbols', params)
}

export function apiTradeUpdateSymbol(params: Record<string, any>): Promise<RespBase> {
  return post('/admin/trade/symbols/update', params)
}

export function apiTradeSetSpotConfig(params: Record<string, any>): Promise<RespBase> {
  return post('/admin/trade/symbols/spot-config', params)
}

export function apiTradeSetContractConfig(params: Record<string, any>): Promise<RespBase> {
  return post('/admin/trade/symbols/contract-config', params)
}

export function apiTradeListOrders(params: Record<string, any>): Promise<RespBase<TradeOrder[]>> {
  return get<TradeOrder[]>('/admin/trade/orders', params)
}

export function apiTradeGetOrder(params: Record<string, any>): Promise<RespBase<any>> {
  return get<any>('/admin/trade/orders/detail', params)
}

export function apiTradeListFills(params: Record<string, any>): Promise<RespBase<any[]>> {
  return get<any[]>('/admin/trade/fills', params)
}

export function apiTradeGetFill(params: Record<string, any>): Promise<RespBase<any>> {
  return get<any>('/admin/trade/fills/detail', params)
}

export function apiTradeListPositions(params: Record<string, any>): Promise<RespBase<any[]>> {
  return get<any[]>('/admin/trade/positions', params)
}

export function apiTradeGetPosition(params: Record<string, any>): Promise<RespBase<any>> {
  return get<any>('/admin/trade/positions/detail', params)
}

export function apiTradeListPositionHistories(params: Record<string, any>): Promise<RespBase<any[]>> {
  return get<any[]>('/admin/trade/position-histories', params)
}

export function apiTradeListMarginAccounts(params: Record<string, any>): Promise<RespBase<any[]>> {
  return get<any[]>('/admin/trade/margin-accounts', params)
}

export function apiTradeListCancelLogs(params: Record<string, any>): Promise<RespBase<any[]>> {
  return get<any[]>('/admin/trade/cancel-logs', params)
}

export function apiTradeGetUserTradeLimit(params: Record<string, any>): Promise<RespBase<any>> {
  return get<any>('/admin/trade/user-trade-limit', params)
}

export function apiTradeSetUserTradeLimit(params: Record<string, any>): Promise<RespBase> {
  return post('/admin/trade/user-trade-limit', params)
}

export function apiTradeGetUserSymbolLimit(params: Record<string, any>): Promise<RespBase<any>> {
  return get<any>('/admin/trade/user-symbol-limit', params)
}

export function apiTradeSetUserSymbolLimit(params: Record<string, any>): Promise<RespBase> {
  return post('/admin/trade/user-symbol-limit', params)
}

export function apiTradeGetUserTradeConfig(params: Record<string, any>): Promise<RespBase<any>> {
  return get<any>('/admin/trade/user-trade-config', params)
}

export function apiTradeSetUserTradeConfig(params: Record<string, any>): Promise<RespBase> {
  return post('/admin/trade/user-trade-config', params)
}

export function apiTradeListRiskLogs(params: Record<string, any>): Promise<RespBase<any[]>> {
  return get<any[]>('/admin/trade/risk-order-check-logs', params)
}

export function apiTradeGetUserLeverageConfig(params: Record<string, any>): Promise<RespBase<any>> {
  return get<any>('/admin/trade/user-leverage-config', params)
}

export function apiTradeSetUserLeverageConfig(params: Record<string, any>): Promise<RespBase> {
  return post('/admin/trade/user-leverage-config', params)
}

export function apiTradeListEvents(params: Record<string, any>): Promise<RespBase<any[]>> {
  return get<any[]>('/admin/trade/events', params)
}

export function apiTradeGetEvent(params: Record<string, any>): Promise<RespBase<any>> {
  return get<any>('/admin/trade/events/detail', params)
}

export function apiTradeRetryEvent(params: Record<string, any>): Promise<RespBase> {
  return post('/admin/trade/events/retry', params)
}

export function apiOptions(): Promise<RespBase<OptionGroup[]>> {
  return get('/admin/trade/options')
}
