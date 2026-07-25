<script setup lang="ts">
import { computed } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useAuthStore } from "@/stores/auth";

const route = useRoute();
const router = useRouter();
const auth = useAuthStore();
const title = computed(() => String(route.meta.title || "流动性管理"));
const allMenus = [
  { path: "/dashboard", icon: "◫", label: "运行总览", perms: ["liquidity:dashboard:view"] },
  { path: "/providers", icon: "⬡", label: "流动性提供方", perms: ["liquidity:provider:list"] },
  { path: "/symbols", icon: "⌁", label: "交易对策略", perms: ["liquidity:strategy:list"] },
  {
    path: "/orders",
    icon: "⇄",
    label: "订单与成交",
    perms: ["liquidity:quote:list", "liquidity:external-order:list"],
  },
  { path: "/hedges", icon: "⟲", label: "对冲任务", perms: ["liquidity:hedge:list"] },
  { path: "/risks", icon: "△", label: "风险事件", perms: ["liquidity:risk:list"] },
  { path: "/reconcile", icon: "≋", label: "外部对账", perms: ["liquidity:reconcile:list"] },
];
const menus = computed(() =>
  allMenus.filter((menu) => menu.perms.some((perm) => auth.hasPerm(perm))),
);
function logout() { auth.logout(); router.replace("/login"); }
</script>

<template>
  <div class="shell">
    <aside>
      <div class="brand"><span class="brand-mark">LQ</span><div><b>LIQUIDITY</b><small>CONTROL CENTER</small></div></div>
      <div class="section-label">运营控制台</div>
      <nav>
        <RouterLink v-for="menu in menus" :key="menu.path" :to="menu.path">
          <span class="nav-icon">{{ menu.icon }}</span><span>{{ menu.label }}</span>
        </RouterLink>
      </nav>
      <div class="side-foot"><span class="pulse"></span>系统连接正常</div>
    </aside>
    <main>
      <header>
        <div><span class="crumb">做市管理 / </span><strong>{{ title }}</strong></div>
        <div class="operator"><span class="environment">PRODUCTION</span><span class="avatar">{{ auth.name.slice(0, 1) }}</span><span>{{ auth.name }}</span><button @click="logout">退出</button></div>
      </header>
      <section class="content"><RouterView /></section>
    </main>
  </div>
</template>

<style scoped>
.shell { display: flex; min-height: 100vh; background: radial-gradient(circle at 76% 0, #10243a 0, #07101d 34%); }
aside { position: fixed; inset: 0 auto 0 0; width: 244px; padding: 25px 18px; border-right: 1px solid rgba(119,146,179,.14); background: rgba(5,13,24,.92); }
.brand { display: flex; align-items: center; gap: 12px; padding: 0 8px 34px; }
.brand-mark { display: grid; place-items: center; width: 40px; height: 40px; border: 1px solid #25d3a2; border-radius: 11px; color: #25d3a2; font-weight: 800; box-shadow: inset 0 0 18px rgba(37,211,162,.12); }
.brand b { display: block; font-size: 14px; letter-spacing: 2.4px; }.brand small { display: block; margin-top: 3px; color: #52647c; font-size: 9px; letter-spacing: 1.7px; }
.section-label { padding: 0 13px 9px; color: #42536a; font-size: 10px; letter-spacing: 1.5px; }
nav { display: grid; gap: 4px; } nav a { display: flex; align-items: center; gap: 12px; padding: 12px 14px; border-radius: 9px; color: #7d8da3; text-decoration: none; font-size: 14px; transition: .2s; }
nav a:hover { color: #d9e5f4; background: rgba(35,57,81,.45); } nav a.router-link-active { color: #32dbac; background: linear-gradient(90deg, rgba(35,211,162,.14), rgba(35,211,162,.02)); box-shadow: inset 2px 0 #2dd4a4; }
.nav-icon { width: 22px; text-align: center; font-size: 17px; }
.side-foot { position: absolute; bottom: 25px; left: 31px; color: #566981; font-size: 11px; }.pulse { display: inline-block; width: 7px; height: 7px; margin-right: 8px; border-radius: 50%; background: #2dd4a4; box-shadow: 0 0 9px #2dd4a4; }
main { width: calc(100% - 244px); margin-left: 244px; } header { height: 68px; display: flex; align-items: center; justify-content: space-between; padding: 0 28px; border-bottom: 1px solid rgba(119,146,179,.13); color: #dfe8f3; background: rgba(7,16,29,.7); backdrop-filter: blur(14px); }
.crumb { color: #53657c; font-size: 13px; }.operator { display: flex; align-items: center; gap: 10px; color: #8494a9; font-size: 12px; }.environment { padding: 5px 8px; border: 1px solid rgba(45,212,164,.24); border-radius: 5px; color: #2dd4a4; font-size: 9px; letter-spacing: 1px; }.avatar { display: grid; place-items: center; width: 29px; height: 29px; border-radius: 8px; background: #1d344c; color: #64e2bf; }.operator button { border: 0; color: #65778e; background: none; cursor: pointer; }
.content { padding: 26px 28px 18px; }
</style>
