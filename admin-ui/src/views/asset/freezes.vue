<template>
  <AssetPanel mode="freezes" :title="t('asset.freezes')" />
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import AssetPanel from './asset-panel.vue'

const { t } = useI18n()
</script>
