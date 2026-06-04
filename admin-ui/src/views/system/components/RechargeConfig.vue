<template>
  <div class="recharge-config">
    <el-form-item :label="t('system.minAmount')">
      <el-input-number v-model="form.minAmount" :min="0" :precision="0" />
    </el-form-item>
    <el-form-item :label="t('system.maxAmount')">
      <el-input-number v-model="form.maxAmount" :min="0" :precision="0" />
    </el-form-item>
    <el-form-item :label="t('system.feeRate')">
      <el-input-number v-model="form.feeRate" :min="0" :precision="0" />
    </el-form-item>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { RechargeConfig } from '@/services/system/ConfigService'

const { t } = useI18n()

const props = defineProps<{
  modelValue: RechargeConfig
}>()

const emit = defineEmits<{
  'update:modelValue': [value: RechargeConfig]
}>()

const form = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value),
})
</script>
