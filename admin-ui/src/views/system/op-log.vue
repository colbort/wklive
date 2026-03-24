<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { usePagination } from '@/composables/usePagination'
import { useLoading } from '@/composables/useLoading'
import { useForm } from '@/composables/useForm'
import { logService } from '@/services'
import type { OpLogItem } from '@/services/system/LogService'

const { t } = useI18n()

// Pagination and list
const {
  pagination,
  updatePagination,
  nextPage: paginationNextPage,
  prevPage: paginationPrevPage,
} = usePagination(10)
const { loading, withLoading } = useLoading()

// Query form
const { form: queryForm } = useForm({
  initialData: {
    username: '',
    method: '',
    path: '',
  },
})

const list_ref = ref<OpLogItem[]>([])

async function fetchList() {
  await withLoading(async () => {
    try {
      const res = await logService.getOperationLogs({
        username: queryForm.username || undefined,
        method: queryForm.method || undefined,
        path: queryForm.path || undefined,
        cursor: pagination.cursor,
        limit: pagination.limit,
      })
      if (res.code !== 0 && res.code !== 200) throw new Error(res.msg)
      list_ref.value = res.data || []
      updatePagination(
        res.total || 0,
        res.hasNext || false,
        res.hasPrev || false,
        res.nextCursor || null,
        res.prevCursor || null,
      )
    } catch (e: any) {
      ElMessage.error(e?.message || t('common.loadFailed'))
    }
  })
}

function onSearch() {
  pagination.cursor = null
  pagination.hasPrev = false
  fetchList()
}

function onReset() {
  queryForm.username = ''
  queryForm.method = ''
  queryForm.path = ''
  pagination.cursor = null
  pagination.hasPrev = false
  fetchList()
}

function nextPage() {
  paginationNextPage()
  fetchList()
}

function prevPage() {
  paginationPrevPage()
  fetchList()
}

onMounted(() => {
  fetchList()
})
</script>

<template>
  <el-card>
    <template #header>{{ t('system.opLog') }}</template>

    <!-- Query Form -->
    <el-form :model="queryForm" inline style="margin-bottom: 16px">
      <el-form-item :label="t('common.username')">
        <el-input
          v-model="queryForm.username"
          :placeholder="t('common.pleaseInputUsername')"
          clearable
          style="width: 200px"
        />
      </el-form-item>

      <el-form-item :label="t('common.method')">
        <el-input
          v-model="queryForm.method"
          :placeholder="t('common.pleaseInputMethod')"
          clearable
          style="width: 200px"
        />
      </el-form-item>

      <el-form-item :label="t('common.path')">
        <el-input
          v-model="queryForm.path"
          :placeholder="t('common.pleaseInputPath')"
          clearable
          style="width: 200px"
        />
      </el-form-item>

      <el-form-item>
        <el-button type="primary" @click="onSearch">{{ t('common.search') }}</el-button>
        <el-button @click="onReset">{{ t('common.reset') }}</el-button>
      </el-form-item>
    </el-form>

    <!-- Table -->
    <el-table :data="list_ref" v-loading="loading" row-key="id" style="margin-bottom: 16px">
      <el-table-column prop="id" :label="t('common.id')" width="70" />
      <el-table-column prop="username" :label="t('common.username')" min-width="120" />
      <el-table-column prop="method" :label="t('common.method')" width="80">
        <template #default="{ row }">
          <el-tag>{{ row.method }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column
        prop="path"
        :label="t('common.path')"
        min-width="150"
        show-overflow-tooltip
      />
      <el-table-column prop="ip" :label="t('common.ipAddress')" min-width="130" />
      <el-table-column
        prop="resp"
        :label="t('common.response')"
        min-width="150"
        show-overflow-tooltip
      />
      <el-table-column prop="costMs" :label="t('common.costMs')" width="110">
        <template #default="{ row }">
          <span style="color: #666">{{ row.costMs }}ms</span>
        </template>
      </el-table-column>
      <el-table-column prop="createdAt" :label="t('common.createdAt')" min-width="170">
        <template #default="{ row }">
          <span style="color: #666">{{
            row.createdAt ? new Date(row.createdAt * 1000).toLocaleString() : '-'
          }}</span>
        </template>
      </el-table-column>
    </el-table>

    <!-- Pagination -->
    <div style="display: flex; justify-content: flex-end; gap: 10px; align-items: center">
      <span>{{ t('common.totalItems', { count: pagination.total }) }}</span>
      <el-button @click="prevPage" :disabled="!pagination.hasPrev">{{
        t('common.prevPage')
      }}</el-button>
      <el-button @click="nextPage" :disabled="!pagination.hasNext">{{
        t('common.nextPage')
      }}</el-button>
      <el-select
        v-model="pagination.limit"
        style="width: 100px"
        @change="
          () => {
            pagination.cursor = null
            pagination.hasPrev = false
            fetchList()
          }
        "
      >
        <el-option label="10" :value="10" />
        <el-option label="20" :value="20" />
        <el-option label="50" :value="50" />
      </el-select>
    </div>
  </el-card>
</template>
