import type { OptionGroup, RespBase } from '@/services'
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
  apiTradeListUserTradeControls,
  apiTradeDisableUserTradeControl,
  apiTradeListUserTradeControlAudits,
  apiTradeListCancelLogs,
  apiTradeListEvents,
  apiTradeListFills,
  apiTradeListMarginSnapshots,
  apiTradeListOrders,
  apiTradeListPositionHistories,
  apiTradeListPositions,
  apiTradeListRiskLogs,
  apiTradeListSymbolLeverageConfigs,
  apiTradeListSymbols,
  apiTradeSetSymbolLeverageConfig,
  apiTradeRetryEvent,
  apiTradeSetContractConfig,
  apiTradeSetSecondsConfig,
  apiTradeSetSpotConfig,
  apiTradeSetUserLeverageConfig,
  apiTradeSetUserSymbolLimit,
  apiTradeSetUserTradeConfig,
  apiTradeSetUserTradeLimit,
  apiTradeUpdateSymbol,
} from '@/api/trade'
import { getCoreOptions } from '@/stores/core'

export type TimeRange = {
  startTime?: number // 开始时间
  endTime?: number // 结束时间
}

export type TradePageReq = { cursor?: number; limit?: number }
export type ContractRiskLimitTier = {
  id: number
  tenantId: number
  symbolId: number
  tierNo: number
  notionalFloor: string
  notionalCap: string
  maxLeverage: number
  initialMarginRate: string
  maintenanceMarginRate: string
  maintenanceAmount: string
  enabled: number
  createTimes: number
  updateTimes: number
}
export type SetContractRiskLimitTierReq = Omit<
  ContractRiskLimitTier,
  'id' | 'tenantId' | 'createTimes' | 'updateTimes'
