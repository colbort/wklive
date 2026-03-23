<template>
  <div class="table-wrapper">
    <el-table
      v-bind="$attrs"
      :data="data"
      stripe
      border
      class="table-content"
    >
      <slot />
      <el-table-column
        v-if="showActions"
        prop="actions"
        label="Actions"
        width="150"
        align="center"
        fixed="right"
      >
        <template #default="{ row }">
          <slot name="actions" :row="row" />
        </template>
      </el-table-column>
    </el-table>
    
    <!-- 分页器 -->
    <div v-if="pagination" style="display:flex; justify-content:flex-end; gap: 10px; align-items: center; margin-top: 12px;">
      <span>总共：{{ pagination.total }} 条</span>
      <el-button @click="$emit('prev')" :disabled="!pagination.hasPrev">上一页</el-button>
      <el-button @click="$emit('next')" :disabled="!pagination.hasNext">下一页</el-button>
      <el-select v-model="pagination.limit" style="width: 100px" @change="() => $emit('change-limit')">
        <el-option label="10" :value="10" />
        <el-option label="20" :value="20" />
        <el-option label="50" :value="50" />
      </el-select>
    </div>
  </div>
</template>

<script setup lang="ts" generic="T extends Record<string, any>">
import { ref, computed } from 'vue'

interface Props<T> {
  data: T[]
  showActions?: boolean
  pagination?: {
    cursor: string | null
    limit: number
    total: number
    hasNext: boolean
    hasPrev: boolean
  }
  pageSizes?: number[]
}

withDefaults(defineProps<Props<T>>(), {
  showActions: true,
  pageSizes: () => [10, 20, 50, 100],
})
</script>

<style scoped>
.table-wrapper {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.table-content {
  width: 100%;
}

.table-pagination {
  display: flex;
  justify-content: flex-end;
  padding: 12px 0;
}
</style>
