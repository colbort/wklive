<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'

import LoginPrompt from '@/components/common/LoginPrompt.vue'
import MarketChartView from '@/components/markets/ChartView.vue'
import MarketQuotesView from '@/components/markets/QuotesView.vue'
import MarketTopTabs from '@/components/markets/TopTabs.vue'
import type { MarketTopTab, MarketTopTabItem } from '@/components/markets/types'
import { getAccessToken } from '@/api/http'
import { useTradingDesk } from '@/composables/useTradingDesk'
import { t } from '@/i18n'
import type { ItickTenantProduct } from '@/types/itick'
import { marketCategoryLabel } from '@/utils/marketCategory'

const topTabs: MarketTopTabItem[] = [
  { key: 'watchlist', label: 'market.watchlist' },
  { key: 'markets', label: 'market.quotes' },
  { key: 'chart', label: 'market.chart' },
]

const activeTopTab = ref<MarketTopTab>('markets')
const marketsPageRef = ref<HTMLElement | null>(null)
const tabsCollapsed = ref(false)
const tabsCollapseProgress = ref(0)
let scrollContainer: HTMLElement | null = null
let collapseRaf = 0

const topTabsHeight = 88

const detailVisible = computed(() => activeTopTab.value === 'chart')
const isLoggedIn = computed(() => Boolean(getAccessToken()))

const {
  selectedCategoryType,
  selectedProductKey,
  selectedIntervalName,
  loadingBootstrap,
  loadingKline,
  depthSnapshot,
  tickSnapshot,
  klineSnapshot,
  viewingLatestKlinePage,
  wsState,
  wsError,
  categories,
  products,
  intervals,
  selectedCategory,
  selectedCategoryCode,
  selectedQuote,
  marketRows,
  loadPreviousKlinePage,
  selectCategory,
  selectProduct,
  selectInterval,
} = useTradingDesk({
  detailVisible,
  tickLimit: 12,
})

function openProductChart(product: ItickTenantProduct) {
  selectProduct(product)
  activeTopTab.value = 'chart'
}

function updateTabsCollapse() {
  collapseRaf = 0

  if (activeTopTab.value !== 'markets') {
    tabsCollapseProgress.value = 0
    tabsCollapsed.value = false
    return
  }

  const pageRect = marketsPageRef.value?.getBoundingClientRect()
  const scrollRect = scrollContainer?.getBoundingClientRect()
  const scrollTop = scrollContainer?.scrollTop || 0
  const pageOffset = pageRect && scrollRect ? pageRect.top - scrollRect.top : -scrollTop
  const pageScroll = Math.max(0, -pageOffset, scrollTop)
  const progress = Math.min(pageScroll / topTabsHeight, 1)

  tabsCollapseProgress.value = progress
  tabsCollapsed.value = progress >= 1
}

function requestTabsCollapseUpdate() {
  if (collapseRaf) return
  collapseRaf = window.requestAnimationFrame(updateTabsCollapse)
}

function bindScrollContainer() {
  scrollContainer?.removeEventListener('scroll', requestTabsCollapseUpdate)
  scrollContainer =
    (marketsPageRef.value?.closest('.page-content') as HTMLElement | null) ||
    document.querySelector<HTMLElement>('.page-content')
  scrollContainer?.addEventListener('scroll', requestTabsCollapseUpdate, {
    passive: true,
  })
  window.addEventListener('scroll', requestTabsCollapseUpdate, { passive: true })
  window.addEventListener('resize', requestTabsCollapseUpdate, { passive: true })
  updateTabsCollapse()
}

onMounted(bindScrollContainer)

onBeforeUnmount(() => {
  scrollContainer?.removeEventListener('scroll', requestTabsCollapseUpdate)
  window.removeEventListener('scroll', requestTabsCollapseUpdate)
  window.removeEventListener('resize', requestTabsCollapseUpdate)
  if (collapseRaf) window.cancelAnimationFrame(collapseRaf)
})

watch(activeTopTab, () => {
  updateTabsCollapse()
})
</script>

<template>
  <section
    ref="marketsPageRef"
    class="markets-page"
    :class="{
      'markets-page--tabs-collapsed': tabsCollapsed,
      'markets-page--chart': activeTopTab === 'chart',
    }"
    :aria-busy="loadingBootstrap"
  >
    <MarketTopTabs
      :tabs="topTabs"
      :active-tab="activeTopTab"
      :collapsed="tabsCollapsed"
      :collapse-progress="tabsCollapseProgress"
      @change="activeTopTab = $event"
    />

    <div v-if="activeTopTab === 'markets'" class="markets-page__content">
      <MarketQuotesView
        :categories="categories"
        :selected-category-type="selectedCategoryType"
        :selected-category-name="marketCategoryLabel(selectedCategory)"
        :selected-category-code="selectedCategoryCode"
        :ws-state="wsState"
        :ws-error="wsError"
        :loading="loadingBootstrap"
        :rows="marketRows"
        :selected-product-key="selectedProductKey"
        :category-pinned="tabsCollapsed"
        @select-category="selectCategory"
        @select-product="openProductChart"
      />
    </div>

    <div v-else-if="activeTopTab === 'watchlist'" class="markets-page__watchlist">
      <div v-if="isLoggedIn" class="watchlist-empty">
        {{ t('market.watchlistPreparing') }}
      </div>
      <LoginPrompt v-else :action-text="t('assets.viewData')" compact />
    </div>

    <div v-else class="markets-page__content markets-page__content--chart">
      <MarketChartView
        :products="products"
        :rows="marketRows"
        :category-name="marketCategoryLabel(selectedCategory)"
        :selected-product-key="selectedProductKey"
        :selected-quote="selectedQuote"
        :kline-snapshot="klineSnapshot"
        :viewing-latest-kline-page="viewingLatestKlinePage"
        :depth-snapshot="depthSnapshot"
        :tick-snapshot="tickSnapshot"
        :loading-kline="loadingKline"
        :intervals="intervals"
        :selected-interval-name="selectedIntervalName"
        @select-product="selectedProductKey = $event"
        @select-interval="selectInterval"
        @load-previous-page="loadPreviousKlinePage"
      />
    </div>
  </section>
</template>

<style scoped>
.markets-page {
  width: 100%;
  max-width: 100%;
  min-height: calc(100dvh - 72px);
  padding: 18px 22px 112px;
  background: var(--page-bg);
  color: var(--text);
}

.markets-page__content {
  min-height: calc(100dvh - 72px);
}

.markets-page__watchlist {
  display: grid;
  min-height: calc(100dvh - 160px);
  place-items: center;
  padding: 24px 18px;
}

.watchlist-empty {
  color: var(--muted);
  font-size: 0.75rem;
}

@media (min-width: 0) {
  .markets-page {
    min-height: 100%;
    padding: 0 0 calc(96px + env(safe-area-inset-bottom));
  }

  .markets-page__content {
    min-height: 100%;
  }

  .markets-page--chart {
    padding-bottom: 0;
  }

  :global(.page-content:has(.markets-page--chart)) {
    padding-bottom: 0;
  }
}
</style>
