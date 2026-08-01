import type { OptionGroup, RespBase } from '@/services'
import {
  apiOptionCreateContract,
  apiOptionGetAccount,
  apiOptionGetBill,
  apiOptionGetContract,
  apiOptionGetExercise,
  apiOptionRetryExercise,
  apiOptionGetMarket,
  apiOptionGetOrder,
  apiOptionGetPosition,
  apiOptionGetSettlement,
  apiOptionGetTrade,
  apiOptionForceCancelContractOrders,
  apiOptionForceCancelComboOrder,
  apiOptionGetAdminComboOrder,
  apiOptionReleaseUserKillSwitch,
  apiOptionGetUserTradingControl,
  apiOptionListTradingControlEvents,
  apiOptionUpsertMMPConfig,
  apiOptionResetMMPConfig,
  apiOptionListMMPConfigs,
  apiOptionCreatePortfolioRiskConfig,
  apiOptionReviewPortfolioRiskConfig,
  apiOptionListPortfolioRiskConfigs,
  apiOptionCreateTradeCorrection,
  apiOptionReviewTradeCorrection,
  apiOptionListTradeCorrections,
  apiOptionRetryAssetInstruction,
  apiOptionGetOperationsOverview,
  apiOptionListAssetInstructions,
  apiOptionListReconciliationIssues,
  apiOptionCreateTradingCalendar,
  apiOptionReviewTradingCalendar,
  apiOptionListTradingCalendars,
  apiOptionHaltContractTrading,
  apiOptionResumeContractTrading,
  apiOptionListTradingHalts,
  apiOptionCreateCorporateAction,
  apiOptionReviewCorporateAction,
  apiOptionListCorporateActions,
  apiOptionListCorporateActionPositions,
  apiOptionCreateContractSeries,
  apiOptionReviewContractSeries,
  apiOptionListContractSeries,
  apiOptionListContractSeriesDetails,
  apiOptionReviewContractSeriesLaunch,
  apiOptionRetrySettlementInstruction,
  apiOptionListPhysicalDeliveryUnits,
  apiOptionRetryPhysicalDeliveryUnit,
  apiOptionRetryTradeEvent,
  apiOptionListRiskAccounts,
  apiOptionListLiquidations,
  apiOptionRetryLiquidation,
  apiOptionListAccounts,
  apiOptionListAdminComboOrders,
  apiOptionListBills,
  apiOptionListContracts,
  apiOptionListExercises,
  apiOptionListMarketSnapshots,
  apiOptionListOrders,
  apiOptionListPositions,
  apiOptionListSettlements,
  apiOptionListSettlementPrices,
  apiOptionCreateSettlementPriceCorrection,
  apiOptionReviewSettlementPrice,
  apiOptionListTrades,
  apiOptionUpdateContract,
  apiOptionUpdateMarket,
} from '@/api/option'
import { getCoreOptions } from '@/stores/core'
export type OptionAdminCommonResp = RespBase

export type TimeRange = {
  startTime?: number // 开始时间
  endTime?: number // 结束时间
}

export type OptionContract = {
  id: number // 主键ID
  tenantId: number // 租户ID
  contractCode: string // 合约编码
  underlyingSymbol: string // 标的符号
  underlyingCoin: string
  settleCoin: string
  quoteCoin: string
  optionType: number
  exerciseStyle: number
  settlementType: number
  strikePrice: string
  contractUnit: string
  minOrderQty: string
  maxOrderQty: string
  priceTick: string
  qtyStep: string
  multiplier: string
  listTime: number
  expireTime: number
  deliverTime: number
  exerciseCutoffTime: number
  autoExerciseThreshold: string
  maxUserLongQty: string
  maxUserShortQty: string
  maxOpenInterest: string
  orderPriceBandRatio: string
  circuitBreakerRatio: string
  greeksMaxAgeSeconds: number
  settlementPriceSource: string
  settlementPriceMethod: string
  settlementWindowSeconds: number
  settlementMinSamples: number
  isAutoExercise: number // 是否自动行权：1是 2否
  status: number
  sort: number // 排序
  remark: string // 备注
  isDeleted: number // 是否删除：1是 2否
  createTimes: number // 创建时间
  updateTimes: number // 更新时间
  makerFeeRate: string
  takerFeeRate: string
  exerciseFeeRate: string
  feeUserId: number
  feeAccountId: number
  sellerMarginMode: number
  initialMarginRate: string
  maintenanceMarginRate: string
  minMarginRate: string
  liquidationFeeRate: string
  insuranceUserId: number
  insuranceAccountId: number
  liquidationDeficitPolicy: number
  physicalDeliveryPolicy: number
  physicalDeliveryCureSeconds: number
  tradingCalendarCode: string
}

export type OptionMarket = {
  id: number // 主键ID
  tenantId: number // 租户ID
  contractId: number // 合约ID
  underlyingPrice: string
  markPrice: string
  lastPrice: string
  bidPrice: string
  askPrice: string
  theoreticalPrice: string
  intrinsicValue: string
  timeValue: string
  iv: string
  delta: string
  gamma: string
  theta: string
  vega: string
  rho: string
  riskFreeRate: string
  pricingModel: string
  snapshotTime: number // 快照时间
  underlyingSnapshotTime: number
  markSnapshotTime: number
  greeksSnapshotTime: number
  createTimes: number // 创建时间
  updateTimes: number // 更新时间
}

