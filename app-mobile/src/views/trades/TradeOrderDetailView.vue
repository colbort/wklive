<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { apiTradeGetOrderDetail } from '@/api/trade'
import CommonPage from '@/components/common/CommonPage.vue'
import { useI18n } from '@/i18n'
import { useOrderEvents } from '@/composables/useOrderEvents'
import type {
  TradeOrder,
  TradeOrderContract,
  TradeOrderSeconds,
  TradeOrderSpot,
} from '@/types/trade'
import { formatAssetDecimalAmount } from '@/utils/assetAmount'

const route = useRoute()
const router = useRouter()
const { locale, t } = useI18n()
const loading = ref(false)
const error = ref('')
const order = ref<TradeOrder | null>(null)
const spot = ref<TradeOrderSpot | null>(null)
const contract = ref<TradeOrderContract | null>(null)
const seconds = ref<TradeOrderSeconds | null>(null)
let detailRefreshInFlight = false
useOrderEvents(
  (event) => {
    if (event.domain !== 'trade') return
    if (event.biz_id && order.value?.id && event.biz_id !== order.value.id) return
    void pollDetail()
  },
  () => {
    void pollDetail()
  },
)

const orderNo = computed(() => String(route.params.orderNo || ''))
const symbol = computed(() => String(route.query.symbol || '--'))
const priceScale = computed(() => normalizeScale(route.query.priceScale, 8))
const qtyScale = computed(() => normalizeScale(route.query.qtyScale, 8))

function normalizeScale(value: unknown, fallback: number) {
  const scale = Number(value)
  return Number.isInteger(scale) && scale >= 0 && scale <= 18 ? scale : fallback
}

function formatTradeDecimal(value: string, scale: number) {
  return formatAssetDecimalAmount(value, scale).replace(/(\.\d*?[1-9])0+$|\.0+$/, '$1')
}

function ok(code: number) {
  return code === 0 || code === 200
}

async function loadDetail(silent = false) {
  if (!orderNo.value) {
    order.value = null
    error.value = t('trade.orderDetailLoadFailed')
    return
  }
  if (!silent) loading.value = true
  if (!silent) error.value = ''
  try {
    const resp = await apiTradeGetOrderDetail({ orderNo: orderNo.value })
    if (!ok(resp.code) || !resp.data?.order) {
      if (!silent || !order.value) {
        error.value = resp.msg || t('trade.orderDetailLoadFailed')
      }
      return
    }
    error.value = ''
    order.value = resp.data.order
    spot.value = resp.data.spot?.id ? resp.data.spot : null
    contract.value = resp.data.contract?.id ? resp.data.contract : null
    seconds.value = resp.data.seconds?.id ? resp.data.seconds : null
  } catch (cause) {
    console.warn('load trade order detail failed', cause)
    if (!silent || !order.value) {
      error.value = t('trade.orderDetailLoadFailed')
    }
  } finally {
    if (!silent) loading.value = false
  }
}

function isTerminalOrder(value: TradeOrder) {
  if ([8, 9, 11, 13, 15, 16].includes(value.displayStatus)) return true
  return [3, 4, 5, 6].includes(value.status)
}

async function pollDetail() {
  if (detailRefreshInFlight || (order.value && isTerminalOrder(order.value))) return
  detailRefreshInFlight = true
  try {
    await loadDetail(true)
  } finally {
    detailRefreshInFlight = false
  }
}

function handleDetailFocus() {
  void pollDetail()
}

function handleDetailVisibilityChange() {
  if (document.visibilityState === 'visible') {
    void pollDetail()
  }
}

