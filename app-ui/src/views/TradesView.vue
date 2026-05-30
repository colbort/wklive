<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'

import { getAccessToken } from '@/api/http'
import {
  apiTradeCancelOrder,
  apiTradeGetLeverageConfig,
  apiTradeGetMarginAccountList,
  apiTradeGetOrderList,
  apiTradeGetPositionList,
  apiTradeGetSymbolDetail,
  apiTradeGetSymbolList,
  apiTradePlaceOrder,
  apiTradeSetLeverage,
} from '@/api/trade'
import DesktopMarketTradeView from '@/components/markets/DesktopMarketTradeView.vue'
import MobileTradeView from '@/components/trades/MobileTradeView.vue'
import { useDevice } from '@/composables/useDevice'
import { useTradingDesk } from '@/composables/useTradingDesk'
import type {
  ContractLeverageConfig,
  ContractMarginAccount,
  ContractPosition,
  PlaceOrderReq,
  TradeOrder,
  TradeSymbol,
  TradeSymbolContract,
  TradeSymbolLeverageConfig,
  TradeSymbolSpot,
} from '@/types/trade'

type SubmitSide = 'buy' | 'sell'
type TradeSymbolDetail = {
  symbol: TradeSymbol | null
  spot: TradeSymbolSpot | null
  contract: TradeSymbolContract | null
  leverageConfigs: TradeSymbolLeverageConfig[]
}

const MARKET_TYPE_SPOT = 1
const MARKET_TYPE_SECONDS_CONTRACT = 2
const MARKET_TYPE_USDT_CONTRACT = 3
const MARKET_TYPE_COIN_CONTRACT = 4
const TRADE_SIDE_BUY = 1
const TRADE_SIDE_SELL = 2
const POSITION_SIDE_NET = 1
const POSITION_SIDE_LONG = 2
const POSITION_SIDE_SHORT = 3
const POSITION_MODE_ONE_WAY = 1
const ORDER_TYPE_LIMIT = 1
const ORDER_TYPE_MARKET = 2
const TIME_IN_FORCE_GTC = 1
const TIME_IN_FORCE_IOC = 2
const ORDER_SOURCE_WEB = 2
const SYMBOL_STATUS_ENABLED = 1
const FALLBACK_LEVERAGE_VALUES = [1, 2, 5, 10, 20, 50, 75, 100, 125]

