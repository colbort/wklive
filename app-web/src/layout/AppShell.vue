<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import type { RouteLocationRaw } from 'vue-router'
import { apiListVisibleCategories, apiListVisibleProducts, getTenantCode } from '@wklive/api'
import { t } from '@/i18n'
import { useLanguagePanel } from '@/composables/useLanguagePanel'
import { useSupportPanel } from '@/composables/useSupportPanel'
import ProductPreviewPanel from '@/components/ProductPreviewPanel.vue'
import type { CategoryPreviewItem } from '@/components/ProductPreviewPanel.vue'
import LanguagePanel from '@/components/LanguagePanel.vue'
import SupportPanel from '@/components/SupportPanel.vue'

import webLogoDark from '../../assets/home/weblogo_dark.png'
import searchIcon from '../../assets/home/search-white.svg'
import menuIcon from '../../assets/home/menu.svg'
import clientDownloadIcon from '../../assets/home/link_1.svg'
import supportIcon from '../../assets/home/link_2.svg'
import languageIcon from '../../assets/home/link_4.svg'
import { ItickTenantProduct } from '@wklive/api/types/itick'

type NavItem = {
  path: RouteLocationRaw
  label: string
}

type CategoryNavItem = NavItem & {
  categoryType: number
}

const fixedNavItems = computed<NavItem[]>(() => [
  { path: '/company-credentials', label: t('nav.companyCredentials') },
  { path: '/whitepaper', label: t('nav.whitepaper') },
  { path: '/regulatory-files', label: t('nav.regulatoryFiles') },
])

const categoryNavItems = ref<CategoryNavItem[]>([])
const isCategoryPreviewOpen = ref(false)
const activeCategoryCode = ref(0)
const previewLoading = ref(false)
const previewError = ref(false)
const previewRequestId = ref(0)
const marketPreviewCache = ref<Record<number, CategoryPreviewItem[]>>({})
const { openLanguagePanel } = useLanguagePanel()
const { openSupportPanel } = useSupportPanel()

const marketPreviewItems = computed(() => marketPreviewCache.value[activeCategoryCode.value] || [])

function buildCategoryPath(categoryCode: string) {
  return {
    path: '/markets',
    query: { categoryCode },
  }
}

function buildProductPath(product: ItickTenantProduct) {
  return {
    path: '/markets',
    query: {
      categoryCode: product.categoryCode,
      market: product.market,
      symbol: product.symbol,
    },
  }
}

function getCoinClass(index: number) {
  const coinClasses = ['coin--btc', 'coin--eth', 'coin--bch', 'coin--xrp', 'coin--ltc', 'coin--doge']
  return coinClasses[index % coinClasses.length]
}

async function loadCategoryPreview(categoryType: number) {
  if (marketPreviewCache.value[categoryType]) return

  const tenantCode = getTenantCode()
  if (!tenantCode) return

  const requestId = previewRequestId.value + 1
  previewRequestId.value = requestId
  previewLoading.value = true
  previewError.value = false

  try {
    const resp = await apiListVisibleProducts({
      tenantCode,
      categoryType: categoryType,
      cursor: 0,
      limit: 20,
    })

    if (requestId !== previewRequestId.value) return

    const products = (resp.data || [])
    marketPreviewCache.value = {
      ...marketPreviewCache.value,
      [categoryType]: products.map((product, index) => ({
        path: buildProductPath(product),
        symbol: product.symbol,
        price: '--',
        change: '--',
        coin: product.baseCoin,
        coinClass: getCoinClass(index),
        icon: product.icon,
      })),
    }
  } catch (error) {
    if (requestId !== previewRequestId.value) return
    previewError.value = true
    console.warn('Failed to load visible products', error)
  } finally {
    if (requestId === previewRequestId.value) {
      previewLoading.value = false
    }
  }
}

function openCategoryPreview(item?: CategoryNavItem) {
  console.warn('Open category preview', item?.categoryType)
  if (item) {
    activeCategoryCode.value = item.categoryType
    void loadCategoryPreview(item.categoryType)
  }
  isCategoryPreviewOpen.value = true
}

function closeCategoryPreview() {
  isCategoryPreviewOpen.value = false
}

onMounted(async () => {
  const tenantCode = getTenantCode()
  if (!tenantCode) return

  try {
    const resp = await apiListVisibleCategories({
      tenantCode,
      cursor: 0,
      limit: 20,
    })
    const categories = resp.data || []

    if (!categories.length) return

    categoryNavItems.value = categories.map((category) => ({
      path: buildCategoryPath(category.categoryCode),
      label: category.categoryName,
      categoryType: category.categoryType,
    }))
  } catch (error) {
    console.warn('Failed to load visible categories', error)
  }
})
</script>

