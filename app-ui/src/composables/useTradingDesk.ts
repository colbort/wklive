import { computed, onBeforeUnmount, onMounted, ref, watch, type ComputedRef } from 'vue'

import { buildItickWsUrl } from '@/api/itick'
import { getTenantCode } from '@/api/http'
import { useItickStore } from '@/stores/itick'
import { useSystemStore } from '@/stores/system'
import type { Interval } from '@/types/core'
import type {
  DepthLevel,
  DepthPayload,
  ItickTenantProduct,
  ItickWsConnectionState,
  ItickWsPongMessage,
  ItickWsServerMessage,
  ItickWsSubscribeMessage,
  ItickWsTopicConfig,
  KlinePayload,
  QuotePayload,
  TickPayload,
} from '@/types/itick'

const DEFAULT_INTERVAL = '1m'
const DEFAULT_K_TYPE = 1
const KLINE_LIMIT = 180
const PING_INTERVAL_MS = 20000
const RECONNECT_DELAY_MS = 2500

export function useTradingDesk(options: {
  detailVisible: ComputedRef<boolean>
  tickLimit?: number
}) {
  const store = useItickStore()
  const systemStore = useSystemStore()

  const tickLimit = options.tickLimit ?? 12
  const selectedCategoryType = ref<number | null>(null)
  const selectedProductKey = ref('')
  const selectedIntervalName = ref(DEFAULT_INTERVAL)
  const loadingBootstrap = ref(false)
  const loadingKline = ref(false)
  const depthSnapshot = ref<DepthPayload | null>(null)
  const tickSnapshot = ref<TickPayload[]>([])
  const klineSnapshot = ref<KlinePayload[]>([])
  const quoteMap = ref<Record<string, QuotePayload>>({})
  const wsState = ref<ItickWsConnectionState>('closed')
  const wsError = ref('')
  const viewingLatestKlinePage = ref(true)
  const wsId = ref('')

  let socket: WebSocket | null = null
  let pingTimer: number | undefined
  let reconnectTimer: number | undefined
  let refreshSocketTimer: number | undefined
  let isUnmounted = false
  let reconnectEnabled = true

  const tenantCode = computed(() => getTenantCode())
  const categories = computed(() => store.categories)
  const products = computed(() => store.products)
  const intervals = computed(() => {
    return systemStore.systemCore.intervals.length
      ? systemStore.systemCore.intervals
      : [{ name: DEFAULT_INTERVAL, kType: DEFAULT_K_TYPE }]
  })
  const selectedCategory = computed(
    () => categories.value.find((item) => item.categoryType === selectedCategoryType.value) ?? null,
  )
  const selectedCategoryCode = computed(() => selectedCategory.value?.categoryCode || '')
  const selectedProduct = computed(
    () => products.value.find((item) => productKey(item) === selectedProductKey.value) ?? products.value[0] ?? null,
  )
  const selectedQuote = computed(() => {
    const product = selectedProduct.value
    return product ? quoteMap.value[productKey(product)] ?? null : null
  })
  const selectedInterval = computed(() => {
    return intervals.value.find((item) => item.name === selectedIntervalName.value) ?? intervals.value[0]
  })
  const placeholderPrice = computed(() => {
    const value = selectedQuote.value?.lastPrice
    if (!value || !Number.isFinite(value)) return '--'
    return formatPrice(value)
  })
  const placeholderChange = computed(() => {
    const quote = selectedQuote.value
    if (!quote || !quote.open) return ''
    return formatPercent(((quote.lastPrice - quote.open) / quote.open) * 100)
  })
  const priceTrend = computed<'up' | 'down'>(() => {
    const quote = selectedQuote.value
    if (!quote || quote.lastPrice >= quote.open) return 'up'
    return 'down'
  })
  const marketRows = computed(() =>
    products.value.map((product) => {
      const quote = quoteMap.value[productKey(product)] ?? null
      const changeRate = getChangeRate(quote)

      return {
        key: productKey(product),
        product,
        quote,
        changeRate,
        direction: changeRate > 0 ? ('up' as const) : changeRate < 0 ? ('down' as const) : ('flat' as const),
      }
    }),
  )
  const desktopProductRows = computed(() =>
    products.value.map((product) => {
      const quote = quoteMap.value[productKey(product)] ?? null
      const changeRate = getChangeRate(quote)

      return {
        key: productKey(product),
        product,
        price: quote ? formatPrice(quote.lastPrice) : '--',
        change: quote ? formatPercent(changeRate) : '--',
        direction: (quote ? (changeRate > 0 ? 'up' : changeRate < 0 ? 'down' : 'flat') : 'flat') as 'up' | 'down' | 'flat',
      }
    }),
  )
  const productSheetRows = computed(() =>
    products.value.map((product) => {
      const key = productKey(product)
      const isSelected = key === selectedProductKey.value
      const quote = quoteMap.value[key] ?? null
      const changeRate = getChangeRate(quote)
      const changeValue = quote ? quote.lastPrice - quote.open : null

      return {
        key,
        product,
        price: quote ? formatPrice(quote.lastPrice) : '--',
        change: quote && changeValue !== null ? `${formatPrice(changeValue)} ${formatPercent(changeRate)}` : '',
        direction: (quote
          ? changeRate > 0
            ? 'up'
            : changeRate < 0
              ? 'down'
              : 'flat'
          : isSelected
            ? priceTrend.value
            : 'flat') as 'up' | 'down' | 'flat',
      }
    }),
  )
  const desktopStats = computed(() => [
    {
      label: '24小时涨跌',
      value: selectedQuote.value
        ? `${formatPrice((selectedQuote.value.lastPrice - selectedQuote.value.open) || 0)}  ${formatPercent(getChangeRate(selectedQuote.value))}`
        : '--',
      down: getChangeRate(selectedQuote.value) < 0,
    },
    { label: '今开', value: formatPrice(selectedQuote.value?.open) },
    { label: '昨收', value: formatPrice(selectedQuote.value?.open) },
    { label: '24小时最高价', value: formatPrice(selectedQuote.value?.high) },
    { label: '24小时最低价', value: formatPrice(selectedQuote.value?.low) },
    { label: '成交额', value: formatCompact(selectedQuote.value?.turnover) },
    { label: '成交量', value: formatCompact(selectedQuote.value?.volume) },
  ])

  watch(selectedCategoryType, async (categoryType) => {
    if (categoryType === null) return
    await loadProducts(categoryType)
  })

  watch(products, (list) => {
    if (!list.length) {
      selectedProductKey.value = ''
      quoteMap.value = {}
      depthSnapshot.value = null
      tickSnapshot.value = []
      klineSnapshot.value = []
      queueSocketRefresh()
      return
    }

    const hasSelected = list.some((item) => productKey(item) === selectedProductKey.value)
    if (!hasSelected) {
      selectedProductKey.value = productKey(list[0])
    }

    queueSocketRefresh()
  }, { deep: true })

  watch(intervals, (list) => {
    if (!list.length) return
    const hasSelected = list.some((item) => item.name === selectedIntervalName.value)
    if (!hasSelected) {
      selectedIntervalName.value = list[0].name
    }
  }, { immediate: true })

  watch(selectedProductKey, async () => {
    depthSnapshot.value = null
    tickSnapshot.value = []
    klineSnapshot.value = []
    if (options.detailVisible.value) {
      await loadSelectedKlinePage(Date.now(), true)
    }
    queueSocketRefresh()
  })

  watch(selectedIntervalName, async () => {
    if (!options.detailVisible.value || !selectedProduct.value) return
    klineSnapshot.value = []
    await loadSelectedKlinePage(Date.now(), true)
    queueSocketRefresh()
  })

  watch(options.detailVisible, async (visible) => {
    queueSocketRefresh()
    if (visible && selectedProduct.value && !klineSnapshot.value.length) {
      await loadSelectedKlinePage(Date.now(), true)
    }
  }, { immediate: true })

  onMounted(async () => {
    await initialize()
  })

  onBeforeUnmount(() => {
    isUnmounted = true
    reconnectEnabled = false
    stopReconnectTimer()
    stopRefreshSocketTimer()
    closeSocket()
  })

  async function initialize() {
    loadingBootstrap.value = true
    try {
      await systemStore.loadSystemCore()
      const list = await store.listVisibleCategories({
        tenantCode: tenantCode.value,
        limit: 20,
      })
      if (selectedCategoryType.value === null) {
        selectedCategoryType.value = list[0]?.categoryType ?? null
      }
    } finally {
      loadingBootstrap.value = false
    }
  }

  async function loadProducts(categoryType: number) {
    loadingBootstrap.value = true
    try {
      await store.listVisibleProducts({
        tenantCode: tenantCode.value,
        categoryType,
        categoryCode: selectedCategory.value?.categoryCode,
        limit: 200,
      })
    } finally {
      loadingBootstrap.value = false
    }
  }

  async function loadSelectedKlinePage(endTs: number, latestPage = false) {
    const product = selectedProduct.value
    if (!product) return

    loadingKline.value = true
    try {
      const categoryCode = product.categoryCode || selectedCategoryCode.value
      if (!categoryCode) return

      const klines = await store.getKline({
        categoryCode,
        market: product.market,
        symbol: product.symbol,
        kType: selectedInterval.value?.kType ?? DEFAULT_K_TYPE,
        endTs,
        limit: KLINE_LIMIT,
      })

      klineSnapshot.value = normalizeKlineList(klines)
      viewingLatestKlinePage.value = latestPage
    } finally {
      loadingKline.value = false
    }
  }

  async function loadPreviousKlinePage() {
    if (loadingKline.value || !klineSnapshot.value.length) return
    const sortedItems = [...klineSnapshot.value].sort((left, right) => right.ts - left.ts)
    const lastItem = sortedItems[sortedItems.length - 1]
    if (!lastItem?.ts) return
    await loadSelectedKlinePage(lastItem.ts, false)
  }

  function selectCategory(categoryType: number) {
    if (selectedCategoryType.value === categoryType) return
    selectedCategoryType.value = categoryType
  }

  function selectProduct(product: ItickTenantProduct) {
    selectedProductKey.value = productKey(product)
  }

  function selectInterval(interval: Interval) {
    if (selectedIntervalName.value === interval.name) return
    selectedIntervalName.value = interval.name
  }

  function queueSocketRefresh() {
    stopRefreshSocketTimer()
    refreshSocketTimer = window.setTimeout(() => {
      if (!products.value.length) {
        closeSocket()
        return
      }
      connectSocket()
    }, 80)
  }

  function createWsId() {
    if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
      return crypto.randomUUID()
    }
    return `ws-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 10)}`
  }

  function connectSocket() {
    wsId.value = createWsId()
    const url = buildItickWsUrl(wsId.value)

    stopReconnectTimer()
    reconnectEnabled = false
    closeSocket()

    wsState.value = 'connecting'
    wsError.value = ''

    const nextSocket = new WebSocket(url)
    socket = nextSocket

    nextSocket.addEventListener('open', () => {
      if (socket !== nextSocket) return
      reconnectEnabled = true
      wsState.value = 'open'
      sendQuoteSubscription()
      sendSelectedDetailSubscription()
      startPingLoop()
    })

    nextSocket.addEventListener('message', (event) => {
      if (socket !== nextSocket) return
      handleSocketMessage(event.data)
    })

    nextSocket.addEventListener('error', () => {
      if (socket !== nextSocket) return
      wsError.value = '实时连接异常，正在尝试恢复。'
    })

    nextSocket.addEventListener('close', () => {
      if (socket !== nextSocket) return
      wsState.value = 'closed'
      stopPingLoop()

      if (!isUnmounted && reconnectEnabled) {
        reconnectTimer = window.setTimeout(() => {
          connectSocket()
        }, RECONNECT_DELAY_MS)
      }
    })
  }

  function closeSocket() {
    stopPingLoop()
    if (socket) {
      const current = socket
      socket = null
      if (current.readyState === WebSocket.OPEN || current.readyState === WebSocket.CONNECTING) {
        current.close()
      }
    }
  }

  function startPingLoop() {
    stopPingLoop()
    pingTimer = window.setInterval(() => {
      sendJson({
        type: 'ping',
        clientTs: Date.now(),
      })
    }, PING_INTERVAL_MS)
  }

  function stopPingLoop() {
    if (pingTimer !== undefined) {
      window.clearInterval(pingTimer)
      pingTimer = undefined
    }
  }

  function stopReconnectTimer() {
    if (reconnectTimer !== undefined) {
      window.clearTimeout(reconnectTimer)
      reconnectTimer = undefined
    }
  }

  function stopRefreshSocketTimer() {
    if (refreshSocketTimer !== undefined) {
      window.clearTimeout(refreshSocketTimer)
      refreshSocketTimer = undefined
    }
  }

  function sendQuoteSubscription() {
    const topics = products.value
      .map((product) => productTopic(product, 'quote'))
      .filter((topic): topic is ItickWsTopicConfig => Boolean(topic))

    if (!topics.length) return

    sendJson({
      type: 'subscribe',
      topics: dedupeTopics(topics),
    })
  }

  function sendSelectedDetailSubscription() {
    if (!options.detailVisible.value) return
    const selected = selectedProduct.value
    if (!selected) return

    const depthTopic = productTopic(selected, 'depth')
    const tickTopic = productTopic(selected, 'tick')
    const klineTopic = productTopic(selected, 'kline')
    if (!depthTopic || !tickTopic || !klineTopic) return

    sendJson({
      type: 'subscribe',
      topics: [
        depthTopic,
        tickTopic,
        {
          ...klineTopic,
          interval: selectedInterval.value?.name || DEFAULT_INTERVAL,
        },
      ],
    })
  }

  function sendJson(payload: ItickWsSubscribeMessage | { type: 'ping'; clientTs: number }) {
    if (!socket || socket.readyState !== WebSocket.OPEN) return
    socket.send(JSON.stringify(payload))
  }

  function handleSocketMessage(raw: string) {
    try {
      const message = JSON.parse(raw) as ItickWsPongMessage | ItickWsServerMessage<unknown>
      if ('type' in message && message.type === 'pong') return
      if (!('topic' in message)) return

      const targetKey = productKeyByFields(message.market || '', message.symbol)
      const currentKey = selectedProduct.value ? productKey(selectedProduct.value) : ''

      switch (message.topic) {
        case 'quote':
          quoteMap.value = {
            ...quoteMap.value,
            [targetKey]: normalizeQuotePayload(message.payload),
          }
          break
        case 'tick':
          if (targetKey === currentKey) {
            tickSnapshot.value = [normalizeTickPayload(message.payload), ...tickSnapshot.value].slice(0, tickLimit)
          }
          break
        case 'depth':
          if (targetKey === currentKey) {
            depthSnapshot.value = normalizeDepthPayload(message.payload)
          }
          break
        case 'kline':
          if (targetKey === currentKey && viewingLatestKlinePage.value) {
            const kline = normalizeKlinePayload(message.payload, message.interval || DEFAULT_INTERVAL)
            klineSnapshot.value = mergeKlines(klineSnapshot.value, kline)
          }
          break
      }
    } catch (error) {
      console.error('handle ws message failed', error)
    }
  }

  function productKey(product: Pick<ItickTenantProduct, 'market' | 'symbol'>) {
    return productKeyByFields(product.market, product.symbol)
  }

  function productKeyByFields(market: string, symbol: string) {
    return `${String(market || '').toUpperCase()}::${String(symbol || '').toUpperCase()}`
  }

  function productTopic(product: ItickTenantProduct, topic: ItickWsTopicConfig['topic']) {
    const categoryCode = product.categoryCode || selectedCategoryCode.value
    if (!categoryCode || !product.market || !product.symbol) return null

    return {
      topic,
      categoryCode,
      market: product.market,
      symbol: product.symbol,
    }
  }

  function dedupeTopics(items: ItickWsTopicConfig[]) {
    const seen = new Set<string>()
    return items.filter((item) => {
      const key = [item.topic, item.categoryCode, item.market, item.symbol, item.interval || ''].join('::')
      if (seen.has(key)) return false
      seen.add(key)
      return true
    })
  }

  function coinGlyph(product: ItickTenantProduct) {
    const coin = product.baseCoin || product.symbol.slice(0, 3) || product.displayName
    return coin.slice(0, 1).toUpperCase()
  }

  return {
    selectedCategoryType,
    selectedProductKey,
    selectedIntervalName,
    loadingBootstrap,
    loadingKline,
    depthSnapshot,
    tickSnapshot,
    klineSnapshot,
    wsState,
    wsError,
    categories,
    products,
    intervals,
    selectedCategory,
    selectedCategoryCode,
    selectedProduct,
    selectedQuote,
    placeholderPrice,
    placeholderChange,
    priceTrend,
    marketRows,
    desktopProductRows,
    productSheetRows,
    desktopStats,
    initialize,
    loadPreviousKlinePage,
    selectCategory,
    selectProduct,
    selectInterval,
    productKey,
    coinGlyph,
  }
}

