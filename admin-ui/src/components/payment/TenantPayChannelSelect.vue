<template>
  <el-select
    v-model="selectedValue"
    :disabled="disabled || !tenantId"
    :placeholder="placeholder || t('common.pleaseSelect')"
    :clearable="clearable"
    filterable
    remote
    :remote-method="loadChannels"
    :loading="loading"
    style="width: 100%"
  >
    <el-option
      v-for="channel in channels"
      :key="channel.id"
      :label="`${channel.channelName || channel.channelCode} (${channel.id})`"
      :value="channel.id"
    />
  </el-select>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { tenantService, type TenantPayChannel } from '@/services'

const props = withDefaults(
  defineProps<{
    modelValue: number | undefined
    tenantId?: number
    disabled?: boolean
    clearable?: boolean
    enabledOnly?: boolean
    placeholder?: string
  }>(),
  {
    tenantId: 0,
    disabled: false,
    clearable: true,
    enabledOnly: true,
    placeholder: '',
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: number | undefined]
  change: [value: number | undefined]
  selected: [value: TenantPayChannel | null]
}>()

const { t } = useI18n()
const loading = ref(false)
const channels = ref<TenantPayChannel[]>([])

const selectedValue = computed({
  get: () => props.modelValue || undefined,
  set: (value) => {
    const nextValue = value == null ? undefined : Number(value)
    emit('update:modelValue', nextValue)
    emit('change', nextValue)
    emit(
      'selected',
      nextValue == null ? null : channels.value.find((item) => item.id === nextValue) || null,
    )
  },
})

async function loadChannels(keyword = '') {
  if (!props.tenantId) {
    channels.value = []
    return
  }
  loading.value = true
  try {
    const res = await tenantService.getTenantChannelList({
      tenantId: props.tenantId,
      keyword: keyword || undefined,
      enabled: props.enabledOnly ? 1 : undefined,
      limit: 100,
    })
    channels.value = res.data || []
  } finally {
    loading.value = false
  }
}

watch(
  () => props.tenantId,
  () => loadChannels(),
  { immediate: true },
)
</script>
