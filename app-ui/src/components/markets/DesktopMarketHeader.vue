<script setup lang='ts'>
import type { ItickTenantProduct } from '@/types/itick'
import type { DesktopStat } from './desktop-types'

defineProps<{
  selectedProduct: ItickTenantProduct | null
  priceTrend: 'up' | 'down'
  placeholderPrice: string
  desktopStats: DesktopStat[]
}>()
</script>

<template>
  <header class="desktop-trade-header">
    <div class="desktop-trade-header__symbol">
      <strong>{{ selectedProduct?.symbol || 'BTC/USDT' }}</strong>
      <span>04-23 05:21 New_York</span>
    </div>

    <div class="desktop-trade-header__price" :class="{ down: priceTrend === 'down' }">
      {{ placeholderPrice }} {{ priceTrend === 'down' ? '↓' : '↑' }}
    </div>

    <div class="desktop-trade-header__stats">
      <span v-for="item in desktopStats" :key="item.label">
        <em>{{ item.label }}</em>
        <strong :class="{ down: item.down }">{{ item.value }}</strong>
      </span>
    </div>

    <button type="button" class="desktop-favorite">☆</button>
  </header>
</template>

<style scoped>
.desktop-trade-header {
  display: grid;
  grid-template-columns: auto auto minmax(0, 1fr) auto;
  align-items: center;
  gap: 28px;
  padding: 20px 22px;
  border-bottom: 1px solid #242633;
}

.desktop-trade-header__symbol {
  display: grid;
  gap: 6px;
}

.desktop-trade-header__symbol strong {
  font-size: 20px;
  font-weight: 600;
}

.desktop-trade-header__symbol span {
  color: #8f929d;
  font-size: 14px;
}

.desktop-trade-header__price {
  color: #09d676;
  font-size: 22px;
  font-weight: 700;
}

.desktop-trade-header__price.down {
  color: #ff574c;
}

.desktop-trade-header__stats {
  display: grid;
  grid-template-columns: repeat(7, minmax(0, 1fr));
  gap: 18px;
}

.desktop-trade-header__stats span {
  display: grid;
  gap: 10px;
}

.desktop-trade-header__stats em {
  color: #8f929d;
  font-size: 12px;
  font-style: normal;
}

.desktop-trade-header__stats strong {
  font-size: 14px;
  font-weight: 600;
}

.desktop-trade-header__stats strong.down {
  color: #ff574c;
}

.desktop-favorite {
  border: 0;
  background: transparent;
  color: #d8dae2;
  font-size: 28px;
}
</style>
