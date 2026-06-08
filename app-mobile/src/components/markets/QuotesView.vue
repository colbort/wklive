<script setup lang="ts">
import { getLocale, useI18n } from '@/i18n'
import type { ItickTenantCategory, ItickTenantProduct, ItickWsConnectionState } from '@/types/itick'
import { marketCategoryCodeLabel, marketCategoryLabel } from '@/utils/marketCategory'
import type { MarketRow } from './types'

defineProps<{
  categories: ItickTenantCategory[]
  selectedCategoryType: number | null
  selectedCategoryName: string
  selectedCategoryCode: string
  wsState: ItickWsConnectionState
  wsError: string
  loading: boolean
  rows: MarketRow[]
  selectedProductKey: string
  categoryPinned?: boolean
}>()

const emit = defineEmits<{
  selectCategory: [categoryType: number]
  selectProduct: [product: ItickTenantProduct]
}>()

const { t } = useI18n()

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
</script>

<template>
  <section class="quotes-view" :class="{ 'quotes-view--category-pinned': categoryPinned }">
    <div class="category-strip">
      <button
        v-for="category in categories"
        :key="category.id"
        type="button"
        class="category-tab app-menu__item"
        :class="{
          'category-tab--active': category.categoryType === selectedCategoryType,
          'app-menu__item--active': category.categoryType === selectedCategoryType,
        }"
        @click="emit('selectCategory', category.categoryType)"
      >
        {{ marketCategoryLabel(category) }}
      </button>
    </div>

    <div class="connection-row">
      <span class="connection-dot" :class="`connection-dot--${wsState}`" />
      <span>{{
        marketCategoryCodeLabel(selectedCategoryCode, selectedCategoryName) ||
          t('market.categoryLoading')
      }}</span>
      <strong>{{ wsError || selectedCategoryCode || t('market.waitingCategoryCode') }}</strong>
    </div>

    <div v-if="loading" class="empty-state">
      {{ t('market.loadingQuotes') }}
    </div>
    <div v-else-if="!rows.length" class="empty-state">
      {{ t('market.noVisibleProducts') }}
    </div>

    <template v-else>
      <button
        v-for="row in rows"
        :key="row.key"
        type="button"
        class="quote-row"
        :class="{
          'quote-row--active': row.key === selectedProductKey,
          'quote-row--down': row.direction === 'down',
        }"
        @click="emit('selectProduct', row.product)"
      >
        <span class="quote-row__icon">
          <img v-if="row.product.icon" :src="row.product.icon" :alt="row.product.symbol">
          <span v-else>{{ productIconText(row.product) }}</span>
        </span>

        <span class="quote-row__name">
          <strong>{{ row.product.symbol }}</strong>
        </span>

        <strong class="quote-row__price">
          {{ row.quote ? formatPrice(row.quote.lastPrice) : '--' }}
        </strong>

        <span class="quote-row__change">
          <strong>{{
            row.quote ? formatPrice(row.quote.lastPrice - row.quote.open) : '--'
          }}</strong>
          <em>{{ row.quote ? formatPercent(row.changeRate) : t('market.waiting') }}</em>
        </span>
      </button>
    </template>
  </section>
</template>

<style scoped>
.quotes-view {
  width: 100%;
  max-width: 100%;
  padding-bottom: 18px;
}

.quotes-view--category-pinned {
  padding-top: 57px;
}

.category-strip {
  position: -webkit-sticky;
  position: sticky;
  top: 0;
  left: 0;
  right: 0;
  z-index: 30;
  width: 100%;
  max-width: 100%;
  box-sizing: border-box;
  display: flex;
  flex-wrap: nowrap;
  gap: 28px;
  overflow-x: auto;
  overflow-y: hidden;
  padding: 0 28px 0;
  border-bottom: 1px solid var(--divider);
  background: var(--page-bg);
  scrollbar-width: none;
  overscroll-behavior-x: contain;
  -webkit-overflow-scrolling: touch;
}

.category-strip::-webkit-scrollbar {
  display: none;
}

.quotes-view--category-pinned .category-strip {
  position: fixed;
  top: 0;
  left: 50%;
  right: auto;
  z-index: 80;
  width: min(100%, var(--app-width, 414px));
  transform: translateX(-50%);
}

.category-tab {
  position: relative;
  flex: 0 0 auto;
  padding: 0 0 16px;
  border: 0;
  background: transparent;
  cursor: pointer;
  line-height: 1.2;
  white-space: nowrap;
}

.category-tab--active::after {
  position: absolute;
  right: 2px;
  bottom: 0;
  left: 2px;
  content: '';
}

.connection-row {
  display: none;
  align-items: center;
  gap: 8px;
  padding: 10px 18px 0;
  color: #818691;
  font-size: 0.6rem;
}

.connection-row strong {
  overflow: hidden;
  color: #5f6570;
  font-weight: 500;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.connection-dot {
  width: 7px;
  height: 7px;
  border-radius: 999px;
  background: #585d69;
}

.connection-dot--connecting {
  background: #f2c94c;
}

.connection-dot--open {
  background: var(--accent);
  box-shadow: 0 0 12px rgba(8, 194, 0, 0.58);
}

.connection-dot--closed {
  background: #e45656;
}

.quote-row {
  display: grid;
  grid-template-columns: 54px minmax(0, 1fr) minmax(88px, 0.42fr) minmax(76px, 0.36fr);
  align-items: center;
  column-gap: 8px;
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

@media (max-width: 390px) {
  .category-strip {
    gap: 24px;
    padding-right: 22px;
    padding-left: 22px;
  }

  .quote-row {
    grid-template-columns: 48px minmax(0, 1fr) minmax(76px, 0.42fr) minmax(68px, 0.36fr);
    column-gap: 6px;
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
}

.quote-row__icon {
  display: grid;
  width: 45px;
  height: 45px;
  place-items: center;
  overflow: hidden;
  border-radius: 50%;
  background: #202631;
  color: #fff;
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
  color: #fff;
  font-size: 0.64rem;
  font-style: normal;
  font-weight: 600;
  text-align: center;
}

.quote-row--down .quote-row__price,
.quote-row--down .quote-row__change strong {
  color: #ff4f43;
}

.quote-row--down .quote-row__change em {
  background: #ff4f43;
  color: #fff;
}

.quote-row--active {
  background: rgba(255, 255, 255, 0.015);
}

.empty-state {
  display: grid;
  place-items: center;
  min-height: 260px;
  padding: 32px 18px;
  color: var(--muted);
  font-size: 0.7rem;
  text-align: center;
}
</style>
