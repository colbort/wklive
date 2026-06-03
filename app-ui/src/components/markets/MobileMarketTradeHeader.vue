<script setup lang='ts'>
import { useI18n } from '@/i18n'
import type { ItickTenantCategory, ItickTenantProduct, QuotePayload } from '@/types/itick'
import { marketCategoryLabel } from '@/utils/marketCategory'

type ProductSheetRow = {
  key: string
  product: ItickTenantProduct
  price: string
  change: string
  direction: 'up' | 'down' | 'flat'
}

defineProps<{
  categories: ItickTenantCategory[]
  selectedCategoryType: number | null
  selectedCategory: ItickTenantCategory | null
  selectedProduct: ItickTenantProduct | null
  selectedProductKey: string
  tradeKind: 'stock' | 'option' | 'forex' | 'commodity' | 'crypto'
  priceTrend: 'up' | 'down'
  placeholderPrice: string
  placeholderChange: string
  selectedQuote: QuotePayload | null
  productMenuOpen: boolean
  productSheetRows: ProductSheetRow[]
  coinGlyph: (product: ItickTenantProduct) => string
}>()

const emit = defineEmits<{
  (e: 'select-category', categoryType: number): void
  (e: 'open-product-menu'): void
  (e: 'close-product-sheet'): void
  (e: 'select-product', product: ItickTenantProduct): void
}>()

const { t } = useI18n()
</script>

<template>
  <div class="mobile-market-trade-header">
    <nav class="trade-categories" :aria-label="t('market.category')">
      <button
        v-for="category in categories"
        :key="category.id"
        type="button"
        :class="{ active: category.categoryType === selectedCategoryType }"
        @click="emit('select-category', category.categoryType)"
      >
        {{ marketCategoryLabel(category) }}
      </button>
    </nav>

    <header class="trade-symbol">
      <button type="button" class="trade-symbol__main" @click="emit('open-product-menu')">
        <strong>{{ selectedProduct?.symbol || t('market.selectProduct') }}</strong>
        <span />
      </button>

      <div v-if="tradeKind === 'stock'" class="trade-symbol__sub">
        {{ selectedProduct?.displayName || selectedProduct?.name || '--' }}
      </div>
      <div v-else class="trade-symbol__quote" :class="priceTrend">
        <span>{{ placeholderPrice }}</span>
        <em>{{ placeholderChange }}</em>
      </div>

      <div class="trade-symbol__icons">
        <button type="button">▮▮</button>
        <button type="button">☆</button>
        <button type="button">⌃</button>
      </div>

      <div v-if="productMenuOpen" class="product-sheet-overlay" @click.self="emit('close-product-sheet')">
        <section class="product-sheet">
          <span class="product-sheet__handle" />

          <header class="product-sheet__header">
            <h3>{{ marketCategoryLabel(selectedCategory) || t('market.product') }}</h3>
            <button type="button" :aria-label="t('common.close')" @click="emit('close-product-sheet')">×</button>
          </header>

          <div class="product-sheet__rows">
            <button
              v-for="row in productSheetRows"
              :key="row.key"
              type="button"
              class="product-sheet-row"
              :class="{
                'product-sheet-row--active': row.key === selectedProductKey,
                'product-sheet-row--down': row.direction === 'down',
              }"
              @click="emit('select-product', row.product)"
            >
              <span class="product-sheet-row__coin">{{ coinGlyph(row.product) }}</span>
              <span class="product-sheet-row__symbol">{{ row.product.symbol }}</span>
              <strong>{{ row.price }}</strong>
              <span class="product-sheet-row__change">
                <em>{{ row.change || '--' }}</em>
                <small>{{ row.change || t('market.waiting') }}</small>
              </span>
            </button>
          </div>

          <div class="product-sheet__footer">
            <span>{{ t('market.productCount', { count: productSheetRows.length }) }}</span>
          </div>
        </section>
      </div>
    </header>
  </div>
</template>

<style scoped>
.mobile-market-trade-header {
  max-width: calc(100% + 36px);
  margin: 0 -18px 8px;
  padding: 0 18px 1px;
}

.trade-categories {
  position: fixed;
  top: 0;
  right: 0;
  left: 0;
  z-index: 25;
  display: flex;
  flex-wrap: nowrap;
  gap: 20px;
  max-width: 100%;
  padding: 10px 22px 8px;
  overflow-x: auto;
  overflow-y: hidden;
  margin-bottom: 0;
  background: #0b0c15;
  scrollbar-width: none;
  -webkit-overflow-scrolling: touch;
}

.trade-categories::-webkit-scrollbar {
  display: none;
}

.trade-categories button,
.trade-symbol button,
.product-sheet__header button,
.product-sheet-row {
  border: 0;
  background: transparent;
  color: inherit;
  font: inherit;
}

.trade-categories button {
  flex: 0 0 auto;
  color: #8f929d;
  font-size: 15px;
  font-weight: 500;
  white-space: nowrap;
}

.trade-categories button.active {
  color: #fff;
  font-size: 17px;
  font-weight: 600;
}

.trade-symbol {
  position: relative;
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 4px 14px;
  margin-bottom: 22px;
}

.trade-symbol__main {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  justify-self: start;
  min-width: 0;
  max-width: 100%;
  padding: 0;
}

