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

export interface CryptoRechargeOrderData {
  order: RechargeOrder
  address: CryptoRechargeAddress
}

export interface CryptoRechargeAddress {
  id: number
  tenantId: number
  userId: number
  walletType: number
  coin: string
  chainCode: number
  address: string
  memo: string
  addressSource: number
  addressType: number
  status: number
  lastUsedTime: number
  createTimes: number
  updateTimes: number
}

export interface CryptoRechargeTx {
  id: number
  tenantId: number
  userId: number
  orderId: number
  orderNo: string
  coin: string
  chainCode: number
  txHash: string
  fromAddress: string
  toAddress: string
  memo: string
  amount: string
  blockHeight: number
  confirmCount: number
  requiredConfirmCount: number
  status: number
  rawData: string
  createTimes: number
  updateTimes: number
}

export interface GetMyRechargeStatReq {}

export interface ListAvailableRechargeChannelsReq {
  rechargeAmount: number
  currency: string
  clientType: number
}

export interface CreateRechargeOrderReq {
  channelId: number
  rechargeAmount: number
  currency: string
  subject?: string
  body?: string
  clientType: number
  clientIp?: string
  bizOrderNo?: string
}

export interface CreateCryptoRechargeOrderReq {
  walletType: number
  coin: string
  chainCode: number
  rechargeAmount: number
  clientType: number
  clientIp?: string
  bizOrderNo?: string
}

export interface GetMyRechargeOrderReq {
  orderNo: string
}

export interface ListMyRechargeOrdersReq extends PageReq {
  status?: number
  orderNo?: string
  createTimeStart?: number
  createTimeEnd?: number
}

export interface CancelMyRechargeOrderReq {
  orderNo: string
}

export interface QueryMyRechargeOrderStatusReq {
  orderNo: string
}

export interface GetMyCryptoRechargeAddressReq {
  walletType: number
  coin: string
  chainCode: number
}

export interface ListMyCryptoRechargeAddressesReq {
  walletType?: number
  coin?: string
  chainCode?: number
}

export interface ListMyCryptoRechargeTxsReq extends PageReq {
  orderNo?: string
  coin?: string
  chainCode?: number
  status?: number
  createTimeStart?: number
  createTimeEnd?: number
}

export interface GetMyCryptoRechargeTxReq {
  id: number
  txHash?: string
}

export interface CreateWithdrawOrderReq {
  amount: number
  currency: string
  address: string
  bankId: number
  remark: string
}

export interface ListMyWithdrawOrdersReq extends PageReq {
  status?: number
}

export interface GetMyWithdrawOrderReq {
  orderNo: string
}
