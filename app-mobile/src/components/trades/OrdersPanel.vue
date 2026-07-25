<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'

import LoginPrompt from '@/components/common/LoginPrompt.vue'
import { useI18n } from '@/i18n'
import type { TradeOrder, TradeSymbol } from '@/types/trade'

type OrderTab = 'open' | 'history'

const props = withDefaults(
  defineProps<{
    showPremarket?: boolean
    orders?: TradeOrder[]
    loading?: boolean
    error?: string
    isLoggedIn?: boolean
    selectedTradeSymbol?: TradeSymbol | null
    cancelingOrderId?: number | null
  }>(),
  {
    showPremarket: false,
    orders: () => [],
    loading: false,
    error: '',
    isLoggedIn: false,
    selectedTradeSymbol: null,
    cancelingOrderId: null,
  },
)

const emit = defineEmits<{
  (e: 'cancel-order', order: TradeOrder): void
  (e: 'refresh'): void
}>()

const { locale, t } = useI18n()
const router = useRouter()
const activeTab = ref<OrderTab>('open')
const openStatuses = new Set([1, 2])
const filteredOrders = computed(() => {
  return props.orders.filter((order) => {
    const isOpen = openStatuses.has(order.status)
    return activeTab.value === 'open' ? isOpen : !isOpen
  })
})

function orderSideText(order: TradeOrder) {
  if (order.productType === 3) {
    return order.secondsDirection === 2 ? t('trade.buyDown') : t('trade.buyUp')
  }
  if (order.positionSide === 2) return t('trade.openLong')
  if (order.positionSide === 3) return t('trade.openShort')
  return order.side === 1 ? t('trade.buy') : t('trade.sell')
}

function statusText(order: TradeOrder) {
  const labels: Record<number, string> = {
    1: t('trade.freezing'),
    2: t('trade.activating'),
    3: t('trade.active'),
    4: t('trade.triggerWaiting'),
    5: t('trade.pending'),
    6: t('trade.partiallyFilled'),
    7: t('trade.settling'),
    8: t('trade.filled'),
    9: t('trade.settled'),
    10: t('trade.canceling'),
    11: t('trade.orderCanceled'),
    12: t('trade.expiring'),
    13: t('trade.orderExpired'),
    14: t('trade.refunding'),
    15: t('trade.refunded'),
    16: t('trade.rejected'),
    17: t('trade.manualReview'),
  }
  return labels[effectiveDisplayStatus(order)] || t('trade.unknown')
}

function effectiveDisplayStatus(order: TradeOrder) {
  if (
    order.productType !== 3 &&
    order.status === 4 &&
    (Number(order.filledQty) > 0 || Number(order.filledAmount) > 0)
  ) {
    return 6
  }
  if (order.displayStatus > 0) return order.displayStatus
  if (order.productType === 3) {
    if (order.status === 3) return 9
    if (order.status === 4) return 15
    if (order.status === 5) return 16
  }
  return (
    {
      1: 5,
      2: 6,
      3: 8,
      4: 11,
      5: 16,
      6: 13,
      7: 1,
      8: 4,
      9: 10,
      10: 12,
      11: 7,
    }[order.status] || 0
  )
}

function statusClass(order: TradeOrder) {
  const status = effectiveDisplayStatus(order)
  if (status === 8 || status === 9) return 'is-success'
  if (status === 11 || status === 13 || status === 15) return 'is-neutral'
  if (status === 16) return 'is-danger'
  if ([1, 2, 3, 4, 5, 6, 7, 10, 12, 14, 17].includes(status)) return 'is-warning'
  return ''
}

function orderTypeText(orderType: number) {
  if (orderType === 1) return t('trade.limit')
  if (orderType === 2) return t('trade.market')
  return t('trade.conditional')
}

