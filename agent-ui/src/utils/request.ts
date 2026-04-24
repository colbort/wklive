/**
 * HTTP 请求工具（增强版）
 * 支持：请求拦截、响应拦截、错误处理、日志
 */

import axios, { AxiosRequestConfig, type AxiosInstance, type Method } from 'axios'
import { useAuthStore } from '@/stores/auth'
import type { RespBase } from '@/services'
import { ENV } from '@/config/environment'
import { logger } from '@/utils/logger'

export const http: AxiosInstance = axios.create({
  baseURL: ENV.API_BASE_URL,
  timeout: ENV.API_TIMEOUT,
  headers: {
    'Content-Type': 'application/json',
  },
})

http.interceptors.request.use(
  (config) => {
    const auth = useAuthStore()
    if (auth.token) {
      config.headers.Authorization = `Bearer ${auth.token}`
    }
    if (ENV.TENANT_CODE) {
      config.headers['X-Tenant-Code'] = ENV.TENANT_CODE
      const url = config.url || ''
      const shouldInjectTenantCode =
        !url.includes('/admin/system/auth/login') &&
        !url.includes('/admin/system/auth/profile') &&
        !url.includes('/admin/system/tenants')

      if (shouldInjectTenantCode) {
        if (config.params && typeof config.params === 'object' && config.params.tenantCode === undefined) {
          config.params = { ...config.params, tenantCode: ENV.TENANT_CODE }
        }
        if (
          config.data &&
          typeof config.data === 'object' &&
          !(config.data instanceof FormData) &&
          config.data.tenantCode === undefined
        ) {
          config.data = { ...config.data, tenantCode: ENV.TENANT_CODE }
        }
      }
    }
    logger.debug(`[${config.method?.toUpperCase()}] ${config.url || ''}`)
    return config
  },
  (error) => {
    logger.error('Request error', error)
    return Promise.reject(error)
  },
)

http.interceptors.response.use(
  (resp) => {
    logger.debug(`Response from ${resp.config.url || ''}`)
    return resp.data
  },
  (error) => {
    const response = error?.response

    if (response?.status === 401) {
      const auth = useAuthStore()
      auth.logout()
      logger.warn('Token expired, redirect to login')
      window.location.assign('/login')
    }

    logger.error(`HTTP ${response?.status || 'unknown'} ${error?.config?.url || ''}`, error)
    return Promise.reject(error)
  },
)

function request<T = any>(
  method: Method,
  url: string,
  options?: {
    params?: any
    data?: any
    config?: AxiosRequestConfig
  },
): Promise<RespBase<T>> {
  const { params, data, config } = options ?? {}
  return http.request({
    method,
    url,
    params,
    data,
    ...(config ?? {}),
  })
}

export function get<T = any>(
  url: string,
  params?: any,
  config?: AxiosRequestConfig,
): Promise<RespBase<T>> {
  return request<T>('GET', url, { params, config })
}

export function post<T = any>(
  url: string,
  data?: any,
  config?: AxiosRequestConfig,
): Promise<RespBase<T>> {
  return request<T>('POST', url, { data, config })
}

export function put<T = any>(
  url: string,
  data?: any,
  config?: AxiosRequestConfig,
): Promise<RespBase<T>> {
  return request<T>('PUT', url, { data, config })
}

export function del<T = any>(
  url: string,
  params?: any,
  config?: AxiosRequestConfig,
): Promise<RespBase<T>> {
  return request<T>('DELETE', url, { params, config })
}

export function patch<T = any>(
  url: string,
  data?: any,
  config?: AxiosRequestConfig,
): Promise<RespBase<T>> {
  return request<T>('PATCH', url, { data, config })
}

export default http
