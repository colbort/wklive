import { apiGetSystemCore } from '@wklive/api/api/core'
import type { OptionsGroup, RespBase } from '@/types/api'

export * from '@wklive/api/api/core'

type CoreOptionsResp = RespBase & { data: OptionsGroup[] }

let coreOptionsPromise: Promise<CoreOptionsResp> | null = null

export function apiGetCoreOptions(): Promise<CoreOptionsResp> {
  if (!coreOptionsPromise) {
    coreOptionsPromise = apiGetSystemCore().then((res) => {
      const data = (res.data as { options?: OptionsGroup[] } | undefined)?.options || []

      return {
        ...res,
        data,
      }
    })
  }

  return coreOptionsPromise
}