export type OptionMarketSnapshot = {
  id: number
  tenantId: number
  contractId: number
  underlyingPrice: string
  markPrice: string
  lastPrice: string
  bidPrice: string
  askPrice: string
  theoreticalPrice: string
  iv: string
  delta: string
  gamma: string
  theta: string
  vega: string
  rho: string
  snapshotTime: number
  sourceType: number
  sourceSnapshotId: string
  createTimes: number
}

export type OptionOrder = {
  id: number // 主键ID
  tenantId: number // 租户ID
  orderNo: string // 订单号
  userId: number
  accountId: number
  contractId: number
  underlyingSymbol: string
  side: number
  positionEffect: number
  orderType: number
  price: string
  qty: string
  filledQty: string
  unfilledQty: string
  avgPrice: string
  turnover: string
  fee: string
  feeCoin: string
  marginAmount: string
  marginCoin: string
  source: number
  clientOrderId: string
  reduceOnly: number // 是否只减仓：1是 2否
  mmp: number // 是否做市商保护单：1是 2否
  mmpGroup: string
  comboOrderId: number
  comboLegNo: number
  portfolioRiskConfigId: number
  portfolioRiskConfigVersion: number
  status: number
  cancelReason: string
  matchTime: number
  cancelTime: number
  createTimes: number // 创建时间
  updateTimes: number // 更新时间
}

export type OptionTrade = {
  id: number // 主键ID
  tenantId: number // 租户ID
  tradeNo: string // 成交号
  contractId: number
  underlyingSymbol: string
  buyOrderId: number
  buyOrderNo: string
  buyUserId: number
  buyAccountId: number
  sellOrderId: number
  sellOrderNo: string
  sellUserId: number
  sellAccountId: number
  price: string
  qty: string
  turnover: string
  buyFee: string
  sellFee: string
  feeCoin: string
  makerSide: number
  matchSequence: number
  comboMatchNo: string
  comboLegNo: number
  tradeTime: number
  createTimes: number // 创建时间
}

export type OptionComboOrder = {
  id: number
  tenantId: number
  comboNo: string
  userId: number
  accountId: number
  clientComboId: string
  strategyKey: string
  inverseStrategyKey: string
  underlyingSymbol: string
  expireTime: number
  settleCoin: string
  quoteCoin: string
  orderType: number
  netPrice: string
  qty: string
  filledQty: string
  unfilledQty: string
  status: number
  payloadHash: string
  cancelReason: string
  cancelTime: number
  createTimes: number
  updateTimes: number
}

export type OptionComboOrderLeg = {
  id: number
  tenantId: number
  comboOrderId: number
  legNo: number
  contractId: number
  side: number
  positionEffect: number
  ratio: number
  price: string
  qty: string
  filledQty: string
  unfilledQty: string
  childOrderId: number
  createTimes: number
  updateTimes: number
}

export type OptionComboOrderDetail = {
  comboOrder: OptionComboOrder
  legs: OptionComboOrderLeg[]
}

export type OptionAdminComboOrderDetail = OptionComboOrderDetail & {
  childOrders: OptionOrderDetail[]
  trades: OptionTradeDetail[]
  assetInstructions: OptionAssetInstruction[]
  tradeTotal: number
  assetInstructionTotal: number
  dataTruncated: boolean
}

export type OptionPosition = {
  id: number
  tenantId: number
  userId: number
  accountId: number
  contractId: number
  underlyingSymbol: string
  side: number
  positionQty: string
  availableQty: string
  frozenQty: string
  openAvgPrice: string
  markPrice: string
  positionValue: string
  marginAmount: string
  maintenanceMargin: string
  unrealizedPnl: string
  realizedPnl: string
  tradeRealizedPnl: string
  settlementRealizedPnl: string
  feePaid: string
  totalReturn: string
  exerciseableQty: string
  status: number
  lastCalcTime: number
  createTimes: number
  updateTimes: number
}

export type OptionExercise = {
  id: number
  tenantId: number
  exerciseNo: string
  clientExerciseId: string
  userId: number
  accountId: number
  contractId: number
  positionId: number
  exerciseType: number
  exerciseQty: string
  strikePrice: string
  settlementPrice: string
  exerciseAmount: string
  profitAmount: string
  fee: string
  feeCoin: string
  status: number
  remark: string
  exerciseTime: number
  finishTime: number
  createTimes: number
  updateTimes: number
}

export type OptionExerciseAssignment = {
  id: number
  tenantId: number
  exerciseId: number
  exerciseNo: string
  longPositionId: number
  shortPositionId: number
  shortUserId: number
  shortAccountId: number
  quantity: string
  payoff: string
  status: number
  instructionNo: string
  createTimes: number
  updateTimes: number
}

export type OptionAssetInstruction = {
  id: number
  tenantId: number
  instructionNo: string
  bizNo: string
  orderId: number
  tradeId: number
  positionId: number
  userId: number
  accountId: number
  action: number
  targetBizNo: string
  coin: string
  amount: string
  stepNo: number
  status: number
  retryCount: number
  nextRetryAt: number
  lastErrorMsg: string
  createTimes: number
  updateTimes: number
  assetFlowNo: string
  reconciliationStatus: number
  reconciledAt: number
  marginLotId: number
  liquidationId: number
  deliveryUnitId: number
  executionGroup: string
}

export type OptionReconciliationIssue = {
  id: number
  tenantId: number
  issueKey: string
  checkType: number
  bizNo: string
  instructionId: number
  expectedValue: string
  actualValue: string
  detail: string
  status: number
  occurrenceCount: number
  resolvedAt: number
  createTimes: number
  updateTimes: number
}

