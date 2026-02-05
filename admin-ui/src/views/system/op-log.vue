<script setup lang="ts">
import { useI18n } from 'vue-i18n'
const { t } = useI18n()
</script>

<template>
  <el-card>
    <template #header>{{ t('system.opLog') }}</template>
    <div style="color:#666;">后续接 /admin/logs/op</div>
  </el-card>
</template>