function statusText(value: TradeOrder) {
  const display: Record<number, string> = {
    1: t('trade.freezing'), 2: t('trade.activating'), 3: t('trade.active'),
    4: t('trade.triggerWaiting'), 5: t('trade.pending'), 6: t('trade.partiallyFilled'),
    7: t('trade.settling'), 8: t('trade.filled'), 9: t('trade.settled'),
    10: t('trade.canceling'), 11: t('trade.orderCanceled'), 12: t('trade.expiring'),
    13: t('trade.orderExpired'), 14: t('trade.refunding'), 15: t('trade.refunded'),
    16: t('trade.rejected'), 17: t('trade.manualReview'),
  }
  let status = value.displayStatus
  if (value.productType !== 3 && value.status === 4 && (Number(value.filledQty) > 0 || Number(value.filledAmount) > 0)) {
    status = 6
  } else if (!status) {
    if (value.productType === 3 && value.status === 3) status = 9
    else if (value.productType === 3 && value.status === 4) status = 15
    else if (value.productType === 3 && value.status === 5) status = 16
    else status = ({ 1: 5, 2: 6, 3: 8, 4: 11, 5: 16, 6: 13, 7: 1, 8: 4, 9: 10, 10: 12, 11: 7 } as Record<number, number>)[value.status] || 0
  }
  return display[status] || t('trade.unknown')
}

function effectiveDisplayStatus(value: TradeOrder) {
  if (
    value.productType !== 3 &&
    value.status === 4 &&
    (Number(value.filledQty) > 0 || Number(value.filledAmount) > 0)
  ) {
    return 6
  }
  if (value.displayStatus > 0) return value.displayStatus
  if (value.productType === 3) {
    if (value.status === 3) return 9
    if (value.status === 4) return 15
    if (value.status === 5) return 16
  }
  return ({ 1: 5, 2: 6, 3: 8, 4: 11, 5: 16, 6: 13, 7: 1, 8: 4, 9: 10, 10: 12, 11: 7 } as Record<number, number>)[value.status] || 0
}

function statusClass(value: TradeOrder) {
  const status = effectiveDisplayStatus(value)
  if (status === 8 || status === 9) return 'is-success'
  if (status === 11 || status === 13 || status === 15) return 'is-neutral'
  if (status === 16) return 'is-danger'
  if ([1, 2, 3, 4, 5, 6, 7, 10, 12, 14, 17].includes(status)) return 'is-warning'
  return ''
}

function productText(value: TradeOrder) {
  if (value.productType === 1) return t('trade.marketModeSpot')
  if (value.productType === 3) return t('trade.marketModeSeconds')
  const kind =
    value.contractType === 2 ? t('trade.contractDelivery') : t('trade.contractPerpetual')
  const valueType =
    value.contractValueType === 2 ? t('trade.contractInverse') : t('trade.contractLinear')
  return `${kind} · ${valueType}`
}

function sideText(value: TradeOrder) {
  if (value.productType === 3) {
    return value.secondsDirection === 2 ? t('trade.buyDown') : t('trade.buyUp')
  }
  if (value.positionSide === 2) return t('trade.openLong')
  if (value.positionSide === 3) return t('trade.openShort')
  return value.side === 1 ? t('trade.buy') : t('trade.sell')
}

function orderTypeText(value: number) {
  if (value === 1) return t('trade.limit')
  if (value === 2) return t('trade.market')
  return t('trade.conditional')
}

function displayOrderQty(value: TradeOrder) {
  if (Number(value.qty) > 0) return formatTradeDecimal(value.qty, qtyScale.value)
  // 现货市价买单按计价币金额下单，委托时没有确定的基础币数量。
  // 成交后以实际 filledQty 作为详情页的数量，避免把原始占位值 0 展示给用户。
  if (
    value.productType === 1 &&
    value.orderType === 2 &&
    value.side === 1 &&
    Number(value.amount) > 0
  ) {
    return Number(value.filledQty) > 0
      ? formatTradeDecimal(value.filledQty, qtyScale.value)
      : '--'
  }
  return value.qty || '0'
}

function displayOrderPrice(value: TradeOrder) {
  if (Number(value.price) > 0) return formatTradeDecimal(value.price, priceScale.value)
  if (value.orderType === 2) {
    return Number(value.avgPrice) > 0
      ? formatTradeDecimal(value.avgPrice, priceScale.value)
      : t('trade.market')
  }
  return value.price || '0'
}

function formatTime(value?: number) {
  if (!value) return '--'
  const ms = value > 9999999999 ? value : value * 1000
  return new Date(ms).toLocaleString(locale.value)
}

