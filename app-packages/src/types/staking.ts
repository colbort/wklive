import type { PageReq } from './api'

export interface StakeProduct {
  id: number
  tenantId: number
  productNo: string
  productName: string
  productType: number
  coinName: string
  coinSymbol: string
  rewardCoinName: string
  rewardCoinSymbol: string
  apr: string
  lockDays: number
  minAmount: string
  maxAmount: string
  stepAmount: string
  totalAmount: string
  stakedAmount: string
  userLimitAmount: string
  interestMode: number
  rewardMode: number
  allowEarlyRedeem: number // 是否允许提前赎回：1是 2否
  earlyRedeemRate: string
  status: number
  sort: number
  remark: string
  createUserId: number
  updateUserId: number
  createTimes: number
  updateTimes: number
}

export interface StakeOrder {
  id: number
  tenantId: number
  orderNo: string
  userId: number
  productId: number
  productNo: string
  productName: string
  productType: number
  coinName: string
  coinSymbol: string
  rewardCoinName: string
  rewardCoinSymbol: string
  stakeAmount: string
  apr: string
  lockDays: number
  interestMode: number
  rewardMode: number
  allowEarlyRedeem: number // 是否允许提前赎回：1是 2否
  earlyRedeemRate: string
  interestDays: number
  startTimes: number
  endTimes: number
  lastRewardTimes: number
  nextRewardTimes: number
  totalReward: string
  pendingReward: string
  redeemAmount: string
  redeemFee: string
  status: number
  redeemType: number
  redeemApplyTimes: number
  redeemTimes: number
  source: number
  remark: string
  createUserId: number
  updateUserId: number
  createTimes: number
  updateTimes: number
}

export interface StakeRewardLog {
  id: number
  tenantId: number
  orderId: number
  orderNo: string
  userId: number
  productId: number
  productName: string
  coinSymbol: string
  rewardCoinSymbol: string
  rewardAmount: string
  beforeReward: string
  afterReward: string
  rewardType: number
  rewardStatus: number
  rewardTimes: number
  remark: string
  createUserId: number
  updateUserId: number
  createTimes: number
  updateTimes: number
}

export interface StakeRedeemLog {
  id: number
  tenantId: number
  orderId: number
  orderNo: string
  userId: number
  productId: number
  redeemNo: string
  redeemType: number
  stakeAmount: string
  redeemAmount: string
  rewardAmount: string
  feeRate: string
  feeAmount: string
  redeemStatus: number
  redeemTimes: number
  remark: string
  createUserId: number
  updateUserId: number
  createTimes: number
  updateTimes: number
}

export interface ProductListReq extends PageReq {
  productType?: number
  coinSymbol?: string
}

export interface ProductDetailReq {
  id: number
}

export interface CreateOrderReq {
  productId: number
  stakeAmount: string
  source: number
  remark?: string
  /** 客户端生成的幂等号，同一业务重试必须复用 */
  requestNo: string
}

export interface MyOrderListReq extends PageReq {
  status?: number
  redeemType?: number
}

export interface MyOrderDetailReq {
  id: number
}

export interface MyRewardLogListReq extends PageReq {
  orderId?: number
  rewardType?: number
}

export interface RedeemReq {
  orderId: number
  redeemType: number
  remark?: string
  /** 客户端生成的幂等号，同一业务重试必须复用 */
  requestNo: string
}

export interface MyRedeemLogListReq extends PageReq {
  orderId?: number
}
