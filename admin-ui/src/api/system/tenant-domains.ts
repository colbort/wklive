import { del, get, post, put } from '@/utils/request'
import type { RespBase } from '@/services'
import type {
  SysTenantDomainCreateReq,
  SysTenantDomainGuestMigrationStats,
  SysTenantDomainItem,
  SysTenantDomainUpdateReq,
} from '@/services/system/TenantDomainsService'

export function apiSysTenantDomainList(tenantId: number): Promise<RespBase<SysTenantDomainItem[]>> {
  return get<SysTenantDomainItem[]>('/admin/system/tenant-domains', { tenantId })
}

export function apiSysTenantDomainGuestMigrationStats(
  tenantId: number,
  sourceOrigin: string,
): Promise<RespBase<SysTenantDomainGuestMigrationStats>> {
  return get<SysTenantDomainGuestMigrationStats>(
    '/admin/system/tenant-domains/guest-migration-stats',
    { tenantId, sourceOrigin },
  )
}

export function apiSysTenantDomainCreate(data: SysTenantDomainCreateReq): Promise<RespBase> {
  return post('/admin/system/tenant-domains', data)
}

export function apiSysTenantDomainUpdate(data: SysTenantDomainUpdateReq): Promise<RespBase> {
  return put('/admin/system/tenant-domains', data)
}

export function apiSysTenantDomainDelete(id: number): Promise<RespBase> {
  return del(`/admin/system/tenant-domains/${id}`)
}
