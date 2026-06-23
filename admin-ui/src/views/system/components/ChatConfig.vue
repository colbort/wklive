<template>
  <div class="chat-config">
    <el-form-item :label="t('system.enabled')">
      <el-switch
        v-model="form.enabled"
        :active-value="1"
        :inactive-value="2"
        :active-text="t('common.enabled')"
        :inactive-text="t('common.disabled')"
      />
    </el-form-item>

    <el-form-item :label="t('system.chatApi')" prop="api">
      <el-input
        v-model="form.api"
        :placeholder="t('system.chatApiPlaceholder')"
      />
    </el-form-item>

    <el-form-item :label="t('system.chatUiUrl')" prop="ui_url">
      <el-input
        v-model="form.chat_ui_url"
        :placeholder="t('system.chatUiUrlPlaceholder')"
      />
    </el-form-item>

    <el-form-item :label="t('system.chatWsUrl')" prop="ws_url">
      <el-input
        v-model="form.chat_ws_url"
        :placeholder="t('system.chatWsUrlPlaceholder')"
      />
    </el-form-item>

    <el-row :gutter="20">
      <el-col :xs="24" :sm="12">
        <el-form-item :label="t('system.apiKey')" prop="api_key">
          <el-input
            v-model="form.api_key"
            :placeholder="t('system.apiKeyPlaceholder')"
          />
        </el-form-item>
      </el-col>
      <el-col :xs="24" :sm="12">
        <el-form-item :label="t('system.apiSecret')" prop="api_secret">
          <el-input
            v-model="form.api_secret"
            :placeholder="t('system.apiSecretPlaceholder')"
            show-password
          />
        </el-form-item>
      </el-col>
    </el-row>

    <el-alert
      :title="t('system.chatConfigTip')"
      type="info"
      show-icon
      :closable="false"
    />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { ChatConfig } from '@/services/system/ConfigService'

const { t } = useI18n()

const props = defineProps<{
  modelValue: ChatConfig
}>()

const emit = defineEmits<{
  'update:modelValue': [value: ChatConfig]
}>()

const form = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value),
})
</script>

<style scoped>
.chat-config {
  width: 100%;
}
</style>
