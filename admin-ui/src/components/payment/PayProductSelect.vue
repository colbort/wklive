<template>
  <el-select
    v-model="selectedValue"
    :disabled="disabled || !platformId"
    :placeholder="placeholder || t('common.pleaseSelect')"
    :clearable="clearable"
    filterable
    remote
    :remote-method="loadProducts"
    :loading="loading"
    style="width: 100%"
  >
    <el-option
      v-for="product in products"
      :key="product.id"
      :label="`${product.productName || product.productCode} (${product.id})`"
      :value="product.id"
    />
  </el-select>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { catalogService, type PayProduct } from '@/services'

const props = withDefaults(
  defineProps<{
    modelValue: number | undefined
    platformId?: number
    disabled?: boolean
    clearable?: boolean
    placeholder?: string
  }>(),
  {
    platformId: 0,
    disabled: false,
    clearable: true,
    placeholder: '',
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: number | undefined]
  change: [value: number | undefined]
  selected: [value: PayProduct | null]
}>()

const { t } = useI18n()
const loading = ref(false)
const products = ref<PayProduct[]>([])

const selectedValue = computed({
  get: () => props.modelValue || undefined,
  set: (value) => {
    const nextValue = value == null ? undefined : Number(value)
    emit('update:modelValue', nextValue)
    emit('change', nextValue)
    emit(
      'selected',
      nextValue == null ? null : products.value.find((item) => item.id === nextValue) || null,
    )
  },
})

async function loadProducts(keyword = '') {
  if (!props.platformId) {
    products.value = []
    return
  }
  loading.value = true
  try {
    const res = await catalogService.getProductList({
      platformId: props.platformId,
      keyword: keyword || undefined,
      enabled: 1,
      limit: 100,
    })
    products.value = res.data || []
  } finally {
    loading.value = false
  }
}

watch(
  () => props.platformId,
  () => loadProducts(),
  { immediate: true },
)
</script>
