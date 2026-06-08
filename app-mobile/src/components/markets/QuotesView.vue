<script setup lang="ts">
import { useI18n } from '@/i18n'
import type { ItickTenantCategory, ItickTenantProduct, ItickWsConnectionState } from '@/types/itick'
import { marketCategoryCodeLabel, marketCategoryLabel } from '@/utils/marketCategory'
import QuoteRow from './QuoteRow.vue'
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
      <QuoteRow
        v-for="row in rows"
        :key="row.key"
        :product="row.product"
        :quote="row.quote"
        :change-rate="row.changeRate"
        :direction="row.direction"
        :active="row.key === selectedProductKey"
        @select="emit('selectProduct', $event)"
      />
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

@media (max-width: 390px) {
  .category-strip {
    gap: 24px;
    padding-right: 22px;
    padding-left: 22px;
  }
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
