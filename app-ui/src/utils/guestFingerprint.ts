export interface GuestFingerprint {
  deviceLocalId: string
  timezone: string
  platform: string
  language: string
  languages: readonly string[]
  userAgent: string
  browserName: string
  browserVersion: string
  browserMajorVersion: string
  osName: string
  deviceType: string
  screenWidth: number
  screenHeight: number
  availWidth: number
  availHeight: number
  innerWidth: number
  innerHeight: number
  colorDepth: number
  pixelRatio: number
  hardwareConcurrency: number
  deviceMemory: number
  maxTouchPoints: number
  cookieEnabled: boolean
  onLine: boolean
  localStorageSupported: boolean
  sessionStorageSupported: boolean
  indexedDBSupported: boolean
}

function safeNumber(value: unknown, defaultValue = 0) {
  return typeof value === 'number' && !Number.isNaN(value) ? value : defaultValue
}

function createLocalId() {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return crypto.randomUUID()
  }

  return `fallback_${Date.now()}_${Math.random().toString(36).slice(2, 10)}`
}

function getScreenInfo() {
  const screenObj = (window.screen || {}) as Partial<Screen>

  return {
    screenWidth: safeNumber(screenObj.width),
    screenHeight: safeNumber(screenObj.height),
    availWidth: safeNumber(screenObj.availWidth),
    availHeight: safeNumber(screenObj.availHeight),
    colorDepth: safeNumber(screenObj.colorDepth),
    pixelRatio: safeNumber(window.devicePixelRatio, 1),
    innerWidth: safeNumber(window.innerWidth),
    innerHeight: safeNumber(window.innerHeight),
  }
}

function getNavigatorInfo() {
  const nav = (window.navigator || {}) as Navigator
  const memoryNavigator = nav as Navigator & { deviceMemory?: number }

  return {
    userAgent: nav.userAgent || '',
    platform: nav.platform || '',
    language: nav.language || '',
    languages: Array.isArray(nav.languages) ? nav.languages : [],
    hardwareConcurrency: safeNumber(nav.hardwareConcurrency),
    deviceMemory: safeNumber(memoryNavigator.deviceMemory),
    maxTouchPoints: safeNumber(nav.maxTouchPoints),
    cookieEnabled: Boolean(nav.cookieEnabled),
    onLine: Boolean(nav.onLine),
  }
}

function getTimezone() {
  try {
    return Intl.DateTimeFormat().resolvedOptions().timeZone || ''
  } catch {
    return ''
  }
}

function getStorageSupport() {
  let localStorageSupported = false
  let sessionStorageSupported = false
  let indexedDBSupported = false

  try {
    localStorageSupported = typeof window.localStorage !== 'undefined'
  } catch {}

  try {
    sessionStorageSupported = typeof window.sessionStorage !== 'undefined'
  } catch {}

  try {
    indexedDBSupported = typeof window.indexedDB !== 'undefined'
  } catch {}

  return {
    localStorageSupported,
    sessionStorageSupported,
    indexedDBSupported,
  }
}

function getDeviceType(userAgent = '', maxTouchPoints = 0, screenWidth = 0) {
  const ua = userAgent.toLowerCase()

  if (/ipad|tablet/.test(ua)) return 'tablet'
  if (/iphone|android.+mobile|mobile/.test(ua)) return 'mobile'
  if (/android/.test(ua) && maxTouchPoints > 0 && screenWidth <= 1024) return 'tablet'
  return 'desktop'
}

function parseUserAgentInfo(userAgent = '') {
  const ua = userAgent.toLowerCase()

  let browserName = 'unknown'
  let browserVersion = ''
  let osName = 'unknown'

  if (ua.includes('edg/')) {
    browserName = 'edge'
    browserVersion = ua.match(/edg\/([\d.]+)/)?.[1] || ''
  } else if (ua.includes('chrome/') && !ua.includes('edg/')) {
    browserName = 'chrome'
    browserVersion = ua.match(/chrome\/([\d.]+)/)?.[1] || ''
  } else if (ua.includes('safari/') && ua.includes('version/')) {
    browserName = 'safari'
    browserVersion = ua.match(/version\/([\d.]+)/)?.[1] || ''
  } else if (ua.includes('firefox/')) {
    browserName = 'firefox'
    browserVersion = ua.match(/firefox\/([\d.]+)/)?.[1] || ''
  }

  if (ua.includes('windows')) {
    osName = 'windows'
  } else if (ua.includes('mac os')) {
    osName = 'macos'
  } else if (ua.includes('android')) {
    osName = 'android'
  } else if (ua.includes('iphone') || ua.includes('ipad') || ua.includes('ios')) {
    osName = 'ios'
  } else if (ua.includes('linux')) {
    osName = 'linux'
  }

  const browserMajorVersion = browserVersion ? browserVersion.split('.')[0] : ''

  return {
    browserName,
    browserVersion,
    browserMajorVersion,
    osName,
  }
}

