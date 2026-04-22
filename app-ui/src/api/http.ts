import axios from 'axios'

const ACCESS_TOKEN_KEY = 'app_access_token'
const REFRESH_TOKEN_KEY = 'app_refresh_token'
const TENANT_ID_KEY = 'app_tenant_id'
const TENANT_CODE_KEY = 'app_tenant_code'

export const http = axios.create({
  baseURL: '/app',
  timeout: 10000,
})

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

export function setTenantCode(tenantCode: string) {
  localStorage.setItem(TENANT_CODE_KEY, tenantCode)
}

export function clearTenantCode() {
  localStorage.removeItem(TENANT_CODE_KEY)
}

function isPlainObject(value: unknown): value is Record<string, unknown> {
  return Object.prototype.toString.call(value) === '[object Object]'
}

function appendTenantScope(target: Record<string, unknown>, url?: string) {
  // const tenantId = getTenantId()
  const tenantCode = getTenantCode()

  // if (tenantId && target.tenantId === undefined) {
  //   target.tenantId = tenantId
  // }

  const needsTenantCode =
    url?.startsWith('/user/login') ||
    url?.startsWith('/user/register') ||
    url?.startsWith('/user/refresh-token')

  if (needsTenantCode && tenantCode && target.tenantCode === undefined) {
    target.tenantCode = tenantCode
  }

  return target
}

http.interceptors.request.use((config) => {
  const accessToken = getAccessToken()
  if (accessToken) {
    config.headers.Authorization = `Bearer ${accessToken}`
  }

  if (isPlainObject(config.params)) {
    config.params = appendTenantScope({ ...config.params }, config.url)
  }

  if (isPlainObject(config.data)) {
    config.data = appendTenantScope({ ...config.data }, config.url)
  }

  return config
})
