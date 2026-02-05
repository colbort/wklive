<script setup lang="ts">
import SideMenu from './side-menu.vue'
import TopBar from './top-bar.vue'
</script>

<template>
  <div class="layout">
    <aside class="sider">
      <div class="brand">Admin UI</div>
      <SideMenu />
    </aside>

    <main class="main">
      <TopBar />
      <div class="content">
        <router-view />
      </div>
    </main>
  </div>
</template>

<style scoped>
.layout { display: flex; height: 100vh; }
.sider { width: 240px; border-right: 1px solid #eee; display: flex; flex-direction: column; }
.brand { height: 56px; display:flex; align-items:center; padding: 0 16px; font-weight: 700; }
.main { flex: 1; display: flex; flex-direction: column; }
.content { padding: 16px; overflow: auto; flex: 1; background: #f7f8fa; }
</style>
