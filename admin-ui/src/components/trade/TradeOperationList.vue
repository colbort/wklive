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
          v-else-if="field === 'status' || field === 'enabled'"
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
          v-else-if="field !== 'bizType' && field !== 'bizId'"
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
          min-width="135"
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
            <template v-else>
              {{ formatValue(column, row[column]) }}
            </template>
          </template>
        </el-table-column>
        <el-table-column
          v-if="kind === 'instructions'"
          :label="t('common.actions')"
          width="110"
          fixed="right"
        >
          <template #default="{ row }">
            <el-button
              v-if="row.status === 4 || row.status === 5"
              v-perm="'trade:operation:settlement-instruction:retry'"
              link
              type="warning"
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
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { usePagination } from '@/composables'
import {
  apiTradeListAssetReservations,
  apiTradeListDeliveryBatches,
  apiTradeListDeliverySettlements,
  apiTradeListFundingBatches,
  apiTradeListFundingSettlements,
  apiTradeListLiquidations,
  apiTradeListRiskTiers,
  apiTradeListSecondsPriceSnapshots,
  apiTradeListSettlementInstructions,
  apiTradeRetrySettlementInstruction,
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
  | 'snapshots'
  | 'reservations'
  | 'instructions'
const props = defineProps<{ kind: Kind }>()
const { t } = useI18n()
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)
const loading = ref(false),
  rows = ref<TradeOperationRecord[]>([])
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
    ],
    api: apiTradeListSettlementInstructions as unknown as ListApi,
  },
}
const filters = configs[props.kind].filters,
  columns = configs[props.kind].columns
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
  Object.keys(query).forEach((k) => (query[k] = k === 'bizType' || k === 'bizId' ? '' : undefined))
  resetAndLoad(load)
}
function filterOptions(field: string) {
  if (field === 'enabled')
    return [
      { value: 1, label: t('common.enabled') },
      { value: 2, label: t('common.disabled') },
    ]
  return Array.from({ length: 12 }, (_, index) => ({
    value: index + 1,
    label: String(index + 1),
  }))
}
function statusTagType(value: number) {
  if (value === 1 || value === 3) return 'success'
  if (value === 4 || value === 5) return 'danger'
  if (value === 2) return 'warning'
  return 'info'
}
function formatStatus(key: string, value: unknown) {
  if (key === 'enabled') return Number(value) === 1 ? t('common.enabled') : t('common.disabled')
  return formatValue(key, value)
}
function formatValue(key: string, value: any) {
  if (value == null || value === '') return '-'
  if (/(Time|Times|At)$/.test(key) && Number(value) > 0)
    return new Date(Number(value)).toLocaleString()
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
onMounted(() => load())
</script>
