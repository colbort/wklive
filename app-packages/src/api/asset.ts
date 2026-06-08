import { authHttp, http } from './http'
import { compactParams } from './utils'
import type { OptionsGroup, RespBase } from '../types/api'
import type {
  AssetCoinConfig,
  AssetFlow,
  AssetFreeze,
  AssetLock,
  AssetUserAsset,
  GetMyAssetReq,
  GetMyAssetSummaryReq,
  ListAssetCoinConfigsReq,
  ListMyAssetFlowsReq,
  ListMyAssetsReq,
  ListMyFreezesReq,
  ListMyLocksReq,
  TransferMyAssetReq,
  TransferMyAssetResp,
  UserAssetSummary,
} from '../types/asset'

export function apiGetAssetOptions(): Promise<
  RespBase & { data: OptionsGroup[] }
> {
  return http.get('/asset/options').then((res: { data: any }) => res.data)
}

export function apiListAssetCoinConfigs(
  params: ListAssetCoinConfigsReq,
): Promise<RespBase & { data: AssetCoinConfig[] }> {
  return authHttp
    .get('/asset/coin-configs', { params: compactParams(params) })
    .then((res: { data: any }) => res.data)
}

export function apiGetMyAssetSummary(
  params: GetMyAssetSummaryReq,
): Promise<RespBase & { data: UserAssetSummary }> {
  return authHttp
    .get('/asset/summary', { params: compactParams(params) })
    .then((res: { data: any }) => res.data)
}

export function apiListMyAssets(
  params: ListMyAssetsReq,
): Promise<RespBase & { data: AssetUserAsset[] }> {
  return authHttp
    .get('/asset/assets', { params: compactParams(params) })
    .then((res: { data: any }) => res.data)
}

export function apiGetMyAsset(
  params: GetMyAssetReq,
): Promise<RespBase & { data: AssetUserAsset }> {
  return authHttp
    .get('/asset/assets/detail', { params: compactParams(params) })
    .then((res: { data: any }) => res.data)
}

export function apiListMyAssetFlows(
  params: ListMyAssetFlowsReq,
): Promise<RespBase & { data: AssetFlow[] }> {
  return authHttp
    .get('/asset/flows', { params: compactParams(params) })
    .then((res: { data: any }) => res.data)
}

export function apiListMyFreezes(
  params: ListMyFreezesReq,
): Promise<RespBase & { data: AssetFreeze[] }> {
  return authHttp
    .get('/asset/freezes', { params: compactParams(params) })
    .then((res: { data: any }) => res.data)
}

export function apiListMyLocks(
  params: ListMyLocksReq,
): Promise<RespBase & { data: AssetLock[] }> {
  return authHttp
    .get('/asset/locks', { params: compactParams(params) })
    .then((res: { data: any }) => res.data)
}

export function apiTransferMyAsset(
  params: TransferMyAssetReq,
): Promise<RespBase & { data: TransferMyAssetResp }> {
  return authHttp
    .post('/asset/transfer', params)
    .then((res: { data: any }) => res.data)
}