.trade-symbol__main strong {
  min-width: 0;
  overflow: hidden;
  font-size: 17px;
  font-weight: 500;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.trade-symbol__main span {
  width: 10px;
  height: 10px;
  transform: rotate(45deg) translateY(-2px);
  border-right: 2px solid currentColor;
  border-bottom: 2px solid currentColor;
}

@media (max-width: 390px) {
  .trade-categories {
    gap: 18px;
    padding-right: 14px;
    padding-left: 14px;
  }

  .mobile-market-trade-header {
    max-width: calc(100% + 28px);
    margin-right: -14px;
    margin-left: -14px;
    padding-right: 14px;
    padding-left: 14px;
  }

  .trade-categories button {
    font-size: 14px;
  }

  .trade-categories button.active {
    font-size: 16px;
  }

  .trade-symbol__icons {
    gap: 8px;
  }

  .trade-symbol__icons button {
    font-size: 20px;
  }

  .product-sheet {
    padding-right: 16px;
    padding-left: 16px;
  }

  .product-sheet-row {
    grid-template-columns: 36px minmax(0, 1fr) minmax(0, 62px) minmax(0, 58px);
    min-height: 76px;
    column-gap: 8px;
  }

  .product-sheet-row__coin {
    width: 34px;
    height: 34px;
    font-size: 17px;
  }

  .product-sheet-row__symbol,
  .product-sheet-row strong,
  .product-sheet-row__change em {
    font-size: 14px;
  }

  .product-sheet-row__change small {
    padding: 4px 6px;
    font-size: 11px;
  }
}

.trade-symbol__sub {
  color: #8f929d;
  font-size: 14px;
}

.trade-symbol__quote {
  color: #0cd977;
  font-size: 14px;
}

.trade-symbol__quote.down {
  color: #ff574c;
}

.trade-symbol__quote span {
  margin-right: 12px;
}

.trade-symbol__quote em {
  font-style: normal;
}

.trade-symbol__icons {
  grid-column: 2;
  grid-row: 1 / span 2;
  display: flex;
  align-items: start;
  gap: 14px;
}

.trade-symbol__icons button {
  color: #fff;
  font-size: 25px;
}

.product-sheet-overlay {
  position: fixed;
  inset: 0;
  z-index: 80;
  display: flex;
  align-items: flex-end;
  justify-content: center;
  padding: 0;
  background: rgba(3, 4, 10, 0.68);
  backdrop-filter: blur(7px);
}

.product-sheet {
  position: relative;
  display: flex;
  flex-direction: column;
  width: min(100%, 640px);
  max-width: 100%;
  max-height: 68dvh;
  padding: 22px 22px 26px;
  overflow: hidden;
  border-radius: 28px 28px 0 0;
  background: #22232c;
  color: #f6f7fb;
  box-shadow: 0 -24px 70px rgba(0, 0, 0, 0.42);
  touch-action: pan-y;
}

.product-sheet__handle {
  display: block;
  width: 54px;
  height: 6px;
  margin: 0 auto 22px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.52);
}

.product-sheet__header {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 42px;
  margin-bottom: 10px;
}

.product-sheet__header h3 {
  margin: 0;
  color: #fff;
  font-size: 22px;
  font-weight: 500;
}

.product-sheet__header button {
  position: absolute;
  top: 42px;
  right: 24px;
  color: #fff;
  font-size: 31px;
  line-height: 1;
  cursor: pointer;
}

.product-sheet__rows {
  flex: 1 1 auto;
  min-height: 0;
  overflow-y: auto;
  overflow-x: hidden;
  overscroll-behavior: contain;
  -webkit-overflow-scrolling: touch;
  touch-action: pan-y;
}

.product-sheet-row {
  display: grid;
  grid-template-columns: 44px minmax(0, 1fr) minmax(0, 72px) minmax(0, 64px);
  align-items: center;
  column-gap: 10px;
  width: 100%;
  min-width: 0;
  min-height: 96px;
  padding: 12px 0;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
  text-align: left;
  cursor: pointer;
}

.product-sheet-row__coin {
  display: grid;
  place-items: center;
  width: 44px;
  height: 44px;
  border-radius: 999px;
  background: linear-gradient(145deg, #4099ff, #67c2ff);
  color: #fff;
  font-size: 21px;
  font-weight: 500;
}

.product-sheet-row:nth-child(4n + 2) .product-sheet-row__coin {
  background: linear-gradient(145deg, #e9ddc9, #fff2d8);
  color: #b8346c;
}

.product-sheet-row:nth-child(4n + 3) .product-sheet-row__coin {
  background: linear-gradient(145deg, #2186dd, #2e9fff);
}

.product-sheet-row:nth-child(4n + 4) .product-sheet-row__coin {
  background: linear-gradient(145deg, #0e52ff, #3888ff);
}

.product-sheet-row__symbol {
  overflow: hidden;
  color: #fff;
  font-size: 17px;
  font-weight: 500;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.product-sheet-row strong {
  min-width: 0;
  overflow: hidden;
  color: #09d676;
  font-size: 16px;
  font-weight: 500;
  text-align: right;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.product-sheet-row__change {
  display: grid;
  justify-items: end;
  gap: 6px;
  min-width: 0;
  overflow: hidden;
}

.product-sheet-row__change em {
  max-width: 100%;
  overflow: hidden;
  color: #09d676;
  font-size: 15px;
  font-style: normal;
  font-weight: 500;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.product-sheet-row__change small {
  width: 100%;
  max-width: 100%;
  min-width: 0;
  padding: 5px 9px;
  overflow: hidden;
  border-radius: 14px;
  background: #06d171;
  color: #fff;
  font-size: 13px;
  text-align: center;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.product-sheet-row--down strong,
.product-sheet-row--down .product-sheet-row__change em {
  color: #ff5148;
}

.product-sheet-row--down .product-sheet-row__change small {
  background: #ff4438;
}

.product-sheet-row--active {
  background: rgba(255, 255, 255, 0.025);
}

.product-sheet__footer {
  display: flex;
  justify-content: center;
  min-height: 26px;
  padding-top: 12px;
  color: #8f929d;
  font-size: 14px;
}
</style>