export type OptionCoinAmount = {
  coin: string
  amount: string
}

export type OptionOperationsOverview = {
  generatedAt: number
  riskStaleSeconds: number
  comboStaleSeconds: number
  assetPendingCount: number
  assetFailedCount: number
  assetManualReviewCount: number
  oldestAssetInstructionTime: number
  openReconciliationCount: number
  oldestReconciliationTime: number
  pendingSettlementPriceCount: number
  staleRiskAccountCount: number
  oldestRiskCalcTime: number
  pendingExerciseCount: number
  oldestExerciseTime: number
  pendingSettlementCount: number
  failedSettlementCount: number
  oldestSettlementTime: number
  pendingLiquidationCount: number
  exceptionLiquidationCount: number
  oldestLiquidationTime: number
  pendingOutboxCount: number
  oldestOutboxTime: number
  pendingInboxCount: number
  oldestInboxTime: number
  physicalExceptionCount: number
  comboStaleCount: number
  comboManualReviewCount: number
  oldestComboExceptionTime: number
  comboInvariantIssueCount: number
  comboIncompleteMatchGroupCount: number
  insuranceLedger: OptionCoinAmount[]
  backstopLiability: OptionCoinAmount[]
  unresolvedDeficit: OptionCoinAmount[]
}

export type OptionPhysicalDeliveryUnit = {
  id: number
  tenantId: number
  deliveryUnitNo: string
  batchId: number
  batchNo: string
  contractId: number
  longPositionId: number
  longUserId: number
  longAccountId: number
  shortPositionId: number
  shortUserId: number
  shortAccountId: number
  quantity: string
  deliveryCoin: string
  deliveryQuantity: string
  paymentCoin: string
  paymentAmount: string
  status: number
  cureDeadline: number
  failedInstructionId: number
  lastErrorMsg: string
  completedAt: number
  createTimes: number
  updateTimes: number
  manualRetryCount: number
  assetInstructions: OptionAssetInstruction[]
}

export type OptionSettlement = {
  id: number
  tenantId: number
  settlementNo: string
  contractId: number
  underlyingSymbol: string
  expireTime: number
  settlementTime: number
  deliveryPrice: string
  theoreticalPrice: string
  iv: string
  isItm: number // 是否实值：1是 2否
  exerciseResult: number
  status: number
  remark: string
  createTimes: number
  updateTimes: number
}

export type OptionSettlementPrice = {
  id: number
  tenantId: number
  contractId: number
  priceSource: string
  windowStart: number
  windowEnd: number
  sampleCount: number
  calculationMethod: string
  deliveryPrice: string
  sourceSnapshotIds: string
  version: number
  status: number
  confirmedBy: number
  confirmedAt: number
  createTimes: number
  updateTimes: number
  supersedesId: number
  changeReason: string
  createdBy: number
}

export type OptionAccount = {
  id: number
  tenantId: number
  userId: number
  accountId: number
  marginCoin: string
  balance: string
  availableBalance: string
  frozenBalance: string
  positionMargin: string
  orderMargin: string
  unrealizedPnl: string
  realizedPnl: string
  riskRate: string
  status: number
  createTimes: number
  updateTimes: number
}

export type OptionBill = {
  id: number
  tenantId: number
  userId: number
  accountId: number
  bizNo: string
  refType: number
  refId: number
  coin: string
  changeAmount: string
  balanceBefore: string
  balanceAfter: string
  remark: string
  createTimes: number
}

export type OptionContractDetail = {
  contract: OptionContract
  market: OptionMarket
}

export type OptionPositionDetail = {
  position: OptionPosition
  contract: OptionContract
  market: OptionMarket
}

export type OptionOrderDetail = {
  order: OptionOrder
  contract: OptionContract
}

export type OptionTradeDetail = {
  trade: OptionTrade
  contract: OptionContract
}

export type OptionExerciseDetail = {
  exercise: OptionExercise
  contract: OptionContract
  assignments: OptionExerciseAssignment[]
  assetInstructions: OptionAssetInstruction[]
}

export type OptionSettlementDetail = {
  settlement: OptionSettlement
  contract: OptionContract
  batch: {
    id: number
    batchNo: string
    instructionCount: number
    successCount: number
    status: number
    lastErrorMsg: string
  }
  positionDetails: Array<{
    positionId: number
    userId: number
    accountId: number
    side: number
    quantity: string
    deliveryCoin: string
    deliveryQuantity: string
    paymentCoin: string
    paymentAmount: string
  }>
  assetInstructions: OptionAssetInstruction[]
}

export type CreateContractReq = Omit<
  OptionContract,
  'id' | 'isDeleted' | 'createTimes' | 'updateTimes'
>

export type UpdateContractReq = Omit<OptionContract, 'createTimes' | 'updateTimes'>

export type OptionForceCancelContractOrdersReq = {
  tenantId: number
  contractId: number
  reason?: string
}

export type OptionReleaseUserKillSwitchReq = {
  tenantId: number
  userId: number
  reason: string
}

export type OptionUserTradingControl = {
  tenantId: number
  userId: number
  killSwitch: number
  reason: string
  activatedAt: number
  releasedAt: number
  updateTimes: number
}

export type OptionTradingControlEvent = {
  id: number
  tenantId: number
  userId: number
  contractId: number
  orderId: number
  eventType: string
  reason: string
  detail: string
  operatorId: number
  createTimes: number
}

