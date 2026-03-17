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
const { pagination, updateTotal } = usePagination(10)
const { loading, withLoading } = useLoading()

// Query form
const { form: queryForm } = useForm({
  initialData: {
    username: '',
    method: '',
    path: '',
  }
})

const list_ref = ref<OpLogItem[]>([])

async function fetchList() {
  await withLoading(async () => {
    try {
      const res = await logService.getOperationLogs({
        username: queryForm.username || undefined,
        method: queryForm.method || undefined,
        path: queryForm.path || undefined,
        page: pagination.page,
        size: pagination.pageSize,
      })
      if (res.code !== 0 && res.code !== 200) throw new Error(res.msg)
      list_ref.value = res.data || []
      updateTotal(res.total || 0)
    } catch (e: any) {
      ElMessage.error(e?.message || t('common.loadFailed'))
    }
  })
}

function onSearch() {
  pagination.page = 1
  fetchList()
}

function onReset() {
  queryForm.username = ''
  queryForm.method = ''
  queryForm.path = ''
  pagination.page = 1
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
    <el-form :model="queryForm" inline style="margin-bottom: 16px;">
  <el-form-item :label="t('common.username')">
        <el-input v-model="queryForm.username" :placeholder="t('common.pleaseInputUsername')" clearable style="width: 200px" />
      </el-form-item>
      
      <el-form-item :label="t('common.method')">
        <el-input v-model="queryForm.method" :placeholder="t('common.pleaseInputMethod')" clearable style="width: 200px" />
      </el-form-item>
      
      <el-form-item :label="t('common.path')">
        <el-input v-model="queryForm.path" :placeholder="t('common.pleaseInputPath')" clearable style="width: 200px" />
      </el-form-item>
      
      <el-form-item>
        <el-button type="primary" @click="onSearch">{{ t('common.search') }}</el-button>
        <el-button @click="onReset">{{ t('common.reset') }}</el-button>
      </el-form-item>
    </el-form>

    <!-- Table -->
    <el-table :data="list_ref" v-loading="loading" row-key="id" style="margin-bottom: 16px;">
      <el-table-column prop="id" :label="t('common.id')" width="70" />
      <el-table-column prop="username" :label="t('common.username')" min-width="120" />
      <el-table-column prop="method" :label="t('common.method')" width="80">
        <template #default="{ row }">
          <el-tag>{{ row.method }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="path" :label="t('common.path')" min-width="150" show-overflow-tooltip />
      <el-table-column prop="ip" :label="t('common.ipAddress')" min-width="130" />
      <el-table-column prop="resp" :label="t('common.response')" min-width="150" show-overflow-tooltip />
      <el-table-column prop="costMs" :label="t('common.costMs')" width="110">
        <template #default="{ row }">
          <span style="color:#666;">{{ row.costMs }}ms</span>
        </template>
      </el-table-column>
      <el-table-column prop="createdAt" :label="t('common.createdAt')" min-width="170">
        <template #default="{ row }">
          <span style="color:#666;">{{ row.createdAt ? new Date(row.createdAt * 1000).toLocaleString() : '-' }}</span>
        </template>
      </el-table-column>
    </el-table>

    <!-- Pagination -->
    <div style="display:flex; justify-content:flex-end;">
      <el-pagination
        background
        layout="total, prev, pager, next, sizes"
        :total="pagination.total"
        :page-size="pagination.pageSize"
        :current-page="pagination.page"
        @update:current-page="(p:number)=>{pagination.page=p; fetchList()}"
        @update:page-size="(s:number)=>{pagination.pageSize=s; pagination.page=1; fetchList()}"
      />
    </div>
  </el-card>
</template>
