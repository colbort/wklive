<script setup lang="ts">
import { computed } from 'vue'

import { useI18n } from '@/i18n'
import type {
  DepthLevel,
  DepthPayload,
  ItickTenantProduct,
  QuotePayload,
  TickPayload,
} from '@/types/itick'

const props = defineProps<{
  selectedProduct: ItickTenantProduct | null
  depthSnapshot: DepthPayload | null
  selectedQuote: QuotePayload | null
  tickSnapshot: TickPayload[]
  placeholderPrice: string
}>()

const { t } = useI18n()

const askRows = computed(() => props.depthSnapshot?.asks ?? [])
const bidRows = computed(() => props.depthSnapshot?.bids ?? [])
const latestPrice = computed(
  () => props.tickSnapshot[0]?.lastPrice || props.selectedQuote?.lastPrice || null,
)
const previousPrice = computed(() => {
  return props.tickSnapshot[1]?.lastPrice || props.selectedQuote?.open || null
})
const fallbackMidPrice = computed(() => {
  if (props.placeholderPrice && props.placeholderPrice !== '--') return props.placeholderPrice
  return formatDepthNumber(bidRows.value[0]?.price ?? askRows.value[0]?.price)
})
const midPrice = computed(() => formatDepthNumber(latestPrice.value) || fallbackMidPrice.value)
const previousMidPrice = computed(() => formatDepthNumber(previousPrice.value) || midPrice.value)
const midDirection = computed<'up' | 'down' | 'flat'>(() => {
  if (!latestPrice.value || !previousPrice.value) return 'flat'
  if (latestPrice.value > previousPrice.value) return 'up'
  if (latestPrice.value < previousPrice.value) return 'down'
  return 'flat'
})
const midArrow = computed(() => {
  if (midDirection.value === 'up') return '↑'
  if (midDirection.value === 'down') return '↓'
  return ''
})

function formatDepthNumber(value: number | null | undefined) {
  if (!value || !Number.isFinite(value)) return '--'
  return Number(value.toFixed(8)).toString()
}

function formatDepthVolume(level: DepthLevel) {
  return formatDepthNumber(level.volume)
}
</script>

<template>
  <aside class="order-book-preview">
    <header>
      <span>{{ t('market.price') }}<br />(USDT)</span>
      <span>{{ t('market.qty') }}<br />({{ selectedProduct?.baseCoin || 'BTC' }})</span>
    </header>
    <div class="depth-tools" aria-hidden="true">
      <i class="depth-tools__bid" />
      <i class="depth-tools__both" />
      <i class="depth-tools__ask" />
    </div>
    <div class="asks">
      <p v-for="level in askRows" :key="`ask-${level.price}-${level.volume}`">
        <span>{{ formatDepthNumber(level.price) }}</span
        ><strong>{{ formatDepthVolume(level) }}</strong>
      </p>
      <p v-if="!askRows.length" class="empty-row"><span>--</span><strong>--</strong></p>
    </div>
    <div class="mid-price" :class="`mid-price--${midDirection}`">
      <div>
        <strong>{{ midPrice }} {{ midArrow }}</strong>
        <span>{{ previousMidPrice }}</span>
      </div>
      <i aria-hidden="true" />
    </div>
    <div class="bids">
      <p v-for="level in bidRows" :key="`bid-${level.price}-${level.volume}`">
        <span>{{ formatDepthNumber(level.price) }}</span
        ><strong>{{ formatDepthVolume(level) }}</strong>
      </p>
      <p v-if="!bidRows.length" class="empty-row"><span>--</span><strong>--</strong></p>
    </div>
  </aside>
</template>

<style scoped>
.order-book-preview {
  min-width: 0;
  overflow: hidden;
}

.order-book-preview header,
.order-book-preview p,
.mid-price {
  display: flex;
  justify-content: space-between;
  gap: 10px;
}

.order-book-preview header {
  padding-bottom: 8px;
  margin-bottom: 10px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
  color: var(--muted);
  font-size: 0.6rem;
  line-height: 1.15;
  text-align: right;
}

.depth-tools {
  display: flex;
  gap: 9px;
  height: 14px;
  margin-bottom: 10px;
}

.depth-tools i {
  display: block;
  width: 14px;
  height: 14px;
  background:
    linear-gradient(#13d383 0 0) 0 0 / 6px 6px no-repeat,
    linear-gradient(#2f85ff 0 0) 8px 0 / 6px 6px no-repeat,
    linear-gradient(#2f85ff 0 0) 0 8px / 6px 6px no-repeat,
    linear-gradient(#13d383 0 0) 8px 8px / 6px 6px no-repeat;
}

.depth-tools__ask {
  filter: hue-rotate(130deg) saturate(1.4);
}

.order-book-preview p {
  position: relative;
  min-height: 22px;
  padding: 1px 6px;
  margin: 0;
  overflow: hidden;
  font-size: 0.65rem;
  font-weight: 700;
  line-height: 20px;
}

.order-book-preview p::after {
  position: absolute;
  top: 0;
  right: 0;
  bottom: 0;
  z-index: -1;
  width: 46%;
  background: rgba(255, 68, 56, 0.14);
  content: '';
}

.order-book-preview .empty-row {
  color: var(--muted);
}

.order-book-preview .empty-row::after {
  display: none;
}

.order-book-preview strong {
  color: var(--text);
  font-weight: 700;
  text-align: right;
}

.asks span {
  color: var(--danger-strong);
}

.bids span,
.mid-price--up {
  color: var(--success);
}

.mid-price--down {
  color: var(--danger-strong);
}

.mid-price--flat {
  color: var(--text);
}

.bids p::after {
  background: rgba(16, 210, 122, 0.14);
}

.mid-price {
  margin: 10px 0 8px;
  font-weight: 700;
}

.mid-price div {
  display: grid;
  gap: 1px;
}

.mid-price strong {
  color: currentColor;
  font-size: 1rem;
  line-height: 1;
  text-align: left;
}

.mid-price span {
  color: var(--muted);
  font-size: 0.65rem;
  text-align: left;
}

.mid-price i {
  min-width: 0;
}

@media (max-width: 390px) {
  .order-book-preview header {
    font-size: 0.55rem;
  }

  .order-book-preview p {
    min-height: 20px;
    padding-right: 4px;
    padding-left: 4px;
    font-size: 0.55rem;
    line-height: 18px;
  }

  .mid-price strong {
    font-size: 0.85rem;
  }
}
</style>