> & { tenantId?: number; id?: number }
export type GetContractRiskLimitTierListReq = TradePageReq & {
  tenantId?: number
  symbolId?: number
  enabled?: number
}
export type GetContractRiskLimitTierListResp = RespBase<ContractRiskLimitTier[]>
export type ContractFundingBatch = {
  id: number
  batchNo: string
  symbolId: number
  fundingRate: string
  markPrice: string
  markSnapshotId: string
  indexPrice: string
  priceSource: string
  formulaVersion: string
  settlementTime: number
  status: number
  totalPositions: number
  settledPositions: number
  lastErrorMsg: string
  createTimes: number
  updateTimes: number
}
export type GetFundingBatchListReq = TradePageReq & {
  tenantId?: number
  symbolId?: number
  status?: number
  timeRange?: TimeRange
}
export type GetFundingBatchListResp = RespBase<ContractFundingBatch[]>
export type InsuranceFundAccount = {
  id: number
  tenantId: number
  symbolId: number
  settleAsset: string
  adlEnabled: number
  status: number
  version: number
  createTimes: number
  updateTimes: number
}
export type SetInsuranceFundAccountReq = Omit<InsuranceFundAccount, 'createTimes' | 'updateTimes'>
export type GetInsuranceFundAccountListReq = TradePageReq & {
  tenantId?: number
  symbolId?: number
  settleAsset?: string
  status?: number
}
export type GetInsuranceFundAccountListResp = RespBase<InsuranceFundAccount[]>
export type TradeMarketSnapshot = {
  id: number
  snapshotId: string
  snapshotKind: string
  symbolId: number
  source: string
  price: string
  markPrice: string
  indexPrice: string
  fundingRate: string
  sourceTimestamp: number
  snapshotTimestamp: number
  revision: number
  formulaVersion: string
  confirmed: number
  rawPayload: string
  createTimes: number
}
export type GetMarketSnapshotListReq = TradePageReq & {
  tenantId?: number
  symbolId?: number
  snapshotKind?: string
  startTime?: number
  endTime?: number
}
export type GetMarketSnapshotListResp = RespBase<TradeMarketSnapshot[]>
export type ContractFundingSettlement = {
  id: number
  settlementNo: string
  batchId: number
  batchNo: string
  symbolId: number
  userId: number
  positionId: number
  positionSide: number
  fundingRate: string
  markPrice: string
  positionQty: string
  feeAsset: string
  feeAmount: string
  settlementTime: number
  status: number
  retryCount: number
  nextRetryAt: number
  lastErrorMsg: string
  settledAt: number
  createTimes: number
  updateTimes: number
}
export type GetFundingSettlementListReq = TradePageReq & {
  tenantId?: number
  batchId?: number
  userId?: number
  positionId?: number
  status?: number
}
export type GetFundingSettlementListResp = RespBase<ContractFundingSettlement[]>
export type ContractDeliveryBatch = {
  id: number
  batchNo: string
  symbolId: number
  settlementPrice: string
  priceSource: string
  priceAlgorithm: string
  formulaVersion: string
  sampleSnapshot: string
  openCutoffTime: number
  matchingStopTime: number
  deliveryTime: number
  status: number
  totalPositions: number
  settledPositions: number
  lastErrorMsg: string
  createTimes: number
  updateTimes: number
}
export type GetDeliveryBatchListReq = TradePageReq & {
  tenantId?: number
  symbolId?: number
  status?: number
  timeRange?: TimeRange
}
export type GetDeliveryBatchListResp = RespBase<ContractDeliveryBatch[]>
export type ContractDeliverySettlement = {
  id: number
  settlementNo: string
  batchId: number
  batchNo: string
  symbolId: number
  userId: number
  positionId: number
  positionSide: number
  settlementPrice: string
  positionQty: string
  realizedPnl: string
  deliveryFee: string
  settleAsset: string
  deliveryTime: number
  status: number
  retryCount: number
  nextRetryAt: number
  lastErrorMsg: string
  settledAt: number
  createTimes: number
  updateTimes: number
}
export type GetDeliverySettlementListReq = TradePageReq & {
  tenantId?: number
  batchId?: number
  userId?: number
  positionId?: number
  status?: number
}
export type GetDeliverySettlementListResp = RespBase<ContractDeliverySettlement[]>
export type ContractLiquidation = {
  id: number
  liquidationNo: string
  positionId: number
  userId: number
  symbolId: number
  positionSide: number
  marginMode: number
  triggerMarkPrice: string
  triggerSnapshotId: string
  triggerIndexPrice: string
  triggerQty: string
  liquidatedQty: string
  maintenanceMargin: string
  accountEquity: string
  bankruptcyPrice: string
  liquidationFee: string
  insuranceFundAmount: string
  adlQty: string
  status: number
  reason: string
  startedAt: number
  completedAt: number
  createTimes: number
  updateTimes: number
}
export type GetLiquidationListReq = TradePageReq & {
  tenantId?: number
  userId?: number
  symbolId?: number
  positionId?: number
  status?: number
  timeRange?: TimeRange
}
export type GetLiquidationListResp = RespBase<ContractLiquidation[]>
export type ContractAccountLiquidation = {
  id: number
  liquidationNo: string
  userId: number
  marginAsset: string
  marginSnapshotId: number
  marginSnapshotVersion: number
  assetVersion: number
  walletBalance: string
  positionMargin: string
  maintenanceMargin: string
  accountEquity: string
  riskRate: string
  grossSettlement: string
  liquidationFee: string
  userCredit: string
  userDebit: string
  deficitAmount: string
  insuranceFundAmount: string
  adlReliefAmount: string
  adlQty: string
  positionCount: number
  status: number
  reason: string
  startedAt: number
  completedAt: number
  version: number
  createTimes: number
  updateTimes: number
}
export type ContractAccountLiquidationItem = {
  id: number
  accountLiquidationId: number
  liquidationNo: string
  positionId: number
  positionVersion: number
  symbolId: number
  positionSide: number
  triggerQty: string
  triggerMarkPrice: string
  triggerSnapshotId: string
  positionMargin: string
  maintenanceMargin: string
  realizedPnl: string
  liquidationFee: string
  deficitAmount: string
  bankruptcyPrice: string
  adlReliefAmount: string
  adlQty: string
  status: number
  createTimes: number
  updateTimes: number
}
export type GetAccountLiquidationListReq = TradePageReq & {
  tenantId?: number
  userId?: number
  marginAsset?: string
  status?: number
  timeRange?: TimeRange
}
export type GetAccountLiquidationListResp = RespBase<ContractAccountLiquidation[]>
export type GetAccountLiquidationDetailReq = { tenantId?: number; id: number }
export type GetAccountLiquidationDetailResp = RespBase<ContractAccountLiquidation> & {
  items: ContractAccountLiquidationItem[]
  settlementInstructions: TradeSettlementInstruction[]
}
export type RetryAccountLiquidationReq = { tenantId?: number; id: number; reason: string }
export type TradeSecondsPriceSnapshot = {
  id: number
  orderId: number
  snapshotType: number
  source: string
  price: string
  quoteTime: number
  receivedAt: number
  algorithm: string
  isSelected: number
  rawPayload: string
  createTimes: number
}
export type GetSecondsPriceSnapshotListReq = TradePageReq & {
  tenantId?: number
  orderId?: number
  snapshotType?: number
}
export type GetSecondsPriceSnapshotListResp = RespBase<TradeSecondsPriceSnapshot[]>
export type TradeAssetReservation = {
  id: number
  orderId: number
  reservationNo: string
  asset: string
  reservedAmount: string
  consumedAmount: string
  releasedAmount: string
  status: number
  retryCount: number
  nextRetryAt: number
  lastErrorMsg: string
  version: number
  createTimes: number
  updateTimes: number
}
export type GetAssetReservationListReq = TradePageReq & {
  tenantId?: number
  orderId?: number
  status?: number
}
export type GetAssetReservationListResp = RespBase<TradeAssetReservation[]>
export type TradeSettlementInstruction = {
  id: number
  instructionNo: string
  bizType: string
  bizId: string
  batchNo: string
  fillId: number
  orderId: number
  positionId: number
  reservationNo: string
  userId: number
  action: number
  asset: string
  amount: string
  stepNo: number
  status: number
  retryCount: number
  nextRetryAt: number
  lastErrorMsg: string
  assetFlowNo: string
  reconciledAt: number
  createTimes: number
  updateTimes: number
}
export type GetSettlementInstructionListReq = TradePageReq & {
  tenantId?: number
  bizType?: string
  bizId?: string
  orderId?: number
  status?: number
}
export type GetSettlementInstructionListResp = RespBase<TradeSettlementInstruction[]>
export type RetrySettlementInstructionReq = { tenantId?: number; id: number; reason: string }
export type ContractReconciliationIssue = {
  id: number
  issueKey: string
  checkType: string
  bizType: string
  bizNo: string
  instructionId: number
  expectedValue: string
  actualValue: string
  detail: string
  status: number
  occurrenceCount: number
  firstSeenAt: number
  lastSeenAt: number
  resolvedAt: number
  operatorId: number
  resolutionReason: string
  createTimes: number
  updateTimes: number
}
export type GetContractReconciliationIssueListReq = TradePageReq & {
  tenantId?: number
  status?: number
  checkType?: string
  bizNo?: string
}
export type GetContractReconciliationIssueListResp = RespBase<ContractReconciliationIssue[]>
export type IgnoreContractReconciliationIssueReq = {
  tenantId?: number
  id: number
  reason: string
}

