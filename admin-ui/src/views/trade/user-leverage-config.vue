<template>
  <div class="module-page">
    <div class="page-header">
      <h2>{{ t('trade.userLeverageConfig') }}</h2>
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
</template>
<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { GetUserLeverageConfigReq, SetUserLeverageConfigReq, tradeService } from '@/services'

const { t } = useI18n()

const submitLoading = ref(false)
const riskQuery = reactive<GetUserLeverageConfigReq>({
  tenantId: 0,
  userId: 0,
  symbolId: 0,
  marketType: 0,
  marginMode: 0,
})
const leverageForm = reactive<SetUserLeverageConfigReq>({
  source: 0,
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
  status: 0,
  remark: '',
})

const loadCurrent = async () => {}

const loadRiskData = async () => {
  submitLoading.value = true
  try {
    Object.assign(
      leverageForm,
      riskQuery,
      (await tradeService.getUserLeverageConfig(riskQuery)).data || {},
    )
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
<style scoped></style>