export type AdminGetUserTradingControlReq = {
  tenantId: number
  userId: number
}

export type ListTradingControlEventsReq = {
  cursor?: number
  limit?: number
  tenantId: number
  userId?: number
  contractId?: number
  eventType?: string
  reason?: string
}

export type OptionMMPConfig = {
  id: number
  tenantId: number
  userId: number
  contractId: number
  groupCode: string
  enabled: number
  qtyThreshold: string
  tradeCountThreshold: number
  lossThreshold: string
  windowSeconds: number
  cooldownSeconds: number
  status: number
  windowStart: number
  accumulatedQty: string
  tradeCount: number
  accumulatedLoss: string
  triggeredAt: number
  cooldownUntil: number
  triggerReason: string
  lastErrorMsg: string
  createdBy: number
  updatedBy: number
  createTimes: number
  updateTimes: number
}

export type UpsertMMPConfigReq = {
  tenantId: number
  userId: number
  contractId: number
  groupCode: string
  enabled: number
  qtyThreshold: string
  tradeCountThreshold: number
  lossThreshold: string
  windowSeconds: number
  cooldownSeconds: number
  reason: string
}

export type ResetMMPConfigReq = {
  tenantId: number
  userId: number
  contractId: number
  groupCode: string
  reason: string
}

export type ListMMPConfigsReq = {
  cursor?: number
  limit?: number
  tenantId: number
  userId?: number
  contractId?: number
  groupCode?: string
  status?: number
}

export type OptionPortfolioRiskConfig = {
  id: number
  tenantId: number
  settleCoin: string
  version: number
  status: number
  modelMethod: number
  initialShockRate: string
  maintenanceShockRate: string
  scenarioShocks: string
  concentrationThreshold: string
  concentrationAddonRate: string
  liquidityAddonRate: string
  effectiveFrom: number
  effectiveUntil: number
  supersedesId: number
  changeReason: string
  evidenceRef: string
  createdBy: number
  reviewedBy: number
  reviewReason: string
  reviewedAt: number
  createTimes: number
  updateTimes: number
}

export type CreatePortfolioRiskConfigReq = {
  tenantId: number
  settleCoin: string
  modelMethod: number
  initialShockRate: string
  maintenanceShockRate: string
  scenarioShocks: string
  concentrationThreshold: string
  concentrationAddonRate: string
  liquidityAddonRate: string
  effectiveFrom: number
  changeReason: string
  evidenceRef: string
  sourceConfigId?: number
}

export type ReviewPortfolioRiskConfigReq = {
  tenantId: number
  configId: number
  approve: boolean
  reason: string
}

export type ListPortfolioRiskConfigsReq = {
  cursor?: number
  limit?: number
  tenantId: number
  settleCoin?: string
  status?: number
}

export type OptionTradeCorrectionLeg = {
  id: number
  legNo: number
  userId: number
  accountId: number
  coin: string
  direction: number
  amount: string
  instructionNo: string
}

export type OptionTradeCorrection = {
  id: number
  tenantId: number
  caseNo: string
  tradeId: number
  contractId: number
  action: number
  status: number
  reason: string
  evidenceRef: string
  requestedBy: number
  reviewedBy: number
  reviewReason: string
  reviewedAt: number
  completedAt: number
  lastErrorMsg: string
  createTimes: number
  updateTimes: number
  legs: OptionTradeCorrectionLeg[]
}

export type TradeCorrectionLegInput = {
  userId: number
  accountId: number
  coin: string
  direction: number
  amount: string
}

export type CreateTradeCorrectionReq = {
  tenantId: number
  tradeId: number
  action: 1
  reason: string
  evidenceRef: string
  legs: TradeCorrectionLegInput[]
}

export type ReviewTradeCorrectionReq = {
  tenantId: number
  correctionId: number
  approve: boolean
  reason: string
}

export type ListTradeCorrectionsReq = {
  cursor?: number
  limit?: number
  tenantId: number
  tradeId?: number
  contractId?: number
  status?: number
}

export type OptionRetryAssetInstructionReq = {
  tenantId: number
  instructionId: number
  reason: string
}

export type GetOperationsOverviewReq = {
  tenantId: number
  riskStaleSeconds?: number
  comboStaleSeconds?: number
}

export type ListAssetInstructionsReq = {
  tenantId: number
  userId?: number
  bizNo?: string
  status?: number
  reconciliationStatus?: number
  cursor?: number
  limit?: number
}

export type ListReconciliationIssuesReq = {
  tenantId: number
  bizNo?: string
  checkType?: number
  status?: number
  cursor?: number
  limit?: number
}

export type OptionTradingCalendarSession = {
  id: number
  tenantId: number
  calendarId: number
  weekday: number
  openSecond: number
  closeSecond: number
  createTimes: number
}

export type OptionTradingCalendarException = {
  id: number
  tenantId: number
  calendarId: number
  exceptionType: number
  startTime: number
  endTime: number
  reason: string
  announcementRef: string
  createTimes: number
}

export type OptionTradingCalendar = {
  id: number
  tenantId: number
  calendarCode: string
  version: number
  status: number
  timezone: string
  effectiveFrom: number
  effectiveUntil: number
  supersedesId: number
  changeReason: string
  evidenceRef: string
  createdBy: number
  reviewedBy: number
  reviewReason: string
  reviewedAt: number
  createTimes: number
  updateTimes: number
  sessions: OptionTradingCalendarSession[]
  exceptions: OptionTradingCalendarException[]
}

