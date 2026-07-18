import axios, { AxiosHeaders, type InternalAxiosRequestConfig } from 'axios'

const VERSION = 'v1'
const LOCATION_JSON = 'JSON'
const LOCATION_QUERY = 'QUERY'
const CONFIG_PATH = '/admin/security/encryption-config'
const SESSION_PATH = '/admin/security/encryption-session'
const PROTECTED_PREFIXES = [
  '/admin/asset/',
  '/admin/itick/',
  '/admin/member/',
  '/admin/option/',
  '/admin/payment/',
  '/admin/staking/',
  '/admin/system/',
  '/admin/trade/',
]

type EncryptionConfig = {
  version: string
  mode: 'DISABLED' | 'OPTIONAL' | 'REQUIRED'
  enabled: boolean
  required: boolean
  rsaKid: string
  publicKey: string
  keyAlgorithm: string
  contentAlgorithm: string
  sessionTtlSeconds: number
  rotateBeforeSeconds: number
  serverTime: number
}

type EncryptionSession = {
  keyId: string
  expiresAt: number
  rotateAfter: number
}

type ApiResponse<T> = {
  code: number
  msg: string
  data?: T
}

type ActiveSession = EncryptionSession & {
  aesKey: CryptoKey
  rsaKid: string
}

export type EncryptionAxiosConfig = InternalAxiosRequestConfig & {
  __requestEncryptionPlainData?: unknown
  __requestEncryptionPlainParams?: unknown
  __requestEncryptionRetry?: boolean
}

const bootstrapHttp = axios.create()

let encryptionConfig: EncryptionConfig | undefined
let configPromise: Promise<EncryptionConfig> | undefined
let activeSession: ActiveSession | undefined
let sessionPromise: Promise<ActiveSession> | undefined
let serverClockOffset = 0

function toBase64Url(bytes: Uint8Array): string {
  let binary = ''
  const chunkSize = 0x8000
  for (let offset = 0; offset < bytes.length; offset += chunkSize) {
    binary += String.fromCharCode(...bytes.subarray(offset, offset + chunkSize))
  }
  return btoa(binary).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/g, '')
}

function fromBase64(value: string): ArrayBuffer {
  const binary = atob(value.replace(/\s/g, ''))
  const bytes = new Uint8Array(binary.length)
  for (let index = 0; index < binary.length; index += 1) bytes[index] = binary.charCodeAt(index)
  return bytes.buffer
}

function toArrayBuffer(bytes: Uint8Array): ArrayBuffer {
  return bytes.buffer.slice(bytes.byteOffset, bytes.byteOffset + bytes.byteLength) as ArrayBuffer
}

function randomBase64Url(size: number): string {
  return toBase64Url(crypto.getRandomValues(new Uint8Array(size)))
}

function adjustedNow(): number {
  return Date.now() + serverClockOffset
}

async function loadEncryptionConfig(baseURL?: string): Promise<EncryptionConfig> {
  if (encryptionConfig) return encryptionConfig
  if (!configPromise) {
    configPromise = bootstrapHttp
      .get<ApiResponse<EncryptionConfig>>(CONFIG_PATH, { baseURL })
      .then((response) => {
        if (response.data.code !== 200 || !response.data.data) {
          throw new Error(response.data.msg || 'Failed to load request encryption config')
        }
        encryptionConfig = response.data.data
        serverClockOffset = encryptionConfig.serverTime - Date.now()
        return encryptionConfig
      })
      .finally(() => {
        configPromise = undefined
      })
  }
  return configPromise
}

async function createSession(config: EncryptionConfig, baseURL?: string): Promise<ActiveSession> {
  if (!crypto?.subtle) throw new Error('Web Crypto API is unavailable')
  if (config.version !== VERSION || !config.rsaKid || !config.publicKey) {
    throw new Error('Invalid request encryption config')
  }

  const rsaKey = await crypto.subtle.importKey(
    'spki',
    fromBase64(config.publicKey),
    { name: 'RSA-OAEP', hash: 'SHA-256' },
    false,
    ['encrypt'],
  )
  const aesKey = await crypto.subtle.generateKey({ name: 'AES-GCM', length: 256 }, true, [
    'encrypt',
  ])
  const rawAESKey = await crypto.subtle.exportKey('raw', aesKey)
  const encryptedKey = await crypto.subtle.encrypt({ name: 'RSA-OAEP' }, rsaKey, rawAESKey)
  const response = await bootstrapHttp.post<ApiResponse<EncryptionSession>>(
    SESSION_PATH,
    {
      version: VERSION,
      rsaKid: config.rsaKid,
      encryptedKey: toBase64Url(new Uint8Array(encryptedKey)),
    },
    { baseURL },
  )
  if (response.data.code !== 200 || !response.data.data) {
    throw new Error(response.data.msg || 'Failed to create request encryption session')
  }
  return { ...response.data.data, aesKey, rsaKid: config.rsaKid }
}