// 交易页：手机端展示交易表单与订单列表，桌面端复用市场交易组合台。
const route = useRoute()
const { isDesktop } = useDevice()
const detailVisible = computed(() => true)
const orderMode = ref<'market' | 'limit'>('market')
const productMenuOpen = ref(false)
const desktopProductsExpanded = ref(false)
const desktopOrderbookExpanded = ref(true)
const authToken = ref(getAccessToken())
const tradeSymbols = ref<TradeSymbol[]>([])
const tradeSymbolDetail = ref<TradeSymbolDetail | null>(null)
const userLeverageConfig = ref<ContractLeverageConfig | null>(null)
const tradeOrders = ref<TradeOrder[]>([])
const tradePositions = ref<ContractPosition[]>([])
const marginAccounts = ref<ContractMarginAccount[]>([])
const loadingTradeSymbols = ref(false)
const loadingTradeDetail = ref(false)
const loadingTradeAccount = ref(false)
const loadingTradeOrders = ref(false)
const tradePrice = ref('')
const tradeQty = ref('')
const tradePercent = ref(0)
const takeProfitPrice = ref('')
const stopLossPrice = ref('')
const marginMode = ref(1)
const leverage = ref(1)
const submittingSide = ref<SubmitSide | null>(null)
const cancelingOrderId = ref<number | null>(null)
const tradeMessage = ref('')
const tradeError = ref('')
const ordersError = ref('')
let tradeSymbolsRequestId = 0
let tradeDetailRequestId = 0
let tradeAccountRequestId = 0
let tradeOrdersRequestId = 0
let leverageConfigRequestId = 0
let saveLeverageRequestId = 0
const {
  selectedCategoryType,
  selectedProductKey,
  selectedIntervalName,
  loadingBootstrap,
  loadingKline,
  depthSnapshot,
  tickSnapshot,
  klineSnapshot,
  categories,
  intervals,
  selectedCategory,
  selectedProduct,
  selectedQuote,
  placeholderPrice,
  placeholderChange,
  priceTrend,
  desktopProductRows,
  productSheetRows,
  desktopStats,
  selectProduct,
  selectInterval,
  coinGlyph,
  productKey,
} = useTradingDesk({
  detailVisible,
  tickLimit: 24,
})
const tradeKind = computed(() => {
  const code =
    `${selectedCategory.value?.categoryCode || ''} ${selectedCategory.value?.categoryName || ''}`.toLowerCase()
  if (code.includes('stock') || code.includes('股票')) return 'stock'
  if (code.includes('option') || code.includes('期权')) return 'option'
  if (code.includes('forex') || code.includes('外汇')) return 'forex'
  if (code.includes('commodity') || code.includes('大宗')) return 'commodity'
  return 'crypto'
})
const isLoggedIn = computed(() => Boolean(authToken.value))
const selectedTradeSymbol = computed(() => matchTradeSymbol())
const selectedTradeSettleAsset = computed(() => {
  return (
    selectedTradeSymbol.value?.settleAsset ||
    selectedTradeSymbol.value?.quoteAsset ||
    selectedProduct.value?.quoteCoin ||
    'USDT'
  )
})
const selectedMarginAccount = computed(() => {
  const settleAsset = selectedTradeSettleAsset.value.toUpperCase()
  return (
    marginAccounts.value.find((account) => account.marginAsset.toUpperCase() === settleAsset) ||
    marginAccounts.value[0] ||
    null
  )
})
const availableBalance = computed(() => selectedMarginAccount.value?.availableBalance || '0')
const longPositionQty = computed(() => {
  const position = tradePositions.value.find((item) => item.positionSide === POSITION_SIDE_LONG)
  return position?.availQty || position?.qty || '0'
})
const shortPositionQty = computed(() => {
  const position = tradePositions.value.find((item) => item.positionSide === POSITION_SIDE_SHORT)
  return position?.availQty || position?.qty || '0'
})
const activeLeverageConfig = computed(() => {
  const symbol = selectedTradeSymbol.value
  if (!symbol) return null
  return (
    tradeSymbolDetail.value?.leverageConfigs.find((config) => {
      return (
        config.status === 1 &&
        config.marketType === symbol.marketType &&
        config.marginMode === marginMode.value
      )
    }) || null
  )
})
const configuredLeverageValues = computed(() => {
  const values = activeLeverageConfig.value?.leverageValues || []
  return Array.from(
    new Set(values.map(Number).filter((value) => Number.isFinite(value) && value > 0)),
  ).sort((left, right) => left - right)
})
const maxTradeLeverage = computed(() => {
  return Math.max(
    1,
    configuredLeverageValues.value[configuredLeverageValues.value.length - 1] || 0,
    Number(activeLeverageConfig.value?.maxLeverage || 0),
    Number(selectedTradeSymbol.value?.maxLeverage || 0),
  )
})
const tradeLeverageValues = computed(() => {
  if (configuredLeverageValues.value.length) return configuredLeverageValues.value

  const maxLeverage = maxTradeLeverage.value
  const values = FALLBACK_LEVERAGE_VALUES.filter((value) => value <= maxLeverage)
  if (!values.includes(maxLeverage)) {
    values.push(maxLeverage)
  }
  return Array.from(new Set(values)).sort((left, right) => left - right)
})
const tradeAvailable = computed(() =>
  Boolean(selectedTradeSymbol.value && !loadingTradeSymbols.value),
)

