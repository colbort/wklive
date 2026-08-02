import { del, get, post, put } from '@/utils/request'
import type {
  AddAssetReq,
  FreezeAssetReq,
  LockAssetReq,
  SubAssetReq,
  UnfreezeAssetReq,
  UnlockAssetReq,
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
  PlatformAccount,
  GetPlatformAccountReq,
  SetPlatformAccountReq,
  AdjustPlatformAccountReq,
  InsuranceCover,
  PlatformBackstopPolicy,
  CreatePlatformBackstopPolicyReq,
  ReviewPlatformBackstopPolicyReq,
  ListPlatformBackstopPoliciesReq,
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

export function apiAdminAddAsset(params: AddAssetReq): Promise<RespBase<AssetChangeResp>> {
  return post<AssetChangeResp>('/admin/asset/add', params)
}

export function apiAdminSubAsset(params: SubAssetReq): Promise<RespBase<AssetChangeResp>> {
  return post<AssetChangeResp>('/admin/asset/sub', params)
}

export function apiAdminFreezeAsset(params: FreezeAssetReq): Promise<RespBase<AssetChangeResp>> {
  return post<AssetChangeResp>('/admin/asset/freeze', params)
}

export function apiAdminUnfreezeAsset(
  params: UnfreezeAssetReq,
): Promise<RespBase<AssetChangeResp>> {
  return post<AssetChangeResp>('/admin/asset/unfreeze', params)
}

export function apiAdminLockAsset(params: LockAssetReq): Promise<RespBase<AssetChangeResp>> {
  return post<AssetChangeResp>('/admin/asset/lock', params)
}

export function apiAdminUnlockAsset(params: UnlockAssetReq): Promise<RespBase<AssetChangeResp>> {
  return post<AssetChangeResp>('/admin/asset/unlock', params)
}

export function apiGetPlatformAccount(
  params: GetPlatformAccountReq,
): Promise<RespBase<PlatformAccount>> {
  return get<PlatformAccount>('/admin/asset/platform-accounts', params)
}

export function apiSetPlatformAccount(
  params: SetPlatformAccountReq,
): Promise<RespBase<PlatformAccount>> {
  return post<PlatformAccount>('/admin/asset/platform-accounts', params)
}

export function apiAdjustPlatformAccount(
  params: AdjustPlatformAccountReq,
): Promise<RespBase<PlatformAccount>> {
  return post<PlatformAccount>('/admin/asset/platform-accounts/adjust', params)
}

export function apiGetInsuranceCover(params: {
  tenantId: number
  liquidationNo: string
}): Promise<RespBase<InsuranceCover>> {
  return get<InsuranceCover>('/admin/asset/insurance-covers', params)
}

export function apiCreatePlatformBackstopPolicy(
  params: CreatePlatformBackstopPolicyReq,
): Promise<RespBase<PlatformBackstopPolicy>> {
  return post<PlatformBackstopPolicy>('/admin/asset/platform-backstop-policies', params)
}

export function apiReviewPlatformBackstopPolicy(
  params: ReviewPlatformBackstopPolicyReq,
): Promise<RespBase<PlatformBackstopPolicy>> {
  const { policyId, ...data } = params
  return post<PlatformBackstopPolicy>(
    `/admin/asset/platform-backstop-policies/${policyId}/review`,
    data,
  )
}

export function apiGetPlatformBackstopPolicy(
  tenantId: number,
  policyId: number,
): Promise<RespBase<PlatformBackstopPolicy>> {
  return get<PlatformBackstopPolicy>(`/admin/asset/platform-backstop-policies/${policyId}`, {
    tenantId,
  })
}

export function apiListPlatformBackstopPolicies(
  params: ListPlatformBackstopPoliciesReq,
): Promise<RespBase<PlatformBackstopPolicy[]>> {
  return get<PlatformBackstopPolicy[]>('/admin/asset/platform-backstop-policies', params)
}
