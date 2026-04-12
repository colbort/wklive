<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink, RouterView, useRoute } from 'vue-router'

import { appNavigation } from '@/config/navigation'
import { useDevice } from '@/composables/useDevice'

const route = useRoute()
const { isDesktop } = useDevice()

const pageTitle = computed(() => String(route.meta.title || 'AVE'))
</script>

<template>
  <div class="app-shell" :class="{ 'app-shell--desktop': isDesktop }">
    <div class="app-shell__aurora app-shell__aurora--left" />
    <div class="app-shell__aurora app-shell__aurora--right" />

    <header class="site-header">
      <RouterLink to="/home" class="site-brand">
        <span class="site-brand__mark">A</span>
        <div>
          <strong>AVE</strong>
          <p>{{ pageTitle }}</p>
        </div>
      </RouterLink>

      <nav v-if="isDesktop" class="site-nav">
        <RouterLink
          v-for="item in appNavigation"
          :key="item.key"
          :to="item.path"
          class="site-nav__item"
          :class="{ 'site-nav__item--active': route.path === item.path }"
        >
          <span>{{ item.label }}</span>
        </RouterLink>
      </nav>

      <div class="site-header__actions">
        <div v-if="isDesktop" class="topbar__badge">Multi-Asset Trading</div>
        <RouterLink to="/profile" class="header-cta">登录</RouterLink>
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
