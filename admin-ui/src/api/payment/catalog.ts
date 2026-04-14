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
  return get<PayPlatform[]>('/admin/payment/platforms/list', params)
}

export function apiPayPlatformDetail(id: number): Promise<RespBase<PayPlatform>> {
  return get<PayPlatform>('/admin/payment/platforms', { id })
}

export function apiPayPlatformCreate(params: CreatePayPlatformReq): Promise<RespBase> {
  return post('/admin/payment/platforms', params)
}

export function apiPayPlatformUpdate(params: UpdatePayPlatformReq): Promise<RespBase> {
  return put('/admin/payment/platforms', params)
}

export function apiPayProductList(params: ListPayProductsReq): Promise<RespBase<PayProduct[]>> {
  return get<PayProduct[]>('/admin/payment/products/list', params)
}

export function apiPayProductDetail(id: number): Promise<RespBase<PayProduct>> {
  return get<PayProduct>('/admin/payment/products', { id })
}

export function apiPayProductCreate(params: CreatePayProductReq): Promise<RespBase> {
  return post('/admin/payment/products', params)
}

export function apiPayProductUpdate(params: UpdatePayProductReq): Promise<RespBase> {
  return put('/admin/payment/products', params)
}
