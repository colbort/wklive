<template>
  <el-select
    v-model="selectedValue"
    :disabled="disabled"
    :placeholder="placeholder || t('common.pleaseSelect')"
    filterable
    remote
    clearable
    :remote-method="loadTenants"
    :loading="loading"
    style="width: 100%"
    @visible-change="handleVisibleChange"
  >
    <el-option v-if="includeSystem" :label="t('system.systemConfigScope')" :value="0" />
    <el-option
      v-for="item in tenants"
      :key="item.id"
      :label="formatTenantLabel(item)"
      :value="item.id"
    />
  </el-select>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { tenantsService, type SysTenantItem } from '@/services'

const props = withDefaults(
  defineProps<{
    modelValue: number
    disabled?: boolean
    includeSystem?: boolean
    placeholder?: string
  }>(),
  {
    disabled: false,
    includeSystem: false,
    placeholder: '',
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: number]
  change: [value: number]
  selected: [value: SysTenantItem | null]
}>()

const { t } = useI18n()
const loading = ref(false)
const tenants = ref<SysTenantItem[]>([])

const selectedValue = computed({
  get: () => props.modelValue,
  set: (value) => {
    const nextValue = Number(value || 0)
    emit('update:modelValue', nextValue)
    emit('change', nextValue)
    emit('selected', tenants.value.find((tenant) => tenant.id === nextValue) || null)
  },
})

function formatTenantLabel(item: SysTenantItem) {
  const name = item.tenantName || item.tenantCode || String(item.id)
  return `${name} (${item.id})`
}

function mergeTenant(item?: SysTenantItem) {
  if (!item?.id || tenants.value.some((tenant) => tenant.id === item.id)) return
  tenants.value = [item, ...tenants.value]
}

async function loadTenants(keyword = '') {
  loading.value = true
  try {
    const res = await tenantsService.getList({
      keyword: keyword || undefined,
      limit: 100,
    })
    tenants.value = res.data || []
    await ensureCurrentTenant()
  } finally {
    loading.value = false
  }
}

async function ensureCurrentTenant() {
  if (!props.modelValue || props.modelValue === 0) return
  const existing = tenants.value.find((tenant) => tenant.id === props.modelValue)
  if (existing) {
    emit('selected', existing)
    return
  }

  const res = await tenantsService.detail({ tenantId: props.modelValue })
  mergeTenant(res.data)
  emit('selected', res.data || null)
}

function handleVisibleChange(visible: boolean) {
  if (visible && tenants.value.length === 0) {
    loadTenants()
  }
}

watch(
  () => props.modelValue,
  () => {
    ensureCurrentTenant()
  },
)

onMounted(() => {
  loadTenants()
})
</script>
