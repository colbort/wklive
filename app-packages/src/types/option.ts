import type { PageReq, TimeRange } from "./api";

export interface OptionContract {
  id: number;
  tenantId: number;
  contractCode: string;
  underlyingSymbol: string;
  underlyingCoin: string;
  settleCoin: string;
  quoteCoin: string;
  optionType: number;
  exerciseStyle: number;
  settlementType: number;
  strikePrice: string;
  contractUnit: string;
  minOrderQty: string;
  maxOrderQty: string;
  priceTick: string;
  qtyStep: string;
  multiplier: string;
  listTime: number;
  expireTime: number;
  deliverTime: number;
  exerciseCutoffTime: number;
  autoExerciseThreshold: string;
  maxUserLongQty: string;
  maxUserShortQty: string;
  maxOpenInterest: string;
  orderPriceBandRatio: string;
  circuitBreakerRatio: string;
  physicalDeliveryPolicy: number;
  physicalDeliveryCureSeconds: number;
  tradingCalendarCode: string;
  isAutoExercise: number; // 是否自动行权：1是 2否
  status: number;
  sort: number;
  remark: string;
  isDeleted: number; // 是否删除：1是 2否
  createTimes: number;
  updateTimes: number;
}

export interface OptionMarket {
  id: number;
  tenantId: number;
  contractId: number;
  underlyingPrice: string;
  markPrice: string;
  lastPrice: string;
  bidPrice: string;
  askPrice: string;
  theoreticalPrice: string;
  intrinsicValue: string;
  timeValue: string;
  iv: string;
  delta: string;
  gamma: string;
  theta: string;
  vega: string;
  rho: string;
  riskFreeRate: string;
  pricingModel: string;
  snapshotTime: number;
  underlyingSnapshotTime: number;
  markSnapshotTime: number;
  greeksSnapshotTime: number;
  createTimes: number;
  updateTimes: number;
}

export interface OptionOrder {
  id: number;
  tenantId: number;
  orderNo: string;
  userId: number;
  accountId: number;
  contractId: number;
  underlyingSymbol: string;
  side: number;
  positionEffect: number;
  orderType: number;
  price: string;
  qty: string;
  filledQty: string;
  unfilledQty: string;
  avgPrice: string;
  turnover: string;
  fee: string;
  feeCoin: string;
  marginAmount: string;
  marginCoin: string;
  source: number;
  clientOrderId: string;
  reduceOnly: number; // 是否只减仓：1是 2否
  mmp: number; // 是否做市商保护单：1是 2否
  mmpGroup: string;
  status: number;
  cancelReason: string;
  matchTime: number;
  cancelTime: number;
  createTimes: number;
  updateTimes: number;
}

export interface OptionTrade {
  id: number;
  tenantId: number;
  tradeNo: string;
  contractId: number;
  underlyingSymbol: string;
  buyOrderId: number;
  buyOrderNo: string;
  buyUid: number;
  buyAccountId: number;
  sellOrderId: number;
  sellOrderNo: string;
  sellUid: number;
  sellAccountId: number;
  price: string;
  qty: string;
  turnover: string;
  buyFee: string;
  sellFee: string;
  feeCoin: string;
  makerSide: number;
  matchSequence: number;
  comboMatchNo: string;
  comboLegNo: number;
  tradeTime: number;
  createTimes: number;
}

export interface OptionPosition {
  id: number;
  tenantId: number;
  userId: number;
  accountId: number;
  contractId: number;
  underlyingSymbol: string;
  side: number;
  positionQty: string;
  availableQty: string;
  frozenQty: string;
  openAvgPrice: string;
  markPrice: string;
  positionValue: string;
  marginAmount: string;
  maintenanceMargin: string;
  unrealizedPnl: string;
  realizedPnl: string;
  tradeRealizedPnl: string;
  settlementRealizedPnl: string;
  feePaid: string;
  totalReturn: string;
  exerciseableQty: string;
  status: number;
  lastCalcTime: number;
  createTimes: number;
  updateTimes: number;
}