watch(
  orderNo,
  () => {
    void loadDetail()
  },
  { immediate: true },
)

onMounted(() => {
  window.addEventListener('focus', handleDetailFocus)
  document.addEventListener('visibilitychange', handleDetailVisibilityChange)
})

onBeforeUnmount(() => {
  window.removeEventListener('focus', handleDetailFocus)
  document.removeEventListener('visibilitychange', handleDetailVisibilityChange)
})
</script>

<template>
  <CommonPage
    class="trade-order-detail-page"
    :title="t('trade.orderDetail')"
    :nav-height="76"
    @back="router.back()"
  >
    <main class="order-detail">
      <p v-if="loading" class="order-detail__state">{{ t('common.loading') }}</p>
      <p v-else-if="error" class="order-detail__state order-detail__state--error">{{ error }}</p>

      <template v-else-if="order">
        <section class="order-detail__hero">
          <div>
            <span>{{ productText(order) }}</span>
            <strong>{{ symbol }}</strong>
            <small>{{ sideText(order) }} / {{ orderTypeText(order.orderType) }}</small>
          </div>
          <b class="order-detail__status" :class="statusClass(order)">{{ statusText(order) }}</b>
        </section>

        <section class="order-detail__card">
          <h2>{{ t('trade.orderInfo') }}</h2>
          <dl>
            <div><dt>{{ t('trade.orderNo') }}</dt><dd>{{ order.orderNo }}</dd></div>
            <div v-if="order.productType !== 3"><dt>{{ t('trade.price') }}</dt><dd>{{ displayOrderPrice(order) }}</dd></div>
            <div v-if="order.productType !== 3"><dt>{{ t('trade.qty') }}</dt><dd>{{ displayOrderQty(order) }}</dd></div>
            <div><dt>{{ t('trade.orderAmount') }}</dt><dd>{{ order.amount || '0' }}</dd></div>
            <div v-if="order.productType !== 3"><dt>{{ t('trade.filledQty') }}</dt><dd>{{ formatTradeDecimal(order.filledQty || '0', qtyScale) }}</dd></div>
            <div v-if="order.productType !== 3"><dt>{{ t('trade.filledAmount') }}</dt><dd>{{ order.filledAmount || '0' }}</dd></div>
            <div v-if="order.productType !== 3"><dt>{{ t('trade.avgPrice') }}</dt><dd>{{ formatTradeDecimal(order.avgPrice || '0', priceScale) }}</dd></div>
            <div><dt>{{ t('trade.fee') }}</dt><dd>{{ order.fee || '0' }} {{ order.feeAsset }}</dd></div>
            <div><dt>{{ t('trade.createTime') }}</dt><dd>{{ formatTime(order.createTimes) }}</dd></div>
            <div><dt>{{ t('trade.updateTime') }}</dt><dd>{{ formatTime(order.updateTimes) }}</dd></div>
            <div v-if="order.cancelReason" class="wide">
              <dt>{{ t('trade.cancelReason') }}</dt><dd>{{ order.cancelReason }}</dd>
            </div>
          </dl>
        </section>

        <section v-if="spot" class="order-detail__card">
          <h2>{{ t('trade.spotSettlement') }}</h2>
          <dl>
            <div><dt>{{ t('trade.frozenAsset') }}</dt><dd>{{ spot.frozenAsset || '--' }}</dd></div>
            <div><dt>{{ t('trade.frozenAmount') }}</dt><dd>{{ spot.frozenAmount || '0' }}</dd></div>
            <div><dt>{{ t('trade.settleAsset') }}</dt><dd>{{ spot.settleAsset || '--' }}</dd></div>
            <div><dt>{{ t('trade.settleAmount') }}</dt><dd>{{ spot.settleAmount || '0' }}</dd></div>
          </dl>
        </section>

        <section v-if="contract" class="order-detail__card">
          <h2>{{ t('trade.contractInfo') }}</h2>
          <dl>
            <div><dt>{{ t('trade.marginMode') }}</dt><dd>{{ contract.marginMode === 2 ? t('trade.isolated') : t('trade.cross') }}</dd></div>
            <div><dt>{{ t('trade.leverage') }}</dt><dd>{{ contract.leverage }}x</dd></div>
            <div><dt>{{ t('trade.margin') }}</dt><dd>{{ contract.marginAmount }} {{ contract.marginAsset }}</dd></div>
            <div><dt>{{ t('trade.liquidationPrice') }}</dt><dd>{{ contract.liquidationPrice || '--' }}</dd></div>
            <div><dt>{{ t('trade.takeProfitPrice') }}</dt><dd>{{ contract.takeProfitPrice || '--' }}</dd></div>
            <div><dt>{{ t('trade.stopLossPrice') }}</dt><dd>{{ contract.stopLossPrice || '--' }}</dd></div>
          </dl>
        </section>

        <section v-if="seconds" class="order-detail__card">
          <h2>{{ t('trade.secondsInfo') }}</h2>
          <dl>
            <div><dt>{{ t('trade.direction') }}</dt><dd>{{ seconds.direction === 2 ? t('trade.buyDown') : t('trade.buyUp') }}</dd></div>
            <div><dt>{{ t('trade.duration') }}</dt><dd>{{ seconds.durationSeconds }}s</dd></div>
            <div><dt>{{ t('trade.investmentAmount') }}</dt><dd>{{ seconds.stakeAmount }} {{ seconds.stakeAsset }}</dd></div>
            <div><dt>{{ t('trade.payoutRate') }}</dt><dd>{{ seconds.payoutRate }}</dd></div>
            <div><dt>{{ t('trade.startPrice') }}</dt><dd>{{ seconds.startPrice || '--' }}</dd></div>
            <div><dt>{{ t('trade.settlementPrice') }}</dt><dd>{{ seconds.settlementPrice || '--' }}</dd></div>
            <div><dt>{{ t('trade.profitAmount') }}</dt><dd>{{ seconds.profitAmount || '0' }}</dd></div>
            <div><dt>{{ t('trade.returnAmount') }}</dt><dd>{{ seconds.returnAmount || '0' }}</dd></div>
          </dl>
        </section>
      </template>
    </main>
  </CommonPage>
