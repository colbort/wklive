import { http } from '@/api/http'
import { compactParams } from '@/api/utils'
import type { RespBase } from '@/types/api'
import type {
  AssetFlow,
  AssetFreeze,
  AssetLock,
  AssetUserAsset,
  GetMyAssetReq,
  GetMyAssetSummaryReq,
  ListMyAssetFlowsReq,
  ListMyAssetsReq,
  ListMyFreezesReq,
  ListMyLocksReq,
  UserAssetSummary,
} from '@/types/asset'

export function apiGetMyAssetSummary(params: GetMyAssetSummaryReq): Promise<RespBase & { data: UserAssetSummary }> {
  return http.get('/asset/summary', { params: compactParams(params) }).then((res) => res.data)
}

export function apiListMyAssets(params: ListMyAssetsReq): Promise<RespBase & { data: AssetUserAsset[] }> {
  return http.get('/asset/assets', { params: compactParams(params) }).then((res) => res.data)
}

export function apiGetMyAsset(params: GetMyAssetReq): Promise<RespBase & { asset: AssetUserAsset }> {
  return http.get('/asset/assets/detail', { params: compactParams(params) }).then((res) => res.data)
}

export function apiListMyAssetFlows(
  params: ListMyAssetFlowsReq,
): Promise<RespBase & { data: AssetFlow[] }> {
  return http.get('/asset/flows', { params: compactParams(params) }).then((res) => res.data)
}

export function apiListMyFreezes(params: ListMyFreezesReq): Promise<RespBase & { data: AssetFreeze[] }> {
  return http.get('/asset/freezes', { params: compactParams(params) }).then((res) => res.data)
}

export function apiListMyLocks(params: ListMyLocksReq): Promise<RespBase & { data: AssetLock[] }> {
  return http.get('/asset/locks', { params: compactParams(params) }).then((res) => res.data)
}
