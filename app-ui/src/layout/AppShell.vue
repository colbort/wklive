<script setup lang='ts'>
import { computed, onMounted, ref, watch } from 'vue'
import { RouterLink, RouterView, useRoute, useRouter } from 'vue-router'

import { appNavigation } from '@/config/navigation'
import { apiGetMyAssetSummary } from '@/api/asset'
import { apiListVisibleCategories, apiListVisibleProducts } from '@/api/itick'
import { getAccessToken, getTenantCode } from '@/api/http'
import { apiGetProfile, apiLogout } from '@/api/userPrivate'
import { useDevice } from '@/composables/useDevice'
import type { AssetUserAsset } from '@/types/asset'
import type { ItickTenantCategory, ItickTenantProduct } from '@/types/itick'
import type { UserProfile } from '@/types/user'

const route = useRoute()
const router = useRouter()
const { isDesktop } = useDevice()

const pageTitle = computed(() => String(route.meta.title || 'AVE'))
const showSiteHeader = computed(() => isDesktop.value || route.name === 'home')
const showMobileTabbar = computed(() => !isDesktop.value && !route.meta.hideTabbar)

const desktopCategories = ref<ItickTenantCategory[]>([])
const desktopProductsMap = ref<Record<number, ItickTenantProduct[]>>({})
const hoveredCategoryType = ref<number | null>(null)
const profile = ref<UserProfile | null>(null)
const userAssets = ref<AssetUserAsset[]>([])
const activeUserPanel = ref('')

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

const userBase = computed(() => profile.value?.base ?? null)
const displayName = computed(() => userBase.value?.nickname || userBase.value?.username || 'GUEST-8437')
const displayUserId = computed(() => userBase.value?.id || 50596163)
const desktopAssetPreview = computed(() => userAssets.value.filter((asset) => asset.walletType === 1).slice(0, 8))

onMounted(async () => {
  if (isDesktop.value) {
    await ensureDesktopCategories()
    await ensureDesktopUserData()
  }
})

watch(isDesktop, async (desktop) => {
  if (!desktop) return
  await ensureDesktopCategories()
  await ensureDesktopUserData()
})

async function ensureDesktopUserData() {
  if (!getAccessToken()) return

  try {
    const [profileResp, assetResp] = await Promise.all([apiGetProfile(), apiGetMyAssetSummary({})])
    if (profileResp.code === 0 || profileResp.code === 200) {
      profile.value = profileResp.data
    }
    if (assetResp.code === 0 || assetResp.code === 200) {
      userAssets.value = assetResp.data?.assets || []
    }
  } catch (error) {
    console.warn('load desktop user data failed', error)
  }
}

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

async function logout() {
  await apiLogout()
  profile.value = null
  userAssets.value = []
  router.push('/profile')
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
        <div v-if="isDesktop" class="site-user-area">
          <div class="site-user-chip">
            <div class="site-user-chip__avatar" />
            <div>
              <strong>{{ displayName }}</strong>
              <span>ID:{{ displayUserId }}</span>
            </div>
            <i class="site-user-chip__bars" />
          </div>

          <div class="site-user-menu">
            <div class="site-user-menu__main">
              <header class="site-user-menu__head">
                <div class="site-user-chip__avatar" />
                <div>
                  <strong>{{ displayName }}</strong>
                  <span>ID:{{ displayUserId }}</span>
                </div>
              </header>

              <nav>
                <RouterLink class="site-user-menu__row" to="/assets" @mouseenter="activeUserPanel = 'assets'">
                  <span>◉</span>
                  <strong>我的资产</strong>
                  <em>›</em>
                </RouterLink>
                <RouterLink class="site-user-menu__row" to="/assets" @mouseenter="activeUserPanel = ''">
                  <span>▣</span>
                  <strong>充值</strong>
                </RouterLink>
                <RouterLink class="site-user-menu__row" to="/assets" @mouseenter="activeUserPanel = ''">
                  <span>▤</span>
                  <strong>提现</strong>
                </RouterLink>
                <RouterLink class="site-user-menu__row" to="/assets" @mouseenter="activeUserPanel = ''">
                  <span>⌘</span>
                  <strong>划转</strong>
                </RouterLink>
                <RouterLink class="site-user-menu__row" to="/assets" @mouseenter="activeUserPanel = ''">
                  <span>▧</span>
                  <strong>资金记录</strong>
                </RouterLink>
                <RouterLink class="site-user-menu__row site-user-menu__row--split" to="/profile" @mouseenter="activeUserPanel = ''">
                  <span>◈</span>
                  <strong>推荐朋友</strong>
                </RouterLink>
                <RouterLink class="site-user-menu__row site-user-menu__row--split" to="/assets" @mouseenter="activeUserPanel = ''">
                  <span>▤</span>
                  <strong>订单中心</strong>
                </RouterLink>
                <RouterLink class="site-user-menu__row site-user-menu__row--split" to="/profile" @mouseenter="activeUserPanel = ''">
                  <span>▰</span>
                  <strong>收款帐户</strong>
                </RouterLink>
                <RouterLink class="site-user-menu__row" to="/profile" @mouseenter="activeUserPanel = ''">
                  <span>◆</span>
                  <strong>安全设置</strong>
                </RouterLink>
                <button class="site-user-menu__row" type="button" @mouseenter="activeUserPanel = ''" @click="logout">
                  <span>↪</span>
                  <strong>退出登录</strong>
                </button>
              </nav>
            </div>

            <aside v-if="activeUserPanel === 'assets'" class="site-user-assets">
              <h3>现金账户</h3>
              <div class="site-user-assets__list">
                <div v-for="asset in desktopAssetPreview" :key="asset.id || asset.coin" class="site-user-assets__row">
                  <span>{{ asset.coin.slice(0, 1) }}</span>
                  <strong>{{ asset.coin }}</strong>
                  <em>{{ asset.availableAmount || asset.totalAmount || '0' }}</em>
                </div>
                <RouterLink class="site-user-assets__more" to="/assets">more+</RouterLink>
              </div>
            </aside>
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

      <nav v-if="showMobileTabbar" class="mobile-tabbar">
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