export type TradingCalendarSessionInput = {
  weekday: number
  openSecond: number
  closeSecond: number
}

export type TradingCalendarExceptionInput = {
  exceptionType: number
  startTime: number
  endTime: number
  reason: string
  announcementRef?: string
}

export type CreateTradingCalendarReq = {
  tenantId: number
  calendarCode: string
  timezone?: string
  effectiveFrom?: number
  changeReason: string
  evidenceRef: string
  sessions?: TradingCalendarSessionInput[]
  exceptions?: TradingCalendarExceptionInput[]
  sourceCalendarId?: number
}

export type ReviewTradingCalendarReq = {
  tenantId: number
  calendarId: number
  approve: boolean
  reason: string
}

export type ListTradingCalendarsReq = {
  tenantId: number
  calendarCode?: string
  status?: number
  cursor?: number
  limit?: number
}

export type OptionTradingHalt = {
  id: number
  tenantId: number
  haltNo: string
  contractId: number
  source: number
  status: number
  reason: string
  evidenceRef: string
  startedAt: number
  createdBy: number
  cancelTotal: number
  cancelSuccess: number
  cancelFailed: number
  lastErrorMsg: string
  liftedAt: number
  liftedBy: number
  liftReason: string
  createTimes: number
  updateTimes: number
}

export type HaltContractTradingReq = {
  tenantId: number
  contractId: number
  reason: string
  evidenceRef: string
}

export type ResumeContractTradingReq = {
  tenantId: number
  haltId: number
  reason: string
}

export type ListTradingHaltsReq = {
  tenantId: number
  contractId?: number
  status?: number
  cursor?: number
  limit?: number
}

export type OptionCorporateActionContract = {
  id: number
  tenantId: number
  actionId: number
  sourceContractId: number
  successorContractId: number
  executionMode: number
  quantityNumerator: string
  quantityDenominator: string
  haltId: number
  status: number
  positionTotal: number
  positionCompleted: number
  positionFailed: number
  lastPositionId: number
  retryCount: number
  lastErrorMsg: string
  createTimes: number
  updateTimes: number
}

export type OptionCorporateAction = {
  id: number
  tenantId: number
  eventNo: string
  externalEventRef: string
  version: number
  underlyingSymbol: string
  actionType: number
  status: number
  announcementTime: number
  exTime: number
  recordTime: number
  effectiveTime: number
  payTime: number
  evidenceRef: string
  evidenceHash: string
  description: string
  createdBy: number
  reviewedBy: number
  reviewReason: string
  reviewedAt: number
  completedAt: number
  lastErrorMsg: string
  createTimes: number
  updateTimes: number
  contracts: OptionCorporateActionContract[]
}

export type CorporateActionContractInput = {
  sourceContractId: number
  successorContractId?: number
  executionMode: number
  quantityNumerator: string
  quantityDenominator: string
}

export type CreateCorporateActionReq = {
  tenantId: number
  eventNo: string
  externalEventRef: string
  underlyingSymbol: string
  actionType: number
  announcementTime: number
  exTime?: number
  recordTime?: number
  effectiveTime: number
  payTime?: number
  evidenceRef: string
  evidenceHash: string
  description: string
  contracts: CorporateActionContractInput[]
}

export type ReviewCorporateActionReq = {
  tenantId: number
  actionId: number
  approve: boolean
  reason: string
}

export type ListCorporateActionsReq = {
  tenantId: number
  underlyingSymbol?: string
  actionType?: number
  status?: number
  cursor?: number
  limit?: number
}

export type OptionCorporateActionPosition = {
  id: number
  tenantId: number
  actionId: number
  actionContractId: number
  sourcePositionId: number
  successorPositionId: number
  userId: number
  accountId: number
  side: number
  sourceQuantity: string
  successorQuantity: string
  sourceAvailableQuantity: string
  successorAvailableQuantity: string
  sourceOpenAvgPrice: string
  successorOpenAvgPrice: string
  sourceEffectiveMultiplier: string
  successorEffectiveMultiplier: string
  costBasisBefore: string
  costBasisAfter: string
  cashDifference: string
  status: number
  retryCount: number
  lastErrorMsg: string
  completedAt: number
  createTimes: number
  updateTimes: number
}

export type OptionContractSeriesExpiry = {
  id: number
  tenantId: number
  seriesId: number
  sequenceNo: number
  cycleCode: string
  listTime: number
  exerciseCutoffTime: number
  expireTime: number
  deliverTime: number
  createTimes: number
}

export type OptionContractSeriesStrikeBand = {
  id: number
  tenantId: number
  seriesId: number
  sequenceNo: number
  lowerStrike: string
  upperStrike: string
  strikeStep: string
  createTimes: number
}

export type OptionContractSeriesDetail = {
  id: number
  tenantId: number
  seriesId: number
  expiryId: number
  optionType: number
  strikePrice: string
  contractCode: string
  contractId: number
  createTimes: number
}

export type OptionContractSeries = {
  id: number
  tenantId: number
  requestKey: string
  seriesCode: string
  version: number
  supersedesId: number
  status: number
  templateContractId: number
  underlyingSymbol: string
  referencePrice: string
  referenceSource: string
  referenceTime: number
  evidenceRef: string
  changeReason: string
  payloadHash: string
  expectedContractCount: number
  generatedContractCount: number
  createdBy: number
  reviewedBy: number
  reviewReason: string
  reviewedAt: number
  generatedAt: number
  launchStatus: number
  launchReviewedBy: number
  launchReviewReason: string
  launchReviewedAt: number
  createTimes: number
  updateTimes: number
  expiries: OptionContractSeriesExpiry[]
  strikeBands: OptionContractSeriesStrikeBand[]
}

