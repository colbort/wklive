<template>
  <div class="itick-config">
    <el-form-item :label="t('system.apiUrl')" prop="api_url">
      <el-input
        v-model="form.api_url"
        :placeholder="t('system.apiUrlPlaceholder') || t('common.pleaseEnter')"
      />
    </el-form-item>
    <el-form-item :label="t('system.apiToken')" prop="api_token">
      <el-input
        v-model="form.api_token"
        :placeholder="t('system.apiTokenPlaceholder') || t('common.pleaseEnter')"
        show-password
      />
    </el-form-item>
    <el-form-item :label="t('system.wsUrl')" prop="ws_url">
      <el-input
        v-model="form.ws_url"
        :placeholder="t('system.wsUrlPlaceholder') || t('common.pleaseEnter')"
      />
    </el-form-item>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { ItickConfig } from '@/services/system/ConfigService'

const { t } = useI18n()

interface Props {
  modelValue: ItickConfig
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'update:modelValue': [value: ItickConfig]
}>()

const form = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value),
})
</script>