function normalizeQuotePayload(payload: unknown): QuotePayload {
  const data = asRecord(payload)
  return {
    lastPrice: toNumber(data.lastPrice),
    open: toNumber(data.open),
    high: toNumber(data.high),
    low: toNumber(data.low),
    volume: toNumber(data.volume),
    turnover: toNumber(data.turnover),
    ts: toNumber(data.ts),
  }
}

function normalizeTickPayload(payload: unknown): TickPayload {
  const data = asRecord(payload)
  return {
    lastPrice: toNumber(data.lastPrice),
    volume: toNumber(data.volume),
    ts: toNumber(data.ts),
  }
}

function normalizeDepthPayload(payload: unknown): DepthPayload {
  const data = asRecord(payload)
  return {
    asks: normalizeDepthLevels(data.asks ?? data.Asks ?? data.ASKS),
    bids: normalizeDepthLevels(data.bids ?? data.Bids ?? data.BIDS),
  }
}

function normalizeDepthLevels(source: unknown): DepthLevel[] {
  if (!Array.isArray(source)) return []
  return source
    .map((item) => {
      const level = asRecord(item)
      return {
        price: toNumber(level.p ?? level.price ?? level.P ?? level.Price),
        volume: toNumber(level.v ?? level.volume ?? level.V ?? level.Volume),
        position: toNumber(level.po ?? level.position ?? level.Po ?? level.Position),
        originVolume: toNumber(level.o ?? level.originVolume ?? level.O ?? level.OriginVolume),
      }
    })
    .filter((item) => item.price > 0)
}

