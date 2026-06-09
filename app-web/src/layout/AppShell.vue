<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import type { RouteLocationRaw } from 'vue-router'
import { apiListVisibleCategories, getTenantCode } from '@wklive/api'
import { locale, setLocale, t } from '@/i18n'
import type { Locale } from '@/i18n'
import { useLanguageDialog } from '@/composables/useLanguageDialog'

import webLogoDark from '../../assets/home/weblogo_dark.png'
import searchIcon from '../../assets/home/search-white.svg'
import menuIcon from '../../assets/home/menu.svg'
import clientDownloadIcon from '../../assets/home/link_1.svg'
import supportIcon from '../../assets/home/link_2.svg'
import languageIcon from '../../assets/home/link_4.svg'

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
const { isLanguageDialogOpen, openLanguageDialog, closeLanguageDialog } = useLanguageDialog()

const navItems = computed(() => [...categoryNavItems.value, ...fixedNavItems.value])

const languageOptions: Array<{
  label: string
  flag: string
  locale?: Locale
}> = [
  { label: 'English', flag: '🇺🇸', locale: 'en-US' },
  { label: 'Español', flag: '🇪🇸' },
  { label: '日本語', flag: '🇯🇵' },
  { label: '한국어', flag: '🇰🇷' },
  { label: 'Русский', flag: '🇷🇺' },
  { label: 'Français', flag: '🇫🇷' },
  { label: 'Português', flag: '🇧🇷' },
  { label: 'Malaysia', flag: '🇲🇾' },
  { label: '中文繁體', flag: '🇨🇳' },
  { label: '中文简体', flag: '🇨🇳', locale: 'zh-CN' },
]

function buildCategoryPath(categoryCode: string) {
  return {
    path: '/markets',
    query: { categoryCode },
  }
}

function selectLanguage(nextLocale?: Locale) {
  if (nextLocale) {
    setLocale(nextLocale)
    closeLanguageDialog()
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
            <button class="menu-panel__item" type="button">
              <img :src="supportIcon" class="menu-panel__icon" alt="">
              {{ t('menu.support') }}
            </button>
            <button class="menu-panel__item" type="button" @click="openLanguageDialog">
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

    <div
      v-if="isLanguageDialogOpen"
      class="language-dialog"
      role="dialog"
      aria-modal="true"
      aria-labelledby="language-dialog-title"
      @click.self="closeLanguageDialog"
    >
      <section class="language-dialog__panel">
        <button
          class="language-dialog__close"
          type="button"
          aria-label="关闭"
          @click="closeLanguageDialog"
        />
        <h2 id="language-dialog-title">
          语言选择
        </h2>
        <label class="language-search">
          <input type="text" placeholder="请输入语言名称">
          <span />
        </label>
        <div class="language-list">
          <button
            v-for="item in languageOptions"
            :key="item.label"
            class="language-option"
            :class="{ 'language-option--active': item.locale === locale }"
            type="button"
            @click="selectLanguage(item.locale)"
          >
            <span class="language-option__flag">{{ item.flag }}</span>
            <span>{{ item.label }}</span>
            <i />
          </button>
        </div>
      </section>
    </div>
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

.language-dialog {
  position: fixed;
  z-index: 80;
  inset: 0;
  display: grid;
  place-items: center;
  padding: var(--px-32);
  background: rgb(0 0 0 / 42%);
}

.language-dialog__panel {
  position: relative;
  width: min(var(--px-420), calc(100vw - var(--px-64)));
  max-height: calc(100vh - var(--px-64));
  padding: var(--px-26) var(--px-26) var(--px-28);
  border-radius: var(--px-28);
  background: rgb(35 36 44);
  box-shadow: 0 var(--px-24) var(--px-80) rgb(0 0 0 / 45%);
  font-size: 16px;
  overflow-y: auto;
}

.language-dialog__panel h2 {
  margin: 0;
  text-align: center;
  font-size: 22px;
  font-weight: var(--font-weight-600);
  line-height: 1.25;
}

.language-dialog__close {
  position: absolute;
  top: var(--px-30);
  right: var(--px-30);
  width: var(--px-24);
  height: var(--px-24);
  border: 0;
  background: transparent;
}

.language-dialog__close::before,
.language-dialog__close::after {
  position: absolute;
  top: 50%;
  left: 50%;
  width: var(--px-22);
  height: var(--px-2);
  border-radius: var(--px-999);
  background: rgb(255 255 255 / 52%);
  content: '';
}

.language-dialog__close::before {
  transform: translate(-50%, -50%) rotate(45deg);
}

.language-dialog__close::after {
  transform: translate(-50%, -50%) rotate(-45deg);
}

.language-search {
  position: relative;
  display: block;
  margin-top: var(--px-46);
}

.language-search input {
  width: 100%;
  height: var(--px-52);
  padding: 0 var(--px-58) 0 var(--px-24);
  border: 1px solid rgb(255 255 255 / 10%);
  border-radius: var(--px-999);
  outline: 0;
  background: rgb(255 255 255 / 6%);
  color: var(--text);
  font-size: 18px;
  font-weight: var(--font-weight-600);
}

.language-search input::placeholder {
  color: rgb(255 255 255 / 34%);
}

.language-search span {
  position: absolute;
  top: 50%;
  right: var(--px-26);
  width: var(--px-20);
  height: var(--px-20);
  border: var(--px-2) solid var(--white);
  border-radius: 50%;
  transform: translateY(-50%);
}

.language-search span::after {
  position: absolute;
  right: calc(var(--px-6) * -1);
  bottom: calc(var(--px-6) * -1);
  width: var(--px-10);
  height: var(--px-2);
  border-radius: var(--px-999);
  background: var(--white);
  content: '';
  transform: rotate(45deg);
  transform-origin: center;
}

.language-list {
  display: grid;
  gap: var(--px-8);
  margin-top: var(--px-24);
}

.language-option {
  display: grid;
  min-height: var(--px-40);
  align-items: center;
  grid-template-columns: var(--px-34) 1fr var(--px-24);
  gap: var(--px-14);
  padding: 0 var(--px-18) 0 var(--px-16);
  border: 0;
  border-radius: var(--px-14);
  background: transparent;
  color: var(--text);
  text-align: left;
  font-size: 18px;
  font-weight: var(--font-weight-600);
  transition:
    background 0.18s ease,
    color 0.18s ease;
}

.language-option:hover {
  background: rgb(16 19 30 / 72%);
}

.language-option__flag {
  display: block;
  width: var(--px-32);
  height: var(--px-32);
  border-radius: 50%;
  font-size: 26px;
  line-height: var(--px-32);
  overflow: hidden;
}

.language-option i {
  display: block;
  width: var(--px-24);
  height: var(--px-24);
  border: var(--px-2) solid rgb(255 255 255 / 8%);
  border-radius: 50%;
}

.language-option--active i {
  border-color: var(--accent);
  box-shadow: inset 0 0 0 var(--px-5) rgb(35 36 44);
  background: var(--accent);
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

  .language-dialog__panel {
    width: min(var(--px-500), calc(100vw - var(--px-64)));
  }

  .language-search {
    margin-top: var(--px-42);
  }

  .language-search input {
    height: var(--px-56);
  }

  .language-list {
    gap: var(--px-14);
  }

  .language-option {
    min-height: var(--px-46);
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
