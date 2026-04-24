<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { RouterLink, RouterView, useRoute, useRouter } from 'vue-router'

import { appNavigation } from '@/config/navigation'
import { apiListVisibleCategories, apiListVisibleProducts } from '@/api/itick'
import { getTenantCode } from '@/api/http'
import { useDevice } from '@/composables/useDevice'
import type { ItickTenantCategory, ItickTenantProduct } from '@/types/itick'

const route = useRoute()
const router = useRouter()
const { isDesktop } = useDevice()

const pageTitle = computed(() => String(route.meta.title || 'AVE'))
const showSiteHeader = computed(() => isDesktop.value || route.name === 'home')

const desktopCategories = ref<ItickTenantCategory[]>([])
const desktopProductsMap = ref<Record<number, ItickTenantProduct[]>>({})
const hoveredCategoryType = ref<number | null>(null)

const desktopDocNav = [
  { key: 'license', label: '公司资质', path: '/home' },
  { key: 'whitepaper', label: '白皮书', path: '/home' },
  { key: 'compliance', label: '监管文件', path: '/home' },
]

const activeDesktopCategoryType = computed(() => {
  if (hoveredCategoryType.value !== null) return hoveredCategoryType.value

  const queryCategoryType = Number(route.query.categoryType)
  if (Number.isFinite(queryCategoryType) && queryCategoryType > 0) {
    return queryCategoryType
  }

  return desktopCategories.value[0]?.categoryType ?? null
})

onMounted(async () => {
  if (isDesktop.value) {
    await ensureDesktopCategories()
  }
})

watch(isDesktop, async (desktop) => {
  if (!desktop) return
  await ensureDesktopCategories()
})

async function ensureDesktopCategories() {
  if (desktopCategories.value.length) return

  const tenantCode = getTenantCode()
  const res = await apiListVisibleCategories({
    tenantCode,
    limit: 20,
  })
  desktopCategories.value = res.data || []

  if (desktopCategories.value[0]) {
    await ensureDesktopProducts(desktopCategories.value[0])
  }
}

async function ensureDesktopProducts(category: ItickTenantCategory) {
  if (desktopProductsMap.value[category.categoryType]?.length) return

  const tenantCode = getTenantCode()
  const res = await apiListVisibleProducts({
    tenantCode,
    categoryType: category.categoryType,
    categoryCode: category.categoryCode,
    limit: 8,
  })

  desktopProductsMap.value = {
    ...desktopProductsMap.value,
    [category.categoryType]: res.data || [],
  }
}

async function handleDesktopCategoryEnter(category: ItickTenantCategory) {
  hoveredCategoryType.value = category.categoryType
  await ensureDesktopProducts(category)
}

function handleDesktopNavLeave() {
  hoveredCategoryType.value = null
}

function coinGlyph(product: ItickTenantProduct) {
  const coin = product.baseCoin || product.symbol.slice(0, 3) || product.displayName
  return coin.slice(0, 1).toUpperCase()
}

function openDesktopProduct(category: ItickTenantCategory, product: ItickTenantProduct) {
  router.push({
    path: '/trades',
    query: {
      categoryType: String(category.categoryType),
      categoryCode: category.categoryCode,
      market: product.market,
      symbol: product.symbol,
    },
  })
}
</script>

<template>
  <div class="app-shell" :class="{ 'app-shell--desktop': isDesktop }">
    <div class="app-shell__aurora app-shell__aurora--left" />
    <div class="app-shell__aurora app-shell__aurora--right" />

    <header v-if="showSiteHeader" class="site-header">
      <RouterLink to="/home" class="site-brand">
        <span class="site-brand__mark">A</span>
        <div>
          <strong>AVE</strong>
          <p v-if="!isDesktop">{{ pageTitle }}</p>
        </div>
      </RouterLink>

      <nav v-if="isDesktop" class="site-nav" @mouseleave="handleDesktopNavLeave">
        <div
          v-for="category in desktopCategories"
          :key="category.id"
          class="site-nav__entry"
          :class="{ 'site-nav__entry--active': category.categoryType === activeDesktopCategoryType }"
          @mouseenter="handleDesktopCategoryEnter(category)"
        >
          <RouterLink
            :to="{
              path: '/trades',
              query: {
                categoryType: String(category.categoryType),
                categoryCode: category.categoryCode,
              },
            }"
            class="site-nav__item"
          >
            <span>{{ category.categoryName }}</span>
          </RouterLink>

          <div v-if="desktopProductsMap[category.categoryType]?.length" class="site-nav__dropdown">
            <div class="site-nav__dropdown-inner">
              <button
                v-for="product in desktopProductsMap[category.categoryType]"
                :key="`${product.market}-${product.symbol}`"
                type="button"
                class="site-nav__market-row"
                @click="openDesktopProduct(category, product)"
              >
                <span class="site-nav__market-badge">{{ coinGlyph(product) }}</span>
                <strong>{{ product.symbol }}</strong>
                <em>{{ product.displayName || product.name || product.market }}</em>
                <small>{{ product.market }}</small>
              </button>
            </div>
          </div>
        </div>

        <div
          v-for="item in desktopDocNav"
          :key="item.key"
          class="site-nav__entry"
        >
          <RouterLink :to="item.path" class="site-nav__item">
            <span>{{ item.label }}</span>
          </RouterLink>
        </div>
      </nav>

      <div class="site-header__actions">
        <button class="site-action-circle">⌕</button>
        <button v-if="!isDesktop" class="site-action-circle">🌐</button>
        <div v-if="isDesktop" class="site-user-chip">
          <div class="site-user-chip__avatar" />
          <div>
            <strong>GUEST-8437</strong>
            <span>ID:50596163</span>
          </div>
        </div>
        <button v-if="isDesktop" class="site-action-plain">☰</button>
        <button class="site-action-circle site-action-circle--menu">☷</button>
      </div>
    </header>

    <div class="app-main">
      <main class="page-content">
        <RouterView />
      </main>

      <nav v-if="!isDesktop" class="mobile-tabbar">
        <RouterLink
          v-for="item in appNavigation"
          :key="item.key"
          :to="item.path"
          class="mobile-tabbar__item"
          :class="{ 'mobile-tabbar__item--active': route.path === item.path }"
        >
          <span class="mobile-tabbar__icon">{{ item.icon }}</span>
          <span>{{ item.label }}</span>
        </RouterLink>
      </nav>
    </div>
  </div>
</template>
