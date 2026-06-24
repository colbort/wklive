<script setup lang="ts">
import { ref } from "vue";
import type { ChatGroup } from "@/types/chat";

defineProps<{
  loading: boolean;
  groups: ChatGroup[];
}>();

const emit = defineEmits<{
  edit: [row: ChatGroup];
  remove: [row: ChatGroup];
  search: [keyword?: string];
  create: [];
}>();

const keyword = ref("");
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
  <div class="table-panel merchant-group-table-panel">
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
  </div>
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
  .merchant-group-table-panel {
    width: 100%;
    min-width: 0;
    overflow: hidden;
  }

  .merchant-group-table-panel :deep(.el-table) {
    width: 100% !important;
    min-width: 0 !important;
  }

  .merchant-group-table-panel :deep(.el-table__inner-wrapper),
  .merchant-group-table-panel :deep(.el-table__body-wrapper),
  .merchant-group-table-panel :deep(.el-scrollbar) {
    min-width: 0;
  }
}
</style>
