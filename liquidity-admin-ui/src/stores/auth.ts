import { computed, ref } from "vue";
import { defineStore } from "pinia";
import { liquidityApi } from "@/api/liquidity";

export const useAuthStore = defineStore("auth", () => {
  const token = ref(localStorage.getItem("liquidity_admin_token") || "");
  const name = ref(localStorage.getItem("liquidity_admin_name") || "做市管理员");
  const isLoggedIn = computed(() => Boolean(token.value));
  async function login(username: string, password: string) {
    const result = await liquidityApi.login({ username, password });
    token.value = result.token;
    name.value = result.name || username;
    localStorage.setItem("liquidity_admin_token", token.value);
    localStorage.setItem("liquidity_admin_name", name.value);
  }
  function logout() {
    token.value = "";
    localStorage.removeItem("liquidity_admin_token");
    localStorage.removeItem("liquidity_admin_name");
  }
  return { token, name, isLoggedIn, login, logout };
});
