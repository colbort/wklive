import { defineStore } from 'pinia'
import { ref } from 'vue'
import { apiGetSystemCore } from '@/api/core'
import type { SystemCore } from '@/types/core'

const DEFAULT_SYSTEM_CORE: SystemCore = {
  isCaptchaEnabled: false,
  isRegisterEnabled: false,
  isGuestEnabled: false,
  isCryptoEnabled: false,
  intervals: [],
}

export const useSystemStore = defineStore('system', () => {
  const systemCore = ref<SystemCore>({ ...DEFAULT_SYSTEM_CORE })
  const loading = ref(false)

  async function loadSystemCore() {
    loading.value = true
    try {
      const res = await apiGetSystemCore()
      if ((res.code === 0 || res.code === 200) && res.data) {
        systemCore.value = res.data
      }
      return systemCore.value
    } finally {
      loading.value = false
    }
  }

  return {
    systemCore,
    loading,
    loadSystemCore,
  }
})
