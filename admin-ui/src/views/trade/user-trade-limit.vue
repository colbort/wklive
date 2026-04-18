<template>
  <div class="module-page">
    <div class="page-header">
      <h2>{{ t('trade.userTradeLimit') }}</h2>
      <div class="header-actions">
        <el-button @click="loadCurrent">
          {{ t('common.refresh') }}
        </el-button>
      </div>
    </div>
    <el-card shadow="never" class="query-card">
      <template #header>
        {{ t('trade.riskQuery') }}
      </template><el-form :model="riskQuery" inline label-width="90px">
        <el-form-item :label="t('trade.tenantId')">
          <el-input-number v-model="riskQuery.tenantId" :min="0" :precision="0" />
        </el-form-item><el-form-item :label="t('trade.userId')">
          <el-input-number v-model="riskQuery.userId" :min="0" :precision="0" />
        </el-form-item><el-form-item :label="t('trade.symbolId')">
          <el-input-number v-model="riskQuery.symbolId" :min="0" :precision="0" />
        </el-form-item><el-form-item :label="t('trade.marketType')">
          <el-input-number v-model="riskQuery.marketType" :min="0" :precision="0" />
        </el-form-item><el-form-item>
          <el-button type="primary" @click="loadRiskData">
            {{ t('trade.loadConfig') }}
          </el-button>
        </el-form-item>
      </el-form>
    </el-card><el-card shadow="never">
      <el-form label-width="120px">
        <el-form-item :label="t('trade.canOpen')">
          <el-switch
            v-model="tradeLimitForm.canOpen"
            :active-value="1"
            :inactive-value="0"
          />
        </el-form-item><el-form-item :label="t('trade.canClose')">
          <el-switch
            v-model="tradeLimitForm.canClose"
            :active-value="1"
            :inactive-value="0"
          />
        </el-form-item><el-form-item :label="t('trade.canCancel')">
          <el-switch
            v-model="tradeLimitForm.canCancel"
            :active-value="1"
            :inactive-value="0"
          />
        </el-form-item><el-form-item :label="t('trade.onlyReduceOnly')">
          <el-switch
            v-model="tradeLimitForm.onlyReduceOnly"
            :active-value="1"
            :inactive-value="0"
          />
        </el-form-item><el-form-item :label="t('trade.tradeEnabled')">
          <el-switch
            v-model="tradeLimitForm.tradeEnabled"
            :active-value="1"
            :inactive-value="0"
          />
        </el-form-item><el-form-item :label="t('trade.maxOpenOrderCount')">
          <el-input-number
            v-model="tradeLimitForm.maxOpenOrderCount"
            :min="0"
            :precision="0"
          />
        </el-form-item><el-form-item :label="t('trade.maxPositionNotional')">
          <el-input v-model="tradeLimitForm.maxPositionNotional" />
        </el-form-item><el-form-item :label="t('trade.operatorId')">
          <el-input-number
            v-model="tradeLimitForm.operatorId"
            :min="0"
            :precision="0"
          />
        </el-form-item><el-button type="primary" :loading="submitLoading" @click="submitTradeLimit">
          {{ t('common.save') }}
        </el-button>
      </el-form>
    </el-card>
  </div>
</template>
<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { GetUserTradeLimitReq, SetUserTradeLimitReq, tradeService } from '@/services'

const { t } = useI18n()

const submitLoading = ref(false)
const riskQuery = reactive<GetUserTradeLimitReq>({
  tenantId: 0,
  userId: 0,
  symbolId: 0,
  marketType: 0,
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

const loadCurrent = async () => {}

const loadRiskData = async () => {
  submitLoading.value = true
  try {
    Object.assign(
      tradeLimitForm,
      riskQuery,
      (await tradeService.getUserTradeLimit(riskQuery)).data || {},
    )
  } finally {
    submitLoading.value = false
  }
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

onMounted(loadCurrent)
</script>
<style scoped></style>
