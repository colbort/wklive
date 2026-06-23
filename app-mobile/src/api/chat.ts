import { http } from '@/api/http'
import {
  collectGuestFingerprint,
  createGuestFingerprintHash,
  getGuestDeviceId,
  setGuestDeviceId,
} from '@/utils/guestFingerprint'
import type { RespBase } from '@wklive/api/types/api'

export type ChatTokenData = {
  chatToken: string
  expireAt: number
  sessionNo: string
  chatUiUrl: string
  chatWsUrl?: string
}

export function apiCreateChatToken(): Promise<RespBase & { data: ChatTokenData }> {
  const fingerprint = collectGuestFingerprint()
  const deviceId = getGuestDeviceId() || `web_${createGuestFingerprintHash(fingerprint)}`
  if (!getGuestDeviceId()) setGuestDeviceId(deviceId)

  return http
    .post('/chat/token', {
      deviceId,
      fingerprint: JSON.stringify(fingerprint),
    })
    .then((res: { data: RespBase & { data: ChatTokenData } }) => res.data)
}