export type ContractSeriesExpiryInput = {
  sequenceNo: number
  cycleCode: string
  listTime: number
  exerciseCutoffTime: number
  expireTime: number
  deliverTime: number
}

export type ContractSeriesStrikeBandInput = {
  sequenceNo: number
  lowerStrike: string
  upperStrike: string
  strikeStep: string
}

export type CreateContractSeriesReq = {
  tenantId: number
  requestKey: string
  seriesCode: string
  contractTemplate: CreateContractReq
  referencePrice: string
  referenceSource: string
  referenceTime: number
  evidenceRef: string
  changeReason: string
  expiries: ContractSeriesExpiryInput[]
  strikeBands: ContractSeriesStrikeBandInput[]
}

export type ReviewContractSeriesReq = {
  tenantId: number
  seriesId: number
  approve: boolean
  reason: string
}

export type ReviewContractSeriesLaunchReq = {
  tenantId: number
  seriesId: number
  approve: boolean
  reason: string
}

export type ListContractSeriesReq = {
  tenantId: number
  seriesCode?: string
  status?: number
  cursor?: number
  limit?: number
}

export type ListContractSeriesDetailsReq = {
  tenantId: number
  seriesId: number
  cursor?: number
  limit?: number
}

export type ListCorporateActionPositionsReq = {
  tenantId: number
  actionId?: number
  actionContractId?: number
  status?: number
  cursor?: number
  limit?: number
}

export type OptionRetrySettlementInstructionReq = {
  tenantId: number
  settlementId: number
  instructionId: number
  reason: string
}

export type ListPhysicalDeliveryUnitsReq = {
  cursor?: number
  limit?: number
  tenantId: number
  contractId?: number
  batchId?: number
  userId?: number
  status?: number
}

export type RetryPhysicalDeliveryUnitReq = {
  tenantId: number
  deliveryUnitId: number
  reason: string
}

export type OptionRetryTradeEventReq = {
  tenantId: number
  eventId: number
}

export type OptionRetryExerciseReq = {
  tenantId: number
  exerciseId: number
}

export type GetContractReq = {
  tenantId?: number // 租户ID
  id?: number // 合约ID
  contractCode?: string // 合约编码
}

export type ListContractsReq = {
  cursor?: number // 游标
  limit?: number // 每页条数
  tenantId?: number // 租户ID
  contractCode?: string
  underlyingSymbol?: string
  optionType?: number
  status?: number
  listTimeRange?: TimeRange // 上线时间范围
  expireTimeRange?: TimeRange // 到期时间范围
}

export type UpdateMarketReq = Omit<
  OptionMarket,
  | 'id'
  | 'createTimes'
  | 'updateTimes'
  | 'underlyingSnapshotTime'
  | 'markSnapshotTime'
  | 'greeksSnapshotTime'
> &
  Partial<Pick<OptionMarket, 'underlyingSnapshotTime' | 'markSnapshotTime' | 'greeksSnapshotTime'>>

export type GetMarketReq = {
  tenantId?: number
  contractId: number
}

export type ListMarketSnapshotsReq = {
  cursor?: number
  limit?: number
  tenantId?: number
  contractId: number
  timeRange?: TimeRange
}

export type GetOrderReq = {
  tenantId?: number
  id?: number
  orderNo?: string
}

export type ListOrdersReq = {
  cursor?: number
  limit?: number
  tenantId?: number
  userId?: number
  accountId?: number
  contractId?: number
  underlyingSymbol?: string
  orderNo?: string
  side?: number
  positionEffect?: number
  orderType?: number
  status?: number
  createTimeRange?: TimeRange
}

export type ListAdminComboOrdersReq = {
  cursor?: number
  limit?: number
  tenantId: number
  userId?: number
  accountId?: number
  comboNo?: string
  underlyingSymbol?: string
  status?: number
  createTimeRange?: TimeRange
}

export type GetAdminComboOrderReq = {
  tenantId: number
  id?: number
  comboNo?: string
}

export type ForceCancelComboOrderReq = {
  tenantId: number
  id?: number
  comboNo?: string
  reason: string
}

export type GetTradeReq = {
  tenantId?: number
  id?: number
  tradeNo?: string
}

export type ListTradesReq = {
  cursor?: number
  limit?: number
  tenantId?: number
  contractId?: number
  userId?: number
  tradeNo?: string
  tradeTimeRange?: TimeRange
}

export type GetPositionReq = {
  tenantId?: number
  id?: number
  positionId?: number
}

export type ListPositionsReq = {
  cursor?: number
  limit?: number
  tenantId?: number
  userId?: number
  accountId?: number
  contractId?: number
  status?: number
}

export type GetExerciseReq = {
  tenantId?: number
  id?: number
  exerciseNo?: string
}

export type ListExercisesReq = {
  cursor?: number
  limit?: number
  tenantId?: number
  userId?: number
  accountId?: number
  contractId?: number
  status?: number
  exerciseTimeRange?: TimeRange
}

export type GetSettlementReq = {
  tenantId?: number
  id?: number
  settlementNo?: string
}

export type ListSettlementsReq = {
  cursor?: number
  limit?: number
  tenantId?: number
  contractId?: number
  status?: number
  expireTimeRange?: TimeRange
}

