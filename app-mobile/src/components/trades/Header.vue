<script setup lang="ts">
import BottomDrawer from '@/components/common/BottomDrawer.vue'
import AppIcon from '@/components/common/AppIcon.vue'
import QuoteRow from '@/components/markets/QuoteRow.vue'
import { useI18n } from '@/i18n'
import type { ItickTenantCategory, ItickTenantProduct } from '@/types/itick'
import { marketCategoryLabel } from '@/utils/marketCategory'

type ProductSheetRow = {
  key: string
  product: ItickTenantProduct
  price: string
  changeValue?: string
  changePercent?: string
  change: string
  direction: 'up' | 'down' | 'flat'
}

defineProps<{
  selectedCategory: ItickTenantCategory | null
  selectedProduct: ItickTenantProduct | null
  selectedProductKey: string
  tradeKind: 'stock' | 'option' | 'forex' | 'commodity' | 'crypto'
  priceTrend: 'up' | 'down' | 'flat'
  placeholderPrice: string
  placeholderChange: string
  productMenuOpen: boolean
  productSheetRows: ProductSheetRow[]
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
      <div v-else class="trade-symbol__quote" :class="`trade-symbol__quote--${priceTrend}`">
        <span>{{ placeholderPrice }}</span>
        <em>{{ placeholderChange }}</em>
      </div>

      <div class="trade-symbol__icons">
        <button type="button" :aria-label="t('trade.pause')">
          <AppIcon name="pause" class="trade-symbol__icon" />
        </button>
        <button type="button" :aria-label="t('market.addWatchlist')">
          <AppIcon name="star" class="trade-symbol__icon" />
        </button>
        <button type="button" :aria-label="t('common.collapse')">
          <AppIcon name="chevron-up" class="trade-symbol__icon" />
        </button>
      </div>

      <BottomDrawer
        :model-value="productMenuOpen"
        :title="marketCategoryLabel(selectedCategory) || t('market.product')"
        :close-label="t('common.close')"
        max-height="68dvh"
        :z-index="80"
        @update:model-value="
          (value) => {
            if (!value) emit('close-product-sheet')
          }
        "
      >
        <div class="product-sheet__rows">
          <QuoteRow
            v-for="row in productSheetRows"
            :key="row.key"
            :product="row.product"
            :price-text="row.price"
            :change-text="row.changeValue || '--'"
            :percent-text="row.changePercent || t('market.waiting')"
            :direction="row.direction"
            :active="row.key === selectedProductKey"
            variant="sheet"
            @select="emit('select-product', $event)"
          />
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
  margin: 10px -18px 8px;
  padding: 0 18px 1px;
}

.trade-symbol button {
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
  font-size: 1rem;
  font-weight: 600;
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
}

.trade-symbol__sub {
  color: var(--muted);
  font-size: 0.7rem;
}

.trade-symbol__quote {
  color: var(--success);
  font-size: 0.7rem;
}

.trade-symbol__quote--up {
  color: var(--success);
}

.trade-symbol__quote--down {
  color: var(--danger-strong);
}

.trade-symbol__quote--flat {
  color: var(--text);
}

.trade-symbol__quote span {
  margin-right: 12px;
}

.trade-symbol__quote em {
  font-style: normal;
}

.trade-symbol__icons {
  gap: 8px;
  grid-column: 2;
  grid-row: 1 / span 2;
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 18px;
  min-height: 1.7rem;
  padding-top: 0.15rem;
}

.trade-symbol__icons button {
  display: grid;
  width: 1.35rem;
  height: 1.35rem;
  place-items: center;
  padding: 0;
  color: var(--text);
}

.trade-symbol__icon {
  width: 1.2rem;
  height: 1.2rem;
}

.trade-symbol__icons button:first-child .trade-symbol__icon {
  width: 1.08rem;
  height: 1.08rem;
}

.product-sheet__rows {
  display: grid;
}

.product-sheet__footer {
  display: flex;
  justify-content: center;
  min-height: 26px;
  padding-top: 12px;
  color: var(--muted);
  font-size: 0.7rem;
}
</style>
