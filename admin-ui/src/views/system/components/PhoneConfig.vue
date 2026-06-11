<template>
  <div class="phone-config">
    <el-form-item :label="t('system.enabled')">
      <el-switch
        v-model="form.enabled"
        :active-value="1"
        :inactive-value="2"
        :active-text="t('common.enabled')"
        :inactive-text="t('common.disabled')"
      />
    </el-form-item>
    <el-row :gutter="20">
      <el-col :xs="24" :sm="10">
        <el-form-item :label="t('system.smsProvider')">
          <el-input v-model="form.provider" :placeholder="t('system.smsProviderPlaceholder')" />
        </el-form-item>
      </el-col>
      <el-col :xs="24" :sm="14">
        <el-form-item :label="t('system.endpoint')">
          <el-input v-model="form.endpoint" :placeholder="t('system.endpointPlaceholder')" />
        </el-form-item>
      </el-col>
    </el-row>
    <el-form-item :label="t('system.httpMethod')">
      <el-select v-model="form.method" class="method-select">
        <el-option label="POST" value="POST" />
        <el-option label="GET" value="GET" />
        <el-option label="PUT" value="PUT" />
      </el-select>
    </el-form-item>
    <el-form-item :label="t('system.headersJson')">
      <el-input
        v-model="form.headers_json"
        type="textarea"
        :rows="4"
        :placeholder="t('system.headersJsonPlaceholder')"
      />
    </el-form-item>
    <el-form-item :label="t('system.bodyTemplate')">
      <el-input
        v-model="form.body_template"
        type="textarea"
        :rows="6"
        :placeholder="t('system.phoneBodyTemplatePlaceholder')"
      />
    </el-form-item>
    <el-alert
      :title="t('system.phoneTemplateTip')"
      type="info"
      show-icon
      :closable="false"
    />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { PhoneConfig } from '@/services/system/ConfigService'

const { t } = useI18n()

const props = defineProps<{
  modelValue: PhoneConfig
}>()

const emit = defineEmits<{
  'update:modelValue': [value: PhoneConfig]
}>()

const form = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value),
})
</script>

<style scoped>
.phone-config {
  width: 100%;
}

.method-select {
  width: 180px;
  max-width: 100%;
}
</style>
