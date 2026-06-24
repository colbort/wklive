<script setup lang="ts">
import { ref } from "vue";
import type { FormInstance } from "element-plus";
import type { ChatAgent, ChatGroup } from "@/types/chat";

interface StatusOption {
  key?: string;
  label: string;
  value: number;
  tagType?: "success" | "info" | "warning" | "danger" | "primary";
}

interface EnabledOption {
  label: string;
  value: number;
}

interface AgentForm {
  username: string;
  password: string;
  nickname: string;
  mobile: string;
  email: string;
  enabled: number;
  autoOnline: boolean;
  maxSessionCount: number;
  groupId?: number;
  welcomeMessage: string;
  remark: string;
}

const props = defineProps<{
  loading: boolean;
  agents: ChatAgent[];
  groups: ChatGroup[];
  statusOptions: StatusOption[];
  visible: boolean;
  title: string;
  dialogMode: "create" | "edit";
  agentForm: AgentForm;
  enabledOptions: EnabledOption[];
}>();

const emit = defineEmits<{
  edit: [row: ChatAgent];
  "status-change": [row: ChatAgent, status: number];
  search: [keyword?: string];
  create: [];
  submit: [];
  "update:visible": [visible: boolean];
}>();

const keyword = ref("");
const formRef = ref<FormInstance>();

function statusText(status: number) {
  return props.statusOptions.find((item) => item.value === status)?.label || "未知";
}

function statusTagType(status: number) {
  return props.statusOptions.find((item) => item.value === status)?.tagType || "info";
}

function groupName(groupId: number) {
  return props.groups.find((item) => item.id === groupId)?.groupName || "-";
}

function onStatusChange(row: ChatAgent, status: string | number) {
  emit("status-change", row, Number(status));
}

const handleStatusChange = (row: ChatAgent) => {
  return (status: string | number) => onStatusChange(row, status);
};

async function submitDialog() {
  await formRef.value?.validate();
  emit("submit");
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
      @keyup.enter="emit('search', keyword)"
      @clear="emit('search')"
    />
    <el-button
      size="small"
      @click="emit('search', keyword)"
    >
      搜索
    </el-button>
    <el-button
      type="primary"
      size="small"
      @click="emit('create')"
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
    >
      <el-table-column
        prop="agentNo"
        label="坐席编号"
        width="130"
      />
      <el-table-column
        prop="chatUserId"
        label="用户 ID"
        width="100"
      />
      <el-table-column
        label="状态"
        width="120"
      >
        <template #default="{ row }">
          <el-tag :type="statusTagType(row.status)">
            {{ statusText(row.status) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column
        label="分组"
        width="140"
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
        width="100"
        fixed="right"
      >
        <template #default="{ row }">
          <el-button
            link
            type="primary"
            @click="emit('edit', row)"
          >
            编辑
          </el-button>
          <el-dropdown @command="handleStatusChange(row)">
            <el-button link>
              状态
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item
                  v-for="item in statusOptions"
                  :key="item.value"
                  :command="item.value"
                >
                  {{ item.label }}
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </template>
      </el-table-column>
    </el-table>
  </div>

  <el-dialog
    class="merchant-agent-edit-dialog"
    :model-value="visible"
    :title="title"
    width="560px"
    destroy-on-close
    @update:model-value="emit('update:visible', $event)"
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
      <el-button @click="emit('update:visible', false)">
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
