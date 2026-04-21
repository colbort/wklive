<script setup lang="ts">
import type { MarketTopTab, MarketTopTabItem } from './types'

defineProps<{
  tabs: MarketTopTabItem[]
  activeTab: MarketTopTab
}>()

const emit = defineEmits<{
  change: [value: MarketTopTab]
}>()
</script>

<template>
  <header class="market-header">
    <nav class="top-tabs" aria-label="市场视图">
      <button
        v-for="tab in tabs"
        :key="tab.key"
        type="button"
        class="top-tab"
        :class="{ 'top-tab--active': activeTab === tab.key }"
        @click="emit('change', tab.key)"
      >
        {{ tab.label }}
      </button>
    </nav>

    <button type="button" class="search-button" aria-label="搜索">
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
  padding: 18px 18px 8px;
}

.top-tabs {
  display: flex;
  gap: 22px;
}

.top-tab,
.search-button {
  border: 0;
  background: transparent;
  color: inherit;
  font: inherit;
  cursor: pointer;
}

.top-tab {
  position: relative;
  padding: 5px 0 9px;
  color: #8f929d;
  font-size: 18px;
  font-weight: 500;
}

.top-tab--active {
  color: #ffffff;
  font-weight: 600;
}

.top-tab--active::after {
  position: absolute;
  right: 3px;
  bottom: 0;
  left: 3px;
  height: 3px;
  border-radius: 999px;
  background: #08c200;
  content: '';
}

.search-button {
  display: grid;
  flex: 0 0 auto;
  place-items: center;
  width: 42px;
  height: 42px;
  border-radius: 999px;
  background: #242631;
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
</style>
