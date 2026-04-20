import type { PageReq, TimeRange } from '@/types/api'

export interface TradeSymbol {
  id: number
  tenantId: number
  symbol: string
  displaySymbol: string
  marketType: number
  baseAsset: string
  quoteAsset: string
  settleAsset: string
  contractType: number
  status: number
  priceScale: number
  qtyScale: number
  minPrice: string
  maxPrice: string
  priceTick: string
  minQty: string
  maxQty: string
  qtyStep: string
  minNotional: string
  maxLeverage: number
  openTime: number
  closeTime: number
  sort: number
  remark: string
  createTimes: number
  updateTimes: number
}

export interface TradeSymbolSpot {
  id: number
  tenantId: number
  symbolId: number
  makerFeeRate: string
  takerFeeRate: string
  buyEnabled: number
  sellEnabled: number
  createTimes: number
  updateTimes: number
}

export interface TradeSymbolContract {
  id: number
  tenantId: number
  symbolId: number
  contractSize: string
  multiplier: string
  maintenanceMarginRate: string
  initialMarginRate: string
  makerFeeRate: string
  takerFeeRate: string
  fundingIntervalMinutes: number
  deliveryTime: number
  supportCross: number
  supportIsolated: number
  buyEnabled: number
  sellEnabled: number
  createTimes: number
  updateTimes: number
}

export interface TradeOrder {
  id: number
  tenantId: number
  orderNo: string
  clientOrderId: string
  userId: number
  symbolId: number
  marketType: number
  side: number
  positionSide: number
  orderType: number
  timeInForce: number
  status: number
  price: string
  qty: string
  amount: string
  filledQty: string
  filledAmount: string
  avgPrice: string
  fee: string
  feeAsset: string
  source: number
  isReduceOnly: number
  isCloseOnly: number
  triggerPrice: string
  triggerType: number
  cancelReason: string
  bizExt: string
  createTimes: number
  updateTimes: number
}

export interface TradeFill {
  id: number
  tenantId: number
  fillNo: string
  orderId: number
  orderNo: string
  userId: number
  symbolId: number
  marketType: number
  side: number
  positionSide: number
  price: string
  qty: string
  amount: string
  fee: string
  feeAsset: string
  liquidityType: number
  realizedPnl: string
  matchTime: number
  createTimes: number
}

export interface ContractPosition {
  id: number
  tenantId: number
  userId: number
  symbolId: number
  marketType: number
  positionSide: number
  marginMode: number
  leverage: number
  qty: string
  availQty: string
  frozenQty: string
  openAvgPrice: string
  markPrice: string
  marginAsset: string
  positionMargin: string
  isolatedMargin: string
  unrealizedPnl: string
  realizedPnl: string
  liquidationPrice: string
  adlRank: number
  version: number
  createTimes: number
  updateTimes: number
}

export interface ContractMarginAccount {
  id: number
  tenantId: number
  userId: number
  marketType: number
  marginAsset: string
  balance: string
  availableBalance: string
  frozenBalance: string
  positionMargin: string
  orderMargin: string
  unrealizedPnl: string
  realizedPnl: string
  version: number
  createTimes: number
  updateTimes: number
}

export interface ContractLeverageConfig {
  id: number
  tenantId: number
  userId: number
  symbolId: number
  marketType: number
  marginMode: number
  positionMode: number
  longLeverage: number
  shortLeverage: number
  maxLeverage: number
  operatorId: number
  source: number
  status: number
  remark: string
  createTimes: number
  updateTimes: number
}

export interface GetSymbolListReq {
  tenantId?: number
  marketType?: number
  status?: number
}

export interface GetSymbolDetailReq {
  tenantId?: number
  symbolId: number
}

export interface PlaceOrderReq {
  tenantId?: number
  symbolId: number
  marketType: number
  side: number
  positionSide: number
  orderType: number
  timeInForce: number
  clientOrderId?: string
  price?: string
  qty?: string
  amount?: string
  isReduceOnly?: number
  isCloseOnly?: number
  triggerPrice?: string
  triggerType?: number
  marginMode?: number
  leverage?: number
  takeProfitPrice?: string
  stopLossPrice?: string
  orderSource?: number
}

export interface CancelOrderReq {
  tenantId?: number
  orderId?: number
  orderNo?: string
  clientOrderId?: string
}

export interface CancelAllOrdersReq {
  tenantId?: number
  marketType?: number
  symbolId?: number
  side?: number
  positionSide?: number
}

export interface GetOrderListReq extends PageReq {
  tenantId?: number
  marketType?: number
  symbolId?: number
  status?: number
  side?: number
  timeRange?: TimeRange
}

export interface GetOrderDetailReq {
  tenantId?: number
  orderId?: number
  orderNo?: string
}

export interface GetFillListReq extends PageReq {
  tenantId?: number
  marketType?: number
  symbolId?: number
  timeRange?: TimeRange
}

export interface GetPositionListReq {
  tenantId?: number
  marketType?: number
  symbolId?: number
}

export interface GetMarginAccountListReq {
  tenantId?: number
  marketType?: number
  marginAsset?: string
}

export interface GetLeverageConfigReq {
  tenantId?: number
  symbolId: number
  marketType: number
  marginMode: number
}

export interface SetLeverageReq {
  tenantId?: number
  symbolId: number
  marketType: number
  marginMode: number
  positionMode: number
  longLeverage: number
  shortLeverage: number
}
