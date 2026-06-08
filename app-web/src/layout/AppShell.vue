<script setup lang="ts">
const navItems = [
  { path: '/home', label: '首页' },
  { path: '/markets', label: '市场' },
  { path: '/trades', label: '交易' },
  { path: '/assets', label: '资产' },
  { path: '/profile', label: '用户' },
]
</script>

<template>
  <div class="app-shell">
    <aside class="sidebar">
      <RouterLink to="/home" class="brand">
        <span class="brand__mark">A</span>
        <strong>AVE</strong>
      </RouterLink>

      <nav class="nav">
        <RouterLink v-for="item in navItems" :key="item.path" :to="item.path" class="nav__item">
          {{ item.label }}
        </RouterLink>
      </nav>
    </aside>

    <main class="content">
      <header class="topbar">
        <div>
          <p>PC Trading Console</p>
          <strong>综合交易工作台</strong>
        </div>
        <RouterLink to="/login" class="login-link">登录</RouterLink>
      </header>

      <RouterView />
    </main>
  </div>
</template>

<style scoped>
.app-shell {
  display: grid;
  grid-template-columns: 248px minmax(0, 1fr);
  min-height: 100vh;
}

.sidebar {
  position: sticky;
  top: 0;
  height: 100vh;
  padding: 28px 18px;
  border-right: 1px solid var(--border);
  background: rgb(10 13 22 / 92%);
  backdrop-filter: blur(20px);
}

.brand {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 0 12px 28px;
  color: var(--text);
  font-size: 30px;
  font-weight: 800;
}

.brand__mark {
  display: grid;
  width: 46px;
  height: 46px;
  place-items: center;
  border-radius: 14px;
  background: linear-gradient(135deg, rgb(63 232 202), rgb(119 83 226));
  color: rgb(8 12 20);
  font-weight: 900;
}

.nav {
  display: grid;
  gap: 8px;
}

.nav__item {
  padding: 13px 14px;
  border-radius: var(--radius-sm);
  color: var(--text-muted);
  font-weight: 600;
}

.nav__item.router-link-active {
  background: var(--accent-soft);
  color: var(--accent);
}

.content {
  min-width: 0;
  padding: 24px;
}

.topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 72px;
  margin-bottom: 20px;
  padding: 0 4px;
}

.topbar p {
  margin: 0 0 4px;
  color: var(--text-muted);
  font-size: 13px;
}

.topbar strong {
  font-size: 24px;
}

.login-link {
  padding: 10px 18px;
  border-radius: 999px;
  background: var(--accent);
  color: rgb(3 10 8);
  font-weight: 800;
}
</style>
