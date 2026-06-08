<script setup lang='ts'>
import BottomDrawer from '@/components/common/BottomDrawer.vue'
import { useI18n } from '@/i18n'
import type { ItickTenantCategory, ItickTenantProduct } from '@/types/itick'
import { marketCategoryLabel } from '@/utils/marketCategory'

type ProductSheetRow = {
  key: string
  product: ItickTenantProduct
  price: string
  change: string
  direction: 'up' | 'down' | 'flat'
}

defineProps<{
  selectedCategory: ItickTenantCategory | null
  selectedProduct: ItickTenantProduct | null
  selectedProductKey: string
  tradeKind: 'stock' | 'option' | 'forex' | 'commodity' | 'crypto'
  priceTrend: 'up' | 'down'
  placeholderPrice: string
  placeholderChange: string
  productMenuOpen: boolean
  productSheetRows: ProductSheetRow[]
  coinGlyph: (product: ItickTenantProduct) => string
}>()

const emit = defineEmits<{
  (e: 'open-product-menu'): void
  (e: 'close-product-sheet'): void
  (e: 'select-product', product: ItickTenantProduct): void
}>()

const { t } = useI18n()
</script>

<template>
  <div class="trade-header">
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

      <BottomDrawer
        :model-value="productMenuOpen"
        :title="marketCategoryLabel(selectedCategory) || t('market.product')"
        :close-label="t('common.close')"
        max-height="68dvh"
        :z-index="80"
        @update:model-value="value => { if (!value) emit('close-product-sheet') }"
      >
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

        <template #footer>
          <div class="product-sheet__footer">
            <span>{{ t('market.productCount', { count: productSheetRows.length }) }}</span>
          </div>
        </template>
      </BottomDrawer>
    </header>
  </div>
</template>

<style scoped>
.trade-header {
  max-width: calc(100% + 36px);
  margin: 0 -18px 8px;
  padding: 0 18px 1px;
}

.trade-symbol button,
.product-sheet-row {
  border: 0;
  background: transparent;
  color: inherit;
  font: inherit;
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
  .trade-header {
    max-width: calc(100% + 28px);
    margin-right: -14px;
    margin-left: -14px;
    padding-right: 14px;
    padding-left: 14px;
  }

  .trade-symbol__icons {
    gap: 8px;
  }

  .trade-symbol__icons button {
    font-size: 20px;
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
  color: var(--muted);
  font-size: 14px;
}

.trade-symbol__quote {
  color: var(--success);
  font-size: 14px;
}

.trade-symbol__quote.down {
  color: var(--danger-strong);
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
  color: var(--text);
  font-size: 25px;
}

.product-sheet__rows {
  display: grid;
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
  color: var(--text);
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
  color: var(--text);
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
  color: var(--text);
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
  color: var(--muted);
  font-size: 14px;
}
</style>
