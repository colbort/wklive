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
  apiStakingUpdateProduct,
} from '@/api/staking'

export type StakeProduct = {
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

export type StakeOrder = {
  id: number
  tenantId: number
  orderNo: string
  uid: number
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

export type StakeRewardLog = {
  id: number
  tenantId: number
  orderId: number
  orderNo: string
  uid: number
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

export type StakeRedeemLog = {
  id: number
  tenantId: number
  orderId: number
  orderNo: string
  uid: number
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

export type AdminProductListReq = {
  cursor?: number
  limit?: number
  tenantId?: number
  productNo?: string
  productName?: string
  coinSymbol?: string
  productType?: number
  status?: number
}

export type AdminProductDetailReq = {
  tenantId?: number
  id: number
}

export type AdminProductCreateReq = {
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
  userLimitAmount: string
  interestMode: number
  rewardMode: number
  allowEarlyRedeem: number
  earlyRedeemRate: string
  status: number
  sort: number
  remark?: string
  operatorUid: number
}

export type AdminProductUpdateReq = AdminProductCreateReq & {
  id: number
}

export type AdminProductChangeStatusReq = {
  tenantId: number
  id: number
  status: number
  operatorUid: number
}

export type AdminOrderListReq = {
  cursor?: number
  limit?: number
  tenantId?: number
  orderNo?: string
  uid?: number
  productId?: number
  productName?: string
  coinSymbol?: string
  status?: number
  redeemType?: number
  source?: number
  startTimesBegin?: number
  startTimesEnd?: number
  endTimesBegin?: number
  endTimesEnd?: number
}

export type AdminOrderDetailReq = {
  tenantId?: number
  id: number
}

export type AdminRewardLogListReq = {
  cursor?: number
  limit?: number
  tenantId?: number
  orderNo?: string
  uid?: number
  productId?: number
  rewardType?: number
  rewardStatus?: number
  rewardTimesBegin?: number
  rewardTimesEnd?: number
}

export type AdminRedeemLogListReq = {
  cursor?: number
  limit?: number
  tenantId?: number
  orderNo?: string
  redeemNo?: string
  uid?: number
  productId?: number
  redeemType?: number
  redeemStatus?: number
  redeemTimesBegin?: number
  redeemTimesEnd?: number
}

export type AdminManualRewardReq = {
  tenantId: number
  orderId: number
  rewardAmount: string
  rewardType: number
  remark?: string
  operatorUid: number
}

export type AdminManualRedeemReq = {
  tenantId: number
  orderId: number
  redeemType: number
  redeemAmount: string
  rewardAmount: string
  feeRate: string
  feeAmount: string
  remark?: string
  operatorUid: number
}

export class StakingService {
  listProducts(params: AdminProductListReq) {
    return apiStakingListProducts(params)
  }

  getProduct(params: AdminProductDetailReq) {
    return apiStakingGetProduct(params)
  }

  createProduct(params: AdminProductCreateReq) {
    return apiStakingCreateProduct(params)
  }

  updateProduct(params: AdminProductUpdateReq) {
    return apiStakingUpdateProduct(params)
  }

  changeProductStatus(params: AdminProductChangeStatusReq) {
    return apiStakingChangeProductStatus(params)
  }

  listOrders(params: AdminOrderListReq) {
    return apiStakingListOrders(params)
  }

  getOrder(params: AdminOrderDetailReq) {
    return apiStakingGetOrder(params)
  }

  listRewardLogs(params: AdminRewardLogListReq) {
    return apiStakingListRewardLogs(params)
  }

  listRedeemLogs(params: AdminRedeemLogListReq) {
    return apiStakingListRedeemLogs(params)
  }

  manualReward(params: AdminManualRewardReq) {
    return apiStakingManualReward(params)
  }

  manualRedeem(params: AdminManualRedeemReq) {
    return apiStakingManualRedeem(params)
  }
}

export const stakingService = new StakingService()
