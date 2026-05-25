<template>
  <div class="module-page">
    <div class="page-header">
      <h2>{{ t('trade.userTradeConfig') }}</h2>
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
          <el-switch v-model="tradeConfigForm.tradeEnabled" :active-value="1" :inactive-value="0" />
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
  </div>
</template>
<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { GetUserTradeConfigReq, SetUserTradeConfigReq, tradeService } from '@/services'

const { t } = useI18n()

const submitLoading = ref(false)
const riskQuery = reactive<GetUserTradeConfigReq>({
  tenantId: 0,
  userId: 0,
  symbolId: 0,
  marketType: 0,
})
const tradeConfigForm = reactive<SetUserTradeConfigReq>({
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

const loadCurrent = async () => {}

const loadRiskData = async () => {
  submitLoading.value = true
  try {
    Object.assign(
      tradeConfigForm,
      riskQuery,
      (await tradeService.getUserTradeConfig(riskQuery)).data || {},
    )
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

onMounted(loadCurrent)
</script>
<style scoped></style>
