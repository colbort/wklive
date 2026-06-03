import { computed, onMounted, ref } from 'vue'

import type { OptionsGroup, OptionsItem, RespBase } from '@/types/api'
import { t } from '@/i18n'

type OptionsLoader = () => Promise<RespBase & { data: OptionsGroup[] }>

function isSuccessCode(code: number) {
  return code === 0 || code === 200
}

export function optionText(option: OptionsItem) {
  const localized = t(`options.${option.code}`)
  if (localized !== `options.${option.code}`) return localized

  return option.code
    .replace(/^(WALLET_TYPE|ORDER_TYPE|TRIGGER_KIND|MARGIN_MODE|TRADE_SIDE|COMMON_STATUS|YES_NO)_/, '')
    .split('_')
    .filter(Boolean)
    .map((part) => part.slice(0, 1) + part.slice(1).toLowerCase())
    .join(' ')
}

export function useOptions(loader: OptionsLoader) {
  const groups = ref<OptionsGroup[]>([])
  const loading = ref(false)
  const error = ref('')

  const groupMap = computed(() => {
    return new Map(groups.value.map((group) => [group.key, group]))
  })

  async function loadOptions() {
    loading.value = true
    error.value = ''
    try {
      const resp = await loader()
      if (!isSuccessCode(resp.code)) {
        error.value = resp.msg || t('options.loadFailed')
        return
      }
      groups.value = resp.data || []
    } catch (err) {
      console.warn('load options failed', err)
      error.value = t('options.loadFailed')
    } finally {
      loading.value = false
    }
  }

  function getGroup(key: string) {
    return groupMap.value.get(key)?.options || []
  }

  function getLabel(key: string, value: number, fallback = '') {
    const option = getGroup(key).find((item) => item.value === value)
    return option ? optionText(option) : fallback
  }

  onMounted(() => {
    void loadOptions()
  })

  return {
    groups,
    loading,
    error,
    loadOptions,
    getGroup,
    getLabel,
  }
}
