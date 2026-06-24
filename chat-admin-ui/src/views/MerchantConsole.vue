<script setup lang="ts">
import {
  computed,
  onBeforeUnmount,
  onMounted,
  reactive,
  ref,
  watch,
} from "vue";
import { useRoute, useRouter } from "vue-router";
import { ElMessage, ElMessageBox } from "element-plus";
import {
  chatAdminWsUrl,
  createAgent,
  createCategory,
  createGroup,
  deleteCategory,
  deleteGroup,
  options as loadOptions,
  pageAgents,
  pageCategories,
  pageGroups,
  updateAgent,
  updateAgentStatus,
  updateCategory,
  updateGroup,
  type CreateChatAgentPayload,
  type UpdateChatAgentPayload,
  type ChatCategoryPayload,
  type ChatGroupPayload,
} from "@/api/chat";
import MerchantAgentTable from "@/components/merchant/MerchantAgentTable.vue";
import MerchantCategoryTable from "@/components/merchant/MerchantCategoryTable.vue";
import MerchantGroupTable from "@/components/merchant/MerchantGroupTable.vue";
import { useAuthStore } from "@/stores/auth";
import type { ChatAgent, ChatCategory, ChatGroup } from "@/types/chat";
import { withOptionLabels, type DisplayOptionItem } from "@/utils/options";

type TabName = "agents" | "categories" | "groups";

const route = useRoute();
const router = useRouter();
const auth = useAuthStore();

const activeTab = computed<TabName>({
  get: () => String(route.meta.activeTab || "agents") as TabName,
  set: (value) => router.push(`/merchant/${value}`),
});

const merchantId = computed(() => auth.user?.merchantId || 0);
const loading = ref(false);
const agents = ref<ChatAgent[]>([]);
const categories = ref<ChatCategory[]>([]);
const groups = ref<ChatGroup[]>([]);
let socket: WebSocket | null = null;
let reconnectTimer: number | null = null;
let reconnectTimes = 0;
let destroyed = false;

const dialogVisible = ref(false);
const dialogMode = ref<"create" | "edit">("create");
const editingId = ref(0);

const agentForm = reactive({
  username: "",
  password: "",
  nickname: "",
  mobile: "",
  email: "",
  enabled: 1,
  autoOnline: false,
  maxSessionCount: 10,
  groupId: undefined as number | undefined,
  welcomeMessage: "",
  remark: "",
});

const categoryForm = reactive({
  parentId: 0,
  categoryCode: "",
  categoryName: "",
  enabled: 1,
  sort: 0,
  remark: "",
});

const groupForm = reactive({
  groupCode: "",
  groupName: "",
  description: "",
  enabled: 1,
  sort: 0,
  remark: "",
});

const dialogTitle = computed(() => {
  const action = dialogMode.value === "create" ? "新增" : "编辑";
  const resource =
    activeTab.value === "agents"
      ? "坐席"
      : activeTab.value === "categories"
        ? "问题分类"
        : "客服分组";
  return `${action}${resource}`;
});
const enabledOptions = [
  { label: "启用", value: 1 },
  { label: "禁用", value: 2 },
];

const defaultAgentStatusOptions: DisplayOptionItem[] = [
  { key: "chat.agent.status.offline", label: "离线", value: 1 },
  { key: "chat.agent.status.online", label: "在线", value: 2 },
  { key: "chat.agent.status.busy", label: "忙碌", value: 3 },
  { key: "chat.agent.status.resting", label: "休息", value: 4 },
];
const statusOptions = ref<DisplayOptionItem[]>(defaultAgentStatusOptions);

watch(
  () => activeTab.value,
  () => {
    loadCurrent();
  },
  { immediate: true },
);

void loadAdminOptions();

onMounted(() => {
  destroyed = false;
  connectWs();
});

onBeforeUnmount(() => {
  destroyed = true;
  clearReconnectTimer();
  if (socket) {
    socket.onclose = null;
    socket.close();
  }
});

async function loadAdminOptions() {
  try {
    const resp = await loadOptions();
    if (resp.data.agentStatuses?.length) {
      statusOptions.value = withOptionLabels(resp.data.agentStatuses);
    }
  } catch {
    statusOptions.value = defaultAgentStatusOptions;
  }
}

