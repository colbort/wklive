<template>
  <div class="module-page">
    <div class="page-header">
      <h2>{{ t('trade.riskControls') }}</h2>
      <div class="header-actions">
        <el-button @click="loadCurrent">
          {{ t('common.refresh') }}
        </el-button>
      </div>
    </div>

    <el-card shadow="never" class="query-card">
      <template #header>
        {{ t('trade.riskQuery') }}
      </template>

      <el-form :model="riskQuery" inline label-width="90px">
        <el-form-item :label="t('trade.tenantId')">
          <el-input-number v-model="riskQuery.tenantId" :min="0" :precision="0" />
        </el-form-item>

        <el-form-item :label="t('trade.userId')">
          <el-input-number v-model="riskQuery.userId" :min="0" :precision="0" />
        </el-form-item>

        <el-form-item :label="t('trade.symbolId')">
          <el-input-number v-model="riskQuery.symbolId" :min="0" :precision="0" />
        </el-form-item>

        <el-form-item :label="t('trade.marketType')">
          <el-input-number v-model="riskQuery.marketType" :min="0" :precision="0" />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="loadRiskData">
            {{ t('trade.loadConfig') }}
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <div class="risk-grid">
      <el-card shadow="never">
        <template #header>
          {{ t('trade.userTradeLimit') }}
        </template>

        <el-form label-width="120px">
          <el-form-item :label="t('trade.canOpen')">
            <el-switch v-model="tradeLimitForm.canOpen" :active-value="1" :inactive-value="0" />
          </el-form-item>

          <el-form-item :label="t('trade.canClose')">
            <el-switch v-model="tradeLimitForm.canClose" :active-value="1" :inactive-value="0" />
          </el-form-item>

          <el-form-item :label="t('trade.canCancel')">
            <el-switch v-model="tradeLimitForm.canCancel" :active-value="1" :inactive-value="0" />
          </el-form-item>

          <el-form-item :label="t('trade.onlyReduceOnly')">
            <el-switch
              v-model="tradeLimitForm.onlyReduceOnly"
              :active-value="1"
              :inactive-value="0"
            />
          </el-form-item>

          <el-form-item :label="t('trade.tradeEnabled')">
            <el-switch
              v-model="tradeLimitForm.tradeEnabled"
              :active-value="1"
              :inactive-value="0"
            />
          </el-form-item>

          <el-form-item :label="t('trade.maxOpenOrderCount')">
            <el-input-number v-model="tradeLimitForm.maxOpenOrderCount" :min="0" :precision="0" />
          </el-form-item>

          <el-form-item :label="t('trade.maxPositionNotional')">
            <el-input v-model="tradeLimitForm.maxPositionNotional" />
          </el-form-item>

          <el-form-item :label="t('trade.operatorId')">
            <el-input-number v-model="tradeLimitForm.operatorId" :min="0" :precision="0" />
          </el-form-item>

          <el-button type="primary" :loading="submitLoading" @click="submitTradeLimit">
            {{ t('common.save') }}
          </el-button>
        </el-form>
      </el-card>

      <el-card shadow="never">
        <template #header>
          {{ t('trade.userSymbolLimit') }}
        </template>

        <el-form label-width="120px">
          <el-form-item :label="t('trade.maxPositionQty')">
            <el-input v-model="symbolLimitForm.maxPositionQty" />
          </el-form-item>

          <el-form-item :label="t('trade.maxOrderQty')">
            <el-input v-model="symbolLimitForm.maxOrderQty" />
          </el-form-item>

          <el-form-item :label="t('trade.minOrderQty')">
            <el-input v-model="symbolLimitForm.minOrderQty" />
          </el-form-item>

          <el-form-item :label="t('trade.priceDeviationRate')">
            <el-input v-model="symbolLimitForm.priceDeviationRate" />
          </el-form-item>

          <el-form-item :label="t('trade.operatorId')">
            <el-input-number v-model="symbolLimitForm.operatorId" :min="0" :precision="0" />
          </el-form-item>

          <el-button type="primary" :loading="submitLoading" @click="submitSymbolLimit">
            {{ t('common.save') }}
          </el-button>
        </el-form>
      </el-card>

      <el-card shadow="never">
        <template #header>
          {{ t('trade.userTradeConfig') }}
        </template>

        <el-form label-width="120px">
          <el-form-item :label="t('trade.positionMode')">
            <el-input-number v-model="tradeConfigForm.positionMode" :min="0" :precision="0" />
          </el-form-item>

          <el-form-item :label="t('trade.marginMode')">
            <el-input-number v-model="tradeConfigForm.marginMode" :min="0" :precision="0" />
          </el-form-item>

          <el-form-item :label="t('trade.defaultLeverage')">
            <el-input-number v-model="tradeConfigForm.defaultLeverage" :min="0" :precision="0" />
          </el-form-item>

          <el-form-item :label="t('trade.tradeEnabled')">
            <el-switch
              v-model="tradeConfigForm.tradeEnabled"
              :active-value="1"
              :inactive-value="0"
            />
          </el-form-item>

          <el-form-item :label="t('trade.reduceOnlyEnabled')">
            <el-switch
              v-model="tradeConfigForm.reduceOnlyEnabled"
              :active-value="1"
              :inactive-value="0"
            />
          </el-form-item>

          <el-button type="primary" :loading="submitLoading" @click="submitTradeConfig">
            {{ t('common.save') }}
          </el-button>
        </el-form>
      </el-card>

      <el-card shadow="never">
        <template #header>
          {{ t('trade.userLeverageConfig') }}
        </template>

        <el-form label-width="120px">
          <el-form-item :label="t('trade.marginMode')">
            <el-input-number v-model="leverageForm.marginMode" :min="0" :precision="0" />
          </el-form-item>

          <el-form-item :label="t('trade.positionMode')">
            <el-input-number v-model="leverageForm.positionMode" :min="0" :precision="0" />
          </el-form-item>

          <el-form-item :label="t('trade.longLeverage')">
            <el-input-number v-model="leverageForm.longLeverage" :min="0" :precision="0" />
          </el-form-item>

          <el-form-item :label="t('trade.shortLeverage')">
            <el-input-number v-model="leverageForm.shortLeverage" :min="0" :precision="0" />
          </el-form-item>

          <el-form-item :label="t('trade.maxLeverage')">
            <el-input-number v-model="leverageForm.maxLeverage" :min="0" :precision="0" />
          </el-form-item>

          <el-form-item :label="t('trade.operatorId')">
            <el-input-number v-model="leverageForm.operatorId" :min="0" :precision="0" />
          </el-form-item>

          <el-button type="primary" :loading="submitLoading" @click="submitLeverage">
            {{ t('common.save') }}
          </el-button>
        </el-form>
      </el-card>
    </div>

    <el-card shadow="never" class="table-card">
      <template #header>
        {{ t('trade.riskLogs') }}
      </template>

      <el-form :model="riskLogQuery" inline label-width="90px" class="query-card-inner">
        <el-form-item :label="t('trade.tenantId')">
          <el-input-number v-model="riskLogQuery.tenantId" :min="0" :precision="0" />
        </el-form-item>

        <el-form-item :label="t('trade.userId')">
          <el-input-number v-model="riskLogQuery.userId" :min="0" :precision="0" />
        </el-form-item>

        <el-form-item :label="t('trade.symbolId')">
          <el-input-number v-model="riskLogQuery.symbolId" :min="0" :precision="0" />
        </el-form-item>

        <el-form-item :label="t('trade.orderNo')">
          <el-input v-model="riskLogQuery.orderNo" clearable />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="loadRiskLogs">
            {{ t('common.search') }}
          </el-button>
        </el-form-item>
      </el-form>

      <el-table v-loading="loading" :data="rows" stripe>
        <el-table-column
          prop="orderNo"
          :label="t('trade.orderNo')"
          min-width="180"
          show-overflow-tooltip
        />

        <el-table-column prop="userId" :label="t('trade.userId')" width="100" />

        <el-table-column prop="symbolId" :label="t('trade.symbolId')" width="100" />

        <el-table-column prop="checkType" :label="t('trade.checkType')" width="100" />

        <el-table-column prop="checkResult" :label="t('trade.checkResult')" width="100" />

        <el-table-column
          prop="rejectMsg"
          :label="t('trade.rejectMsg')"
          min-width="220"
          show-overflow-tooltip
        />

        <el-table-column :label="t('common.actions')" width="100" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="showDetail(row)">
              {{ t('option.detail') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="detailVisible" :title="t('option.detail')" width="760px">
      <pre class="detail-pre">{{ JSON.stringify(detailData, null, 2) }}</pre>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import {
  SetUserSymbolLimitReq,
  SetUserTradeLimitReq,
  tradeService,
  type RiskOrderCheckLog,
} from '@/services'

const { t } = useI18n()

interface RiskQuery {
  tenantId: number
  userId: number
  symbolId: number
  marketType: number
  marginMode: number
}

interface RiskLogQuery {
  tenantId: number
  userId: number
  symbolId: number
  orderNo: string
  limit: number
}

interface TradeConfigForm {
  tenantId: number
  userId: number
  marketType: number
  symbolId: number
  positionMode: number
  marginMode: number
  defaultLeverage: number
  tradeEnabled: number
  reduceOnlyEnabled: number
}

interface LeverageForm {
  tenantId: number
  userId: number
  symbolId: number
  marketType: number
  marginMode: number
  positionMode: number
  longLeverage: number
  shortLeverage: number
  maxLeverage: number
  operatorId: number
  source: number
  status: number
  remark: string
}

const loading = ref(false)
const submitLoading = ref(false)
const rows = ref<RiskOrderCheckLog[]>([])
const detailVisible = ref(false)
const detailData = ref<RiskOrderCheckLog | null>(null)

const riskQuery = reactive<RiskQuery>({
  tenantId: 0,
  userId: 0,
  symbolId: 0,
  marketType: 0,
  marginMode: 0,
})

const riskLogQuery = reactive<RiskLogQuery>({
  tenantId: 0,
  userId: 0,
  symbolId: 0,
  orderNo: '',
  limit: 100,
})

const tradeLimitForm = reactive<SetUserTradeLimitReq>({
  tenantId: 0,
  userId: 0,
  marketType: 0,
  canOpen: 1,
  canClose: 1,
  canCancel: 1,
  canTriggerOrder: 1,
  canApiTrade: 1,
  tradeEnabled: 1,
  onlyReduceOnly: 0,
  maxOpenOrderCount: 0,
  maxOrderCountPerDay: 0,
  maxCancelCountPerDay: 0,
  maxOpenNotional: '',
  maxPositionNotional: '',
  riskLevel: 0,
  operatorId: 0,
  source: 0,
  status: 1,
  effectiveStartTime: 0,
  effectiveEndTime: 0,
  remark: '',
})

const symbolLimitForm = reactive<SetUserSymbolLimitReq>({
  tenantId: 0,
  userId: 0,
  symbolId: 0,
  marketType: 0,
  maxPositionQty: '',
  maxPositionNotional: '',
  maxOpenOrders: 0,
  maxOrderQty: '',
  maxOrderNotional: '',
  minOrderQty: '',
  minOrderNotional: '',
  maxLongPositionQty: '',
  maxShortPositionQty: '',
  priceDeviationRate: '',
  operatorId: 0,
  source: 0,
  status: 0,
  effectiveStartTime: 0,
  effectiveEndTime: 0,
  remark: '',
})

const tradeConfigForm = reactive<TradeConfigForm>({
  tenantId: 0,
  userId: 0,
  marketType: 0,
  symbolId: 0,
  positionMode: 0,
  marginMode: 0,
  defaultLeverage: 1,
  tradeEnabled: 1,
  reduceOnlyEnabled: 0,
})

const leverageForm = reactive<LeverageForm>({
  tenantId: 0,
  userId: 0,
  symbolId: 0,
  marketType: 0,
  marginMode: 0,
  positionMode: 0,
  longLeverage: 1,
  shortLeverage: 1,
  maxLeverage: 1,
  operatorId: 0,
  source: 0,
  status: 0,
  remark: '',
})

const pickList = (res: any) => res?.data || res?.list || []

const loadCurrent = async () => {
  await loadRiskLogs()
}

const loadRiskData = async () => {
  submitLoading.value = true
  try {
    Object.assign(
      tradeLimitForm,
      riskQuery,
      (await tradeService.getUserTradeLimit(riskQuery)).data || {},
    )
    Object.assign(
      symbolLimitForm,
      riskQuery,
      (await tradeService.getUserSymbolLimit(riskQuery)).data || {},
    )
    Object.assign(
      tradeConfigForm,
      riskQuery,
      (await tradeService.getUserTradeConfig(riskQuery)).data || {},
    )
    Object.assign(
      leverageForm,
      riskQuery,
      (await tradeService.getUserLeverageConfig(riskQuery)).data || {},
    )
    await loadRiskLogs()
  } finally {
    submitLoading.value = false
  }
}

const loadRiskLogs = async () => {
  loading.value = true
  try {
    rows.value = pickList(await tradeService.listRiskLogs(riskLogQuery))
  } finally {
    loading.value = false
  }
}

const showDetail = (row: RiskOrderCheckLog) => {
  detailData.value = row
  detailVisible.value = true
}

const submitTradeLimit = async () => {
  submitLoading.value = true
  try {
    await tradeService.setUserTradeLimit(tradeLimitForm)
    ElMessage.success(t('trade.saveSuccessTradeLimit'))
  } finally {
    submitLoading.value = false
  }
}

const submitSymbolLimit = async () => {
  submitLoading.value = true
  try {
    await tradeService.setUserSymbolLimit(symbolLimitForm)
    ElMessage.success(t('trade.saveSuccessSymbolLimit'))
  } finally {
    submitLoading.value = false
  }
}

const submitTradeConfig = async () => {
  submitLoading.value = true
  try {
    await tradeService.setUserTradeConfig(tradeConfigForm)
    ElMessage.success(t('trade.saveSuccessTradeConfig'))
  } finally {
    submitLoading.value = false
  }
}

const submitLeverage = async () => {
  submitLoading.value = true
  try {
    await tradeService.setUserLeverageConfig(leverageForm)
    ElMessage.success(t('trade.saveSuccessLeverage'))
  } finally {
    submitLoading.value = false
  }
}

onMounted(loadCurrent)
</script>

<style scoped>
.detail-pre {
  white-space: pre-wrap;
  word-break: break-all;
}

.risk-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
  margin-bottom: 16px;
}

.query-card-inner {
  margin-bottom: 16px;
}
</style>
