<script setup lang="ts">
import { ref } from "vue";
import type { ChatCategory } from "@/types/chat";

defineProps<{
  loading: boolean;
  categories: ChatCategory[];
}>();

const emit = defineEmits<{
  edit: [row: ChatCategory];
  remove: [row: ChatCategory];
  search: [keyword?: string];
  create: [];
}>();

const keyword = ref("");
</script>

<template>
  <div class="table-panel">
    <div class="table-actions">
      <el-input
        v-model="keyword"
        clearable
        placeholder="搜索分类"
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
        新增分类
      </el-button>
    </div>
    <el-table
      v-loading="loading"
      :data="categories"
      height="100%"
    >
      <el-table-column
        prop="categoryCode"
        label="分类编码"
        width="180"
      />
      <el-table-column
        prop="categoryName"
        label="分类名称"
        min-width="180"
      />
      <el-table-column
        prop="parentId"
        label="父级 ID"
        width="100"
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
  gap: 12px;
  min-height: 0;
}

.table-actions {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}
</style>
