import { del, get, post, put } from '@/utils/request'
import type {
  AdminAddAssetReq,
  AdminFreezeAssetReq,
  AdminLockAssetReq,
  AdminSubAssetReq,
  AdminUnfreezeAssetReq,
  AdminUnlockAssetReq,
  AssetChangeResp,
  AssetCoinConfig,
  AssetFlow,
  AssetFreeze,
  AssetLock,
  AssetUserAsset,
  CreateAssetCoinConfigReq,
  GetAssetCoinConfigReq,
  PageAssetCoinConfigsReq,
  PageAssetFlowsReq,
  PageAssetFreezesReq,
  PageAssetLocksReq,
  PageUserAssetsReq,
  RespBase,
  UpdateAssetCoinConfigReq,
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

export function apiPageAssetCoinConfigs(
  params: PageAssetCoinConfigsReq,
): Promise<RespBase<AssetCoinConfig[]>> {
  return get<AssetCoinConfig[]>('/admin/asset/coin-configs', params)
}

export function apiGetAssetCoinConfig(
  params: GetAssetCoinConfigReq,
): Promise<RespBase<AssetCoinConfig>> {
  const { id, ...query } = params
  return get<AssetCoinConfig>(`/admin/asset/coin-configs/${id}`, query)
}

export function apiCreateAssetCoinConfig(
  params: CreateAssetCoinConfigReq,
): Promise<RespBase<AssetCoinConfig>> {
  return post<AssetCoinConfig>('/admin/asset/coin-configs', params)
}

export function apiUpdateAssetCoinConfig(
  params: UpdateAssetCoinConfigReq,
): Promise<RespBase<AssetCoinConfig>> {
  const { id, ...data } = params
  return put<AssetCoinConfig>(`/admin/asset/coin-configs/${id}`, data)
}

export function apiDeleteAssetCoinConfig(
  id: number,
  params?: { tenantId?: number },
): Promise<RespBase> {
  return del('/admin/asset/coin-configs/' + id, params)
}

export function apiAdminAddAsset(params: AdminAddAssetReq): Promise<RespBase<AssetChangeResp>> {
  return post<AssetChangeResp>('/admin/asset/add', params)
}

export function apiAdminSubAsset(params: AdminSubAssetReq): Promise<RespBase<AssetChangeResp>> {
  return post<AssetChangeResp>('/admin/asset/sub', params)
}

export function apiAdminFreezeAsset(
  params: AdminFreezeAssetReq,
): Promise<RespBase<AssetChangeResp>> {
  return post<AssetChangeResp>('/admin/asset/freeze', params)
}

export function apiAdminUnfreezeAsset(
  params: AdminUnfreezeAssetReq,
): Promise<RespBase<AssetChangeResp>> {
  return post<AssetChangeResp>('/admin/asset/unfreeze', params)
}

export function apiAdminLockAsset(params: AdminLockAssetReq): Promise<RespBase<AssetChangeResp>> {
  return post<AssetChangeResp>('/admin/asset/lock', params)
}

export function apiAdminUnlockAsset(
  params: AdminUnlockAssetReq,
): Promise<RespBase<AssetChangeResp>> {
  return post<AssetChangeResp>('/admin/asset/unlock', params)
}
