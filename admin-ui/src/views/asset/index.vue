<template>
  <AssetPanel mode="user-assets" :title="t('asset.management')" :show-tabs="true" />
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import AssetPanel from './asset-panel.vue'

const { t } = useI18n()
</script>