watch(
  () => route.query,
  (query) => {
    const categoryType = Number(query.categoryType)
    if (Number.isFinite(categoryType) && categoryType > 0) {
      selectedCategoryType.value = categoryType
    }

    const market = String(query.market || '')
    const symbol = String(query.symbol || '')
    if (market && symbol) {
      selectedProductKey.value = productKey({ market, symbol })
    }
  },
  { immediate: true },
)
watch(selectedProductKey, () => {
  productMenuOpen.value = false
})
watch(
  () => selectedTradeSymbol.value?.id || 0,
  (symbolId) => {
    tradeDetailRequestId += 1
    tradeAccountRequestId += 1
    tradeOrdersRequestId += 1
    tradeMessage.value = ''
    tradeError.value = ''
    ordersError.value = ''
    tradePrice.value = ''
    tradeQty.value = ''
    tradePercent.value = 0
    takeProfitPrice.value = ''
    stopLossPrice.value = ''
    tradeSymbolDetail.value = null
    userLeverageConfig.value = null
    tradeOrders.value = []
    tradePositions.value = []
    marginAccounts.value = []

    if (!symbolId) return
    void loadTradeSymbolDetail(symbolId)
    void refreshTradeAccount()
  },
  { immediate: true },
)
watch(isLoggedIn, () => {
  void refreshTradeAccount()
  void loadUserLeverageConfig()
})
watch([maxTradeLeverage, configuredLeverageValues], () => {
  const nextLeverage = clampLeverage(leverage.value || defaultTradeLeverage())
  if (leverage.value !== nextLeverage) {
    leverage.value = nextLeverage
  }
})
watch(
  [() => selectedTradeSymbol.value?.id || 0, marginMode],
  () => {
    void loadUserLeverageConfig()
  },
  { immediate: true },
)

onMounted(() => {
  syncAuthState()
  void loadTradeSymbols()
  window.addEventListener('focus', syncAuthState)
  document.addEventListener('visibilitychange', syncAuthState)
})

onBeforeUnmount(() => {
  window.removeEventListener('focus', syncAuthState)
  document.removeEventListener('visibilitychange', syncAuthState)
})

function closeProductSheet() {
  productMenuOpen.value = false
}

function toggleDesktopProducts() {
  desktopProductsExpanded.value = !desktopProductsExpanded.value
}

function toggleDesktopOrderbook() {
  desktopOrderbookExpanded.value = !desktopOrderbookExpanded.value
}

function syncAuthState() {
  authToken.value = getAccessToken()
}

function isSuccessCode(code: number) {
  return code === 0 || code === 200
}

function isContractMarket(marketType: number) {
  return [
    MARKET_TYPE_SECONDS_CONTRACT,
    MARKET_TYPE_USDT_CONTRACT,
    MARKET_TYPE_COIN_CONTRACT,
  ].includes(marketType)
}

function normalizeSymbolText(value: string) {
  return String(value || '')
    .toUpperCase()
    .replace(/[^A-Z0-9]/g, '')
}

function productCandidates() {
  const product = selectedProduct.value
  if (!product) return new Set<string>()

  return new Set(
    [
      product.symbol,
      product.code,
      product.name,
      product.displayName,
      `${product.baseCoin}${product.quoteCoin}`,
      `${product.baseCoin}/${product.quoteCoin}`,
    ]
      .map(normalizeSymbolText)
      .filter(Boolean),
  )
}

function symbolCandidates(symbol: TradeSymbol) {
  return [
    symbol.symbol,
    symbol.displaySymbol,
    `${symbol.baseAsset}${symbol.quoteAsset}`,
    `${symbol.baseAsset}/${symbol.quoteAsset}`,
    `${symbol.baseAsset}${symbol.settleAsset}`,
  ]
    .map(normalizeSymbolText)
    .filter(Boolean)
}

function marketPreference(symbol: TradeSymbol) {
  if (tradeKind.value === 'crypto') {
    if (symbol.marketType === MARKET_TYPE_USDT_CONTRACT) return 0
    if (symbol.marketType === MARKET_TYPE_COIN_CONTRACT) return 1
    if (symbol.marketType === MARKET_TYPE_SPOT) return 2
  }

  if (symbol.marketType === MARKET_TYPE_SPOT) return 0
  return 1
}

function matchTradeSymbol() {
  const candidates = productCandidates()
  if (!candidates.size) return null

  return (
    tradeSymbols.value
      .filter((symbol) => symbolCandidates(symbol).some((candidate) => candidates.has(candidate)))
      .sort((left, right) => marketPreference(left) - marketPreference(right))[0] || null
  )
}

