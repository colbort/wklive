import { http } from '@/api/http'
import { compactParams } from '@/api/utils'
import type { RespBase } from '@/types/api'
import type {
  CancelAllOrdersReq,
  CancelOrderReq,
  ContractLeverageConfig,
  ContractMarginAccount,
  ContractPosition,
  GetFillListReq,
  GetLeverageConfigReq,
  GetMarginAccountListReq,
  GetOrderDetailReq,
  GetOrderListReq,
  GetPositionListReq,
  GetSymbolDetailReq,
  GetSymbolListReq,
  PlaceOrderReq,
  SetLeverageReq,
  TradeFill,
  TradeOrder,
  TradeSymbol,
  TradeSymbolContract,
  TradeSymbolSpot,
} from '@/types/trade'

export function apiTradeGetSymbolList(params: GetSymbolListReq): Promise<RespBase & { list: TradeSymbol[] }> {
  return http.get('/trade/symbols', { params: compactParams(params) }).then((res) => res.data)
}

export function apiTradeGetSymbolDetail(
  params: GetSymbolDetailReq,
): Promise<RespBase & { symbol: TradeSymbol; spot: TradeSymbolSpot; contract: TradeSymbolContract }> {
  return http.get('/trade/symbols/detail', { params: compactParams(params) }).then((res) => res.data)
}

export function apiTradePlaceOrder(params: PlaceOrderReq): Promise<RespBase & { order: TradeOrder }> {
  return http.post('/trade/orders', params).then((res) => res.data)
}

export function apiTradeCancelOrder(params: CancelOrderReq): Promise<RespBase> {
  return http.post('/trade/orders/cancel', params).then((res) => res.data)
}

export function apiTradeCancelAllOrders(
  params: CancelAllOrdersReq,
): Promise<RespBase & { affectedCount: number }> {
  return http.post('/trade/orders/cancel-all', params).then((res) => res.data)
}

export function apiTradeGetOrderList(params: GetOrderListReq): Promise<RespBase & { list: TradeOrder[] }> {
  return http.get('/trade/orders', { params: compactParams(params) }).then((res) => res.data)
}

export function apiTradeGetOrderDetail(
  params: GetOrderDetailReq,
): Promise<RespBase & { order: TradeOrder; spot: TradeSymbolSpot; contract: TradeSymbolContract }> {
  return http.get('/trade/orders/detail', { params: compactParams(params) }).then((res) => res.data)
}

export function apiTradeGetFillList(params: GetFillListReq): Promise<RespBase & { list: TradeFill[] }> {
  return http.get('/trade/fills', { params: compactParams(params) }).then((res) => res.data)
}

export function apiTradeGetPositionList(
  params: GetPositionListReq,
): Promise<RespBase & { list: ContractPosition[] }> {
  return http.get('/trade/positions', { params: compactParams(params) }).then((res) => res.data)
}

export function apiTradeGetMarginAccountList(
  params: GetMarginAccountListReq,
): Promise<RespBase & { list: ContractMarginAccount[] }> {
  return http.get('/trade/margin-accounts', { params: compactParams(params) }).then((res) => res.data)
}

export function apiTradeGetLeverageConfig(
  params: GetLeverageConfigReq,
): Promise<RespBase & { data: ContractLeverageConfig }> {
  return http.get('/trade/leverage-config', { params: compactParams(params) }).then((res) => res.data)
}

export function apiTradeSetLeverage(params: SetLeverageReq): Promise<RespBase> {
  return http.post('/trade/leverage', params).then((res) => res.data)
}
