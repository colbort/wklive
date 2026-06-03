<script setup lang="ts">
import { computed, ref } from 'vue'

import { useI18n } from '@/i18n'
import type { TradeOrder, TradeSymbol } from '@/types/trade'

type OrderTab = 'open' | 'history'

const props = withDefaults(defineProps<{
  orders?: TradeOrder[]
  loading?: boolean
  error?: string
  isLoggedIn?: boolean
  selectedTradeSymbol?: TradeSymbol | null
  cancelingOrderId?: number | null
}>(), {
  orders: () => [],
  loading: false,
  error: '',
  isLoggedIn: false,
  selectedTradeSymbol: null,
  cancelingOrderId: null,
})

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
  <section class="desktop-bottom-panel">
    <header class="desktop-bottom-panel__tabs">
      <button type="button" :class="{ active: activeTab === 'open' }" @click="activeTab = 'open'">{{ t('trade.openOrders') }}</button>
      <button type="button" :class="{ active: activeTab === 'history' }" @click="activeTab = 'history'">{{ t('trade.orderHistory') }}</button>
      <button type="button" class="desktop-bottom-panel__refresh" @click="emit('refresh')">{{ t('trade.refresh') }}</button>
    </header>

    <div class="desktop-bottom-panel__table">
      <div class="desktop-bottom-panel__head">
        <span>{{ t('trade.contractName') }}</span>
        <span>{{ t('trade.direction') }}</span>
        <span>{{ t('trade.price') }}</span>
        <span>{{ t('trade.qtyFilled') }}</span>
        <span>{{ t('trade.avgPrice') }}</span>
        <span>{{ t('trade.status') }}</span>
        <span>{{ t('trade.time') }}</span>
        <span>{{ t('trade.action') }}</span>
      </div>

      <div v-if="!isLoggedIn" class="desktop-bottom-panel__empty">
        <div class="desktop-bottom-panel__empty-icon">▤</div>
        <p>{{ t('trade.loginOrRegisterViewData') }}</p>
      </div>
      <div v-else-if="!selectedTradeSymbol" class="desktop-bottom-panel__empty">
        <div class="desktop-bottom-panel__empty-icon">▤</div>
        <p>{{ t('trade.unavailable') }}</p>
      </div>
      <div v-else-if="loading" class="desktop-bottom-panel__empty">
        <div class="desktop-bottom-panel__empty-icon">▤</div>
        <p>{{ t('trade.orderLoading') }}</p>
      </div>
      <div v-else-if="error" class="desktop-bottom-panel__empty desktop-bottom-panel__empty--error">
        <div class="desktop-bottom-panel__empty-icon">▤</div>
        <p>{{ error }}</p>
      </div>
      <div v-else-if="!filteredOrders.length" class="desktop-bottom-panel__empty">
        <div class="desktop-bottom-panel__empty-icon">▤</div>
        <p>{{ t('common.none') }}</p>
      </div>

      <div v-else class="desktop-bottom-panel__body">
        <div v-for="order in filteredOrders" :key="order.id || order.orderNo" class="desktop-bottom-panel__row">
          <span>{{ selectedTradeSymbol?.displaySymbol || selectedTradeSymbol?.symbol || '--' }}</span>
          <span>{{ orderSideText(order) }} / {{ orderTypeText(order.orderType) }}</span>
          <span>{{ order.price || t('trade.market') }}</span>
          <span>{{ order.qty }} / {{ order.filledQty || '0' }}</span>
          <span>{{ order.avgPrice || '--' }}</span>
          <span>{{ statusText(order.status) }}</span>
          <span>{{ formatTime(order.createTimes) }}</span>
          <button
            v-if="openStatuses.has(order.status)"
            type="button"
            :disabled="cancelingOrderId === order.id"
            @click="emit('cancel-order', order)"
          >
            {{ cancelingOrderId === order.id ? t('trade.canceling') : t('trade.cancel') }}
          </button>
          <span v-else>--</span>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.desktop-bottom-panel {
  display: grid;
  grid-template-rows: auto minmax(0, 1fr);
  min-height: 0;
  border-top: 1px solid #242633;
  background: #0b0c15;
}

.desktop-bottom-panel__tabs {
  display: flex;
  gap: 28px;
  padding: 14px 18px 0;
}

.desktop-bottom-panel__tabs button {
  position: relative;
  border: 0;
  background: transparent;
  color: #8f929d;
  font: inherit;
  font-size: 15px;
  font-weight: 500;
  padding-bottom: 16px;
}

.desktop-bottom-panel__tabs button.active {
  color: #fff;
}

.desktop-bottom-panel__tabs button.active::after {
  position: absolute;
  right: 0;
  bottom: 0;
  left: 0;
  height: 4px;
  border-radius: 999px;
  background: #fff;
  content: '';
}

.desktop-bottom-panel__refresh {
  margin-left: auto;
}

.desktop-bottom-panel__table {
  display: grid;
  grid-template-rows: auto minmax(0, 1fr);
  min-height: 0;
}

.desktop-bottom-panel__head {
  display: grid;
  grid-template-columns: 1.2fr 0.75fr 0.65fr 0.9fr 0.7fr 0.7fr 0.9fr 0.5fr;
  gap: 18px;
  padding: 16px 18px;
  border-top: 1px solid #242633;
  color: #8f929d;
  font-size: 13px;
}

.desktop-bottom-panel__empty {
  display: grid;
  place-items: center;
  align-content: center;
  gap: 12px;
  color: #7d828e;
}

.desktop-bottom-panel__empty--error {
  color: #ff6b5f;
}

.desktop-bottom-panel__empty-icon {
  width: 72px;
  height: 72px;
  display: grid;
  place-items: center;
  border-radius: 999px;
  background: radial-gradient(circle at 50% 35%, rgba(255, 255, 255, 0.16), rgba(255, 255, 255, 0.02));
  color: #cfd3dc;
  font-size: 28px;
}

.desktop-bottom-panel__body {
  min-height: 0;
  overflow-y: auto;
}

.desktop-bottom-panel__row {
  display: grid;
  grid-template-columns: 1.2fr 0.75fr 0.65fr 0.9fr 0.7fr 0.7fr 0.9fr 0.5fr;
  gap: 18px;
  align-items: center;
  min-height: 44px;
  padding: 0 18px;
  border-top: 1px solid #181b25;
  color: #d7dbe5;
  font-size: 13px;
}

.desktop-bottom-panel__row span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.desktop-bottom-panel__row button {
  border: 0;
  background: transparent;
  color: #02b904;
  font: inherit;
  font-size: 13px;
}

.desktop-bottom-panel__row button:disabled {
  color: #7d828e;
}
</style>
