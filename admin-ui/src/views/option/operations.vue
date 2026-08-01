<template>
  <div class="module-page">
    <CrudQueryCard :model="query" @search="refreshAll" @reset="resetQuery">
      <el-form-item :label="t('option.tenantId')">
        <TenantSelect v-model="query.tenantId" class="tenant-select-filter" />
      </el-form-item>
      <el-form-item :label="t('option.riskStaleThreshold')">
        <el-input-number
          v-model="query.riskStaleSeconds"
          :min="10"
          :max="300"
          :step="10"
        />
      </el-form-item>
      <el-form-item :label="t('option.comboStaleThreshold')">
        <el-input-number
          v-model="query.comboStaleSeconds"
          :min="10"
          :max="300"
          :step="10"
        />
      </el-form-item>
      <el-form-item :label="t('option.bizNo')">
        <el-input v-model="query.bizNo" clearable />
      </el-form-item>
      <el-form-item :label="t('option.comboNo')">
        <el-input v-model="query.comboNo" clearable />
      </el-form-item>
      <el-form-item :label="t('option.comboStatus')">
        <el-select v-model="query.comboStatus" clearable style="width: 160px">
          <el-option
            v-for="status in comboStatuses"
            :key="status.value"
            :label="status.label"
            :value="status.value"
          />
        </el-select>
      </el-form-item>
      <el-form-item :label="t('option.instructionStatus')">
        <el-select v-model="query.instructionStatus" clearable style="width: 150px">
          <el-option :label="t('option.pending')" :value="1" />
          <el-option :label="t('option.failed')" :value="4" />
          <el-option :label="t('option.manualReview')" :value="5" />
        </el-select>
      </el-form-item>
    </CrudQueryCard>

    <el-alert
      v-if="!query.tenantId"
      :title="t('option.selectTenantForOperations')"
      type="info"
      :closable="false"
    />

    <template v-else>
      <el-card v-loading="loading" shadow="never" class="table-card">
        <template #header>
          <div class="card-header">
            <span>{{ t('option.operationsOverview') }}</span>
            <span class="generated-at">{{ formatTime(overview?.generatedAt) }}</span>
          </div>
        </template>
        <div class="metric-grid">
          <div v-for="metric in metrics" :key="metric.label" class="metric">
            <span class="metric-label">{{ metric.label }}</span>
            <strong :class="{ danger: metric.danger }">{{ metric.value }}</strong>
            <small>{{ formatOldest(metric.oldest) }}</small>
          </div>
        </div>
        <el-divider>{{ t('option.insuranceAndBackstop') }}</el-divider>
        <el-descriptions :column="3" border>
          <el-descriptions-item :label="t('option.insuranceLedger')">
            {{ formatCoinAmounts(overview?.insuranceLedger) }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.backstopLiability')">
            {{ formatCoinAmounts(overview?.backstopLiability) }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.unresolvedDeficit')">
            {{ formatCoinAmounts(overview?.unresolvedDeficit) }}
          </el-descriptions-item>
        </el-descriptions>
      </el-card>

      <el-card v-loading="loading" shadow="never" class="table-card">
        <template #header>
          {{ t('option.comboOrderWorkbench') }}
        </template>
        <el-table :data="comboOrders" stripe>
          <el-table-column prop="comboOrder.comboNo" :label="t('option.comboNo')" min-width="210" />
          <el-table-column prop="comboOrder.userId" :label="t('option.userId')" width="100" />
          <el-table-column prop="comboOrder.accountId" :label="t('option.accountId')" width="110" />
          <el-table-column
            prop="comboOrder.underlyingSymbol"
            :label="t('option.underlyingSymbol')"
            min-width="130"
          />
          <el-table-column prop="comboOrder.netPrice" :label="t('option.netPrice')" width="120" />
          <el-table-column prop="comboOrder.qty" :label="t('option.quantity')" width="110" />
          <el-table-column prop="comboOrder.filledQty" :label="t('option.filledQty')" width="110" />
          <el-table-column :label="t('option.comboLegs')" min-width="300">
            <template #default="{ row }">
              {{ formatComboLegs(row) }}
            </template>
          </el-table-column>
          <el-table-column :label="t('common.status')" width="120">
            <template #default="{ row }">
              <el-tag :type="comboStatusType(row.comboOrder.status)">
                {{ comboStatusLabel(row.comboOrder.status) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column :label="t('common.actions')" width="150" fixed="right">
            <template #default="{ row }">
              <el-button
                v-perm="'option:operations:combo-view'"
                link
                type="primary"
                @click="openComboDetail(row)"
              >
                {{ t('option.detail') }}
              </el-button>
              <el-button
                v-if="[1, 2, 3].includes(row.comboOrder.status)"
                v-perm="'option:operations:combo-cancel'"
                link
                type="danger"
                @click="forceCancelCombo(row)"
              >
                {{ t('option.forceCancelCombo') }}
              </el-button>
            </template>
          </el-table-column>
        </el-table>
        <CursorPagination
          v-model:limit="comboPage.limit"
          :total="comboPage.total"
          :has-prev="comboPage.hasPrev"
          :has-next="comboPage.hasNext"
          @prev="comboPrev"
          @next="comboNext"
          @limit-change="comboReset"
        />
      </el-card>

      <el-card v-loading="loading" shadow="never" class="table-card">
        <template #header>
          {{ t('option.assetInstructionWorkbench') }}
        </template>
        <el-table :data="instructions" stripe>
          <el-table-column
            prop="instructionNo"
            :label="t('option.instructionNo')"
            min-width="220"
          />
          <el-table-column prop="bizNo" :label="t('option.bizNo')" min-width="180" />
          <el-table-column prop="userId" :label="t('option.userId')" width="100" />
          <el-table-column prop="coin" :label="t('option.coin')" width="90" />
          <el-table-column prop="amount" :label="t('option.amount')" min-width="130" />
          <el-table-column prop="stepNo" :label="t('option.stepNo')" width="80" />
          <el-table-column prop="status" :label="t('common.status')" width="90" />
          <el-table-column prop="retryCount" :label="t('option.retryCount')" width="90" />
          <el-table-column
            prop="lastErrorMsg"
            :label="t('option.lastErrorMsg')"
            min-width="220"
            show-overflow-tooltip
          />
          <el-table-column :label="t('common.actions')" width="100" fixed="right">
            <template #default="{ row }">
              <el-button
                v-if="[4, 5].includes(row.status) && !row.deliveryUnitId"
                v-perm="'option:operations:asset-retry'"
                link
                type="warning"
                @click="retryInstruction(row)"
              >
                {{ t('option.retry') }}
              </el-button>
            </template>
          </el-table-column>
        </el-table>
        <CursorPagination
          v-model:limit="instructionPage.limit"
          :total="instructionPage.total"
          :has-prev="instructionPage.hasPrev"
          :has-next="instructionPage.hasNext"
          @prev="instructionPrev"
          @next="instructionNext"
          @limit-change="instructionReset"
        />
      </el-card>

      <el-card v-loading="loading" shadow="never" class="table-card">
        <template #header>
          {{ t('option.reconciliationWorkbench') }}
        </template>
        <el-table :data="issues" stripe>
          <el-table-column prop="issueKey" :label="t('option.issueKey')" min-width="220" />
          <el-table-column prop="bizNo" :label="t('option.bizNo')" min-width="170" />
          <el-table-column prop="checkType" :label="t('option.checkType')" width="100" />
          <el-table-column prop="status" :label="t('common.status')" width="90" />
          <el-table-column
            prop="occurrenceCount"
            :label="t('option.occurrenceCount')"
            width="100"
          />
          <el-table-column
            prop="expectedValue"
            :label="t('option.expectedValue')"
            min-width="180"
            show-overflow-tooltip
          />
          <el-table-column
            prop="actualValue"
            :label="t('option.actualValue')"
            min-width="180"
            show-overflow-tooltip
          />
          <el-table-column
            prop="detail"
            :label="t('option.detail')"
            min-width="240"
            show-overflow-tooltip
          />
        </el-table>
        <CursorPagination
          v-model:limit="issuePage.limit"
          :total="issuePage.total"
          :has-prev="issuePage.hasPrev"
          :has-next="issuePage.hasNext"
          @prev="issuePrev"
          @next="issueNext"
          @limit-change="issueReset"
        />
      </el-card>
    </template>

    <el-drawer
      v-model="comboDrawerVisible"
      :title="t('option.comboOrderDetail')"
      size="80%"
      destroy-on-close
    >
      <template v-if="comboDetail">
        <el-alert
          v-if="comboDetail.dataTruncated"
          :title="t('option.comboDetailTruncated')"
          type="warning"
          :closable="false"
          show-icon
        />
        <el-descriptions :column="3" border class="detail-section">
          <el-descriptions-item :label="t('option.comboNo')">
            {{ comboDetail.comboOrder.comboNo }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.userId')">
            {{ comboDetail.comboOrder.userId }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.accountId')">
            {{ comboDetail.comboOrder.accountId }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.netPrice')">
            {{ comboDetail.comboOrder.netPrice }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.quantity')">
            {{ comboDetail.comboOrder.qty }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('common.status')">
            {{ comboStatusLabel(comboDetail.comboOrder.status) }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.cancelReason')" :span="3">
            {{ comboDetail.comboOrder.cancelReason || '-' }}
          </el-descriptions-item>
        </el-descriptions>

        <el-divider>{{ t('option.comboLegs') }}</el-divider>
        <el-table :data="comboDetail.legs" stripe>
          <el-table-column prop="legNo" :label="t('option.legNo')" width="80" />
          <el-table-column prop="contractId" :label="t('option.contractId')" width="120" />
          <el-table-column :label="t('option.side')" width="90">
            <template #default="{ row }">{{ row.side === 1 ? 'BUY' : 'SELL' }}</template>
          </el-table-column>
          <el-table-column prop="ratio" :label="t('option.ratio')" width="80" />
          <el-table-column prop="price" :label="t('option.price')" />
          <el-table-column prop="qty" :label="t('option.quantity')" />
          <el-table-column prop="filledQty" :label="t('option.filledQty')" />
          <el-table-column prop="childOrderId" :label="t('option.childOrderId')" min-width="120" />
        </el-table>

        <el-divider>{{ t('option.shadowOrders') }}</el-divider>
        <el-table :data="comboDetail.childOrders" stripe>
          <el-table-column prop="order.orderNo" :label="t('option.orderNo')" min-width="210" />
          <el-table-column prop="order.contractId" :label="t('option.contractId')" width="120" />
          <el-table-column prop="order.marginAmount" :label="t('option.marginAmount')" width="140" />
          <el-table-column prop="order.marginCoin" :label="t('option.coin')" width="100" />
          <el-table-column prop="order.status" :label="t('common.status')" width="90" />
          <el-table-column prop="order.cancelReason" :label="t('option.cancelReason')" min-width="180" />
        </el-table>

        <el-divider>
          {{ t('option.comboTrades') }} ({{ comboDetail.tradeTotal }})
        </el-divider>
        <el-table :data="comboDetail.trades" stripe>
          <el-table-column prop="trade.comboMatchNo" :label="t('option.comboMatchNo')" min-width="210" />
          <el-table-column prop="trade.comboLegNo" :label="t('option.legNo')" width="80" />
          <el-table-column prop="trade.contractId" :label="t('option.contractId')" width="120" />
          <el-table-column prop="trade.price" :label="t('option.price')" width="120" />
          <el-table-column prop="trade.qty" :label="t('option.quantity')" width="120" />
          <el-table-column prop="trade.tradeTime" :label="t('option.tradeTime')" min-width="170">
            <template #default="{ row }">{{ formatTime(row.trade.tradeTime) }}</template>
          </el-table-column>
        </el-table>

        <el-divider>
          {{ t('option.assetInstructions') }} ({{ comboDetail.assetInstructionTotal }})
        </el-divider>
        <el-table :data="comboDetail.assetInstructions" stripe>
          <el-table-column prop="instructionNo" :label="t('option.instructionNo')" min-width="220" />
          <el-table-column prop="orderId" :label="t('option.childOrderId')" width="120" />
          <el-table-column prop="action" :label="t('option.action')" width="90" />
          <el-table-column prop="coin" :label="t('option.coin')" width="90" />
          <el-table-column prop="amount" :label="t('option.amount')" width="130" />
          <el-table-column prop="status" :label="t('common.status')" width="90" />
          <el-table-column prop="lastErrorMsg" :label="t('option.lastErrorMsg')" min-width="220" />
        </el-table>
      </template>
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { usePagination } from '@/composables'
import {
  optionService,
  type OptionAssetInstruction,
  type OptionAdminComboOrderDetail,
  type OptionComboOrderDetail,
  type OptionCoinAmount,
  type OptionOperationsOverview,
  type OptionReconciliationIssue,
} from '@/services'
import TenantSelect from '@/components/TenantSelect.vue'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'

const { t } = useI18n()
const loading = ref(false)
const overview = ref<OptionOperationsOverview>()
const instructions = ref<OptionAssetInstruction[]>([])
const issues = ref<OptionReconciliationIssue[]>([])
const comboOrders = ref<OptionComboOrderDetail[]>([])
const comboDetail = ref<OptionAdminComboOrderDetail>()
const comboDrawerVisible = ref(false)
const query = reactive({
  tenantId: undefined as number | undefined,
  riskStaleSeconds: 60,
  comboStaleSeconds: 60,
  bizNo: '',
  comboNo: '',
  comboStatus: undefined as number | undefined,
  instructionStatus: undefined as number | undefined,
})
const instructionPagination = usePagination<number>(20)
const issuePagination = usePagination<number>(20)
const comboPagination = usePagination<number>(20)
const instructionPage = instructionPagination.pagination
const issuePage = issuePagination.pagination
const comboPage = comboPagination.pagination

const comboStatuses = computed(() =>
  [1, 2, 3, 4, 5, 6, 7, 8].map((value) => ({
    value,
    label: comboStatusLabel(value),
  })),
)

const metrics = computed(() => {
  const value = overview.value
  if (!value) return []
  return [
    {
      label: t('option.assetExceptions'),
      value: value.assetFailedCount + value.assetManualReviewCount,
      oldest: value.oldestAssetInstructionTime,
      danger: value.assetFailedCount + value.assetManualReviewCount > 0,
    },
    {
      label: t('option.openReconciliation'),
      value: value.openReconciliationCount,
      oldest: value.oldestReconciliationTime,
      danger: value.openReconciliationCount > 0,
    },
    {
      label: t('option.pendingSettlementPrices'),
      value: value.pendingSettlementPriceCount,
      oldest: 0,
      danger: value.pendingSettlementPriceCount > 0,
    },
    {
      label: t('option.staleRiskAccounts'),
      value: value.staleRiskAccountCount,
      oldest: value.oldestRiskCalcTime,
      danger: value.staleRiskAccountCount > 0,
    },
    {
      label: t('option.exerciseBacklog'),
      value: value.pendingExerciseCount,
      oldest: value.oldestExerciseTime,
      danger: value.pendingExerciseCount > 0,
    },
    {
      label: t('option.settlementBacklog'),
      value: value.pendingSettlementCount + value.failedSettlementCount,
      oldest: value.oldestSettlementTime,
      danger: value.failedSettlementCount > 0,
    },
    {
      label: t('option.liquidationBacklog'),
      value: value.pendingLiquidationCount + value.exceptionLiquidationCount,
      oldest: value.oldestLiquidationTime,
      danger: value.exceptionLiquidationCount > 0,
    },
    {
      label: t('option.eventBacklog'),
      value: value.pendingOutboxCount + value.pendingInboxCount,
      oldest:
        value.oldestOutboxTime && value.oldestInboxTime
          ? Math.min(value.oldestOutboxTime, value.oldestInboxTime)
          : value.oldestOutboxTime || value.oldestInboxTime,
      danger: value.pendingOutboxCount + value.pendingInboxCount > 0,
    },
    {
      label: t('option.physicalDeliveryExceptions'),
      value: value.physicalExceptionCount,
      oldest: 0,
      danger: value.physicalExceptionCount > 0,
    },
    {
      label: t('option.comboOrderExceptions'),
      value: value.comboStaleCount + value.comboManualReviewCount,
      oldest: value.oldestComboExceptionTime,
      danger: value.comboStaleCount + value.comboManualReviewCount > 0,
    },
    {
      label: t('option.comboIntegrityIssues'),
      value: value.comboInvariantIssueCount + value.comboIncompleteMatchGroupCount,
      oldest: 0,
      danger: value.comboInvariantIssueCount + value.comboIncompleteMatchGroupCount > 0,
    },
  ]
})

const loadOverview = async () => {
  if (!query.tenantId) return
  overview.value = (
    await optionService.getOperationsOverview({
      tenantId: query.tenantId,
      riskStaleSeconds: query.riskStaleSeconds,
      comboStaleSeconds: query.comboStaleSeconds,
    })
  ).data
}

const loadInstructions = async () => {
  if (!query.tenantId) return
  const res = await optionService.listAssetInstructions({
    tenantId: query.tenantId,
    bizNo: query.bizNo || undefined,
    status: query.instructionStatus,
    cursor: instructionPage.cursor,
    limit: instructionPage.limit,
  })
  instructions.value = res.data || []
  instructionPagination.updateFromResponse(res)
}

const loadIssues = async () => {
  if (!query.tenantId) return
  const res = await optionService.listReconciliationIssues({
    tenantId: query.tenantId,
    bizNo: query.bizNo || undefined,
    cursor: issuePage.cursor,
    limit: issuePage.limit,
  })
  issues.value = res.data || []
  issuePagination.updateFromResponse(res)
}

const loadComboOrders = async () => {
  if (!query.tenantId) return
  const res = await optionService.listAdminComboOrders({
    tenantId: query.tenantId,
    comboNo: query.comboNo || undefined,
    status: query.comboStatus,
    cursor: comboPage.cursor,
    limit: comboPage.limit,
  })
  comboOrders.value = res.data || []
  comboPagination.updateFromResponse(res)
}

const refreshAll = async () => {
  if (!query.tenantId) return
  loading.value = true
  try {
    await Promise.all([loadOverview(), loadComboOrders(), loadInstructions(), loadIssues()])
  } finally {
    loading.value = false
  }
}

const resetQuery = () => {
  query.tenantId = undefined
  query.riskStaleSeconds = 60
  query.comboStaleSeconds = 60
  query.bizNo = ''
  query.comboNo = ''
  query.comboStatus = undefined
  query.instructionStatus = undefined
  overview.value = undefined
  instructions.value = []
  issues.value = []
  comboOrders.value = []
  comboDetail.value = undefined
  comboDrawerVisible.value = false
  instructionPagination.reset()
  issuePagination.reset()
  comboPagination.reset()
}

const retryInstruction = async (row: OptionAssetInstruction) => {
  const { value } = await ElMessageBox.prompt(t('option.retryReason'), t('option.retry'), {
    inputType: 'textarea',
    inputValidator: (input) => {
      const reason = input?.trim() || ''
      if (!reason) return t('option.retryReasonRequired')
      return Array.from(reason).length <= 64 || t('option.retryReasonTooLong')
    },
  })
  await optionService.retryAssetInstruction({
    tenantId: row.tenantId,
    instructionId: row.id,
    reason: value.trim(),
  })
  ElMessage.success(t('common.success'))
  await refreshAll()
}

const openComboDetail = async (row: OptionComboOrderDetail) => {
  const tenantId = query.tenantId
  if (!tenantId) return
  comboDetail.value = (
    await optionService.getAdminComboOrder({
      tenantId,
      id: row.comboOrder.id,
    })
  ).data
  comboDrawerVisible.value = true
}

const forceCancelCombo = async (row: OptionComboOrderDetail) => {
  const tenantId = query.tenantId
  if (!tenantId) return
  const { value } = await ElMessageBox.prompt(
    t('option.forceCancelComboPrompt'),
    t('option.forceCancelCombo'),
    {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      inputPattern: /\S+/,
      inputErrorMessage: t('option.forceCancelReasonRequired'),
      inputValidator: (input) =>
        input.trim().length <= 200 || t('option.forceCancelReasonTooLong'),
    },
  )
  await optionService.forceCancelComboOrder({
    tenantId,
    id: row.comboOrder.id,
    reason: value.trim(),
  })
  ElMessage.success(t('common.success'))
  await loadComboOrders()
  if (comboDrawerVisible.value) {
    await openComboDetail(row)
  }
}

const instructionReset = () => instructionPagination.resetAndLoad(loadInstructions)
const instructionPrev = () => instructionPagination.prevAndLoad(loadInstructions)
const instructionNext = () => instructionPagination.nextAndLoad(loadInstructions)
const issueReset = () => issuePagination.resetAndLoad(loadIssues)
const issuePrev = () => issuePagination.prevAndLoad(loadIssues)
const issueNext = () => issuePagination.nextAndLoad(loadIssues)
const comboReset = () => comboPagination.resetAndLoad(loadComboOrders)
const comboPrev = () => comboPagination.prevAndLoad(loadComboOrders)
const comboNext = () => comboPagination.nextAndLoad(loadComboOrders)
const comboStatusLabel = (status: number) =>
  ({
    1: t('option.comboFunding'),
    2: t('option.comboActive'),
    3: t('option.comboPartFilled'),
    4: t('option.comboFilled'),
    5: t('option.comboCanceling'),
    6: t('option.comboCanceled'),
    7: t('option.comboRejected'),
    8: t('option.manualReview'),
  })[status] || String(status)
const comboStatusType = (status: number) => {
  if ([4].includes(status)) return 'success'
  if ([6, 7].includes(status)) return 'info'
  if ([5, 8].includes(status)) return 'danger'
  if ([1, 3].includes(status)) return 'warning'
  return 'primary'
}
const formatComboLegs = (row: OptionComboOrderDetail) =>
  row.legs
    .map(
      (leg) =>
        `#${leg.legNo} ${leg.contractId} ${leg.side === 1 ? 'BUY' : 'SELL'} ${leg.ratio}@${leg.price}`,
    )
    .join(' | ')
const formatTime = (value?: number) => (value ? new Date(value * 1000).toLocaleString() : '-')
const formatOldest = (value?: number) =>
  value ? `${t('option.oldest')}: ${formatTime(value)}` : '-'
const formatCoinAmounts = (items?: OptionCoinAmount[]) =>
  items?.length ? items.map((item) => `${item.amount} ${item.coin}`).join(', ') : '-'
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.generated-at,
.metric-label,
.metric small {
  color: var(--el-text-color-secondary);
}
.metric-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 12px;
}
.metric {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 14px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 6px;
}
.metric strong {
  font-size: 24px;
}
.metric strong.danger {
  color: var(--el-color-danger);
}
.detail-section {
  margin-top: 12px;
}
</style>
