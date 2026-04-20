import { http } from '@/api/http'
import { compactParams } from '@/api/utils'
import type { ApiResp } from '@/types/api'
import type {
  BatchGetQuoteReq,
  BatchGetQuoteResp,
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

export async function apiBatchGetQuote(
  params: BatchGetQuoteReq,
): Promise<ApiResp<BatchGetQuoteResp>> {
  const { data } = await http.get<ApiResp<BatchGetQuoteResp>>('/itick/batch/quote', {
    params: {
      ...compactParams({
        categoryCode: params.categoryCode,
        market: params.market,
      }),
      data: JSON.stringify(params.data),
    },
  })
  return data
}

export function buildItickWsUrl(baseUrl?: string) {
  // const resolvedBaseUrl = baseUrl || import.meta.env.VITE_API_BASE_URL || window.location.origin
  // const parsed = new URL(resolvedBaseUrl)
  // const protocol = parsed.protocol === 'https:' ? 'wss:' : 'ws:'

  // return `${protocol}//${parsed.host}/app/user/ws/itick`
  return 'ws://127.0.0.1:6666/app/itick/ws/itick'
}
