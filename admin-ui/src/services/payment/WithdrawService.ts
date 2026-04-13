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
  id: number // 提现订单ID
  tenantId: number // 租户ID
  userId: number // 用户ID
  orderNo: string // 平台订单号
  bizOrderNo: string // 业务订单号
  platformId: number // 平台ID
  productId: number // 产品ID
  accountId: number // 账号ID
  channelId: number // 通道ID
  currency: string // 币种
  amount: number // 订单金额，单位分
  feeAmount: number // 手续费金额，单位分
  actualAmount: number // 实际到账金额，单位分
  clientType: number // 客户端类型：1APP 2H5 3WEB
  clientIp: string // 客户端IP
  status: number // 状态：1待处理 2处理中 3成功 4失败 5已关闭
  thirdTradeNo: string // 三方交易号
  thirdOrderNo: string // 三方订单号
  requestData: string // 请求快照(JSON)
  responseData: string // 响应快照(JSON)
  notifyData: string // 回调数据(JSON)
  processTime: number // 处理时间
  notifyTime: number // 回调时间
  closeTime: number // 关闭时间
  remark: string // 备注
  createTimes: number // 创建时间
  updateTimes: number // 更新时间
}

export type ListWithdrawOrdersReq = {
  tenantId?: number // 租户ID
  userId?: number // 用户ID
  orderNo?: string // 平台订单号
  cursor?: string | null // 分页游标
  limit?: number // 分页大小
}

export type ListWithdrawNotifyLogsReq = {
  tenantId?: number // 租户ID
  orderNo?: string // 平台订单号
  orderId?: number // 订单ID
  platformId?: number // 平台ID
  channelId?: number // 通道ID
  notifyStatus?: number // 处理状态
  signResult?: number // 验签结果
  createTimeStart?: number // 创建时间开始
  createTimeEnd?: number // 创建时间结束
  cursor?: string | null // 分页游标
  limit?: number // 分页大小
}

export class WithdrawService {
  getWithdrawOrderList(params: ListWithdrawOrdersReq): Promise<RespBase<WithdrawOrder[]>> {
    return apiWithdrawOrderList(params)
  }
  getWithdrawOrderDetail(orderNo: string, tenantId: number): Promise<RespBase<WithdrawOrder>> {
    return apiWithdrawOrderDetail(orderNo, tenantId)
  }
  auditWithdrawOrder(params: {
    tenantId: number // 租户ID
    orderNo: string // 平台订单号
    approve: number // 审核结果：1通过 2拒绝
    remark?: string // 备注
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