function isPositiveDecimal(value: string) {
  const text = value.trim()
  if (!/^\d+(\.\d+)?$/.test(text)) return false
  return Number(text) > 0
}

function divideAmountTextBy100(value: string) {
  const text = String(value ?? '').trim()
  if (!/^-?\d+(\.\d+)?$/.test(text)) return value

  const negative = text.startsWith('-')
  const unsignedText = negative ? text.slice(1) : text
  const [integerPart, fractionPart = ''] = unsignedText.split('.')
  const digits = `${integerPart}${fractionPart}`.replace(/^0+(?=\d)/, '') || '0'
  const decimalPlaces = fractionPart.length + 2
  const paddedDigits = digits.padStart(decimalPlaces + 1, '0')
  const integerEndIndex = paddedDigits.length - decimalPlaces
  const nextInteger = paddedDigits.slice(0, integerEndIndex).replace(/^0+(?=\d)/, '') || '0'
  const nextFraction = paddedDigits.slice(integerEndIndex).replace(/0+$/, '')
  const nextValue = nextFraction ? `${nextInteger}.${nextFraction}` : nextInteger

  if (nextValue === '0') return '0'
  return negative ? `-${nextValue}` : nextValue
}

function normalizeTradeOrderAmounts(order: TradeOrder): TradeOrder {
  return {
    ...order,
    amount: divideAmountTextBy100(order.amount),
    filledAmount: divideAmountTextBy100(order.filledAmount),
    fee: divideAmountTextBy100(order.fee),
  }
}

function normalizeContractPositionAmounts(position: ContractPosition): ContractPosition {
  return {
    ...position,
    positionMargin: divideAmountTextBy100(position.positionMargin),
    isolatedMargin: divideAmountTextBy100(position.isolatedMargin),
    unrealizedPnl: divideAmountTextBy100(position.unrealizedPnl),
    realizedPnl: divideAmountTextBy100(position.realizedPnl),
  }
}

function normalizeMarginAccountAmounts(account: ContractMarginAccount): ContractMarginAccount {
  return {
    ...account,
    balance: divideAmountTextBy100(account.balance),
    availableBalance: divideAmountTextBy100(account.availableBalance),
    frozenBalance: divideAmountTextBy100(account.frozenBalance),
    positionMargin: divideAmountTextBy100(account.positionMargin),
    orderMargin: divideAmountTextBy100(account.orderMargin),
    unrealizedPnl: divideAmountTextBy100(account.unrealizedPnl),
    realizedPnl: divideAmountTextBy100(account.realizedPnl),
  }
}

function createClientOrderId() {
  return `web-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 10)}`
}

function orderPositionSide(symbol: TradeSymbol, side: SubmitSide) {
  if (!isContractMarket(symbol.marketType)) return POSITION_SIDE_NET
  return side === 'buy' ? POSITION_SIDE_LONG : POSITION_SIDE_SHORT
}

function clampLeverage(value: number) {
  const nextValue = Number.isFinite(value) && value > 0 ? Math.round(value) : 1
  const leverageValues = tradeLeverageValues.value
  if (leverageValues.length) {
    return leverageValues.reduce((closest, current) => {
      return Math.abs(current - nextValue) < Math.abs(closest - nextValue) ? current : closest
    }, leverageValues[0])
  }
  return Math.min(maxTradeLeverage.value, Math.max(1, nextValue))
}

function defaultTradeLeverage() {
  return clampLeverage(
    userLeverageConfig.value?.longLeverage ||
      userLeverageConfig.value?.shortLeverage ||
      activeLeverageConfig.value?.defaultLeverage ||
      tradeLeverageValues.value[0] ||
      1,
  )
}

function updateMarginMode(value: number) {
  marginMode.value = value
}

function updateLeverage(value: number) {
  const nextLeverage = clampLeverage(value)
  leverage.value = nextLeverage
  void saveUserLeverageConfig(nextLeverage)
}

