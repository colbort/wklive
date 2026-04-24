import { get } from '@/utils/request'
import type { OptionGroup, RespBase } from '@/services'

export function apiOptions(): Promise<RespBase<OptionGroup[]>> {
  return get('/admin/system/options')
}
