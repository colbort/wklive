<template>
  <div class="module-page">
    <CrudQueryCard :model="query" @search="loadOverview" @reset="resetQuery">
      <el-form-item v-if="authStore.isSystemAdmin" :label="t('option.tenantId')">
        <TenantSelect v-model="query.tenantId" class="tenant-select-filter" />
      </el-form-item>
      <el-form-item :label="t('option.riskStaleThreshold')">
        <el-input-number v-model="query.riskStaleSeconds" :min="10" :max="300" :step="10" />
      </el-form-item>
      <el-form-item :label="t('option.comboStaleThreshold')">
        <el-input-number v-model="query.comboStaleSeconds" :min="10" :max="300" :step="10" />
      </el-form-item>
    </CrudQueryCard>

    <el-card v-loading="loading" shadow="never" class="table-card overview-card">
      <div class="overview-scroll">
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
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'
import TenantSelect from '@/components/TenantSelect.vue'
import { optionService, type OptionCoinAmount, type OptionOperationsOverview } from '@/services'
import { useAuthStore } from '@/stores/auth'

const { t } = useI18n()
const authStore = useAuthStore()
const loading = ref(false)
const overview = ref<OptionOperationsOverview>()
const query = reactive({
  tenantId: (authStore.isTenantUser ? authStore.profileTenantId : undefined) as number | undefined,
  riskStaleSeconds: 60,
  comboStaleSeconds: 60,
})

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

async function loadOverview() {
  loading.value = true
  try {
    overview.value = (
      await optionService.getOperationsOverview({
        tenantId: query.tenantId,
        riskStaleSeconds: query.riskStaleSeconds,
        comboStaleSeconds: query.comboStaleSeconds,
      })
    ).data
  } finally {
    loading.value = false
  }
}

function resetQuery() {
  query.tenantId = authStore.isTenantUser ? authStore.profileTenantId || undefined : undefined
  query.riskStaleSeconds = 60
  query.comboStaleSeconds = 60
  overview.value = undefined
  void loadOverview()
}

const formatTime = (value?: number) => (value ? new Date(value * 1000).toLocaleString() : '-')
const formatOldest = (value?: number) =>
  value ? `${t('option.oldest')}: ${formatTime(value)}` : '-'
const formatCoinAmounts = (items?: OptionCoinAmount[]) =>
  items?.length ? items.map((item) => `${item.amount} ${item.coin}`).join(', ') : '-'

onMounted(loadOverview)
</script>

<style scoped>
.metric-label,
.metric small {
  color: var(--el-text-color-secondary);
}
.overview-scroll {
  height: 100%;
  overflow: auto;
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
</style>