async function ensureSession(config: EncryptionConfig, baseURL?: string): Promise<ActiveSession> {
  if (
    activeSession &&
    activeSession.rsaKid === config.rsaKid &&
    adjustedNow() < activeSession.rotateAfter
  ) {
    return activeSession
  }
  if (!sessionPromise) {
    sessionPromise = createSession(config, baseURL)
      .then((session) => {
        activeSession = session
        return session
      })
      .finally(() => {
        sessionPromise = undefined
      })
  }
  return sessionPromise
}

function requestLocation(
  config: InternalAxiosRequestConfig,
): typeof LOCATION_JSON | typeof LOCATION_QUERY | undefined {
  const method = (config.method || '').toUpperCase()
  const pathname = new URL(axios.getUri(config), window.location.origin).pathname
  if (!PROTECTED_PREFIXES.some((prefix) => pathname.startsWith(prefix))) return undefined
  if (['POST', 'PUT', 'PATCH'].includes(method)) return LOCATION_JSON
  if (['GET', 'DELETE'].includes(method)) return LOCATION_QUERY
  return undefined
}

function buildAAD(
  keyId: string,
  timestamp: string,
  nonce: string,
  method: string,
  bindingTarget: string,
  location: typeof LOCATION_JSON | typeof LOCATION_QUERY,
): Uint8Array {
  return new TextEncoder().encode(
    [VERSION, location, keyId, timestamp, nonce, method.toUpperCase(), bindingTarget].join('\n'),
  )
}

export async function encryptAxiosRequest(
  rawConfig: InternalAxiosRequestConfig,
): Promise<InternalAxiosRequestConfig> {
  const location = requestLocation(rawConfig)
  if (!location) return rawConfig
  if (typeof FormData !== 'undefined' && rawConfig.data instanceof FormData) {
    return rawConfig
  }

  const config = rawConfig as EncryptionAxiosConfig
  const encryption = await loadEncryptionConfig(config.baseURL)
  if (!encryption.enabled || encryption.mode === 'DISABLED') return config

  const session = await ensureSession(encryption, config.baseURL)
  const timestamp = String(Math.floor(adjustedNow() / 1000))
  const nonce = randomBase64Url(18)
  const iv = crypto.getRandomValues(new Uint8Array(12))
  const originalURL = new URL(axios.getUri(config), window.location.origin)
  const bindingTarget =
    location === LOCATION_QUERY
      ? `${originalURL.pathname}?data`
      : originalURL.pathname + originalURL.search
  const aad = buildAAD(
    session.keyId,
    timestamp,
    nonce,
    config.method || '',
    bindingTarget,
    location,
  )
  let plaintext: Uint8Array
  if (location === LOCATION_QUERY) {
    config.__requestEncryptionPlainParams ??= config.params
    plaintext = new TextEncoder().encode(originalURL.search.slice(1))
  } else {
    const plainData = config.__requestEncryptionPlainData ?? config.data ?? {}
    config.__requestEncryptionPlainData = plainData
    plaintext = new TextEncoder().encode(JSON.stringify(plainData))
  }
  const cipherText = await crypto.subtle.encrypt(
    {
      name: 'AES-GCM',
      iv: toArrayBuffer(iv),
      additionalData: toArrayBuffer(aad),
      tagLength: 128,
    },
    session.aesKey,
    toArrayBuffer(plaintext),
  )

  if (location === LOCATION_QUERY) {
    const cipherBytes = new Uint8Array(cipherText)
    const packed = new Uint8Array(iv.length + cipherBytes.length)
    packed.set(iv)
    packed.set(cipherBytes, iv.length)
    config.params = { data: toBase64Url(packed) }
  } else {
    config.data = {
      iv: toBase64Url(iv),
      cipherText: toBase64Url(new Uint8Array(cipherText)),
    }
  }
  config.headers = AxiosHeaders.from(config.headers)
  if (location === LOCATION_JSON) config.headers.set('Content-Type', 'application/json')
  config.headers.set('X-Encryption-Version', VERSION)
  config.headers.set('X-Encryption-Location', location)
  config.headers.set('X-Encryption-Key-Id', session.keyId)
  config.headers.set('X-Timestamp', timestamp)
  config.headers.set('X-Nonce', nonce)
  return config
}

export function resetRequestEncryptionSession(refetchConfig = false): void {
  activeSession = undefined
  sessionPromise = undefined
  if (refetchConfig) {
    encryptionConfig = undefined
    configPromise = undefined
  }
}

export function isEncryptionSessionExpiredError(error: unknown): boolean {
  if (!axios.isAxiosError(error)) return false
  return error.response?.data?.msg === 'ENCRYPTION_KEY_EXPIRED'
}