async function loadUserLeverageConfig() {
  const requestId = ++leverageConfigRequestId
  const symbol = selectedTradeSymbol.value

  if (!isLoggedIn.value || !symbol || !isContractMarket(symbol.marketType)) {
    userLeverageConfig.value = null
    leverage.value = defaultTradeLeverage()
    return
  }

  try {
    const resp = await apiTradeGetLeverageConfig({
      symbolId: symbol.id,
      marketType: symbol.marketType,
      marginMode: marginMode.value,
    })
    if (requestId !== leverageConfigRequestId) return

    if (isSuccessCode(resp.code) && resp.data) {
      userLeverageConfig.value = resp.data
    } else {
      userLeverageConfig.value = null
    }
    leverage.value = defaultTradeLeverage()
  } catch (error) {
    console.warn('load leverage config failed', error)
    if (requestId === leverageConfigRequestId) {
      userLeverageConfig.value = null
      leverage.value = defaultTradeLeverage()
    }
  }
}

async function saveUserLeverageConfig(nextLeverage: number) {
  const requestId = ++saveLeverageRequestId
  const symbol = selectedTradeSymbol.value
  if (!isLoggedIn.value || !symbol || !isContractMarket(symbol.marketType)) return

  try {
    const resp = await apiTradeSetLeverage({
      symbolId: symbol.id,
      marketType: symbol.marketType,
      marginMode: marginMode.value,
      positionMode: POSITION_MODE_ONE_WAY,
      longLeverage: nextLeverage,
      shortLeverage: nextLeverage,
    })
    if (requestId !== saveLeverageRequestId) return

    if (!isSuccessCode(resp.code)) {
      tradeError.value = resp.msg || '杠杆设置失败'
      return
    }
    await loadUserLeverageConfig()
  } catch (error) {
    console.warn('set leverage failed', error)
    if (requestId === saveLeverageRequestId) {
      tradeError.value = '杠杆设置失败'
    }
  }
}

async function loadTradeSymbols() {
  const requestId = ++tradeSymbolsRequestId
  loadingTradeSymbols.value = true

  try {
    const resp = await apiTradeGetSymbolList({ status: SYMBOL_STATUS_ENABLED })
    if (requestId !== tradeSymbolsRequestId) return

    if (!isSuccessCode(resp.code)) {
      tradeSymbols.value = []
      tradeError.value = resp.msg || '交易对加载失败'
      return
    }

    tradeSymbols.value = resp.data || []
  } catch (error) {
    console.warn('load trade symbols failed', error)
    if (requestId === tradeSymbolsRequestId) {
      tradeSymbols.value = []
      tradeError.value = '交易对加载失败'
    }
  } finally {
    if (requestId === tradeSymbolsRequestId) {
      loadingTradeSymbols.value = false
    }
  }
}

async function loadTradeSymbolDetail(symbolId: number) {
  const requestId = ++tradeDetailRequestId
  loadingTradeDetail.value = true

  try {
    const resp = await apiTradeGetSymbolDetail({ symbolId })
    if (requestId !== tradeDetailRequestId) return

    if (!isSuccessCode(resp.code)) {
      tradeSymbolDetail.value = null
      tradeError.value = resp.msg || '交易对详情加载失败'
      return
    }

    tradeSymbolDetail.value = {
      symbol: resp.symbol || selectedTradeSymbol.value,
      spot: resp.spot || null,
      contract: resp.contract || null,
      leverageConfigs: resp.leverageConfigs || [],
    }
    leverage.value = defaultTradeLeverage()
    void loadUserLeverageConfig()
  } catch (error) {
    console.warn('load trade symbol detail failed', error)
    if (requestId === tradeDetailRequestId) {
      tradeSymbolDetail.value = null
      tradeError.value = '交易对详情加载失败'
    }
  } finally {
    if (requestId === tradeDetailRequestId) {
      loadingTradeDetail.value = false
    }
  }
}

