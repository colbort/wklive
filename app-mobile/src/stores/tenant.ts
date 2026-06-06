import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import {
  clearTenantCode,
  getTenantCode,
  setTenantCode,
} from '@/api/http'

export const useTenantStore = defineStore('tenant', () => {
  const tenantCode = ref(getTenantCode())

  const hasTenantCode = computed(() => Boolean(tenantCode.value))

  function hydrateFromEnv() {
    if (!tenantCode.value) {
      tenantCode.value = import.meta.env.VITE_TENANT_CODE || ''
      if (tenantCode.value) {
        setTenantCode(tenantCode.value)
      }
    }
  }

  function updateTenantCode(value?: string | null) {
    if (!value) return
    tenantCode.value = value
    setTenantCode(value)
  }

  function resetTenant() {
    tenantCode.value = import.meta.env.VITE_TENANT_CODE || ''
    clearTenantCode()
    if (tenantCode.value) {
      setTenantCode(tenantCode.value)
    }
  }

  hydrateFromEnv()

  return {
    tenantCode,
    hasTenantCode,
    hydrateFromEnv,
    updateTenantCode,
    resetTenant,
  }
})
