import { authHttp, http } from './http'
import { compactParams } from './utils'
import type { OptionsGroup, RespBase } from '../types/api'
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
  TradeSymbolLeverageConfig,
  TradeSymbolSpot,
} from '../types/trade'

export function apiGetTradeOptions(): Promise<
  RespBase & { data: OptionsGroup[] }
> {
  return http.get('/trade/options').then((res: { data: any }) => res.data)
}

export function apiTradeGetSymbolList(
  params: GetSymbolListReq,
): Promise<RespBase & { data: TradeSymbol[] }> {
  return http
    .get('/trade/symbols', { params: compactParams(params) })
    .then((res: { data: any }) => res.data)
}

export function apiTradeGetSymbolDetail(params: GetSymbolDetailReq): Promise<
  RespBase & {
    data: {
      symbol: TradeSymbol
      spot: TradeSymbolSpot
      contract: TradeSymbolContract
      leverageConfigs: TradeSymbolLeverageConfig[]
    }
  }
> {
  return http
    .get('/trade/symbols/detail', { params: compactParams(params) })
    .then((res: { data: any }) => res.data)
}

export function apiTradePlaceOrder(
  params: PlaceOrderReq,
): Promise<RespBase & { data: TradeOrder }> {
  return authHttp
    .post('/trade/orders', params)
    .then((res: { data: any }) => res.data)
}

export function apiTradeCancelOrder(params: CancelOrderReq): Promise<RespBase> {
  return authHttp
    .post('/trade/orders/cancel', params)
    .then((res: { data: any }) => res.data)
}

export function apiTradeCancelAllOrders(
  params: CancelAllOrdersReq,
): Promise<RespBase & { data: number }> {
  return authHttp
    .post('/trade/orders/cancel-all', params)
    .then((res: { data: any }) => res.data)
}

export function apiTradeGetOrderList(
  params: GetOrderListReq,
): Promise<RespBase & { data: TradeOrder[] }> {
  return authHttp
    .get('/trade/orders', { params: compactParams(params) })
    .then((res: { data: any }) => res.data)
}

export function apiTradeGetOrderDetail(params: GetOrderDetailReq): Promise<
  RespBase & {
    data: {
      order: TradeOrder
      spot: TradeSymbolSpot
      contract: TradeSymbolContract
    }
  }
> {
  return authHttp
    .get('/trade/orders/detail', { params: compactParams(params) })
    .then((res: { data: any }) => res.data)
}

export function apiTradeGetFillList(
  params: GetFillListReq,
): Promise<RespBase & { data: TradeFill[] }> {
  return authHttp
    .get('/trade/fills', { params: compactParams(params) })
    .then((res: { data: any }) => res.data)
}

export function apiTradeGetPositionList(
  params: GetPositionListReq,
): Promise<RespBase & { data: ContractPosition[] }> {
  return authHttp
    .get('/trade/positions', { params: compactParams(params) })
    .then((res: { data: any }) => res.data)
}

export function apiTradeGetMarginAccountList(
  params: GetMarginAccountListReq,
): Promise<RespBase & { data: ContractMarginAccount[] }> {
  return authHttp
    .get('/trade/margin-accounts', { params: compactParams(params) })
    .then((res: { data: any }) => res.data)
}

export function apiTradeGetLeverageConfig(
  params: GetLeverageConfigReq,
): Promise<RespBase & { data: ContractLeverageConfig }> {
  return authHttp
    .get('/trade/leverage-config', { params: compactParams(params) })
    .then((res: { data: any }) => res.data)
}

export function apiTradeSetLeverage(params: SetLeverageReq): Promise<RespBase> {
  return authHttp
    .post('/trade/leverage', params)
    .then((res: { data: any }) => res.data)
}
