<template>
  <div class="module-page">
    <CrudQueryCard :model="query" @search="loadList" @reset="resetQuery">
      <el-form-item :label="t('trade.tenantId')">
        <TenantSelect v-model="query.tenantId" class="tenant-select-filter" />
      </el-form-item>

      <el-form-item :label="t('trade.userId')">
        <UserSelect v-model="query.userId" :tenant-id="query.tenantId || undefined" />
      </el-form-item>

      <el-form-item :label="t('trade.symbolId')">
        <SymbolSelect
          v-model="query.symbolId"
          :tenant-id="query.tenantId || undefined"
          :product-type="query.productType || undefined"
        />
      </el-form-item>

      <el-form-item :label="t('trade.productType')">
        <el-select v-model="query.productType" clearable class="query-field">
          <el-option
            v-for="item in productTypeOptions"
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

        <el-table-column
          :label="`${t('trade.tenantId')}/${t('trade.userId')}/${t('trade.symbolId')}`"
          min-width="200"
          align="center"
        >
          <template #default="{ row }">
            {{ row.tenantId }} / {{ row.userId }} / {{ row.symbolId }}
          </template>
        </el-table-column>

        <el-table-column :label="t('trade.productType')" min-width="100">
          <template #default="{ row }">
            <el-tag size="small" effect="light">
              {{ optionLabel('productType', row.productType) }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column :label="t('trade.side')" width="100">
          <template #default="{ row }">
            <el-tag size="small" :type="orderDirectionTagType(row)" effect="light">
              {{ orderDirectionLabel(row) }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column :label="t('trade.orderType')" width="120" align="center">
          <template #default="{ row }">
            {{ orderTypeLabel(row) }}
          </template>
        </el-table-column>

        <el-table-column :label="t('trade.triggerKind')" width="120" align="center">
          <template #default="{ row }">
            {{ optionLabel('triggerKind', row.triggerKind) }}
          </template>
        </el-table-column>

        <el-table-column :label="t('trade.price')" min-width="120" align="right">
          <template #default="{ row }">
            {{ displayOrderPrice(row) }}
          </template>
        </el-table-column>

        <el-table-column :label="t('trade.qty')" min-width="120" align="right">
          <template #default="{ row }">
            {{ displayOrderQty(row) }}
          </template>
        </el-table-column>

        <el-table-column
          v-if="hasMatchableOrders"
          :label="t('trade.filledQtyAmount')"
          min-width="150"
          align="center"
        >
          <template #default="{ row }">
            <span v-if="row.productType === 3" class="muted">-</span>
            <span v-else>
              {{ displayAmount(row.filledQty) }} / {{ displayAmount(row.filledAmount) }}
            </span>
          </template>
        </el-table-column>

        <el-table-column :label="t('trade.status')" width="150">
          <template #default="{ row }">
            <el-tag
              size="small"
              :type="orderDisplayStatusTagType(effectiveOrderDisplayStatus(row))"
              effect="light"
            >
              {{ orderDisplayStatusLabel(effectiveOrderDisplayStatus(row)) }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column :label="t('trade.updateTimes')" min-width="170">
          <template #default="{ row }">
            {{ formatDate(row.updateTimes) }}
          </template>
        </el-table-column>

        <el-table-column
          :label="t('common.actions')"
          align="center"
          width="110"
          fixed="right"
        >
          <template #default="{ row }">
            <el-button
              v-perm="'trade:order:detail'"
              link
              type="primary"
              @click="showDetail(row)"
            >
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
            <el-tag
              :type="orderDisplayStatusTagType(effectiveOrderDisplayStatus(detailData))"
              effect="light"
            >
              {{ orderDisplayStatusLabel(effectiveOrderDisplayStatus(detailData)) }}
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
            <el-descriptions-item :label="t('trade.productType')">
              {{ optionLabel('productType', detailData.productType) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.positionSide')">
              {{ optionLabel('positionSide', detailData.positionSide) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.side')">
              <el-tag size="small" :type="orderDirectionTagType(detailData)" effect="light">
                {{ orderDirectionLabel(detailData) }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.source')">
              {{ optionLabel('orderSourceType', detailData.source) }}
            </el-descriptions-item>
          </el-descriptions>

          <el-descriptions
            v-if="detailSpot"
            :title="t('trade.spotOrderDetails')"
            :column="2"
            border
          >
            <el-descriptions-item :label="t('trade.frozenAsset')">
              {{ detailSpot.frozenAsset || '-' }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.frozenAmount')">
              {{ displayAmount(detailSpot.frozenAmount) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.settleAsset')">
              {{ detailSpot.settleAsset || '-' }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.settleAmount')">
              {{ displayAmount(detailSpot.settleAmount) }}
            </el-descriptions-item>
          </el-descriptions>

          <el-descriptions
            v-if="detailContract"
            :title="
              detailData.contractType === 1
                ? t('trade.perpetualOrderDetails')
                : t('trade.deliveryOrderDetails')
            "
            :column="2"
            border
          >
            <el-descriptions-item :label="t('trade.marginMode')">
              {{ optionLabel('marginMode', detailContract.marginMode) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.leverage')">
              {{ detailContract.leverage }}x
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.marginAsset')">
              {{ detailContract.marginAsset || '-' }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.marginAmount')">
              {{ displayAmount(detailContract.marginAmount) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.closePositionType')">
              {{ closePositionTypeLabel(detailContract.closePositionType) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.reservedCloseQty')">
              {{ displayAmount(detailContract.reservedCloseQty) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.liquidationPrice')">
              {{ displayAmount(detailContract.liquidationPrice) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.riskPrice')">
              {{ displayAmount(detailContract.riskPrice) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.takeProfitPrice')">
              {{ displayAmount(detailContract.takeProfitPrice) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.stopLossPrice')">
              {{ displayAmount(detailContract.stopLossPrice) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.riskTierId')" :span="2">
              {{ detailContract.riskTierId || '-' }}
            </el-descriptions-item>
          </el-descriptions>

          <el-descriptions
            v-if="detailSeconds"
            :title="t('trade.secondsDetails')"
            :column="2"
            border
          >
            <el-descriptions-item :label="t('trade.side')">
              {{ secondsDirectionLabel(detailSeconds.direction) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.durationSeconds')">
              {{ detailSeconds.durationSeconds }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.stakeAmount')">
              {{ displayAmount(detailSeconds.stakeAmount) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.stakeAsset')">
              {{ detailSeconds.stakeAsset || '-' }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.payoutRate')">
              {{ displayAmount(detailSeconds.payoutRate) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.secondsFeeRate')">
              {{ displayAmount(detailSeconds.feeRate) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.secondsSettlementStatus')">
              {{ secondsSettlementStatusLabel(detailSeconds.settlementStatus) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.secondsResult')">
              {{ secondsResultLabel(detailSeconds.result) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.startPrice')">
              {{ displayAmount(detailSeconds.startPrice) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.startPriceTime')">
              {{ formatTimestamp(detailSeconds.startPriceTime) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.startPriceSource')" :span="2">
              {{ detailSeconds.startPriceSource || '-' }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.frozenAt')">
              {{ formatTimestamp(detailSeconds.frozenAt) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.activatedAt')">
              {{ formatTimestamp(detailSeconds.activatedAt) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.expireTime')">
              {{ formatTimestamp(detailSeconds.expireTime) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.settledAt')">
              {{ formatTimestamp(detailSeconds.settledAt) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.settlementPrice')">
              {{ displayAmount(detailSeconds.settlementPrice) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.settlementPriceTime')">
              {{ formatTimestamp(detailSeconds.settlementPriceTime) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.settlementPriceSource')" :span="2">
              {{ detailSeconds.settlementPriceSource || '-' }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.profitAmount')">
              {{ displayAmount(detailSeconds.profitAmount) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.feeAmount')">
              {{ displayAmount(detailSeconds.feeAmount) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.returnAmount')">
              {{ displayAmount(detailSeconds.returnAmount) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.priceAlgorithm')">
              {{ detailSeconds.priceAlgorithm || '-' }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.reservationNo')" :span="2">
              {{ detailSeconds.reservationNo || '-' }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.settlementReason')" :span="2">
              {{ detailSeconds.settlementReason || '-' }}
            </el-descriptions-item>
          </el-descriptions>

          <el-descriptions
            v-else
            :title="t('trade.orderParams')"
            :column="2"
            border
          >
            <el-descriptions-item :label="t('trade.orderType')">
              {{ orderTypeLabel(detailData) }}
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
            <el-descriptions-item :label="t('trade.isClosePosition')">
              {{ yesNoLabel(detailData.isClosePosition) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.canceledQty')">
              {{ displayAmount(detailData.canceledQty) }}
            </el-descriptions-item>
          </el-descriptions>

          <el-descriptions
            v-if="!detailSeconds"
            :title="t('trade.fillInfo')"
            :column="2"
            border
          >
            <el-descriptions-item :label="t('trade.filledQty')">
              {{ displayAmount(detailData.filledQty) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.filledAmount')">
              {{ displayAmount(detailData.filledAmount) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.fillProgress')" :span="2">
              <div class="progress-row">
                <el-progress :percentage="fillProgress" :stroke-width="10" :show-text="false" />
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

          <el-descriptions
            v-if="!detailSeconds"
            :title="t('trade.triggerAndCancel')"
            :column="2"
            border
          >
            <el-descriptions-item :label="t('trade.triggerPrice')">
              {{ displayAmount(detailData.triggerPrice) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.triggerType')">
              {{ optionLabel('triggerType', detailData.triggerType) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.cancelReason')" :span="2">
              {{ detailData.cancelReason || '-' }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.ocoGroupNo')">
              {{ detailData.ocoGroupNo || '-' }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.expireAt')">
              {{ formatTimestamp(detailData.expireAt) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.triggeredAt')">
              {{ formatTimestamp(detailData.triggeredAt) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.completionReason')">
              {{ detailData.completionReason || '-' }}
            </el-descriptions-item>
          </el-descriptions>

          <el-descriptions :title="t('trade.timeAndExt')" :column="2" border>
            <el-descriptions-item :label="t('trade.createTimes')">
              {{ formatDate(detailData.createTimes) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.updateTimes')">
              {{ formatDate(detailData.updateTimes) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.version')">
              {{ detailData.version }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.requestHash')">
              {{ detailData.requestHash || '-' }}
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
import Decimal from 'decimal.js'
import { usePagination } from '@/composables'
import {
  tradeService,
  type OptionGroup,
  type OptionItem,
  type TradeOrder,
  type TradeOrderContract,
  type TradeOrderSeconds,
  type TradeOrderSpot,
} from '@/services'
import { formatDate } from '@/utils'
import { findOptionGroup, getOptionLabel } from '@/utils/options'
import TenantSelect from '@/components/TenantSelect.vue'
import UserSelect from '@/components/UserSelect.vue'
import SymbolSelect from '@/components/SymbolSelect.vue'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'

const { t } = useI18n()
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)

type OrderQuery = {
  tenantId?: number
  userId?: number
  symbolId?: number
  productType?: number
  status?: number
  keyword: string
}

const fallbackOptions: Record<string, OptionItem[]> = {
  productType: [
    { value: 1, code: 'PRODUCT_TYPE_SPOT' },
    { value: 2, code: 'PRODUCT_TYPE_DERIVATIVE' },
    { value: 3, code: 'PRODUCT_TYPE_SECONDS' },
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
    { value: 9, code: 'ORDER_STATUS_CANCELING' },
    { value: 10, code: 'ORDER_STATUS_EXPIRING' },
    { value: 11, code: 'ORDER_STATUS_SETTLEMENT_PENDING' },
  ],
  orderDisplayStatus: [
    { value: 1, code: 'ORDER_DISPLAY_STATUS_FREEZING' },
    { value: 2, code: 'ORDER_DISPLAY_STATUS_ACTIVATING' },
    { value: 3, code: 'ORDER_DISPLAY_STATUS_ACTIVE' },
    { value: 4, code: 'ORDER_DISPLAY_STATUS_TRIGGER_WAITING' },
    { value: 5, code: 'ORDER_DISPLAY_STATUS_PENDING' },
    { value: 6, code: 'ORDER_DISPLAY_STATUS_PART_FILLED' },
    { value: 7, code: 'ORDER_DISPLAY_STATUS_SETTLING' },
    { value: 8, code: 'ORDER_DISPLAY_STATUS_FILLED' },
    { value: 9, code: 'ORDER_DISPLAY_STATUS_SETTLED' },
    { value: 10, code: 'ORDER_DISPLAY_STATUS_CANCELING' },
    { value: 11, code: 'ORDER_DISPLAY_STATUS_CANCELED' },
    { value: 12, code: 'ORDER_DISPLAY_STATUS_EXPIRING' },
    { value: 13, code: 'ORDER_DISPLAY_STATUS_EXPIRED' },
    { value: 14, code: 'ORDER_DISPLAY_STATUS_REFUNDING' },
    { value: 15, code: 'ORDER_DISPLAY_STATUS_REFUNDED' },
    { value: 16, code: 'ORDER_DISPLAY_STATUS_REJECTED' },
    { value: 17, code: 'ORDER_DISPLAY_STATUS_MANUAL_REVIEW' },
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
  orderSourceType: [
    { value: 1, code: 'ORDER_SOURCE_TYPE_APP' },
    { value: 2, code: 'ORDER_SOURCE_TYPE_WEB' },
    { value: 3, code: 'ORDER_SOURCE_TYPE_API' },
    { value: 4, code: 'ORDER_SOURCE_TYPE_SYSTEM' },
    { value: 5, code: 'ORDER_SOURCE_TYPE_LIQUIDITY' },
  ],
}

const loading = ref(false)
const detailLoading = ref(false)
const rows = ref<TradeOrder[]>([])
const detailVisible = ref(false)
const detailData = ref<TradeOrder | null>(null)
const detailSpot = ref<TradeOrderSpot | null>(null)
const detailContract = ref<TradeOrderContract | null>(null)
const detailSeconds = ref<TradeOrderSeconds | null>(null)
const optionGroups = ref<OptionGroup[]>([])

const query = reactive<OrderQuery>({
  tenantId: undefined,
  userId: undefined,
  symbolId: undefined,
  productType: undefined,
  status: undefined,
  keyword: '',
})

const detailTitle = computed(() => `${t('trade.orders')}${t('option.detail')}`)
const productTypeOptions = computed(() => optionItems('productType'))
const orderStatusOptions = computed(() => optionItems('orderStatus'))
const fillProgress = computed(() => calcFillProgress(detailData.value))
const hasMatchableOrders = computed(() =>
  rows.value.some(
    (order) =>
      order.productType !== 3 &&
      (isPositiveAmount(order.filledQty) || isPositiveAmount(order.filledAmount)),
  ),
)

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
      productType: query.productType || undefined,
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
  query.productType = undefined
  query.status = undefined
  query.keyword = ''
  resetAndLoad(loadList)
}

const showDetail = async (row: TradeOrder) => {
  detailVisible.value = true
  detailLoading.value = true
  detailData.value = row
  detailSpot.value = null
  detailContract.value = null
  detailSeconds.value = null
  try {
    const res = await tradeService.getOrder({ tenantId: row.tenantId, id: row.id })
    detailData.value = res.data?.order || row
    detailSpot.value = res.data?.spot?.id ? res.data.spot : null
    detailContract.value = res.data?.contract?.id ? res.data.contract : null
    detailSeconds.value = res.data?.seconds?.id ? res.data.seconds : null
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

function orderDirectionLabel(order: TradeOrder) {
  if (order.productType === 3) {
    if (order.secondsDirection === 1) return t('trade.secondsUp')
    if (order.secondsDirection === 2) return t('trade.secondsDown')
    return '-'
  }
  if (order.side === 1) return t('options.SIDE_BUY')
  if (order.side === 2) return t('options.SIDE_SELL')
  return '-'
}

function secondsDirectionLabel(direction: number) {
  if (direction === 1) return t('trade.secondsUp')
  if (direction === 2) return t('trade.secondsDown')
  return '-'
}

function closePositionTypeLabel(value: number) {
  if (value === 0) return t('trade.closePositionNormal')
  if (value === 1) return t('trade.closePositionLong')
  if (value === 2) return t('trade.closePositionShort')
  return String(value)
}

function secondsSettlementStatusLabel(status: number) {
  const labels: Record<number, string> = {
    0: t('trade.secondsPendingFreeze'),
    1: t('trade.secondsActivating'),
    2: t('trade.secondsActive'),
    3: t('trade.secondsSettling'),
    4: t('trade.secondsSettled'),
    5: t('trade.secondsRefunding'),
    6: t('trade.secondsRefunded'),
    7: t('trade.secondsManualReview'),
  }
  return labels[status] || String(status)
}

function secondsResultLabel(result: number) {
  const labels: Record<number, string> = {
    0: t('trade.secondsResultPending'),
    1: t('trade.secondsResultWin'),
    2: t('trade.secondsResultLose'),
    3: t('trade.secondsResultDraw'),
    4: t('trade.secondsResultVoid'),
  }
  return labels[result] || String(result)
}

function formatTimestamp(value?: number) {
  return value && value > 0 ? formatDate(value) : '-'
}

function orderDirectionTagType(order: TradeOrder) {
  return order.productType === 3 ? sideTagType(order.secondsDirection) : sideTagType(order.side)
}

function orderTypeLabel(order: TradeOrder) {
  return order.productType === 3
    ? t('trade.secondsOrder')
    : optionLabel('orderType', order.orderType)
}

function orderDisplayStatusLabel(status: number) {
  return optionLabel('orderDisplayStatus', status)
}

function effectiveOrderDisplayStatus(order: TradeOrder) {
  // 撤销的是未成交的剩余量；已有成交记录时不能展示为“已撤单”。
  if (
    order.productType !== 3 &&
    order.status === 4 &&
    (isPositiveAmount(order.filledQty) || isPositiveAmount(order.filledAmount))
  ) {
    return 6
  }

  if (order.displayStatus > 0) return order.displayStatus

  // 兼容滚动发布期间尚未返回 displayStatus 的旧版 RPC/API。
  if (order.productType === 3) {
    if (order.status === 3) return 9
    if (order.status === 4) return 15
    if (order.status === 5) return 16
  }

  const legacyMapping: Record<number, number> = {
    1: 5,
    2: 6,
    3: 8,
    4: 11,
    5: 16,
    6: 13,
    7: 1,
    8: 4,
    9: 10,
    10: 12,
    11: 7,
  }
  return legacyMapping[order.status] || 0
}

function orderDisplayStatusTagType(status: number) {
  if (status === 8 || status === 9) return 'success'
  if (status === 11 || status === 13 || status === 15) return 'info'
  if (status === 16) return 'danger'
  if ([1, 2, 3, 4, 5, 6, 7, 10, 12, 14, 17].includes(status)) return 'warning'
  return ''
}

function displayAmount(value?: string | number) {
  if (value === undefined || value === null || value === '') return '-'
  return String(value)
}

function displayOrderPrice(order: TradeOrder) {
  if (order.orderType !== 2) return displayAmount(order.price)
  if (isPositiveAmount(order.avgPrice)) return displayAmount(order.avgPrice)
  return optionLabel('orderType', 2)
}

function displayOrderQty(order: TradeOrder) {
  if (isPositiveAmount(order.qty)) return displayAmount(order.qty)
  if (isPositiveAmount(order.filledQty)) return displayAmount(order.filledQty)
  return '-'
}

function isPositiveAmount(value?: string | number) {
  try {
    const amount = new Decimal(value || 0)
    return amount.isFinite() && amount.isPositive()
  } catch {
    return false
  }
}

function yesNoLabel(value?: number) {
  if (value === undefined || value === null) return '-'
  return value === 1 ? t('users.yes') : t('users.no')
}

function calcFillProgress(order: TradeOrder | null) {
  if (!order) return 0

  const calculate = (
    filled: string | number | undefined,
    total: string | number | undefined,
  ): number | null => {
    try {
      const totalValue = new Decimal(total || 0)
      const filledValue = new Decimal(filled || 0)
      if (!totalValue.isFinite() || !totalValue.isPositive() || !filledValue.isFinite()) return null

      const percentage = Decimal.max(0, Decimal.min(filledValue.div(totalValue).mul(100), 100))
        .toDecimalPlaces(0, Decimal.ROUND_HALF_UP)
        .toNumber()
      return Number.isFinite(percentage) ? percentage : null
    } catch {
      return null
    }
  }

  return calculate(order.filledQty, order.qty) ?? calculate(order.filledAmount, order.amount) ?? 0
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
