<template>
  <el-card v-if="card" shadow="never" class="query-card crud-query-card">
    <template v-if="$slots.header" #header>
      <slot name="header" />
    </template>

    <el-form
      :model="model"
      inline
      :label-width="labelWidth"
      class="crud-query-form"
    >
      <slot />

      <el-form-item v-if="showActions && (showSearch || showReset)" class="crud-query-actions">
        <el-button v-if="showSearch" type="primary" @click="emit('search')">
          {{ searchText || t('common.search') }}
        </el-button>
        <el-button v-if="showReset" @click="emit('reset')">
          {{ resetText || t('common.reset') }}
        </el-button>
      </el-form-item>

      <el-form-item v-if="$slots.actions" class="crud-query-actions crud-query-extra-actions">
        <slot name="actions" />
      </el-form-item>
    </el-form>
  </el-card>

  <div v-else class="query-card crud-query-card crud-query-card--plain">
    <el-form
      :model="model"
      inline
      :label-width="labelWidth"
      class="crud-query-form"
    >
      <slot />

      <el-form-item v-if="showActions && (showSearch || showReset)" class="crud-query-actions">
        <el-button v-if="showSearch" type="primary" @click="emit('search')">
          {{ searchText || t('common.search') }}
        </el-button>
        <el-button v-if="showReset" @click="emit('reset')">
          {{ resetText || t('common.reset') }}
        </el-button>
      </el-form-item>

      <el-form-item v-if="$slots.actions" class="crud-query-actions crud-query-extra-actions">
        <slot name="actions" />
      </el-form-item>
    </el-form>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'

withDefaults(
  defineProps<{
    model?: Record<string, unknown>
    labelWidth?: string
    showSearch?: boolean
    showReset?: boolean
    showActions?: boolean
    card?: boolean
    searchText?: string
    resetText?: string
  }>(),
  {
    model: () => ({}),
    labelWidth: '90px',
    showSearch: true,
    showReset: true,
    showActions: true,
    card: true,
    searchText: '',
    resetText: '',
  },
)

const emit = defineEmits<{
  search: []
  reset: []
}>()

const { t } = useI18n()
</script>
