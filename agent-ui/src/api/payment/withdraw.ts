import { get, post } from '@/utils/request'
import type {
  RespBase,
  WithdrawOrder,
  PayNotifyLog,
  ListWithdrawOrdersReq,
  ListWithdrawNotifyLogsReq,
} from '@/services'

export function apiWithdrawOrderList(
  params: ListWithdrawOrdersReq,
): Promise<RespBase<WithdrawOrder[]>> {
  return get<WithdrawOrder[]>('/admin/payment/withdraw-orders', params)
}

export function apiWithdrawOrderDetail(
  orderNo: string,
  tenantId: number,
): Promise<RespBase<WithdrawOrder>> {
  return get<WithdrawOrder>(`/admin/payment/withdraw-order/${orderNo}`, { tenantId })
}

export function apiAuditWithdrawOrder(params: {
  tenantId: number
  orderNo: string
  approve: number
  remark?: string
}): Promise<RespBase> {
  return post(`/admin/payment/withdraw-order/${params.orderNo}/audit`, params)
}

export function apiWithdrawNotifyLogList(
  params: ListWithdrawNotifyLogsReq,
): Promise<RespBase<PayNotifyLog[]>> {
  return get<PayNotifyLog[]>('/admin/payment/withdraw-notify-logs', params)
}

export function apiWithdrawNotifyLogDetail(
  id: number,
  tenantId: number,
): Promise<RespBase<PayNotifyLog>> {
  return get<PayNotifyLog>(`/admin/payment/withdraw-notify-log/${id}`, { tenantId })
}
