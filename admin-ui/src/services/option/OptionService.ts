import type { RespBase } from '@/services'
import {
  apiOptionCreateContract,
  apiOptionGetAccount,
  apiOptionGetBill,
  apiOptionGetContract,
  apiOptionGetExercise,
  apiOptionGetMarket,
  apiOptionGetOrder,
  apiOptionGetPosition,
  apiOptionGetSettlement,
  apiOptionGetTrade,
  apiOptionListAccounts,
  apiOptionListBills,
  apiOptionListContracts,
  apiOptionListExercises,
  apiOptionListMarketSnapshots,
  apiOptionListOrders,
  apiOptionListPositions,
  apiOptionListSettlements,
  apiOptionListTrades,
  apiOptionUpdateContract,
  apiOptionUpdateMarket,
} from '@/api/option'

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
  sort: number // 排序
  remark: string // 备注
  isDeleted: number // 是否删除
  createTimes: number // 创建时间
  updateTimes: number // 更新时间
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
  createTimes: number
}

export type OptionOrder = {
  id: number // 主键ID
  tenantId: number // 租户ID
  orderNo: string // 订单号
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
  createTimes: number // 创建时间
}

export type OptionPosition = {
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

export type OptionExercise = {
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
  isItm: number
  exerciseResult: number
  status: number
  remark: string
  createTimes: number
  updateTimes: number
}

export type OptionAccount = {
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

export type OptionBill = {
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
}

export type OptionSettlementDetail = {
  settlement: OptionSettlement
  contract: OptionContract
}

export type CreateContractReq = Omit<OptionContract, 'id' | 'isDeleted' | 'createTimes' | 'updateTimes'>

export type UpdateContractReq = Omit<OptionContract, 'createTimes' | 'updateTimes'>

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

export type UpdateMarketReq = Omit<OptionMarket, 'id' | 'createTimes' | 'updateTimes'>

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
  uid?: number
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
  uid?: number
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
  uid?: number
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
  uid?: number
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

export type GetAccountReq = {
  tenantId?: number
  id?: number
  uid?: number
  accountId?: number
}

export type ListAccountsReq = {
  cursor?: number
  limit?: number
  tenantId?: number
  uid?: number
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
  uid?: number
  accountId?: number
  refType?: number
  createTimeRange?: TimeRange
}

export class OptionService {
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

  listSettlements(params: ListSettlementsReq) {
    return apiOptionListSettlements(params)
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
