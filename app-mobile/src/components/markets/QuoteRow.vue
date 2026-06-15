<script setup lang="ts">
import { computed } from 'vue'

import { getLocale, useI18n } from '@/i18n'
import { useSystemStore } from '@/stores/system'
import type { ItickTenantProduct, QuotePayload } from '@/types/itick'
import { resolveSystemAssetUrl } from '@/utils/assetUrl'

const props = withDefaults(
  defineProps<{
    product: ItickTenantProduct
    quote?: QuotePayload | null
    changeRate?: number
    priceText?: string
    changeText?: string
    percentText?: string
    direction?: 'up' | 'down' | 'flat'
    active?: boolean
    variant?: 'page' | 'sheet'
  }>(),
  {
    quote: null,
    changeRate: undefined,
    priceText: '',
    changeText: '',
    percentText: '',
    direction: 'flat',
    variant: 'page',
  },
)

const emit = defineEmits<{
  select: [product: ItickTenantProduct]
}>()

const { t } = useI18n()
const systemStore = useSystemStore()
const iconUrl = computed(() => resolveIconUrl(props.product.icon))

function formatNumber(value?: number | null, digits = 2) {
  if (value === null || value === undefined || !Number.isFinite(value)) return '--'

  return new Intl.NumberFormat(getLocale(), {
    minimumFractionDigits: digits,
    maximumFractionDigits: digits,
  }).format(value)
}

function formatPrice(value?: number | null) {
  if (value === null || value === undefined || !Number.isFinite(value)) return '--'

  return formatNumber(value, Math.abs(value) >= 1 ? 4 : 8)
}

function formatPercent(value: number) {
  return `${value >= 0 ? '+' : ''}${value.toFixed(2)}%`
}

function productIconText(product: ItickTenantProduct) {
  return (product.baseCoin || product.symbol || product.code || '?').slice(0, 2).toUpperCase()
}

function resolveIconUrl(icon?: string) {
  return resolveSystemAssetUrl(systemStore.systemCore.assetUrl, icon)
}

function displayPrice() {
  return props.priceText || (props.quote ? formatPrice(props.quote.lastPrice) : '--')
}

function displayChange() {
  if (props.changeText) return props.changeText
  return props.quote ? formatPrice(props.quote.lastPrice - props.quote.open) : '--'
}

function displayPercent() {
  if (props.percentText) return props.percentText
  return props.quote && props.changeRate !== undefined
    ? formatPercent(props.changeRate)
    : t('market.waiting')
}
</script>

<template>
  <button
    type="button"
    class="quote-row"
    :class="{
      'quote-row--active': active,
      'quote-row--down': direction === 'down',
      'quote-row--sheet': variant === 'sheet',
    }"
    @click="emit('select', product)"
  >
    <span class="quote-row__icon">
      <img v-if="product.icon" :src="iconUrl" :alt="product.symbol">
      <span v-else>{{ productIconText(product) }}</span>
    </span>

    <span class="quote-row__name">
      <strong>{{ product.symbol }}</strong>
    </span>

    <strong class="quote-row__price">
      {{ displayPrice() }}
    </strong>

    <span class="quote-row__change">
      <strong>{{ displayChange() }}</strong>
      <em>{{ displayPercent() }}</em>
    </span>
  </button>
</template>

<style scoped>
.quote-row {
  display: grid;
  grid-template-columns: 55px minmax(0, 1fr) minmax(88px, 0.42fr) minmax(76px, 0.36fr);
  align-items: center;
  column-gap: 6px;
  width: calc(100% - 36px);
  min-height: 86px;
  margin: 0 18px;
  padding: 15px 0;
  border: 0;
  border-bottom: 1px solid var(--divider);
  background: transparent;
  color: inherit;
  cursor: pointer;
  font: inherit;
  text-align: left;
}

.quote-row__icon {
  display: grid;
  width: 55px;
  height: 55px;
  place-items: center;
  overflow: hidden;
  border-radius: 50%;
  background: var(--panel-bg-muted);
  color: var(--text-strong);
  font-size: 0.72rem;
  font-weight: 700;
}

.quote-row__icon img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.quote-row__name {
  display: grid;
  min-width: 0;
}

.quote-row__name strong {
  overflow: hidden;
  color: var(--text);
  font-size: 0.9rem;
  font-weight: 600;
  line-height: 1.12;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.quote-row__price {
  overflow: hidden;
  color: var(--success);
  font-size: 0.86rem;
  font-weight: 600;
  text-align: right;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.quote-row__change {
  display: grid;
  gap: 6px;
  justify-items: end;
}

.quote-row__change strong {
  overflow: hidden;
  max-width: 100%;
  color: var(--success);
  font-size: 0.82rem;
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.quote-row__change em {
  min-width: 54px;
  padding: 4px 8px;
  border-radius: 999px;
  background: var(--success);
  color: var(--text-strong);
  font-size: 0.64rem;
  font-style: normal;
  font-weight: 600;
  text-align: center;
}

.quote-row--down .quote-row__price,
.quote-row--down .quote-row__change strong {
  color: var(--danger-strong);
}

.quote-row--down .quote-row__change em {
  background: var(--danger-strong);
  color: var(--text-strong);
}

.quote-row--active {
  background: var(--row-hover-bg);
}

.quote-row--sheet {
  width: 100%;
  min-height: 86px;
  margin: 0;
}

@media (max-width: 390px) {
  .quote-row {
    grid-template-columns: 55px minmax(0, 1fr) minmax(76px, 0.42fr) minmax(68px, 0.36fr);
    column-gap: 5px;
    width: calc(100% - 32px);
    min-height: 80px;
    margin: 0 16px;
  }

  .quote-row__price,
  .quote-row__change strong {
    font-size: 0.76rem;
  }

  .quote-row__change em {
    font-size: 0.6rem;
  }

  .quote-row--sheet {
    width: 100%;
    margin: 0;
  }
}
</style>