export type ListSettlementPricesReq = {
  cursor?: number
  limit?: number
  tenantId?: number
  contractId?: number
  status?: number
}

export type CreateSettlementPriceCorrectionReq = {
  tenantId: number
  contractId: number
  deliveryPrice: string
  sourceSnapshotIds: string
  reason: string
}

export type ReviewSettlementPriceReq = {
  tenantId: number
  settlementPriceId: number
  approve: boolean
  reason?: string
}

export type GetAccountReq = {
  tenantId?: number
  id?: number
  userId?: number
  accountId?: number
}

export type ListAccountsReq = {
  cursor?: number
  limit?: number
  tenantId?: number
  userId?: number
  accountId?: number
  marginCoin?: string
  status?: number
}

export type GetBillReq = {
  tenantId?: number
  id?: number
  bizNo?: string
}

export type ListBillsReq = {
  cursor?: number
  limit?: number
  tenantId?: number
  userId?: number
  accountId?: number
  refType?: number
  createTimeRange?: TimeRange
}

export type OptionRiskAccount = {
  id: number
  tenantId: number
  userId: number
  accountId: number
  settleCoin: string
  equity: string
  netOptionValue: string
  positionMargin: string
  maintenanceMargin: string
  unrealizedPnl: string
  riskRate: string
  status: number
  portfolioRiskMethod: number
  portfolioRiskConfigId: number
  portfolioRiskConfigVersion: number
  portfolioScenarioLoss: string
  portfolioShortFloor: string
  portfolioConcentrationAddon: string
  portfolioLiquidityAddon: string
  lastCalcTime: number
  createTimes: number
  updateTimes: number
}

export type OptionLiquidation = {
  id: number
  tenantId: number
  liquidationNo: string
  userId: number
  accountId: number
  contractId: number
  positionId: number
  quantity: string
  markPrice: string
  maintenanceMargin: string
  equity: string
  deficitAmount: string
  liquidationFee: string
  status: number
  retryCount: number
  lastErrorMsg: string
  collateralAmount: string
  insuranceFundAmount: string
  remainingDeficit: string
  takeoverPositionId: number
  completedAt: number
  insuranceAttempt: number
  backstopAmount: string
  deficitResolution: number
  createTimes: number
  updateTimes: number
}

export type ListOptionRiskAccountsReq = {
  cursor?: number
  limit?: number
  tenantId?: number
  userId?: number
  accountId?: number
  settleCoin?: string
  status?: number
}

export type ListOptionLiquidationsReq = {
  cursor?: number
  limit?: number
  tenantId?: number
  userId?: number
  accountId?: number
  contractId?: number
  positionId?: number
  status?: number
}

export type OptionRetryLiquidationReq = {
  tenantId: number
  liquidationId: number
}

export class OptionService {
  getOptions(): Promise<RespBase<OptionGroup[]>> {
    return getCoreOptions()
  }

  listContracts(params: ListContractsReq) {
    return apiOptionListContracts(params)
  }

  getContract(params: GetContractReq) {
    return apiOptionGetContract(params)
  }

  createContract(params: CreateContractReq) {
    return apiOptionCreateContract(params)
  }

  updateContract(params: UpdateContractReq) {
    return apiOptionUpdateContract(params)
  }

  forceCancelContractOrders(params: OptionForceCancelContractOrdersReq) {
    return apiOptionForceCancelContractOrders(params)
  }

  releaseUserKillSwitch(params: OptionReleaseUserKillSwitchReq) {
    return apiOptionReleaseUserKillSwitch(params)
  }

  getUserTradingControl(params: AdminGetUserTradingControlReq) {
    return apiOptionGetUserTradingControl(params)
  }

  listTradingControlEvents(params: ListTradingControlEventsReq) {
    return apiOptionListTradingControlEvents(params)
  }

  upsertMMPConfig(params: UpsertMMPConfigReq) {
    return apiOptionUpsertMMPConfig(params)
  }

  resetMMPConfig(params: ResetMMPConfigReq) {
    return apiOptionResetMMPConfig(params)
  }

  listMMPConfigs(params: ListMMPConfigsReq) {
    return apiOptionListMMPConfigs(params)
  }

  createPortfolioRiskConfig(params: CreatePortfolioRiskConfigReq) {
    return apiOptionCreatePortfolioRiskConfig(params)
  }

  reviewPortfolioRiskConfig(params: ReviewPortfolioRiskConfigReq) {
    return apiOptionReviewPortfolioRiskConfig(params)
  }

  listPortfolioRiskConfigs(params: ListPortfolioRiskConfigsReq) {
    return apiOptionListPortfolioRiskConfigs(params)
  }

  createTradeCorrection(params: CreateTradeCorrectionReq) {
    return apiOptionCreateTradeCorrection(params)
  }

  reviewTradeCorrection(params: ReviewTradeCorrectionReq) {
    return apiOptionReviewTradeCorrection(params)
  }

  listTradeCorrections(params: ListTradeCorrectionsReq) {
    return apiOptionListTradeCorrections(params)
  }

  retryAssetInstruction(params: OptionRetryAssetInstructionReq) {
    return apiOptionRetryAssetInstruction(params)
  }

  getOperationsOverview(params: GetOperationsOverviewReq) {
    return apiOptionGetOperationsOverview(params)
  }

  listAssetInstructions(params: ListAssetInstructionsReq) {
    return apiOptionListAssetInstructions(params)
  }

  listReconciliationIssues(params: ListReconciliationIssuesReq) {
    return apiOptionListReconciliationIssues(params)
  }

