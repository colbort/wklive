import { h } from 'vue'
import { ElButton, ElNotification } from 'element-plus'
import { ENV } from '@/config/environment'
import { useAuthStore } from '@/stores/auth'
import { logger } from '@/utils/logger'

export const ADMIN_NOTIFICATION_EVENT_TYPES = {
  USER_IDENTITY_SUBMIT: 'user_identity_submit',
  RECHARGE: 'recharge',
  WITHDRAW: 'withdraw',
  CONTRACT_RECONCILIATION: 'contract_reconciliation',
  PRICE_ENGINE_INPUT: 'price_engine_input',
  SNAPSHOT_OUTBOX: 'snapshot_outbox',
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

type AdminNotificationAckResult = {
  action: 'ack_result'
  ok: boolean
  eventType?: string
  alertKey?: string
  tenantId?: number
  acknowledgedAt?: number
  acknowledgedBy?: number
  error?: string
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
  [ADMIN_NOTIFICATION_EVENT_TYPES.CONTRACT_RECONCILIATION]: '发现合约对账差异',
  [ADMIN_NOTIFICATION_EVENT_TYPES.PRICE_ENGINE_INPUT]: '价格引擎输入异常',
  [ADMIN_NOTIFICATION_EVENT_TYPES.SNAPSHOT_OUTBOX]: '行情快照队列异常',
}

function buildWebsocketUrl() {
  const apiBase = new URL(ENV.API_BASE_URL, window.location.origin)
  const url = new URL('/admin/ws/notifications', apiBase)
  url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:'
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

    const socket = new WebSocket(buildWebsocketUrl(), [
      'wklive-admin-notifications',
      `bearer.${auth.token}`,
    ])
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
    let payload: AdminNotificationEvent | AdminNotificationAckResult
    try {
      payload = JSON.parse(raw)
    } catch (error) {
      logger.warn('Invalid admin notification payload', error)
      return
    }

    if ('action' in payload && payload.action === 'ack_result') {
      ElNotification({
        title: payload.ok ? '告警已确认' : '告警确认失败',
        message: payload.ok ? '确认回执已保存，后续自动升级已停止。' : payload.error || '请重试。',
        type: payload.ok ? 'success' : 'error',
        duration: 5000,
      })
      return
    }

    const event = payload as AdminNotificationEvent
    if (!isActionEvent(event.type)) return

    this.playVoiceReminder(event)

    ElNotification({
      title: event.title || '管理通知',
      message: this.notificationMessage(event),
      type: event.level || 'info',
      duration: this.requiresAcknowledgement(event) ? 0 : 8000,
      showClose: true,
    })
  }

  private notificationMessage(event: AdminNotificationEvent) {
    if (!this.requiresAcknowledgement(event)) return event.message

    return h('div', { class: 'admin-notification-message' }, [
      h('div', event.message),
      h(
        ElButton,
        {
          type: 'primary',
          size: 'small',
          style: 'margin-top: 10px',
          onClick: () => this.acknowledge(event),
        },
        () => '确认收到',
      ),
    ])
  }

  private requiresAcknowledgement(event: AdminNotificationEvent) {
    const operational =
      event.type === ADMIN_NOTIFICATION_EVENT_TYPES.CONTRACT_RECONCILIATION ||
      event.type === ADMIN_NOTIFICATION_EVENT_TYPES.PRICE_ENGINE_INPUT ||
      event.type === ADMIN_NOTIFICATION_EVENT_TYPES.SNAPSHOT_OUTBOX
    return operational && event.data?.state === 'firing' && Boolean(this.alertKey(event))
  }

  private alertKey(event: AdminNotificationEvent) {
    if (event.bizNo) return event.bizNo
    const value = event.data?.alertKey
    return typeof value === 'string' ? value : ''
  }

  private acknowledge(event: AdminNotificationEvent) {
    if (!this.socket || this.socket.readyState !== WebSocket.OPEN) {
      ElNotification.error({
        title: '告警确认失败',
        message: '通知连接已断开，请等待重连后重试。',
      })
      return
    }
    this.socket.send(
      JSON.stringify({
        action: 'ack',
        eventType: event.type,
        alertKey: this.alertKey(event),
        tenantId: event.tenantId || 0,
        reason: '后台值班人员通过 Admin WebSocket 手动确认',
      }),
    )
  }

  private playVoiceReminder(event: AdminNotificationEvent) {
    this.playPromptTone()

    if (!('speechSynthesis' in window)) return

    try {
      const text =
        voiceTextMap[event.type as AdminNotificationEventType] ||
        event.title ||
        event.message ||
        '有新的管理通知'
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
