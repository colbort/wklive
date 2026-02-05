<script setup lang="ts">
import { useI18n } from 'vue-i18n'
const { t } = useI18n()
</script>

<template>
  <el-card>
    <template #header>{{ t('system.roles') }}</template>

    <div style="display:flex; gap: 8px; flex-wrap: wrap;">
      <el-button type="primary" v-perm="'sys:role:add'">{{ t('perms.sys:role:add') }}</el-button>
      <el-button v-perm="'sys:role:update'">{{ t('perms.sys:role:update') }}</el-button>
      <el-button type="danger" v-perm="'sys:role:delete'">{{ t('perms.sys:role:delete') }}</el-button>
      <el-button v-perm="'sys:role:grant'">{{ t('perms.sys:role:grant') }}</el-button>
    </div>
  </el-card>
</template>
