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
  id: number // 主键ID
  tenantId: number // 租户ID
  userId: number // 用户ID
  successOrderCount: number // 成功充值笔数
  successTotalAmount: number // 成功累计充值金额，单位分
  todaySuccessAmount: number // 今日成功充值金额，单位分
  todaySuccessCount: number // 今日成功充值次数
  firstSuccessTime: number // 首次成功充值时间
  lastSuccessTime: number // 最近成功充值时间
  createTimes: number // 创建时间
  updateTimes: number // 更新时间
}

export type RechargeOrder = {
  id: number // 充值订单ID
  tenantId: number // 租户ID
  userId: number // 用户ID
  orderNo: string // 平台订单号
  bizOrderNo: string // 业务订单号
  platformId: number // 平台ID
  productId: number // 产品ID
  accountId: number // 账号ID
  channelId: number // 通道ID
  currency: string // 币种
  orderAmount: number // 订单金额，单位分
  payAmount: number // 实际支付金额，单位分
  feeAmount: number // 手续费金额，单位分
  subject: string // 标题
  body: string // 描述
  clientType: number // 客户端类型：1APP 2H5 3WEB
  clientIp: string // 客户端IP
  status: number // 状态：1待支付 2支付中 3成功 4失败 5已关闭 6已退款
  thirdTradeNo: string // 三方交易号
  thirdOrderNo: string // 三方订单号
  payUrl: string // 支付链接
  qrContent: string // 二维码内容
  requestData: string // 请求快照(JSON)
  responseData: string // 响应快照(JSON)
  notifyData: string // 回调数据(JSON)
  expireTime: number // 过期时间
  paidTime: number // 支付时间
  notifyTime: number // 回调时间
  closeTime: number // 关闭时间
  remark: string // 备注
  createTimes: number // 创建时间
  updateTimes: number // 更新时间
}

export type PayNotifyLog = {
  id: number // 回调日志ID
  tenantId: number // 租户ID
  orderId: number // 订单ID
  orderNo: string // 平台订单号
  platformId: number // 平台ID
  channelId: number // 通道ID
  notifyStatus: number // 处理状态：1待处理 2成功 3失败
  notifyBody: string // 回调原文
  signResult: number // 验签结果：1未校验 2通过 3失败
  processResult: string // 处理结果
  errorMessage: string // 错误信息
  notifyTime: number // 回调时间
  createTimes: number // 创建时间
}

export type GetUserRechargeStatReq = {
  tenantId: number // 租户ID
  userId: number // 用户ID
}

export type ListUserRechargeStatsReq = {
  tenantId?: number // 租户ID
  userId?: number // 用户ID
  successTotalAmountMin?: number // 累计充值金额最小值，单位分
  successTotalAmountMax?: number // 累计充值金额最大值，单位分
  cursor?: string | null // 分页游标
  limit?: number // 分页大小
}

export type ListRechargeOrdersReq = {
  tenantId?: number // 租户ID
  userId?: number // 用户ID
  platformId?: number // 平台ID
  productId?: number // 产品ID
  accountId?: number // 账号ID
  channelId?: number // 通道ID
  orderNo?: string // 平台订单号
  bizOrderNo?: string // 业务订单号
  thirdTradeNo?: string // 三方交易号
  status?: number // 订单状态
  createTimeStart?: number // 创建时间开始
  createTimeEnd?: number // 创建时间结束
  cursor?: string | null // 分页游标
  limit?: number // 分页大小
}

export type ListRechargeNotifyLogsReq = {
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
