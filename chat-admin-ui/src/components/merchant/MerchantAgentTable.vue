<script setup lang="ts">
import { ref } from "vue";
import type { ChatAgent, ChatGroup } from "@/types/chat";

interface StatusOption {
  key?: string;
  label: string;
  value: number;
  tagType?: "success" | "info" | "warning" | "danger" | "primary";
}

const props = defineProps<{
  loading: boolean;
  agents: ChatAgent[];
  groups: ChatGroup[];
  statusOptions: StatusOption[];
}>();

const emit = defineEmits<{
  edit: [row: ChatAgent];
  "status-change": [row: ChatAgent, status: number];
  search: [keyword?: string];
  create: [];
}>();

const keyword = ref("");

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
</script>

<template>
  <div
    class="table-actions"
    style="width: 100%; margin-bottom: 10px;"
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
  <div class="table-panel">
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
</template>

<style scoped>
.table-panel {
  display: grid;
  gap: 12px;
  min-height: 0;
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
</style>
