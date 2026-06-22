<script setup lang="ts">
import type { ChatGroup } from "@/types/chat";

defineProps<{
  loading: boolean;
  groups: ChatGroup[];
}>();

const emit = defineEmits<{
  edit: [row: ChatGroup];
  remove: [row: ChatGroup];
}>();
</script>

<template>
  <el-table
    v-loading="loading"
    :data="groups"
    height="100%"
  >
    <el-table-column
      prop="groupCode"
      label="分组编码"
      width="180"
    />
    <el-table-column
      prop="groupName"
      label="分组名称"
      width="160"
    />
    <el-table-column
      prop="description"
      label="描述"
      min-width="220"
      show-overflow-tooltip
    />
    <el-table-column
      prop="sort"
      label="排序"
      width="90"
    />
    <el-table-column
      label="状态"
      width="110"
    >
      <template #default="{ row }">
        <el-tag :type="row.enabled === 1 ? 'success' : 'info'">
          {{ row.enabled === 1 ? "启用" : "禁用" }}
        </el-tag>
      </template>
    </el-table-column>
    <el-table-column
      prop="remark"
      label="备注"
      min-width="160"
      show-overflow-tooltip
    />
    <el-table-column
      label="操作"
      width="160"
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
        <el-button
          link
          type="danger"
          @click="emit('remove', row)"
        >
          删除
        </el-button>
      </template>
    </el-table-column>
  </el-table>
</template>
