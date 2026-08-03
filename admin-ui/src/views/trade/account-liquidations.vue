<template>
  <div class="module-page">
    <CrudQueryCard :model="query" @search="load" @reset="resetQuery">
      <el-form-item :label="t('trade.tenantId')">
        <TenantSelect v-model="query.tenantId" />
      </el-form-item>
      <el-form-item :label="t('trade.userId')">
        <UserSelect v-model="query.userId" :tenant-id="query.tenantId" />
      </el-form-item>
      <el-form-item :label="t('trade.marginAsset')">
        <el-input v-model="query.marginAsset" clearable />
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
          prop="liquidationNo"
          :label="t('trade.liquidationNo')"
          min-width="210"
          show-overflow-tooltip
        />
        <el-table-column prop="userId" :label="t('trade.userId')" min-width="100" />
        <el-table-column prop="marginAsset" :label="t('trade.marginAsset')" min-width="120" />
        <el-table-column prop="accountEquity" :label="t('trade.accountEquity')" min-width="140" />
        <el-table-column
          prop="maintenanceMargin"
          :label="t('trade.maintenanceMargin')"
          min-width="160"
        />
        <el-table-column prop="riskRate" :label="t('trade.riskRate')" min-width="120" />
        <el-table-column prop="positionCount" :label="t('trade.positionCount')" min-width="130" />
        <el-table-column
          prop="grossSettlement"
          :label="t('trade.grossSettlement')"
          min-width="150"
        />
        <el-table-column prop="deficitAmount" :label="t('trade.deficitAmount')" min-width="140" />
        <el-table-column
          prop="insuranceFundAmount"
          :label="t('trade.insuranceFundAmount')"
          min-width="170"
        />
        <el-table-column
          prop="adlReliefAmount"
          :label="t('trade.adlReliefAmount')"
          min-width="150"
        />
        <el-table-column prop="adlQty" :label="t('trade.adlQty')" min-width="120" />
        <el-table-column prop="liquidationFee" :label="t('trade.liquidationFee')" min-width="140" />
        <el-table-column prop="status" :label="t('trade.status')" min-width="150">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)">
              {{ statusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          prop="reason"
          :label="t('trade.reason')"
          min-width="180"
          show-overflow-tooltip
        />
        <el-table-column prop="completedAt" :label="t('trade.completedAt')" min-width="190">
          <template #default="{ row }">
            {{ formatTime(row.completedAt) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('common.actions')" width="150" fixed="right">
          <template #default="{ row }">
            <el-button
              v-perm="'trade:account-liquidation:detail'"
              link
              type="primary"
              @click="openDetail(row)"
            >
              {{ t('common.detail') }}
            </el-button>
            <el-button
              v-perm="'trade:account-liquidation:retry'"
              link
              type="warning"
              :disabled="row.status !== 5"
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

    <el-dialog
      v-model="detailVisible"
      :title="t('trade.accountLiquidationDetail')"
      width="86%"
      destroy-on-close
    >
      <div v-loading="detailLoading">
        <el-descriptions v-if="detail" :column="3" border>
          <el-descriptions-item
            v-for="field in summaryFields"
            :key="field"
            :label="t(`trade.${field}`)"
          >
            <el-tag v-if="field === 'status'" :type="statusTagType(detail.status)">
              {{ statusLabel(detail.status) }}
            </el-tag>
            <template v-else>
              {{ formatValue(field, detail[field]) }}
            </template>
          </el-descriptions-item>
        </el-descriptions>
        <el-divider>{{ t('trade.accountLiquidationItems') }}</el-divider>
        <el-table :data="items" stripe>
          <el-table-column prop="positionId" :label="t('trade.positionId')" min-width="130" />
          <el-table-column prop="symbolId" :label="t('trade.symbolId')" min-width="130" />
          <el-table-column prop="triggerQty" :label="t('trade.triggerQty')" min-width="130" />
          <el-table-column
            prop="triggerMarkPrice"
            :label="t('trade.triggerMarkPrice')"
            min-width="150"
          />
          <el-table-column
            prop="positionMargin"
            :label="t('trade.positionMargin')"
            min-width="150"
          />
          <el-table-column prop="realizedPnl" :label="t('trade.realizedPnl')" min-width="140" />
          <el-table-column
            prop="liquidationFee"
            :label="t('trade.liquidationFee')"
            min-width="140"
          />
          <el-table-column prop="deficitAmount" :label="t('trade.deficitAmount')" min-width="140" />
          <el-table-column
            prop="bankruptcyPrice"
            :label="t('trade.bankruptcyPrice')"
            min-width="150"
          />
          <el-table-column
            prop="adlReliefAmount"
            :label="t('trade.adlReliefAmount')"
            min-width="150"
          />
          <el-table-column prop="adlQty" :label="t('trade.adlQty')" min-width="120" />
          <el-table-column prop="status" :label="t('trade.status')" min-width="100" />
        </el-table>
        <el-divider>{{ t('trade.settlementInstructions') }}</el-divider>
        <el-table :data="instructions" stripe>
          <el-table-column prop="instructionNo" :label="t('trade.instructionNo')" min-width="220" />
          <el-table-column prop="action" :label="t('trade.action')" min-width="130">
            <template #default="{ row }">
              {{ instructionActionLabel(row.action) }}
            </template>
          </el-table-column>
          <el-table-column prop="amount" :label="t('trade.amount')" min-width="130" />
          <el-table-column prop="status" :label="t('trade.status')" min-width="100" />
          <el-table-column prop="retryCount" :label="t('trade.retryCount')" min-width="110" />
          <el-table-column prop="assetFlowNo" :label="t('trade.assetFlowNo')" min-width="220" />
          <el-table-column
            prop="lastErrorMsg"
            :label="t('trade.lastErrorMsg')"
            min-width="220"
            show-overflow-tooltip
          />
        </el-table>
      </div>
    </el-dialog>
  </div>
</template>
<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'
import TenantSelect from '@/components/TenantSelect.vue'
import UserSelect from '@/components/UserSelect.vue'
import { usePagination } from '@/composables'
import { formatDate } from '@/utils'
import { getOptionValueLabel } from '@/utils/options'
import {
  apiTradeGetAccountLiquidationDetail,
  apiTradeListAccountLiquidations,
  apiTradeRetryAccountLiquidation,
} from '@/api/trade'
import {
  tradeService,
  type ContractAccountLiquidation,
  type ContractAccountLiquidationItem,
  type OptionGroup,
  type TradeSettlementInstruction,
} from '@/services'

const { t } = useI18n()
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)
const query = reactive<{
  tenantId?: number
  userId?: number
  marginAsset: string
  status?: number
}>({ marginAsset: '' })
const rows = ref<ContractAccountLiquidation[]>([])
const loading = ref(false)
const detailVisible = ref(false)
const detailLoading = ref(false)
const detail = ref<ContractAccountLiquidation | null>(null)
const items = ref<ContractAccountLiquidationItem[]>([])
const instructions = ref<TradeSettlementInstruction[]>([])
const optionGroups = ref<OptionGroup[]>([])
const statusOptions = [1, 2, 3, 4, 5, 6, 7].map((value) => ({ value, label: '' }))
function statusLabel(value: number) {
  return (
    (
      {
        1: t('trade.accountLiquidationPending'),
        2: t('trade.accountLiquidationAssetSettling'),
        3: t('trade.accountLiquidationClosing'),
        4: t('trade.accountLiquidationCompleted'),
        5: t('trade.accountLiquidationManual'),
        6: t('trade.accountLiquidationInsuranceFund'),
        7: t('trade.accountLiquidationADL'),
      } as Record<number, string>
    )[value] || String(value)
  )
}
statusOptions.forEach((item) => {
  item.label = statusLabel(item.value)
})
function statusTagType(value: number): 'success' | 'warning' | 'danger' | 'info' {
  if (value === 4) return 'success'
  if (value === 5) return 'danger'
  if ([1, 2, 3, 6, 7].includes(value)) return 'warning'
  return 'info'
}
async function load() {
  loading.value = true
  try {
    const res = await apiTradeListAccountLiquidations({
      tenantId: query.tenantId,
      userId: query.userId,
      marginAsset: query.marginAsset || undefined,
      status: query.status,
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
  Object.assign(query, {
    tenantId: undefined,
    userId: undefined,
    marginAsset: '',
    status: undefined,
  })
  resetAndLoad(load)
}
async function openDetail(row: ContractAccountLiquidation) {
  detailVisible.value = true
  detailLoading.value = true
  try {
    const res = await apiTradeGetAccountLiquidationDetail({ tenantId: query.tenantId, id: row.id })
    detail.value = res.data || null
    items.value = res.items || []
    instructions.value = res.settlementInstructions || []
  } finally {
    detailLoading.value = false
  }
}
async function retry(row: ContractAccountLiquidation) {
  const { value } = await ElMessageBox.prompt(t('trade.retryReason'), t('trade.retry'), {
    inputValidator: (text) => !!text?.trim() || t('trade.retryReasonRequired'),
  })
  await apiTradeRetryAccountLiquidation({
    tenantId: query.tenantId,
    id: row.id,
    reason: value.trim(),
  })
  ElMessage.success(t('common.success'))
  await load()
}
type SummaryField = Extract<keyof ContractAccountLiquidation, string>
const summaryFields: SummaryField[] = [
  'liquidationNo',
  'userId',
  'marginAsset',
  'marginSnapshotId',
  'assetVersion',
  'walletBalance',
  'positionMargin',
  'maintenanceMargin',
  'accountEquity',
  'riskRate',
  'grossSettlement',
  'liquidationFee',
  'userCredit',
  'userDebit',
  'deficitAmount',
  'insuranceFundAmount',
  'adlReliefAmount',
  'adlQty',
  'positionCount',
  'status',
  'reason',
  'startedAt',
  'completedAt',
]
function formatValue(key: string, value: unknown) {
  if (value == null || value === '') return '-'
  if (/(Time|Times|At)$/.test(key) && Number(value) > 0) return formatDate(Number(value))
  return value
}
function instructionActionLabel(value: number) {
  return (
    getOptionValueLabel(optionGroups.value, 'settlementInstructionAction', value, t) ||
    String(value)
  )
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
