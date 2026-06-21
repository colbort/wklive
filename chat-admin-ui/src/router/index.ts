import { createRouter, createWebHistory } from "vue-router";
import { useAuthStore } from "@/stores/auth";

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: "/login",
      name: "Login",
      component: () => import("@/views/LoginView.vue"),
      meta: { public: true },
    },
    {
      path: "/",
      component: () => import("@/views/AppShell.vue"),
      children: [
        { path: "", redirect: "/merchant/agents" },
        {
          path: "merchant/agents",
          name: "MerchantAgents",
          component: () => import("@/views/MerchantConsole.vue"),
          meta: { role: 1, activeTab: "agents" },
        },
        {
          path: "merchant/categories",
          name: "MerchantCategories",
          component: () => import("@/views/MerchantConsole.vue"),
          meta: { role: 1, activeTab: "categories" },
        },
        {
          path: "merchant/groups",
          name: "MerchantGroups",
          component: () => import("@/views/MerchantConsole.vue"),
          meta: { role: 1, activeTab: "groups" },
        },
        {
          path: "agent/workbench",
          name: "AgentWorkbench",
          component: () => import("@/views/AgentWorkbench.vue"),
          meta: { role: 2 },
        },
      ],
    },
    { path: "/:pathMatch(.*)*", redirect: "/" },
  ],
});

router.beforeEach(async (to) => {
  const auth = useAuthStore();
  if (to.meta.public) return true;
  if (!auth.isLoggedIn)
    return { path: "/login", query: { redirect: to.fullPath } };

  if (!auth.profileLoaded) {
    try {
      await auth.fetchProfile();
    } catch {
      auth.clear();
      return { path: "/login", query: { redirect: to.fullPath } };
    }
  }

  if (to.path === "/") return auth.homePath;

  const requiredRole = to.meta.role;
  if (requiredRole === 1 && !auth.isMerchant)
    return { path: auth.homePath, replace: true };
  if (requiredRole === 2 && !auth.isAgent)
    return { path: auth.homePath, replace: true };
  return true;
});
