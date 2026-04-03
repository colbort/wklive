import type { RespBase, BaseService } from '@/services'

// ===== ITICK服务 =====

class TenantCategoriesService implements BaseService {}

export { TenantCategoriesService }
export const tenantCategoriesService = new TenantCategoriesService()
