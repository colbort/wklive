import type { RespBase } from '@/services'
import {
  apiAuditWithdrawOrder,
  apiWithdrawNotifyLogDetail,
  apiWithdrawNotifyLogList,
  apiWithdrawOrderDetail,
  apiWithdrawOrderList,
} from '@/api/payment/withdraw'
import type { PayNotifyLog } from './RechargeService'

export type WithdrawOrder = {
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

export type ListWithdrawOrdersReq = {
  tenantId?: number
  userId?: number
  orderNo?: string
  cursor?: string | null
  limit?: number
}

export type ListWithdrawNotifyLogsReq = {
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

export class WithdrawService {
  getWithdrawOrderList(params: ListWithdrawOrdersReq): Promise<RespBase<WithdrawOrder[]>> {
    return apiWithdrawOrderList(params)
  }
  getWithdrawOrderDetail(orderNo: string, tenantId: number): Promise<RespBase<WithdrawOrder>> {
    return apiWithdrawOrderDetail(orderNo, tenantId)
  }
  auditWithdrawOrder(params: {
    tenantId: number
    orderNo: string
    approve: number
    remark?: string
  }): Promise<RespBase> {
    return apiAuditWithdrawOrder(params)
  }
  getWithdrawNotifyLogList(
    params: ListWithdrawNotifyLogsReq,
  ): Promise<RespBase<PayNotifyLog[]>> {
    return apiWithdrawNotifyLogList(params)
  }
  getWithdrawNotifyLogDetail(id: number, tenantId: number): Promise<RespBase<PayNotifyLog>> {
    return apiWithdrawNotifyLogDetail(id, tenantId)
  }
}

export const withdrawService = new WithdrawService()
