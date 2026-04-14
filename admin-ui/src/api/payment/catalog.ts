import { get, post, put } from '@/utils/request'
import type {
  RespBase,
  PayPlatformItem,
  PayPlatform,
  PayProduct,
  CreatePayPlatformReq,
  UpdatePayPlatformReq,
  ListPayPlatformsReq,
  CreatePayProductReq,
  UpdatePayProductReq,
  ListPayProductsReq,
} from '@/services'

export function apiPayPlatforms(): Promise<RespBase<PayPlatformItem[]>> {
  return get<PayPlatformItem[]>('/admin/payment/platforms/options')
}

export function apiPayPlatformList(params: ListPayPlatformsReq): Promise<RespBase<PayPlatform[]>> {
  return get<PayPlatform[]>('/admin/payment/platforms', params)
}

export function apiPayPlatformDetail(id: number): Promise<RespBase<PayPlatform>> {
  return get<PayPlatform>('/admin/payment/platform', { id })
}

export function apiPayPlatformCreate(params: CreatePayPlatformReq): Promise<RespBase> {
  return post('/admin/payment/platform', params)
}

export function apiPayPlatformUpdate(params: UpdatePayPlatformReq): Promise<RespBase> {
  return put('/admin/payment/platform', params)
}

export function apiPayProductList(params: ListPayProductsReq): Promise<RespBase<PayProduct[]>> {
  return get<PayProduct[]>('/admin/payment/products', params)
}

export function apiPayProductDetail(id: number): Promise<RespBase<PayProduct>> {
  return get<PayProduct>('/admin/payment/product', { id })
}

export function apiPayProductCreate(params: CreatePayProductReq): Promise<RespBase> {
  return post('/admin/payment/product', params)
}

export function apiPayProductUpdate(params: UpdatePayProductReq): Promise<RespBase> {
  return put('/admin/payment/product', params)
}