export interface OptionExercise {
  id: number;
  tenantId: number;
  exerciseNo: string;
  userId: number;
  accountId: number;
  contractId: number;
  positionId: number;
  exerciseType: number;
  exerciseQty: string;
  strikePrice: string;
  settlementPrice: string;
  exerciseAmount: string;
  profitAmount: string;
  fee: string;
  feeCoin: string;
  status: number;
  remark: string;
  exerciseTime: number;
  finishTime: number;
  createTimes: number;
  updateTimes: number;
}

export interface OptionExerciseInstruction {
  id: number;
  tenantId: number;
  userId: number;
  accountId: number;
  contractId: number;
  positionId: number;
  clientInstructionId: string;
  instructionType: number;
  version: number;
  status: number;
  supersedesId: number;
  cutoffTime: number;
  createTimes: number;
  updateTimes: number;
}

export interface OptionUserTradingControl {
  tenantId: number;
  userId: number;
  killSwitch: number;
  reason: string;
  activatedAt: number;
  releasedAt: number;
  updateTimes: number;
}

export interface OptionMMPConfig {
  id: number;
  tenantId: number;
  userId: number;
  contractId: number;
  groupCode: string;
  enabled: number;
  qtyThreshold: string;
  tradeCountThreshold: number;
  lossThreshold: string;
  windowSeconds: number;
  cooldownSeconds: number;
  status: number;
  windowStart: number;
  accumulatedQty: string;
  tradeCount: number;
  accumulatedLoss: string;
  triggeredAt: number;
  cooldownUntil: number;
  triggerReason: string;
  lastErrorMsg: string;
  createdBy: number;
  updatedBy: number;
  createTimes: number;
  updateTimes: number;
}

export interface OptionAccount {
  id: number;
  tenantId: number;
  userId: number;
  accountId: number;
  marginCoin: string;
  balance: string;
  availableBalance: string;
  frozenBalance: string;
  positionMargin: string;
  orderMargin: string;
  unrealizedPnl: string;
  realizedPnl: string;
  riskRate: string;
  status: number;
  createTimes: number;
  updateTimes: number;
}

export interface OptionBill {
  id: number;
  tenantId: number;
  userId: number;
  accountId: number;
  bizNo: string;
  refType: number;
  refId: number;
  coin: string;
  changeAmount: string;
  balanceBefore: string;
  balanceAfter: string;
  remark: string;
  createTimes: number;
}

export interface OptionContractDetail {
  contract: OptionContract;
  market: OptionMarket;
}

export interface OptionMarketStatistics {
  contractId: number;
  volume24h: string;
  turnover24h: string;
  tradeCount24h: number;
  openInterest: string;
  longOpenInterest: string;
  shortOpenInterest: string;
  oiBalanced: boolean;
  statisticsWindowStart: number;
  statisticsAsOf: number;
  positionAsOf: number;
  source: "OPTION_TRADES_AND_SETTLED_POSITIONS";
  openInterestMethod: "MAX_LONG_SHORT";
}

export interface OptionChainLeg {
  contract: OptionContract;
  market: OptionMarket;
  statistics: OptionMarketStatistics;
}

export interface OptionChainRow {
  strikePrice: string;
  call?: OptionChainLeg;
  put?: OptionChainLeg;
}

export interface OptionOrderBookLevel {
  price: string;
  qty: string;
  orderCount: number;
}

export interface OptionOrderBook {
  contractId: number;
  lastMatchSequence: number;
  generatedAt: number;
  source: "OPTION_ACTIVE_LIMIT_ORDERS";
  bids: OptionOrderBookLevel[];
  asks: OptionOrderBookLevel[];
}

export interface ListOptionChainReq {
  underlyingSymbol: string;
  expireTime: number;
  status?: number;
}

export interface GetOrderBookReq {
  contractId: number;
  depthLimit?: number;
}

export interface OptionComboOrder {
  id: number;
  tenantId: number;
  comboNo: string;
  userId: number;
  accountId: number;
  clientComboId: string;
  strategyKey: string;
  inverseStrategyKey: string;
  underlyingSymbol: string;
  expireTime: number;
  settleCoin: string;
  quoteCoin: string;
  orderType: number;
  netPrice: string;
  qty: string;
  filledQty: string;
  unfilledQty: string;
  status: number;
  payloadHash: string;
  cancelReason: string;
  cancelTime: number;
  createTimes: number;
  updateTimes: number;
}

