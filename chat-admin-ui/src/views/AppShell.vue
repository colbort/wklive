<script setup lang="ts">
import { computed, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useAuthStore } from "@/stores/auth";

const auth = useAuthStore();
const route = useRoute();
const router = useRouter();
const merchantDrawerVisible = ref(false);

const merchantMenu = [
  { path: "/merchant/agents", label: "坐席管理" },
  { path: "/merchant/groups", label: "客服分组" },
  { path: "/merchant/categories", label: "问题分类" },
];

const currentMenuLabel = computed(
  () =>
    merchantMenu.find((item) => item.path === route.path)?.label || "客服管理",
);
const merchantName = computed(
  () => auth.user?.nickname || auth.user?.username || "商户",
);
const merchantAvatarText = computed(() => merchantName.value.slice(0, 1));

async function logout() {
  await auth.logout();
  router.replace("/login");
}

function openMerchantMenu() {
  merchantDrawerVisible.value = true;
}

function goMerchant(path: string) {
  merchantDrawerVisible.value = false;
  router.push(path);
}
</script>

<template>
  <div
    v-if="auth.isMerchant"
    class="merchant-shell"
  >
    <aside class="merchant-sidebar">
      <div class="brand">
        <button
          class="brand-mark brand-button"
          type="button"
          @click="openMerchantMenu"
        >
          <span class="brand-initials">CS</span>
          <el-avatar
            class="brand-avatar"
            :size="38"
            :src="auth.user?.avatarUrl"
          >
            {{ merchantAvatarText }}
          </el-avatar>
        </button>
        <div>
          <div class="brand-title">
            <span class="desktop-brand-text">客服工作台</span>
            <span class="mobile-brand-text">{{ currentMenuLabel }}</span>
          </div>
          <div class="brand-subtitle">
            <span class="desktop-brand-text">商户后台</span>
            <span class="mobile-brand-text">{{ merchantName }}</span>
          </div>
        </div>
      </div>

      <nav class="nav">
        <button
          v-for="item in merchantMenu"
          :key="item.path"
          class="nav-item"
          :class="{ active: route.path === item.path }"
          type="button"
          @click="goMerchant(item.path)"
        >
          {{ item.label }}
        </button>
      </nav>

      <div class="sidebar-settings">
        <el-dropdown
          trigger="click"
          @command="(command: string) => command === 'logout' && logout()"
        >
          <button
            class="nav-item settings-trigger"
            type="button"
          >
            设置
          </button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="logout">
                退出登录
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </aside>

    <section class="merchant-main">
      <header class="merchant-topbar">
        <h1>{{ currentMenuLabel }}</h1>
        <div class="merchant-profile">
          <div class="profile-text">
            <strong>{{ auth.user?.nickname }}</strong>
          </div>
          <el-avatar
            :size="36"
            :src="auth.user?.avatarUrl"
          >
            {{ auth.user?.nickname?.slice(0, 1) || "商" }}
          </el-avatar>
        </div>
      </header>

      <main class="merchant-content">
        <RouterView />
      </main>
    </section>

    <el-drawer
      v-model="merchantDrawerVisible"
      class="merchant-menu-drawer"
      direction="ltr"
      size="260px"
      :with-header="false"
    >
      <div class="drawer-brand">
        <div class="brand-mark">
          CS
        </div>
        <div>
          <div class="brand-title">
            客服工作台
          </div>
          <div class="brand-subtitle">
            商户后台
          </div>
        </div>
      </div>
      <nav class="nav drawer-nav">
        <button
          v-for="item in merchantMenu"
          :key="item.path"
          class="nav-item"
          :class="{ active: route.path === item.path }"
          type="button"
          @click="goMerchant(item.path)"
        >
          {{ item.label }}
        </button>
      </nav>
    </el-drawer>
  </div>

  <div
    v-else
    class="agent-shell"
  >
    <header class="agent-topbar">
      <div class="brand">
        <div class="brand-mark">
          CS
        </div>
        <div>
          <div class="brand-title">
            客服工作台
          </div>
          <div class="brand-subtitle">
            坐席端
          </div>
        </div>
      </div>
      <div class="profile">
        <div class="profile-text">
          <strong>{{ auth.user?.nickname }}</strong>
          <span>{{ auth.agent?.agentNo }}</span>
        </div>
        <el-button @click="logout">
          退出
        </el-button>
      </div>
    </header>

    <main class="agent-content">
      <RouterView />
    </main>
  </div>
</template>