</template>

<style scoped>
.trade-order-detail-page {
  height: 100vh;
  height: 100dvh;
  min-height: 100vh;
  min-height: 100dvh;
}

.order-detail {
  display: grid;
  gap: 14px;
  padding: 16px;
}

.order-detail__state {
  padding: 60px 16px;
  text-align: center;
  color: var(--muted);
}

.order-detail__state--error {
  color: var(--danger);
}

.order-detail__hero,
.order-detail__card {
  border: 1px solid var(--divider-soft);
  border-radius: 14px;
  background: var(--page-bg-soft);
}

.order-detail__hero {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  padding: 18px;
}

.order-detail__hero div {
  display: grid;
  gap: 5px;
}

.order-detail__hero span,
.order-detail__hero small {
  color: var(--muted);
}

.order-detail__status {
  align-self: flex-start;
  padding: 4px 8px;
  border-radius: 5px;
}

.order-detail__status.is-success {
  color: #52c41a;
  background: rgba(82, 196, 26, 0.12);
}

.order-detail__status.is-warning {
  color: #faad14;
  background: rgba(250, 173, 20, 0.12);
}

.order-detail__status.is-neutral {
  color: #9298a3;
  background: rgba(146, 152, 163, 0.12);
}

.order-detail__status.is-danger {
  color: #ff5b57;
  background: rgba(255, 91, 87, 0.12);
}

.order-detail__card {
  padding: 16px;
}

.order-detail__card h2 {
  margin: 0 0 14px;
  font-size: 0.9rem;
}

.order-detail__card dl {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px 12px;
  margin: 0;
}

.order-detail__card dl div {
  min-width: 0;
}

.order-detail__card dl .wide {
  grid-column: 1 / -1;
}

.order-detail__card dt {
  margin-bottom: 5px;
  color: var(--muted);
  font-size: 0.7rem;
}

.order-detail__card dd {
  overflow-wrap: anywhere;
  margin: 0;
  color: var(--text);
  font-size: 0.75rem;
}
</style>
