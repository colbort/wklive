import { get, post, put } from '@/utils/request'
import type { RespBase } from '@/services'

export type AuthorityRegistry = {
  id: number
  authority: string
  providerCode: string
  producerType: string
  allowedKinds: string[]
  status: number
  version: number
  createTimes: number
  updateTimes: number
}

export type SetAuthorityRegistryReq = {
  id?: number
  authority: string
  providerCode: string
  producerType: string
  allowedKinds: string[]
  status: 1 | 2
  version?: number
}

export type ListAuthorityRegistriesReq = {
  authority?: string
  providerCode?: string
  producerType?: string
  snapshotKind?: string
  status?: number
  cursor?: number
  limit?: number
}

export type PriceFormulaComponent = {
  authority: string
  snapshotKind: string
  categoryCode?: string
  market?: string
  symbol: string
  weight: string
}

export type CreatePriceFormulaReq = {
  formulaNo: string
  authority: string
  snapshotKind: string
  categoryCode?: string
  market?: string
  symbol: string
  algorithm: number
  formulaVersion: string
  components: PriceFormulaComponent[]
  maxLookbackMs: number
  maxDeviationBps: number
  minInputCount: number
  intervalMs: number
  activate: boolean
}

export type PriceFormula = CreatePriceFormulaReq & {
  id: number
  lastTargetTime: number
  status: number
  version: number
  createTimes: number
  updateTimes: number
}

export type ListPriceFormulasReq = {
  authority?: string
  snapshotKind?: string
  categoryCode?: string
  market?: string
  symbol?: string
  status?: number
  cursor?: number
  limit?: number
}

export type SnapshotOutbox = {
  id: number
  snapshotId: string
  status: number
  retryCount: number
  nextRetryAt: number
  lastErrorMsg: string
  redisPublishedAt: number
  eventPublishedAt: number
  createTimes: number
  updateTimes: number
}

export type ListSnapshotOutboxReq = {
  status?: number
  snapshotId?: string
  cursor?: number
  limit?: number
}

export function apiListAuthorityRegistries(
  params: ListAuthorityRegistriesReq,
): Promise<RespBase<AuthorityRegistry[]>> {
  return get<AuthorityRegistry[]>('/admin/market/authorities', params)
}

export function apiSetAuthorityRegistry(
  params: SetAuthorityRegistryReq,
): Promise<RespBase<AuthorityRegistry>> {
  return post<AuthorityRegistry>('/admin/market/authorities', params)
}

export function apiListPriceFormulas(
  params: ListPriceFormulasReq,
): Promise<RespBase<PriceFormula[]>> {
  return get<PriceFormula[]>('/admin/market/price-formulas', params)
}

export function apiCreatePriceFormula(
  params: CreatePriceFormulaReq,
): Promise<RespBase<PriceFormula>> {
  return post<PriceFormula>('/admin/market/price-formulas', params)
}

export function apiChangePriceFormulaStatus(id: number, status: 1 | 3): Promise<RespBase> {
  return put(`/admin/market/price-formulas/${id}/status`, { status })
}

export function apiListSnapshotOutbox(
  params: ListSnapshotOutboxReq,
): Promise<RespBase<SnapshotOutbox[]>> {
  return get<SnapshotOutbox[]>('/admin/market/snapshot-outbox', params)
}

export function apiRetrySnapshotOutbox(id: number): Promise<RespBase> {
  return post(`/admin/market/snapshot-outbox/${id}/retry`)
}

export function apiRevokeAuthoritativeSnapshot(params: {
  snapshotId: string
  replacementSnapshotId?: string
  reason: string
}): Promise<RespBase> {
  return post('/admin/market/authoritative-snapshots/revoke', params)
}
