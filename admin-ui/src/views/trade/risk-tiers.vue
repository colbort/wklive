<template>
  <TradeOperationList :key="version" kind="riskTiers">
    <template #actions>
      <el-button v-perm="'trade:risk-tier:update'" type="primary" @click="visible = true">
        {{ t('trade.addRiskTier') }}
      </el-button>
    </template>
  </TradeOperationList>
  <el-dialog v-model="visible" :title="t('trade.addRiskTier')" width="620px">
    <el-form label-width="150px">
      <el-form-item :label="t('trade.symbolId')">
        <el-input-number v-model="form.symbolId" :min="1" />
      </el-form-item><el-form-item :label="t('trade.tierNo')">
        <el-input-number v-model="form.tierNo" :min="1" />
      </el-form-item><el-form-item :label="t('trade.notionalFloor')">
        <el-input v-model="form.notionalFloor" />
      </el-form-item><el-form-item :label="t('trade.notionalCap')">
        <el-input v-model="form.notionalCap" />
      </el-form-item><el-form-item :label="t('trade.maxLeverage')">
        <el-input-number v-model="form.maxLeverage" :min="1" />
      </el-form-item><el-form-item :label="t('trade.initialMarginRate')">
        <el-input v-model="form.initialMarginRate" />
      </el-form-item><el-form-item :label="t('trade.maintenanceMarginRate')">
        <el-input v-model="form.maintenanceMarginRate" />
      </el-form-item><el-form-item :label="t('trade.maintenanceAmount')">
        <el-input v-model="form.maintenanceAmount" />
      </el-form-item>
    </el-form><template #footer>
      <el-button @click="visible = false">
        {{ t('common.cancel') }}
      </el-button><el-button type="primary" @click="save">
        {{ t('common.confirm') }}
      </el-button>
    </template>
  </el-dialog>
</template>
<script setup lang="ts">
import { reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { apiTradeSetRiskTier } from '@/api/trade'
import TradeOperationList from '@/components/trade/TradeOperationList.vue'
const { t } = useI18n()
const visible = ref(false),
  version = ref(0)
const form = reactive({
  symbolId: 0,
  tierNo: 1,
  notionalFloor: '0',
  notionalCap: '0',
  maxLeverage: 1,
  initialMarginRate: '0',
  maintenanceMarginRate: '0',
  maintenanceAmount: '0',
  enabled: 1,
})
async function save() {
  await apiTradeSetRiskTier(form)
  ElMessage.success(t('common.success'))
  visible.value = false
  version.value++
}
</script>
