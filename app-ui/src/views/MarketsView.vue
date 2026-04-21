<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'

import MarketChartView from '@/components/markets/MarketChartView.vue'
import MarketQuotesView from '@/components/markets/MarketQuotesView.vue'
import MarketTopTabs from '@/components/markets/MarketTopTabs.vue'
import type { MarketTopTab, MarketTopTabItem } from '@/components/markets/types'
import { getAccessToken, getTenantId } from '@/api/http'
import { buildItickWsUrl } from '@/api/itick'
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

const store = useItickStore()
const systemStore = useSystemStore()

const DEFAULT_INTERVAL = '1m'
const DEFAULT_K_TYPE = 1
const KLINE_LIMIT = 34
const TICK_LIMIT = 12
const PING_INTERVAL_MS = 20000
const RECONNECT_DELAY_MS = 2500

const topTabs: MarketTopTabItem[] = [
  { key: 'watchlist', label: '自选' },
  { key: 'markets', label: '行情' },
  { key: 'chart', label: '图表' },
]

const activeTopTab = ref<MarketTopTab>('markets')
const selectedCategoryType = ref<number | null>(null)
const selectedProductKey = ref('')
const selectedIntervalName = ref(DEFAULT_INTERVAL)
const depthSnapshot = ref<DepthPayload | null>(null)
const tickSnapshot = ref<TickPayload[]>([])
const klineSnapshot = ref<KlinePayload[]>([])
const quoteMap = ref<Record<string, QuotePayload>>({})
const wsState = ref<ItickWsConnectionState>('closed')
const wsError = ref('')
const loadingBootstrap = ref(false)
const loadingKline = ref(false)
const viewingLatestKlinePage = ref(true)
const wsId = ref('')

let socket: WebSocket | null = null
let pingTimer: number | undefined
let reconnectTimer: number | undefined
let refreshSocketTimer: number | undefined
let isUnmounted = false
let reconnectEnabled = true
let openingProductChart = false

const tenantId = computed(() => getTenantId() ?? Number(import.meta.env.VITE_TENANT_ID || 0))
const isLoggedIn = computed(() => Boolean(getAccessToken()))
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
  () => products.value.find((item) => productKey(item) === selectedProductKey.value) ?? null,
)

const selectedQuote = computed(() => {
  const product = selectedProduct.value
  return product ? quoteMap.value[productKey(product)] ?? null : null
})

const selectedInterval = computed(() => {
  return intervals.value.find((item) => item.name === selectedIntervalName.value) ?? intervals.value[0]
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
      direction: changeRate > 0 ? 'up' as const : changeRate < 0 ? 'down' as const : 'flat' as const,
    }
  }),
)

