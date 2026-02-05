<script setup lang="ts">
import { useI18n } from 'vue-i18n'
const { t } = useI18n()
</script>

<template>
  <el-card>
    <template #header>{{ t('system.menus') }}</template>

    <div style="display:flex; gap: 8px; flex-wrap: wrap;">
      <el-button type="primary" v-perm="'sys:menu:add'">{{ t('perms.sys:menu:add') }}</el-button>
      <el-button v-perm="'sys:menu:update'">{{ t('perms.sys:menu:update') }}</el-button>
      <el-button type="danger" v-perm="'sys:menu:delete'">{{ t('perms.sys:menu:delete') }}</el-button>
    </div>
  </el-card>
</template>