async function refreshTradeAccount() {
  syncAuthState()

  const accountRequestId = ++tradeAccountRequestId
  const ordersRequestId = ++tradeOrdersRequestId
  const symbol = selectedTradeSymbol.value
  if (!isLoggedIn.value || !symbol) {
    tradeOrders.value = []
    tradePositions.value = []
    marginAccounts.value = []
    loadingTradeAccount.value = false
    loadingTradeOrders.value = false
    return
  }

  loadingTradeAccount.value = true
  loadingTradeOrders.value = true
  ordersError.value = ''

  const orderParams = {
    marketType: symbol.marketType,
    symbolId: symbol.id,
    limit: 100,
  }

  try {
    const [ordersResult, positionsResult, accountsResult] = await Promise.allSettled([
      apiTradeGetOrderList(orderParams),
      apiTradeGetPositionList({
        marketType: symbol.marketType,
        symbolId: symbol.id,
      }),
      apiTradeGetMarginAccountList({
        marketType: symbol.marketType,
        marginAsset: selectedTradeSettleAsset.value,
      }),
    ])

    if (accountRequestId !== tradeAccountRequestId) return

    if (ordersRequestId === tradeOrdersRequestId) {
      if (ordersResult.status === 'fulfilled' && isSuccessCode(ordersResult.value.code)) {
        tradeOrders.value = (ordersResult.value.data || []).map(normalizeTradeOrderAmounts)
      } else {
        tradeOrders.value = []
        ordersError.value =
          ordersResult.status === 'fulfilled'
            ? ordersResult.value.msg || '委托加载失败'
            : '委托加载失败'
      }
    }

    if (positionsResult.status === 'fulfilled' && isSuccessCode(positionsResult.value.code)) {
      tradePositions.value = (positionsResult.value.data || []).map(normalizeContractPositionAmounts)
    } else {
      tradePositions.value = []
    }

    if (accountsResult.status === 'fulfilled' && isSuccessCode(accountsResult.value.code)) {
      marginAccounts.value = (accountsResult.value.data || []).map(normalizeMarginAccountAmounts)
    } else {
      marginAccounts.value = []
    }
  } catch (error) {
    console.warn('refresh trade account failed', error)
    if (accountRequestId === tradeAccountRequestId) {
      if (ordersRequestId === tradeOrdersRequestId) {
        tradeOrders.value = []
        ordersError.value = '委托加载失败'
      }
      tradePositions.value = []
      marginAccounts.value = []
    }
  } finally {
    if (accountRequestId === tradeAccountRequestId) {
      loadingTradeAccount.value = false
    }
    if (ordersRequestId === tradeOrdersRequestId) {
      loadingTradeOrders.value = false
    }
  }
}

async function refreshTradeOrders() {
  syncAuthState()

  const requestId = ++tradeOrdersRequestId
  const symbol = selectedTradeSymbol.value
  if (!isLoggedIn.value || !symbol) {
    tradeOrders.value = []
    loadingTradeOrders.value = false
    return
  }

  ordersError.value = ''

  try {
    const resp = await apiTradeGetOrderList({
      marketType: symbol.marketType,
      symbolId: symbol.id,
      limit: 100,
    })
    if (requestId !== tradeOrdersRequestId) return

    if (!isSuccessCode(resp.code)) {
      if (!tradeOrders.value.length) {
        ordersError.value = resp.msg || '委托加载失败'
      }
      return
    }

    tradeOrders.value = (resp.data || []).map(normalizeTradeOrderAmounts)
  } catch (error) {
    console.warn('refresh trade orders failed', error)
    if (requestId === tradeOrdersRequestId) {
      if (!tradeOrders.value.length) {
        ordersError.value = '委托加载失败'
      }
    }
  } finally {
    if (requestId === tradeOrdersRequestId) {
      loadingTradeOrders.value = false
    }
  }
}

