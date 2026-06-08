<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink, RouterView, useRoute } from 'vue-router'

import AppIcon from '@/components/common/AppIcon.vue'
import { useI18n } from '@/i18n'

type AppNavIcon = 'nav-home' | 'nav-market' | 'nav-trade' | 'nav-assets' | 'nav-profile'

type AppNavItem = {
  key: string
  labelKey: string
  path: string
  icon: AppNavIcon
}

const appNavigation: AppNavItem[] = [
  { key: 'home', labelKey: 'nav.home', path: '/home', icon: 'nav-home' },
  { key: 'markets', labelKey: 'nav.markets', path: '/markets', icon: 'nav-market' },
  { key: 'trade', labelKey: 'nav.trade', path: '/trades', icon: 'nav-trade' },
  { key: 'wallet', labelKey: 'nav.wallet', path: '/assets', icon: 'nav-assets' },
  { key: 'profile', labelKey: 'nav.profile', path: '/profile', icon: 'nav-profile' },
]

const route = useRoute()
const { t } = useI18n()

const isHomeRoute = computed(() => route.name === 'home')
const tabbarOverride = computed(() => {
  const value = Array.isArray(route.query.tabbar) ? route.query.tabbar[0] : route.query.tabbar
  if (value === '1' || value === 'true' || value === 'show') return true
  if (value === '0' || value === 'false' || value === 'hide') return false
  return null
})
const showTabbar = computed(() => {
  if (tabbarOverride.value !== null) return tabbarOverride.value
  return !route.meta.hideTabbar
})
</script>

<template>
  <div
    class="app-shell"
    :class="{
      'app-shell--home': isHomeRoute,
      'app-shell--tabbar-visible': showTabbar,
    }"
  >
    <div class="app-shell__aurora app-shell__aurora--left" />
    <div class="app-shell__aurora app-shell__aurora--right" />

    <div class="app-main">
      <main class="page-content">
        <RouterView />
      </main>

      <nav v-if="showTabbar" class="app-tabbar">
        <RouterLink
          v-for="item in appNavigation"
          :key="item.key"
          :to="item.path"
          class="app-tabbar__item"
          :class="[
            `app-tabbar__item--${item.key}`,
            { 'app-tabbar__item--active': route.path === item.path },
          ]"
        >
          <span class="app-tabbar__icon">
            <AppIcon :name="item.icon" class="app-tabbar__icon-svg" />
          </span>
          <span>{{ t(item.labelKey) }}</span>
        </RouterLink>
      </nav>
    </div>
  </div>
</template>
