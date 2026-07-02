<script setup lang="ts">
import {
  computed,
  nextTick,
  onBeforeUnmount,
  onMounted,
  reactive,
  ref,
} from "vue";
import { ElMessage, ElMessageBox, type FormInstance } from "element-plus";
import {
  chatAdminWsUrl,
  createAgent,
  deleteAgent,
  options as loadOptions,
  pageAgents,
  pageGroups,
  updateAgent,
  type CreateChatAgentPayload,
  type UpdateChatAgentPayload,
} from "@/api/chat";
import { chatEventType, type ChatEventType } from "@/api/constant";
import { useAuthStore } from "@/stores/auth";
import type { ChatAgent, ChatGroup } from "@/types/chat";
import {
  optionGroup,
  withOptionLabels,
  type DisplayOptionItem,
} from "@/utils/options";

const auth = useAuthStore();
const merchantId = computed(() => auth.user?.merchantId || 0);

const enabledOptions = [
  { label: "启用", value: 1 },
  { label: "禁用", value: 2 },
];
const defaultAgentStatusOptions: DisplayOptionItem[] = [
  { code: "CHAT_AGENT_STATUS_OFFLINE", label: "离线", value: 1, tagType: "info" },
  { code: "CHAT_AGENT_STATUS_ONLINE", label: "在线", value: 2, tagType: "success" },
  { code: "CHAT_AGENT_STATUS_BUSY", label: "忙碌", value: 3, tagType: "warning" },
  { code: "CHAT_AGENT_STATUS_RESTING", label: "休息", value: 4, tagType: "info" },
];

const loading = ref(false);
const agents = ref<ChatAgent[]>([]);
const groups = ref<ChatGroup[]>([]);
const statusOptions = ref<DisplayOptionItem[]>(defaultAgentStatusOptions);
const keyword = ref("");
const formRef = ref<FormInstance>();
const dialogVisible = ref(false);
const dialogMode = ref<"create" | "edit">("create");
const editingId = ref(0);
let socket: WebSocket | null = null;
let reconnectTimer: number | null = null;
let reconnectTimes = 0;
let destroyed = false;

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

const dialogTitle = computed(
  () => `${dialogMode.value === "create" ? "新增" : "编辑"}坐席`,
);

onMounted(() => {
  destroyed = false;
  void loadAdminOptions();
  void loadCurrent();
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
    const agentStatuses = optionGroup(resp.data.options, "chatAgentStatus");
    if (agentStatuses.length) {
      statusOptions.value = withOptionLabels(agentStatuses);
    }
  } catch {
    statusOptions.value = defaultAgentStatusOptions;
  }
}

async function loadCurrent(searchKeyword = keyword.value) {
  if (!merchantId.value) return;
  loading.value = true;
  try {
    agents.value = (
      await pageAgents({
        merchantId: merchantId.value,
        limit: 100,
        keyword: searchKeyword || undefined,
      })
    ).data;
    groups.value = (
      await pageGroups({ merchantId: merchantId.value, limit: 100 })
    ).data;
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
      eventType?: ChatEventType;
      agent?: ChatAgent;
    };
    if (
      (event.eventType !== chatEventType.SYSTEM_NOTICE &&
        event.eventType !== chatEventType.AGENT_LEAVE) ||
      !event.agent
    )
      return;
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
  } else {
    void loadCurrent();
  }
  if (auth.agent?.id === agent.id) {
    auth.agent = {
      ...auth.agent,
      ...agent,
    };
  }
}

function resetForm() {
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
  nextTick(() => formRef.value?.clearValidate());
}

function openCreate() {
  dialogMode.value = "create";
  resetForm();
  dialogVisible.value = true;
}

