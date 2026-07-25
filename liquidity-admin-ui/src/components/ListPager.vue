<script setup lang="ts">
defineProps<{
  total: number;
  canPrevious: boolean;
  canNext: boolean;
  loading?: boolean;
}>();

const limit = defineModel<number>("limit", { required: true });

defineEmits<{
  previous: [];
  next: [];
  limitChange: [];
}>();
</script>

<template>
  <footer v-if="total > 0" class="list-pager">
    <span>共 {{ total }} 条</span>
    <el-button :disabled="!canPrevious || loading" @click="$emit('previous')">上一页</el-button>
    <el-button type="primary" :disabled="!canNext || loading" @click="$emit('next')">下一页</el-button>
    <el-select v-model="limit" style="width: 110px" @change="$emit('limitChange')">
      <el-option :value="20" label="20条/页" />
      <el-option :value="50" label="50条/页" />
      <el-option :value="100" label="100条/页" />
    </el-select>
  </footer>
</template>
