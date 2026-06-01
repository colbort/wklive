import type { PageReq, TimeRange } from '@/types/api'

export interface AssetUserAsset {
  id: number
  tenantId: number
  userId: number
  walletType: number
  coin: string
  totalAmount: string
  availableAmount: string
  frozenAmount: string
  lockedAmount: string
  status: number
  version: number
  remark: string
  createTimes: number
  updateTimes: number
}

export interface AssetCoinConfig {
  id: number
  tenantId: number
  walletType: number
  coin: string
  symbol: string
  coinName: string
  coinType: number
  chainCode: number
  iconUrl: string
  iconText: string
  iconBgColor: string
  decimalPlaces: number
  appVisible: number
  rechargeEnabled: number
  withdrawEnabled: number
  transferEnabled: number
  status: number
  sort: number
  remark: string
  createTimes: number
  updateTimes: number
}

export interface AssetFlow {
  id: number
  flowNo: string
  tenantId: number
  userId: number
  walletType: number
  coin: string
  bizType: number
  sceneType: number
  opType: number
  bizId: number
  bizNo: string
  changeAmount: string
  beforeTotalAmount: string
  afterTotalAmount: string
  beforeAvailableAmount: string
  afterAvailableAmount: string
  beforeFrozenAmount: string
  afterFrozenAmount: string
  beforeLockedAmount: string
  afterLockedAmount: string
  balanceSnapshotVersion: number
  changeType: string
  remark: string
  createTimes: number
  updateTimes: number
}

export interface AssetFreeze {
  id: number
  freezeNo: string
  tenantId: number
  userId: number
  walletType: number
  coin: string
  bizType: number
  sceneType: number
  bizId: number
  bizNo: string
  amount: string
  usedAmount: string
  unfreezeAmount: string
  remainAmount: string
  status: number
  expireTime: number
  remark: string
  createTimes: number
  updateTimes: number
}

export interface AssetLock {
  id: number
  lockNo: string
  tenantId: number
  userId: number
  walletType: number
  coin: string
  bizType: number
  sceneType: number
  bizId: number
  bizNo: string
  amount: string
  unlockAmount: string
  remainAmount: string
  status: number
  startTime: number
  endTime: number
  remark: string
  createTimes: number
  updateTimes: number
}

export interface UserAssetSummary {
  tenantId: number
  userId: number
  totalAssetUsdt: string
  totalAvailableUsdt: string
  totalFrozenUsdt: string
  totalLockedUsdt: string
  assets: AssetUserAsset[]
}

export interface TransferMyAssetReq {
  fromWalletType: number
  toWalletType: number
  coin: string
  amount: string
  remark?: string
}

export interface TransferMyAssetResp {
  fromAsset: AssetUserAsset
  toAsset: AssetUserAsset
}

export interface GetMyAssetSummaryReq {
}

export interface ListAssetCoinConfigsReq {
  walletType: number
  operationType: number
  coinType?: number
}

export interface ListMyAssetsReq {
  walletType?: number
  coin?: string
}

export interface GetMyAssetReq {
  walletType: number
  coin: string
}

export interface ListMyAssetFlowsReq extends PageReq {
  walletType?: number
  coin?: string
  bizType?: number
  sceneType?: number
  timeRange?: TimeRange
}

export interface ListMyFreezesReq extends PageReq {
  walletType: number
  coin?: string
  status?: number
}

export interface ListMyLocksReq extends PageReq {
  walletType: number
  coin?: string
  status?: number
}
