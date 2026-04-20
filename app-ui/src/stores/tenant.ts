import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import {
  clearTenantId,
  clearTenantCode,
  getTenantCode,
  getTenantId,
  setTenantCode,
  setTenantId,
} from '@/api/http'

export const useTenantStore = defineStore('tenant', () => {
  const tenantId = ref<number | null>(getTenantId())
  const tenantCode = ref(getTenantCode())

  const hasTenantId = computed(() => tenantId.value !== null && tenantId.value > 0)
  const hasTenantCode = computed(() => Boolean(tenantCode.value))

  function hydrateFromEnv() {
    if (!tenantId.value) {
      const envTenantId = Number(import.meta.env.VITE_TENANT_ID)
      if (Number.isFinite(envTenantId) && envTenantId > 0) {
        tenantId.value = envTenantId
        setTenantId(envTenantId)
      }
    }

    if (!tenantCode.value) {
      tenantCode.value = import.meta.env.VITE_TENANT_CODE || ''
      if (tenantCode.value) {
        setTenantCode(tenantCode.value)
      }
    }
  }

  function updateTenantId(value?: number | null) {
    if (!value || value <= 0) return
    tenantId.value = value
    setTenantId(value)
  }

  function updateTenantCode(value?: string | null) {
    if (!value) return
    tenantCode.value = value
    setTenantCode(value)
  }

  function resetTenant() {
    const envTenantId = Number(import.meta.env.VITE_TENANT_ID)
    tenantId.value = Number.isFinite(envTenantId) && envTenantId > 0 ? envTenantId : null
    tenantCode.value = import.meta.env.VITE_TENANT_CODE || ''
    clearTenantId()
    clearTenantCode()
    if (tenantId.value) {
      setTenantId(tenantId.value)
    }
    if (tenantCode.value) {
      setTenantCode(tenantCode.value)
    }
  }

  hydrateFromEnv()

  return {
    tenantId,
    tenantCode,
    hasTenantId,
    hasTenantCode,
    hydrateFromEnv,
    updateTenantId,
    updateTenantCode,
    resetTenant,
  }
})
