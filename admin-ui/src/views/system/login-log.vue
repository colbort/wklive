<script setup lang="ts">
import { useI18n } from 'vue-i18n'
const { t } = useI18n()
</script>

<template>
  <el-card>
    <template #header>{{ t('system.loginLog') }}</template>
    <div style="color:#666;">后续接 /admin/logs/login</div>
  </el-card>
</template>
