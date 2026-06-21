import { computed, ref } from "vue";
import { defineStore } from "pinia";
import { login, logout as logoutApi, profile, type LoginReq } from "@/api/chat";
import type { ChatAgent, ChatUser } from "@/types/chat";

const storageKey = "chat-admin-ui:auth";

interface AuthSnapshot {
  token: string;
  refreshToken: string;
  expireTime: number;
  user: ChatUser;
  agent?: ChatAgent;
}

function loadSnapshot(): AuthSnapshot | null {
  const raw = window.localStorage.getItem(storageKey);
  if (!raw) return null;
  try {
    return JSON.parse(raw) as AuthSnapshot;
  } catch {
    return null;
  }
}

export const useAuthStore = defineStore("auth", () => {
  const initial = loadSnapshot();
  const token = ref(initial?.token || "");
  const refreshToken = ref(initial?.refreshToken || "");
  const expireTime = ref(initial?.expireTime || 0);
  const user = ref<ChatUser | null>(initial?.user || null);
  const agent = ref<ChatAgent | null>(initial?.agent || null);
  const profileLoaded = ref(false);

  const isLoggedIn = computed(() => Boolean(token.value && user.value));
  const isMerchant = computed(() => user.value?.userType === 1);
  const isAgent = computed(() => user.value?.userType === 2);
  const homePath = computed(() =>
    isAgent.value ? "/agent/workbench" : "/merchant/agents",
  );

  function persist() {
    if (!token.value || !user.value) {
      window.localStorage.removeItem(storageKey);
      return;
    }
    window.localStorage.setItem(
      storageKey,
      JSON.stringify({
        token: token.value,
        refreshToken: refreshToken.value,
        expireTime: expireTime.value,
        user: user.value,
        agent: agent.value || undefined,
      }),
    );
  }

  function clear() {
    token.value = "";
    refreshToken.value = "";
    expireTime.value = 0;
    user.value = null;
    agent.value = null;
    profileLoaded.value = false;
    persist();
  }

  async function loginWithPassword(req: LoginReq) {
    const resp = await login(req);
    token.value = resp.data.token.accessToken;
    refreshToken.value = resp.data.token.refreshToken;
    expireTime.value = resp.data.token.expireTime;
    user.value = resp.data.user;
    agent.value = resp.data.agent || null;
    profileLoaded.value = true;
    persist();
  }

  async function fetchProfile() {
    if (!token.value) return;
    const resp = await profile();
    user.value = resp.user;
    agent.value = resp.agent || null;
    profileLoaded.value = true;
    persist();
  }

  async function logout() {
    try {
      if (token.value) {
        await logoutApi();
      }
    } finally {
      clear();
    }
  }

  return {
    token,
    refreshToken,
    expireTime,
    user,
    agent,
    profileLoaded,
    isLoggedIn,
    isMerchant,
    isAgent,
    homePath,
    loginWithPassword,
    fetchProfile,
    logout,
    clear,
  };
});
