/**
 * HTTP 请求工具（增强版）
 * 支持：请求拦截、响应拦截、错误处理、日志
 */

import axios, { AxiosRequestConfig, AxiosInstance, Method } from 'axios'
import { useAuthStore } from '@/stores'
import { RespBase } from '@/services'
import { logger } from '@/utils/logger'

export const http: AxiosInstance = axios.create({
  baseURL: 'http://localhost:8888', // ✅ admin-api:8888
  timeout: 15000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// ==================== 请求拦截器 ====================
http.interceptors.request.use(
  (config) => {
    const auth = useAuthStore()
    if (auth.token) {
      config.headers.Authorization = `Bearer ${auth.token}`
    }
    logger.debug(`📤 [${config.method?.toUpperCase()}] ${config.url}`)
    return config
  },
  (error) => {
    logger.error('Request error:', error)
    return Promise.reject(error)
  },
)

// ==================== 响应拦截器 ====================
http.interceptors.response.use(
  (resp) => {
    logger.debug(`✅ Response from ${resp.config.url}`)
    return resp.data
  },
  (error) => {
    const { response } = error
    if (error?.response?.status === 401) {
      const auth = useAuthStore()
      auth.logout()
      logger.warn('Token expired, redirect to login')
      location.href = '/login'
    }
    logger.error(`❌ [${response?.status}] ${error.config?.url}`)
    return Promise.reject(error)
  },
)

// ==================== 通用 request ==================
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

// ==================== 基础方法 ====================

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
