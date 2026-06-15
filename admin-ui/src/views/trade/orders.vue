<template>
  <div class="module-page">
    <CrudQueryCard
      :model="query"
      label-width="auto"
      @search="loadList"
      @reset="resetQuery"
    >
      <el-form-item :label="t('trade.tenantId')">
        <TenantSelect v-model="query.tenantId" class="tenant-select-filter" />
      </el-form-item>

      <el-form-item :label="t('trade.userId')">
        <el-input-number v-model="query.userId" :min="0" :precision="0" />
      </el-form-item>

      <el-form-item :label="t('trade.symbolId')">
        <el-input-number v-model="query.symbolId" :min="0" :precision="0" />
      </el-form-item>

      <el-form-item :label="t('trade.marketType')">
        <el-select v-model="query.marketType" clearable class="query-field">
          <el-option
            v-for="item in marketTypeOptions"
            :key="item.value"
            :label="optionItemLabel(item)"
            :value="item.value"
          />
        </el-select>
      </el-form-item>

      <el-form-item :label="t('trade.status')">
        <el-select v-model="query.status" clearable class="query-field">
          <el-option
            v-for="item in orderStatusOptions"
            :key="item.value"
            :label="optionItemLabel(item)"
            :value="item.value"
          />
        </el-select>
      </el-form-item>

      <el-form-item :label="t('common.keyword')">
        <el-input v-model="query.keyword" clearable class="query-keyword" />
      </el-form-item>
    </CrudQueryCard>

    <el-card shadow="never" class="table-card">
      <el-table v-loading="loading" :data="rows" stripe>
        <el-table-column
          prop="orderNo"
          :label="t('trade.orderNo')"
          min-width="180"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            <div class="order-no-cell">
              <span>{{ row.orderNo || '-' }}</span>
              <span v-if="row.clientOrderId" class="muted">{{ row.clientOrderId }}</span>
            </div>
          </template>
        </el-table-column>

        <el-table-column prop="tenantId" :label="t('trade.tenantId')" width="100" />
        <el-table-column prop="userId" :label="t('trade.userId')" width="100" />
        <el-table-column prop="symbolId" :label="t('trade.symbolId')" width="100" />

        <el-table-column :label="t('trade.marketType')" min-width="130">
          <template #default="{ row }">
            <el-tag size="small" effect="light">
              {{ optionLabel('marketType', row.marketType) }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column :label="t('trade.side')" width="100">
          <template #default="{ row }">
            <el-tag size="small" :type="sideTagType(row.side)" effect="light">
              {{ optionLabel('tradeSide', row.side) }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column :label="t('trade.orderType')" width="120">
          <template #default="{ row }">
            {{ optionLabel('orderType', row.orderType) }}
          </template>
        </el-table-column>

        <el-table-column :label="t('trade.triggerKind')" width="120">
          <template #default="{ row }">
            {{ optionLabel('triggerKind', row.triggerKind) }}
          </template>
        </el-table-column>

        <el-table-column :label="t('trade.price')" min-width="120" align="right">
          <template #default="{ row }">
            {{ displayAmount(row.price) }}
          </template>
        </el-table-column>

        <el-table-column :label="t('trade.qty')" min-width="120" align="right">
          <template #default="{ row }">
            {{ displayAmount(row.qty) }}
          </template>
        </el-table-column>

        <el-table-column :label="t('trade.filled')" min-width="150" align="right">
          <template #default="{ row }">
            <div class="amount-stack">
              <span>{{ displayAmount(row.filledQty) }}</span>
              <span class="muted">{{ displayAmount(row.filledAmount) }}</span>
            </div>
          </template>
        </el-table-column>

        <el-table-column :label="t('trade.status')" width="150">
          <template #default="{ row }">
            <el-tag size="small" :type="orderStatusTagType(row.status)" effect="light">
              {{ optionLabel('orderStatus', row.status) }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column :label="t('trade.updateTimes')" min-width="170">
          <template #default="{ row }">
            {{ formatDate(row.updateTimes) }}
          </template>
        </el-table-column>

        <el-table-column :label="t('common.actions')" width="110" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="showDetail(row)">
              <el-icon><View /></el-icon>
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

    <el-drawer v-model="detailVisible" :title="detailTitle" size="820px">
      <div v-loading="detailLoading">
        <el-empty v-if="!detailData" :description="t('common.noData')" />

        <div v-else class="detail-layout">
          <div class="detail-header">
            <div>
              <div class="detail-title">
                {{ detailData.orderNo || '-' }}
              </div>
              <div class="detail-subtitle">
                {{ detailData.clientOrderId || '-' }}
              </div>
            </div>
            <el-tag :type="orderStatusTagType(detailData.status)" effect="light">
              {{ optionLabel('orderStatus', detailData.status) }}
            </el-tag>
          </div>

          <el-descriptions :column="2" border>
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
            <el-descriptions-item :label="t('trade.marketType')">
              {{ optionLabel('marketType', detailData.marketType) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.positionSide')">
              {{ optionLabel('positionSide', detailData.positionSide) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.side')">
              <el-tag size="small" :type="sideTagType(detailData.side)" effect="light">
                {{ optionLabel('tradeSide', detailData.side) }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.source')">
              {{ optionLabel('sourceType', detailData.source) }}
            </el-descriptions-item>
          </el-descriptions>

          <el-descriptions :title="t('trade.orderParams')" :column="2" border>
            <el-descriptions-item :label="t('trade.orderType')">
              {{ optionLabel('orderType', detailData.orderType) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.triggerKind')">
              {{ optionLabel('triggerKind', detailData.triggerKind) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.timeInForce')">
              {{ optionLabel('timeInForce', detailData.timeInForce) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.price')">
              {{ displayAmount(detailData.price) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.qty')">
              {{ displayAmount(detailData.qty) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.amount')">
              {{ displayAmount(detailData.amount) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.avgPrice')">
              {{ displayAmount(detailData.avgPrice) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.isReduceOnly')">
              {{ yesNoLabel(detailData.isReduceOnly) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.isCloseOnly')">
              {{ yesNoLabel(detailData.isCloseOnly) }}
            </el-descriptions-item>
          </el-descriptions>

          <el-descriptions :title="t('trade.fillInfo')" :column="2" border>
            <el-descriptions-item :label="t('trade.filledQty')">
              {{ displayAmount(detailData.filledQty) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.filledAmount')">
              {{ displayAmount(detailData.filledAmount) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.fillProgress')" :span="2">
              <div class="progress-row">
                <el-progress :percentage="fillProgress" :stroke-width="10" />
                <span>{{ fillProgress }}%</span>
              </div>
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.fee')">
              {{ displayAmount(detailData.fee) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.feeAsset')">
              {{ detailData.feeAsset || '-' }}
            </el-descriptions-item>
          </el-descriptions>

          <el-descriptions :title="t('trade.triggerAndCancel')" :column="2" border>
            <el-descriptions-item :label="t('trade.triggerPrice')">
              {{ displayAmount(detailData.triggerPrice) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.triggerType')">
              {{ optionLabel('triggerType', detailData.triggerType) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.cancelReason')" :span="2">
              {{ detailData.cancelReason || '-' }}
            </el-descriptions-item>
          </el-descriptions>

          <el-descriptions :title="t('trade.timeAndExt')" :column="2" border>
            <el-descriptions-item :label="t('trade.createTimes')">
              {{ formatDate(detailData.createTimes) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.updateTimes')">
              {{ formatDate(detailData.updateTimes) }}
            </el-descriptions-item>
            <el-descriptions-item v-if="detailData.bizExt" :label="t('trade.bizExt')" :span="2">
              <pre class="detail-code">{{ formatJsonText(detailData.bizExt) }}</pre>
            </el-descriptions-item>
          </el-descriptions>
        </div>
      </div>
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { View } from '@element-plus/icons-vue'
import { usePagination } from '@/composables'
import { tradeService, type OptionGroup, type OptionItem, type TradeOrder } from '@/services'
import { formatDate } from '@/utils'
import { findOptionGroup, getOptionLabel } from '@/utils/options'
import TenantSelect from '@/components/TenantSelect.vue'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'

const { t } = useI18n()
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)

type OrderQuery = {
  tenantId?: number
  userId?: number
  symbolId?: number
  marketType?: number
  status?: number
  keyword: string
}

const fallbackOptions: Record<string, OptionItem[]> = {
  marketType: [
    { value: 1, code: 'MARKET_TYPE_SPOT' },
    { value: 2, code: 'MARKET_TYPE_SECONDS_CONTRACT' },
    { value: 3, code: 'MARKET_TYPE_USDT_CONTRACT' },
    { value: 4, code: 'MARKET_TYPE_COIN_CONTRACT' },
  ],
  tradeSide: [
    { value: 1, code: 'TRADE_SIDE_BUY' },
    { value: 2, code: 'TRADE_SIDE_SELL' },
  ],
  positionSide: [
    { value: 1, code: 'POSITION_SIDE_NET' },
    { value: 2, code: 'POSITION_SIDE_LONG' },
    { value: 3, code: 'POSITION_SIDE_SHORT' },
  ],
  orderType: [
    { value: 1, code: 'ORDER_TYPE_LIMIT' },
    { value: 2, code: 'ORDER_TYPE_MARKET' },
  ],
  triggerKind: [
    { value: 0, code: 'TRIGGER_KIND_NONE' },
    { value: 1, code: 'TRIGGER_KIND_CONDITIONAL' },
    { value: 2, code: 'TRIGGER_KIND_TAKE_PROFIT' },
    { value: 3, code: 'TRIGGER_KIND_STOP_LOSS' },
  ],
  timeInForce: [
    { value: 1, code: 'TIME_IN_FORCE_GTC' },
    { value: 2, code: 'TIME_IN_FORCE_IOC' },
    { value: 3, code: 'TIME_IN_FORCE_FOK' },
    { value: 4, code: 'TIME_IN_FORCE_POST_ONLY' },
  ],
  orderStatus: [
    { value: 1, code: 'ORDER_STATUS_PENDING' },
    { value: 2, code: 'ORDER_STATUS_PART_FILLED' },
    { value: 3, code: 'ORDER_STATUS_FILLED' },
    { value: 4, code: 'ORDER_STATUS_CANCELED' },
    { value: 5, code: 'ORDER_STATUS_REJECTED' },
    { value: 6, code: 'ORDER_STATUS_EXPIRED' },
    { value: 7, code: 'ORDER_STATUS_FREEZING' },
    { value: 8, code: 'ORDER_STATUS_TRIGGER_WAITING' },
  ],
  triggerType: [
    { value: 1, code: 'TRIGGER_TYPE_LAST_PRICE' },
    { value: 2, code: 'TRIGGER_TYPE_MARK_PRICE' },
    { value: 3, code: 'TRIGGER_TYPE_INDEX_PRICE' },
  ],
  sourceType: [
    { value: 1, code: 'SOURCE_TYPE_SYSTEM' },
    { value: 2, code: 'SOURCE_TYPE_USER' },
    { value: 3, code: 'SOURCE_TYPE_ADMIN' },
    { value: 4, code: 'SOURCE_TYPE_TASK' },
  ],
}

const loading = ref(false)
const detailLoading = ref(false)
const rows = ref<TradeOrder[]>([])
const detailVisible = ref(false)
const detailData = ref<TradeOrder | null>(null)
const optionGroups = ref<OptionGroup[]>([])

const query = reactive<OrderQuery>({
  tenantId: undefined,
  userId: undefined,
  symbolId: undefined,
  marketType: undefined,
  status: undefined,
  keyword: '',
})

const detailTitle = computed(() => `${t('trade.orders')}${t('option.detail')}`)
const marketTypeOptions = computed(() => optionItems('marketType'))
const orderStatusOptions = computed(() => optionItems('orderStatus'))
const fillProgress = computed(() => calcFillProgress(detailData.value))

const optionItems = (key: string) => {
  const options = findOptionGroup(optionGroups.value, key)
  return options.length ? options : fallbackOptions[key] || []
}

const optionItemLabel = (item: OptionItem) => getOptionLabel(t, item.code, item.value)

const optionLabel = (key: string, value?: number | string) => {
  if (value === undefined || value === null || value === '') return '-'
  const option = optionItems(key).find((item) => String(item.value) === String(value))
  return option ? optionItemLabel(option) : String(value)
}

const loadList = async () => {
  loading.value = true
  try {
    const res = await tradeService.listOrders({
      tenantId: query.tenantId || undefined,
      userId: query.userId || undefined,
      symbolId: query.symbolId || undefined,
      marketType: query.marketType || undefined,
      status: query.status || undefined,
      keyword: query.keyword || undefined,
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
  query.tenantId = undefined
  query.userId = undefined
  query.symbolId = undefined
  query.marketType = undefined
  query.status = undefined
  query.keyword = ''
  resetAndLoad(loadList)
}

const showDetail = async (row: TradeOrder) => {
  detailVisible.value = true
  detailLoading.value = true
  detailData.value = row
  try {
    const res = await tradeService.getOrder({ tenantId: row.tenantId, id: row.id })
    detailData.value = res.data || row
  } finally {
    detailLoading.value = false
  }
}

const loadOptions = async () => {
  optionGroups.value = (await tradeService.getOptions()).data || []
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

function sideTagType(side: number) {
  if (side === 1) return 'success'
  if (side === 2) return 'danger'
  return 'info'
}

function orderStatusTagType(status: number) {
  if (status === 3) return 'success'
  if (status === 2) return 'warning'
  if (status === 4 || status === 6) return 'info'
  if (status === 5) return 'danger'
  if (status === 7 || status === 8) return 'warning'
  return ''
}

function displayAmount(value?: string | number) {
  if (value === undefined || value === null || value === '') return '-'
  return String(value)
}

function yesNoLabel(value?: number) {
  if (value === undefined || value === null) return '-'
  return value === 1 ? t('users.yes') : t('users.no')
}

function calcFillProgress(order: TradeOrder | null) {
  if (!order) return 0
  const filledQty = Number(order.filledQty || 0)
  const qty = Number(order.qty || 0)
  if (qty > 0) return Math.min(100, Math.round((filledQty / qty) * 100))

  const filledAmount = Number(order.filledAmount || 0)
  const amount = Number(order.amount || 0)
  if (amount > 0) return Math.min(100, Math.round((filledAmount / amount) * 100))
  return 0
}

function formatJsonText(value: string) {
  try {
    return JSON.stringify(JSON.parse(value), null, 2)
  } catch {
    return value
  }
}

onMounted(async () => {
  await loadOptions()
  await loadList()
})
</script>

<style scoped>
.query-field {
  width: 160px;
}

.query-keyword {
  width: 220px;
}

.order-no-cell,
.amount-stack {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.muted {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.detail-layout {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.detail-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding-bottom: 4px;
}

.detail-title {
  color: var(--el-text-color-primary);
  font-size: 18px;
  font-weight: 600;
  line-height: 1.4;
}

.detail-subtitle {
  margin-top: 4px;
  color: var(--el-text-color-secondary);
  font-size: 13px;
}

.progress-row {
  display: grid;
  grid-template-columns: 1fr auto;
  align-items: center;
  gap: 12px;
  width: 100%;
}

.detail-code {
  margin: 0;
  max-height: 260px;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-word;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono', monospace;
  font-size: 12px;
  line-height: 1.5;
  color: #334155;
}
</style>
