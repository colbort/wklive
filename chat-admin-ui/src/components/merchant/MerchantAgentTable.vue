<script setup lang="ts">
import type { ChatAgent, ChatGroup } from "@/types/chat";

interface StatusOption {
  label: string;
  value: number;
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
}>();

function statusText(status: number) {
  return props.statusOptions.find((item) => item.value === status)?.label || "未知";
}

function groupName(groupId: number) {
  return props.groups.find((item) => item.id === groupId)?.groupName || "-";
}
</script>

<template>
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
        <el-tag
          :type="
            row.status === 2
              ? 'success'
              : row.status === 3
                ? 'warning'
                : 'info'
          "
        >
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
      width="220"
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
        <el-dropdown @command="(status: string | number) => emit('status-change', row, Number(status))">
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
</template>