async function submitTradeOrder(side: SubmitSide) {
  if (submittingSide.value) return

  syncAuthState()
  tradeMessage.value = ''
  tradeError.value = ''

  const symbol = selectedTradeSymbol.value
  if (!isLoggedIn.value) {
    tradeError.value = '请先登录后再交易'
    return
  }
  if (!symbol) {
    tradeError.value = loadingTradeSymbols.value
      ? '交易对加载中，请稍后再试'
      : '当前品种未配置交易接口'
    return
  }

  const qty = tradeQty.value.trim()
  if (!isPositiveDecimal(qty)) {
    tradeError.value = '请输入有效数量'
    return
  }

  const isLimitOrder = orderMode.value === 'limit'
  const price = tradePrice.value.trim()
  if (isLimitOrder && !isPositiveDecimal(price)) {
    tradeError.value = '请输入有效价格'
    return
  }

  const params: PlaceOrderReq = {
    symbolId: symbol.id,
    marketType: symbol.marketType,
    side: side === 'buy' ? TRADE_SIDE_BUY : TRADE_SIDE_SELL,
    positionSide: orderPositionSide(symbol, side),
    orderType: isLimitOrder ? ORDER_TYPE_LIMIT : ORDER_TYPE_MARKET,
    timeInForce: isLimitOrder ? TIME_IN_FORCE_GTC : TIME_IN_FORCE_IOC,
    clientOrderId: createClientOrderId(),
    qty,
    orderSource: ORDER_SOURCE_WEB,
  }

  if (isLimitOrder) {
    params.price = price
  }

  if (isContractMarket(symbol.marketType)) {
    params.marginMode = marginMode.value
    params.leverage = clampLeverage(leverage.value)
    if (isPositiveDecimal(takeProfitPrice.value)) {
      params.takeProfitPrice = takeProfitPrice.value.trim()
    }
    if (isPositiveDecimal(stopLossPrice.value)) {
      params.stopLossPrice = stopLossPrice.value.trim()
    }
  }

  submittingSide.value = side
  try {
    const resp = await apiTradePlaceOrder(params)
    if (!isSuccessCode(resp.code)) {
      tradeError.value = resp.msg || '下单失败'
      return
    }

    tradeMessage.value = resp.order?.orderNo ? `委托已提交：${resp.order.orderNo}` : '委托已提交'
    tradeQty.value = ''
    tradePercent.value = 0
    await refreshTradeAccount()
  } catch (error) {
    console.warn('place trade order failed', error)
    tradeError.value = '下单失败，请稍后重试'
  } finally {
    submittingSide.value = null
  }
}

async function cancelTradeOrder(order: TradeOrder) {
  if (cancelingOrderId.value) return

  syncAuthState()
  if (!isLoggedIn.value) {
    ordersError.value = '请先登录'
    return
  }

  cancelingOrderId.value = order.id || null
  ordersError.value = ''
  tradeMessage.value = ''

  try {
    const resp = await apiTradeCancelOrder({
      orderId: order.id || undefined,
      orderNo: order.orderNo || undefined,
    })

    if (!isSuccessCode(resp.code)) {
      ordersError.value = resp.msg || '撤单失败'
      return
    }

    tradeMessage.value = '撤单已提交'
    await refreshTradeAccount()
  } catch (error) {
    console.warn('cancel trade order failed', error)
    ordersError.value = '撤单失败，请稍后重试'
  } finally {
    cancelingOrderId.value = null
  }
}
</script>

