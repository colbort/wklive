import axios, { type AxiosError, type AxiosResponse, type InternalAxiosRequestConfig } from 'axios'

const ACCESS_TOKEN_KEY = 'app_access_token'
const REFRESH_TOKEN_KEY = 'app_refresh_token'
const TENANT_ID_KEY = 'app_tenant_id'
const TENANT_CODE_KEY = 'app_tenant_code'
const DEFAULT_API_BASE_PATH = '/app'

function trimTrailingSlash(value: string) {
  return value.replace(/\/+$/, '')
}

function normalizeLeadingSlash(value: string) {
  return value.startsWith('/') ? value : `/${value}`
}

function resolveApiBaseURL() {
  const apiBaseUrl = import.meta.env.VITE_API_BASE_URL?.trim()??''
  const apiBasePath = normalizeLeadingSlash(import.meta.env.VITE_API_BASE_PATH || DEFAULT_API_BASE_PATH)
  const target = import.meta.env.VITE_APP_TARGET

  if (!apiBaseUrl || !target) return apiBasePath

  return `${trimTrailingSlash(apiBaseUrl)}${apiBasePath}`
}

export const http = axios.create({
  baseURL: resolveApiBaseURL(),
  timeout: 10000,
})

const refreshHttp = axios.create({
  baseURL: resolveApiBaseURL(),
  timeout: 10000,
})

type RetryableRequestConfig = InternalAxiosRequestConfig & {
  _retry?: boolean
}

let refreshTokenTask: Promise<string | null> | null = null

export function getAccessToken() {
  return localStorage.getItem(ACCESS_TOKEN_KEY)
}

export function setAccessToken(token: string) {
  localStorage.setItem(ACCESS_TOKEN_KEY, token)
}

export function clearAccessToken() {
  localStorage.removeItem(ACCESS_TOKEN_KEY)
}

export function getRefreshToken() {
  return localStorage.getItem(REFRESH_TOKEN_KEY)
}

export function setRefreshToken(token: string) {
  localStorage.setItem(REFRESH_TOKEN_KEY, token)
}

export function clearRefreshToken() {
  localStorage.removeItem(REFRESH_TOKEN_KEY)
}

export function getTenantCode() {
  return localStorage.getItem(TENANT_CODE_KEY) || import.meta.env.VITE_TENANT_CODE || ''
}

export function getTenantId() {
  return localStorage.getItem(TENANT_ID_KEY) || import.meta.env.VITE_TENANT_ID || ''
}

export function setTenantId(tenantId: string | number) {
  localStorage.setItem(TENANT_ID_KEY, String(tenantId))
}

export function clearTenantId() {
  localStorage.removeItem(TENANT_ID_KEY)
}

export function setTenantCode(tenantCode: string) {
  localStorage.setItem(TENANT_CODE_KEY, tenantCode)
}

export function clearTenantCode() {
  localStorage.removeItem(TENANT_CODE_KEY)
}

function clearAuthTokens() {
  clearAccessToken()
  clearRefreshToken()
}

function isPlainObject(value: unknown): value is Record<string, unknown> {
  return Object.prototype.toString.call(value) === '[object Object]'
}

function appendTenantScope(target: Record<string, unknown>, url?: string) {
  const tenantCode = getTenantCode()

  const needsTenantCode =
    url?.startsWith('/user/login') ||
    url?.startsWith('/user/register') ||
    url?.startsWith('/user/refresh-token')

  if (needsTenantCode && tenantCode && target.tenantCode === undefined) {
    target.tenantCode = tenantCode
  }

  return target
}

function stripUserTenantScope(value: unknown): unknown {
  if (Array.isArray(value)) {
    return value.map((item) => stripUserTenantScope(item))
  }

  if (!isPlainObject(value)) {
    return value
  }

  return Object.fromEntries(
    Object.entries(value)
      .filter(([key]) => key !== 'userId' && key !== 'tenantId')
      .map(([key, childValue]) => [key, stripUserTenantScope(childValue)]),
  )
}

http.interceptors.request.use((config) => {
  const accessToken = getAccessToken()
  if (accessToken) {
    config.headers.Authorization = `Bearer ${accessToken}`
  }

  const tenantId = getTenantId()
  const tenantCode = getTenantCode()
  if (tenantId) {
    config.headers['x-tenant-id'] = tenantId
  }
  if (tenantCode) {
    config.headers['x-tenant-code'] = tenantCode
  }

  if (isPlainObject(config.params)) {
    config.params = stripUserTenantScope(appendTenantScope({ ...config.params }, config.url))
  }

  if (isPlainObject(config.data)) {
    config.data = stripUserTenantScope(appendTenantScope({ ...config.data }, config.url))
  }

  return config
})

function isRefreshTokenRequest(url?: string) {
  return Boolean(url?.startsWith('/user/refresh-token'))
}

function getResponseCode(data: unknown) {
  if (!isPlainObject(data)) return undefined
  return Number(data.code)
}

function readTokenPayload(data: unknown) {
  if (!isPlainObject(data)) return null

  const directToken = data.token
  if (isPlainObject(directToken)) return directToken

  const nestedData = data.data
  if (isPlainObject(nestedData) && isPlainObject(nestedData.token)) return nestedData.token

  return null
}

function refreshAccessToken() {
  if (!refreshTokenTask) {
    refreshTokenTask = (async () => {
      const refreshToken = getRefreshToken()
      if (!refreshToken) return null

      const tenantCode = getTenantCode()
      const res = await refreshHttp.post('/user/refresh-token', {
        refreshToken,
        ...(tenantCode ? { tenantCode } : {}),
      })
      const token = readTokenPayload(res.data)
      const accessToken = typeof token?.accessToken === 'string' ? token.accessToken : ''
      const nextRefreshToken = typeof token?.refreshToken === 'string' ? token.refreshToken : ''

      if (!accessToken) return null
      setAccessToken(accessToken)
      if (nextRefreshToken) setRefreshToken(nextRefreshToken)
      return accessToken
    })().finally(() => {
      refreshTokenTask = null
    })
  }

  return refreshTokenTask
}

async function retryWithRefreshedToken(config?: InternalAxiosRequestConfig) {
  const originalConfig = config as RetryableRequestConfig | undefined
  if (!originalConfig || originalConfig._retry || isRefreshTokenRequest(originalConfig.url)) {
    clearAuthTokens()
    return Promise.reject(new Error('Unauthorized'))
  }

  originalConfig._retry = true
  const accessToken = await refreshAccessToken()
  if (!accessToken) {
    clearAuthTokens()
    return Promise.reject(new Error('Unauthorized'))
  }

  originalConfig.headers = originalConfig.headers || {}
  originalConfig.headers.Authorization = `Bearer ${accessToken}`
  return http(originalConfig)
}

http.interceptors.response.use(
  (response: AxiosResponse) => {
    if (getResponseCode(response.data) === 401) {
      return retryWithRefreshedToken(response.config)
    }
    return response
  },
  (error: AxiosError) => {
    if (error.response?.status === 401) {
      return retryWithRefreshedToken(error.config)
    }
    return Promise.reject(error)
  },
)
