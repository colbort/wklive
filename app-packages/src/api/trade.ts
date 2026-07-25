import { authHttp, http } from './http'
import { compactParams } from './utils'
import type { RespBase } from '../types/api'
import type {
  CancelAllOrdersReq,
  TradeCancelOrderReq,
  ContractLeverageConfig,
  ContractMarginSnapshot,
  ContractPosition,
  GetFillListReq,
  GetLeverageConfigReq,
  GetMarginSnapshotListReq,
  TradeGetOrderDetailReq,
  GetOrderListReq,
  GetPositionListReq,
  GetSymbolDetailReq,
  GetSymbolListReq,
  TradePlaceOrderReq,
  SetLeverageReq,
  TradeFill,
  TradeOrder,
  TradeOrderContract,
  TradeOrderSpot,
  TradeSymbol,
  TradeSymbolContract,
  TradeSymbolLeverageConfig,
  TradeSymbolSeconds,
  TradeSymbolSpot,
  TradeOrderSeconds,
} from '../types/trade'

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
      secondsConfigs: TradeSymbolSeconds[]
    }
  }
> {
  return http
    .get('/trade/symbols/detail', { params: compactParams(params) })
    .then((res: { data: any }) => res.data)
}

export function apiTradePlaceOrder(
  params: TradePlaceOrderReq,
): Promise<RespBase & { data: TradeOrder }> {
  return authHttp
    .post('/trade/orders', params)
    .then((res: { data: any }) => res.data)
}

export function apiTradeCancelOrder(params: TradeCancelOrderReq): Promise<RespBase> {
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

export function apiTradeGetOrderDetail(params: TradeGetOrderDetailReq): Promise<
  RespBase & {
    data: {
      order: TradeOrder
      spot: TradeOrderSpot
      contract: TradeOrderContract
      seconds: TradeOrderSeconds
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

export function apiTradeGetMarginSnapshotList(
  params: GetMarginSnapshotListReq,
): Promise<RespBase & { data: ContractMarginSnapshot[] }> {
  return authHttp
    .get('/trade/margin-snapshots', { params: compactParams(params) })
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
