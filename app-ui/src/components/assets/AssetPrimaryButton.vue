<script setup lang="ts">
defineProps<{
  label: string
}>()
</script>

<template>
  <button type="button" class="asset-primary-button">{{ label }}</button>
</template>

<style scoped>
.asset-primary-button {
  width: 100%;
  min-height: 60px;
  border: 0;
  border-radius: 16px;
  background: #02b904;
  color: #fff;
  font: inherit;
  font-size: 18px;
  font-weight: 800;
}
</style>
