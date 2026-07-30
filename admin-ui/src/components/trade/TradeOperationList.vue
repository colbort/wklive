<template>
  <div class="module-page">
    <CrudQueryCard :model="query" @search="load" @reset="reset">
      <el-form-item :label="t('trade.tenantId')">
        <TenantSelect v-model="query.tenantId" class="tenant-select-filter" />
      </el-form-item>
      <el-form-item v-for="field in filters" :key="field" :label="t(`trade.${field}`)">
        <SymbolSelect
          v-if="field === 'symbolId'"
          v-model="query[field]"
          :tenant-id="query.tenantId || undefined"
        />
        <UserSelect
          v-else-if="field === 'userId'"
          v-model="query[field]"
          :tenant-id="query.tenantId || undefined"
        />
        <el-select
          v-else-if="field === 'status' || field === 'enabled' || field === 'snapshotType'"
          v-model="query[field]"
          :placeholder="t('common.pleaseSelect')"
          clearable
          class="query-field"
        >
          <el-option
            v-for="item in filterOptions(field)"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
        <el-input-number
          v-else-if="!['bizType', 'bizId', 'checkType', 'bizNo'].includes(field)"
          v-model="query[field]"
          :min="0"
          controls-position="right"
          class="query-field"
        />
        <el-input
          v-else
          v-model="query[field]"
          clearable
          class="query-field"
        />
      </el-form-item>
      <template v-if="$slots.actions" #actions>
        <slot name="actions" />
      </template>
    </CrudQueryCard>
    <el-card shadow="never" class="table-card">
      <el-table v-loading="loading" :data="rows" stripe>
        <el-table-column
          v-for="column in columns"
          :key="column"
          :prop="column"
          :label="t(`trade.${column}`)"
          :min-width="columnWidth(column)"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            <el-tag
              v-if="column === 'status' || column === 'enabled'"
              size="small"
              :type="statusTagType(Number(row[column]))"
              effect="light"
            >
              {{ formatStatus(column, row[column]) }}
            </el-tag>
            <el-tag
              v-else-if="column === 'snapshotType'"
              size="small"
              :type="snapshotTypeTagType(Number(row[column]))"
              effect="light"
            >
              {{ snapshotTypeLabel(Number(row[column])) }}
            </el-tag>
            <el-tag
              v-else-if="column === 'isSelected'"
              size="small"
              :type="Number(row[column]) === 1 ? 'success' : 'info'"
              effect="light"
            >
              {{ Number(row[column]) === 1 ? t('common.yes') : t('common.no') }}
            </el-tag>
            <template v-else>
              {{ formatValue(column, row[column]) }}
            </template>
          </template>
        </el-table-column>
        <el-table-column
          v-if="
            kind === 'instructions' ||
              kind === 'reconciliationIssues' ||
              kind === 'accountLiquidations'
          "
          :label="t('common.actions')"
          width="170"
          fixed="right"
        >
          <template #default="{ row }">
            <el-button
              v-if="kind === 'accountLiquidations'"
              v-perm="'trade:account-liquidation:detail'"
              link
              type="primary"
              @click="openAccountLiquidationDetail(row)"
            >
              {{ t('common.detail') }}
            </el-button>
            <el-button
              v-if="kind === 'accountLiquidations' && row.status === 5"
              v-perm="'trade:account-liquidation:retry'"
              link
              type="warning"
              @click="retryAccountLiquidation(row)"
            >
              {{ t('trade.retry') }}
            </el-button>
            <el-button
              v-if="kind === 'instructions' && (row.status === 4 || row.status === 5)"
              v-perm="'trade:operation:settlement-instruction:retry'"
              link
              type="warning"
              @click="retry(row)"
            >
              {{ t('trade.retry') }}
            </el-button>
            <el-button
              v-if="kind === 'reconciliationIssues' && row.status === 1"
              v-perm="'trade:operation:reconciliation-issue:ignore'"
              link
              type="warning"
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
    <el-dialog
      v-model="accountDetailVisible"
      :title="t('trade.accountLiquidationDetail')"
      width="86%"
      destroy-on-close
    >
      <div v-loading="accountDetailLoading">
        <el-descriptions v-if="accountDetail" :column="3" border>
          <el-descriptions-item
            v-for="field in accountSummaryFields"
            :key="field"
            :label="t(`trade.${field}`)"
          >
            <el-tag
              v-if="field === 'status'"
              :type="statusTagType(Number(accountDetail[field]))"
              effect="light"
              size="small"
            >
              {{ formatStatus(field, accountDetail[field]) }}
            </el-tag>
            <template v-else>
              {{ formatValue(field, accountDetail[field]) }}
            </template>
          </el-descriptions-item>
        </el-descriptions>
        <el-divider>{{ t('trade.accountLiquidationItems') }}</el-divider>
        <el-table :data="accountItems" stripe>
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
        <el-table :data="accountInstructions" stripe>
          <el-table-column prop="instructionNo" :label="t('trade.instructionNo')" min-width="220" />
          <el-table-column prop="action" :label="t('trade.action')" min-width="130">
            <template #default="{ row }">
              {{ optionValueLabel('settlementInstructionAction', Number(row.action)) }}
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
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'
import SymbolSelect from '@/components/SymbolSelect.vue'
import TenantSelect from '@/components/TenantSelect.vue'
import UserSelect from '@/components/UserSelect.vue'
import { usePagination } from '@/composables'
import { formatDate } from '@/utils'
import { findFormOptionGroup, getOptionLabel, getOptionValueLabel } from '@/utils/options'
import {
  tradeService,
  type ContractAccountLiquidation,
  type ContractAccountLiquidationItem,
  type OptionGroup,
  type TradeSettlementInstruction,
} from '@/services'
import {
  apiTradeGetAccountLiquidationDetail,
  apiTradeListAccountLiquidations,
  apiTradeListAssetReservations,
  apiTradeListDeliveryBatches,
  apiTradeListDeliverySettlements,
  apiTradeListFundingBatches,
  apiTradeListFundingSettlements,
  apiTradeListLiquidations,
  apiTradeListRiskTiers,
  apiTradeListSecondsPriceSnapshots,
  apiTradeListSettlementInstructions,
  apiTradeListContractReconciliationIssues,
  apiTradeIgnoreContractReconciliationIssue,
  apiTradeRetrySettlementInstruction,
  apiTradeRetryAccountLiquidation,
} from '@/api/trade'
type TradeOperationRecord = Record<string, string | number | null | undefined>
type ListApi = (params: Record<string, unknown>) => Promise<any>

