import axios, { AxiosRequestConfig, Method } from 'axios'
import { useAuthStore } from '@/stores/auth'
import { ApiResp } from '@/api/types'

export const http = axios.create({
  baseURL: 'http://localhost:8888', // ✅ admin-api:8888
  timeout: 15000,
})

// ---------------- 请求拦截 ----------------
http.interceptors.request.use((config) => {
  const auth = useAuthStore()
  if (auth.token) config.headers.Authorization = `Bearer ${auth.token}`
  return config
})

// ---------------- 响应拦截 ----------------
http.interceptors.response.use(
  (resp) => resp.data,
  (err) => {
    if (err?.response?.status === 401) {
      const auth = useAuthStore()
      auth.logout()
      location.href = '/login'
    }
    return Promise.reject(err)
  },
)


// ---------------- 通用 request ----------------
function request<T = any>(
  method: Method,
  url: string,
  options?: {
    params?: any
    data?: any
    config?: AxiosRequestConfig
  },
): Promise<ApiResp<T>> {
  const { params, data, config } = options ?? {}
  return http.request({
    method,
    url,
    params,
    data,
    ...(config ?? {}),
  })
}

// ---------------- 基础方法 ----------------

export function get<T = any>(
  url: string,
  params?: any,
  config?: AxiosRequestConfig,
) : Promise<ApiResp<T>> {
  return request<T>('GET', url, { params, config })
}

export function post<T = any>(
  url: string,
  data?: any,
  config?: AxiosRequestConfig,
) : Promise<ApiResp<T>> {
  return request<T>('POST', url, { data, config })
}

export function put<T = any>(
  url: string,
  data?: any,
  config?: AxiosRequestConfig,
) : Promise<ApiResp<T>> {
  return request<T>('PUT', url, { data, config })
}

export function del<T = any>(
  url: string,
  params?: any,
  config?: AxiosRequestConfig,
) : Promise<ApiResp<T>> {
  return request<T>('DELETE', url, { params, config })
}
