import type { PageReq, TimeRange } from './api'

export interface TradeSymbol {
  id: number
  tenantId: number
  categoryType: number
  market: string
  symbol: string
  displaySymbol: string
  productType: number
  baseAsset: string
  quoteAsset: string
  settleAsset: string
  contractType: number
  contractValueType: number
  marginAsset: string
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
  maxNotional: string
  listingTime: number
  tradingStartTime: number
  tradingEndTime: number
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

export interface TradeSymbolSeconds {
  id: number
  tenantId: number
  symbolId: number
  durationSeconds: number
  payoutRate: string
  drawRule: number
  startPriceSource: string
  settlementPriceSource: string
  quoteValidityMs: number
  minStake: string
  maxStake: string
  upEnabled: number
  downEnabled: number
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
  fundingRateCap: string
  fundingRateFloor: string
  indexSymbol: string
  markPriceSource: string
  settlementPriceSource: string
  openLongEnabled: number
  openShortEnabled: number
  closeLongEnabled: number
  closeShortEnabled: number
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
  productType: number
  contractType: number
  contractValueType: number
  side: number
  positionSide: number
  orderType: number
  timeInForce: number
  status: number
  displayStatus: number
  price: string
  qty: string
  amount: string
  filledQty: string
  filledAmount: string
  avgPrice: string
  fee: string
  feeAsset: string
  source: number
  isReduceOnly: number // 是否只减仓：1是 2否
  triggerPrice: string
  triggerType: number
  triggerKind: number
  cancelReason: string
  bizExt: string
  createTimes: number
  updateTimes: number
  secondsDirection: number
  durationSeconds: number
}

export interface TradeOrderSeconds {
  id: number
  tenantId: number
  orderId: number
  direction: number
  durationSeconds: number
  stakeAsset: string
  stakeAmount: string
  payoutRate: string
  startPrice: string
  startPriceTime: number
  expireTime: number
  settlementPrice: string
  settlementPriceTime: number
  result: number
  feeRate: string
  frozenAt: number
  activatedAt: number
  startPriceSource: string
  settlementPriceSource: string
  priceAlgorithm: string
  profitAmount: string
  feeAmount: string
  returnAmount: string
  settlementStatus: number
  reservationNo: string
  settlementReason: string
  settledAt: number
  version: number
  createTimes: number
  updateTimes: number
}

export interface TradeOrderSpot {
  id: number
  tenantId: number
  orderId: number
  frozenAsset: string
  frozenAmount: string
  settleAsset: string
  settleAmount: string
  createTimes: number
  updateTimes: number
}

export interface TradeOrderContract {
  id: number
  tenantId: number
  orderId: number
  marginMode: number
  leverage: number
  marginAsset: string
  marginAmount: string
  closePositionType: number
  liquidationPrice: string
  takeProfitPrice: string
  stopLossPrice: string
  reservedCloseQty: string
  riskPrice: string
  riskTierId: number
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
  productType: number
  contractType: number
  contractValueType: number
  matchNo: string
  settlementStatus: number
  settlementRetryCount: number
  settledAt: number
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
  contractType: number
  contractValueType: number
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

export interface ContractMarginSnapshot {
  id: number
  tenantId: number
  userId: number
  marginAsset: string
  walletBalance: string
  availableBalance: string
  frozenBalance: string
  positionMargin: string
  orderMargin: string
  unrealizedPnl: string
  realizedPnl: string
  version: number
  sourceEventNo: string
  snapshotTime: number
  createTimes: number
  updateTimes: number
}

export interface ContractLeverageConfig {
  id: number
  tenantId: number
  userId: number
  symbolId: number
  marginMode: number
  longLeverage: number
  shortLeverage: number
  operatorId: number
  source: number
  enabled: number
  remark: string
  createTimes: number
  updateTimes: number
}

export interface TradeSymbolLeverageConfig {
  id: number
  tenantId: number
  symbolId: number
  marginMode: number
  leverageValues: number[]
  defaultLeverage: number
  enabled: number
  sort: number
  remark: string
  createTimes: number
  updateTimes: number
}

export interface GetSymbolListReq {
  productType?: number
  status?: number
  categoryType?: number
  market?: string
}

export interface GetSymbolDetailReq {
  symbolId: number
}

export interface TradePlaceOrderReq {
  symbolId: number
  side: number
  positionSide: number
  orderType: number
  timeInForce: number
  clientOrderId?: string
  price?: string
  qty?: string
  amount?: string
  isReduceOnly?: number // 是否只减仓：1是 2否
  triggerPrice?: string
  triggerType?: number
  triggerKind?: number
  marginMode?: number
  leverage?: number
  takeProfitPrice?: string
  stopLossPrice?: string
  orderSource?: number
  secondsDirection?: number
  durationSeconds?: number
}

export interface TradeCancelOrderReq {
  orderId?: number
  orderNo?: string
  clientOrderId?: string
}

export interface CancelAllOrdersReq {
  productType?: number
  symbolId?: number
  side?: number
  positionSide?: number
}

export interface GetOrderListReq extends PageReq {
  productType?: number
  symbolId?: number
  status?: number
  side?: number
  timeRange?: TimeRange
}

export interface TradeGetOrderDetailReq {
  orderId?: number
  orderNo?: string
}

export interface GetFillListReq extends PageReq {
  productType?: number
  symbolId?: number
  timeRange?: TimeRange
}

export interface GetPositionListReq {
  contractType?: number
  symbolId?: number
}

export interface GetMarginSnapshotListReq {
  marginAsset?: string
}

export interface GetLeverageConfigReq {
  symbolId: number
  marginMode: number
}

export interface SetLeverageReq {
  symbolId: number
  marginMode: number
  longLeverage: number
  shortLeverage: number
}
