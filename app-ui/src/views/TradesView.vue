<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'

import TradeOrderTabs from '@/components/trades/TradeOrderTabs.vue'
import { useItickStore } from '@/stores/itick'
import type { ItickTenantProduct } from '@/types/itick'
import { getTenantCode } from '@/api/http'

const store = useItickStore()

const selectedCategoryType = ref<number | null>(null)
const selectedProductKey = ref('')
const orderMode = ref<'market' | 'limit'>('market')
const productMenuOpen = ref(false)
const loading = ref(false)

const tenantCode = computed(() => getTenantCode())
const categories = computed(() => store.categories)
const products = computed(() => store.products)
const selectedCategory = computed(
  () => categories.value.find((item) => item.categoryType === selectedCategoryType.value) ?? null,
)
const selectedProduct = computed(
  () => products.value.find((item) => productKey(item) === selectedProductKey.value) ?? products.value[0] ?? null,
)
const tradeKind = computed(() => {
  const code = `${selectedCategory.value?.categoryCode || ''} ${selectedCategory.value?.categoryName || ''}`.toLowerCase()
  if (code.includes('stock') || code.includes('股票')) return 'stock'
  if (code.includes('option') || code.includes('期权')) return 'option'
  if (code.includes('forex') || code.includes('外汇')) return 'forex'
  if (code.includes('commodity') || code.includes('大宗')) return 'commodity'
  return 'crypto'
})
const placeholderPrice = computed(() => {
  if (tradeKind.value === 'stock') return '--'
  if (tradeKind.value === 'option') return '4781.55'
  if (tradeKind.value === 'forex') return '1.1765'
  if (tradeKind.value === 'commodity') return '4779.59'
  return '76704.54'
})
const placeholderChange = computed(() => {
  if (tradeKind.value === 'stock') return ''
  if (tradeKind.value === 'option') return '-0.8%'
  if (tradeKind.value === 'forex') return '-0.16%'
  if (tradeKind.value === 'commodity') return '-0.84%'
  return '+1.33%'
})
const priceTrend = computed(() => (placeholderChange.value.startsWith('-') ? 'down' : 'up'))
const productSheetRows = computed(() =>
  products.value.map((product) => {
    const key = productKey(product)
    const isSelected = key === selectedProductKey.value

    return {
      key,
      product,
      price: isSelected ? placeholderPrice.value : '--',
      change: isSelected ? placeholderChange.value : '',
      direction: isSelected ? priceTrend.value : 'flat',
    }
  }),
)

watch(selectedCategoryType, async (categoryType) => {
  if (categoryType === null) return
  await loadProducts(categoryType)
})

watch(products, (list) => {
  if (!list.length) {
    selectedProductKey.value = ''
    return
  }

  const hasSelected = list.some((item) => productKey(item) === selectedProductKey.value)
  if (!hasSelected) {
    selectedProductKey.value = productKey(list[0])
  }
})

onMounted(async () => {
  loading.value = true
  try {
    const list = await store.listVisibleCategories({
      tenantCode: tenantCode.value,
      limit: 20,
    })
    selectedCategoryType.value = list[0]?.categoryType ?? null
  } finally {
    loading.value = false
  }
})

async function loadProducts(categoryType: number) {
  loading.value = true
  try {
    await store.listVisibleProducts({
      tenantCode: tenantCode.value,
      categoryType,
      categoryCode: selectedCategory.value?.categoryCode,
      limit: 200,
    })
  } finally {
    loading.value = false
  }
}

function productKey(product: Pick<ItickTenantProduct, 'market' | 'symbol'>) {
  return `${String(product.market || '').toUpperCase()}::${String(product.symbol || '').toUpperCase()}`
}

function selectProduct(product: ItickTenantProduct) {
  selectedProductKey.value = productKey(product)
  productMenuOpen.value = false
}

function closeProductSheet() {
  productMenuOpen.value = false
}

function coinGlyph(product: ItickTenantProduct) {
  const coin = product.baseCoin || product.symbol.slice(0, 3) || product.displayName
  return coin.slice(0, 1).toUpperCase()
}
</script>

