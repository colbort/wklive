<template>
  <div class="system-core-config">
    <el-form-item :label="t('system.configValue')" prop="site_name">
      <el-input
        v-model="form.site_name"
        :placeholder="t('system.siteName') || t('common.pleaseEnter')"
      />
    </el-form-item>
    <el-form-item :label="t('system.siteLogo')" prop="site_logo">
      <el-input
        v-model="form.site_logo"
        :placeholder="t('system.siteLogo') || t('common.pleaseEnter')"
      />
    </el-form-item>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { SystemCore } from '@/services/system/ConfigService'

const { t } = useI18n()

interface Props {
  modelValue: SystemCore
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'update:modelValue': [value: SystemCore]
}>()

const form = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

import { computed } from 'vue'
</script>

<style scoped>
.system-core-config {
  width: 100%;
}
</style>
