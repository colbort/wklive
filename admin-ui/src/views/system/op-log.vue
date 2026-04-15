<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { usePagination } from '@/composables/usePagination'
import { useLoading } from '@/composables/useLoading'
import { useForm } from '@/composables/useForm'
import { formatDate } from '@/utils'
import { logService, type OptionGroup } from '@/services'
import type { OpLogItem } from '@/services/system/LogService'
import { findOptionGroup, getOptionLabel } from '@/utils/options'

const { t } = useI18n()
const optionGroups = ref<OptionGroup[]>([])
const methodOptions = computed(() => findOptionGroup(optionGroups.value, 'method'))

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

async function fetchOptions() {
  try {
    const res = await logService.getOptions()
    if (res.code !== 0 && res.code !== 200) throw new Error(res.msg)
    optionGroups.value = res.data || []
  } catch (e: any) {
    ElMessage.error(e?.message || t('common.loadFailed'))
  }
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
  fetchOptions()
  fetchList()
})
</script>

<template>
  <el-card>
    <template #header>
      {{ t('system.opLog') }}
    </template>

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
        <el-select
          v-model="queryForm.method"
          :placeholder="t('common.pleaseSelect')"
          clearable
          style="width: 200px"
        >
          <el-option
            v-for="item in methodOptions"
            :key="item.value"
            :label="getOptionLabel(t, item.code, item.value)"
            :value="item.code"
          />
        </el-select>
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
        <el-button type="primary" @click="onSearch">
          {{ t('common.search') }}
        </el-button>
        <el-button @click="onReset">
          {{ t('common.reset') }}
        </el-button>
      </el-form-item>
    </el-form>

    <!-- Table -->
    <el-table
      v-loading="loading"
      :data="list_ref"
      row-key="id"
      style="margin-bottom: 16px"
    >
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
      <el-table-column prop="createTimes" :label="t('common.createTimes')" min-width="170">
        <template #default="{ row }">
          <span style="color: #666">{{ formatDate(row.createTimes) }}</span>
        </template>
      </el-table-column>
    </el-table>

    <!-- Pagination -->
    <div style="display: flex; justify-content: flex-end; gap: 10px; align-items: center">
      <span>{{ t('common.totalItems', { count: pagination.total }) }}</span>
      <el-button :disabled="!pagination.hasPrev" @click="prevPage">
        {{ t('common.prevPage') }}
      </el-button>
      <el-button :disabled="!pagination.hasNext" @click="nextPage">
        {{ t('common.nextPage') }}
      </el-button>
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