  createTradingCalendar(params: CreateTradingCalendarReq) {
    return apiOptionCreateTradingCalendar(params)
  }

  reviewTradingCalendar(params: ReviewTradingCalendarReq) {
    return apiOptionReviewTradingCalendar(params)
  }

  listTradingCalendars(params: ListTradingCalendarsReq) {
    return apiOptionListTradingCalendars(params)
  }

  haltContractTrading(params: HaltContractTradingReq) {
    return apiOptionHaltContractTrading(params)
  }

  resumeContractTrading(params: ResumeContractTradingReq) {
    return apiOptionResumeContractTrading(params)
  }

  listTradingHalts(params: ListTradingHaltsReq) {
    return apiOptionListTradingHalts(params)
  }

  createCorporateAction(params: CreateCorporateActionReq) {
    return apiOptionCreateCorporateAction(params)
  }

  reviewCorporateAction(params: ReviewCorporateActionReq) {
    return apiOptionReviewCorporateAction(params)
  }

  listCorporateActions(params: ListCorporateActionsReq) {
    return apiOptionListCorporateActions(params)
  }

  listCorporateActionPositions(params: ListCorporateActionPositionsReq) {
    return apiOptionListCorporateActionPositions(params)
  }

  createContractSeries(params: CreateContractSeriesReq) {
    return apiOptionCreateContractSeries(params)
  }

  reviewContractSeries(params: ReviewContractSeriesReq) {
    return apiOptionReviewContractSeries(params)
  }

  listContractSeries(params: ListContractSeriesReq) {
    return apiOptionListContractSeries(params)
  }

  listContractSeriesDetails(params: ListContractSeriesDetailsReq) {
    return apiOptionListContractSeriesDetails(params)
  }

  reviewContractSeriesLaunch(params: ReviewContractSeriesLaunchReq) {
    return apiOptionReviewContractSeriesLaunch(params)
  }

  retrySettlementInstruction(params: OptionRetrySettlementInstructionReq) {
    return apiOptionRetrySettlementInstruction(params)
  }

  listPhysicalDeliveryUnits(params: ListPhysicalDeliveryUnitsReq) {
    return apiOptionListPhysicalDeliveryUnits(params)
  }

  retryPhysicalDeliveryUnit(params: RetryPhysicalDeliveryUnitReq) {
    return apiOptionRetryPhysicalDeliveryUnit(params)
  }

  retryTradeEvent(params: OptionRetryTradeEventReq) {
    return apiOptionRetryTradeEvent(params)
  }

  listRiskAccounts(params: ListOptionRiskAccountsReq) {
    return apiOptionListRiskAccounts(params)
  }

  listLiquidations(params: ListOptionLiquidationsReq) {
    return apiOptionListLiquidations(params)
  }

  retryLiquidation(params: OptionRetryLiquidationReq) {
    return apiOptionRetryLiquidation(params)
  }

  getMarket(params: GetMarketReq) {
    return apiOptionGetMarket(params)
  }

  updateMarket(params: UpdateMarketReq) {
    return apiOptionUpdateMarket(params)
  }

  listMarketSnapshots(params: ListMarketSnapshotsReq) {
    return apiOptionListMarketSnapshots(params)
  }

  listOrders(params: ListOrdersReq) {
    return apiOptionListOrders(params)
  }

  listAdminComboOrders(params: ListAdminComboOrdersReq) {
    return apiOptionListAdminComboOrders(params)
  }

  getAdminComboOrder(params: GetAdminComboOrderReq) {
    return apiOptionGetAdminComboOrder(params)
  }

  forceCancelComboOrder(params: ForceCancelComboOrderReq) {
    return apiOptionForceCancelComboOrder(params)
  }

  getOrder(params: GetOrderReq) {
    return apiOptionGetOrder(params)
  }

  listTrades(params: ListTradesReq) {
    return apiOptionListTrades(params)
  }

  getTrade(params: GetTradeReq) {
    return apiOptionGetTrade(params)
  }

  listPositions(params: ListPositionsReq) {
    return apiOptionListPositions(params)
  }

  getPosition(params: GetPositionReq) {
    return apiOptionGetPosition(params)
  }

  listExercises(params: ListExercisesReq) {
    return apiOptionListExercises(params)
  }

  getExercise(params: GetExerciseReq) {
    return apiOptionGetExercise(params)
  }

  retryExercise(params: OptionRetryExerciseReq) {
    return apiOptionRetryExercise(params)
  }

  listSettlements(params: ListSettlementsReq) {
    return apiOptionListSettlements(params)
  }

  listSettlementPrices(params: ListSettlementPricesReq) {
    return apiOptionListSettlementPrices(params)
  }

  createSettlementPriceCorrection(params: CreateSettlementPriceCorrectionReq) {
    return apiOptionCreateSettlementPriceCorrection(params)
  }

  reviewSettlementPrice(params: ReviewSettlementPriceReq) {
    return apiOptionReviewSettlementPrice(params)
  }

  getSettlement(params: GetSettlementReq) {
    return apiOptionGetSettlement(params)
  }

  listAccounts(params: ListAccountsReq) {
    return apiOptionListAccounts(params)
  }

  getAccount(params: GetAccountReq) {
    return apiOptionGetAccount(params)
  }

  listBills(params: ListBillsReq) {
    return apiOptionListBills(params)
  }

  getBill(params: GetBillReq) {
    return apiOptionGetBill(params)
  }
}

export const optionService = new OptionService()
