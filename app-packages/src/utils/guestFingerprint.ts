import FingerprintJS from '@fingerprintjs/fingerprintjs'

export interface GuestFingerprint {
  visitorId: string
  version: string
  confidence: number
}

let fingerprintAgentPromise: ReturnType<typeof FingerprintJS.load> | undefined

function getFingerprintAgent() {
  fingerprintAgentPromise ??= FingerprintJS.load()
  return fingerprintAgentPromise
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
  } catch { /* empty */ }
}

export function setGuestId(guestId: string) {
  try {
    localStorage.setItem('guest_id', guestId || '')
  } catch { /* empty */ }
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
  } catch { /* empty */ }
}

export async function collectGuestFingerprint(): Promise<GuestFingerprint> {
  const agent = await getFingerprintAgent()
  const result = await agent.get()

  return {
    visitorId: result.visitorId,
    version: result.version,
    confidence: result.confidence.score,
  }
}

export function createGuestFingerprintHash(fingerprint: GuestFingerprint) {
  return fingerprint.visitorId
}
