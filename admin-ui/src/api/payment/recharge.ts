import { get, post } from '@/utils/request'
import type {
  RespBase,
  UserRechargeStat,
  RechargeOrder,
  PayNotifyLog,
  GetUserRechargeStatReq,
  ListUserRechargeStatsReq,
  ListRechargeOrdersReq,
  ListRechargeNotifyLogsReq,
} from '@/services'

export function apiUserRechargeStatDetail(
  params: GetUserRechargeStatReq,
): Promise<RespBase<UserRechargeStat>> {
  return get<UserRechargeStat>('/admin/payment/user-recharge-stats', params)
}

export function apiUserRechargeStatList(
  params: ListUserRechargeStatsReq,
): Promise<RespBase<UserRechargeStat[]>> {
  return get<UserRechargeStat[]>('/admin/payment/user-recharge-stats', params)
}

export function apiRechargeOrderList(
  params: ListRechargeOrdersReq,
): Promise<RespBase<RechargeOrder[]>> {
  return get<RechargeOrder[]>('/admin/payment/recharge-orders', params)
}

export function apiRechargeOrderDetail(
  orderNo: string,
  tenantId: number,
): Promise<RespBase<RechargeOrder>> {
  return get<RechargeOrder>(`/admin/payment/recharge-order/${orderNo}`, { tenantId })
}

export function apiCloseRechargeOrder(
  orderNo: string,
  params: { tenantId: number; remark?: string },
): Promise<RespBase> {
  return post(`/admin/payment/recharge-order/${orderNo}/close`, params)
}

export function apiManualSuccessRechargeOrder(
  orderNo: string,
  params: { tenantId: number; thirdTradeNo?: string; payAmount: number; remark?: string },
): Promise<RespBase> {
  return post(`/admin/payment/recharge-order/${orderNo}/manual-success`, params)
}

export function apiRetryRechargeNotify(
  orderNo: string,
  params: { tenantId: number },
): Promise<RespBase> {
  return post(`/admin/payment/recharge-order/${orderNo}/retry-notify`, params)
}

export function apiRechargeNotifyLogList(
  params: ListRechargeNotifyLogsReq,
): Promise<RespBase<PayNotifyLog[]>> {
  return get<PayNotifyLog[]>('/admin/payment/recharge-notify-logs', params)
}

export function apiRechargeNotifyLogDetail(
  id: number,
  tenantId: number,
): Promise<RespBase<PayNotifyLog>> {
  return get<PayNotifyLog>(`/admin/payment/recharge-notify-log/${id}`, { tenantId })
}
