<script setup lang="ts">
import type { RouteLocationRaw } from 'vue-router'

export type CategoryPreviewItem = {
  path: RouteLocationRaw
  symbol: string
  price: string
  change: string
  coin: string
  coinClass: string
  icon?: string
}

defineProps<{
  items: CategoryPreviewItem[]
  loading: boolean
  error: boolean
}>()

defineEmits<{
  mouseenter: []
  mouseleave: []
}>()
</script>

<template>
  <div class="category-preview" @mouseenter="$emit('mouseenter')" @mouseleave="$emit('mouseleave')">
    <RouterLink
      v-for="item in items"
      :key="item.symbol"
      :to="item.path"
      class="category-preview__row"
    >
      <span class="category-preview__coin" :class="item.coinClass">
        <img v-if="item.icon" :src="item.icon" alt="">
        <span v-else>{{ item.symbol.slice(0, 1) }}</span>
      </span>
      <span class="category-preview__symbol">{{ item.symbol }}</span>
      <span class="category-preview__market">
        <span>{{ item.price }}</span>
        <span>{{ item.change }}</span>
      </span>
    </RouterLink>
    <div v-if="loading" class="category-preview__state">
      Loading...
    </div>
    <div v-else-if="error" class="category-preview__state">
      --
    </div>
    <div v-else-if="!items.length" class="category-preview__state">
      --
    </div>
  </div>
</template>

<style scoped>
.category-preview {
  position: absolute;
  z-index: 35;
  top: calc(100% + var(--px-10));
  left: var(--px-82);
  width: var(--px-520);
  max-height: calc(100vh - var(--px-100));
  padding: var(--px-28) var(--px-26);
  overflow: hidden auto;
  border: 1px solid rgb(255 255 255 / 10%);
  border-radius: var(--px-28);
  background: rgb(36 37 45);
  box-shadow: 0 var(--px-24) var(--px-80) rgb(0 0 0 / 36%);
}

.category-preview::before {
  position: absolute;
  top: calc(var(--px-10) * -1);
  left: 0;
  width: 100%;
  height: var(--px-12);
  content: '';
}

.category-preview__row {
  display: grid;
  grid-template-columns: var(--px-64) 1fr auto;
  align-items: center;
  min-height: var(--px-112);
  gap: var(--px-18);
  padding: 0 var(--px-10);
  border-bottom: 1px solid rgb(255 255 255 / 8%);
  color: var(--text);
  transition: background 0.18s ease;
}

.category-preview__row:last-child {
  border-bottom: 0;
}

.category-preview__row:hover {
  background: rgb(255 255 255 / 6%);
}

.category-preview__coin {
  display: grid;
  width: var(--px-54);
  height: var(--px-54);
  place-items: center;
  border-radius: 50%;
  color: var(--white);
  font-size: var(--font-size-30);
  font-weight: var(--font-weight-900);
  line-height: 1;
}

.category-preview__coin img {
  display: block;
  width: 100%;
  height: 100%;
  border-radius: 50%;
  object-fit: cover;
}

.coin--btc {
  background: rgb(249 143 21);
}

.coin--eth {
  background: rgb(139 102 245);
}

.coin--bch {
  background: rgb(132 191 109);
}

.coin--xrp {
  background: rgb(25 34 41);
}

.coin--ltc {
  background: rgb(222 142 22);
}

.coin--doge {
  background: rgb(207 180 25);
}

.category-preview__symbol {
  font-size: var(--font-size-24);
  font-weight: var(--font-weight-900);
}

.category-preview__market {
  display: grid;
  justify-items: end;
  gap: var(--px-8);
  color: rgb(255 76 66);
  font-size: var(--font-size-24);
  font-weight: var(--font-weight-900);
  line-height: 1;
}

.category-preview__market span:last-child {
  padding: var(--px-8) var(--px-10);
  border-radius: var(--px-8);
  background: rgb(255 76 66);
  color: var(--white);
  font-size: var(--font-size-18);
}

.category-preview__state {
  display: grid;
  min-height: var(--px-180);
  place-items: center;
  color: var(--text-muted);
  font-size: var(--font-size-18);
  font-weight: var(--font-weight-700);
}
</style>
