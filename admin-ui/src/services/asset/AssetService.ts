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
  id: number // 主键ID
  tenantId: number // 租户ID
  userId: number // 用户ID
  walletType: number // 钱包类型
  coin: string // 币种
  totalAmount: string // 总资产
  availableAmount: string // 可用资产
  frozenAmount: string // 冻结资产
  lockedAmount: string // 锁定资产
  status: number // 状态
  version: number // 乐观锁版本号
  remark: string // 备注
  createTimes: number // 创建时间
  updateTimes: number // 更新时间
}

export type TimeRange = {
  startTime?: number // 开始时间
  endTime?: number // 结束时间
}

export type AssetFlow = {
  id: number // 主键ID
  flowNo: string // 流水单号
  tenantId: number // 租户ID
  userId: number // 用户ID
  walletType: number // 钱包类型
  coin: string // 币种
  bizType: number // 业务类型
  sceneType: number // 业务场景
  opType: number // 操作方向
  bizId: number // 业务ID
  bizNo: string // 业务单号
  changeAmount: string // 变动金额
  beforeTotalAmount: string // 变动前总资产
  afterTotalAmount: string // 变动后总资产
  beforeAvailableAmount: string // 变动前可用资产
  afterAvailableAmount: string // 变动后可用资产
  beforeFrozenAmount: string // 变动前冻结资产
  afterFrozenAmount: string // 变动后冻结资产
  beforeLockedAmount: string // 变动前锁定资产
  afterLockedAmount: string // 变动后锁定资产
  balanceSnapshotVersion: number // 快照版本号
  changeType: string // 变更类型
  remark: string // 备注
  createTimes: number // 创建时间
  updateTimes: number // 更新时间
}

export type AssetFreeze = {
  id: number // 主键ID
  freezeNo: string // 冻结单号
  tenantId: number // 租户ID
  userId: number // 用户ID
  walletType: number // 钱包类型
  coin: string // 币种
  bizType: number // 业务类型
  sceneType: number // 业务场景
  bizId: number // 业务ID
  bizNo: string // 业务单号
  amount: string // 冻结金额
  usedAmount: string // 已使用金额
  unfreezeAmount: string // 已解冻金额
  remainAmount: string // 剩余冻结金额
  status: number // 状态
  expireTime: number // 过期时间
  remark: string // 备注
  createTimes: number // 创建时间
  updateTimes: number // 更新时间
}

export type AssetLock = {
  id: number // 主键ID
  lockNo: string // 锁仓单号
  tenantId: number // 租户ID
  userId: number // 用户ID
  walletType: number // 钱包类型
  coin: string // 币种
  bizType: number // 业务类型
  sceneType: number // 业务场景
  bizId: number // 业务ID
  bizNo: string // 业务单号
  amount: string // 锁仓金额
  unlockAmount: string // 已解锁金额
  remainAmount: string // 剩余锁仓金额
  status: number // 状态
  startTime: number // 开始时间
  endTime: number // 结束时间
  remark: string // 备注
  createTimes: number // 创建时间
  updateTimes: number // 更新时间
}
export type AssetChangeResp = {
  bizNo: string // 业务单号
  asset: AssetUserAsset // 变更后的资产
}

export type PageUserAssetsReq = {
  cursor?: number // 游标
  limit?: number // 每页条数
  tenantId?: number // 租户ID
  userId?: number // 用户ID
  walletType?: number // 钱包类型
  coin?: string // 币种
  status?: number // 状态
}

export type GetUserAssetDetailReq = {
  tenantId: number // 租户ID
  userId: number // 用户ID
  walletType: number // 钱包类型
  coin: string // 币种
}

export type PageAssetFlowsReq = {
  cursor?: number // 游标
  limit?: number // 每页条数
  tenantId?: number // 租户ID
  userId?: number // 用户ID
  walletType?: number // 钱包类型
  coin?: string // 币种
  bizType?: number // 业务类型
  sceneType?: number // 业务场景
  bizNo?: string // 业务单号
  timeRange?: TimeRange // 时间范围
}

export type PageAssetFreezesReq = {
  cursor?: number // 游标
  limit?: number // 每页条数
  tenantId?: number // 租户ID
  userId?: number // 用户ID
  walletType?: number // 钱包类型
  coin?: string // 币种
  bizType?: number // 业务类型
  bizNo?: string // 业务单号
  status?: number // 状态
}

export type PageAssetLocksReq = {
  cursor?: number // 游标
  limit?: number // 每页条数
  tenantId?: number // 租户ID
  userId?: number // 用户ID
  walletType?: number // 钱包类型
  coin?: string // 币种
  bizType?: number // 业务类型
  bizNo?: string // 业务单号
  status?: number // 状态
}

export type AdminAddAssetReq = {
  tenantId: number // 租户ID
  userId: number // 用户ID
  walletType: number // 钱包类型
  coin: string // 币种
  amount: string // 变更金额
  bizNo: string // 业务单号
  remark?: string // 备注
  operatorId: number // 操作人ID
}

export type AdminSubAssetReq = AdminAddAssetReq

export type AdminFreezeAssetReq = AdminAddAssetReq

export type AdminUnfreezeAssetReq = {
  tenantId: number // 租户ID
  freezeNo: string // 冻结单号
  amount: string // 解冻金额
  bizNo: string // 业务单号
  remark?: string // 备注
  operatorId: number // 操作人ID
}

export type AdminLockAssetReq = AdminAddAssetReq

export type AdminUnlockAssetReq = {
  tenantId: number // 租户ID
  lockNo: string // 锁仓单号
  amount: string // 解锁金额
  bizNo: string // 业务单号
  remark?: string // 备注
  operatorId: number // 操作人ID
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
