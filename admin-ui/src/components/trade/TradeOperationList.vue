<template>
  <div>
    <CrudQueryCard :model="query" @search="load" @reset="reset">
      <el-form-item :label="t('trade.tenantId')">
        <TenantSelect v-model="query.tenantId" class="tenant-select-filter" />
      </el-form-item>
      <el-form-item v-for="field in filters" :key="field" :label="t(`trade.${field}`)">
        <el-input-number
          v-if="field !== 'bizType' && field !== 'bizId'"
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
            {{ formatValue(column, row[column]) }}
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
      <div class="list-footer">
        <span>{{ t('trade.total') }}: {{ total }}</span><el-button :disabled="!nextCursor" @click="load(nextCursor)">
          {{ t('trade.nextPage') }}
        </el-button>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'
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
const loading = ref(false),
  rows = ref<TradeOperationRecord[]>([]),
  total = ref(0),
  nextCursor = ref(0)
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
function params(cursor = 0) {
  const p: Record<string, any> = { cursor, limit: 20 }
  for (const k of ['tenantId', ...filters]) if (query[k] !== '' && query[k] != null) p[k] = query[k]
  return p
}
async function load(cursor = 0) {
  loading.value = true
  try {
    const res = await configs[props.kind].api(params(cursor))
    rows.value = res.data || []
    total.value = res.total || 0
    nextCursor.value = res.nextCursor || 0
  } finally {
    loading.value = false
  }
}
function reset() {
  Object.keys(query).forEach((k) => (query[k] = k === 'bizType' || k === 'bizId' ? '' : undefined))
  load()
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

<style scoped>
.list-footer {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  gap: 16px;
  margin-top: 16px;
}
</style>
