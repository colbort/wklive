<script setup lang="ts">
import { useAuthStore } from '@/stores/auth'
const auth = useAuthStore()
</script>

<template>
  <el-card>
    <template #header>首页</template>
    <div>欢迎，{{ auth.user?.nickname || auth.user?.username || '-' }}</div>
    <div style="margin-top: 8px; color: #666;">
      本项目已内置多语言（vue-i18n），当前默认中文；后期你只需要补充 /src/i18n/locales/*.ts 即可。
    </div>
  </el-card>
</template>
