<script setup lang="ts">
const navItems = [
  { path: '/markets', label: '加密货币合约' },
  { path: '/markets', label: '股票' },
  { path: '/markets', label: '外汇' },
  { path: '/markets', label: '大宗商品' },
  { path: '/trades', label: '期权合约' },
  { path: '/profile', label: '公司资质' },
  { path: '/profile', label: '白皮书' },
  { path: '/profile', label: '监管文件' },
]
</script>

<template>
  <div class="app-shell">
    <header class="site-header">
      <RouterLink to="/home" class="brand" aria-label="AVE 首页">
        <span class="brand__mark"><span></span></span>
        <strong>AVE</strong>
      </RouterLink>

      <nav class="nav" aria-label="主导航">
        <RouterLink v-for="item in navItems" :key="`${item.path}-${item.label}`" :to="item.path">
          {{ item.label }}
        </RouterLink>
      </nav>

      <div class="header-actions">
        <button class="icon-button" type="button" aria-label="搜索">
          <span class="search-icon"></span>
        </button>
        <span class="divider"></span>
        <RouterLink to="/login" class="pill pill--ghost">模拟注册</RouterLink>
        <RouterLink to="/login" class="pill pill--outline">登录</RouterLink>
        <RouterLink to="/login" class="pill pill--solid">注册</RouterLink>
        <span class="divider"></span>
        <button class="icon-button icon-button--menu" type="button" aria-label="菜单">
          <span></span>
          <span></span>
          <span></span>
        </button>
      </div>
    </header>

    <main class="content">
      <RouterView />
    </main>
  </div>
</template>

<style scoped>
:global(html),
:global(body),
:global(#app) {
  width: 100%;
  min-height: 100%;
  margin: 0;
  border-radius: 0;
  background: var(--bg, #090c14);
}

:global(body) {
  overflow-x: auto;
}

.app-shell {
  width: 100%;
  min-width: 100vw;
  min-height: 100vh;
  border-radius: 0;
  background: var(--bg, #090c14);
  overflow: visible;
}

.site-header {
  position: sticky;
  z-index: 20;
  top: 0;
  display: flex;
  align-items: center;
  height: 88px;
  padding: 0 32px;
  border-bottom: 1px solid rgb(255 255 255 / 5%);
  border-radius: 0;
  background: rgb(10 12 22 / 96%);
  backdrop-filter: blur(18px);
}

.brand {
  display: flex;
  flex: 0 0 auto;
  align-items: center;
  gap: 12px;
  color: var(--text);
  font-size: 37px;
  font-weight: 900;
  line-height: 1;
}

.brand__mark {
  position: relative;
  display: grid;
  width: 54px;
  height: 54px;
  place-items: center;
  border-radius: 13px;
  background: linear-gradient(135deg, #2ed8c6 0%, #31bd86 48%, #8155e7 100%);
}

.brand__mark::before {
  position: absolute;
  inset: 8px 11px 9px;
  border-radius: 7px;
  background: rgb(10 12 22);
  clip-path: polygon(50% 0, 100% 84%, 50% 57%, 0 84%);
  content: "";
}

.brand__mark span {
  position: absolute;
  width: 20px;
  height: 20px;
  transform: translateY(8px) rotate(45deg);
  border-radius: 3px;
  background: linear-gradient(135deg, #27dec7, #8d62ed);
}

.nav {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 34px;
  margin-left: 56px;
  padding-right: 38px;
  border-right: 1px solid rgb(255 255 255 / 7%);
  color: var(--text);
  font-size: 21px;
  font-weight: 800;
  white-space: nowrap;
}

.nav a {
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
  gap: 14px;
  margin-left: auto;
}

.divider {
  width: 1px;
  height: 48px;
  margin: 0 12px;
  background: rgb(255 255 255 / 7%);
}

.icon-button {
  display: grid;
  width: 62px;
  height: 62px;
  place-items: center;
  border: 0;
  border-radius: 50%;
  background: rgb(29 32 44);
  color: var(--text);
}

.search-icon {
  width: 21px;
  height: 21px;
  border: 2px solid currentColor;
  border-radius: 50%;
}

.search-icon::after {
  display: block;
  width: 9px;
  height: 2px;
  margin: 16px 0 0 15px;
  transform: rotate(45deg);
  border-radius: 2px;
  background: currentColor;
  content: "";
}

.pill {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 112px;
  height: 54px;
  padding: 0 28px;
  border-radius: 999px;
  font-size: 21px;
  font-weight: 800;
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
  color: #fff;
}

.icon-button--menu {
  gap: 5px;
  border: 1px solid rgb(255 255 255 / 18%);
  background: transparent;
}

.icon-button--menu span {
  display: block;
  width: 25px;
  height: 2px;
  border-radius: 2px;
  background: var(--text);
}

.content {
  min-width: 0;
  border-radius: 0;
}

@media (max-width: 1500px) {
  .site-header {
    padding: 0 20px;
  }

  .nav {
    gap: 20px;
    margin-left: 34px;
    font-size: 18px;
  }

  .pill {
    min-width: 96px;
    padding: 0 20px;
    font-size: 18px;
  }
}
</style>
