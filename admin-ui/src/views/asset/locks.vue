<template>
  <AssetPanel mode="locks" :title="t('asset.locks')" />
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import AssetPanel from './asset-panel.vue'

const { t } = useI18n()
</script>
