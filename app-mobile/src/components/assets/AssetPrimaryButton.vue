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
  background: var(--accent);
  color: var(--text);
  font: inherit;
  font-size: 0.9rem;
  font-weight: 800;
}
</style>
