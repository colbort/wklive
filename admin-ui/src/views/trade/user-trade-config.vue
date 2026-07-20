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
        <SymbolSelect
          v-model="riskQuery.symbolId"
          :tenant-id="riskQuery.tenantId || undefined"
          :product-type="riskQuery.productType || undefined"
        />
      </el-form-item>
      <el-form-item :label="t('trade.productType')">
        <el-input-number v-model="riskQuery.productType" :min="0" :precision="0" />
      </el-form-item>
    </CrudQueryCard>
    <el-card shadow="never">
      <el-form label-width="120px">
        <el-form-item :label="t('trade.tradeEnabled')">
          <el-switch v-model="tradeConfigForm.tradeEnabled" :active-value="1" :inactive-value="2" />
        </el-form-item>
        <el-button
          v-perm="'trade:user-trade-config:update'"
          type="primary"
          :loading="submitLoading"
          @click="submitTradeConfig"
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
import { GetUserTradeConfigReq, SetUserTradeConfigReq, tradeService } from '@/services'
import TenantSelect from '@/components/TenantSelect.vue'
import UserSelect from '@/components/UserSelect.vue'
import SymbolSelect from '@/components/SymbolSelect.vue'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'

const { t } = useI18n()

const submitLoading = ref(false)
const riskQuery = reactive<GetUserTradeConfigReq>({
  tenantId: undefined as number | undefined,
  userId: undefined as number | undefined,
  symbolId: undefined as number | undefined,
  productType: undefined as number | undefined,
})
const tradeConfigForm = reactive<SetUserTradeConfigReq>({
  tenantId: 0,
  userId: 0,
  productType: 0,
  symbolId: 0,
  tradeEnabled: 1,
})

const loadList = async () => {
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

function resetQuery() {
  void loadList()
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

onMounted(loadList)
</script>
