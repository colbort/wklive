import { createRouter, createWebHistory } from "vue-router";
import { useAuthStore } from "@/stores/auth";

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: "/login", component: () => import("@/views/LoginView.vue"), meta: { public: true } },
    {
      path: "/",
      component: () => import("@/layout/AppShell.vue"),
      children: [
        { path: "", redirect: "/dashboard" },
        { path: "dashboard", component: () => import("@/views/DashboardView.vue"), meta: { title: "运行总览" } },
        { path: "providers", component: () => import("@/views/ProvidersView.vue"), meta: { title: "流动性提供方" } },
        { path: "symbols", component: () => import("@/views/SymbolConfigsView.vue"), meta: { title: "交易对策略" } },
        { path: "orders", component: () => import("@/views/OrdersView.vue"), meta: { title: "订单与成交" } },
        { path: "hedges", component: () => import("@/views/ResourceView.vue"), props: { resource: "hedges" }, meta: { title: "对冲任务" } },
        { path: "risks", component: () => import("@/views/ResourceView.vue"), props: { resource: "risks" }, meta: { title: "风险事件" } },
        { path: "reconcile", component: () => import("@/views/ResourceView.vue"), props: { resource: "reconcile" }, meta: { title: "外部对账" } },
      ],
    },
    { path: "/:pathMatch(.*)*", redirect: "/" },
  ],
});

router.beforeEach((to) => {
  if (to.meta.public) return true;
  if (!useAuthStore().isLoggedIn) return { path: "/login", query: { redirect: to.fullPath } };
  return true;
});
