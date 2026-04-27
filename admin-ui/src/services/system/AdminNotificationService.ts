import { ElNotification } from 'element-plus'
import { ENV } from '@/config/environment'
import { useAuthStore } from '@/stores'
import { logger } from '@/utils/logger'

export const ADMIN_NOTIFICATION_EVENT_TYPES = {
  USER_IDENTITY_SUBMIT: 'user_identity_submit',
  RECHARGE: 'recharge',
  WITHDRAW: 'withdraw',
} as const

export type AdminNotificationEventType =
  (typeof ADMIN_NOTIFICATION_EVENT_TYPES)[keyof typeof ADMIN_NOTIFICATION_EVENT_TYPES]

export type AdminNotificationLevel = 'info' | 'warning' | 'error'

export type AdminNotificationEvent = {
  id: string
  type: AdminNotificationEventType | string
  level: AdminNotificationLevel
  title: string
  message: string
  source?: string
  tenantId?: number
  userId?: number
  bizNo?: string
  data?: Record<string, unknown>
  createdAt: number
}

type AudioWindow = Window &
  typeof globalThis & {
    webkitAudioContext?: typeof AudioContext
  }

const reconnectBaseDelay = 1000
const reconnectMaxDelay = 30000
const unlockEvents = ['click', 'keydown', 'touchstart'] as const
const voiceTextMap: Record<AdminNotificationEventType, string> = {
  [ADMIN_NOTIFICATION_EVENT_TYPES.USER_IDENTITY_SUBMIT]: '有新的实名认证提交',
  [ADMIN_NOTIFICATION_EVENT_TYPES.RECHARGE]: '有新的充值订单',
  [ADMIN_NOTIFICATION_EVENT_TYPES.WITHDRAW]: '有新的提现订单',
}

function buildWebsocketUrl(token: string) {
  const url = new URL('/admin/ws/notifications', ENV.API_BASE_URL)
  url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:'
  url.searchParams.set('token', token)
  return url.toString()
}

function isActionEvent(type: string) {
  return Object.values(ADMIN_NOTIFICATION_EVENT_TYPES).includes(type as AdminNotificationEventType)
}

class AdminNotificationService {
  private socket: WebSocket | null = null
  private reconnectTimer: number | null = null
  private reconnectAttempts = 0
  private manuallyClosed = false
  private audioContext: AudioContext | null = null
  private audioUnlocked = false
  private unlockHandler = () => {
    this.unlockAudio()
  }

  connect() {
    const auth = useAuthStore()
    if (!auth.token) return
    if (
      this.socket &&
      (this.socket.readyState === WebSocket.OPEN || this.socket.readyState === WebSocket.CONNECTING)
    ) {
      return
    }

    this.manuallyClosed = false
    this.clearReconnectTimer()
    this.bindUnlockEvents()

    const socket = new WebSocket(buildWebsocketUrl(auth.token))
    this.socket = socket

    socket.onopen = () => {
      this.reconnectAttempts = 0
      logger.info('Admin notification websocket connected')
    }

    socket.onmessage = (message) => {
      this.handleMessage(message.data)
    }

    socket.onerror = (error) => {
      logger.warn('Admin notification websocket error', error)
    }

    socket.onclose = () => {
      if (this.socket === socket) {
        this.socket = null
      }
      this.scheduleReconnect()
    }
  }

  disconnect() {
    this.manuallyClosed = true
    this.clearReconnectTimer()
    this.unbindUnlockEvents()

    if (this.socket) {
      this.socket.close()
      this.socket = null
    }

    if (this.audioContext) {
      this.audioContext.close().catch((error) => {
        logger.warn('Close admin notification audio context failed', error)
      })
      this.audioContext = null
    }
    this.audioUnlocked = false
  }

  private handleMessage(raw: string) {
    let event: AdminNotificationEvent
    try {
      event = JSON.parse(raw)
    } catch (error) {
      logger.warn('Invalid admin notification payload', error)
      return
    }

    if (!isActionEvent(event.type)) return

    this.playVoiceReminder(event)

    ElNotification({
      title: event.title || '管理通知',
      message: event.message,
      type: event.level || 'info',
      duration: 8000,
      showClose: true,
    })
  }

  private playVoiceReminder(event: AdminNotificationEvent) {
    this.playPromptTone()

    if (!('speechSynthesis' in window)) return

    try {
      const text =
        voiceTextMap[event.type as AdminNotificationEventType] || event.title || event.message || '有新的管理通知'
      const utterance = new SpeechSynthesisUtterance(text)
      utterance.lang = 'zh-CN'
      utterance.rate = 1
      utterance.pitch = 1
      utterance.volume = 1
      window.speechSynthesis.cancel()
      window.speechSynthesis.speak(utterance)
    } catch (error) {
      logger.warn('Play admin notification voice failed', error)
    }
  }

  private playPromptTone() {
    const audioContext = this.getAudioContext()
    if (!audioContext) return

    try {
      const oscillator = audioContext.createOscillator()
      const gain = audioContext.createGain()
      const now = audioContext.currentTime

      oscillator.type = 'sine'
      oscillator.frequency.setValueAtTime(880, now)
      oscillator.frequency.setValueAtTime(1175, now + 0.16)
      gain.gain.setValueAtTime(0.001, now)
      gain.gain.exponentialRampToValueAtTime(0.16, now + 0.02)
      gain.gain.exponentialRampToValueAtTime(0.001, now + 0.36)

      oscillator.connect(gain)
      gain.connect(audioContext.destination)
      oscillator.start(now)
      oscillator.stop(now + 0.38)
    } catch (error) {
      logger.warn('Play admin notification tone failed', error)
    }
  }

  private bindUnlockEvents() {
    if (this.audioUnlocked) return

    unlockEvents.forEach((eventName) => {
      window.addEventListener(eventName, this.unlockHandler, { passive: true })
    })
  }

  private unbindUnlockEvents() {
    unlockEvents.forEach((eventName) => {
      window.removeEventListener(eventName, this.unlockHandler)
    })
  }

  private unlockAudio() {
    const audioContext = this.getAudioContext()
    if (!audioContext) return

    audioContext
      .resume()
      .then(() => {
        this.audioUnlocked = true
        this.unbindUnlockEvents()
      })
      .catch((error) => {
        logger.warn('Unlock admin notification audio failed', error)
      })
  }

  private getAudioContext() {
    const AudioContext = window.AudioContext || (window as AudioWindow).webkitAudioContext
    if (!AudioContext) return null

    if (!this.audioContext) {
      this.audioContext = new AudioContext()
    }

    if (this.audioContext.state === 'suspended' && this.audioUnlocked) {
      this.audioContext.resume().catch((error) => {
        logger.warn('Resume admin notification audio context failed', error)
      })
    }

    return this.audioContext
  }

  private scheduleReconnect() {
    if (this.manuallyClosed) return

    const auth = useAuthStore()
    if (!auth.token) return

    const delay = Math.min(reconnectBaseDelay * 2 ** this.reconnectAttempts, reconnectMaxDelay)
    this.reconnectAttempts += 1
    this.reconnectTimer = window.setTimeout(() => {
      this.reconnectTimer = null
      this.connect()
    }, delay)
  }

  private clearReconnectTimer() {
    if (this.reconnectTimer !== null) {
      window.clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }
  }
}

export const adminNotificationService = new AdminNotificationService()