function openEdit(row: ChatAgent) {
  dialogMode.value = "edit";
  resetForm();
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

async function submitDialog() {
  if (!merchantId.value) return;
  await formRef.value?.validate();

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

  dialogVisible.value = false;
  ElMessage.success("保存成功");
  await loadCurrent();
}

async function removeAgent(row: ChatAgent) {
  await ElMessageBox.confirm(`确认删除坐席「${row.agentNo}」？`, "删除确认", {
    type: "warning",
  });
  await deleteAgent(row.id, merchantId.value);
  ElMessage.success("删除成功");
  await loadCurrent();
}

function statusText(status: number) {
  return statusOptions.value.find((item) => item.value === status)?.label || "未知";
}

function statusTagType(status: number) {
  return statusOptions.value.find((item) => item.value === status)?.tagType || "info";
}

function groupName(groupId: number) {
  return groups.value.find((item) => item.id === groupId)?.groupName || "-";
}

</script>

<template>
  <div
    class="table-actions"
    style="width: 100%;"
  >
    <el-input
      v-model="keyword"
      clearable
      placeholder="搜索坐席"
      size="small"
      @keyup.enter="loadCurrent(keyword)"
      @clear="loadCurrent('')"
    />
    <el-button
      size="small"
      @click="loadCurrent(keyword)"
    >
      搜索
    </el-button>
    <el-button
      type="primary"
      size="small"
      @click="openCreate"
    >
      新增坐席
    </el-button>
  </div>
  <div
    class="table-panel merchant-agent-table-panel"
    style="width: 100%;"
  >
    <el-table
      v-loading="loading"
      :data="agents"
      height="100%"
      style="text-align: center;"
    >
      <el-table-column
        prop="agentNo"
        label="坐席编号"
        width="80"
      />
      <el-table-column
        label="状态"
        width="80"
      >
        <template #default="{ row }">
          <el-tag :type="statusTagType(row.status)">
            {{ statusText(row.status) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column
        label="分组"
        width="100"
      >
        <template #default="{ row }">
          {{ groupName(row.groupId) }}
        </template>
      </el-table-column>
      <el-table-column
        label="接待量"
        width="130"
      >
        <template #default="{ row }">
          {{ row.currentSessionCount }} / {{ row.maxSessionCount }}
        </template>
      </el-table-column>
      <el-table-column
        label="登录上线"
        width="110"
      >
        <template #default="{ row }">
          <el-tag :type="row.autoOnline === 1 ? 'success' : 'info'">
            {{ row.autoOnline === 1 ? "自动" : "手动" }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column
        prop="welcomeMessage"
        label="欢迎语"
        min-width="260"
        show-overflow-tooltip
      />
      <el-table-column
        prop="remark"
        label="备注"
        width="160"
        show-overflow-tooltip
      />
      <el-table-column
        label="操作"
        width="110"
        fixed="right"
      >
        <template #default="{ row }">
          <el-button
            link
            type="primary"
            @click="openEdit(row)"
          >
            编辑
          </el-button>
          <el-button
            link
            type="danger"
            @click="removeAgent(row)"
          >
            删除
          </el-button>
        </template>
      </el-table-column>
    </el-table>
  </div>

  <el-dialog
    v-model="dialogVisible"
    class="merchant-agent-edit-dialog"
    :title="dialogTitle"
    width="560px"
    destroy-on-close
  >
    <el-form
      ref="formRef"
      label-width="104px"
      :model="agentForm"
    >
      <el-form-item
        v-if="dialogMode === 'create'"
        label="登录账号"
        prop="username"
        :rules="[{ required: true, message: '请输入登录账号' }]"
      >
        <el-input v-model="agentForm.username" />
      </el-form-item>
      <el-form-item
        v-if="dialogMode === 'create'"
        label="登录密码"
        prop="password"
        :rules="[{ required: true, message: '请输入登录密码' }]"
      >
        <el-input
          v-model="agentForm.password"
          show-password
          type="password"
        />
      </el-form-item>
      <el-form-item
        v-if="dialogMode === 'create'"
        label="昵称"
        prop="nickname"
        :rules="[{ required: true, message: '请输入昵称' }]"
      >
        <el-input v-model="agentForm.nickname" />
      </el-form-item>
      <el-form-item
        v-if="dialogMode === 'create'"
        label="手机号"
      >
        <el-input v-model="agentForm.mobile" />
      </el-form-item>
      <el-form-item
        v-if="dialogMode === 'create'"
        label="邮箱"
      >
        <el-input v-model="agentForm.email" />
      </el-form-item>
      <el-form-item
        v-if="dialogMode === 'create'"
        label="账号状态"
      >
        <el-segmented
          v-model="agentForm.enabled"
          :options="enabledOptions"
        />
      </el-form-item>
      <el-form-item label="客服分组">
        <el-select
          v-model="agentForm.groupId"
          clearable
          class="full-input"
        >
          <el-option
            v-for="item in groups"
            :key="item.id"
            :label="item.groupName"
            :value="item.id"
          />
        </el-select>
      </el-form-item>
      <el-form-item
        label="最大接待"
        prop="maxSessionCount"
        :rules="[{ required: true, message: '请输入最大接待数' }]"
      >
        <el-input-number
          v-model="agentForm.maxSessionCount"
          :min="1"
          :max="999"
          class="full-input"
        />
      </el-form-item>
      <el-form-item label="登录上线">
        <el-switch
          v-model="agentForm.autoOnline"
          active-text="自动上线"
          inactive-text="手动上线"
        />
      </el-form-item>
      <el-form-item label="欢迎语">
        <el-input
          v-model="agentForm.welcomeMessage"
          type="textarea"
          :rows="3"
        />
      </el-form-item>
      <el-form-item label="备注">
        <el-input
          v-model="agentForm.remark"
          type="textarea"
          :rows="2"
        />
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="dialogVisible = false">
        取消
      </el-button>
      <el-button
        type="primary"
        @click="submitDialog"
      >
        保存
      </el-button>
    </template>
  </el-dialog>
</template>

<style scoped>
.table-panel {
  display: grid;
  flex: 1 1 auto;
  gap: 12px;
  min-height: 0;
  overflow: hidden;
  border: 1px solid #e6e9ef;
  border-radius: 8px;
  background: #fff;
}

.table-actions {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: nowrap;
}

.table-actions :deep(.el-input) {
  flex: 1;
  min-width: 0;
}

.table-actions :deep(.el-button) {
  flex: none;
}

@media (max-width: 760px) {
  .merchant-agent-table-panel {
    width: 100%;
    min-width: 0;
    overflow: hidden;
  }

  .merchant-agent-table-panel :deep(.el-table) {
    width: 100% !important;
    min-width: 0 !important;
  }

  .merchant-agent-table-panel :deep(.el-table__inner-wrapper),
  .merchant-agent-table-panel :deep(.el-table__body-wrapper),
  .merchant-agent-table-panel :deep(.el-scrollbar) {
    min-width: 0;
  }
}
</style>

<style>
@media (max-width: 760px) {
  .merchant-agent-edit-dialog {
    position: absolute;
    right: 0;
    bottom: 0;
    left: 0;
    display: flex;
    width: 100% !important;
    max-height: 88dvh;
    flex-direction: column;
    margin: 0 !important;
    border-radius: 16px 16px 0 0;
  }

  .merchant-agent-edit-dialog .el-dialog__header {
    flex: 0 0 auto;
    margin-right: 0;
    padding: 16px 18px 10px;
  }

  .merchant-agent-edit-dialog .el-dialog__body {
    flex: 1 1 auto;
    min-height: 0;
    overflow: auto;
    padding: 8px 18px 12px;
  }

  .merchant-agent-edit-dialog .el-dialog__footer {
    flex: 0 0 auto;
    padding: 12px 18px 16px;
    border-top: 1px solid #e6e9ef;
  }

  .merchant-agent-edit-dialog .el-form-item {
    margin-bottom: 16px;
  }

  .merchant-agent-edit-dialog .el-form-item__label {
    width: 88px !important;
    padding-right: 10px;
  }

  .merchant-agent-edit-dialog .el-form-item__content {
    min-width: 0;
  }

  .merchant-agent-edit-dialog .el-input-number,
  .merchant-agent-edit-dialog .el-select,
  .merchant-agent-edit-dialog .el-segmented {
    width: 100%;
  }

  .merchant-agent-edit-dialog .el-dialog__footer .el-button {
    min-width: 96px;
  }
}
</style>
