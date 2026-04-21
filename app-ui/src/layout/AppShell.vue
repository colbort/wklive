<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink, RouterView, useRoute } from 'vue-router'

import { appNavigation } from '@/config/navigation'
import { useDevice } from '@/composables/useDevice'

const route = useRoute()
const { isDesktop } = useDevice()

const pageTitle = computed(() => String(route.meta.title || 'AVE'))
const showSiteHeader = computed(() => route.name === 'home')
const desktopNav = [
  '加密货币合约',
  '股票',
  '外汇',
  '大宗商品',
  '期权合约',
  '公司资质',
  '白皮书',
  '监管文件',
]
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

      <nav v-if="isDesktop" class="site-nav">
        <a
          v-for="item in desktopNav"
          :key="item"
          href="#"
          class="site-nav__item"
        >
          <span>{{ item }}</span>
        </a>
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
