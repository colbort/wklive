<template>
  <div class="module-page">
    <CrudQueryCard
      :model="riskQuery"
      :show-actions="false"
      @search="loadList"
      @reset="resetQuery"
    >
      <el-form-item :label="t('trade.tenantId')">
        <TenantSelect v-model="riskQuery.tenantId" class="tenant-select-filter" />
      </el-form-item>
      <el-form-item :label="t('trade.userId')">
        <UserSelect v-model="riskQuery.userId" :tenant-id="riskQuery.tenantId || undefined" />
      </el-form-item>
      <el-form-item :label="t('trade.symbolId')">
        <el-input-number v-model="riskQuery.symbolId" :min="0" :precision="0" />
      </el-form-item>
      <el-form-item :label="t('trade.marketType')">
        <el-input-number v-model="riskQuery.marketType" :min="0" :precision="0" />
      </el-form-item>
    </CrudQueryCard>
    <el-card shadow="never">
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
            :inactive-value="2"
          />
        </el-form-item>
        <el-form-item :label="t('trade.tradeEnabled')">
          <el-switch v-model="tradeLimitForm.tradeEnabled" :active-value="1" :inactive-value="2" />
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
        <el-button
          v-perm="'trade:user-trade-limit:update'"
          type="primary"
          :loading="submitLoading"
          @click="submitTradeLimit"
        >
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
import TenantSelect from '@/components/TenantSelect.vue'
import UserSelect from '@/components/UserSelect.vue'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'

const { t } = useI18n()

const submitLoading = ref(false)
const riskQuery = reactive<GetUserTradeLimitReq>({
  tenantId: undefined as number | undefined,
  userId: undefined as number | undefined,
  symbolId: undefined as number | undefined,
  marketType: undefined as number | undefined,
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
  onlyReduceOnly: 2,
  maxOpenOrderCount: 0,
  maxOrderCountPerDay: 0,
  maxCancelCountPerDay: 0,
  maxOpenNotional: '',
  maxPositionNotional: '',
  riskLevel: 0,
  operatorId: 0,
  source: 0,
  enabled: 1,
  effectiveStartTime: 0,
  effectiveEndTime: 0,
  remark: '',
})

const loadList = async () => {
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

function resetQuery() {
  void loadList()
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

onMounted(loadList)
</script>
