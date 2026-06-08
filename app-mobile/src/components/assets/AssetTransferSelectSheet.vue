<script setup lang="ts">
import { computed, ref, watch } from 'vue'

import AssetCoinIcon from '@/components/assets/AssetCoinIcon.vue'
import BottomDrawer from '@/components/common/BottomDrawer.vue'
import { useI18n } from '@/i18n'
import type { AssetCoinConfig, AssetUserAsset } from '@/types/asset'
import { formatAssetMinorAmount } from '@/utils/assetAmount'

const props = withDefaults(
  defineProps<{
    title: string
    modelValue: boolean
    source?: 'assets' | 'configs'
    excludedWalletType?: number
    excludedCoin?: string
    walletTypes: Array<{ value: number; label: string }>
    selectedWalletType: number
    selectedCoin: string
    assets: AssetUserAsset[]
    configs: AssetCoinConfig[]
  }>(),
  {
    source: 'assets',
    excludedWalletType: 0,
    excludedCoin: '',
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  select: [payload: { walletType: number; coin: string }]
}>()

const query = ref('')
const activeWalletType = ref(props.selectedWalletType)
const { t } = useI18n()

function coinKey(coin: string) {
  return String(coin || '').toUpperCase()
}

function isExcludedCoin(coin: string) {
  return (
    activeWalletType.value === props.excludedWalletType &&
    Boolean(props.excludedCoin) &&
    coinKey(coin) === coinKey(props.excludedCoin)
  )
}

const coins = computed(() => {
  const enabledConfigs = props.configs.filter(
    (config) => config.walletType === activeWalletType.value,
  )
  const configMap = new Map<string, AssetCoinConfig>()
  enabledConfigs.forEach((config) => {
    const key = coinKey(config.coin)
    if (!configMap.has(key)) configMap.set(key, config)
  })
  const assetMap = new Map<string, AssetUserAsset>()
  props.assets
    .filter((asset) => asset.walletType === activeWalletType.value)
    .forEach((asset) => {
      const key = coinKey(asset.coin)
      if (!assetMap.has(key)) assetMap.set(key, asset)
    })

  const rows =
    props.source === 'configs'
      ? Array.from(configMap.values()).map((config) => {
          const asset = assetMap.get(coinKey(config.coin))
          return {
            coin: config.coin,
            config,
            amountLabel: t('assetFlow.balance'),
            availableAmount: formatAssetMinorAmount(
              asset?.availableAmount || asset?.totalAmount || '0',
              config.decimalPlaces,
            ),
          }
        })
      : props.assets
          .filter((asset) => asset.walletType === activeWalletType.value)
          .map((asset) => ({
            coin: asset.coin,
            config: configMap.get(coinKey(asset.coin)),
            amountLabel: t('assetFlow.available'),
            availableAmount: formatAssetMinorAmount(
              asset.availableAmount || asset.totalAmount || '0',
              configMap.get(coinKey(asset.coin))?.decimalPlaces,
            ),
          }))

  const keyword = query.value.trim().toUpperCase()
  return rows.filter(
    (row) => !isExcludedCoin(row.coin) && (!keyword || row.coin.toUpperCase().includes(keyword)),
  )
})

watch(
  () => props.modelValue,
  (visible) => {
    if (visible) {
      activeWalletType.value = props.selectedWalletType
      query.value = ''
    }
  },
)

function close() {
  emit('update:modelValue', false)
}

function selectCoin(coin: string) {
  emit('select', { walletType: activeWalletType.value, coin })
  close()
}
</script>

<template>
  <BottomDrawer
    :model-value="modelValue"
    :title="title"
    :close-label="t('common.close')"
    max-height="84dvh"
    :z-index="110"
    @update:model-value="emit('update:modelValue', $event)"
  >
    <div class="asset-select-group">
      <h3>{{ t('assetFlow.chooseAccount') }}</h3>
      <div class="asset-select-accounts">
        <button
          v-for="account in walletTypes"
          :key="account.value"
          type="button"
          :class="{ active: activeWalletType === account.value }"
          @click="activeWalletType = account.value"
        >
          {{ account.label }}
        </button>
      </div>
    </div>

    <div class="asset-select-group">
      <h3>{{ t('assetFlow.chooseCoin') }}</h3>
      <label class="asset-select-search">
        <span>⌕</span>
        <input v-model="query" :placeholder="t('assetFlow.searchCoin')">
      </label>
    </div>

    <div class="asset-select-list">
      <button
        v-for="row in coins"
        :key="`${activeWalletType}-${row.coin}`"
        type="button"
        class="asset-select-row"
        @click="selectCoin(row.coin)"
      >
        <AssetCoinIcon :coin="row.coin" :config="row.config" />
        <strong>{{ row.coin }}</strong>
        <span>{{ row.amountLabel }} <b>{{ row.availableAmount }}</b></span>
        <em v-if="activeWalletType === selectedWalletType && row.coin === selectedCoin">✓</em>
      </button>
      <p v-if="!coins.length" class="asset-select-empty">
        {{ t('assets.noCoins') }}
      </p>
    </div>
  </BottomDrawer>
</template>

<style scoped>
button,
input {
  border: 0;
  background: transparent;
  color: inherit;
  font: inherit;
}

.asset-select-group {
  display: grid;
  gap: 16px;
  margin-bottom: 30px;
}

.asset-select-group h3 {
  margin: 0;
  font-size: 0.85rem;
  font-weight: 800;
}

.asset-select-accounts {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
}

.asset-select-accounts button {
  min-height: 54px;
  border-radius: 18px;
  background: #41434d;
  font-size: 0.9rem;
  font-weight: 800;
}

.asset-select-accounts button.active {
  background: var(--accent);
}

.asset-select-search {
  display: grid;
  grid-template-columns: 34px minmax(0, 1fr);
  align-items: center;
  min-height: 64px;
  padding: 0 20px;
  border-radius: 18px;
  background: #41434d;
}

.asset-select-search span {
  color: var(--text);
  font-size: 1.3rem;
}

.asset-select-search input {
  min-width: 0;
  outline: 0;
  font-size: 0.85rem;
}

.asset-select-search input::placeholder {
  color: var(--muted);
}

.asset-select-list {
  display: grid;
  margin: 0 -26px;
}

.asset-select-row {
  display: grid;
  grid-template-columns: 42px minmax(0, 1fr) auto 28px;
  align-items: center;
  gap: 12px;
  min-height: 72px;
  padding: 0 26px;
  border-bottom: 1px solid #333541;
  text-align: left;
}

.asset-select-row strong {
  font-size: 0.9rem;
}

.asset-select-row span {
  color: var(--text);
  font-size: 0.75rem;
}

.asset-select-row b {
  font-size: 0.95rem;
}

.asset-select-row em {
  color: var(--accent);
  font-size: 1.1rem;
  font-style: normal;
  font-weight: 800;
  text-align: right;
}

.asset-select-empty {
  margin: 26px;
  color: var(--muted);
  font-size: 0.7rem;
  text-align: center;
}
</style>
