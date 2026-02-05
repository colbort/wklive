<script setup lang="ts">
import { useRoute } from 'vue-router'
const route = useRoute()
</script>

<template>
  <el-result icon="error" title="缺少页面组件" sub-title="后端返回的 component 在前端不存在">
    <template #extra>
      <div style="color:#666;">missing: {{ (route.meta as any).missing }}</div>
      <el-button style="margin-top: 12px" @click="$router.push('/')">回到首页</el-button>
    </template>
  </el-result>
</template>
