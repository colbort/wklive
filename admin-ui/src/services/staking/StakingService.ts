import type { OptionGroup, RespBase } from '@/services'
import {
  apiStakingChangeProductStatus,
  apiStakingCreateProduct,
  apiStakingGetOrder,
  apiStakingGetProduct,
  apiStakingListOrders,
  apiStakingListProducts,
  apiStakingListRedeemLogs,
  apiStakingListRewardLogs,
  apiStakingManualRedeem,
  apiStakingManualReward,
  apiStakingListOperations,
  apiStakingRetryOperation,
  apiStakingListReconciliations,
  apiStakingUpdateProduct,
} from '@/api/staking'
import { getCoreOptions } from '@/stores/core'

export type StakeProduct = {
  id: number // 主键ID
  tenantId: number // 租户ID
  productNo: string // 产品编号
  productName: string // 产品名称
  productType: number // 产品类型
  coinName: string // 质押币种名称
  coinSymbol: string // 质押币种符号
  rewardCoinName: string // 奖励币种名称
  rewardCoinSymbol: string // 奖励币种符号
  apr: string // 年化收益率
  lockDays: number // 锁仓天数
  minAmount: string // 最小质押金额
  maxAmount: string // 最大质押金额
  stepAmount: string // 递增步长
  totalAmount: string // 产品总额度
  stakedAmount: string // 已质押额度
  userLimitAmount: string // 单用户限额
  interestMode: number // 计息模式
  rewardMode: number // 奖励模式
  allowEarlyRedeem: number // 是否允许提前赎回：1是 2否
  earlyRedeemRate: string // 提前赎回费率
  status: number // 状态
  sort: number // 排序
  remark: string // 备注
  createUserId: number // 创建人ID
  updateUserId: number // 更新人ID
  createTimes: number // 创建时间
  updateTimes: number // 更新时间
}

export type StakeOrder = {
  id: number // 主键ID
  tenantId: number // 租户ID
  orderNo: string // 订单号
  userId: number // 用户ID
  productId: number // 产品ID
  productNo: string // 产品编号
  productName: string // 产品名称
  productType: number // 产品类型
  coinName: string // 质押币种名称
  coinSymbol: string // 质押币种符号
  rewardCoinName: string // 奖励币种名称
  rewardCoinSymbol: string // 奖励币种符号
  stakeAmount: string // 质押金额
  apr: string // 年化收益率
  lockDays: number // 锁仓天数
  interestMode: number // 计息模式
  rewardMode: number // 奖励模式
  allowEarlyRedeem: number // 是否允许提前赎回：1是 2否
  earlyRedeemRate: string // 提前赎回费率
  interestDays: number // 已计息天数
  startTimes: number // 开始时间
  endTimes: number // 结束时间
  lastRewardTimes: number // 上次发奖时间
  nextRewardTimes: number // 下次发奖时间
  totalReward: string // 累计奖励
  pendingReward: string // 待发奖励
  redeemAmount: string // 已赎回金额
  redeemFee: string // 赎回手续费
  status: number // 状态
  redeemType: number // 赎回类型
  redeemApplyTimes: number // 赎回申请时间
  redeemTimes: number // 赎回完成时间
  source: number // 来源
  remark: string // 备注
  createUserId: number // 创建人ID
  updateUserId: number // 更新人ID
  createTimes: number // 创建时间
  updateTimes: number // 更新时间
  requestNo?: string // 请求幂等号
  activeOperationNo?: string // 当前资金操作号
  version?: number // 乐观锁版本
}

export type StakeRewardLog = {
  id: number // 主键ID
  tenantId: number // 租户ID
  orderId: number // 订单ID
  orderNo: string // 订单号
  userId: number // 用户ID
  productId: number // 产品ID
  productName: string // 产品名称
  coinSymbol: string // 质押币种符号
  rewardCoinSymbol: string // 奖励币种符号
  rewardAmount: string // 奖励金额
  beforeReward: string // 发放前累计奖励
  afterReward: string // 发放后累计奖励
  rewardType: number // 奖励类型
  rewardStatus: number // 奖励状态
  rewardTimes: number // 奖励时间
  remark: string // 备注
  createUserId: number // 创建人ID
  updateUserId: number // 更新人ID
  createTimes: number // 创建时间
  updateTimes: number // 更新时间
  operationNo?: string // 资金操作号
}

