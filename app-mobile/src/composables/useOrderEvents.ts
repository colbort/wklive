import { onBeforeUnmount, onMounted } from 'vue'

import { getAccessToken } from '@/api/http'
import { buildUserEventStreamUrl } from '@/api/userEvent'

export type PrivateOrderEvent = {
  id: string
  type: string
  domain: 'trade' | 'option' | 'staking' | string
  tenant_id: number
  user_id: number
  symbol_id?: number
  product_type?: number
  biz_id?: number
  biz_no?: string
  occurred_at: number
}

type UserEventStreamMessage = {
  type: string
  data?: PrivateOrderEvent
}

const INITIAL_RECONNECT_DELAY = 1000
const MAX_RECONNECT_DELAY = 15000

export function useOrderEvents(
  onOrderChanged: (event: PrivateOrderEvent) => void,
  onConnected?: () => void,
) {
  let controller: AbortController | null = null
  let reconnectTimer: number | undefined
  let reconnectDelay = INITIAL_RECONNECT_DELAY
  let disposed = false

  function connectionId() {
    if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
      return crypto.randomUUID()
    }
    return `events-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 10)}`
  }

  async function connect() {
    stopReconnect()
    closeStream()
    const token = getAccessToken()
    if (!token || disposed || document.visibilityState !== 'visible') return

    const nextController = new AbortController()
    let shouldReconnect = true
    controller = nextController
    try {
      const response = await fetch(buildUserEventStreamUrl(connectionId()), {
        method: 'GET',
        headers: {
          Accept: 'text/event-stream',
          Authorization: `Bearer ${token}`,
        },
        cache: 'no-store',
        signal: nextController.signal,
      })
      if (response.status === 401 || response.status === 403) {
        shouldReconnect = false
      }
      if (!response.ok || !response.body) {
        throw new Error(`user event stream failed: ${response.status}`)
      }

      await readEventStream(response.body, (payload) => {
        if (controller !== nextController) return
        if (payload.type === 'order.changed' && payload.data) {
          onOrderChanged(payload.data)
        } else if (payload.type === 'connected') {
          reconnectDelay = INITIAL_RECONNECT_DELAY
          onConnected?.()
        }
      })
    } catch (error) {
      if (!(error instanceof DOMException && error.name === 'AbortError')) {
        // HTTP refresh remains the source of truth while the stream reconnects.
      }
    } finally {
      if (controller === nextController) {
        controller = null
        if (shouldReconnect) scheduleReconnect()
      }
    }
  }

  function reconnect() {
    reconnectDelay = INITIAL_RECONNECT_DELAY
    void connect()
  }

  function scheduleReconnect() {
    if (disposed || reconnectTimer !== undefined || document.visibilityState !== 'visible') return
    reconnectTimer = window.setTimeout(() => {
      reconnectTimer = undefined
      void connect()
    }, reconnectDelay)
    reconnectDelay = Math.min(MAX_RECONNECT_DELAY, reconnectDelay * 2)
  }

  function stopReconnect() {
    if (reconnectTimer !== undefined) {
      window.clearTimeout(reconnectTimer)
      reconnectTimer = undefined
    }
  }

  function closeStream() {
    if (!controller) return
    const current = controller
    controller = null
    current.abort()
  }

  function handleVisibilityChange() {
    if (document.visibilityState === 'visible') {
      reconnect()
    } else {
      stopReconnect()
      closeStream()
    }
  }

  onMounted(() => {
    disposed = false
    document.addEventListener('visibilitychange', handleVisibilityChange)
    void connect()
  })

  onBeforeUnmount(() => {
    disposed = true
    document.removeEventListener('visibilitychange', handleVisibilityChange)
    stopReconnect()
    closeStream()
  })

  return { reconnect }
}

async function readEventStream(
  stream: ReadableStream<Uint8Array>,
  onMessage: (message: UserEventStreamMessage) => void,
) {
  const reader = stream.getReader()
  const decoder = new TextDecoder()
  let buffer = ''

  while (true) {
    const { done, value } = await reader.read()
    buffer += decoder.decode(value, { stream: !done }).replace(/\r\n/g, '\n')

    let boundary = buffer.indexOf('\n\n')
    while (boundary >= 0) {
      const block = buffer.slice(0, boundary)
      buffer = buffer.slice(boundary + 2)
      const data = block
        .split('\n')
        .filter((line) => line.startsWith('data:'))
        .map((line) => line.slice(5).trimStart())
        .join('\n')
      if (data) {
        try {
          onMessage(JSON.parse(data) as UserEventStreamMessage)
        } catch {
          // Ignore malformed events and continue reading the stream.
        }
      }
      boundary = buffer.indexOf('\n\n')
    }

    if (done) return
  }
}
