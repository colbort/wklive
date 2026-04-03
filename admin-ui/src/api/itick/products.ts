import { get, post, put, del } from '@/utils/request'

import type {
  RespBase,
  ListProductsReq,
  ItickProduct,
  CreateProductReq,
  UpdateProductReq,
  GetProductKlineReq,
  Kline,
} from '@/services'


export function apiItickProductList(params: ListProductsReq): Promise<RespBase<ItickProduct[]>> {
    return get<ItickProduct[]>(`/admin/itick/products`, params)
}

export function apiItickProductCreate(params: CreateProductReq): Promise<RespBase> {
    return post(`/admin/itick/products`, params)
}

export function apiItickProductUpdate(params: UpdateProductReq): Promise<RespBase> {
    return put(`/admin/itick/products`, params)
}

export function apiItickProductDetail(id: number): Promise<RespBase<ItickProduct>> {
    return get<ItickProduct>(`/admin/itick/products/${id}`)
}

export function apiItickProductKline(params: GetProductKlineReq): Promise<RespBase<Kline[]>> {
    return get<Kline[]>(`/admin/itick/product/kline`, params)
}
