<script setup lang='ts'>
import { useI18n } from '@/i18n'
import type { MarketTopTab, MarketTopTabItem } from './types'

defineProps<{
  tabs: MarketTopTabItem[]
  activeTab: MarketTopTab
  collapsed?: boolean
  collapseProgress?: number
}>()

const emit = defineEmits<{
  change: [value: MarketTopTab]
}>()

const { t } = useI18n()
</script>

<template>
  <header
    class="market-header"
    :class="{ 'market-header--collapsed': collapsed || collapseProgress === 1 }"
    :style="`--market-top-collapse: ${collapseProgress || 0}`"
  >
    <nav class="top-tabs" :aria-label="t('market.viewLabel')">
      <button
        v-for="tab in tabs"
        :key="tab.key"
        type="button"
        class="top-tab app-menu__item"
        :class="{
          'top-tab--active': activeTab === tab.key,
          'app-menu__item--active': activeTab === tab.key,
        }"
        @click="emit('change', tab.key)"
      >
        {{ t(tab.label) }}
      </button>
    </nav>

    <button type="button" class="search-button" :aria-label="t('common.search')">
      <span />
    </button>
  </header>
</template>

<style scoped>
.market-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  box-sizing: border-box;
  max-height: calc(74px * (1 - var(--market-top-collapse, 0)));
  overflow: hidden;
  padding:
    calc(16px * (1 - var(--market-top-collapse, 0)))
    28px
    calc(8px * (1 - var(--market-top-collapse, 0)));
  min-height: calc(64px * (1 - var(--market-top-collapse, 0)));
  background: var(--page-bg);
  opacity: calc(1 - var(--market-top-collapse, 0));
  transform: translateY(calc(-10px * var(--market-top-collapse, 0)));
  transition: opacity 0.08s linear;
}

.market-header--collapsed {
  max-height: 0;
  min-height: 0;
  padding-top: 0;
  padding-bottom: 0;
  opacity: 0;
  pointer-events: none;
}

.top-tabs {
  display: flex;
  gap: var(--menu-gap);
}

.top-tab {
  border: 0;
  background: transparent;
  cursor: pointer;
}

.top-tab {
  position: relative;
  padding: .2rem 0 .45rem;
  line-height: 1.2;
}

.top-tab--active::after {
  position: absolute;
  right: .15rem;
  bottom: 0;
  left: .15rem;
  content: '';
}

.search-button {
  display: grid;
  flex: 0 0 auto;
  place-items: center;
  width: 46px;
  height: 46px;
  border: 0;
  border-radius: 999px;
  background: #242631;
  color: inherit;
  font: inherit;
  cursor: pointer;
}

.search-button span {
  width: 17px;
  height: 17px;
  border: 2px solid #fff;
  border-radius: 999px;
}

.search-button span::after {
  display: block;
  width: 8px;
  height: 2px;
  margin: 13px 0 0 13px;
  transform: rotate(45deg);
  border-radius: 999px;
  background: #fff;
  content: '';
}

@media (max-width: 390px) {
  .market-header {
    padding-right: 22px;
    padding-left: 22px;
  }

  .top-tabs {
    gap: var(--menu-gap);
  }

  .search-button {
    width: 44px;
    height: 44px;
  }
}
</style>