<template>
  <section class="trade-page" :aria-busy="loading">
    <nav class="trade-categories" aria-label="交易分类">
      <button
        v-for="category in categories"
        :key="category.id"
        type="button"
        :class="{ active: category.categoryType === selectedCategoryType }"
        @click="selectedCategoryType = category.categoryType"
      >
        {{ category.categoryName }}
      </button>
    </nav>

    <header class="trade-symbol">
      <button type="button" class="trade-symbol__main" @click="productMenuOpen = true">
        <strong>{{ selectedProduct?.symbol || '选择产品' }}</strong>
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

      <div v-if="productMenuOpen" class="product-sheet-overlay" @click.self="closeProductSheet">
        <section class="product-sheet">
          <span class="product-sheet__handle" />

          <header class="product-sheet__header">
            <h3>{{ selectedCategory?.categoryName || '产品' }}</h3>
            <button type="button" aria-label="关闭" @click="closeProductSheet">×</button>
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
              @click="selectProduct(row.product)"
            >
              <span class="product-sheet-row__coin">{{ coinGlyph(row.product) }}</span>
              <span class="product-sheet-row__symbol">{{ row.product.symbol }}</span>
              <strong>{{ row.price }}</strong>
              <span class="product-sheet-row__change">
                <em>{{ row.change || '--' }}</em>
                <small>{{ row.change || '等待' }}</small>
              </span>
            </button>
          </div>

          <div class="product-sheet__footer">
            <span>共 {{ productSheetRows.length }} 个产品</span>
          </div>
        </section>
      </div>
    </header>

    <section v-if="tradeKind === 'stock'" class="stock-panel">
      <div class="stock-alert">
        <span>!</span>
        <strong>该品种休市中，期间暂停交易</strong>
      </div>

      <div class="inner-tabs">
        <button class="active" type="button">普通交易</button>
        <button type="button">融资融券</button>
        <button type="button">盘前</button>
      </div>

      <div class="mode-switch compact">
        <button type="button" :class="{ active: orderMode === 'market' }" @click="orderMode = 'market'">市价</button>
        <button type="button" :class="{ active: orderMode === 'limit' }" @click="orderMode = 'limit'">限价</button>
      </div>

      <div class="trade-input">数量</div>
      <div class="percent-bar"><i /></div>
      <div class="percent-labels"><span>0%</span><span>25%</span><span>50%</span><span>75%</span><span>100%</span></div>
      <button class="wide-action" type="button">登录/注册</button>
      <TradeOrderTabs show-premarket />
    </section>

    <section v-else-if="tradeKind === 'option'" class="option-panel">
      <div class="mini-chart">
        <svg viewBox="0 0 320 88" aria-label="走势">
          <path d="M0 50 C28 48 26 40 54 42 C88 47 82 28 119 25 C152 20 143 32 176 28 C218 26 199 64 231 63 C266 62 252 49 320 53" />
        </svg>
      </div>

      <h3>时间</h3>
      <div class="duration-grid">
        <button class="active" type="button"><strong>30S</strong><span>30%</span></button>
        <button type="button"><strong>60S</strong><span>40%</span></button>
        <button type="button"><strong>90S</strong><span>50%</span></button>
        <button type="button"><strong>120S</strong><span>60%</span></button>
        <button type="button"><strong>180S</strong><span>70%</span></button>
        <button type="button"><strong>360S</strong><span>80%</span></button>
      </div>

      <h3>投资额</h3>
      <div class="trade-input split"><span>>=100</span><strong>USD</strong></div>
      <div class="percent-bar"><i /></div>
      <div class="percent-labels"><span>0%</span><span>25%</span><span>50%</span><span>75%</span><span>100%</span></div>
      <div class="buyable"><span>可买</span><strong>0 USD</strong></div>
      <div class="dual-actions"><button type="button">登录</button><button type="button">注册</button></div>
      <TradeOrderTabs />
    </section>

    <section v-else class="contract-panel" :class="{ 'contract-panel--crypto': tradeKind === 'crypto' }">
      <div class="trade-form">
        <div class="mode-switch">
          <button type="button" :class="{ active: orderMode === 'market' }" @click="orderMode = 'market'">市价</button>
          <button type="button" :class="{ active: orderMode === 'limit' }" @click="orderMode = 'limit'">限价</button>
        </div>

        <div class="form-row">
          <button type="button">全仓⌄</button>
          <button type="button">1X⌄</button>
        </div>

        <div class="trade-input">
          数量({{ tradeKind === 'forex' ? selectedProduct?.symbol : selectedProduct?.baseCoin || selectedProduct?.symbol || 'BTC' }})
        </div>
        <div class="percent-bar"><i /></div>
        <div class="percent-labels"><span>0%</span><span>25%</span><span>50%</span><span>75%</span><span>100%</span></div>

        <div class="account-lines">
          <span>可用</span><strong>0 USD</strong>
          <span>换算</span><strong>{{ tradeKind === 'crypto' ? '1 手 = 1 BTC' : `1 手 = 1 ${selectedProduct?.symbol || ''}` }}</strong>
        </div>

        <label class="checkbox-line"><i />止盈/止损</label>
        <div class="account-lines">
          <span>可开多</span><strong>0 手</strong>
          <span>保证金</span><strong>0 USD</strong>
        </div>
        <button class="wide-action" type="button">登录</button>

        <label class="checkbox-line"><i />止盈/止损</label>
        <div class="account-lines">
          <span>可开空</span><strong>0 手</strong>
          <span>保证金</span><strong>0 USD</strong>
        </div>
        <button class="wide-action" type="button">注册</button>
      </div>

      <aside v-if="tradeKind === 'crypto'" class="order-book-preview">
        <header><span>价格<br />(USDT)</span><span>数量<br />({{ selectedProduct?.baseCoin || 'BTC' }})</span></header>
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

      <TradeOrderTabs />
    </section>
  </section>
