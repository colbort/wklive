import type { RespBase, BaseService } from '@/services'

// ===== ITICK服务 =====

class TenantProductsService implements BaseService {}

export { TenantProductsService }
export const tenantProductsService = new TenantProductsService()
