<template>
  <el-select
    v-model="selectedValue"
    :disabled="disabled || !tenantId || !platformId"
    :placeholder="placeholder || t('common.pleaseSelect')"
    :clearable="clearable"
    filterable
    remote
    :remote-method="loadAccounts"
    :loading="loading"
    style="width: 100%"
  >
    <el-option
      v-for="account in accounts"
      :key="account.id"
      :label="`${account.accountName || account.accountCode} (${account.id})`"
      :value="account.id"
    />
  </el-select>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { tenantService, type TenantPayAccount } from '@/services'

const props = withDefaults(
  defineProps<{
    modelValue: number | undefined
    tenantId?: number
    platformId?: number
    disabled?: boolean
    clearable?: boolean
    placeholder?: string
  }>(),
  {
    tenantId: 0,
    platformId: 0,
    disabled: false,
    clearable: true,
    placeholder: '',
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: number | undefined]
  change: [value: number | undefined]
  selected: [value: TenantPayAccount | null]
}>()

const { t } = useI18n()
const loading = ref(false)
const accounts = ref<TenantPayAccount[]>([])

const selectedValue = computed({
  get: () => props.modelValue || undefined,
  set: (value) => {
    const nextValue = value == null ? undefined : Number(value)
    emit('update:modelValue', nextValue)
    emit('change', nextValue)
    emit(
      'selected',
      nextValue == null ? null : accounts.value.find((item) => item.id === nextValue) || null,
    )
  },
})

async function loadAccounts(keyword = '') {
  if (!props.tenantId || !props.platformId) {
    accounts.value = []
    return
  }
  loading.value = true
  try {
    const res = await tenantService.getTenantAccountList({
      tenantId: props.tenantId,
      platformId: props.platformId,
      keyword: keyword || undefined,
      enabled: 1,
      limit: 100,
    })
    accounts.value = res.data || []
  } finally {
    loading.value = false
  }
}

watch(
  () => [props.tenantId, props.platformId],
  () => loadAccounts(),
  { immediate: true },
)
</script>
