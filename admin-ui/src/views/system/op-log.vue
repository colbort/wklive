<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { usePagination } from '@/composables/usePagination'
import { useLoading } from '@/composables/useLoading'
import { useForm } from '@/composables/useForm'
// import { apiOpLogList } from '@/api/system/logs' // TODO: Implement API

const { t } = useI18n()

// Example structure - implement when API ready
type OpLog = {
  id: number
  username: string
  operation: string
  result: number // 0: failed, 1: success
  ip: string
  createdAt: number
}

// Pagination and list
const { pagination, updateTotal } = usePagination(10)
const { loading, withLoading } = useLoading()

// Query form
const { form: queryForm } = useForm({
  initialData: {
    username: '',
    operation: '',
    result: undefined as number | undefined,
  }
})

const list_ref = ref<OpLog[]>([])

async function fetchList() {
  await withLoading(async () => {
    try {
      // TODO: Implement API call when /admin/logs/op endpoint is ready
      // const res = await apiOpLogList({
      //   username: queryForm.username || undefined,
      //   operation: queryForm.operation || undefined,
      //   result: queryForm.result,
      //   page: pagination.page,
      //   size: pagination.pageSize,
      // })
      // if (res.code !== 0 && res.code !== 200) throw new Error(res.msg)
      // list_ref.value = res.data || []
      // updateTotal(res.total || 0)
      
      ElMessage.info('操作日志功能待实现 - Need to implement /admin/logs/op API')
    } catch (e: any) {
      ElMessage.error(e?.message || '加载失败')
    }
  })
}

function onSearch() {
  pagination.page = 1
  fetchList()
}

function onReset() {
  queryForm.username = ''
  queryForm.operation = ''
  queryForm.result = undefined
  pagination.page = 1
  fetchList()
}

onMounted(() => {
  // fetchList()
})
</script>

<template>
  <el-card>
    <template #header>{{ t('system.opLog') }}</template>
    
    <!-- Query Form -->
    <el-form :model="queryForm" inline style="margin-bottom: 16px;">
      <el-form-item label="用户名">
        <el-input v-model="queryForm.username" placeholder="请输入用户名" clearable style="width: 220px" />
      </el-form-item>
      
      <el-form-item label="操作">
        <el-input v-model="queryForm.operation" placeholder="请输入操作" clearable style="width: 220px" />
      </el-form-item>
      
      <el-form-item label="结果">
        <el-select v-model="queryForm.result" placeholder="请选择结果" clearable style="width: 140px">
          <el-option label="成功" :value="1" />
          <el-option label="失败" :value="0" />
        </el-select>
      </el-form-item>
      
      <el-form-item>
        <el-button type="primary" @click="onSearch">{{ t('common.search') }}</el-button>
        <el-button @click="onReset">{{ t('common.reset') }}</el-button>
      </el-form-item>
    </el-form>

    <!-- Table -->
    <el-table :data="list_ref" v-loading="loading" row-key="id" style="margin-bottom: 16px;">
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="username" label="用户名" min-width="120" />
      <el-table-column prop="operation" label="操作" min-width="150" />
      <el-table-column prop="result" label="结果" width="100">
        <template #default="{ row }">
          <el-tag :type="row.result === 1 ? 'success' : 'danger'">
            {{ row.result === 1 ? '成功' : '失败' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="ip" label="IP地址" min-width="130" />
      <el-table-column prop="createdAt" label="创建时间" min-width="170">
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

    <!-- Placeholder -->
    <div style="color:#666; text-align:center; padding:40px;">
      后续接 /admin/logs/op 接口 - 实现后取消注释上方 API 调用
    </div>
  </el-card>
</template>