</template>

<style scoped>
.trade-page {
  width: 100%;
  max-width: 100%;
  min-height: calc(100dvh - 72px);
  padding: 18px 22px 112px;
  overflow-x: hidden;
  background: #0b0c15;
  color: #f6f7fb;
}

.trade-categories {
  display: flex;
  flex-wrap: wrap;
  gap: 28px;
  margin-bottom: 28px;
}

.trade-categories button,
.trade-symbol button,
.mode-switch button,
.form-row button,
.wide-action,
.dual-actions button,
.inner-tabs button,
.duration-grid button,
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
  font-size: 19px;
  font-weight: 500;
  white-space: nowrap;
}

.trade-categories button.active {
  color: #fff;
  font-size: 22px;
  font-weight: 600;
}

.trade-symbol {
  position: relative;
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 4px 14px;
  margin-bottom: 22px;
}

.trade-symbol__main {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  justify-self: start;
  padding: 0;
}

.trade-symbol__main strong {
  font-size: 17px;
  font-weight: 500;
}

.trade-symbol__main span {
  width: 10px;
  height: 10px;
  transform: rotate(45deg) translateY(-2px);
  border-right: 2px solid currentColor;
  border-bottom: 2px solid currentColor;
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
  max-height: 68dvh;
  padding: 22px 22px 26px;
  overflow: hidden;
  border-radius: 28px 28px 0 0;
  background: #22232c;
  color: #f6f7fb;
  box-shadow: 0 -24px 70px rgba(0, 0, 0, 0.42);
  touch-action: pan-y;
  max-width: 100%;
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

.mode-switch {
  display: grid;
  grid-template-columns: 1fr 1fr;
  min-height: 58px;
  margin-bottom: 18px;
  overflow: hidden;
  border-radius: 999px;
  background: #242631;
}

.mode-switch.compact {
  max-width: 260px;
}

.mode-switch button {
  color: #8f929d;
  font-size: 17px;
}

.mode-switch button.active {
  border-radius: 999px;
  background: #02b904;
  color: #fff;
}

.contract-panel--crypto {
  display: grid;
  grid-template-columns: minmax(0, 1.1fr) minmax(180px, 0.9fr);
  gap: 18px;
}

.contract-panel {
  min-width: 0;
}

.trade-form,
.stock-panel,
.option-panel {
  min-width: 0;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 14px;
  margin-bottom: 18px;
}

.form-row button,
.trade-input {
  min-height: 58px;
  border-radius: 12px;
  background: #242631;
  color: #f6f7fb;
  text-align: left;
}

.form-row button,
.trade-input {
  padding: 0 18px;
}

.trade-input {
  display: flex;
  align-items: center;
  margin-bottom: 18px;
  color: #8f929d;
  font-size: 17px;
}

.trade-input.split {
  justify-content: space-between;
}

.trade-input.split strong {
  color: #fff;
  font-weight: 500;
}

.percent-bar {
  height: 18px;
  margin-bottom: 10px;
  border-radius: 999px;
  background: linear-gradient(90deg, #1c1f2a 0 24%, transparent 24% 25%, #1c1f2a 25% 49%, transparent 49% 50%, #1c1f2a 50% 74%, transparent 74% 75%, #1c1f2a 75%);
}

.percent-bar i {
  display: block;
  width: 18px;
  height: 18px;
  border-radius: 999px;
  background: #02b904;
}

.percent-labels {
  display: flex;
  justify-content: space-between;
  margin-bottom: 24px;
  color: #8f929d;
  font-size: 14px;
}

.account-lines {
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 8px 14px;
  margin-bottom: 18px;
  color: #8f929d;
  font-size: 14px;
}

.account-lines strong {
  color: #fff;
  font-weight: 500;
}

.checkbox-line {
  display: flex;
  align-items: center;
  gap: 10px;
  margin: 8px 0 16px;
  font-size: 17px;
}

.checkbox-line i {
  width: 20px;
  height: 20px;
  border: 1px solid #f6f7fb;
  border-radius: 4px;
}

.wide-action,
.dual-actions button {
  min-height: 54px;
  border-radius: 12px;
  background: #181b25;
  color: #fff;
  font-size: 17px;
}

.wide-action {
  width: 100%;
  margin-bottom: 24px;
}

.dual-actions {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 18px;
  margin-bottom: 90px;
}

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

.stock-alert {
  display: flex;
  align-items: center;
  gap: 14px;
  margin: 0 -22px 22px;
  padding: 16px 28px;
  background: #282a34;
}

.stock-alert span {
  display: grid;
  place-items: center;
  width: 28px;
  height: 28px;
  border-radius: 999px;
  background: #ffa51f;
  color: #0b0c15;
}

.stock-alert strong {
  font-size: 17px;
  font-weight: 500;
}

.inner-tabs {
  display: flex;
  gap: 28px;
  margin-bottom: 24px;
  border-bottom: 1px solid #242633;
}

.inner-tabs button {
  position: relative;
  padding: 0 0 14px;
  color: #8f929d;
  font-size: 18px;
  font-weight: 500;
}

.inner-tabs button.active {
  color: #fff;
}

.inner-tabs button.active::after {
  position: absolute;
  right: 6px;
  bottom: 0;
  left: 6px;
  height: 3px;
  border-radius: 999px;
  background: #02b904;
  content: '';
}

.mini-chart {
  height: 110px;
  margin: 10px 0 34px;
}

.mini-chart svg {
  width: 100%;
  height: 100%;
}

.mini-chart path {
  fill: rgba(2, 185, 4, 0.18);
  stroke: #08c964;
  stroke-width: 3;
}

.option-panel h3 {
  margin: 0 0 16px;
  color: #8f929d;
  font-size: 17px;
  font-weight: 500;
}

.duration-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 14px;
  margin-bottom: 34px;
}

.duration-grid button {
  display: grid;
  gap: 6px;
  min-height: 84px;
  place-items: center;
  border-radius: 22px;
  background: #242631;
}

.duration-grid button.active {
  background: #02b904;
}

.duration-grid strong {
  font-size: 17px;
  font-weight: 600;
}

.duration-grid span {
  color: #9ca0aa;
}

.buyable {
  display: flex;
  justify-content: space-between;
  margin: 26px 0;
  color: #8f929d;
}

.buyable strong {
  color: #fff;
  font-weight: 500;
}

.trade-order-tabs {
  grid-column: 1 / -1;
  margin: 6px -22px 0;
  border-top: 1px solid #242633;
}

@media (max-width: 959px) {
  .trade-page {
    min-height: 100%;
    padding: 16px 16px calc(112px + env(safe-area-inset-bottom));
  }

  .contract-panel--crypto {
    grid-template-columns: minmax(0, 1fr);
  }

  .trade-categories {
    gap: 14px 20px;
  }

  .trade-symbol {
    grid-template-columns: minmax(0, 1fr) auto;
  }

  .trade-symbol__main,
  .trade-symbol__sub,
  .trade-symbol__quote,
  .account-lines strong,
  .order-book-preview strong {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .trade-symbol__main strong,
  .account-lines strong,
  .order-book-preview strong {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .order-book-preview p {
    font-size: 12px;
  }
}
</style>
