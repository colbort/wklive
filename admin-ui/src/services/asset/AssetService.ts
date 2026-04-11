import type { RespBase } from '@/services'
import {
  apiAdminAddAsset,
  apiAdminFreezeAsset,
  apiAdminLockAsset,
  apiAdminSubAsset,
  apiAdminUnfreezeAsset,
  apiAdminUnlockAsset,
  apiGetUserAssetDetail,
  apiPageAssetFlows,
  apiPageAssetFreezes,
  apiPageAssetLocks,
  apiPageUserAssets,
} from '@/api/asset'

export type AssetUserAsset = {
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
  createTime: number
  updateTime: number
}

export type TimeRange = {
  startTime?: number
  endTime?: number
}

export type AssetFlow = {
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
  remark: string
  createTime: number
  updateTime: number
}

export type AssetFreeze = {
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
  createTime: number
  updateTime: number
}

export type AssetLock = {
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
  createTime: number
  updateTime: number
}
export type AssetChangeResp = {
  bizNo: string
  asset: AssetUserAsset
}

export type PageUserAssetsReq = {
  cursor?: number
  limit?: number
  tenantId?: number
  userId?: number
  walletType?: number
  coin?: string
  status?: number
}

export type GetUserAssetDetailReq = {
  tenantId: number
  userId: number
  walletType: number
  coin: string
}

export type PageAssetFlowsReq = {
  cursor?: number
  limit?: number
  tenantId?: number
  userId?: number
  walletType?: number
  coin?: string
  bizType?: number
  sceneType?: number
  bizNo?: string
  timeRange?: TimeRange
}

export type PageAssetFreezesReq = {
  cursor?: number
  limit?: number
  tenantId?: number
  userId?: number
  walletType?: number
  coin?: string
  bizType?: number
  bizNo?: string
  status?: number
}

export type PageAssetLocksReq = {
  cursor?: number
  limit?: number
  tenantId?: number
  userId?: number
  walletType?: number
  coin?: string
  bizType?: number
  bizNo?: string
  status?: number
}

export type AdminAddAssetReq = {
  tenantId: number
  userId: number
  walletType: number
  coin: string
  amount: string
  bizNo: string
  remark?: string
  operatorId: number
}

export type AdminSubAssetReq = AdminAddAssetReq

export type AdminFreezeAssetReq = AdminAddAssetReq

export type AdminUnfreezeAssetReq = {
  tenantId: number
  freezeNo: string
  amount: string
  bizNo: string
  remark?: string
  operatorId: number
}

export type AdminLockAssetReq = AdminAddAssetReq

export type AdminUnlockAssetReq = {
  tenantId: number
  lockNo: string
  amount: string
  bizNo: string
  remark?: string
  operatorId: number
}

export class AssetService {
  getUserAssets(params: PageUserAssetsReq): Promise<RespBase<AssetUserAsset[]>> {
    return apiPageUserAssets(params)
  }

  getUserAssetDetail(params: GetUserAssetDetailReq) {
    return apiGetUserAssetDetail(params)
  }

  getFlows(params: PageAssetFlowsReq): Promise<RespBase<AssetFlow[]>> {
    return apiPageAssetFlows(params)
  }

  getFreezes(params: PageAssetFreezesReq): Promise<RespBase<AssetFreeze[]>> {
    return apiPageAssetFreezes(params)
  }

  getLocks(params: PageAssetLocksReq): Promise<RespBase<AssetLock[]>> {
    return apiPageAssetLocks(params)
  }

  addAsset(params: AdminAddAssetReq) {
    return apiAdminAddAsset(params)
  }

  subAsset(params: AdminSubAssetReq) {
    return apiAdminSubAsset(params)
  }

  freezeAsset(params: AdminFreezeAssetReq) {
    return apiAdminFreezeAsset(params)
  }

  unfreezeAsset(params: AdminUnfreezeAssetReq) {
    return apiAdminUnfreezeAsset(params)
  }

  lockAsset(params: AdminLockAssetReq) {
    return apiAdminLockAsset(params)
  }

  unlockAsset(params: AdminUnlockAssetReq) {
    return apiAdminUnlockAsset(params)
  }
}

export const assetService = new AssetService()
