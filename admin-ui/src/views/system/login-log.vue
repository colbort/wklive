<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { usePagination } from '@/composables/usePagination'
import { useLoading } from '@/composables/useLoading'
import { useForm } from '@/composables/useForm'
import { apiLoginLogList } from '@/api/system/logs'
import type { LoginLogItem } from '@/types/system/logs'

const { t } = useI18n()

// Pagination and list
const { pagination, updateTotal } = usePagination(10)
const { loading, withLoading } = useLoading()

// Query form
const { form: queryForm } = useForm({
  initialData: {
    username: '',
    success: undefined as number | undefined,
  }
})

const list_ref = ref<LoginLogItem[]>([])

async function fetchList() {
  await withLoading(async () => {
    try {
      const res = await apiLoginLogList({
        username: queryForm.username || undefined,
        success: queryForm.success,
      // note: backend field named success        page: pagination.page,
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
  queryForm.success = undefined
  pagination.page = 1
  fetchList()
}

onMounted(() => {
  fetchList()
})
</script>

<template>
  <el-card>
    <template #header>{{ t('system.loginLog') }}</template>
    
    <!-- Query Form -->
    <el-form :model="queryForm" inline style="margin-bottom: 16px;">
      <el-form-item :label="t('common.username')">
        <el-input v-model="queryForm.username" :placeholder="t('common.pleaseInputUsername')" clearable style="width: 220px" />
      </el-form-item>
      
      <el-form-item :label="t('common.result')">
        <el-select v-model="queryForm.success" :placeholder="t('common.pleaseSelectResult')" clearable style="width: 140px">
          <el-option :label="t('common.success')" :value="1" />
          <el-option :label="t('common.failed')" :value="0" />
        </el-select>
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
      <el-table-column prop="ip" :label="t('common.loginIP')" min-width="130" />
      <el-table-column prop="ua" :label="t('common.userAgent')" min-width="180" show-overflow-tooltip />
      <el-table-column prop="success" :label="t('common.result')" width="100">
        <template #default="{ row }">
          <el-tag :type="row.success === 1 ? 'success' : 'danger'">
            {{ row.success === 1 ? t('common.loginSuccess') : t('common.loginFailed') }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="msg" :label="t('common.failureReason')" min-width="180" show-overflow-tooltip>
        <template #default="{ row }">
          <span v-if="row.success !== 1">{{ row.msg }}
          </span>
          <span v-else style="color:#999;">-</span>
        </template>
      </el-table-column>
      <el-table-column prop="loginAt" :label="t('common.loginTime')" min-width="170">
        <template #default="{ row }">
          <span style="color:#666;">{{ row.loginAt ? new Date(row.loginAt * 1000).toLocaleString() : '-' }}</span>
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
