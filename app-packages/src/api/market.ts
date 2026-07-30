import { getApiBasePath, getApiBaseUrl, http } from './http'
import { compactParams } from './utils'
import type { ApiResp, RespBase } from '../types/api'
import type {
  GetKlineReq,
  GetKlineResp,
  GetQuoteReq,
  GetQuoteResp,
  ListVisibleCategoriesReq,
  ListVisibleCategoriesResp,
  ListVisibleProductsReq,
  ListVisibleProductsResp,
} from '../types/market'

export async function apiListVisibleCategories(
  params: ListVisibleCategoriesReq,
): Promise<ApiResp<ListVisibleCategoriesResp>> {
  const { data } = await http.get<ApiResp<ListVisibleCategoriesResp>>('/market/categories', {
    params: compactParams(params),
  })
  return data
}

export async function apiListVisibleProducts(
  params: ListVisibleProductsReq,
): Promise<ApiResp<ListVisibleProductsResp>> {
  const { data } = await http.get<ApiResp<ListVisibleProductsResp>>('/market/products', {
    params: compactParams(params),
  })
  return data
}

export async function apiGetKline(params: GetKlineReq): Promise<ApiResp<GetKlineResp>> {
  const { data } = await http.get<ApiResp<GetKlineResp>>('/market/kline', {
    params: compactParams(params),
  })
  return data
}

export async function apiGetQuote(params: GetQuoteReq): Promise<ApiResp<GetQuoteResp>> {
  const { data } = await http.get<ApiResp<GetQuoteResp>>('/market/quote', {
    params: compactParams(params),
  })
  return data
}

function isLoopbackHost(hostname: string) {
  return ['localhost', '127.0.0.1', '0.0.0.0', '::1', '[::1]'].includes(hostname)
}

function resolveMarketWsBaseUrl(baseUrl?: string) {
  if (baseUrl) return baseUrl

  const envBaseUrl = getApiBaseUrl()
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

export function buildMarketWsUrl(id: string, baseUrl?: string) {
  const resolvedBaseUrl = resolveMarketWsBaseUrl(baseUrl)
  const parsed = new URL(resolvedBaseUrl)
  const protocol = parsed.protocol === 'https:' ? 'wss:' : 'ws:'
  const apiBasePath = getApiBasePath().replace(/\/+$/, '')

  return `${protocol}//${parsed.host}${apiBasePath}/market/ws/${id}`
}
