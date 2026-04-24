<script setup lang="ts">
import type { ItickTenantProduct } from '@/types/itick'

defineProps<{
  selectedProduct: ItickTenantProduct | null
}>()
</script>

<template>
  <aside class="order-book-preview">
    <header>
      <span>价格<br />(USDT)</span>
      <span>数量<br />({{ selectedProduct?.baseCoin || 'BTC' }})</span>
    </header>
    <div class="asks">
      <p v-for="price in ['76715.78', '76715.64', '76715.18', '76710.34', '76709.95', '76708.34']" :key="price">
        <span>{{ price }}</span><strong>0.1658497455</strong>
      </p>
    </div>
    <div class="mid-price">76707.92 ↑</div>
    <div class="bids">
      <p v-for="price in ['76704.54', '76704.53', '76704.46', '76701.73', '76700.53']" :key="price">
        <span>{{ price }}</span><strong>0.029772174</strong>
      </p>
    </div>
  </aside>
</template>

<style scoped>
.order-book-preview {
  min-width: 0;
}

.order-book-preview header,
.order-book-preview p {
  display: flex;
  justify-content: space-between;
  gap: 10px;
}

.order-book-preview header {
  margin-bottom: 14px;
  color: #8f929d;
  font-size: 14px;
}

.order-book-preview p {
  margin: 0 0 8px;
  font-size: 14px;
}

.order-book-preview strong {
  color: #fff;
  font-weight: 500;
}

.asks span {
  color: #ff574c;
}

.bids span,
.mid-price {
  color: #0cd977;
}

.mid-price {
  margin: 16px 0;
  font-size: 26px;
  font-weight: 600;
}
</style>
