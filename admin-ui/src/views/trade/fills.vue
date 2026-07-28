<template>
  <div class="module-page">
    <CrudQueryCard :model="currentQuery" @search="loadList" @reset="resetQuery">
      <el-form-item :label="t('trade.tenantId')">
        <TenantSelect v-model="currentQuery.tenantId" class="tenant-select-filter" />
      </el-form-item>
      <el-form-item :label="t('trade.userId')">
        <UserSelect v-model="currentQuery.userId" :tenant-id="currentQuery.tenantId || undefined" />
      </el-form-item>
      <el-form-item :label="t('trade.symbolId')">
        <SymbolSelect
          v-model="currentQuery.symbolId"
          :tenant-id="currentQuery.tenantId || undefined"
        />
      </el-form-item>
      <el-form-item :label="t('common.keyword')">
        <el-input v-model="currentQuery.keyword" clearable />
      </el-form-item>
    </CrudQueryCard>

    <el-card shadow="never" class="table-card">
      <el-table v-loading="loading" :data="rows" stripe>
        <el-table-column
          v-for="column in currentColumns"
          :key="column.prop"
          :prop="column.prop"
          :label="column.label"
          :min-width="column.width || 140"
          show-overflow-tooltip
        />

        <el-table-column
          :label="t('common.actions')"
          align="center"
          width="100"
          fixed="right"
        >
          <template #default="{ row }">
            <el-button
              v-perm="'trade:fill:detail'"
              link
              type="primary"
              @click="showDetail(row)"
            >
              {{ t('option.detail') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <CursorPagination
        v-model:limit="pagination.limit"
        :total="pagination.total"
        :has-prev="pagination.hasPrev"
        :has-next="pagination.hasNext"
        @prev="handlePrevPage"
        @next="handleNextPage"
        @limit-change="handleLimitChange"
      />
    </el-card>

    <el-dialog v-model="detailVisible" :title="t('trade.fillDetail')" width="960px">
      <template v-if="detailData">
        <div class="detail-section-title">
          {{ t('trade.basicInfo') }}
        </div>
        <el-descriptions :column="3" border>
          <el-descriptions-item :label="t('trade.fillNo')" :span="2">
            {{ detailData.fillNo }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('trade.id')">
            {{ detailData.id }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('trade.tenantId')">
            {{ detailData.tenantId }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('trade.userId')">
            {{ detailData.userId }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('trade.symbolId')">
            {{ detailData.symbolId }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('trade.orderNo')" :span="2">
            {{ detailData.orderNo }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('trade.orderId')">
            {{ detailData.orderId }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('trade.matchNo')" :span="2">
            {{ detailData.matchNo }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('trade.productType')">
            {{ optionLabel('productType', detailData.productType) }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('trade.contractType')">
            {{ optionLabel('contractType', detailData.contractType) }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('trade.contractValueType')">
            {{ optionLabel('contractValueType', detailData.contractValueType) }}
          </el-descriptions-item>
        </el-descriptions>

        <div class="detail-section-title">
          {{ t('trade.fillInfo') }}
        </div>
        <el-descriptions :column="3" border>
          <el-descriptions-item :label="t('trade.side')">
            <el-tag :type="detailData.side === 1 ? 'success' : 'danger'">
              {{ enumLabel(tradeSideLabels, detailData.side) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item :label="t('trade.positionSide')">
            {{ enumLabel(positionSideLabels, detailData.positionSide) }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('trade.liquidityType')">
            {{ enumLabel(liquidityTypeLabels, detailData.liquidityType) }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('trade.price')">
            {{ detailData.price }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('trade.qty')">
            {{ detailData.qty }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('trade.amount')">
            {{ detailData.amount }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('trade.fee')">
            {{ detailData.fee }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('trade.feeAsset')">
            {{ detailData.feeAsset || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('trade.realizedPnl')">
            {{ detailData.realizedPnl }}
          </el-descriptions-item>
        </el-descriptions>

        <div class="detail-section-title">
          {{ t('trade.settlementInfo') }}
        </div>
        <el-descriptions :column="3" border>
          <el-descriptions-item :label="t('trade.settlementStatus')">
            <el-tag :type="settlementTagType(detailData.settlementStatus)">
              {{ enumLabel(settlementStatusLabels, detailData.settlementStatus) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item :label="t('trade.settlementRetryCount')">
            {{ detailData.settlementRetryCount }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('trade.settledAt')">
            {{ formatDate(detailData.settledAt) }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('trade.matchTime')">
            {{ formatDate(detailData.matchTime) }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('trade.createTimes')" :span="2">
            {{ formatDate(detailData.createTimes) }}
          </el-descriptions-item>
        </el-descriptions>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { usePagination } from '@/composables'
import { tradeService, type TradeFill } from '@/services'
import { formatDate } from '@/utils'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'
import TenantSelect from '@/components/TenantSelect.vue'
import UserSelect from '@/components/UserSelect.vue'
import SymbolSelect from '@/components/SymbolSelect.vue'

const { t } = useI18n()
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)

interface CurrentQuery {
  tenantId: number | undefined
  userId: number | undefined
  symbolId: number | undefined
  keyword: string
  limit: number
}

interface CurrentColumn {
  prop: string
  label: string
  width?: number
}

const loading = ref(false)
const rows = ref<TradeFill[]>([])
const detailVisible = ref(false)
const detailData = ref<TradeFill | null>(null)

const currentQuery = reactive<CurrentQuery>({
  tenantId: undefined,
  userId: undefined,
  symbolId: undefined,
  keyword: '',
  limit: 20,
})

const currentColumns: CurrentColumn[] = [
  { prop: 'fillNo', label: t('trade.fillNo'), width: 180 },
  { prop: 'orderNo', label: t('trade.orderNo'), width: 180 },
  { prop: 'userId', label: t('trade.userId'), width: 100 },
  { prop: 'price', label: t('trade.price') },
  { prop: 'qty', label: t('trade.qty') },
]

const tradeSideLabels: Record<number, string> = {
  1: t('options.TRADE_SIDE_BUY'),
  2: t('options.TRADE_SIDE_SELL'),
}
const productTypeLabels: Record<number, string> = {
  1: t('options.PRODUCT_TYPE_SPOT'),
  2: t('options.PRODUCT_TYPE_DERIVATIVE'),
  3: t('options.PRODUCT_TYPE_SECONDS'),
}
const contractTypeLabels: Record<number, string> = {
  0: t('options.CONTRACT_TYPE_NOT_APPLICABLE'),
  1: t('options.CONTRACT_TYPE_PERPETUAL'),
  2: t('options.CONTRACT_TYPE_DELIVERY'),
}
const contractValueTypeLabels: Record<number, string> = {
  0: t('options.CONTRACT_VALUE_TYPE_NOT_APPLICABLE'),
  1: t('options.CONTRACT_VALUE_TYPE_LINEAR'),
  2: t('options.CONTRACT_VALUE_TYPE_INVERSE'),
}
const positionSideLabels: Record<number, string> = {
  1: t('options.POSITION_SIDE_NET'),
  2: t('options.POSITION_SIDE_LONG'),
  3: t('options.POSITION_SIDE_SHORT'),
}
const liquidityTypeLabels: Record<number, string> = {
  1: t('options.LIQUIDITY_TYPE_MAKER'),
  2: t('options.LIQUIDITY_TYPE_TAKER'),
}
const settlementStatusLabels: Record<number, string> = {
  1: t('options.FILL_SETTLEMENT_STATUS_PENDING'),
  2: t('options.FILL_SETTLEMENT_STATUS_PROCESSING'),
  3: t('options.FILL_SETTLEMENT_STATUS_SETTLED'),
  4: t('options.FILL_SETTLEMENT_STATUS_FAILED'),
  5: t('options.FILL_SETTLEMENT_STATUS_MANUAL_REVIEW'),
}

function enumLabel(labels: Record<number, string>, value: number) {
  return labels[value] || String(value)
}

function optionLabel(group: string, value: number) {
  const groups: Record<string, Record<number, string>> = {
    productType: productTypeLabels,
    contractType: contractTypeLabels,
    contractValueType: contractValueTypeLabels,
  }
  return enumLabel(groups[group] || {}, value)
}

function settlementTagType(status: number): 'success' | 'warning' | 'danger' | 'info' {
  if (status === 3) return 'success'
  if (status === 4 || status === 5) return 'danger'
  if (status === 1 || status === 2) return 'warning'
  return 'info'
}

const loadList = async () => {
  loading.value = true
  try {
    const res = await tradeService.listFills({
      ...currentQuery,
      cursor: pagination.cursor,
      limit: pagination.limit,
    })
    rows.value = res?.data || []
    updateFromResponse(res)
  } finally {
    loading.value = false
  }
}

const resetQuery = () => {
  currentQuery.tenantId = undefined
  currentQuery.userId = undefined
  currentQuery.symbolId = undefined
  currentQuery.keyword = ''
  currentQuery.limit = 100
  loadList()
}

const showDetail = async (row: TradeFill) => {
  detailData.value =
    (await tradeService.getFill({ tenantId: row.tenantId, id: row.id })).data || row
  detailVisible.value = true
}

function handleLimitChange() {
  resetAndLoad(loadList)
}

function handlePrevPage() {
  prevAndLoad(loadList)
}

function handleNextPage() {
  nextAndLoad(loadList)
}

onMounted(loadList)
</script>

<style scoped>
.detail-section-title {
  margin: 20px 0 12px;
  color: var(--el-text-color-primary);
  font-size: 15px;
  font-weight: 600;
}

.detail-section-title:first-child {
  margin-top: 0;
}
</style>
