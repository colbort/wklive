<template>
  <div class="module-page">
    <CrudQueryCard :model="query" @search="load" @reset="resetQuery">
      <el-form-item :label="t('trade.tenantId')">
        <TenantSelect v-model="query.tenantId" />
      </el-form-item>
      <el-form-item :label="t('trade.status')">
        <el-select v-model="query.status" clearable>
          <el-option :label="t('trade.reconciliationOpen')" :value="1" />
          <el-option :label="t('trade.reconciliationResolved')" :value="2" />
          <el-option :label="t('trade.reconciliationIgnored')" :value="3" />
        </el-select>
      </el-form-item>
      <el-form-item :label="t('trade.checkType')">
        <el-input v-model="query.checkType" clearable />
      </el-form-item>
      <el-form-item :label="t('trade.bizNo')">
        <el-input v-model="query.bizNo" clearable />
      </el-form-item>
    </CrudQueryCard>
    <el-card shadow="never" class="table-card">
      <el-table v-loading="loading" :data="rows" stripe>
        <el-table-column
          prop="issueKey"
          :label="t('trade.issueKey')"
          min-width="220"
          show-overflow-tooltip
        />
        <el-table-column prop="checkType" :label="t('trade.checkType')" min-width="140" />
        <el-table-column prop="bizType" :label="t('trade.bizType')" min-width="120" />
        <el-table-column
          prop="bizNo"
          :label="t('trade.bizNo')"
          min-width="200"
          show-overflow-tooltip
        />
        <el-table-column prop="instructionId" :label="t('trade.instructionId')" min-width="130" />
        <el-table-column
          prop="expectedValue"
          :label="t('trade.expectedValue')"
          min-width="150"
          show-overflow-tooltip
        />
        <el-table-column
          prop="actualValue"
          :label="t('trade.actualValue')"
          min-width="150"
          show-overflow-tooltip
        />
        <el-table-column
          prop="detail"
          :label="t('trade.detail')"
          min-width="220"
          show-overflow-tooltip
        />
        <el-table-column prop="status" :label="t('trade.status')" min-width="110">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)">
              {{ statusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          prop="occurrenceCount"
          :label="t('trade.occurrenceCount')"
          min-width="130"
        />
        <el-table-column prop="firstSeenAt" :label="t('trade.firstSeenAt')" min-width="190">
          <template #default="{ row }">
            {{ formatTime(row.firstSeenAt) }}
          </template>
        </el-table-column>
        <el-table-column prop="lastSeenAt" :label="t('trade.lastSeenAt')" min-width="190">
          <template #default="{ row }">
            {{ formatTime(row.lastSeenAt) }}
          </template>
        </el-table-column>
        <el-table-column
          prop="resolutionReason"
          :label="t('trade.resolutionReason')"
          min-width="180"
          show-overflow-tooltip
        />
        <el-table-column :label="t('common.actions')" width="110" fixed="right">
          <template #default="{ row }">
            <el-button
              v-perm="'trade:operation:reconciliation-issue:ignore'"
              link
              type="warning"
              :disabled="row.status !== 1"
              @click="ignoreIssue(row)"
            >
              {{ t('trade.ignoreIssue') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
      <CursorPagination
        v-model:limit="pagination.limit"
        :total="pagination.total"
        :has-prev="pagination.hasPrev"
        :has-next="pagination.hasNext"
        @prev="prevAndLoad(load)"
        @next="nextAndLoad(load)"
        @limit-change="resetAndLoad(load)"
      />
    </el-card>
  </div>
</template>
<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'
import TenantSelect from '@/components/TenantSelect.vue'
import { usePagination } from '@/composables'
import { formatDate } from '@/utils'
import {
  apiTradeIgnoreContractReconciliationIssue,
  apiTradeListContractReconciliationIssues,
} from '@/api/trade'
import type { ContractReconciliationIssue } from '@/services'

const { t } = useI18n()
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)
const query = reactive<{ tenantId?: number; status?: number; checkType: string; bizNo: string }>({
  checkType: '',
  bizNo: '',
})
const rows = ref<ContractReconciliationIssue[]>([])
const loading = ref(false)
function statusLabel(value: number) {
  if (value === 1) return t('trade.reconciliationOpen')
  if (value === 2) return t('trade.reconciliationResolved')
  if (value === 3) return t('trade.reconciliationIgnored')
  return String(value)
}
function statusTagType(value: number): 'success' | 'danger' | 'info' {
  if (value === 1) return 'danger'
  if (value === 2) return 'success'
  return 'info'
}
async function load() {
  loading.value = true
  try {
    const res = await apiTradeListContractReconciliationIssues({
      tenantId: query.tenantId,
      status: query.status,
      checkType: query.checkType || undefined,
      bizNo: query.bizNo || undefined,
      cursor: pagination.cursor,
      limit: pagination.limit,
    })
    rows.value = res.data || []
    updateFromResponse(res)
  } finally {
    loading.value = false
  }
}
function resetQuery() {
  Object.assign(query, { tenantId: undefined, status: undefined, checkType: '', bizNo: '' })
  resetAndLoad(load)
}
async function ignoreIssue(row: ContractReconciliationIssue) {
  const { value } = await ElMessageBox.prompt(t('trade.ignoreReason'), t('trade.ignoreIssue'), {
    inputValidator: (text) => {
      if (!text?.trim()) return t('trade.ignoreReasonRequired')
      if (new TextEncoder().encode(text.trim()).length > 500) return t('trade.ignoreReasonTooLong')
      return true
    },
  })
  await apiTradeIgnoreContractReconciliationIssue({
    tenantId: query.tenantId,
    id: row.id,
    reason: value.trim(),
  })
  ElMessage.success(t('common.success'))
  await load()
}
function formatTime(value: number) {
  return value > 0 ? formatDate(value) : '-'
}
onMounted(load)
</script>
