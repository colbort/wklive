<script setup lang="ts">
import { computed, ref } from 'vue'

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
const activeTab = ref<OrderTab>('open')
const openStatuses = new Set([1, 2])
const filteredOrders = computed(() => {
  return props.orders.filter((order) => {
    const isOpen = openStatuses.has(order.status)
    return activeTab.value === 'open' ? isOpen : !isOpen
  })
})

function orderSideText(order: TradeOrder) {
  if (order.positionSide === 2) return t('trade.openLong')
  if (order.positionSide === 3) return t('trade.openShort')
  return order.side === 1 ? t('trade.buy') : t('trade.sell')
}

function statusText(status: number) {
  const labels: Record<number, string> = {
    1: t('trade.pending'),
    2: t('trade.partiallyFilled'),
    3: t('trade.filled'),
    4: t('trade.canceled'),
    5: t('trade.rejected'),
    6: t('trade.expired'),
  }
  return labels[status] || t('trade.unknown')
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
      <button v-if="showPremarket" type="button">{{ t('trade.premarketOrders') }}</button>
      <button class="trade-orders-panel__refresh" type="button" @click="emit('refresh')">
        {{ t('trade.refresh') }}
      </button>
    </div>

    <LoginPrompt v-if="!isLoggedIn" :action-text="t('assets.viewData')" compact />
    <p v-else-if="!selectedTradeSymbol" class="trade-orders-panel__state">
      {{ t('trade.unavailable') }}
    </p>
    <p v-else-if="loading" class="trade-orders-panel__state">{{ t('trade.orderLoading') }}</p>
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
      >
        <div>
          <strong>{{
            selectedTradeSymbol?.displaySymbol || selectedTradeSymbol?.symbol || '--'
          }}</strong>
          <span>{{ orderSideText(order) }} / {{ orderTypeText(order.orderType) }}</span>
        </div>
        <div>
          <strong>{{ order.price || t('trade.market') }}</strong>
          <span>{{ order.qty }} / {{ order.filledQty || '0' }}</span>
        </div>
        <div>
          <strong>{{ statusText(order.status) }}</strong>
          <span>{{ formatTime(order.createTimes) }}</span>
        </div>
        <button
          v-if="openStatuses.has(order.status)"
          type="button"
          :disabled="cancelingOrderId === order.id"
          @click="emit('cancel-order', order)"
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
  border-top: 1px solid rgba(255, 255, 255, 0.08);
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
  color: #ff6b5f;
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
  background: #151823;
}

.trade-orders-panel__item div {
  display: grid;
  gap: 4px;
  min-width: 0;
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
