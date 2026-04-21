<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'

import { getTenantId } from '@/api/http'
import { buildItickWsUrl } from '@/api/itick'
import { useItickStore } from '@/stores/itick'
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

const DEFAULT_INTERVAL = '1m'
const KLINE_LIMIT = 18
const TICK_LIMIT = 12
const PING_INTERVAL_MS = 20000
const RECONNECT_DELAY_MS = 2500

const selectedCategoryType = ref<number | null>(null)
const selectedProductKey = ref('')
const depthSnapshot = ref<DepthPayload | null>(null)
const tickSnapshot = ref<TickPayload[]>([])
const klineSnapshot = ref<KlinePayload[]>([])
const quoteMap = ref<Record<string, QuotePayload>>({})
const wsState = ref<ItickWsConnectionState>('closed')
const wsError = ref('')
const loadingBootstrap = ref(false)

let socket: WebSocket | null = null
let pingTimer: number | undefined
let reconnectTimer: number | undefined
let refreshSocketTimer: number | undefined
let isUnmounted = false
let reconnectEnabled = true

const tenantId = computed(() => getTenantId() ?? Number(import.meta.env.VITE_TENANT_ID || 0))
const categories = computed(() => store.categories)
const products = computed(() => store.products)

const selectedCategory = computed(
  () => categories.value.find((item) => item.categoryType === selectedCategoryType.value) ?? null,
)

const selectedProduct = computed(
  () => products.value.find((item) => productKey(item) === selectedProductKey.value) ?? null,
)

const selectedCategoryCode = computed(() => {
  return selectedCategory.value?.categoryCode || ''
})

const selectedQuote = computed(() => {
  const product = selectedProduct.value
  return product ? quoteMap.value[productKey(product)] ?? null : null
})

const marketRows = computed(() =>
  products.value.map((product) => {
    const quote = quoteMap.value[productKey(product)] ?? null
    const changeRate = quote && quote.open > 0 ? ((quote.lastPrice - quote.open) / quote.open) * 100 : 0

    return {
      key: productKey(product),
      product,
      quote,
      changeRate,
      direction: changeRate > 0 ? 'up' : changeRate < 0 ? 'down' : 'flat',
    }
  }),
)

const stats = computed(() => {
  const quote = selectedQuote.value
  if (!quote) {
    return [
      { label: '24H高', value: '--' },
      { label: '24H低', value: '--' },
      { label: '24H量', value: '--' },
      { label: '成交额', value: '--' },
    ]
  }

  return [
    { label: '24H高', value: formatNumber(quote.high) },
    { label: '24H低', value: formatNumber(quote.low) },
    { label: '24H量', value: formatCompact(quote.volume) },
    { label: '成交额', value: formatCompact(quote.turnover) },
  ]
})

const selectedPriceChange = computed(() => {
  const quote = selectedQuote.value
  if (!quote || !quote.open) {
    return 0
  }
  return ((quote.lastPrice - quote.open) / quote.open) * 100
})

const selectedTrendClass = computed(() => {
  if (selectedPriceChange.value > 0) return 'up'
  if (selectedPriceChange.value < 0) return 'down'
  return 'flat'
})

watch(selectedCategoryType, async (categoryType) => {
  if (!categoryType) {
    return
  }

  await loadProducts(categoryType)
}, { immediate: false })

watch(products, async (list) => {
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
    selectedProductKey.value = ''
  }

  queueSocketRefresh()
}, { deep: true })

