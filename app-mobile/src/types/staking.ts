import type { PageReq } from '@/types/api'

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
  allowEarlyRedeem: number
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
  allowEarlyRedeem: number
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

export interface AppProductListReq extends PageReq {
  productType?: number
  coinSymbol?: string
}

export interface AppProductDetailReq {
  id: number
}

export interface AppCreateOrderReq {
  productId: number
  stakeAmount: string
  source: number
  remark?: string
}

export interface AppMyOrderListReq extends PageReq {
  status?: number
  redeemType?: number
}

export interface AppMyOrderDetailReq {
  id: number
}

export interface AppMyRewardLogListReq extends PageReq {
  orderId?: number
  rewardType?: number
}

export interface AppRedeemReq {
  orderId: number
  redeemType: number
  remark?: string
}

export interface AppMyRedeemLogListReq extends PageReq {
  orderId?: number
}