export interface OptionComboOrderLeg {
  id: number;
  tenantId: number;
  comboOrderId: number;
  legNo: number;
  contractId: number;
  side: number;
  positionEffect: number;
  ratio: number;
  price: string;
  qty: string;
  filledQty: string;
  unfilledQty: string;
  childOrderId: number;
  createTimes: number;
  updateTimes: number;
}

export interface OptionComboOrderDetail {
  comboOrder: OptionComboOrder;
  legs: OptionComboOrderLeg[];
}

export interface OptionComboOrderLegInput {
  contractId: number;
  side: number;
  positionEffect: number;
  ratio: number;
  price: string;
}

export interface OptionPlaceComboOrderReq {
  accountId: number;
  clientComboId: string;
  orderType: number;
  qty: string;
  netPrice: string;
  legs: OptionComboOrderLegInput[];
}

export interface OptionCancelComboOrderReq {
  accountId: number;
  comboOrderId?: number;
  comboNo?: string;
}

export interface OptionGetComboOrderReq {
  accountId: number;
  comboOrderId?: number;
  comboNo?: string;
}

export interface OptionListComboOrdersReq extends PageReq {
  accountId: number;
  status?: number;
}

export interface OptionPositionDetail {
  position: OptionPosition;
  contract: OptionContract;
  market: OptionMarket;
}

export interface OptionOrderDetail {
  order: OptionOrder;
  contract: OptionContract;
}

export interface OptionTradeDetail {
  trade: OptionTrade;
  contract: OptionContract;
}

export interface OptionExerciseDetail {
  exercise: OptionExercise;
  contract: OptionContract;
}

export interface ListContractsReq extends PageReq {
  underlyingSymbol?: string;
  optionType?: number;
  status?: number;
}

export interface GetContractDetailReq {
  contractId: number;
}

export interface OptionPlaceOrderReq {
  accountId: number;
  contractId: number;
  side: number;
  positionEffect: number;
  orderType: number;
  price: string;
  qty: string;
  clientOrderId?: string;
  reduceOnly?: number; // 是否只减仓：1是 2否
  protectionPrice?: string;
  maxTurnover?: string;
  mmp?: number;
  mmpGroup?: string;
}

export interface OptionCancelOrderReq {
  accountId: number;
  orderId?: number;
  orderNo?: string;
}

export interface OptionGetOrderDetailReq {
  accountId: number;
  orderId?: number;
  orderNo?: string;
}

export interface ListCurrentOrdersReq extends PageReq {
  accountId: number;
  contractId?: number;
  side?: number;
}

export interface ListHistoryOrdersReq extends PageReq {
  accountId: number;
  contractId?: number;
  status?: number;
  createTimeRange?: TimeRange;
}

export interface ListTradesReq extends PageReq {
  accountId: number;
  contractId?: number;
  tradeTimeRange?: TimeRange;
}

export interface ListPositionsReq extends PageReq {
  accountId: number;
  status?: number;
}

export interface GetPositionDetailReq {
  accountId: number;
  positionId: number;
}

export interface ExerciseReq {
  accountId: number;
  positionId: number;
  contractId: number;
  exerciseQty: string;
  clientExerciseId: string;
}

export interface SetExerciseInstructionReq {
  accountId: number;
  positionId: number;
  contractId: number;
  instructionType: number;
  clientInstructionId: string;
}

export interface GetExerciseInstructionReq {
  accountId: number;
  positionId: number;
}

export interface ActivateKillSwitchReq {
  reason?: string;
}

export interface GetMMPConfigReq {
  contractId: number;
  groupCode: string;
}

export interface ListExercisesReq extends PageReq {
  accountId: number;
  contractId?: number;
  status?: number;
  exerciseTimeRange?: TimeRange;
}

export interface ListAccountsReq {
  accountId?: number;
}

export interface ListBillsReq extends PageReq {
  accountId?: number;
  refType?: number;
  createTimeRange?: TimeRange;
}