<template>
  <section class="trade-page" :aria-busy="loadingBootstrap">
    <DesktopMarketTradeView
      v-if="isDesktop"
      :selected-product="selectedProduct"
      :price-trend="priceTrend"
      :placeholder-price="placeholderPrice"
      :desktop-stats="desktopStats"
      :desktop-product-rows="desktopProductRows"
      :selected-product-key="selectedProductKey"
      :desktop-products-expanded="desktopProductsExpanded"
      :desktop-orderbook-expanded="desktopOrderbookExpanded"
      :selected-quote="selectedQuote"
      :depth-snapshot="depthSnapshot"
      :tick-snapshot="tickSnapshot"
      :kline-snapshot="klineSnapshot"
      :intervals="intervals"
      :selected-interval-name="selectedIntervalName"
      :loading-kline="loadingKline"
      :order-mode="orderMode"
      :selected-trade-symbol="selectedTradeSymbol"
      :trade-symbol-detail="tradeSymbolDetail"
      :trade-symbol-loading="loadingTradeSymbols || loadingTradeDetail"
      :is-logged-in="isLoggedIn"
      :trade-available="tradeAvailable"
      :trade-price="tradePrice"
      :trade-qty="tradeQty"
      :trade-percent="tradePercent"
      :margin-mode="marginMode"
      :leverage="leverage"
      :max-leverage="maxTradeLeverage"
      :leverage-values="tradeLeverageValues"
      :settle-asset="selectedTradeSettleAsset"
      :available-balance="availableBalance"
      :long-position-qty="longPositionQty"
      :short-position-qty="shortPositionQty"
      :trade-message="tradeMessage"
      :trade-error="tradeError"
      :submitting-side="submittingSide"
      :trade-orders="tradeOrders"
      :orders-loading="loadingTradeOrders"
      :orders-error="ordersError"
      :canceling-order-id="cancelingOrderId"
      :coin-glyph="coinGlyph"
      @toggle-products="toggleDesktopProducts"
      @toggle-orderbook="toggleDesktopOrderbook"
      @select-product="selectProduct"
      @select-interval="selectInterval"
      @update:order-mode="orderMode = $event"
      @update:trade-price="tradePrice = $event"
      @update:trade-qty="tradeQty = $event"
      @update:trade-percent="tradePercent = $event"
      @update:margin-mode="updateMarginMode"
      @update:leverage="updateLeverage"
      @submit-order="submitTradeOrder"
      @cancel-order="cancelTradeOrder"
      @refresh-orders="refreshTradeOrders"
    />
    <MobileTradeView
      v-else
      :categories="categories"
      :selected-category-type="selectedCategoryType"
      :selected-category="selectedCategory"
      :selected-product="selectedProduct"
      :selected-product-key="selectedProductKey"
      :trade-kind="tradeKind"
      :price-trend="priceTrend"
      :placeholder-price="placeholderPrice"
      :placeholder-change="placeholderChange"
      :selected-quote="selectedQuote"
      :depth-snapshot="depthSnapshot"
      :product-menu-open="productMenuOpen"
      :product-sheet-rows="productSheetRows"
      :order-mode="orderMode"
      :selected-trade-symbol="selectedTradeSymbol"
      :trade-symbol-detail="tradeSymbolDetail"
      :trade-symbol-loading="loadingTradeSymbols || loadingTradeDetail"
      :is-logged-in="isLoggedIn"
      :trade-available="tradeAvailable"
      :trade-price="tradePrice"
      :trade-qty="tradeQty"
      :trade-percent="tradePercent"
      :margin-mode="marginMode"
      :leverage="leverage"
      :max-leverage="maxTradeLeverage"
      :leverage-values="tradeLeverageValues"
      :take-profit-price="takeProfitPrice"
      :stop-loss-price="stopLossPrice"
      :settle-asset="selectedTradeSettleAsset"
      :available-balance="availableBalance"
      :long-position-qty="longPositionQty"
      :short-position-qty="shortPositionQty"
      :trade-message="tradeMessage"
      :trade-error="tradeError"
      :submitting-side="submittingSide"
      :trade-orders="tradeOrders"
      :orders-loading="loadingTradeOrders"
      :orders-error="ordersError"
      :canceling-order-id="cancelingOrderId"
      :coin-glyph="coinGlyph"
      @select-category="selectedCategoryType = $event"
      @open-product-menu="productMenuOpen = true"
      @close-product-sheet="closeProductSheet"
      @select-product="selectProduct"
      @update:order-mode="orderMode = $event"
      @update:trade-price="tradePrice = $event"
      @update:trade-qty="tradeQty = $event"
      @update:trade-percent="tradePercent = $event"
      @update:margin-mode="updateMarginMode"
      @update:leverage="updateLeverage"
      @update:take-profit-price="takeProfitPrice = $event"
      @update:stop-loss-price="stopLossPrice = $event"
      @submit-order="submitTradeOrder"
      @cancel-order="cancelTradeOrder"
      @refresh-orders="refreshTradeOrders"
    />
  </section>
</template>

<style scoped>
.trade-page {
  width: 100%;
  max-width: 100%;
  min-height: calc(100dvh - 72px);
  padding: 18px 22px 112px;
  overflow-x: hidden;
  background: #0b0c15;
  color: #f6f7fb;
}
</style>
