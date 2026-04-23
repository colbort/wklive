import { http } from '@/api/http'
import { compactParams } from '@/api/utils'
import type { ApiResp } from '@/types/api'
import type {
  GetKlineReq,
  GetKlineResp,
  GetQuoteReq,
  GetQuoteResp,
  ListVisibleCategoriesReq,
  ListVisibleCategoriesResp,
  ListVisibleProductsReq,
  ListVisibleProductsResp,
} from '@/types/itick'

export async function apiListVisibleCategories(
  params: ListVisibleCategoriesReq,
): Promise<ApiResp<ListVisibleCategoriesResp>> {
  const { data } = await http.get<ApiResp<ListVisibleCategoriesResp>>('/itick/categories', {
    params: compactParams(params),
  })
  return data
}

export async function apiListVisibleProducts(
  params: ListVisibleProductsReq,
): Promise<ApiResp<ListVisibleProductsResp>> {
  const { data } = await http.get<ApiResp<ListVisibleProductsResp>>('/itick/products', {
    params: compactParams(params),
  })
  return data
}

export async function apiGetKline(params: GetKlineReq): Promise<ApiResp<GetKlineResp>> {
  const { data } = await http.get<ApiResp<GetKlineResp>>('/itick/kline', {
    params: compactParams(params),
  })
  return data
}

export async function apiGetQuote(params: GetQuoteReq): Promise<ApiResp<GetQuoteResp>> {
  const { data } = await http.get<ApiResp<GetQuoteResp>>('/itick/quote', {
    params: compactParams(params),
  })
  return data
}

function isLoopbackHost(hostname: string) {
  return ['localhost', '127.0.0.1', '0.0.0.0', '::1', '[::1]'].includes(hostname)
}

function resolveItickWsBaseUrl(baseUrl?: string) {
  if (baseUrl) return baseUrl

  const envBaseUrl = import.meta.env.VITE_API_BASE_URL
  if (!envBaseUrl) return window.location.origin

  try {
    const parsedEnvBaseUrl = new URL(envBaseUrl)
    if (isLoopbackHost(parsedEnvBaseUrl.hostname) && !isLoopbackHost(window.location.hostname)) {
      return window.location.origin
    }
  } catch {
    return window.location.origin
  }

  return envBaseUrl
}

export function buildItickWsUrl(id: string, baseUrl?: string) {
  const resolvedBaseUrl = resolveItickWsBaseUrl(baseUrl)
  const parsed = new URL(resolvedBaseUrl)
  const protocol = parsed.protocol === 'https:' ? 'wss:' : 'ws:'
  const apiBasePath = (import.meta.env.VITE_API_BASE_PATH || '/app').replace(/\/+$/, '')

  return `${protocol}//${parsed.host}${apiBasePath}/itick/ws/${id}`
}