<template>
  <div class="app-shell">
    <header class="site-header" @mouseleave="closeCategoryPreview">
      <RouterLink to="/home" class="brand" aria-label="AVE">
        <img :src="webLogoDark" alt="AVE">
      </RouterLink>

      <nav class="nav" aria-label="Navigation">
        <RouterLink
          v-for="item in categoryNavItems"
          :key="item.label"
          :to="item.path"
          class="nav__link nav__link--category"
          @mouseenter="openCategoryPreview(item)"
          @focus="openCategoryPreview(item)"
        >
          {{ item.label }}
        </RouterLink>
        <RouterLink
          v-for="item in fixedNavItems"
          :key="item.label"
          :to="item.path"
          class="nav__link"
          @mouseenter="closeCategoryPreview"
          @focus="closeCategoryPreview"
        >
          {{ item.label }}
        </RouterLink>
      </nav>

      <ProductPreviewPanel
        v-show="isCategoryPreviewOpen"
        :items="marketPreviewItems"
        :loading="previewLoading"
        :error="previewError"
        @mouseenter="openCategoryPreview()"
        @mouseleave="closeCategoryPreview"
      />

      <div class="header-actions" @mouseenter="closeCategoryPreview">
        <button class="icon-button" type="button" :aria-label="t('actions.search')">
          <img :src="searchIcon" alt="">
        </button>
        <span class="divider" />
        <RouterLink to="/login" class="pill pill--ghost">
          {{ t('actions.demoRegister') }}
        </RouterLink>
        <RouterLink to="/login" class="pill pill--outline">
          {{ t('actions.login') }}
        </RouterLink>
        <RouterLink to="/login" class="pill pill--solid">
          {{ t('actions.register') }}
        </RouterLink>
        <span class="divider" />
        <div class="menu-popover">
          <button
            class="icon-button icon-button--menu"
            type="button"
            :aria-label="t('actions.menu')"
          >
            <img :src="menuIcon" alt="">
          </button>
          <div class="menu-panel">
            <button class="menu-panel__item" type="button">
              <img :src="clientDownloadIcon" class="menu-panel__icon" alt="">
              {{ t('menu.download') }}
            </button>
            <button class="menu-panel__item" type="button" @click="openSupportPanel">
              <img :src="supportIcon" class="menu-panel__icon" alt="">
              {{ t('menu.support') }}
            </button>
            <button class="menu-panel__item" type="button" @click="openLanguagePanel">
              <img :src="languageIcon" class="menu-panel__icon" alt="">
              {{ t('menu.language') }}
            </button>
          </div>
        </div>
      </div>
    </header>

    <main class="content">
      <RouterView />
    </main>

    <LanguagePanel />
    <SupportPanel />
  </div>
</template>

<style scoped>
.app-shell {
  width: 100%;
  min-height: 100vh;
  background: var(--bg);
}

.site-header {
  position: sticky;
  z-index: 20;
  top: 0;
  display: flex;
  align-items: center;
  gap: var(--px-20);
  height: var(--px-76);
  padding: 0 var(--px-28);
  border-bottom: 1px solid rgb(255 255 255 / 5%);
  background: rgb(10 12 22 / 96%);
  backdrop-filter: blur(18px);
}

.brand {
  display: flex;
  flex: 0 0 auto;
  align-items: center;
}

.brand img {
  display: block;
  width: var(--px-100);
  height: auto;
}

.nav {
  display: flex;
  flex: 1 1 auto;
  min-width: 0;
  align-items: center;
  gap: var(--px-28);
  padding-left: var(--px-28);
  padding-right: var(--px-28);
  border-right: 1px solid rgb(255 255 255 / 7%);
  color: var(--text);
  font-size: var(--font-size-18);
  font-weight: var(--font-weight-800);
  overflow-x: auto;
  overscroll-behavior-inline: contain;
  scrollbar-width: none;
  white-space: nowrap;
}

.nav::-webkit-scrollbar {
  display: none;
}

.nav a {
  flex: 0 0 auto;
  transition:
    background 0.2s ease,
    color 0.2s ease,
    opacity 0.2s ease;
}

.nav a:hover,
.nav a.router-link-active {
  color: var(--accent);
}

.nav__link {
  display: inline-flex;
  align-items: center;
  height: 100%;
  padding: 0 var(--px-6);
}

.nav__link--category:hover,
.nav__link--category:focus-visible {
  background: rgb(255 255 255 / 8%);
}

.header-actions {
  display: flex;
  flex: 0 0 auto;
  align-items: center;
  gap: var(--px-6);
}

