<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import type { RouteLocationRaw } from 'vue-router'
import { apiListVisibleCategories, getTenantCode } from '@wklive/api'
import { currentLanguageLabel, t, toggleLocale } from '@/i18n'

import webLogoDark from '../../assets/home/weblogo_dark.png'

type NavItem = {
  path: RouteLocationRaw
  label: string
}

const fixedNavItems = computed<NavItem[]>(() => [
  { path: '/company-credentials', label: t('nav.companyCredentials') },
  { path: '/whitepaper', label: t('nav.whitepaper') },
  { path: '/regulatory-files', label: t('nav.regulatoryFiles') },
])

const categoryNavItems = ref<NavItem[]>([])

const navItems = computed(() => [...categoryNavItems.value, ...fixedNavItems.value])

function buildCategoryPath(categoryCode: string) {
  return {
    path: '/markets',
    query: { categoryCode },
  }
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
    }))
  } catch (error) {
    console.warn('Failed to load visible categories', error)
  }
})
</script>

<template>
  <div class="app-shell">
    <header class="site-header">
      <RouterLink to="/home" class="brand" aria-label="AVE">
        <img :src="webLogoDark" alt="AVE">
      </RouterLink>

      <nav class="nav" aria-label="Navigation">
        <RouterLink v-for="item in navItems" :key="item.label" :to="item.path">
          {{ item.label }}
        </RouterLink>
      </nav>

      <div class="header-actions">
        <button class="icon-button" type="button" :aria-label="t('actions.search')">
          <span class="search-icon" />
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
        <button
          class="lang-toggle"
          type="button"
          :aria-label="t('actions.switchLanguage')"
          @click="toggleLocale"
        >
          {{ currentLanguageLabel }}
        </button>
        <span class="divider" />
        <button class="icon-button icon-button--menu" type="button" :aria-label="t('actions.menu')">
          <span />
          <span />
          <span />
        </button>
      </div>
    </header>

    <main class="content">
      <RouterView />
    </main>
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
  gap: var(--px-28);
  height: var(--px-88);
  padding: 0 var(--px-32);
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
  width: var(--px-148);
  height: auto;
}

.nav {
  display: flex;
  flex: 1 1 auto;
  min-width: 0;
  align-items: center;
  gap: var(--px-34);
  padding-right: var(--px-38);
  border-right: 1px solid rgb(255 255 255 / 7%);
  color: var(--text);
  font-size: var(--font-size-21);
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
    color 0.2s ease,
    opacity 0.2s ease;
}

.nav a:hover,
.nav a.router-link-active {
  color: var(--accent);
}

.header-actions {
  display: flex;
  flex: 0 0 auto;
  align-items: center;
  gap: var(--px-10);
}

.divider {
  width: 1px;
  height: var(--px-42);
  margin: 0 var(--px-8);
  background: rgb(255 255 255 / 7%);
}

.icon-button {
  display: grid;
  width: var(--px-54);
  height: var(--px-54);
  place-items: center;
  border: 0;
  border-radius: 50%;
  background: rgb(29 32 44);
  color: var(--text);
}

.search-icon {
  width: var(--px-21);
  height: var(--px-21);
  border: var(--px-2) solid currentColor;
  border-radius: 50%;
}

.search-icon::after {
  display: block;
  width: var(--px-9);
  height: var(--px-2);
  margin: calc(var(--px-21) - var(--px-5)) 0 0 calc(var(--px-21) - var(--px-6));
  transform: rotate(45deg);
  border-radius: var(--px-2);
  background: currentColor;
  content: '';
}

.pill {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: var(--px-96);
  height: var(--px-50);
  padding: 0 var(--px-22);
  border-radius: var(--px-999);
  font-size: var(--font-size-19);
  font-weight: var(--font-weight-800);
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
  min-width: var(--px-50);
  height: var(--px-50);
  padding: 0 var(--px-10);
  border: 1px solid rgb(255 255 255 / 14%);
  border-radius: var(--px-999);
  background: transparent;
  color: var(--text);
  font-size: var(--font-size-18);
  font-weight: var(--font-weight-800);
}

.icon-button--menu {
  gap: var(--px-5);
  border: 1px solid rgb(255 255 255 / 18%);
  background: transparent;
}

.icon-button--menu span {
  display: block;
  width: var(--px-25);
  height: var(--px-2);
  border-radius: var(--px-2);
  background: var(--text);
}

.content {
  min-width: 0;
}

@media (max-width: 1500px) {
  .site-header {
    gap: var(--px-24);
    padding: 0 var(--px-20);
  }

  .brand img {
    width: var(--px-132);
  }

  .nav {
    gap: var(--px-20);
    padding-right: var(--px-24);
    font-size: var(--font-size-18);
  }

  .header-actions {
    gap: var(--px-8);
  }

  .icon-button {
    width: var(--px-50);
    height: var(--px-50);
  }

  .pill {
    min-width: var(--px-88);
    padding: 0 var(--px-16);
    font-size: var(--font-size-18);
  }
}

@media (max-width: 1280px) {
  .site-header {
    gap: var(--px-18);
    padding: 0 var(--px-16);
  }

  .brand img {
    width: var(--px-120);
  }

  .nav {
    gap: var(--px-18);
  }

  .icon-button {
    width: var(--px-46);
    height: var(--px-46);
  }

  .pill {
    min-width: var(--px-78);
    height: var(--px-46);
    padding: 0 var(--px-14);
  }

  .lang-toggle {
    min-width: var(--px-46);
    height: var(--px-46);
  }
}
</style>
