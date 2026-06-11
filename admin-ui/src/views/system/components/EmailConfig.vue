<template>
  <div class="email-config">
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
      <el-col :xs="24" :sm="14">
        <el-form-item :label="t('system.smtpHost')">
          <el-input v-model="form.smtp_host" :placeholder="t('system.smtpHostPlaceholder')" />
        </el-form-item>
      </el-col>
      <el-col :xs="24" :sm="10">
        <el-form-item :label="t('system.smtpPort')">
          <el-input-number
            v-model="form.smtp_port"
            :min="0"
            :precision="0"
            class="full-control"
          />
        </el-form-item>
      </el-col>
    </el-row>
    <el-row :gutter="20">
      <el-col :xs="24" :sm="12">
        <el-form-item :label="t('system.smtpUsername')">
          <el-input v-model="form.username" :placeholder="t('common.pleaseEnter')" />
        </el-form-item>
      </el-col>
      <el-col :xs="24" :sm="12">
        <el-form-item :label="t('system.smtpPassword')">
          <el-input v-model="form.password" :placeholder="t('common.pleaseEnter')" show-password />
        </el-form-item>
      </el-col>
    </el-row>
    <el-row :gutter="20">
      <el-col :xs="24" :sm="12">
        <el-form-item :label="t('system.fromEmail')">
          <el-input v-model="form.from_email" :placeholder="t('system.fromEmailPlaceholder')" />
        </el-form-item>
      </el-col>
      <el-col :xs="24" :sm="12">
        <el-form-item :label="t('system.fromName')">
          <el-input v-model="form.from_name" :placeholder="t('system.fromNamePlaceholder')" />
        </el-form-item>
      </el-col>
    </el-row>
    <el-form-item :label="t('system.subjectTemplate')">
      <el-input
        v-model="form.subject_template"
        :placeholder="t('system.subjectTemplatePlaceholder')"
      />
    </el-form-item>
    <el-form-item :label="t('system.bodyTemplate')">
      <el-input
        v-model="form.body_template"
        type="textarea"
        :rows="5"
        :placeholder="t('system.emailBodyTemplatePlaceholder')"
      />
    </el-form-item>
    <el-alert
      :title="t('system.verificationTemplateTip')"
      type="info"
      show-icon
      :closable="false"
    />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { EmailConfig } from '@/services/system/ConfigService'

const { t } = useI18n()

const props = defineProps<{
  modelValue: EmailConfig
}>()

const emit = defineEmits<{
  'update:modelValue': [value: EmailConfig]
}>()

const form = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value),
})
</script>

<style scoped>
.email-config {
  width: 100%;
}

.full-control {
  width: 100%;
}
</style>
