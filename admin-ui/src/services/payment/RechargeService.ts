import type { RespBase } from '@/services'
import {
  apiCloseRechargeOrder,
  apiManualSuccessRechargeOrder,
  apiRechargeNotifyLogDetail,
  apiRechargeNotifyLogList,
  apiRechargeOrderDetail,
  apiRechargeOrderList,
  apiRetryRechargeNotify,
  apiUserRechargeStatDetail,
  apiUserRechargeStatList,
} from '@/api/payment/recharge'

export type UserRechargeStat = {
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

export type RechargeOrder = {
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

export type PayNotifyLog = {
  id: number
  tenantId: number
  orderId: number
  orderNo: string
  platformId: number
  channelId: number
  notifyStatus: number
  notifyBody: string
  signResult: number
  processResult: string
  errorMessage: string
  notifyTime: number
  createTimes: number
}

export type GetUserRechargeStatReq = {
  tenantId: number
  userId: number
}

export type ListUserRechargeStatsReq = {
  tenantId?: number
  userId?: number
  successTotalAmountMin?: number
  successTotalAmountMax?: number
  cursor?: string | null
  limit?: number
}

export type ListRechargeOrdersReq = {
  tenantId?: number
  userId?: number
  platformId?: number
  productId?: number
  accountId?: number
  channelId?: number
  orderNo?: string
  bizOrderNo?: string
  thirdTradeNo?: string
  status?: number
  createTimeStart?: number
  createTimeEnd?: number
  cursor?: string | null
  limit?: number
}

export type ListRechargeNotifyLogsReq = {
  tenantId?: number
  orderNo?: string
  orderId?: number
  platformId?: number
  channelId?: number
  notifyStatus?: number
  signResult?: number
  createTimeStart?: number
  createTimeEnd?: number
  cursor?: string | null
  limit?: number
}

export class RechargeService {
  getUserRechargeStat(params: GetUserRechargeStatReq): Promise<RespBase<UserRechargeStat>> {
    return apiUserRechargeStatDetail(params)
  }
  getUserRechargeStatList(
    params: ListUserRechargeStatsReq,
  ): Promise<RespBase<UserRechargeStat[]>> {
    return apiUserRechargeStatList(params)
  }
  getRechargeOrderList(params: ListRechargeOrdersReq): Promise<RespBase<RechargeOrder[]>> {
    return apiRechargeOrderList(params)
  }
  getRechargeOrderDetail(orderNo: string, tenantId: number): Promise<RespBase<RechargeOrder>> {
    return apiRechargeOrderDetail(orderNo, tenantId)
  }
  closeRechargeOrder(orderNo: string, tenantId: number, remark?: string): Promise<RespBase> {
    return apiCloseRechargeOrder(orderNo, { tenantId, remark })
  }
  manualSuccessRechargeOrder(
    orderNo: string,
    params: { tenantId: number; thirdTradeNo?: string; payAmount: number; remark?: string },
  ): Promise<RespBase> {
    return apiManualSuccessRechargeOrder(orderNo, params)
  }
  retryRechargeNotify(orderNo: string, tenantId: number): Promise<RespBase> {
    return apiRetryRechargeNotify(orderNo, { tenantId })
  }
  getRechargeNotifyLogList(
    params: ListRechargeNotifyLogsReq,
  ): Promise<RespBase<PayNotifyLog[]>> {
    return apiRechargeNotifyLogList(params)
  }
  getRechargeNotifyLogDetail(id: number, tenantId: number): Promise<RespBase<PayNotifyLog>> {
    return apiRechargeNotifyLogDetail(id, tenantId)
  }
}

export const rechargeService = new RechargeService()
