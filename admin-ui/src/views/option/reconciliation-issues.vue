<template>
  <div class="module-page">
    <CrudQueryCard :model="query" @search="search" @reset="resetQuery">
      <el-form-item v-if="authStore.isSystemAdmin" :label="t('option.tenantId')">
        <TenantSelect v-model="query.tenantId" class="tenant-select-filter" />
      </el-form-item>
      <el-form-item :label="t('option.bizNo')">
        <el-input v-model="query.bizNo" clearable />
      </el-form-item>
    </CrudQueryCard>

    <el-card v-loading="loading" shadow="never" class="table-card">
      <el-table :data="rows" stripe>
        <el-table-column
          v-if="authStore.isSystemAdmin"
          prop="tenantId"
          :label="t('option.tenantId')"
          width="110"
        />
        <el-table-column prop="issueKey" :label="t('option.issueKey')" min-width="220" />
        <el-table-column prop="bizNo" :label="t('option.bizNo')" min-width="170" />
        <el-table-column prop="checkType" :label="t('option.checkType')" width="110" />
        <el-table-column prop="status" :label="t('common.status')" width="90" />
        <el-table-column prop="occurrenceCount" :label="t('option.occurrenceCount')" width="110" />
        <el-table-column
          prop="expectedValue"
          :label="t('option.expectedValue')"
          min-width="180"
          show-overflow-tooltip
        />
        <el-table-column
          prop="actualValue"
          :label="t('option.actualValue')"
          min-width="180"
          show-overflow-tooltip
        />
        <el-table-column
          prop="detail"
          :label="t('option.detail')"
          min-width="240"
          show-overflow-tooltip
        />
      </el-table>
      <CursorPagination
        v-model:limit="page.limit"
        :total="page.total"
        :has-prev="page.hasPrev"
        :has-next="page.hasNext"
        @prev="pagination.prevAndLoad(loadRows)"
        @next="pagination.nextAndLoad(loadRows)"
        @limit-change="pagination.resetAndLoad(loadRows)"
      />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'
import TenantSelect from '@/components/TenantSelect.vue'
import { usePagination } from '@/composables'
import { optionService, type OptionReconciliationIssue } from '@/services'
import { useAuthStore } from '@/stores/auth'

const { t } = useI18n()
const authStore = useAuthStore()
const loading = ref(false)
const rows = ref<OptionReconciliationIssue[]>([])
const pagination = usePagination<number>(20)
const page = pagination.pagination
const query = reactive({
  tenantId: (authStore.isTenantUser ? authStore.profileTenantId : undefined) as number | undefined,
  bizNo: '',
})

async function loadRows() {
  loading.value = true
  try {
    const response = await optionService.listReconciliationIssues({
      tenantId: query.tenantId,
      bizNo: query.bizNo || undefined,
      cursor: page.cursor,
      limit: page.limit,
    })
    rows.value = response.data || []
    pagination.updateFromResponse(response)
  } finally {
    loading.value = false
  }
}

function search() {
  pagination.reset()
  void loadRows()
}

function resetQuery() {
  query.tenantId = authStore.isTenantUser ? authStore.profileTenantId || undefined : undefined
  query.bizNo = ''
  pagination.reset()
  void loadRows()
}

onMounted(loadRows)
</script>