export type StakeRedeemLog = {
  id: number // 主键ID
  tenantId: number // 租户ID
  orderId: number // 订单ID
  orderNo: string // 订单号
  userId: number // 用户ID
  productId: number // 产品ID
  redeemNo: string // 赎回单号
  redeemType: number // 赎回类型
  stakeAmount: string // 原始质押金额
  redeemAmount: string // 赎回金额
  rewardAmount: string // 奖励金额
  feeRate: string // 手续费率
  feeAmount: string // 手续费金额
  redeemStatus: number // 赎回状态
  redeemTimes: number // 赎回时间
  remark: string // 备注
  createUserId: number // 创建人ID
  updateUserId: number // 更新人ID
  createTimes: number // 创建时间
  updateTimes: number // 更新时间
}

export type StakeOperation = {
  id: number
  tenantId: number
  userId: number
  orderId: number
  orderNo: string
  operationNo: string
  requestNo: string
  operationType: number
  principalAmount: string
  rewardAmount: string
  feeAmount: string
  principalStatus: number
  rewardStatus: number
  feeStatus: number
  status: number
  periodEnd: number
  retryCount: number
  nextRetryAt: number
  lastError: string
  operatorUserId: number
  remark: string
  version: number
  createTimes: number
  updateTimes: number
}

export type StakeReconciliation = {
  id: number
  tenantId: number
  reconciliationDate: number
  coinSymbol: string
  activePrincipal: string
  productStaked: string
  positionStaked: string
  assetLocked: string
  rewardLogAmount: string
  rewardPlatformAmount: string
  feeLogAmount: string
  feePlatformAmount: string
  productDiff: string
  positionDiff: string
  lockDiff: string
  rewardDiff: string
  feeDiff: string
  status: number
  detail: string
  createTimes: number
  updateTimes: number
}

export type ProductListReq = {
  cursor?: number // 游标
  limit?: number // 每页条数
  tenantId?: number // 租户ID
  productNo?: string // 产品编号
  productName?: string // 产品名称
  coinSymbol?: string // 币种符号
  productType?: number // 产品类型
  status?: number // 状态
}

export type ProductDetailReq = {
  tenantId?: number // 租户ID
  id: number // 产品ID
}

export type ProductCreateReq = {
  tenantId: number // 租户ID
  productNo: string // 产品编号
  productName: string // 产品名称
  productType: number // 产品类型
  coinName: string // 质押币种名称
  coinSymbol: string // 质押币种符号
  rewardCoinName: string // 奖励币种名称
  rewardCoinSymbol: string // 奖励币种符号
  apr: string // 年化收益率
  lockDays: number // 锁仓天数
  minAmount: string // 最小质押金额
  maxAmount: string // 最大质押金额
  stepAmount: string // 递增步长
  totalAmount: string // 产品总额度
  userLimitAmount: string // 单用户限额
  interestMode: number // 计息模式
  rewardMode: number // 奖励模式
  allowEarlyRedeem: number // 是否允许提前赎回：1是 2否
  earlyRedeemRate: string // 提前赎回费率
  status: number // 状态
  sort: number // 排序
  remark?: string // 备注
}

export type ProductUpdateReq = ProductCreateReq & {
  id: number
}

export type ProductChangeStatusReq = {
  tenantId: number // 租户ID
  id: number // 产品ID
  status: number // 状态
}