watch(selectedProductKey, async () => {
  depthSnapshot.value = null
  tickSnapshot.value = []
  klineSnapshot.value = []

  await bootstrapSelectedProduct()
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

async function bootstrapSelectedProduct() {
  const product = selectedProduct.value
  if (!product) {
    return
  }

  try {
    const categoryCode = product.categoryCode || selectedCategoryCode.value
    if (!categoryCode) {
      return
    }

    const klines = await store.getKline({
      categoryCode,
      market: product.market,
      symbol: product.symbol,
      kType: 1,
      limit: KLINE_LIMIT,
    })

    klineSnapshot.value = klines
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
  } catch (error) {
    console.error('bootstrap kline failed', error)
  }
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

function stopRefreshSocketTimer() {
  if (refreshSocketTimer !== undefined) {
    window.clearTimeout(refreshSocketTimer)
    refreshSocketTimer = undefined
  }
}

function connectSocket() {
  console.log('=================================================. 9999')
  const url = buildItickWsUrl()

  stopReconnectTimer()
  reconnectEnabled = false
  closeSocket()

  wsState.value = 'connecting'
  wsError.value = ''

  console.log('itick ws url =>', url)
  socket = new WebSocket(url)

  socket.addEventListener('open', () => {
    console.log('=================================================. 666666')
    reconnectEnabled = true
    wsState.value = 'open'
    sendQuoteSubscription()
    sendSelectedDetailSubscription()
    startPingLoop()
  })

  socket.addEventListener('message', (event) => {
    console.log('=================================================. 5555555')
    handleSocketMessage(event.data)
  })

  socket.addEventListener('error', () => {
    console.log('=================================================. 88888')
    wsError.value = '实时连接异常，正在尝试恢复。'
  })

  socket.addEventListener('close', () => {
    console.log('=================================================. 777777')
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

function stopReconnectTimer() {
  if (reconnectTimer !== undefined) {
    window.clearTimeout(reconnectTimer)
    reconnectTimer = undefined
  }
}

function startPingLoop() {
  console.log('=================================================. 22222')
  stopPingLoop()
  pingTimer = window.setInterval(() => {
    console.log('=================================================.  4444')
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

function sendQuoteSubscription() {
  const topics = products.value
    .map((product) => productTopic(product, 'quote'))
    .filter((topic): topic is ItickWsTopicConfig => Boolean(topic))

  if (!topics.length) {
    return
  }

  const payload: ItickWsSubscribeMessage = {
    type: 'subscribe',
    topics: dedupeTopics(topics),
  }

  sendJson(payload)
}

function sendSelectedDetailSubscription() {
  const selected = selectedProduct.value
  if (!selected) {
    return
  }

  const depthTopic = productTopic(selected, 'depth')
  const tickTopic = productTopic(selected, 'tick')
  const klineTopic = productTopic(selected, 'kline')
  if (!depthTopic || !tickTopic || !klineTopic) {
    return
  }

  const payload: ItickWsSubscribeMessage = {
    type: 'subscribe',
    topics: [
      {
        ...depthTopic,
      },
      {
        ...tickTopic,
      },
      {
        ...klineTopic,
        interval: DEFAULT_INTERVAL,
      },
    ],
  }

  sendJson(payload)
}

function sendJson(payload: ItickWsSubscribeMessage | { type: 'ping'; clientTs: number }) {
  if (!socket || socket.readyState !== WebSocket.OPEN) {
    return
  }

  socket.send(JSON.stringify(payload))
}

function handleSocketMessage(raw: string) {
  try {
    const message = JSON.parse(raw) as ItickWsPongMessage | ItickWsServerMessage<unknown>

    if ('type' in message && message.type === 'pong') {
      return
    }

    if (!('topic' in message)) {
      return
    }

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
        if (targetKey === currentKey) {
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
  if (selectedCategoryType.value === categoryType) {
    return
  }
  selectedCategoryType.value = categoryType
}

function selectProduct(product: ItickTenantProduct) {
  selectedProductKey.value = productKey(product)
}

function productKey(product: Pick<ItickTenantProduct, 'market' | 'symbol'>) {
  return productKeyByFields(product.market, product.symbol)
}

function productKeyByFields(market: string, symbol: string) {
  return `${String(market || '').toUpperCase()}::${String(symbol || '').toUpperCase()}`
}

function productTopic(product: ItickTenantProduct, topic: ItickWsTopicConfig['topic']) {
  const categoryCode = product.categoryCode || selectedCategoryCode.value
  if (!categoryCode || !product.market || !product.symbol) {
    return null
  }

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
    if (seen.has(key)) {
      return false
    }
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
    asks: normalizeDepthLevels(data.asks ?? data.Asks),
    bids: normalizeDepthLevels(data.bids ?? data.Bids),
  }
}

function normalizeDepthLevels(source: unknown): DepthLevel[] {
  if (!Array.isArray(source)) {
    return []
  }

  return source
    .map((item) => {
      const level = asRecord(item)
      return {
        price: toNumber(level.p ?? level.price),
        volume: toNumber(level.v ?? level.volume),
        position: toNumber(level.po ?? level.position),
        originVolume: toNumber(level.o ?? level.originVolume),
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

function formatNumber(value?: number | null, digits = 2) {
  if (value === null || value === undefined || !Number.isFinite(value)) {
    return '--'
  }

  return new Intl.NumberFormat('zh-CN', {
    minimumFractionDigits: digits,
    maximumFractionDigits: digits,
  }).format(value)
}

function formatCompact(value?: number | null) {
  if (value === null || value === undefined || !Number.isFinite(value)) {
    return '--'
  }

  return new Intl.NumberFormat('zh-CN', {
    notation: 'compact',
    maximumFractionDigits: 2,
  }).format(value)
}

function formatPercent(value: number) {
  return `${value >= 0 ? '+' : ''}${value.toFixed(2)}%`
}

function formatTime(ts: number) {
  if (!ts) {
    return '--'
  }

  return new Intl.DateTimeFormat('zh-CN', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    month: '2-digit',
    day: '2-digit',
  }).format(ts)
}
</script>

<template>
  <section class="markets-page">
    <section class="markets-hero">
      <div>
        <p class="markets-hero__eyebrow">实时市场</p>
        <h1>分类筛选、产品切换与 WebSocket 行情订阅一次到位</h1>
        <p class="markets-hero__text">
          先加载可见分类，再按分类拉取产品列表。选中产品后自动订阅报价、盘口、成交与 1 分钟 K 线。
        </p>
      </div>

      <div class="markets-hero__status">
        <span class="markets-hero__dot" :class="`markets-hero__dot--${wsState}`" />
        <strong>{{ wsState === 'open' ? '实时已连接' : wsState === 'connecting' ? '连接中' : '已断开' }}</strong>
        <span>{{ selectedCategory?.categoryName || '未选择分类' }}</span>
        <small>{{ wsError || `categoryCode: ${selectedCategoryCode || '--'}` }}</small>
      </div>
    </section>

    <section class="markets-layout">
      <aside class="markets-sidebar panel-card">
        <div class="markets-sidebar__header">
          <h2>分类</h2>
          <span>{{ categories.length }} 个可见分组</span>
        </div>

        <button
          v-for="category in categories"
          :key="category.id"
          type="button"
          class="markets-category"
          :class="{ 'markets-category--active': category.categoryType === selectedCategoryType }"
          @click="selectCategory(category.categoryType)"
        >
          <strong>{{ category.categoryName }}</strong>
          <span>分类类型 #{{ category.categoryType }}</span>
        </button>
      </aside>

      <section class="markets-main">
        <article class="panel-card markets-board">
          <div class="markets-board__header">
            <div>
              <h2>{{ selectedProduct?.displayName || selectedProduct?.symbol || '选择产品查看详情' }}</h2>
              <p>
                {{ selectedProduct?.market || '--' }} · {{ selectedProduct?.symbol || '--' }}
                <span v-if="selectedProduct?.quoteCoin">/ {{ selectedProduct.quoteCoin }}</span>
              </p>
            </div>

            <div class="markets-board__price" :class="selectedTrendClass">
              <strong>{{ selectedQuote ? formatNumber(selectedQuote.lastPrice, 4) : '--' }}</strong>
              <span>{{ formatPercent(selectedPriceChange) }}</span>
            </div>
          </div>

          <div class="markets-stats">
            <article v-for="item in stats" :key="item.label" class="markets-stats__item">
              <span>{{ item.label }}</span>
              <strong>{{ item.value }}</strong>
            </article>
          </div>

          <p v-if="!selectedProduct" class="markets-board__hint">
            当前未选中具体品种，页面会先为当前分类下全部产品订阅实时报价。
          </p>
        </article>

        <article class="panel-card markets-products">
          <div class="markets-products__header">
            <h2>产品列表</h2>
            <span>{{ marketRows.length }} 个品种</span>
          </div>

          <div v-if="loadingBootstrap" class="markets-empty">正在加载市场数据...</div>

          <div v-else-if="!marketRows.length" class="markets-empty">当前分类暂无可见产品。</div>

          <button
            v-for="row in marketRows"
            v-else
            :key="row.key"
            type="button"
            class="markets-product-row"
            :class="{
              'markets-product-row--active': row.key === selectedProductKey,
              'markets-product-row--up': row.direction === 'up',
              'markets-product-row--down': row.direction === 'down',
            }"
            @click="selectProduct(row.product)"
          >
            <div>
              <strong>{{ row.product.displayName || row.product.symbol }}</strong>
              <span>{{ row.product.market }} · {{ row.product.symbol }}</span>
            </div>

            <div class="markets-product-row__quote">
              <strong>{{ row.quote ? formatNumber(row.quote.lastPrice, 4) : '--' }}</strong>
              <span>{{ row.quote ? formatPercent(row.changeRate) : '等待推送' }}</span>
            </div>
          </button>
        </article>

        <section class="markets-detail-grid">
          <article class="panel-card markets-depth">
            <div class="markets-section__header">
              <h3>盘口深度</h3>
              <span>{{ selectedProduct?.symbol || '--' }}</span>
            </div>

            <div class="markets-depth__columns">
              <div>
                <h4>卖盘</h4>
                <div v-if="!depthSnapshot?.asks.length" class="markets-mini-empty">等待深度推送</div>
                <div v-for="(item, index) in depthSnapshot?.asks.slice(0, 6)" :key="`ask-${index}`" class="markets-depth__row markets-depth__row--ask">
                  <span>{{ formatNumber(item.price, 4) }}</span>
                  <strong>{{ formatCompact(item.volume) }}</strong>
                </div>
              </div>

              <div>
                <h4>买盘</h4>
                <div v-if="!depthSnapshot?.bids.length" class="markets-mini-empty">等待深度推送</div>
                <div v-for="(item, index) in depthSnapshot?.bids.slice(0, 6)" :key="`bid-${index}`" class="markets-depth__row markets-depth__row--bid">
                  <span>{{ formatNumber(item.price, 4) }}</span>
                  <strong>{{ formatCompact(item.volume) }}</strong>
                </div>
              </div>
            </div>
          </article>

          <article class="panel-card markets-ticks">
            <div class="markets-section__header">
              <h3>实时成交</h3>
              <span>{{ tickSnapshot.length }} 条</span>
            </div>

            <div v-if="!tickSnapshot.length" class="markets-mini-empty">等待成交推送</div>

            <div v-for="(item, index) in tickSnapshot" v-else :key="`${item.ts}-${index}`" class="markets-ticks__row">
              <div>
                <strong>{{ formatNumber(item.lastPrice, 4) }}</strong>
                <span>{{ formatTime(item.ts) }}</span>
              </div>
              <strong>{{ formatCompact(item.volume) }}</strong>
            </div>
          </article>

          <article class="panel-card markets-klines">
            <div class="markets-section__header">
              <h3>1 分钟 K 线</h3>
              <span>{{ klineSnapshot.length }} 根</span>
            </div>

            <div v-if="!klineSnapshot.length" class="markets-mini-empty">等待 K 线推送</div>

            <div v-for="item in klineSnapshot" v-else :key="item.ts" class="markets-kline-row">
              <div>
                <strong>{{ formatTime(item.ts) }}</strong>
                <span>{{ item.interval }}</span>
              </div>
              <div>
                <span>O {{ formatNumber(item.open, 4) }}</span>
                <span>H {{ formatNumber(item.high, 4) }}</span>
                <span>L {{ formatNumber(item.low, 4) }}</span>
                <strong>C {{ formatNumber(item.close, 4) }}</strong>
              </div>
            </div>
          </article>
        </section>
      </section>
    </section>
  </section>
</template>

<style scoped>
.markets-page {
  display: flex;
  flex-direction: column;
  gap: 24px;
  padding-bottom: 40px;
}

.markets-hero,
.panel-card {
  position: relative;
  border: 1px solid rgba(150, 161, 207, 0.18);
  border-radius: 28px;
  background:
    linear-gradient(140deg, rgba(51, 63, 55, 0.24) 0%, rgba(16, 18, 31, 0.92) 45%, rgba(14, 16, 24, 0.98) 100%),
    linear-gradient(180deg, rgba(255, 255, 255, 0.03), rgba(255, 255, 255, 0));
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.06);
}

.markets-hero {
  display: flex;
  justify-content: space-between;
  gap: 24px;
  padding: 30px 32px;
}

.markets-hero__eyebrow {
  margin: 0 0 10px;
  color: #08c200;
  font-size: 13px;
  letter-spacing: 0.24em;
  text-transform: uppercase;
}

.markets-hero h1 {
  margin: 0;
  max-width: 760px;
  color: #f7f9ff;
  font-size: clamp(32px, 4vw, 48px);
  line-height: 1.08;
}

.markets-hero__text {
  margin: 18px 0 0;
  max-width: 720px;
  color: rgba(228, 233, 255, 0.68);
  font-size: 16px;
  line-height: 1.7;
}

.markets-hero__status {
  display: grid;
  gap: 8px;
  min-width: 260px;
  align-self: flex-start;
  padding: 18px 20px;
  border-radius: 22px;
  background: rgba(255, 255, 255, 0.03);
  color: rgba(228, 233, 255, 0.72);
}

.markets-hero__status strong {
  color: #f7f9ff;
  font-size: 18px;
}

.markets-hero__status small {
  color: rgba(228, 233, 255, 0.52);
}

.markets-hero__dot {
  width: 12px;
  height: 12px;
  border-radius: 999px;
}

.markets-hero__dot--connecting {
  background: #f2c94c;
}

.markets-hero__dot--open {
  background: #08c200;
  box-shadow: 0 0 18px rgba(8, 194, 0, 0.45);
}

.markets-hero__dot--closed {
  background: #eb5757;
}

.markets-layout {
  display: grid;
  grid-template-columns: 300px minmax(0, 1fr);
  gap: 24px;
}

.panel-card {
  padding: 22px;
}

.markets-sidebar {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.markets-sidebar__header,
.markets-products__header,
.markets-board__header,
.markets-section__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.markets-sidebar__header h2,
.markets-products__header h2,
.markets-board__header h2,
.markets-section__header h3 {
  margin: 0;
  color: #f7f9ff;
}

.markets-sidebar__header span,
.markets-products__header span,
.markets-section__header span,
.markets-board__header p {
  color: rgba(228, 233, 255, 0.58);
}

.markets-board__header p {
  margin: 8px 0 0;
}

.markets-category,
.markets-product-row {
  width: 100%;
  border: 1px solid rgba(139, 150, 194, 0.12);
  border-radius: 20px;
  background: rgba(255, 255, 255, 0.02);
  color: inherit;
  cursor: pointer;
  transition:
    transform 0.2s ease,
    border-color 0.2s ease,
    background 0.2s ease;
}

.markets-category {
  display: grid;
  gap: 6px;
  padding: 16px 18px;
  text-align: left;
}

.markets-category strong,
.markets-product-row strong,
.markets-board__price strong,
.markets-stats__item strong,
.markets-depth__row strong,
.markets-ticks__row strong,
.markets-kline-row strong {
  color: #f7f9ff;
}

.markets-category span,
.markets-product-row span,
.markets-stats__item span,
.markets-depth__row span,
.markets-ticks__row span,
.markets-kline-row span,
.markets-empty,
.markets-mini-empty {
  color: rgba(228, 233, 255, 0.6);
}

.markets-category:hover,
.markets-product-row:hover {
  transform: translateY(-1px);
  border-color: rgba(8, 194, 0, 0.35);
}

.markets-category--active,
.markets-product-row--active {
  border-color: rgba(8, 194, 0, 0.48);
  background: linear-gradient(135deg, rgba(8, 194, 0, 0.16), rgba(255, 255, 255, 0.04));
}

.markets-main {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.markets-board__price {
  display: grid;
  gap: 6px;
  justify-items: end;
}

.markets-board__price strong {
  font-size: clamp(28px, 4vw, 42px);
}

.markets-board__price.up span,
.markets-product-row--up .markets-product-row__quote span,
.markets-depth__row--bid span {
  color: #08c200;
}

.markets-board__price.down span,
.markets-product-row--down .markets-product-row__quote span,
.markets-depth__row--ask span {
  color: #ff6b6b;
}

.markets-board__price.flat span {
  color: rgba(228, 233, 255, 0.72);
}

.markets-stats {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
  margin-top: 22px;
}

.markets-stats__item {
  display: grid;
  gap: 8px;
  padding: 18px;
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.03);
}

.markets-products {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.markets-product-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 16px 18px;
  text-align: left;
}

.markets-product-row__quote {
  display: grid;
  gap: 6px;
  justify-items: end;
}

.markets-empty,
.markets-mini-empty {
  display: grid;
  place-items: center;
  min-height: 120px;
  border-radius: 20px;
  border: 1px dashed rgba(139, 150, 194, 0.18);
}

.markets-detail-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 24px;
}

.markets-depth__columns {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 18px;
}

.markets-depth__columns h4 {
  margin: 0 0 14px;
  color: rgba(228, 233, 255, 0.8);
}

.markets-depth__row,
.markets-ticks__row,
.markets-kline-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 10px 0;
  border-bottom: 1px solid rgba(139, 150, 194, 0.08);
}

.markets-ticks,
.markets-klines {
  max-height: 520px;
  overflow: auto;
}

.markets-ticks__row div,
.markets-kline-row div {
  display: grid;
  gap: 4px;
}

.markets-kline-row > div:last-child {
  justify-items: end;
}

@media (max-width: 1200px) {
  .markets-layout {
    grid-template-columns: 1fr;
  }

  .markets-detail-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 768px) {
  .markets-page {
    gap: 18px;
    padding-bottom: 120px;
  }

  .markets-hero,
  .panel-card {
    border-radius: 24px;
  }

  .markets-hero {
    flex-direction: column;
    padding: 24px 20px;
  }

  .markets-hero__status {
    width: 100%;
    min-width: 0;
  }

  .panel-card {
    padding: 18px;
  }

  .markets-board__header,
  .markets-sidebar__header,
  .markets-products__header,
  .markets-section__header {
    align-items: flex-start;
    flex-direction: column;
  }

  .markets-board__price,
  .markets-product-row__quote,
  .markets-kline-row > div:last-child {
    justify-items: start;
  }

  .markets-stats {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .markets-product-row,
  .markets-kline-row,
  .markets-ticks__row {
    align-items: flex-start;
    flex-direction: column;
  }

  .markets-depth__columns {
    grid-template-columns: 1fr;
  }
}
</style>
