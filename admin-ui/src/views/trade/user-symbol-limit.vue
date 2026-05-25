<template>
  <div class="module-page">
    <div class="page-header">
      <h2>{{ t('trade.userSymbolLimit') }}</h2>
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
    <el-card shadow="never">
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
  </div>
</template>
<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { GetUserSymbolLimitReq, SetUserSymbolLimitReq, tradeService } from '@/services'

const { t } = useI18n()

const submitLoading = ref(false)
const riskQuery = reactive<GetUserSymbolLimitReq>({
  tenantId: 0,
  userId: 0,
  symbolId: 0,
  marketType: 0,
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

const loadCurrent = async () => {}

const loadRiskData = async () => {
  submitLoading.value = true
  try {
    Object.assign(
      symbolLimitForm,
      riskQuery,
      (await tradeService.getUserSymbolLimit(riskQuery)).data || {},
    )
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

onMounted(loadCurrent)
</script>
<style scoped></style>
