import { get, post } from '@/utils/request'
import type {
  BizTradeEvent,
  ContractLeverageConfig,
  ContractMarginAccount,
  ContractPosition,
  ContractPositionHistory,
  CreateSymbolReq,
  GetCancelLogListAdminReq,
  GetFillDetailAdminReq,
  GetFillListAdminReq,
  GetMarginAccountListAdminReq,
  GetOrderDetailAdminReq,
  GetOrderListAdminReq,
  GetPositionDetailAdminReq,
  GetPositionHistoryListAdminReq,
  GetPositionListAdminReq,
  GetRiskOrderCheckLogListReq,
  GetSymbolDetailAdminReq,
  GetSymbolListAdminReq,
  GetTradeEventDetailReq,
  GetTradeEventListReq,
  GetUserLeverageConfigReq,
  GetUserSymbolLimitReq,
  GetUserTradeConfigReq,
  GetUserTradeLimitReq,
  OptionGroup,
  RespBase,
  RetryTradeEventReq,
  RiskOrderCheckLog,
  RiskUserSymbolLimit,
  RiskUserTradeLimit,
  SetContractSymbolConfigReq,
  SetSpotSymbolConfigReq,
  SetUserLeverageConfigReq,
  SetUserSymbolLimitReq,
  SetUserTradeConfigReq,
  SetUserTradeLimitReq,
  TradeCancelLog,
  TradeFill,
  TradeOrder,
  TradeSymbol,
  TradeUserConfig,
  UpdateSymbolReq,
} from '@/services'

export function apiTradeListSymbols(params: GetSymbolListAdminReq): Promise<RespBase<TradeSymbol[]>> {
  return get<TradeSymbol[]>('/admin/trade/symbols', params)
}

export function apiTradeGetSymbol(params: GetSymbolDetailAdminReq): Promise<RespBase<TradeSymbol>> {
  return get<TradeSymbol>('/admin/trade/symbols/detail', params)
}

export function apiTradeCreateSymbol(params: CreateSymbolReq): Promise<RespBase> {
  return post('/admin/trade/symbols', params)
}

export function apiTradeUpdateSymbol(params: UpdateSymbolReq): Promise<RespBase> {
  return post('/admin/trade/symbols/update', params)
}

export function apiTradeSetSpotConfig(params: SetSpotSymbolConfigReq): Promise<RespBase> {
  return post('/admin/trade/symbols/spot-config', params)
}

export function apiTradeSetContractConfig(params: SetContractSymbolConfigReq): Promise<RespBase> {
  return post('/admin/trade/symbols/contract-config', params)
}

export function apiTradeListOrders(params: GetOrderListAdminReq): Promise<RespBase<TradeOrder[]>> {
  return get<TradeOrder[]>('/admin/trade/orders', params)
}

export function apiTradeGetOrder(params: GetOrderDetailAdminReq): Promise<RespBase<TradeOrder>> {
  return get<TradeOrder>('/admin/trade/orders/detail', params)
}

export function apiTradeListFills(params: GetFillListAdminReq): Promise<RespBase<TradeFill[]>> {
  return get<TradeFill[]>('/admin/trade/fills', params)
}

export function apiTradeGetFill(params: GetFillDetailAdminReq): Promise<RespBase<TradeFill>> {
  return get<TradeFill>('/admin/trade/fills/detail', params)
}

export function apiTradeListPositions(params: GetPositionListAdminReq): Promise<RespBase<ContractPosition[]>> {
  return get<ContractPosition[]>('/admin/trade/positions', params)
}

export function apiTradeGetPosition(params: GetPositionDetailAdminReq): Promise<RespBase<ContractPosition>> {
  return get<ContractPosition>('/admin/trade/positions/detail', params)
}

export function apiTradeListPositionHistories(
  params: GetPositionHistoryListAdminReq,
): Promise<RespBase<ContractPositionHistory[]>> {
  return get<ContractPositionHistory[]>('/admin/trade/position-histories', params)
}

export function apiTradeListMarginAccounts(params: GetMarginAccountListAdminReq): Promise<RespBase<ContractMarginAccount[]>> {
  return get<ContractMarginAccount[]>('/admin/trade/margin-accounts', params)
}

export function apiTradeListCancelLogs(params: GetCancelLogListAdminReq): Promise<RespBase<TradeCancelLog[]>> {
  return get<TradeCancelLog[]>('/admin/trade/cancel-logs', params)
}

export function apiTradeGetUserTradeLimit(params: GetUserTradeLimitReq): Promise<RespBase<RiskUserTradeLimit>> {
  return get<RiskUserTradeLimit>('/admin/trade/user-trade-limit', params)
}

export function apiTradeSetUserTradeLimit(params: SetUserTradeLimitReq): Promise<RespBase> {
  return post('/admin/trade/user-trade-limit', params)
}

export function apiTradeGetUserSymbolLimit(params: GetUserSymbolLimitReq): Promise<RespBase<RiskUserSymbolLimit>> {
  return get<RiskUserSymbolLimit>('/admin/trade/user-symbol-limit', params)
}

export function apiTradeSetUserSymbolLimit(params: SetUserSymbolLimitReq): Promise<RespBase> {
  return post('/admin/trade/user-symbol-limit', params)
}

export function apiTradeGetUserTradeConfig(params: GetUserTradeConfigReq): Promise<RespBase<TradeUserConfig>> {
  return get<TradeUserConfig>('/admin/trade/user-trade-config', params)
}

export function apiTradeSetUserTradeConfig(params: SetUserTradeConfigReq): Promise<RespBase> {
  return post('/admin/trade/user-trade-config', params)
}

export function apiTradeListRiskLogs(params: GetRiskOrderCheckLogListReq): Promise<RespBase<RiskOrderCheckLog[]>> {
  return get<RiskOrderCheckLog[]>('/admin/trade/risk-order-check-logs', params)
}

export function apiTradeGetUserLeverageConfig(params: GetUserLeverageConfigReq): Promise<RespBase<ContractLeverageConfig>> {
  return get<ContractLeverageConfig>('/admin/trade/user-leverage-config', params)
}

export function apiTradeSetUserLeverageConfig(params: SetUserLeverageConfigReq): Promise<RespBase> {
  return post('/admin/trade/user-leverage-config', params)
}

export function apiTradeListEvents(params: GetTradeEventListReq): Promise<RespBase<BizTradeEvent[]>> {
  return get<BizTradeEvent[]>('/admin/trade/events', params)
}

export function apiTradeGetEvent(params: GetTradeEventDetailReq): Promise<RespBase<BizTradeEvent>> {
  return get<BizTradeEvent>('/admin/trade/events/detail', params)
}

export function apiTradeRetryEvent(params: RetryTradeEventReq): Promise<RespBase> {
  return post('/admin/trade/events/retry', params)
}

export function apiOptions(): Promise<RespBase<OptionGroup[]>> {
  return get('/admin/trade/options')
}
