<script setup lang='ts'>
import type { ItickTenantProduct } from '@/types/itick'
import { useI18n } from '@/i18n'
import type { DesktopProductRow } from './desktop-types'

defineProps<{
  rows: DesktopProductRow[]
  selectedProductKey: string
  coinGlyph: (product: ItickTenantProduct) => string
}>()

const emit = defineEmits<{
  (e: 'select-product', product: ItickTenantProduct): void
}>()

const { t } = useI18n()
</script>

<template>
  <aside class="desktop-products-panel">
    <header class="desktop-products-panel__head">
      <nav>
        <button type="button">{{ t('market.favorite') }}</button>
        <button type="button" class="active">{{ t('market.all') }}</button>
      </nav>
    </header>

    <div class="desktop-products-panel__list">
      <button
        v-for="row in rows"
        :key="row.key"
        type="button"
        class="desktop-product-row"
        :class="{ 'desktop-product-row--active': row.key === selectedProductKey }"
        @click="emit('select-product', row.product)"
      >
        <span class="desktop-product-row__coin">{{ coinGlyph(row.product) }}</span>
        <span class="desktop-product-row__symbol">{{ row.product.symbol }}</span>
        <span class="desktop-product-row__quote">
          <strong>{{ row.price }}</strong>
          <small :class="{ down: row.direction === 'down' }">{{ row.change }}</small>
        </span>
      </button>
    </div>
  </aside>
</template>

<style scoped>
.desktop-products-panel {
  display: grid;
  grid-template-rows: auto 1fr auto;
  width: 100%;
  min-height: 0;
  border-right: 1px solid #242633;
}

.desktop-products-panel__head {
  padding: 18px 16px 10px;
  border-bottom: 1px solid #242633;
}

.desktop-products-panel__head nav {
  display: flex;
  gap: 24px;
}

.desktop-products-panel__head button {
  position: relative;
  border: 0;
  background: transparent;
  color: #8f929d;
  font: inherit;
  font-size: 17px;
  font-weight: 500;
}

.desktop-products-panel__head button.active {
  color: #fff;
}

.desktop-products-panel__head button.active::after {
  position: absolute;
  right: 0;
  bottom: -12px;
  left: 0;
  height: 3px;
  border-radius: 999px;
  background: #fff;
  content: '';
}

.desktop-products-panel__list {
  min-height: 0;
  overflow-y: auto;
  overscroll-behavior: contain;
}

.desktop-product-row {
  display: grid;
  grid-template-columns: 44px minmax(0, 1fr) minmax(96px, 108px);
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 14px 14px;
  border: 0;
  border-bottom: 1px solid #242633;
  background: transparent;
  color: inherit;
  text-align: left;
  font: inherit;
}

.desktop-product-row__coin {
  display: grid;
  place-items: center;
  width: 44px;
  height: 44px;
  border-radius: 999px;
  background: linear-gradient(145deg, #ff9a16, #f2b728);
  color: #fff;
  font-size: 22px;
  font-weight: 700;
}

.desktop-product-row:nth-child(4n + 2) .desktop-product-row__coin {
  background: linear-gradient(145deg, #8b66ff, #a582ff);
}

.desktop-product-row:nth-child(4n + 3) .desktop-product-row__coin {
  background: linear-gradient(145deg, #85bf62, #97d579);
}

.desktop-product-row:nth-child(4n + 4) .desktop-product-row__coin {
  background: linear-gradient(145deg, #1d2430, #101722);
}

.desktop-product-row__symbol {
  overflow: hidden;
  font-size: 14px;
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.desktop-product-row strong {
  color: #ff574c;
  font-size: 14px;
  font-weight: 700;
  text-align: right;
}

.desktop-product-row__quote {
  display: grid;
  justify-items: end;
  gap: 6px;
  min-width: 0;
}

.desktop-product-row small {
  min-width: 0;
  padding: 4px 7px;
  border-radius: 10px;
  background: #ff574c;
  color: #fff;
  font-size: 11px;
  text-align: center;
  white-space: nowrap;
}
</style>
