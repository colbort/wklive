import { ref } from 'vue'
import { getSystemCore } from '@/stores/core'
import { logger } from '@/utils/logger'
import { buildAssetUrl } from '@/utils/file-url'

export type SystemCoreState = {
  siteName: string
  siteLogo: string
}

const DEFAULT_SYSTEM_CORE: SystemCoreState = {
  siteName: 'Admin UI',
  siteLogo: '',
}

function normalizeSystemCore(data: any): SystemCoreState {
  return {
    siteName: data?.siteName || data?.site_name || DEFAULT_SYSTEM_CORE.siteName,
    siteLogo: data?.siteLogo || data?.site_logo || DEFAULT_SYSTEM_CORE.siteLogo,
  }
}

function ensureFavicon(href: string) {
  let icon = document.querySelector("link[rel~='icon']") as HTMLLinkElement | null
  if (!icon) {
    icon = document.createElement('link')
    icon.rel = 'icon'
    document.head.appendChild(icon)
  }
  icon.href = href
}

export function applySystemBranding(core: SystemCoreState) {
  if (core.siteName) {
    document.title = core.siteName
  }
  if (core.siteLogo) {
    ensureFavicon(buildAssetUrl(core.siteLogo))
  }
}

export function useSystemCore(initial?: Partial<SystemCoreState>) {
  const systemCore = ref<SystemCoreState>({
    ...DEFAULT_SYSTEM_CORE,
    ...initial,
  })

  async function loadSystemCore() {
    try {
      const res = await getSystemCore()
      if (res?.code === 200 && res.data) {
        systemCore.value = normalizeSystemCore(res.data)
      }
      return systemCore.value
    } catch (error) {
      logger.warn('Failed to load system core config', error)
      return systemCore.value
    }
  }

  return {
    systemCore,
    loadSystemCore,
  }
}