export type TradeSymbol = {
  id: number // 主键ID
  tenantId: number // 租户ID
  symbol: string // 交易对编码
  displaySymbol: string // 展示名称
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
  sort: number // 排序
  remark: string // 备注
  createTimes: number // 创建时间
  updateTimes: number // 更新时间
}

export type TradeSymbolSpot = {
  id: number
  tenantId: number
  symbolId: number
  makerFeeRate: string
  takerFeeRate: string
  buyEnabled: number
  sellEnabled: number
  createTimes: number // 创建时间
  updateTimes: number // 更新时间
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
  fundingIntervalMinutes: number // 资金费率间隔分钟数
  deliveryTime: number // 交割时间
  supportCross: number // 全仓支持状态：0不支持 1支持
  supportIsolated: number // 逐仓支持状态：0不支持 1支持
  fundingRateCap: string
  fundingRateFloor: string
  fundingRateSource: string
  indexSymbol: string
  markPriceSource: string
  settlementPriceSource: string
  openCutoffTime: number
  matchingStopTime: number
  settlementWindowSeconds: number
  settlementPriceAlgorithm: string
  deliveryFeeRate: string
  liquidationFeeRate: string
  openLongEnabled: number
  openShortEnabled: number
  closeLongEnabled: number
  closeShortEnabled: number
  createTimes: number // 创建时间
  updateTimes: number // 更新时间
}

