<template>
  <div class="module-page">
    <CrudQueryCard :model="riskQuery" @search="loadList" @reset="resetQuery">
      <el-form-item :label="t('trade.tenantId')">
        <TenantSelect v-model="riskQuery.tenantId" class="tenant-select-filter" />
      </el-form-item>
      <el-form-item :label="t('trade.userId')">
        <UserSelect v-model="riskQuery.userId" :tenant-id="riskQuery.tenantId || undefined" />
      </el-form-item>
      <el-form-item :label="t('trade.symbolId')">
        <SymbolSelect
          v-model="riskQuery.symbolId"
          :tenant-id="riskQuery.tenantId || undefined"
          :product-type="riskQuery.productType || undefined"
        />
      </el-form-item>
      <el-form-item :label="t('trade.productType')">
        <el-input-number v-model="riskQuery.productType" :min="0" :precision="0" />
      </el-form-item>
    </CrudQueryCard>
    <el-card shadow="never" class="table-card">
      <el-table v-loading="loading" :data="rows" stripe>
        <el-table-column
          prop="orderNo"
          :label="t('trade.orderNo')"
          min-width="180"
          show-overflow-tooltip
        />
        <el-table-column prop="userId" :label="t('trade.userId')" width="100" />
        <el-table-column prop="symbolId" :label="t('trade.symbolId')" width="100" />
        <el-table-column prop="checkType" :label="t('trade.checkType')" min-width="130">
          <template #default="{ row }">
            {{ checkTypeLabel(row.checkType) }}
          </template>
        </el-table-column>
        <el-table-column prop="checkResult" :label="t('trade.checkResult')" width="110">
          <template #default="{ row }">
            <el-tag :type="checkResultTagType(row.checkResult)">
              {{ checkResultLabel(row.checkResult) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          prop="rejectMsg"
          :label="t('trade.rejectMsg')"
          min-width="220"
          show-overflow-tooltip
        />
        <el-table-column
          :label="t('common.actions')"
          align="center"
          width="100"
          fixed="right"
        >
          <template #default="{ row }">
            <el-button link type="primary" @click="((detailData = row), (detailVisible = true))">
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
    <el-dialog v-model="detailVisible" :title="t('trade.riskCheckLogDetail')" width="900px">
      <template v-if="detailData">
        <div class="detail-section-title">
          {{ t('trade.basicInfo') }}
        </div>
        <el-descriptions :column="3" border>
          <el-descriptions-item :label="t('trade.id')">
            {{ detailData.id }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('trade.tenantId')">
            {{ detailData.tenantId }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('trade.userId')">
            {{ detailData.userId }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('trade.orderNo')" :span="2">
            {{ detailData.orderNo || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('trade.clientOrderId')">
            {{ detailData.clientOrderId || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('trade.symbolId')">
            {{ detailData.symbolId }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('trade.productType')">
            {{ productTypeLabel(detailData.productType) }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('trade.source')">
            {{ sourceLabel(detailData.source) }}
          </el-descriptions-item>
        </el-descriptions>

        <div class="detail-section-title">
          {{ t('trade.riskCheckInfo') }}
        </div>
        <el-descriptions :column="3" border>
          <el-descriptions-item :label="t('trade.checkType')">
            {{ checkTypeLabel(detailData.checkType) }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('trade.checkResult')">
            <el-tag :type="checkResultTagType(detailData.checkResult)">
              {{ checkResultLabel(detailData.checkResult) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item :label="t('trade.operatorId')">
            {{ detailData.operatorId || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('trade.requestPrice')">
            {{ detailData.requestPrice || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('trade.requestQty')">
            {{ detailData.requestQty || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('trade.requestAmount')">
            {{ detailData.requestAmount || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('trade.rejectCode')">
            {{ detailData.rejectCode || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('trade.rejectMsg')" :span="2">
            {{ detailData.rejectMsg || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('trade.createTimes')" :span="3">
            {{ formatDate(detailData.createTimes) }}
          </el-descriptions-item>
        </el-descriptions>

        <template v-if="detailData.checkSnapshot">
          <div class="detail-section-title">
            {{ t('trade.checkSnapshot') }}
          </div>
          <pre class="snapshot-pre">{{ formatSnapshot(detailData.checkSnapshot) }}</pre>
        </template>
      </template>
    </el-dialog>
  </div>
</template>
<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { usePagination } from '@/composables'
import { tradeService, type RiskOrderCheckLog } from '@/services'
import { formatDate } from '@/utils'
import TenantSelect from '@/components/TenantSelect.vue'
import UserSelect from '@/components/UserSelect.vue'
import SymbolSelect from '@/components/SymbolSelect.vue'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'

const { t } = useI18n()
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)

const loading = ref(false)
interface RiskQuery {
  tenantId?: number
  userId?: number
  symbolId?: number
  productType?: number
}

const rows = ref<RiskOrderCheckLog[]>([])
const detailVisible = ref(false)
const detailData = ref<RiskOrderCheckLog | null>(null)
const riskQuery = reactive<RiskQuery>({
  tenantId: undefined,
  userId: undefined,
  symbolId: undefined,
  productType: undefined,
})

const checkTypeLabels: Record<number, string> = {
  0: t('options.RISK_CHECK_TYPE_UNKNOWN'),
  1: t('options.RISK_CHECK_TYPE_BALANCE'),
  2: t('options.RISK_CHECK_TYPE_MARGIN'),
  3: t('options.RISK_CHECK_TYPE_POSITION'),
  4: t('options.RISK_CHECK_TYPE_TRADE_PERMISSION'),
  5: t('options.RISK_CHECK_TYPE_PRICE_PROTECT'),
  6: t('options.RISK_CHECK_TYPE_QTY_LIMIT'),
  7: t('options.RISK_CHECK_TYPE_NOTIONAL_LIMIT'),
  8: t('options.RISK_CHECK_TYPE_FREQUENCY_LIMIT'),
}
const checkResultLabels: Record<number, string> = {
  0: t('options.RISK_CHECK_RESULT_UNKNOWN'),
  1: t('options.RISK_CHECK_RESULT_PASS'),
  2: t('options.RISK_CHECK_RESULT_REJECT'),
  3: t('options.RISK_CHECK_RESULT_WARN_PASS'),
}
const productTypeLabels: Record<number, string> = {
  1: t('options.PRODUCT_TYPE_SPOT'),
  2: t('options.PRODUCT_TYPE_DERIVATIVE'),
  3: t('options.PRODUCT_TYPE_SECONDS'),
}
const sourceLabels: Record<number, string> = {
  1: t('options.SOURCE_TYPE_SYSTEM'),
  2: t('options.SOURCE_TYPE_USER'),
  3: t('options.SOURCE_TYPE_ADMIN'),
  4: t('options.SOURCE_TYPE_TASK'),
}

const valueLabel = (labels: Record<number, string>, value: number) => labels[value] || String(value)
const checkTypeLabel = (value: number) => valueLabel(checkTypeLabels, value)
const checkResultLabel = (value: number) => valueLabel(checkResultLabels, value)
const productTypeLabel = (value: number) => valueLabel(productTypeLabels, value)
const sourceLabel = (value: number) => valueLabel(sourceLabels, value)

function checkResultTagType(result: number): 'success' | 'warning' | 'danger' | 'info' {
  if (result === 1) return 'success'
  if (result === 2) return 'danger'
  if (result === 3) return 'warning'
  return 'info'
}

function formatSnapshot(value: string) {
  try {
    return JSON.stringify(JSON.parse(value), null, 2)
  } catch {
    return value
  }
}

const loadList = async () => {
  loading.value = true
  try {
    const res = await tradeService.listRiskLogs({
      ...riskQuery,
      cursor: pagination.cursor,
      limit: pagination.limit,
      count: pagination.total,
    })
    rows.value = res?.data || []
    updateFromResponse(res)
  } finally {
    loading.value = false
  }
}

function resetQuery() {
  riskQuery.tenantId = undefined
  riskQuery.userId = undefined
  riskQuery.symbolId = undefined
  riskQuery.productType = undefined
  resetAndLoad(loadList)
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

.snapshot-pre {
  max-height: 260px;
  margin: 0;
  padding: 16px;
  overflow: auto;
  border-radius: 8px;
  background: var(--el-fill-color-light);
  white-space: pre-wrap;
  word-break: break-all;
}
</style>
