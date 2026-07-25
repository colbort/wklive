<template>
  <el-select
    v-model="selectedValue"
    :disabled="disabled"
    :placeholder="placeholder || t('common.pleaseSelect')"
    :clearable="clearable"
    filterable
    remote
    :remote-method="loadPlatforms"
    :loading="loading"
    style="width: 100%"
    @visible-change="handleVisibleChange"
  >
    <el-option
      v-for="platform in platforms"
      :key="platform.id"
      :label="formatPlatformLabel(platform)"
      :value="platform.id"
    />
  </el-select>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { catalogService, type PayPlatform } from '@/services'

const props = withDefaults(
  defineProps<{
    modelValue: number | undefined
    disabled?: boolean
    clearable?: boolean
    enabledOnly?: boolean
    placeholder?: string
  }>(),
  {
    disabled: false,
    clearable: true,
    enabledOnly: true,
    placeholder: '',
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: number | undefined]
  change: [value: number | undefined]
  selected: [value: PayPlatform | null]
}>()

const { t } = useI18n()
const loading = ref(false)
const platforms = ref<PayPlatform[]>([])

const selectedValue = computed({
  get: () => props.modelValue || undefined,
  set: (value) => {
    const nextValue = value === undefined || value === null ? undefined : Number(value)
    emit('update:modelValue', nextValue)
    emit('change', nextValue)
    emit(
      'selected',
      nextValue === undefined
        ? null
        : platforms.value.find((platform) => platform.id === nextValue) || null,
    )
  },
})

function formatPlatformLabel(platform: PayPlatform) {
  const name = platform.platformName || platform.platformCode || String(platform.id)
  return `${name} (${platform.id})`
}

function mergePlatform(platform?: PayPlatform) {
  if (!platform?.id || platforms.value.some((item) => item.id === platform.id)) return
  platforms.value = [platform, ...platforms.value]
}

async function ensureCurrentPlatform() {
  const platformId = selectedValue.value
  if (!platformId || platforms.value.some((platform) => platform.id === platformId)) return

  const res = await catalogService.getPlatformDetail(platformId)
  mergePlatform(res.data)
}

async function loadPlatforms(keyword = '') {
  loading.value = true
  try {
    const res = await catalogService.getPlatformList({
      keyword: keyword || undefined,
      enabled: props.enabledOnly ? 1 : undefined,
      limit: 100,
    })
    platforms.value = res.data || []
    await ensureCurrentPlatform()
  } finally {
    loading.value = false
  }
}

function handleVisibleChange(visible: boolean) {
  if (visible && platforms.value.length === 0) loadPlatforms()
}

watch(
  () => props.modelValue,
  () => ensureCurrentPlatform(),
)

onMounted(() => loadPlatforms())
</script>