export type TradeSymbolSeconds = {
  id: number
  tenantId: number
  symbolId: number
  durationSeconds: number
  payoutRate: string
  feeRate: string
  drawRule: number
  startPriceSource: string
  settlementPriceSource: string
  quoteValidityMs: number
  settlementWindowMs: number
  settlementPriceAlgorithm: string
  drawTolerance: string
  maxExposureAmount: string
  minStake: string
  maxStake: string
  upEnabled: number
  downEnabled: number
  createTimes: number
  updateTimes: number
}

export type TradeSymbolLeverageConfig = {
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

export type TradeSymbolDetailData = {
  symbol: TradeSymbol
  spot?: TradeSymbolSpot
  contract?: TradeSymbolContract
  secondsConfigs?: TradeSymbolSeconds[]
  leverageConfigs?: TradeSymbolLeverageConfig[]
}

export type TradeSymbolDetailResp = RespBase<TradeSymbolDetailData>

export type TradeUserConfig = {
  id: number
  tenantId: number
  userId: number
  productType: number
  symbolId: number
  tradeEnabled: number // 交易开关：1启用 2禁用
  createTimes: number // 创建时间
  updateTimes: number // 更新时间
}

export type TradeOrder = {
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
  requestHash: string
  canceledQty: string
  isClosePosition: number
  ocoGroupNo: string
  expireAt: number
  triggeredAt: number
  completionReason: string
  cancelReason: string
  bizExt: string
  version: number
  secondsDirection: number
  durationSeconds: number
  createTimes: number // 创建时间
  updateTimes: number // 更新时间
}

export type TradeOrderSeconds = {
  id: number
  tenantId: number
  orderId: number
  direction: number
  durationSeconds: number
  stakeAsset: string
  stakeAmount: string
  payoutRate: string
  feeRate: string
  frozenAt: number
  activatedAt: number
  startPrice: string
  startPriceTime: number
  startPriceSource: string
  expireTime: number
  settlementPrice: string
  settlementPriceTime: number
  settlementPriceSource: string
  priceAlgorithm: string
  result: number
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

export type TradeOrderSpot = {
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

export type TradeOrderContract = {
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

export type TradeOrderDetailData = {
  order: TradeOrder
  spot?: TradeOrderSpot
  contract?: TradeOrderContract
  seconds?: TradeOrderSeconds
}

export type TradeFill = {
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
  contractType: number
  contractValueType: number
  positionSide: number
  positionMode: number
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
  maintenanceMargin: string
  bankruptcyPrice: string
  riskRate: string
  status: number
  adlRank: number
  lastFundingTime: number
  closedAt: number
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
  contractType: number
  contractValueType: number
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

export type ContractMarginSnapshot = {
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
  createTimes: number
  updateTimes: number
  sourceEventNo: string
  snapshotTime: number
}

export type ContractLeverageConfig = {
  id: number
  tenantId: number
  userId: number
  symbolId: number
  marginMode: number
  longLeverage: number
  shortLeverage: number
  operatorId: number
  source: number
  enabled: number // 启用状态
  remark: string // 备注
  createTimes: number // 创建时间
  updateTimes: number // 更新时间
}

export type RiskUserTradeLimit = {
  id: number
  tenantId: number
  userId: number
  productType: number
  canOpen: number // 开仓权限：0禁止 1允许
  canClose: number // 平仓权限：0禁止 1允许
  canCancel: number // 撤单权限：0禁止 1允许
  canTriggerOrder: number // 条件单权限：0禁止 1允许
  canApiTrade: number // API交易权限：0禁止 1允许
  tradeEnabled: number // 交易开关：1启用 2禁用
  onlyReduceOnly: number // 仅减仓开关：1启用 2禁用
  maxOpenOrderCount: number
  maxOrderCountPerDay: number
  maxCancelCountPerDay: number
  maxOpenNotional: string
  maxPositionNotional: string
  riskLevel: number
  operatorId: number
  source: number
  enabled: number // 启用状态
  effectiveStartTime: number
  effectiveEndTime: number
  remark: string
  createTimes: number
  updateTimes: number
  contractType: number
  controlMode: number
  version: number
}

export type RiskUserSymbolLimit = {
  id: number
  tenantId: number
  userId: number
  symbolId: number
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
  enabled: number
  effectiveStartTime: number
  effectiveEndTime: number
  remark: string
  createTimes: number
  updateTimes: number
  controlMode: number
  version: number
}

export type UserTradeControlEntry = {
  scopeType: number
  productLimit?: RiskUserTradeLimit
  symbolLimit?: RiskUserSymbolLimit
}

export type TradeUserControlAudit = {
  id: number
  tenantId: number
  controlId: number
  scopeType: number
  userId: number
  changeType: number
  beforeJson: string
  afterJson: string
  operatorId: number
  source: number
  reason: string
  requestId: string
  createTimes: number
}

export type ListUserTradeControlsReq = TradePageReq & {
  tenantId?: number
  userId?: number
  productType?: number
  contractType?: number
  symbolId?: number
  enabled?: number
  scopeType?: number
}

export type DisableUserTradeControlReq = {
  tenantId: number
  controlId: number
  scopeType: number
  expectedVersion: number
  reason: string
}

export type ListUserTradeControlAuditsReq = TradePageReq & {
  tenantId?: number
  controlId?: number
  userId?: number
  scopeType?: number
}

export type RiskOrderCheckLog = {
  id: number
  tenantId: number
  orderNo: string
  clientOrderId: string
  userId: number
  symbolId: number
  productType: number
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
  productType: number
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
  consumer: string
  payloadVersion: number
  claimedBy: string
  claimedAt: number
  deliveredAt: number
}

export type CreateSymbolReq = Omit<TradeSymbol, 'id' | 'createTimes' | 'updateTimes'>

export type UpdateSymbolReq = {
  tenantId: number // 租户ID
  id: number // 交易对ID
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
  maxNotional: string
  listingTime: number
  tradingStartTime: number
  tradingEndTime: number
  sort: number
  remark?: string // 备注
}

export type GetSymbolListAdminReq = {
  cursor?: number // 游标
  limit?: number // 每页条数
  tenantId?: number // 租户ID
  productType?: number
  status?: number // 状态
  keyword?: string // 关键字
}

export type GetSymbolDetailAdminReq = {
  tenantId?: number // 租户ID
  id: number // 交易对ID
}

export type SetSpotSymbolConfigReq = Omit<TradeSymbolSpot, 'id' | 'createTimes' | 'updateTimes'>

export type SetContractSymbolConfigReq = Omit<
  TradeSymbolContract,
  'id' | 'createTimes' | 'updateTimes'
>

export type SetSecondsSymbolConfigReq = Omit<
  TradeSymbolSeconds,
  'id' | 'createTimes' | 'updateTimes'
>

export type SetSymbolLeverageConfigReq = {
  tenantId: number
  symbolId: number
  marginMode: number
  leverageValues: number[]
  defaultLeverage: number
  enabled: number
  sort: number
  remark: string
}

export type GetSymbolLeverageConfigListReq = {
  cursor?: number
  limit?: number
  tenantId?: number
  symbolId?: number
  marginMode?: number
  enabled?: number
}

export type GetOrderListAdminReq = {
  cursor?: number
  limit?: number
  tenantId?: number
  userId?: number
  productType?: number
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
  productType?: number
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
  contractType?: number
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
  contractType?: number
  positionId?: number
  actionType?: number
  timeRange?: TimeRange
}

export type GetMarginSnapshotListAdminReq = {
  cursor?: number
  limit?: number
  tenantId?: number
  userId?: number
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

export type SetUserTradeLimitReq = Omit<
  RiskUserTradeLimit,
  'id' | 'createTimes' | 'updateTimes' | 'version'
> & { expectedVersion: number }

export type SetUserSymbolLimitReq = Omit<
  RiskUserSymbolLimit,
  'id' | 'createTimes' | 'updateTimes' | 'version'
> & { expectedVersion: number }

export type GetUserTradeLimitReq = {
  tenantId?: number
  userId?: number
  productType?: number
  contractType?: number
}

export type GetUserSymbolLimitReq = {
  tenantId?: number
  userId?: number
  symbolId?: number
}

export type SetUserTradeConfigReq = {
  tenantId: number // 租户ID
  userId: number // 用户ID
  productType: number
  symbolId: number // 交易对ID
  tradeEnabled: number // 交易开关：1启用 2禁用
}

export type GetUserTradeConfigReq = {
  tenantId?: number
  userId?: number
  productType?: number
  symbolId?: number
}

export type GetRiskOrderCheckLogListReq = {
  cursor?: number
  limit?: number
  tenantId?: number
  userId?: number
  symbolId?: number
  orderNo?: string
  productType?: number
  checkType?: number
  checkResult?: number
  timeRange?: TimeRange
}

export type SetUserLeverageConfigReq = Omit<
  ContractLeverageConfig,
  'id' | 'createTimes' | 'updateTimes'
>

export type GetUserLeverageConfigReq = {
  tenantId?: number
  userId?: number
  symbolId?: number
  marginMode?: number
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
  tenantId: number // 租户ID
  id: number // 事件ID
  operatorId: number // 操作人ID
  eventNo: string
}

export class TradeService {
  getOptions(): Promise<RespBase<OptionGroup[]>> {
    return getCoreOptions()
  }

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

  setSecondsConfig(params: SetSecondsSymbolConfigReq) {
    return apiTradeSetSecondsConfig(params)
  }

  listSymbolLeverageConfigs(params: GetSymbolLeverageConfigListReq) {
    return apiTradeListSymbolLeverageConfigs(params)
  }

  setSymbolLeverageConfig(params: SetSymbolLeverageConfigReq) {
    return apiTradeSetSymbolLeverageConfig(params)
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

  listMarginSnapshots(params: GetMarginSnapshotListAdminReq) {
    return apiTradeListMarginSnapshots(params)
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

  listUserTradeControls(params: ListUserTradeControlsReq) {
    return apiTradeListUserTradeControls(params)
  }

  disableUserTradeControl(params: DisableUserTradeControlReq) {
    return apiTradeDisableUserTradeControl(params)
  }

  listUserTradeControlAudits(params: ListUserTradeControlAuditsReq) {
    return apiTradeListUserTradeControlAudits(params)
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
