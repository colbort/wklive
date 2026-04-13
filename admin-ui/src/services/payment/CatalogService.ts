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
  id: number // 平台ID
  platformCode: string // 平台编码
  platformName: string // 平台名称
  platformType: number // 平台类型：1三方支付 2银行转账 3链上支付 4人工代收
  notifyUrl: string // 统一异步通知地址
  returnUrl: string // 默认同步跳转地址
  icon: string // 图标
  status: number // 状态：1启用 2停用
  remark: string // 备注
  createTimes: number // 创建时间
  updateTimes: number // 更新时间
}

export type PayProduct = {
  id: number // 产品ID
  platformId: number // 平台ID
  productCode: string // 产品编码
  productName: string // 产品名称
  sceneType: number // 场景：1APP 2H5 3WEB 4收银台 5链上
  currency: string // 币种
  status: number // 状态：1启用 2停用
  remark: string // 备注
  createTimes: number // 创建时间
  updateTimes: number // 更新时间
}

export type CreatePayPlatformReq = {
  platformCode: string // 平台编码
  platformName: string // 平台名称
  platformType: number // 平台类型：1三方支付 2银行转账 3链上支付 4人工代收
  notifyUrl?: string // 统一异步通知地址
  returnUrl?: string // 默认同步跳转地址
  icon?: string // 图标
  status: number // 状态：1启用 2停用
  remark?: string // 备注
}

export type UpdatePayPlatformReq = {
  id: number // 平台ID
  platformName: string // 平台名称
  platformType: number // 平台类型：1三方支付 2银行转账 3链上支付 4人工代收
  notifyUrl?: string // 统一异步通知地址
  returnUrl?: string // 默认同步跳转地址
  icon?: string // 图标
  status: number // 状态：1启用 2停用
  remark?: string // 备注
}

export type ListPayPlatformsReq = {
  keyword?: string // 关键字
  platformCode?: string // 平台编码
  status?: number // 状态：1启用 2停用
  platformType?: number // 平台类型
  cursor?: string | null // 分页游标
  limit?: number // 分页大小
}

export type CreatePayProductReq = {
  platformId: number // 平台ID
  productCode: string // 产品编码
  productName: string // 产品名称
  sceneType: number // 场景：1APP 2H5 3WEB 4收银台 5链上
  currency: string // 币种
  status: number // 状态：1启用 2停用
  remark?: string // 备注
}

export type UpdatePayProductReq = {
  id: number // 产品ID
  productName: string // 产品名称
  sceneType: number // 场景：1APP 2H5 3WEB 4收银台 5链上
  currency: string // 币种
  status: number // 状态：1启用 2停用
  remark?: string // 备注
}

export type ListPayProductsReq = {
  platformId?: number // 平台ID
  keyword?: string // 关键字
  productCode?: string // 产品编码
  status?: number // 状态：1启用 2停用
  sceneType?: number // 场景类型
  cursor?: string | null // 分页游标
  limit?: number // 分页大小
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
