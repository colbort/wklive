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
    <el-form-item label="K线校正间隔(分钟)">
      <el-input-number v-model="form.reconcile_interval_minutes" :min="1" :max="60" />
    </el-form-item>
    <el-form-item label="校正窗口K线数量">
      <el-input-number v-model="form.reconcile_window_bars" :min="5" :max="500" />
    </el-form-item>
    <el-form-item label="缺口扫描间隔(分钟)">
      <el-input-number v-model="form.gap_scan_interval_minutes" :min="5" :max="1440" />
    </el-form-item>
    <el-form-item label="缺口修复批量大小">
      <el-input-number v-model="form.repair_batch_size" :min="1" :max="100" />
    </el-form-item>
    <el-form-item label="当前K线桶TTL(分钟)">
      <el-input-number v-model="form.building_bucket_ttl_minutes" :min="10" :max="1440" />
    </el-form-item>
    <el-form-item label="WS K线失效时间(秒)">
      <el-input-number v-model="form.ws_kline_stale_seconds" :min="5" :max="300" />
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