.divider {
  width: 1px;
  height: var(--px-32);
  margin: 0 var(--px-4);
  background: rgb(255 255 255 / 7%);
}

.icon-button {
  display: grid;
  width: var(--px-36);
  height: var(--px-36);
  place-items: center;
  border: 0;
  border-radius: 50%;
  background: rgb(29 32 44);
  color: var(--text);
}

.icon-button img {
  display: block;
  width: var(--px-16);
  height: var(--px-16);
  object-fit: contain;
}

.pill {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: var(--px-78);
  height: var(--px-40);
  padding: 0 var(--px-16);
  border-radius: var(--px-999);
  font-size: var(--font-size-15);
  font-weight: var(--font-weight-700);
}

.pill--ghost {
  background: rgb(29 32 44);
  color: var(--text);
}

.pill--outline {
  border: 1px solid var(--accent);
  color: var(--accent);
}

.pill--solid {
  background: var(--accent);
  color: var(--white);
}

.lang-toggle {
  min-width: var(--px-40);
  height: var(--px-40);
  padding: 0 var(--px-8);
  border: 1px solid rgb(255 255 255 / 14%);
  border-radius: var(--px-999);
  background: transparent;
  color: var(--text);
  font-size: var(--font-size-15);
  font-weight: var(--font-weight-700);
}

.icon-button--menu {
  border: 1px solid rgb(255 255 255 / 18%);
  background: transparent;
}

.icon-button--menu img {
  width: var(--px-18);
  height: var(--px-18);
}

.menu-popover {
  position: relative;
  display: flex;
}

.menu-popover::after {
  position: absolute;
  top: 100%;
  right: 0;
  width: var(--px-260);
  height: var(--px-18);
  content: '';
}

.menu-panel {
  position: absolute;
  z-index: 30;
  top: calc(100% + var(--px-14));
  right: 0;
  display: grid;
  width: var(--px-260);
  padding: var(--px-18) var(--px-28);
  border: 1px solid rgb(255 255 255 / 8%);
  border-radius: var(--px-28);
  background: rgb(11 13 24 / 98%);
  box-shadow: 0 var(--px-24) var(--px-80) rgb(0 0 0 / 35%);
  opacity: 0;
  pointer-events: none;
  transform: translateY(calc(var(--px-8) * -1));
  transition:
    opacity 0.18s ease,
    transform 0.18s ease;
}

.menu-popover:hover .menu-panel,
.menu-popover:focus-within .menu-panel {
  opacity: 1;
  pointer-events: auto;
  transform: translateY(0);
}

.menu-panel__item {
  display: flex;
  align-items: center;
  gap: var(--px-16);
  min-height: var(--px-70);
  padding: 0;
  border: 0;
  border-bottom: 1px solid rgb(255 255 255 / 8%);
  background: transparent;
  color: var(--text);
  font-size: var(--font-size-21);
  font-weight: var(--font-weight-800);
}

.menu-panel__item:last-child {
  border-bottom: 0;
}

.menu-panel__item:hover {
  color: var(--accent);
}

.menu-panel__icon {
  display: grid;
  width: var(--px-36);
  height: var(--px-36);
  flex: 0 0 auto;
  place-items: center;
  border: 1px solid rgb(255 255 255 / 12%);
  border-radius: 50%;
  color: var(--text);
  font-size: var(--font-size-18);
  line-height: 1;
}

.content {
  min-width: 0;
}

@media (max-width: 1500px) {
  .site-header {
    gap: var(--px-18);
    height: var(--px-70);
    padding: 0 var(--px-18);
  }

  .brand img {
    width: var(--px-120);
  }

  .nav {
    gap: var(--px-18);
    padding-left: var(--px-18);
    padding-right: var(--px-20);
    font-size: var(--font-size-18);
  }

  .header-actions {
    gap: var(--px-5);
  }

  .icon-button {
    width: var(--px-34);
    height: var(--px-34);
  }

  .pill {
    min-width: var(--px-70);
    height: var(--px-38);
    padding: 0 var(--px-12);
  }

  .lang-toggle {
    min-width: var(--px-38);
    height: var(--px-38);
  }
}

@media (max-width: 1280px) {
  .site-header {
    gap: var(--px-14);
    height: var(--px-64);
    padding: 0 var(--px-14);
  }

  .brand img {
    width: var(--px-110);
  }

  .nav {
    gap: var(--px-16);
    font-size: var(--font-size-15);
  }

  .icon-button {
    width: var(--px-32);
    height: var(--px-32);
  }

  .pill {
    min-width: var(--px-64);
    height: var(--px-36);
    padding: 0 var(--px-10);
  }

  .lang-toggle {
    min-width: var(--px-36);
    height: var(--px-36);
  }
}
</style>