function formatTime(value: number) {
  if (!value) return '--'
  const ms = value > 9999999999 ? value : value * 1000
  return new Date(ms).toLocaleString(locale.value, {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function productText(order: TradeOrder) {
  if (order.productType === 1) return t('trade.marketModeSpot')
  if (order.productType === 3) return t('trade.marketModeSeconds')
  const contract =
    order.contractType === 2 ? t('trade.contractDelivery') : t('trade.contractPerpetual')
  const valueType =
    order.contractValueType === 2 ? t('trade.contractInverse') : t('trade.contractLinear')
  return `${contract} · ${valueType}`
}

function primaryValue(order: TradeOrder) {
  if (order.productType === 3) return order.amount || '0'
  if (order.orderType === 2) return order.avgPrice && order.avgPrice !== '0' ? order.avgPrice : t('trade.market')
  return order.price || '--'
}

function secondaryValue(order: TradeOrder) {
  if (order.productType === 3) return `${order.durationSeconds || 0}s`
  return `${order.qty || '0'} / ${order.filledQty || '0'}`
}

function orderSummaryText(order: TradeOrder) {
  if (order.productType === 3) return `${orderSideText(order)} / ${t('trade.duration')}`
  return `${orderSideText(order)} / ${orderTypeText(order.orderType)}`
}

function openOrderDetail(order: TradeOrder) {
  void router.push({
    name: 'trade-order-detail',
    params: { orderNo: order.orderNo },
    query: {
      symbol: props.selectedTradeSymbol?.displaySymbol || props.selectedTradeSymbol?.symbol || '',
    },
  })
}
</script>

<template>
  <section class="trade-orders-panel">
    <div class="trade-orders-panel__nav">
      <button :class="{ active: activeTab === 'open' }" type="button" @click="activeTab = 'open'">
        {{ t('trade.openOrders') }}
      </button>
      <button
        :class="{ active: activeTab === 'history' }"
        type="button"
        @click="activeTab = 'history'"
      >
        {{ t('trade.historyOrders') }}
      </button>
      <button v-if="showPremarket" type="button">
        {{ t('trade.premarketOrders') }}
      </button>
      <button class="trade-orders-panel__refresh" type="button" @click="emit('refresh')">
        {{ t('trade.refresh') }}
      </button>
    </div>

    <LoginPrompt v-if="!isLoggedIn" :action-text="t('assets.viewData')" compact />
    <p v-else-if="!selectedTradeSymbol" class="trade-orders-panel__state">
      {{ t('trade.unavailable') }}
    </p>
    <p v-else-if="loading" class="trade-orders-panel__state">
      {{ t('trade.orderLoading') }}
    </p>
    <p v-else-if="error" class="trade-orders-panel__state trade-orders-panel__state--error">
      {{ error }}
    </p>
    <p v-else-if="!filteredOrders.length" class="trade-orders-panel__state">
      {{ t('common.none') }}
    </p>

    <ul v-else class="trade-orders-panel__list">
      <li
        v-for="order in filteredOrders"
        :key="order.id || order.orderNo"
        class="trade-orders-panel__item"
        tabindex="0"
        role="button"
        @click="openOrderDetail(order)"
        @keydown.enter="openOrderDetail(order)"
      >
        <div>
          <strong>{{
            selectedTradeSymbol?.displaySymbol || selectedTradeSymbol?.symbol || '--'
          }}</strong>
          <span>{{ productText(order) }}</span>
        </div>
        <div>
          <strong>{{ orderSummaryText(order) }}</strong>
          <span>{{ primaryValue(order) }} · {{ secondaryValue(order) }}</span>
        </div>
        <div>
          <strong class="trade-orders-panel__status" :class="statusClass(order)">
            {{ statusText(order) }}
          </strong>
          <span>{{ formatTime(order.createTimes) }}</span>
        </div>
        <button
          v-if="openStatuses.has(order.status)"
          type="button"
          :disabled="cancelingOrderId === order.id"
          @click.stop="emit('cancel-order', order)"
        >
          {{ cancelingOrderId === order.id ? t('trade.canceling') : t('trade.cancel') }}
        </button>
      </li>
    </ul>
  </section>
</template>

<style scoped>
.trade-orders-panel__nav {
  display: flex;
  gap: 22px;
  padding: 20px 0 0;
  border-top: 1px solid var(--divider-soft);
  align-items: flex-start;
}

button {
  position: relative;
  padding: 0 0 14px;
  border: 0;
  background: transparent;
  color: var(--muted);
  font: inherit;
  font-size: 0.85rem;
  font-weight: 700;
}

.trade-orders-panel__refresh {
  margin-left: auto;
  font-size: 0.7rem;
}

button.active {
  color: var(--text);
}

button.active::after {
  position: absolute;
  right: 6px;
  bottom: 0;
  left: 6px;
  height: 3px;
  border-radius: 999px;
  background: var(--accent);
  content: '';
}

.trade-orders-panel__state {
  display: grid;
  min-height: 88px;
  margin: 0;
  place-items: center;
  color: var(--muted);
  font-size: 0.7rem;
}

.trade-orders-panel__state--error {
  color: var(--danger);
}

.trade-orders-panel__list {
  display: grid;
  gap: 10px;
  margin: 0;
  padding: 14px 0 0;
  list-style: none;
}

.trade-orders-panel__item {
  display: grid;
  grid-template-columns: minmax(0, 1.15fr) minmax(0, 0.95fr) minmax(0, 0.9fr) auto;
  gap: 10px;
  align-items: center;
  min-width: 0;
  padding: 12px;
  border-radius: 8px;
  background: var(--page-bg-soft);
  cursor: pointer;
  outline: none;
}

.trade-orders-panel__item:active,
.trade-orders-panel__item:focus-visible {
  box-shadow: inset 3px 0 0 var(--accent);
}

.trade-orders-panel__item div {
  display: grid;
  gap: 4px;
  min-width: 0;
}

.trade-orders-panel__item > div:last-of-type {
  justify-items: end;
  text-align: right;
}

.trade-orders-panel__item strong,
.trade-orders-panel__item span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.trade-orders-panel__item strong {
  color: var(--text);
  font-size: 0.65rem;
  font-weight: 600;
}

.trade-orders-panel__status {
  display: inline-flex;
  width: fit-content;
  padding: 2px 6px;
  border-radius: 4px;
}

.trade-orders-panel__status.is-success {
  color: #52c41a;
  background: rgba(82, 196, 26, 0.12);
}

.trade-orders-panel__status.is-warning {
  color: #faad14;
  background: rgba(250, 173, 20, 0.12);
}

.trade-orders-panel__status.is-neutral {
  color: #9298a3;
  background: rgba(146, 152, 163, 0.12);
}

.trade-orders-panel__status.is-danger {
  color: #ff5b57;
  background: rgba(255, 91, 87, 0.12);
}

.trade-orders-panel__item span {
  color: var(--muted);
  font-size: 0.6rem;
}

.trade-orders-panel__item button {
  padding: 0;
  color: var(--accent);
  font-size: 0.65rem;
}

.trade-orders-panel__item button:disabled {
  color: var(--muted);
}

@media (max-width: 390px) {
  .trade-orders-panel__nav {
    gap: 14px;
  }

  button {
    font-size: 0.75rem;
  }

  .trade-orders-panel__item {
    grid-template-columns: 1fr 1fr;
  }
}
</style>