watch(selectedCategoryType, async (categoryType) => {
  if (!categoryType) return
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

watch(selectedProductKey, async () => {
  depthSnapshot.value = null
  tickSnapshot.value = []
  klineSnapshot.value = []
  if (activeTopTab.value === 'chart' && !openingProductChart) {
    await loadSelectedKlinePage(Date.now(), true)
  }
  queueSocketRefresh()
})

watch(activeTopTab, async (tab) => {
  if (tab === 'chart' && selectedProduct.value) {
    depthSnapshot.value = null
    tickSnapshot.value = []
    await loadSelectedKlinePage(Date.now(), true)
  }

  queueSocketRefresh()
})

watch(intervals, (list) => {
  if (!list.length) return
  const hasSelected = list.some((item) => item.name === selectedIntervalName.value)
  if (!hasSelected) {
    selectedIntervalName.value = list[0].name
  }
}, { immediate: true })

watch(selectedIntervalName, async () => {
  if (activeTopTab.value !== 'chart' || !selectedProduct.value) return

  klineSnapshot.value = []
  await loadSelectedKlinePage(Date.now(), true)
  queueSocketRefresh()
})

onMounted(async () => {
  await initializeMarketPage()
})

onBeforeUnmount(() => {
  isUnmounted = true
  reconnectEnabled = false
  stopReconnectTimer()
  stopRefreshSocketTimer()
  closeSocket()
})

async function initializeMarketPage() {
  loadingBootstrap.value = true
  try {
    const list = await store.listVisibleCategories({
      tenantId: tenantId.value,
      limit: 20,
    })

    if (list.length > 0) {
      selectedCategoryType.value = list[0].categoryType
    }
  } finally {
    loadingBootstrap.value = false
  }
}

async function loadProducts(categoryType: number) {
  loadingBootstrap.value = true
  try {
    await store.listVisibleProducts({
      tenantId: tenantId.value,
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
  } catch (error) {
    console.error('bootstrap kline failed', error)
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

function connectSocket() {
  wsId.value = crypto.randomUUID()
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
  if (activeTopTab.value !== 'chart') return

  const selected = selectedProduct.value
  if (!selected) return

  const depthTopic = productTopic(selected, 'depth')
  const tickTopic = productTopic(selected, 'tick')
  const klineTopic = productTopic(selected, 'kline')
  if (!depthTopic || !tickTopic || !klineTopic) return

  const payload: ItickWsSubscribeMessage = {
    type: 'subscribe',
    topics: [
      depthTopic,
      tickTopic,
      {
        ...klineTopic,
        interval: selectedInterval.value?.name || DEFAULT_INTERVAL,
      },
    ],
  }

  sendJson(payload)
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
          tickSnapshot.value = [normalizeTickPayload(message.payload), ...tickSnapshot.value].slice(0, TICK_LIMIT)
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

function selectCategory(categoryType: number) {
  if (selectedCategoryType.value === categoryType) return
  selectedCategoryType.value = categoryType
}

function selectProduct(product: ItickTenantProduct) {
  selectedProductKey.value = productKey(product)
}

function openProductChart(product: ItickTenantProduct) {
  openingProductChart = true
  selectedProductKey.value = productKey(product)
  activeTopTab.value = 'chart'
  window.setTimeout(() => {
    openingProductChart = false
  })
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
  current.forEach((item) => {
    map.set(item.ts, item)
  })
  map.set(latest.ts, latest)

  return Array.from(map.values())
    .sort((left, right) => right.ts - left.ts)
    .slice(0, KLINE_LIMIT)
}

function normalizeKlineList(items: Array<{
  open: number
  high: number
  low: number
  close: number
  volume: number
  turnover: number
  ts: number
}>) {
  return items
    .map((item) => ({
      interval: selectedInterval.value?.name || DEFAULT_INTERVAL,
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
  if (value && typeof value === 'object') {
    return value as Record<string, unknown>
  }
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

function selectInterval(interval: Interval) {
  if (selectedIntervalName.value === interval.name) return
  selectedIntervalName.value = interval.name
}
</script>

<template>
  <section class="markets-page">
    <div class="market-phone">
      <MarketTopTabs
        :tabs="topTabs"
        :active-tab="activeTopTab"
        @change="activeTopTab = $event"
      />

      <main class="market-content">
        <section v-if="activeTopTab === 'watchlist'" class="watch-panel">
          <div class="watch-panel__icon">★</div>
          <h2>{{ isLoggedIn ? '暂无自选品种' : '登录后查看自选' }}</h2>
          <p>{{ isLoggedIn ? '在图表页点亮星标后，自选会显示在这里。' : '自选行情会跟随账号同步，需要先登录。' }}</p>
        </section>

        <MarketQuotesView
          v-else-if="activeTopTab === 'markets'"
          :categories="categories"
          :selected-category-type="selectedCategoryType"
          :selected-category-name="selectedCategory?.categoryName || ''"
          :selected-category-code="selectedCategoryCode"
          :ws-state="wsState"
          :ws-error="wsError"
          :loading="loadingBootstrap"
          :rows="marketRows"
          :selected-product-key="selectedProductKey"
          @select-category="selectCategory"
          @select-product="openProductChart"
        />

        <MarketChartView
          v-else
          :products="products"
          :rows="marketRows"
          :category-name="selectedCategory?.categoryName || ''"
          :selected-product-key="selectedProductKey"
          :selected-quote="selectedQuote"
          :kline-snapshot="klineSnapshot"
          :depth-snapshot="depthSnapshot"
          :tick-snapshot="tickSnapshot"
          :loading-kline="loadingKline"
          :intervals="intervals"
          :selected-interval-name="selectedIntervalName"
          @select-product="selectedProductKey = $event"
          @select-interval="selectInterval"
          @load-previous-page="loadPreviousKlinePage"
        />
      </main>
    </div>
  </section>
</template>

<style scoped>
.markets-page {
  display: flex;
  justify-content: center;
  min-height: calc(100dvh - 120px);
  padding: 12px 0 72px;
  background: #080910;
}

.market-phone {
  width: min(100%, 640px);
  height: calc(100dvh - 132px);
  min-height: 560px;
  overflow-x: hidden;
  overflow-y: auto;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 8px;
  background: #0b0c15;
  color: #f6f7fb;
  box-shadow: 0 18px 60px rgba(0, 0, 0, 0.3);
  scrollbar-width: none;
}

.market-phone::-webkit-scrollbar {
  display: none;
}

.market-content {
  min-height: 560px;
}

.watch-panel {
  display: grid;
  place-items: center;
  align-content: center;
  gap: 10px;
  min-height: 320px;
  padding: 32px 18px;
  color: #8f929d;
  text-align: center;
}

.watch-panel__icon {
  display: grid;
  place-items: center;
  width: 56px;
  height: 56px;
  border: 1px solid #2a2d38;
  border-radius: 999px;
  color: #fff;
  font-size: 24px;
}

.watch-panel h2 {
  margin: 0;
  color: #fff;
  font-size: 18px;
  font-weight: 500;
}

.watch-panel p {
  max-width: 300px;
  margin: 0;
  font-size: 14px;
  line-height: 1.65;
}

@media (max-width: 768px) {
  .markets-page {
    padding: 0 0 72px;
  }

  .market-phone {
    width: 100%;
    height: calc(100dvh - 72px);
    min-height: 0;
    border-right: 0;
    border-left: 0;
    border-radius: 0;
  }
}
</style>
