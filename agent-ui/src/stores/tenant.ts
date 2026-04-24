import { defineStore } from 'pinia'
import { ENV } from '@/config/environment'
import { tenantsService, type SysTenantItem } from '@/services'

type TenantState = {
  ready: boolean
  loading: boolean
  tenantCode: string
  tenantId: number
  tenantName: string
  detail: SysTenantItem | null
}

export const useTenantStore = defineStore('tenant', {
  state: (): TenantState => ({
    ready: false,
    loading: false,
    tenantCode: ENV.TENANT_CODE,
    tenantId: 0,
    tenantName: '',
    detail: null,
  }),
  actions: {
    async ensureLoaded(force = false) {
      if (this.ready && !force) return this.detail
      if (this.loading) return this.detail
      if (!this.tenantCode) {
        throw new Error('VITE_TENANT_CODE is required for agent-ui')
      }

      this.loading = true
      try {
        const res = await tenantsService.detail({ tenantCode: this.tenantCode })
        if (res.code !== 200 || !res.data) {
          throw new Error(res.msg || `Tenant ${this.tenantCode} not found`)
        }
        this.detail = res.data
        this.tenantId = res.data.id
        this.tenantName = res.data.tenantName
        this.ready = true
        return this.detail
      } finally {
        this.loading = false
      }
    },
  },
})
