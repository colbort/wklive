<template>
  <div class="withdraw-config">
    <el-form-item :label="t('system.minAmount')">
      <el-input-number v-model="form.minAmount" :min="0" :precision="0" />
    </el-form-item>
    <el-form-item :label="t('system.maxAmount')">
      <el-input-number v-model="form.maxAmount" :min="0" :precision="0" />
    </el-form-item>
    <el-form-item :label="t('system.feeRate')">
      <el-input-number v-model="form.feeRate" :min="0" :precision="0" />
    </el-form-item>
    <el-form-item :label="t('system.dailyLimitPerUser')">
      <el-input-number v-model="form.dailyLimitPerUser" :min="0" :precision="0" />
    </el-form-item>
    <el-form-item :label="t('system.dailyAmountLimitPerUser')">
      <el-input-number v-model="form.dailyAmountLimitPerUser" :min="0" :precision="0" />
    </el-form-item>
    <el-form-item :label="t('system.allowedTimeRange')">
      <el-input v-model="form.allowedTimeRange" :placeholder="t('system.allowedTimeRangeTip')" />
    </el-form-item>
    <el-form-item :label="t('system.pendingWithdrawalLimitPerUser')">
      <el-input-number v-model="form.pendingWithdrawalLimitPerUser" :min="0" :precision="0" />
    </el-form-item>
    <el-form-item :label="t('system.freeWithdrawTimesPerDay')">
      <el-input-number v-model="form.freeWithdrawTimesPerDay" :min="0" :precision="0" />
    </el-form-item>
  </div>
</template>

<script setup lang='ts'>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { WithdrawConfig } from '@/services/system/ConfigService'

const { t } = useI18n()

const props = defineProps<{
  modelValue: WithdrawConfig
}>()

const emit = defineEmits<{
  'update:modelValue': [value: WithdrawConfig]
}>()

const form = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value),
})
</script>