type Kind =
  | 'riskTiers'
  | 'fundingBatches'
  | 'fundingSettlements'
  | 'deliveryBatches'
  | 'deliverySettlements'
  | 'liquidations'
  | 'accountLiquidations'
  | 'snapshots'
  | 'reservations'
  | 'instructions'
  | 'reconciliationIssues'
const props = defineProps<{ kind: Kind }>()
const { t } = useI18n()
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)
const loading = ref(false),
  rows = ref<TradeOperationRecord[]>([])
const accountDetailVisible = ref(false)
const accountDetailLoading = ref(false)
const accountDetail = ref<ContractAccountLiquidation | null>(null)
const accountItems = ref<ContractAccountLiquidationItem[]>([])
const accountInstructions = ref<TradeSettlementInstruction[]>([])
const optionGroups = ref<OptionGroup[]>([])
const query = reactive<Record<string, any>>({
  tenantId: undefined,
  symbolId: undefined,
  batchId: undefined,
  userId: undefined,
  positionId: undefined,
  orderId: undefined,
  status: undefined,
  enabled: undefined,
  snapshotType: undefined,
  bizType: '',
  bizId: '',
  checkType: '',
  bizNo: '',
  marginAsset: '',
})
const configs: Record<Kind, { filters: string[]; columns: string[]; api: ListApi }> = {
  riskTiers: {
    filters: ['symbolId', 'enabled'],
    columns: [
      'id',
      'symbolId',
      'tierNo',
      'notionalFloor',
      'notionalCap',
      'maxLeverage',
      'initialMarginRate',
      'maintenanceMarginRate',
      'enabled',
      'updateTimes',
    ],
    api: apiTradeListRiskTiers as unknown as ListApi,
  },
  fundingBatches: {
    filters: ['symbolId', 'status'],
    columns: [
      'batchNo',
      'symbolId',
      'fundingRate',
      'markPrice',
      'settlementTime',
      'status',
      'settledPositions',
      'totalPositions',
      'lastErrorMsg',
    ],
    api: apiTradeListFundingBatches as unknown as ListApi,
  },
  fundingSettlements: {
    filters: ['batchId', 'userId', 'positionId', 'status'],
    columns: [
      'settlementNo',
      'batchNo',
      'userId',
      'positionId',
      'positionSide',
      'feeAsset',
      'feeAmount',
      'status',
      'settledAt',
    ],
    api: apiTradeListFundingSettlements as unknown as ListApi,
  },
  deliveryBatches: {
    filters: ['symbolId', 'status'],
    columns: [
      'batchNo',
      'symbolId',
      'settlementPrice',
      'priceSource',
      'priceAlgorithm',
      'formulaVersion',
      'sampleSnapshot',
      'deliveryTime',
      'status',
      'settledPositions',
      'totalPositions',
      'lastErrorMsg',
    ],
    api: apiTradeListDeliveryBatches as unknown as ListApi,
  },
  deliverySettlements: {
    filters: ['batchId', 'userId', 'positionId', 'status'],
    columns: [
      'settlementNo',
      'batchNo',
      'userId',
      'positionId',
      'settlementPrice',
      'realizedPnl',
      'deliveryFee',
      'status',
      'settledAt',
    ],
    api: apiTradeListDeliverySettlements as unknown as ListApi,
  },
  liquidations: {
    filters: ['userId', 'symbolId', 'positionId', 'status'],
    columns: [
      'liquidationNo',
      'userId',
      'symbolId',
      'positionId',
      'positionSide',
      'triggerSnapshotId',
      'triggerQty',
      'liquidatedQty',
      'liquidationFee',
      'status',
      'reason',
      'completedAt',
    ],
    api: apiTradeListLiquidations as unknown as ListApi,
  },
  accountLiquidations: {
    filters: ['userId', 'marginAsset', 'status'],
    columns: [
      'liquidationNo',
      'userId',
      'marginAsset',
      'accountEquity',
      'maintenanceMargin',
      'riskRate',
      'positionCount',
      'grossSettlement',
      'deficitAmount',
      'insuranceFundAmount',
      'adlReliefAmount',
      'adlQty',
      'liquidationFee',
      'status',
      'reason',
      'completedAt',
    ],
    api: apiTradeListAccountLiquidations as unknown as ListApi,
  },
  snapshots: {
    filters: ['orderId', 'snapshotType'],
    columns: [
      'id',
      'orderId',
      'snapshotType',
      'source',
      'price',
      'quoteTime',
      'receivedAt',
      'algorithm',
      'isSelected',
    ],
    api: apiTradeListSecondsPriceSnapshots as unknown as ListApi,
  },
  reservations: {
    filters: ['orderId', 'status'],
    columns: [
      'reservationNo',
      'orderId',
      'asset',
      'reservedAmount',
      'consumedAmount',
      'releasedAmount',
      'status',
      'retryCount',
      'lastErrorMsg',
      'updateTimes',
    ],
    api: apiTradeListAssetReservations as unknown as ListApi,
  },
  instructions: {
    filters: ['bizType', 'bizId', 'orderId', 'status'],
    columns: [
      'instructionNo',
      'bizType',
      'bizId',
      'orderId',
      'positionId',
      'userId',
      'action',
      'asset',
      'amount',
      'status',
      'retryCount',
      'lastErrorMsg',
      'assetFlowNo',
      'reconciledAt',
    ],
    api: apiTradeListSettlementInstructions as unknown as ListApi,
  },
  reconciliationIssues: {
    filters: ['status', 'checkType', 'bizNo'],
    columns: [
      'issueKey',
      'checkType',
      'bizType',
      'bizNo',
      'instructionId',
      'expectedValue',
      'actualValue',
      'detail',
      'status',
      'occurrenceCount',
      'firstSeenAt',
      'lastSeenAt',
      'resolvedAt',
      'operatorId',
      'resolutionReason',
    ],
    api: apiTradeListContractReconciliationIssues as unknown as ListApi,
  },
}
const filters = configs[props.kind].filters,
  columns = configs[props.kind].columns
