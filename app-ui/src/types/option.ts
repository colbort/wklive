import type { PageReq, TimeRange } from '@/types/api'

export interface OptionContract {
  id: number
  tenantId: number
  contractCode: string
  underlyingSymbol: string
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
  isAutoExercise: number
  status: number
  sort: number
  remark: string
  isDeleted: number
  createTimes: number
  updateTimes: number
}

export interface OptionMarket {
  id: number
  tenantId: number
  contractId: number
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
  snapshotTime: number
  createTimes: number
  updateTimes: number
}

export interface OptionOrder {
  id: number
  tenantId: number
  orderNo: string
  uid: number
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
  source: number
  clientOrderId: string
  reduceOnly: number
  mmp: number
  status: number
  cancelReason: string
  matchTime: number
  cancelTime: number
  createTimes: number
  updateTimes: number
}

export interface OptionTrade {
  id: number
  tenantId: number
  tradeNo: string
  contractId: number
  underlyingSymbol: string
  buyOrderId: number
  buyOrderNo: string
  buyUid: number
  buyAccountId: number
  sellOrderId: number
  sellOrderNo: string
  sellUid: number
  sellAccountId: number
  price: string
  qty: string
  turnover: string
  buyFee: string
  sellFee: string
  feeCoin: string
  makerSide: number
  tradeTime: number
  createTimes: number
}

export interface OptionPosition {
  id: number
  tenantId: number
  uid: number
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
  exerciseableQty: string
  status: number
  lastCalcTime: number
  createTimes: number
  updateTimes: number
}

export interface OptionExercise {
  id: number
  tenantId: number
  exerciseNo: string
  uid: number
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

export interface OptionAccount {
  id: number
  tenantId: number
  uid: number
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

export interface OptionBill {
  id: number
  tenantId: number
  uid: number
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

export interface OptionContractDetail {
  contract: OptionContract
  market: OptionMarket
}

export interface OptionPositionDetail {
  position: OptionPosition
  contract: OptionContract
  market: OptionMarket
}

export interface OptionOrderDetail {
  order: OptionOrder
  contract: OptionContract
}

export interface OptionTradeDetail {
  trade: OptionTrade
  contract: OptionContract
}

export interface OptionExerciseDetail {
  exercise: OptionExercise
  contract: OptionContract
}

export interface AppListContractsReq extends PageReq {
  tenantId?: number
  underlyingSymbol?: string
  optionType?: number
  status?: number
}

export interface AppGetContractDetailReq {
  tenantId?: number
  contractId: number
}

export interface AppPlaceOrderReq {
  tenantId?: number
  accountId: number
  contractId: number
  side: number
  positionEffect: number
  orderType: number
  price: string
  qty: string
  clientOrderId?: string
  reduceOnly?: number
  mmp?: number
}

export interface AppCancelOrderReq {
  tenantId?: number
  accountId: number
  orderId?: number
  orderNo?: string
}

export interface AppGetOrderDetailReq {
  tenantId?: number
  accountId: number
  orderId?: number
  orderNo?: string
}

export interface AppListCurrentOrdersReq extends PageReq {
  tenantId?: number
  accountId: number
  contractId?: number
  side?: number
}

export interface AppListHistoryOrdersReq extends PageReq {
  tenantId?: number
  accountId: number
  contractId?: number
  status?: number
  createTimeRange?: TimeRange
}

export interface AppListTradesReq extends PageReq {
  tenantId?: number
  accountId: number
  contractId?: number
  tradeTimeRange?: TimeRange
}

export interface AppListPositionsReq extends PageReq {
  tenantId?: number
  accountId: number
  status?: number
}

export interface AppGetPositionDetailReq {
  tenantId?: number
  accountId: number
  positionId: number
}

export interface AppExerciseReq {
  tenantId?: number
  accountId: number
  positionId: number
  contractId: number
  exerciseQty: string
}

export interface AppListExercisesReq extends PageReq {
  tenantId?: number
  accountId: number
  contractId?: number
  status?: number
  exerciseTimeRange?: TimeRange
}

export interface AppListAccountsReq {
  tenantId?: number
  accountId?: number
}

export interface AppListBillsReq extends PageReq {
  tenantId?: number
  accountId?: number
  refType?: number
  createTimeRange?: TimeRange
}
