import { computed, onMounted, ref } from 'vue'

import type { OptionsGroup, OptionsItem, RespBase } from '@/types/api'

type OptionsLoader = () => Promise<RespBase & { data: OptionsGroup[] }>

const CODE_LABELS: Record<string, string> = {
  WALLET_TYPE_SPOT: '现金账户',
  WALLET_TYPE_FUNDING: '股票账户',
  WALLET_TYPE_CONTRACT: '合约账户',
  WALLET_TYPE_EARN: '理财账户',
  WALLET_TYPE_OPTION: '期权账户',
  ORDER_TYPE_LIMIT: '限价',
  ORDER_TYPE_MARKET: '市价',
  ORDER_TYPE_CONDITIONAL: '条件单',
  ORDER_TYPE_TAKE_PROFIT: '止盈',
  ORDER_TYPE_STOP_LOSS: '止损',
  MARGIN_MODE_CROSS: '全仓',
  MARGIN_MODE_ISOLATED: '逐仓',
  TRADE_SIDE_BUY: '买入',
  TRADE_SIDE_SELL: '卖出',
  COMMON_STATUS_ENABLED: '启用',
  COMMON_STATUS_DISABLED: '禁用',
  YES_NO_YES: '是',
  YES_NO_NO: '否',
}

function isSuccessCode(code: number) {
  return code === 0 || code === 200
}

export function optionText(option: OptionsItem) {
  if (CODE_LABELS[option.code]) return CODE_LABELS[option.code]

  return option.code
    .replace(/^(WALLET_TYPE|ORDER_TYPE|MARGIN_MODE|TRADE_SIDE|COMMON_STATUS|YES_NO)_/, '')
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
        error.value = resp.msg || 'options 加载失败'
        return
      }
      groups.value = resp.data || []
    } catch (err) {
      console.warn('load options failed', err)
      error.value = 'options 加载失败'
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
