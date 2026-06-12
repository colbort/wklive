<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { OptionGroup } from '@/services'
import { formatDate } from '@/utils'
import { getOptionValueLabel } from '@/utils/options'

type DetailRecord = Record<string, unknown>

const props = withDefaults(
  defineProps<{
    data?: object | null
    optionGroups?: OptionGroup[]
    columns?: number
  }>(),
  {
    data: null,
    optionGroups: () => [],
    columns: 2,
  },
)

const { t } = useI18n()

const commonLabels = new Set([
  'id',
  'tenantId',
  'userId',
  'status',
  'remark',
  'sort',
  'icon',
  'createTimes',
  'updateTimes',
])

const optionGroupByKey: Record<string, string> = {
  platformType: 'platformType',
  sceneType: 'sceneType',
  feeType: 'feeType',
  openStatus: 'openStatus',
  clientType: 'clientType',
  notifyStatus: 'notifyStatus',
  signResult: 'signResult',
  chainCode: 'chainCode',
  visible: 'visible',
  isDefault: 'yesNo',
  status: 'status',
}

const booleanKeys = new Set(['isDefault', 'visible', 'allowNewUser', 'allowOldUser'])

const entries = computed(() =>
  Object.entries((props.data || {}) as DetailRecord).filter(([, value]) => value !== undefined),
)

function translate(key: string) {
  const namespaces = commonLabels.has(key) ? ['common', 'payment'] : ['payment', 'common']
  for (const namespace of namespaces) {
    const localeKey = `${namespace}.${key}`
    const translated = t(localeKey)
    if (translated !== localeKey) return translated
  }
  return key
}

function isTimeKey(key: string) {
  return key === 'createTimes' || key === 'updateTimes' || /(?:Time|At)$/.test(key)
}

function isLongText(value: unknown) {
  return (
    typeof value === 'string' && (value.length > 80 || value.includes('{') || value.includes('['))
  )
}

function formatJsonText(value: string) {
  try {
    return JSON.stringify(JSON.parse(value), null, 2)
  } catch {
    return value
  }
}

function displayValue(key: string, value: unknown) {
  if (value === null || value === '') return '-'
  if (isTimeKey(key) && typeof value === 'number') return formatDate(value)
  if (booleanKeys.has(key)) return Number(value) === 1 ? t('users.yes') : t('users.no')

  const groupKey =
    key === 'status' && props.data && 'orderNo' in props.data
      ? 'payOrderStatus'
      : optionGroupByKey[key]
  if (groupKey && props.optionGroups.length) {
    const label = getOptionValueLabel(props.optionGroups, groupKey, value as string | number, t)
    if (String(label) !== String(value) && label !== 0) return label
  }

  if (typeof value === 'object') return JSON.stringify(value, null, 2)
  return String(value)
}
</script>

<template>
  <el-empty v-if="!data" :description="t('common.noData')" />
  <el-descriptions v-else :column="columns" border>
    <el-descriptions-item
      v-for="[key, value] in entries"
      :key="key"
      :label="translate(key)"
      :span="isLongText(value) || typeof value === 'object' ? columns : 1"
    >
      <pre v-if="isLongText(value)" class="detail-code">{{ formatJsonText(String(value)) }}</pre>
      <pre v-else-if="typeof value === 'object' && value !== null" class="detail-code">{{
        displayValue(key, value)
      }}</pre>
      <span v-else class="detail-value">{{ displayValue(key, value) }}</span>
    </el-descriptions-item>
  </el-descriptions>
</template>

<style scoped>
.detail-value {
  overflow-wrap: anywhere;
}

.detail-code {
  margin: 0;
  max-height: 260px;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-word;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono', monospace;
  font-size: 12px;
  line-height: 1.5;
  color: #334155;
}
</style>
