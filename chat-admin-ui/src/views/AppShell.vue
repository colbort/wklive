<script setup lang="ts">
import { computed } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useAuthStore } from "@/stores/auth";

const auth = useAuthStore();
const route = useRoute();
const router = useRouter();

const merchantMenu = [
  { path: "/merchant/agents", label: "坐席管理" },
  { path: "/merchant/groups", label: "客服分组" },
  { path: "/merchant/categories", label: "问题分类" },
];

const agentMenu = [{ path: "/agent/workbench", label: "接待工作台" }];

const menu = computed(() => (auth.isMerchant ? merchantMenu : agentMenu));

async function logout() {
  await auth.logout();
  router.replace("/login");
}
</script>

<template>
  <div class="app-shell">
    <aside class="sidebar">
      <div class="brand">
        <div class="brand-mark">
          CS
        </div>
        <div>
          <div class="brand-title">
            客服工作台
          </div>
          <div class="brand-subtitle">
            {{ auth.isMerchant ? "商户后台" : "坐席端" }}
          </div>
        </div>
      </div>

      <nav class="nav">
        <button
          v-for="item in menu"
          :key="item.path"
          class="nav-item"
          :class="{ active: route.path === item.path }"
          type="button"
          @click="router.push(item.path)"
        >
          {{ item.label }}
        </button>
      </nav>
    </aside>

    <main class="main">
      <header class="topbar">
        <div>
          <h1>{{ route.meta.role === 2 ? "接待工作台" : "客服管理" }}</h1>
        </div>
        <div class="profile">
          <div class="profile-text">
            <strong>{{ auth.user?.nickname }}</strong>
            <span>{{
              auth.isMerchant ? "客服商户" : auth.agent?.agentNo
            }}</span>
          </div>
          <el-button @click="logout">
            退出
          </el-button>
        </div>
      </header>

      <RouterView />
    </main>
  </div>
</template>