export type OrderListReq = {
  cursor?: number // 游标
  limit?: number // 每页条数
  tenantId?: number // 租户ID
  orderNo?: string // 订单号
  userId?: number // 用户ID
  productId?: number // 产品ID
  productName?: string // 产品名称
  coinSymbol?: string // 币种符号
  status?: number // 状态
  redeemType?: number // 赎回类型
  source?: number // 来源
  startTimesBegin?: number // 开始时间起
  startTimesEnd?: number // 开始时间止
  endTimesBegin?: number // 结束时间起
  endTimesEnd?: number // 结束时间止
}

export type OrderDetailReq = {
  tenantId?: number // 租户ID
  id: number // 订单ID
}

export type RewardLogListReq = {
  cursor?: number // 游标
  limit?: number // 每页条数
  tenantId?: number // 租户ID
  orderNo?: string // 订单号
  userId?: number // 用户ID
  productId?: number // 产品ID
  rewardType?: number // 奖励类型
  rewardStatus?: number // 奖励状态
  rewardTimesBegin?: number // 奖励开始时间
  rewardTimesEnd?: number // 奖励结束时间
}

export type RedeemLogListReq = {
  cursor?: number // 游标
  limit?: number // 每页条数
  tenantId?: number // 租户ID
  orderNo?: string // 订单号
  redeemNo?: string // 赎回单号
  userId?: number // 用户ID
  productId?: number // 产品ID
  redeemType?: number // 赎回类型
  redeemStatus?: number // 赎回状态
  redeemTimesBegin?: number // 赎回开始时间
  redeemTimesEnd?: number // 赎回结束时间
}

export type ManualRewardReq = {
  tenantId: number // 租户ID
  orderId: number // 订单ID
  rewardAmount: string // 奖励金额
  rewardType: number // 奖励类型
  remark?: string // 备注
  requestNo: string // 请求幂等号
}

export type ManualRedeemReq = {
  tenantId: number // 租户ID
  orderId: number // 订单ID
  redeemType: number // 赎回类型
  redeemAmount: string // 赎回金额
  rewardAmount: string // 奖励金额
  feeRate: string // 手续费率
  feeAmount: string // 手续费金额
  remark?: string // 备注
  requestNo: string // 请求幂等号
}

export type OperationListReq = {
  cursor?: number
  limit?: number
  tenantId?: number
  operationNo?: string
  orderNo?: string
  userId?: number
  operationType?: number
  status?: number
}

export type OperationRetryReq = {
  tenantId: number
  id: number
  reason: string
}

export type ReconciliationListReq = {
  cursor?: number
  limit?: number
  tenantId?: number
  reconciliationDate?: number
  coinSymbol?: string
  status?: number
}

export class StakingService {
  getOptions(): Promise<RespBase<OptionGroup[]>> {
    return getCoreOptions()
  }

  listProducts(params: ProductListReq) {
    return apiStakingListProducts(params)
  }

  getProduct(params: ProductDetailReq) {
    return apiStakingGetProduct(params)
  }

  createProduct(params: ProductCreateReq) {
    return apiStakingCreateProduct(params)
  }

  updateProduct(params: ProductUpdateReq) {
    return apiStakingUpdateProduct(params)
  }

  changeProductStatus(params: ProductChangeStatusReq) {
    return apiStakingChangeProductStatus(params)
  }

  listOrders(params: OrderListReq) {
    return apiStakingListOrders(params)
  }

  getOrder(params: OrderDetailReq) {
    return apiStakingGetOrder(params)
  }

  listRewardLogs(params: RewardLogListReq) {
    return apiStakingListRewardLogs(params)
  }

  listRedeemLogs(params: RedeemLogListReq) {
    return apiStakingListRedeemLogs(params)
  }

  manualReward(params: ManualRewardReq) {
    return apiStakingManualReward(params)
  }

  manualRedeem(params: ManualRedeemReq) {
    return apiStakingManualRedeem(params)
  }

  listOperations(params: OperationListReq) {
    return apiStakingListOperations(params)
  }

  retryOperation(params: OperationRetryReq) {
    return apiStakingRetryOperation(params)
  }

  listReconciliations(params: ReconciliationListReq) {
    return apiStakingListReconciliations(params)
  }
}

export const stakingService = new StakingService()
