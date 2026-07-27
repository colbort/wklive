import { getApiBasePath, getApiBaseUrl } from './http'

function isLoopbackHost(hostname: string) {
  return ['localhost', '127.0.0.1', '0.0.0.0', '::1', '[::1]'].includes(hostname)
}

export function buildUserEventStreamUrl(id: string, baseUrl?: string) {
  let resolvedBaseUrl = baseUrl || getApiBaseUrl() || window.location.origin
  try {
    const parsedBaseUrl = new URL(resolvedBaseUrl)
    if (isLoopbackHost(parsedBaseUrl.hostname) && !isLoopbackHost(window.location.hostname)) {
      resolvedBaseUrl = window.location.origin
    }
  } catch {
    resolvedBaseUrl = window.location.origin
  }
  const parsed = new URL(resolvedBaseUrl)
  const apiBasePath = getApiBasePath().replace(/\/+$/, '')
  return `${parsed.protocol}//${parsed.host}${apiBasePath}/events/stream/${encodeURIComponent(id)}`
}
