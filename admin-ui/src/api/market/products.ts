import { get, post, put } from '@/utils/request'

import type {
  RespBase,
  ListProductsReq,
  MarketProduct,
  CreateProductReq,
  UpdateProductReq,
  GetProductKlineReq,
  Kline,
  SyncProductKlineHistoryReq,
  SyncProductKlineHistoryResp,
} from '@/services'

export function apiMarketProductList(params: ListProductsReq): Promise<RespBase<MarketProduct[]>> {
  return get<MarketProduct[]>('/admin/market/products', params)
}

export function apiSyncMarketProductKlineHistory(
  params: SyncProductKlineHistoryReq,
): Promise<SyncProductKlineHistoryResp> {
  return post(
    '/admin/market/product/kline/sync-history',
    params,
  ) as Promise<SyncProductKlineHistoryResp>
}

export function apiMarketProductCreate(params: CreateProductReq): Promise<RespBase> {
  return post('/admin/market/products', params)
}

export function apiMarketProductUpdate(params: UpdateProductReq): Promise<RespBase> {
  return put('/admin/market/products', params)
}

export function apiMarketProductDetail(id: number): Promise<RespBase<MarketProduct>> {
  return get<MarketProduct>(`/admin/market/products/${id}`)
}

export function apiMarketProductKline(params: GetProductKlineReq): Promise<RespBase<Kline[]>> {
  return get<Kline[]>('/admin/market/product/kline', params)
}
