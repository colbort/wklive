<template>
  <AssetPanel mode="user-assets" :title="t('asset.userAssets')" />
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import AssetPanel from './asset-panel.vue'

const { t } = useI18n()
</script>
