import type { RespBase } from '@/services'
import {
  apiPayPlatformCreate,
  apiPayPlatformDetail,
  apiPayPlatformList,
  apiPayPlatformUpdate,
  apiPayProductCreate,
  apiPayProductDetail,
  apiPayProductList,
  apiPayProductUpdate,
} from '@/api/payment/catalog'

export type PayPlatform = {
  id: number
  platformCode: string
  platformName: string
  platformType: number
  notifyUrl: string
  returnUrl: string
  icon: string
  status: number
  remark: string
  createTimes: number
  updateTimes: number
}

export type PayProduct = {
  id: number
  platformId: number
  productCode: string
  productName: string
  sceneType: number
  currency: string
  status: number
  remark: string
  createTimes: number
  updateTimes: number
}

export type CreatePayPlatformReq = {
  platformCode: string
  platformName: string
  platformType: number
  notifyUrl?: string
  returnUrl?: string
  icon?: string
  status: number
  remark?: string
}

export type UpdatePayPlatformReq = {
  id: number
  platformName: string
  platformType: number
  notifyUrl?: string
  returnUrl?: string
  icon?: string
  status: number
  remark?: string
}

export type ListPayPlatformsReq = {
  keyword?: string
  platformCode?: string
  status?: number
  platformType?: number
  cursor?: string | null
  limit?: number
}

export type CreatePayProductReq = {
  platformId: number
  productCode: string
  productName: string
  sceneType: number
  currency: string
  status: number
  remark?: string
}

export type UpdatePayProductReq = {
  id: number
  productName: string
  sceneType: number
  currency: string
  status: number
  remark?: string
}

export type ListPayProductsReq = {
  platformId?: number
  keyword?: string
  productCode?: string
  status?: number
  sceneType?: number
  cursor?: string | null
  limit?: number
}

export class CatalogService {
  async getPlatformList(params: ListPayPlatformsReq): Promise<RespBase<PayPlatform[]>> {
    return apiPayPlatformList(params)
  }

  async getPlatformDetail(id: number): Promise<RespBase<PayPlatform>> {
    return apiPayPlatformDetail(id)
  }

  async createPlatform(params: CreatePayPlatformReq): Promise<RespBase> {
    return apiPayPlatformCreate(params)
  }

  async updatePlatform(id: string | number, params: Omit<UpdatePayPlatformReq, 'id'>): Promise<RespBase> {
    return apiPayPlatformUpdate({ id: Number(id), ...params })
  }

  async getProductList(params: ListPayProductsReq): Promise<RespBase<PayProduct[]>> {
    return apiPayProductList(params)
  }

  async getProductDetail(id: number): Promise<RespBase<PayProduct>> {
    return apiPayProductDetail(id)
  }

  async createProduct(params: CreatePayProductReq): Promise<RespBase> {
    return apiPayProductCreate(params)
  }

  async updateProduct(id: string | number, params: Omit<UpdatePayProductReq, 'id'>): Promise<RespBase> {
    return apiPayProductUpdate({ id: Number(id), ...params })
  }
}

export const catalogService = new CatalogService()
