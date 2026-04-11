import {
  apiTradeCreateSymbol,
  apiTradeGetEvent,
  apiTradeGetFill,
  apiTradeGetOrder,
  apiTradeGetPosition,
  apiTradeGetSymbol,
  apiTradeGetUserLeverageConfig,
  apiTradeGetUserSymbolLimit,
  apiTradeGetUserTradeConfig,
  apiTradeGetUserTradeLimit,
  apiTradeListCancelLogs,
  apiTradeListEvents,
  apiTradeListFills,
  apiTradeListMarginAccounts,
  apiTradeListOrders,
  apiTradeListPositionHistories,
  apiTradeListPositions,
  apiTradeListRiskLogs,
  apiTradeListSymbols,
  apiTradeRetryEvent,
  apiTradeSetContractConfig,
  apiTradeSetSpotConfig,
  apiTradeSetUserLeverageConfig,
  apiTradeSetUserSymbolLimit,
  apiTradeSetUserTradeConfig,
  apiTradeSetUserTradeLimit,
  apiTradeUpdateSymbol,
} from '@/api/trade'

export type TimeRange = {
  startTime?: number
  endTime?: number
}

export type TradeSymbol = {
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

export type TradeSymbolSpot = {
  id: number
  tenantId: number
  symbolId: number
  makerFeeRate: string
  takerFeeRate: string
  buyEnabled: boolean
  sellEnabled: boolean
  createTimes: number
  updateTimes: number
}

export type TradeSymbolContract = {
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
  supportCross: boolean
  supportIsolated: boolean
  buyEnabled: boolean
  sellEnabled: boolean
  createTimes: number
  updateTimes: number
}

export type TradeUserConfig = {
  id: number
  tenantId: number
  userId: number
  marketType: number
  symbolId: number
  positionMode: number
  marginMode: number
  defaultLeverage: number
  tradeEnabled: boolean
  reduceOnlyEnabled: boolean
  createTimes: number
  updateTimes: number
}

export type TradeOrder = {
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
  isReduceOnly: boolean
  isCloseOnly: boolean
  triggerPrice: string
  triggerType: number
  cancelReason: string
  bizExt: string
  createTimes: number
  updateTimes: number
}

export type TradeFill = {
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

export type TradeCancelLog = {
  id: number
  tenantId: number
  orderId: number
  orderNo: string
  userId: number
  cancelSource: number
  cancelReason: string
  createTimes: number
}

export type ContractPosition = {
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

export type ContractPositionHistory = {
  id: number
  tenantId: number
  positionId: number
  userId: number
  symbolId: number
  marketType: number
  positionSide: number
  actionType: number
  beforeQty: string
  afterQty: string
  beforeAvailQty: string
  afterAvailQty: string
  beforeFrozenQty: string
  afterFrozenQty: string
  beforeOpenAvgPrice: string
  afterOpenAvgPrice: string
  beforePositionMargin: string
  afterPositionMargin: string
  beforeIsolatedMargin: string
  afterIsolatedMargin: string
  beforeUnrealizedPnl: string
  afterUnrealizedPnl: string
  realizedPnlDelta: string
  feeDelta: string
  feeAsset: string
  markPrice: string
  refOrderId: number
  refFillId: number
  operatorId: number
  source: number
  remark: string
  createTimes: number
}

export type ContractMarginAccount = {
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

export type ContractLeverageConfig = {
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

export type RiskUserTradeLimit = {
  id: number
  tenantId: number
  userId: number
  marketType: number
  canOpen: boolean
  canClose: boolean
  canCancel: boolean
  canTriggerOrder: boolean
  canApiTrade: boolean
  tradeEnabled: boolean
  onlyReduceOnly: boolean
  maxOpenOrderCount: number
  maxOrderCountPerDay: number
  maxCancelCountPerDay: number
  maxOpenNotional: string
  maxPositionNotional: string
  riskLevel: number
  operatorId: number
  source: number
  status: number
  effectiveStartTime: number
  effectiveEndTime: number
  remark: string
  createTimes: number
  updateTimes: number
}

export type RiskUserSymbolLimit = {
  id: number
  tenantId: number
  userId: number
  symbolId: number
  marketType: number
  maxPositionQty: string
  maxPositionNotional: string
  maxOpenOrders: number
  maxOrderQty: string
  maxOrderNotional: string
  minOrderQty: string
  minOrderNotional: string
  maxLongPositionQty: string
  maxShortPositionQty: string
  priceDeviationRate: string
  operatorId: number
  source: number
  status: number
  effectiveStartTime: number
  effectiveEndTime: number
  remark: string
  createTimes: number
  updateTimes: number
}

export type RiskOrderCheckLog = {
  id: number
  tenantId: number
  orderNo: string
  clientOrderId: string
  userId: number
  symbolId: number
  marketType: number
  checkType: number
  checkResult: number
  rejectCode: string
  rejectMsg: string
  requestPrice: string
  requestQty: string
  requestAmount: string
  operatorId: number
  source: number
  checkSnapshot: string
  createTimes: number
}

export type BizTradeEvent = {
  id: number
  tenantId: number
  eventNo: string
  eventType: string
  bizId: string
  bizType: string
  userId: number
  symbolId: number
  marketType: number
  operatorId: number
  source: number
  eventStatus: number
  retryCount: number
  maxRetryCount: number
  nextRetryAt: number
  lastErrorMsg: string
  payload: string
  extData: string
  createTimes: number
  updateTimes: number
}

export type CreateSymbolReq = Omit<TradeSymbol, 'id' | 'createTimes' | 'updateTimes'>

export type UpdateSymbolReq = {
  tenantId: number
  id: number
  displaySymbol: string
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
  remark?: string
}

export type GetSymbolListAdminReq = {
  cursor?: number
  limit?: number
  tenantId?: number
  marketType?: number
  status?: number
  keyword?: string
}

export type GetSymbolDetailAdminReq = {
  tenantId?: number
  id: number
}

export type SetSpotSymbolConfigReq = Omit<TradeSymbolSpot, 'id' | 'createTimes' | 'updateTimes'>

export type SetContractSymbolConfigReq = Omit<TradeSymbolContract, 'id' | 'createTimes' | 'updateTimes'>

export type GetOrderListAdminReq = {
  cursor?: number
  limit?: number
  tenantId?: number
  userId?: number
  marketType?: number
  symbolId?: number
  status?: number
  keyword?: string
  timeRange?: TimeRange
}

export type GetOrderDetailAdminReq = {
  tenantId?: number
  id: number
}

export type GetFillListAdminReq = {
  cursor?: number
  limit?: number
  tenantId?: number
  userId?: number
  marketType?: number
  symbolId?: number
  timeRange?: TimeRange
}

export type GetFillDetailAdminReq = {
  tenantId?: number
  id: number
}

export type GetPositionListAdminReq = {
  cursor?: number
  limit?: number
  tenantId?: number
  userId?: number
  marketType?: number
  symbolId?: number
}

export type GetPositionDetailAdminReq = {
  tenantId?: number
  id: number
}

export type GetPositionHistoryListAdminReq = {
  cursor?: number
  limit?: number
  tenantId?: number
  userId?: number
  symbolId?: number
  marketType?: number
  positionId?: number
  actionType?: number
  timeRange?: TimeRange
}

export type GetMarginAccountListAdminReq = {
  cursor?: number
  limit?: number
  tenantId?: number
  userId?: number
  marketType?: number
  marginAsset?: string
}

export type GetCancelLogListAdminReq = {
  cursor?: number
  limit?: number
  tenantId?: number
  userId?: number
  orderId?: number
  orderNo?: string
  cancelSource?: number
  timeRange?: TimeRange
}

export type SetUserTradeLimitReq = Omit<RiskUserTradeLimit, 'id' | 'createTimes' | 'updateTimes'>

export type SetUserSymbolLimitReq = Omit<RiskUserSymbolLimit, 'id' | 'createTimes' | 'updateTimes'>

export type GetUserTradeLimitReq = {
  tenantId?: number
  userId: number
  marketType: number
}

export type GetUserSymbolLimitReq = {
  tenantId?: number
  userId: number
  symbolId: number
  marketType: number
}

export type SetUserTradeConfigReq = {
  tenantId: number
  userId: number
  marketType: number
  symbolId: number
  positionMode: number
  marginMode: number
  defaultLeverage: number
  tradeEnabled: boolean
  reduceOnlyEnabled: boolean
}

export type GetUserTradeConfigReq = {
  tenantId?: number
  userId: number
  marketType: number
  symbolId?: number
}

export type GetRiskOrderCheckLogListReq = {
  cursor?: number
  limit?: number
  tenantId?: number
  userId?: number
  symbolId?: number
  marketType?: number
  checkType?: number
  checkResult?: number
  timeRange?: TimeRange
}

export type SetUserLeverageConfigReq = Omit<ContractLeverageConfig, 'id' | 'createTimes' | 'updateTimes'>

export type GetUserLeverageConfigReq = {
  tenantId?: number
  userId: number
  symbolId: number
  marketType: number
  marginMode: number
}

export type GetTradeEventListReq = {
  cursor?: number
  limit?: number
  tenantId?: number
  eventType?: string
  bizType?: string
  bizId?: string
  eventStatus?: number
  timeRange?: TimeRange
}

export type GetTradeEventDetailReq = {
  tenantId?: number
  id: number
}

export type RetryTradeEventReq = {
  tenantId: number
  id: number
  operatorId: number
}

export class TradeService {
  listSymbols(params: GetSymbolListAdminReq) {
    return apiTradeListSymbols(params)
  }

  getSymbol(params: GetSymbolDetailAdminReq) {
    return apiTradeGetSymbol(params)
  }

  createSymbol(params: CreateSymbolReq) {
    return apiTradeCreateSymbol(params)
  }

  updateSymbol(params: UpdateSymbolReq) {
    return apiTradeUpdateSymbol(params)
  }

  setSpotConfig(params: SetSpotSymbolConfigReq) {
    return apiTradeSetSpotConfig(params)
  }

  setContractConfig(params: SetContractSymbolConfigReq) {
    return apiTradeSetContractConfig(params)
  }

  listOrders(params: GetOrderListAdminReq) {
    return apiTradeListOrders(params)
  }

  getOrder(params: GetOrderDetailAdminReq) {
    return apiTradeGetOrder(params)
  }

  listFills(params: GetFillListAdminReq) {
    return apiTradeListFills(params)
  }

  getFill(params: GetFillDetailAdminReq) {
    return apiTradeGetFill(params)
  }

  listPositions(params: GetPositionListAdminReq) {
    return apiTradeListPositions(params)
  }

  getPosition(params: GetPositionDetailAdminReq) {
    return apiTradeGetPosition(params)
  }

  listPositionHistories(params: GetPositionHistoryListAdminReq) {
    return apiTradeListPositionHistories(params)
  }

  listMarginAccounts(params: GetMarginAccountListAdminReq) {
    return apiTradeListMarginAccounts(params)
  }

  listCancelLogs(params: GetCancelLogListAdminReq) {
    return apiTradeListCancelLogs(params)
  }

  getUserTradeLimit(params: GetUserTradeLimitReq) {
    return apiTradeGetUserTradeLimit(params)
  }

  setUserTradeLimit(params: SetUserTradeLimitReq) {
    return apiTradeSetUserTradeLimit(params)
  }

  getUserSymbolLimit(params: GetUserSymbolLimitReq) {
    return apiTradeGetUserSymbolLimit(params)
  }

  setUserSymbolLimit(params: SetUserSymbolLimitReq) {
    return apiTradeSetUserSymbolLimit(params)
  }

  getUserTradeConfig(params: GetUserTradeConfigReq) {
    return apiTradeGetUserTradeConfig(params)
  }

  setUserTradeConfig(params: SetUserTradeConfigReq) {
    return apiTradeSetUserTradeConfig(params)
  }

  listRiskLogs(params: GetRiskOrderCheckLogListReq) {
    return apiTradeListRiskLogs(params)
  }

  getUserLeverageConfig(params: GetUserLeverageConfigReq) {
    return apiTradeGetUserLeverageConfig(params)
  }

  setUserLeverageConfig(params: SetUserLeverageConfigReq) {
    return apiTradeSetUserLeverageConfig(params)
  }

  listEvents(params: GetTradeEventListReq) {
    return apiTradeListEvents(params)
  }

  getEvent(params: GetTradeEventDetailReq) {
    return apiTradeGetEvent(params)
  }

  retryEvent(params: RetryTradeEventReq) {
    return apiTradeRetryEvent(params)
  }
}

export const tradeService = new TradeService()
