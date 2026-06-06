<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink, RouterView, useRoute } from 'vue-router'

import AppIcon from '@/components/common/AppIcon.vue'
import { appNavigation } from '@/config/navigation'
import { useI18n } from '@/i18n'

const route = useRoute()
const { t, toggleLocale } = useI18n()

const pageTitle = computed(() => String(route.meta.title || 'AVE'))
const isHomeRoute = computed(() => route.name === 'home')
const showSiteHeader = computed(() => isHomeRoute.value)
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

    <header v-if="showSiteHeader" class="site-header">
      <RouterLink to="/home" class="site-brand">
        <span class="site-brand__mark">A</span>
        <div>
          <strong>AVE</strong>
          <p v-if="!isHomeRoute">
            {{ pageTitle }}
          </p>
        </div>
      </RouterLink>

      <div class="site-header__actions">
        <button class="site-action-circle" :aria-label="t('common.search')">
          ⌕
        </button>
        <button
          class="site-action-circle"
          type="button"
          :aria-label="t('common.language')"
          @click="toggleLocale"
        >
          🌐
        </button>
      </div>
    </header>

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
