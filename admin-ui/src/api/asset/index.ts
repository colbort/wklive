import { get, post } from '@/utils/request'
import type {
  AdminAddAssetReq,
  AdminFreezeAssetReq,
  AdminLockAssetReq,
  AdminSubAssetReq,
  AdminUnfreezeAssetReq,
  AdminUnlockAssetReq,
  AssetChangeResp,
  AssetFlow,
  AssetFreeze,
  AssetLock,
  AssetUserAsset,
  PageAssetFlowsReq,
  PageAssetFreezesReq,
  PageAssetLocksReq,
  PageUserAssetsReq,
  RespBase,
} from '@/services'

export function apiPageUserAssets(params: PageUserAssetsReq): Promise<RespBase<AssetUserAsset[]>> {
  return get<AssetUserAsset[]>('/admin/asset/user-assets', params)
}

export function apiGetUserAssetDetail(params: {
  tenantId: number
  userId: number
  walletType: number
  coin: string
}): Promise<RespBase<AssetUserAsset>> {
  return get<AssetUserAsset>('/admin/asset/user-assets/detail', params)
}

export function apiPageAssetFlows(params: PageAssetFlowsReq): Promise<RespBase<AssetFlow[]>> {
  return get<AssetFlow[]>('/admin/asset/flows', params)
}

export function apiPageAssetFreezes(params: PageAssetFreezesReq): Promise<RespBase<AssetFreeze[]>> {
  return get<AssetFreeze[]>('/admin/asset/freezes', params)
}

export function apiPageAssetLocks(params: PageAssetLocksReq): Promise<RespBase<AssetLock[]>> {
  return get<AssetLock[]>('/admin/asset/locks', params)
}

export function apiAdminAddAsset(params: AdminAddAssetReq): Promise<RespBase<AssetChangeResp>> {
  return post<AssetChangeResp>('/admin/asset/add', params)
}

export function apiAdminSubAsset(params: AdminSubAssetReq): Promise<RespBase<AssetChangeResp>> {
  return post<AssetChangeResp>('/admin/asset/sub', params)
}

export function apiAdminFreezeAsset(params: AdminFreezeAssetReq): Promise<RespBase<AssetChangeResp>> {
  return post<AssetChangeResp>('/admin/asset/freeze', params)
}

export function apiAdminUnfreezeAsset(params: AdminUnfreezeAssetReq): Promise<RespBase<AssetChangeResp>> {
  return post<AssetChangeResp>('/admin/asset/unfreeze', params)
}

export function apiAdminLockAsset(params: AdminLockAssetReq): Promise<RespBase<AssetChangeResp>> {
  return post<AssetChangeResp>('/admin/asset/lock', params)
}

export function apiAdminUnlockAsset(params: AdminUnlockAssetReq): Promise<RespBase<AssetChangeResp>> {
  return post<AssetChangeResp>('/admin/asset/unlock', params)
}