async function loadCurrent(keyword?: string) {
  if (!merchantId.value) return;
  loading.value = true;
  try {
    const params = {
      merchantId: merchantId.value,
      limit: 100,
      keyword: keyword || undefined,
    };
    if (activeTab.value === "agents") {
      agents.value = (await pageAgents(params)).data;
      if (!groups.value.length) {
        groups.value = (
          await pageGroups({ merchantId: merchantId.value, limit: 100 })
        ).data;
      }
    } else if (activeTab.value === "categories") {
      categories.value = (await pageCategories(params)).data;
    } else {
      groups.value = (await pageGroups(params)).data;
    }
  } finally {
    loading.value = false;
  }
}

function connectWs() {
  if (!auth.token || !merchantId.value) return;
  clearReconnectTimer();
  if (socket) {
    socket.onclose = null;
    socket.close();
  }
  socket = new WebSocket(
    chatAdminWsUrl({
      token: auth.token,
      merchantId: merchantId.value,
    }),
  );
  socket.onopen = () => {
    reconnectTimes = 0;
  };
  socket.onclose = () => {
    scheduleReconnect();
  };
  socket.onmessage = (event) => {
    handleWsMessage(event.data);
  };
}

function scheduleReconnect() {
  if (destroyed || !auth.token || !merchantId.value) return;
  clearReconnectTimer();
  const delays = [1000, 2000, 5000, 10000, 15000];
  const delay = delays[Math.min(reconnectTimes, delays.length - 1)];
  reconnectTimes += 1;
  reconnectTimer = window.setTimeout(() => {
    connectWs();
  }, delay);
}

function clearReconnectTimer() {
  if (!reconnectTimer) return;
  window.clearTimeout(reconnectTimer);
  reconnectTimer = null;
}

function handleWsMessage(payload: string) {
  try {
    const event = JSON.parse(payload) as {
      type?: string;
      agent?: ChatAgent;
    };
    if (event.type !== "chat.agent.status.updated" || !event.agent) return;
    upsertAgent(event.agent);
  } catch {
    // ignore invalid push payload
  }
}

function upsertAgent(agent: ChatAgent) {
  const index = agents.value.findIndex((item) => item.id === agent.id);
  if (index >= 0) {
    agents.value[index] = {
      ...agents.value[index],
      ...agent,
    };
  } else if (activeTab.value === "agents") {
    void loadCurrent();
  }
  if (auth.agent?.id === agent.id) {
    auth.agent = {
      ...auth.agent,
      ...agent,
    };
  }
}

function resetForms() {
  editingId.value = 0;
  Object.assign(agentForm, {
    username: "",
    password: "",
    nickname: "",
    mobile: "",
    email: "",
    enabled: 1,
    autoOnline: false,
    maxSessionCount: 10,
    groupId: undefined,
    welcomeMessage: "",
    remark: "",
  });
  Object.assign(categoryForm, {
    parentId: 0,
    categoryCode: "",
    categoryName: "",
    enabled: 1,
    sort: 0,
    remark: "",
  });
  Object.assign(groupForm, {
    groupCode: "",
    groupName: "",
    description: "",
    enabled: 1,
    sort: 0,
    remark: "",
  });
}

function openCreate() {
  dialogMode.value = "create";
  resetForms();
  dialogVisible.value = true;
}

function openAgentEdit(row: ChatAgent) {
  dialogMode.value = "edit";
  resetForms();
  editingId.value = row.id;
  Object.assign(agentForm, {
    username: "",
    password: "",
    nickname: "",
    mobile: "",
    email: "",
    enabled: 1,
    autoOnline: row.autoOnline === 1,
    maxSessionCount: row.maxSessionCount || 10,
    groupId: row.groupId || undefined,
    welcomeMessage: row.welcomeMessage,
    remark: row.remark,
  });
  dialogVisible.value = true;
}

function openCategoryEdit(row: ChatCategory) {
  dialogMode.value = "edit";
  resetForms();
  editingId.value = row.id;
  Object.assign(categoryForm, {
    parentId: row.parentId,
    categoryCode: row.categoryCode,
    categoryName: row.categoryName,
    enabled: row.enabled,
    sort: row.sort,
    remark: row.remark,
  });
  dialogVisible.value = true;
}

function openGroupEdit(row: ChatGroup) {
  dialogMode.value = "edit";
  resetForms();
  editingId.value = row.id;
  Object.assign(groupForm, {
    groupCode: row.groupCode,
    groupName: row.groupName,
    description: row.description,
    enabled: row.enabled,
    sort: row.sort,
    remark: row.remark,
  });
  dialogVisible.value = true;
}