const statusOptionGroup = computed(() => {
  if (props.kind === 'reservations') return 'assetReservationStatus'
  if (props.kind === 'instructions') return 'settlementInstructionStatus'
  return ''
})
function params() {
  const p: Record<string, any> = { cursor: pagination.cursor, limit: pagination.limit }
  for (const k of ['tenantId', ...filters]) if (query[k] !== '' && query[k] != null) p[k] = query[k]
  return p
}
async function load() {
  loading.value = true
  try {
    const res = await configs[props.kind].api(params())
    rows.value = res.data || []
    updateFromResponse(res)
  } finally {
    loading.value = false
  }
}
function reset() {
  Object.keys(query).forEach(
    (k) =>
      (query[k] = ['bizType', 'bizId', 'checkType', 'bizNo', 'marginAsset'].includes(k)
        ? ''
        : undefined),
  )
  resetAndLoad(load)
}
function filterOptions(field: string) {
  if (field === 'enabled')
    return [
      { value: 1, label: t('common.enabled') },
      { value: 2, label: t('common.disabled') },
    ]
  if (field === 'snapshotType') return optionSelectItems('secondsPriceSnapshotType')
  if (field === 'status' && props.kind === 'reconciliationIssues')
    return [
      { value: 1, label: t('trade.reconciliationOpen') },
      { value: 2, label: t('trade.reconciliationResolved') },
      { value: 3, label: t('trade.reconciliationIgnored') },
    ]
  if (field === 'status' && props.kind === 'accountLiquidations')
    return [
      { value: 1, label: t('trade.accountLiquidationPending') },
      { value: 2, label: t('trade.accountLiquidationAssetSettling') },
      { value: 3, label: t('trade.accountLiquidationClosing') },
      { value: 4, label: t('trade.accountLiquidationCompleted') },
      { value: 5, label: t('trade.accountLiquidationManual') },
      { value: 6, label: t('trade.accountLiquidationInsuranceFund') },
      { value: 7, label: t('trade.accountLiquidationADL') },
    ]
  if (field === 'status' && statusOptionGroup.value)
    return optionSelectItems(statusOptionGroup.value)
  return Array.from({ length: 12 }, (_, index) => ({
    value: index + 1,
    label: String(index + 1),
  }))
}
function optionSelectItems(group: string) {
  return findFormOptionGroup(optionGroups.value, group).map((item) => ({
    value: item.value,
    label: getOptionLabel(t, item.code, item.value),
  }))
}
function optionValueLabel(group: string, value: number) {
  return getOptionValueLabel(optionGroups.value, group, value, t) || String(value)
}
function snapshotTypeLabel(value: number) {
  return optionValueLabel('secondsPriceSnapshotType', value)
}
function snapshotTypeTagType(value: number): 'primary' | 'warning' | 'success' | 'info' {
  if (value === 1) return 'primary'
  if (value === 2) return 'warning'
  if (value === 3) return 'success'
  return 'info'
}
function columnWidth(column: string) {
  if (column === 'source') return 220
  if (/(Time|Times|At)$/.test(column)) return 210
  if (/(No)$/.test(column)) return 210
  if (column === 'snapshotType') return 150
  if (column === 'algorithm') return 150
  if (column === 'isSelected') return 110
  if (column === 'status') return 130
  return 135
}
function statusTagType(value: number) {
  if (props.kind === 'reservations') {
    if (value === 2 || value === 4 || value === 6) return 'success'
    if (value === 7) return 'danger'
    if (value === 1 || value === 3 || value === 5) return 'warning'
    return 'info'
  }
  if (props.kind === 'instructions') {
    if (value === 3) return 'success'
    if (value === 4 || value === 5) return 'danger'
    if (value === 1 || value === 2) return 'warning'
    return 'info'
  }
  if (props.kind === 'reconciliationIssues') {
    if (value === 2) return 'success'
    if (value === 3) return 'info'
    if (value === 1) return 'danger'
    return 'info'
  }
  if (props.kind === 'accountLiquidations') {
    if (value === 4) return 'success'
    if (value === 5) return 'danger'
    if ((value >= 1 && value <= 3) || value === 6 || value === 7) return 'warning'
    return 'info'
  }
  if (value === 1 || value === 3) return 'success'
  if (value === 4 || value === 5) return 'danger'
  if (value === 2) return 'warning'
  return 'info'
}
function formatStatus(key: string, value: unknown) {
  if (key === 'enabled') return Number(value) === 1 ? t('common.enabled') : t('common.disabled')
  if (key === 'status' && statusOptionGroup.value)
    return optionValueLabel(statusOptionGroup.value, Number(value))
  if (key === 'status' && props.kind === 'reconciliationIssues') {
    if (Number(value) === 1) return t('trade.reconciliationOpen')
    if (Number(value) === 2) return t('trade.reconciliationResolved')
    if (Number(value) === 3) return t('trade.reconciliationIgnored')
  }
  if (key === 'status' && props.kind === 'accountLiquidations') {
    if (Number(value) === 1) return t('trade.accountLiquidationPending')
    if (Number(value) === 2) return t('trade.accountLiquidationAssetSettling')
    if (Number(value) === 3) return t('trade.accountLiquidationClosing')
    if (Number(value) === 4) return t('trade.accountLiquidationCompleted')
    if (Number(value) === 5) return t('trade.accountLiquidationManual')
    if (Number(value) === 6) return t('trade.accountLiquidationInsuranceFund')
    if (Number(value) === 7) return t('trade.accountLiquidationADL')
  }
  return formatValue(key, value)
}
function formatValue(key: string, value: any) {
  if (value == null || value === '') return '-'
  if (key === 'action' && props.kind === 'instructions')
    return optionValueLabel('settlementInstructionAction', Number(value))
  if (/(Time|Times|At)$/.test(key) && Number(value) > 0) return formatDate(Number(value))
  return value
}
async function retry(row: TradeOperationRecord) {
  const { value } = await ElMessageBox.prompt(t('trade.retryReason'), t('trade.retry'), {
    inputValidator: (v) => !!v || t('trade.retryReasonRequired'),
  })
  await apiTradeRetrySettlementInstruction({ id: Number(row.id), reason: value })
  ElMessage.success(t('common.success'))
  load()
}
async function ignoreIssue(row: TradeOperationRecord) {
  const { value } = await ElMessageBox.prompt(t('trade.ignoreReason'), t('trade.ignoreIssue'), {
    inputValidator: (v) => {
      if (!v?.trim()) return t('trade.ignoreReasonRequired')
      if (new TextEncoder().encode(v.trim()).length > 500) return t('trade.ignoreReasonTooLong')
      return true
    },
  })
  await apiTradeIgnoreContractReconciliationIssue({
    id: Number(row.id),
    tenantId: query.tenantId,
    reason: value.trim(),
  })
  ElMessage.success(t('common.success'))
  load()
}
type AccountSummaryField = Extract<keyof ContractAccountLiquidation, string>
const accountSummaryFields: AccountSummaryField[] = [
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
async function openAccountLiquidationDetail(row: TradeOperationRecord) {
  accountDetailVisible.value = true
  accountDetailLoading.value = true
  try {
    const res = await apiTradeGetAccountLiquidationDetail({
      id: Number(row.id),
      tenantId: query.tenantId,
    })
    accountDetail.value = res.data ?? null
    accountItems.value = res.items || []
    accountInstructions.value = res.settlementInstructions || []
  } finally {
    accountDetailLoading.value = false
  }
}
async function retryAccountLiquidation(row: TradeOperationRecord) {
  const { value } = await ElMessageBox.prompt(t('trade.retryReason'), t('trade.retry'), {
    inputValidator: (v) => !!v?.trim() || t('trade.retryReasonRequired'),
  })
  await apiTradeRetryAccountLiquidation({
    id: Number(row.id),
    tenantId: query.tenantId,
    reason: value.trim(),
  })
  ElMessage.success(t('common.success'))
  await load()
}
onMounted(async () => {
  const optionsResponse = await tradeService.getOptions()
  optionGroups.value = optionsResponse.data || []
  await load()
})
</script>
