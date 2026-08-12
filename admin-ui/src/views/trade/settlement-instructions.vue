<template>
  <div class="module-page">
    <CrudQueryCard :model="query" @search="load" @reset="resetQuery">
      <el-form-item :label="t('trade.tenantId')">
        <TenantSelect v-model="query.tenantId" />
      </el-form-item>
      <el-form-item :label="t('trade.bizType')">
        <el-input v-model="query.bizType" clearable />
      </el-form-item>
      <el-form-item :label="t('trade.bizId')">
        <el-input v-model="query.bizId" clearable />
      </el-form-item>
      <el-form-item :label="t('trade.orderId')">
        <el-input-number v-model="query.orderId" :min="0" :precision="0" />
      </el-form-item>
      <el-form-item :label="t('trade.status')">
        <el-select v-model="query.status" clearable>
          <el-option
            v-for="item in statusOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
      </el-form-item>
    </CrudQueryCard>
    <el-card shadow="never" class="table-card">
      <el-table v-loading="loading" :data="rows" stripe>
        <el-table-column
          prop="instructionNo"
          :label="t('trade.instructionNo')"
          min-width="220"
          show-overflow-tooltip
        />
        <el-table-column prop="bizType" :label="t('trade.bizType')" min-width="120" />
        <el-table-column
          prop="bizId"
          :label="t('trade.bizId')"
          min-width="180"
          show-overflow-tooltip
        />
        <el-table-column prop="orderId" :label="t('trade.orderId')" min-width="110" />
        <el-table-column prop="positionId" :label="t('trade.positionId')" min-width="110" />
        <el-table-column prop="userId" :label="t('trade.userId')" min-width="100" />
        <el-table-column prop="action" :label="t('trade.action')" min-width="140">
          <template #default="{ row }">
            {{ actionLabel(row.action) }}
          </template>
        </el-table-column>
        <el-table-column prop="asset" :label="t('trade.asset')" min-width="100" />
        <el-table-column prop="amount" :label="t('trade.amount')" min-width="130" />
        <el-table-column prop="status" :label="t('trade.status')" min-width="130">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)">
              {{ statusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="retryCount" :label="t('trade.retryCount')" min-width="110" />
        <el-table-column
          prop="lastErrorMsg"
          :label="t('trade.lastErrorMsg')"
          min-width="220"
          show-overflow-tooltip
        />
        <el-table-column
          prop="assetFlowNo"
          :label="t('trade.assetFlowNo')"
          min-width="210"
          show-overflow-tooltip
        />
        <el-table-column prop="reconciledAt" :label="t('trade.reconciledAt')" min-width="190">
          <template #default="{ row }">
            {{ formatTime(row.reconciledAt) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('common.actions')" width="90" fixed="right">
          <template #default="{ row }">
            <el-button
              v-perm="'trade:operation:settlement-instruction:retry'"
              link
              type="warning"
              :disabled="!canRetry(row)"
              @click="retry(row)"
            >
              {{ t('trade.retry') }}
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
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'
import TenantSelect from '@/components/TenantSelect.vue'
import { usePagination } from '@/composables'
import { formatDate } from '@/utils'
import { findFormOptionGroup, getOptionLabel, getOptionValueLabel } from '@/utils/options'
import { apiTradeListSettlementInstructions, apiTradeRetrySettlementInstruction } from '@/api/trade'
import { tradeService, type OptionGroup, type TradeSettlementInstruction } from '@/services'

const { t } = useI18n()
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)
const query = reactive<{
  tenantId?: number
  bizType: string
  bizId: string
  orderId?: number
  status?: number
}>({ bizType: '', bizId: '' })
const rows = ref<TradeSettlementInstruction[]>([])
const optionGroups = ref<OptionGroup[]>([])
const loading = ref(false)
const statusOptions = computed(() =>
  findFormOptionGroup(optionGroups.value, 'settlementInstructionStatus').map((item) => ({
    value: item.value,
    label: getOptionLabel(t, item.code, item.value),
  })),
)
function statusLabel(value: number) {
  return (
    getOptionValueLabel(optionGroups.value, 'settlementInstructionStatus', value, t) ||
    String(value)
  )
}
function actionLabel(value: number) {
  return (
    getOptionValueLabel(optionGroups.value, 'settlementInstructionAction', value, t) ||
    String(value)
  )
}
function statusTagType(value: number): 'success' | 'warning' | 'danger' | 'info' {
  if (value === 3) return 'success'
  if (value === 4 || value === 5) return 'danger'
  if (value === 1 || value === 2) return 'warning'
  return 'info'
}
async function load() {
  loading.value = true
  try {
    const res = await apiTradeListSettlementInstructions({
      tenantId: query.tenantId,
      bizType: query.bizType || undefined,
      bizId: query.bizId || undefined,
      orderId: query.orderId,
      status: query.status,
      cursor: pagination.cursor,
      limit: pagination.limit,
      count: pagination.total,
    })
    rows.value = res.data || []
    updateFromResponse(res)
  } finally {
    loading.value = false
  }
}
function resetQuery() {
  Object.assign(query, {
    tenantId: undefined,
    bizType: '',
    bizId: '',
    orderId: undefined,
    status: undefined,
  })
  resetAndLoad(load)
}
function canRetry(row: TradeSettlementInstruction) {
  return row.status === 4 || row.status === 5
}
async function retry(row: TradeSettlementInstruction) {
  const { value } = await ElMessageBox.prompt(t('trade.retryReason'), t('trade.retry'), {
    inputValidator: (text) => !!text?.trim() || t('trade.retryReasonRequired'),
  })
  await apiTradeRetrySettlementInstruction({
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
onMounted(async () => {
  const res = await tradeService.getOptions()
  optionGroups.value = res.data || []
  await load()
})
</script>