async function submitDialog() {
  if (!merchantId.value) return;

  if (activeTab.value === "agents") {
    if (dialogMode.value === "create") {
      const payload: CreateChatAgentPayload = {
        maxSessionCount: agentForm.maxSessionCount,
        groupId: agentForm.groupId,
        welcomeMessage: agentForm.welcomeMessage,
        remark: agentForm.remark,
        username: agentForm.username,
        password: agentForm.password,
        nickname: agentForm.nickname,
        mobile: agentForm.mobile,
        email: agentForm.email,
        enabled: agentForm.enabled,
        autoOnline: agentForm.autoOnline ? 1 : 2,
      };
      await createAgent(payload);
    } else {
      const payload: UpdateChatAgentPayload = {
        merchantId: merchantId.value,
        maxSessionCount: agentForm.maxSessionCount,
        groupId: agentForm.groupId,
        welcomeMessage: agentForm.welcomeMessage,
        autoOnline: agentForm.autoOnline ? 1 : 2,
        remark: agentForm.remark,
      };
      await updateAgent(editingId.value, payload);
    }
  } else if (activeTab.value === "categories") {
    const payload: ChatCategoryPayload = {
      merchantId: merchantId.value,
      parentId: categoryForm.parentId,
      categoryCode: categoryForm.categoryCode || undefined,
      categoryName: categoryForm.categoryName,
      enabled: categoryForm.enabled,
      sort: categoryForm.sort,
      remark: categoryForm.remark,
    };
    if (dialogMode.value === "create") {
      await createCategory(payload);
    } else {
      await updateCategory(editingId.value, payload);
    }
  } else {
    const payload: ChatGroupPayload = {
      merchantId: merchantId.value,
      groupCode: groupForm.groupCode || undefined,
      groupName: groupForm.groupName,
      description: groupForm.description,
      enabled: groupForm.enabled,
      sort: groupForm.sort,
      remark: groupForm.remark,
    };
    if (dialogMode.value === "create") {
      await createGroup(payload);
    } else {
      await updateGroup(editingId.value, payload);
    }
  }

  dialogVisible.value = false;
  ElMessage.success("保存成功");
  await loadCurrent();
}

async function changeAgentStatus(row: ChatAgent, status: number) {
  await updateAgentStatus(row.id, {
    status,
  });
  ElMessage.success("状态已更新");
  await loadCurrent();
}

async function removeCategory(row: ChatCategory) {
  await ElMessageBox.confirm(
    `确认删除分类「${row.categoryName}」？`,
    "删除确认",
    {
      type: "warning",
    },
  );
  await deleteCategory(row.id, merchantId.value);
  ElMessage.success("删除成功");
  await loadCurrent();
}

async function removeGroup(row: ChatGroup) {
  await ElMessageBox.confirm(`确认删除分组「${row.groupName}」？`, "删除确认", {
    type: "warning",
  });
  await deleteGroup(row.id, merchantId.value);
  ElMessage.success("删除成功");
  await loadCurrent();
}
</script>

<template>
  <section class="console-page">
    <div class="table-wrap">
      <MerchantAgentTable
        v-if="activeTab === 'agents'"
        :loading="loading"
        :agents="agents"
        :groups="groups"
        :status-options="statusOptions"
        :visible="dialogVisible"
        :title="dialogTitle"
        :dialog-mode="dialogMode"
        :agent-form="agentForm"
        :enabled-options="enabledOptions"
        @edit="openAgentEdit"
        @status-change="changeAgentStatus"
        @search="loadCurrent"
        @create="openCreate"
        @update:visible="dialogVisible = $event"
        @submit="submitDialog"
      />

      <MerchantCategoryTable
        v-else-if="activeTab === 'categories'"
        :loading="loading"
        :categories="categories"
        :visible="dialogVisible"
        :title="dialogTitle"
        :dialog-mode="dialogMode"
        :category-form="categoryForm"
        :enabled-options="enabledOptions"
        @edit="openCategoryEdit"
        @remove="removeCategory"
        @search="loadCurrent"
        @create="openCreate"
        @update:visible="dialogVisible = $event"
        @submit="submitDialog"
      />

      <MerchantGroupTable
        v-else
        :loading="loading"
        :groups="groups"
        :visible="dialogVisible"
        :title="dialogTitle"
        :dialog-mode="dialogMode"
        :group-form="groupForm"
        :enabled-options="enabledOptions"
        @edit="openGroupEdit"
        @remove="removeGroup"
        @search="loadCurrent"
        @create="openCreate"
        @update:visible="dialogVisible = $event"
        @submit="submitDialog"
      />
    </div>
  </section>
</template>
