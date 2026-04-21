<script setup lang="ts">
import type { ItickTenantCategory, ItickTenantProduct, ItickWsConnectionState } from '@/types/itick'
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
}>()

const emit = defineEmits<{
  selectCategory: [categoryType: number]
  selectProduct: [product: ItickTenantProduct]
}>()

function coinGlyph(product: ItickTenantProduct) {
  const coin = product.baseCoin || product.symbol.slice(0, 3) || product.displayName
  return coin.slice(0, 1).toUpperCase()
}

function formatNumber(value?: number | null, digits = 2) {
  if (value === null || value === undefined || !Number.isFinite(value)) return '--'

  return new Intl.NumberFormat('zh-CN', {
    minimumFractionDigits: digits,
    maximumFractionDigits: digits,
  }).format(value)
}

function formatPercent(value: number) {
  return `${value >= 0 ? '+' : ''}${value.toFixed(2)}%`
}
</script>

<template>
  <section class="quotes-view">
    <div class="category-strip">
      <button
        v-for="category in categories"
        :key="category.id"
        type="button"
        class="category-tab"
        :class="{ 'category-tab--active': category.categoryType === selectedCategoryType }"
        @click="emit('selectCategory', category.categoryType)"
      >
        {{ category.categoryName }}
      </button>
    </div>

    <div class="connection-row">
      <span class="connection-dot" :class="`connection-dot--${wsState}`" />
      <span>{{ selectedCategoryName || '分类加载中' }}</span>
      <strong>{{ wsError || selectedCategoryCode || '等待 categoryCode' }}</strong>
    </div>

    <div v-if="loading" class="empty-state">正在加载行情...</div>
    <div v-else-if="!rows.length" class="empty-state">当前分类暂无可见产品。</div>

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
        <span class="coin-mark">{{ coinGlyph(row.product) }}</span>

        <span class="quote-row__name">
          <strong>{{ row.product.symbol }}</strong>
          <small>{{ row.product.market }} · {{ row.product.categoryCode || selectedCategoryCode }}</small>
        </span>

        <strong class="quote-row__price">
          {{ row.quote ? formatNumber(row.quote.lastPrice, 4) : '--' }}
        </strong>

        <span class="quote-row__change">
          <strong>{{ row.quote ? formatNumber(row.quote.lastPrice - row.quote.open, 4) : '--' }}</strong>
          <em>{{ row.quote ? formatPercent(row.changeRate) : '等待' }}</em>
        </span>
      </button>
    </template>
  </section>
</template>

<style scoped>
.quotes-view {
  padding-bottom: 18px;
}

.category-strip {
  display: flex;
  gap: 18px;
  overflow-x: auto;
  padding: 14px 18px 0;
  border-bottom: 1px solid #21232e;
}

.category-tab,
.quote-row {
  border: 0;
  background: transparent;
  color: inherit;
  font: inherit;
  cursor: pointer;
}

.category-tab {
  position: relative;
  flex: 0 0 auto;
  padding: 0 0 14px;
  color: #8c8f99;
  font-size: 17px;
  font-weight: 500;
  white-space: nowrap;
}

.category-tab--active {
  color: #ffffff;
  font-weight: 600;
}

.category-tab--active::after {
  position: absolute;
  right: 2px;
  bottom: 0;
  left: 2px;
  height: 3px;
  border-radius: 999px 999px 0 0;
  background: #08c200;
  content: '';
}

.connection-row {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 18px 0;
  color: #818691;
  font-size: 12px;
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
  background: #08c200;
  box-shadow: 0 0 12px rgba(8, 194, 0, 0.58);
}

.connection-dot--closed {
  background: #e45656;
}

.quote-row {
  display: grid;
  grid-template-columns: 52px minmax(96px, 1fr) minmax(82px, 0.75fr) minmax(76px, 0.65fr);
  align-items: center;
  width: calc(100% - 36px);
  min-height: 82px;
  margin: 0 18px;
  padding: 12px 0;
  border-bottom: 1px solid #20222d;
  text-align: left;
}

.coin-mark {
  display: grid;
  place-items: center;
  width: 38px;
  height: 38px;
  border-radius: 999px;
  background: linear-gradient(145deg, #ff9a16, #f2b728);
  color: #fff;
  font-size: 19px;
  font-weight: 500;
}

.quote-row:nth-of-type(4n + 1) .coin-mark {
  background: linear-gradient(145deg, #8b66ff, #a582ff);
}

.quote-row:nth-of-type(4n + 2) .coin-mark {
  background: linear-gradient(145deg, #85bf62, #97d579);
}

.quote-row:nth-of-type(4n + 3) .coin-mark {
  background: linear-gradient(145deg, #1d2430, #101722);
}

.quote-row__name {
  display: grid;
  gap: 4px;
  min-width: 0;
}

.quote-row__name strong {
  overflow: hidden;
  color: #fff;
  font-size: 17px;
  font-weight: 500;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.quote-row__name small {
  color: #5f6570;
  font-size: 11px;
}

.quote-row__price {
  color: #09d676;
  font-size: 17px;
  font-weight: 600;
  text-align: right;
}

.quote-row__change {
  display: grid;
  gap: 5px;
  justify-items: end;
}

.quote-row__change strong {
  color: #09d676;
  font-size: 15px;
  font-weight: 500;
}

.quote-row__change em {
  min-width: 62px;
  padding: 5px 8px;
  border-radius: 13px;
  background: #06d171;
  color: #fff;
  font-size: 13px;
  font-style: normal;
  font-weight: 500;
  text-align: center;
}

.quote-row--down .quote-row__price,
.quote-row--down .quote-row__change strong {
  color: #ff5959;
}

.quote-row--down .quote-row__change em {
  background: #ff4d4d;
}

.quote-row--active {
  background: rgba(255, 255, 255, 0.015);
}

.empty-state {
  display: grid;
  place-items: center;
  min-height: 260px;
  padding: 32px 18px;
  color: #8f929d;
  font-size: 14px;
  text-align: center;
}
</style>
