import { computed, ref } from "vue";
import { defineStore } from "pinia";
import {
  liquidityApi,
  type MenuNode,
  type ProfileUser,
} from "@/api/liquidity";

export const useAuthStore = defineStore("auth", () => {
  const token = ref(localStorage.getItem("liquidity_admin_token") || "");
  const exp = ref(Number(localStorage.getItem("liquidity_admin_exp") || 0));
  const user = ref<ProfileUser | null>(null);
  const menus = ref<MenuNode[]>([]);
  const perms = ref<string[]>([]);
  const roleIds = ref<number[]>([]);
  const isProfileLoaded = ref(false);

  const isLoggedIn = computed(() => Boolean(token.value));
  const name = computed(() => user.value?.nickname || user.value?.username || "做市管理员");
  const hasPerm = (perm: string) => perms.value.includes(perm);

  async function login(username: string, password: string, googleCode?: string) {
    const result = await liquidityApi.login({ username, password, googleCode });
    if (result.code !== 200 || !result.data?.token) {
      throw new Error(result.msg || "登录失败");
    }
    if (result.data.appScope !== 2) {
      throw new Error("当前账号不属于做市管理后台");
    }

    token.value = result.data.token;
    exp.value = result.data.exp;
    localStorage.setItem("liquidity_admin_token", token.value);
    localStorage.setItem("liquidity_admin_exp", String(exp.value));

    try {
      await fetchProfile();
    } catch (error) {
      logout();
      throw error;
    }
  }

  async function fetchProfile() {
    const result = await liquidityApi.profile();
    if (result.code !== 200 || !result.data?.user) {
      throw new Error(result.msg || "获取用户信息失败");
    }
    if (result.data.user.appScope !== 2) {
      throw new Error("当前账号不属于做市管理后台");
    }

    user.value = result.data.user;
    menus.value = result.data.menus || [];
    perms.value = result.data.perms || [];
    roleIds.value = result.data.roleIds || [];
    isProfileLoaded.value = true;
  }

  function logout() {
    token.value = "";
    exp.value = 0;
    user.value = null;
    menus.value = [];
    perms.value = [];
    roleIds.value = [];
    isProfileLoaded.value = false;
    localStorage.removeItem("liquidity_admin_token");
    localStorage.removeItem("liquidity_admin_exp");
    localStorage.removeItem("liquidity_admin_name");
  }

  return {
    token,
    exp,
    user,
    menus,
    perms,
    roleIds,
    isProfileLoaded,
    isLoggedIn,
    name,
    hasPerm,
    login,
    fetchProfile,
    logout,
  };
});
