import type { PageReq } from '@/types/api'

export interface UserRechargeStat {
  id: number
  tenantId: number
  userId: number
  successOrderCount: number
  successTotalAmount: number
  todaySuccessAmount: number
  todaySuccessCount: number
  firstSuccessTime: number
  lastSuccessTime: number
  createTimes: number
  updateTimes: number
}

export interface AvailableRechargeChannel {
  channelId: number
  channelCode: string
  channelName: string
  displayName: string
  icon: string
  currency: string
  singleMinAmount: number
  singleMaxAmount: number
  feeType: number
  feeRate: string
  feeFixedAmount: number
  platformId: number
  productId: number
  accountId: number
  userSuccessTotalAmount: number
}

export interface RechargeOrder {
  id: number
  tenantId: number
  userId: number
  orderNo: string
  bizOrderNo: string
  platformId: number
  productId: number
  accountId: number
  channelId: number
  currency: string
  orderAmount: number
  payAmount: number
  feeAmount: number
  subject: string
  body: string
  clientType: number
  clientIp: string
  status: number
  thirdTradeNo: string
  thirdOrderNo: string
  payUrl: string
  qrContent: string
  requestData: string
  responseData: string
  notifyData: string
  expireTime: number
  paidTime: number
  notifyTime: number
  closeTime: number
  remark: string
  createTimes: number
  updateTimes: number
}

export interface WithdrawOrder {
  id: number
  tenantId: number
  userId: number
  orderNo: string
  bizOrderNo: string
  currency: string
  amount: number
  feeAmount: number
  actualAmount: number
  clientType: number
  clientIp: string
  status: number
  thirdTradeNo: string
  thirdOrderNo: string
  requestData: string
  responseData: string
  notifyData: string
  processTime: number
  notifyTime: number
  closeTime: number
  remark: string
  createTimes: number
  updateTimes: number
}

export interface GetMyRechargeStatReq {
  tenantId?: number
}

export interface ListAvailableRechargeChannelsReq {
  tenantId?: number
  rechargeAmount: number
  currency: string
  clientType: number
}

export interface CreateRechargeOrderReq {
  tenantId?: number
  channelId: number
  rechargeAmount: number
  currency: string
  subject?: string
  body?: string
  clientType: number
  clientIp?: string
  bizOrderNo?: string
}

export interface GetMyRechargeOrderReq {
  tenantId?: number
  orderNo: string
}

export interface ListMyRechargeOrdersReq extends PageReq {
  tenantId?: number
  status?: number
  orderNo?: string
  createTimeStart?: number
  createTimeEnd?: number
}

export interface CancelMyRechargeOrderReq {
  tenantId?: number
  orderNo: string
}

export interface QueryMyRechargeOrderStatusReq {
  tenantId?: number
  orderNo: string
}

export interface CreateWithdrawOrderReq {
  tenantId?: number
  userId?: number
  amount: number
  currency: string
  address: string
  bankId: number
  remark: string
}

export interface ListMyWithdrawOrdersReq extends PageReq {
  tenantId?: number
  userId?: number
  status?: number
}

export interface GetMyWithdrawOrderReq {
  orderNo: string
}
