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
        <el-button
          v-perm="'trade:user-leverage-config:update'"
          type="primary"
          :loading="submitLoading"
          @click="submitLeverage"
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
import { GetUserLeverageConfigReq, SetUserLeverageConfigReq, tradeService } from '@/services'
import TenantSelect from '@/components/TenantSelect.vue'
import UserSelect from '@/components/UserSelect.vue'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'

const { t } = useI18n()

const submitLoading = ref(false)
const riskQuery = reactive<GetUserLeverageConfigReq>({
  tenantId: undefined as number | undefined,
  userId: undefined as number | undefined,
  symbolId: undefined as number | undefined,
  marketType: undefined as number | undefined,
  marginMode: undefined as number | undefined,
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
  enabled: 1,
  remark: '',
})

const loadList = async () => {
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

function resetQuery() {
  void loadList()
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

onMounted(loadList)
</script>
