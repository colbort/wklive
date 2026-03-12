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
    <el-pagination
      v-if="pagination"
      v-model:current-page="pagination.page"
      v-model:page-size="pagination.pageSize"
      :page-sizes="pageSizes"
      :total="pagination.total"
      layout="total, sizes, prev, pager, next, jumper"
      class="table-pagination"
    />
  </div>
</template>

<script setup lang="ts" generic="T extends Record<string, any>">
import { ref, computed } from 'vue'

interface Props<T> {
  data: T[]
  showActions?: boolean
  pagination?: {
    page: number
    pageSize: number
    total: number
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