function normalizeKlinePayload(payload: unknown, interval: string): KlinePayload {
  const data = asRecord(payload)
  return {
    interval,
    open: toNumber(data.open),
    high: toNumber(data.high),
    low: toNumber(data.low),
    close: toNumber(data.close),
    volume: toNumber(data.volume),
    turnover: toNumber(data.turnover),
    ts: toNumber(data.ts),
  }
}

function mergeKlines(current: KlinePayload[], latest: KlinePayload) {
  const map = new Map<number, KlinePayload>()
  current.forEach((item) => map.set(item.ts, item))
  map.set(latest.ts, latest)
  return Array.from(map.values())
    .sort((left, right) => right.ts - left.ts)
    .slice(0, KLINE_LIMIT)
}

function normalizeKlineList(items: Array<{ open: number; high: number; low: number; close: number; volume: number; turnover: number; ts: number }>) {
  return items
    .map((item) => ({
      interval: DEFAULT_INTERVAL,
      open: item.open,
      high: item.high,
      low: item.low,
      close: item.close,
      volume: item.volume,
      turnover: item.turnover,
      ts: item.ts,
    }))
    .sort((left, right) => right.ts - left.ts)
    .slice(0, KLINE_LIMIT)
}

function asRecord(value: unknown): Record<string, unknown> {
  if (value && typeof value === 'object') return value as Record<string, unknown>
  return {}
}

function toNumber(value: unknown) {
  const next = Number(value)
  return Number.isFinite(next) ? next : 0
}

function getChangeRate(quote?: QuotePayload | null) {
  if (!quote || !quote.open) return 0
  return ((quote.lastPrice - quote.open) / quote.open) * 100
}

function formatNumber(value?: number | null, digits = 2) {
  if (value === null || value === undefined || !Number.isFinite(value)) return '--'
  return new Intl.NumberFormat('zh-CN', {
    minimumFractionDigits: digits,
    maximumFractionDigits: digits,
  }).format(value)
}

function formatPrice(value?: number | null) {
  if (value === null || value === undefined || !Number.isFinite(value)) return '--'
  return formatNumber(value, Math.abs(value) >= 1 ? 4 : 8)
}

function formatCompact(value?: number | null) {
  if (value === null || value === undefined || !Number.isFinite(value)) return '--'
  return new Intl.NumberFormat('zh-CN', {
    notation: 'compact',
    maximumFractionDigits: 2,
  }).format(value)
}

function formatPercent(value: number) {
  return `${value >= 0 ? '+' : ''}${value.toFixed(2)}%`
}