function stableStringify(value: unknown): string {
  if (Array.isArray(value)) {
    return `[${value.map((item) => stableStringify(item)).join(',')}]`
  }
  if (value && typeof value === 'object') {
    return `{${Object.keys(value)
      .sort()
      .map((key) => `${JSON.stringify(key)}:${stableStringify((value as Record<string, unknown>)[key])}`)
      .join(',')}}`
  }
  return JSON.stringify(value) ?? 'null'
}

function hashString(value: string) {
  let hash = 0x811c9dc5
  for (let index = 0; index < value.length; index += 1) {
    hash ^= value.charCodeAt(index)
    hash = Math.imul(hash, 0x01000193)
  }
  return `fp_${(hash >>> 0).toString(16).padStart(8, '0')}`
}

export function getOrCreateDeviceLocalId() {
  const key = 'device_local_id'

  try {
    let value = localStorage.getItem(key)
    if (value) return value

    value = createLocalId()
    localStorage.setItem(key, value)
    return value
  } catch {
    return createLocalId()
  }
}

export function getGuestId() {
  try {
    return localStorage.getItem('guest_id') || ''
  } catch {
    return ''
  }
}

export function getGuestDeviceId() {
  try {
    return localStorage.getItem('guest_device_id') || ''
  } catch {
    return ''
  }
}

export function setGuestDeviceId(deviceId: string) {
  try {
    localStorage.setItem('guest_device_id', deviceId || '')
  } catch {}
}

export function setGuestId(guestId: string) {
  try {
    localStorage.setItem('guest_id', guestId || '')
  } catch {}
}

export function getGuestToken() {
  try {
    return localStorage.getItem('guest_token') || ''
  } catch {
    return ''
  }
}

export function setGuestToken(token: string) {
  try {
    localStorage.setItem('guest_token', token || '')
  } catch {}
}

export function collectGuestFingerprint(): GuestFingerprint {
  const navigatorInfo = getNavigatorInfo()
  const screenInfo = getScreenInfo()
  const timezone = getTimezone()
  const storageInfo = getStorageSupport()
  const uaInfo = parseUserAgentInfo(navigatorInfo.userAgent)
  const deviceType = getDeviceType(
    navigatorInfo.userAgent,
    navigatorInfo.maxTouchPoints,
    screenInfo.screenWidth,
  )

  return {
    deviceLocalId: getOrCreateDeviceLocalId(),

    timezone,

    platform: navigatorInfo.platform,
    language: navigatorInfo.language,
    languages: navigatorInfo.languages,

    userAgent: navigatorInfo.userAgent,
    browserName: uaInfo.browserName,
    browserVersion: uaInfo.browserVersion,
    browserMajorVersion: uaInfo.browserMajorVersion,
    osName: uaInfo.osName,
    deviceType,

    screenWidth: screenInfo.screenWidth,
    screenHeight: screenInfo.screenHeight,
    availWidth: screenInfo.availWidth,
    availHeight: screenInfo.availHeight,
    innerWidth: screenInfo.innerWidth,
    innerHeight: screenInfo.innerHeight,
    colorDepth: screenInfo.colorDepth,
    pixelRatio: screenInfo.pixelRatio,

    hardwareConcurrency: navigatorInfo.hardwareConcurrency,
    deviceMemory: navigatorInfo.deviceMemory,
    maxTouchPoints: navigatorInfo.maxTouchPoints,

    cookieEnabled: navigatorInfo.cookieEnabled,
    onLine: navigatorInfo.onLine,

    localStorageSupported: storageInfo.localStorageSupported,
    sessionStorageSupported: storageInfo.sessionStorageSupported,
    indexedDBSupported: storageInfo.indexedDBSupported,
  }
}

export function createGuestFingerprintHash(fingerprint = collectGuestFingerprint()) {
  return hashString(stableStringify(fingerprint))
}
