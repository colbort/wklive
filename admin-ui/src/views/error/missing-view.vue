<script setup lang="ts">
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
const route = useRoute()
const { t } = useI18n()
</script>

<template>
  <el-result
    icon="error"
    :title="t('error.missingComponent')"
    :sub-title="t('error.missingComponentSubtitle')"
  >
    <template #extra>
      <div style="color: #666">missing: {{ (route.meta as any).missing }}</div>
      <el-button style="margin-top: 12px" @click="$router.push('/')">{{
        t('common.goHome')
      }}</el-button>
    </template>
  </el-result>
</template>
