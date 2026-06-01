<script setup lang='ts'>
import { computed, ref, watch } from 'vue'

import AssetCoinIcon from '@/components/assets/AssetCoinIcon.vue'
import type { AssetCoinConfig, AssetUserAsset } from '@/types/asset'
import { formatAssetMinorAmount } from '@/utils/assetAmount'

const props = defineProps<{
  title: string
  modelValue: boolean
  walletTypes: Array<{ value: number; label: string }>
  selectedWalletType: number
  selectedCoin: string
  assets: AssetUserAsset[]
  configs: AssetCoinConfig[]
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  select: [payload: { walletType: number; coin: string }]
}>()

const query = ref('')
const activeWalletType = ref(props.selectedWalletType)

function coinKey(coin: string) {
  return String(coin || '').toUpperCase()
}

const coins = computed(() => {
  const enabledConfigs = props.configs.filter((config) => config.walletType === activeWalletType.value)
  const configMap = new Map<string, AssetCoinConfig>()
  enabledConfigs.forEach((config) => {
    const key = coinKey(config.coin)
    if (!configMap.has(key)) configMap.set(key, config)
  })
  const enabledCoins = new Set(configMap.keys())
  const assetMap = new Map<string, AssetUserAsset>()
  props.assets
    .filter((asset) => asset.walletType === activeWalletType.value)
    .filter((asset) => enabledCoins.has(coinKey(asset.coin)))
    .forEach((asset) => {
      const key = coinKey(asset.coin)
      if (!assetMap.has(key)) assetMap.set(key, asset)
    })

  const rows = Array.from(assetMap.values()).map((asset) => ({
      coin: asset.coin,
      config: configMap.get(coinKey(asset.coin)),
      availableAmount: formatAssetMinorAmount(
        asset.availableAmount || asset.totalAmount || '0',
        configMap.get(coinKey(asset.coin))?.decimalPlaces,
      ),
    }))

  const configuredRows = Array.from(configMap.values())
    .filter((config) => !assetMap.has(coinKey(config.coin)))
    .map((config) => ({
      coin: config.coin,
      config,
      availableAmount: formatAssetMinorAmount('0', config.decimalPlaces),
    }))

  const keyword = query.value.trim().toUpperCase()
  return [...rows, ...configuredRows].filter((row) => !keyword || row.coin.toUpperCase().includes(keyword))
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
  <Teleport to="body">
    <div v-if="modelValue" class="asset-select-overlay" role="presentation" @click.self="close">
      <section class="asset-select-sheet" role="dialog" aria-modal="true" :aria-label="title">
        <i class="asset-select-sheet__handle" />
        <button type="button" class="asset-select-sheet__close" aria-label="关闭" @click="close">×</button>
        <h2>{{ title }}</h2>

        <div class="asset-select-group">
          <h3>选择账户</h3>
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
          <h3>选择币种</h3>
          <label class="asset-select-search">
            <span>⌕</span>
            <input v-model="query" placeholder="币种搜索" />
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
            <span>可用 <b>{{ row.availableAmount }}</b></span>
            <em v-if="activeWalletType === selectedWalletType && row.coin === selectedCoin">✓</em>
          </button>
          <p v-if="!coins.length" class="asset-select-empty">暂无可选币种</p>
        </div>
      </section>
    </div>
  </Teleport>
</template>

<style scoped>
button,
input {
  border: 0;
  background: transparent;
  color: inherit;
  font: inherit;
}

.asset-select-overlay {
  position: fixed;
  inset: 0;
  z-index: 110;
  display: grid;
  align-items: end;
  background: rgba(0, 0, 0, 0.66);
  backdrop-filter: blur(12px);
}

.asset-select-sheet {
  position: relative;
  max-height: 84dvh;
  overflow: hidden auto;
  padding: 16px 26px calc(18px + env(safe-area-inset-bottom));
  border-radius: 24px 24px 0 0;
  background: #262832;
  color: #fff;
}

.asset-select-sheet__handle {
  display: block;
  width: 42px;
  height: 5px;
  margin: 4px auto 24px;
  border-radius: 999px;
  background: #a2a3aa;
}

.asset-select-sheet__close {
  position: absolute;
  top: 28px;
  right: 28px;
  font-size: 32px;
  line-height: 1;
}

.asset-select-sheet h2 {
  margin: 0 0 34px;
  font-size: 22px;
  font-weight: 800;
  text-align: center;
}

.asset-select-group {
  display: grid;
  gap: 16px;
  margin-bottom: 30px;
}

.asset-select-group h3 {
  margin: 0;
  font-size: 17px;
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
  font-size: 18px;
  font-weight: 800;
}

.asset-select-accounts button.active {
  background: #02b904;
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
  color: #fff;
  font-size: 26px;
}

.asset-select-search input {
  min-width: 0;
  outline: 0;
  font-size: 17px;
}

.asset-select-search input::placeholder {
  color: #9b9da6;
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
  font-size: 18px;
}

.asset-select-row span {
  color: #fff;
  font-size: 15px;
}

.asset-select-row b {
  font-size: 19px;
}

.asset-select-row em {
  color: #02b904;
  font-size: 22px;
  font-style: normal;
  font-weight: 800;
  text-align: right;
}

.asset-select-empty {
  margin: 26px;
  color: #9b9da6;
  font-size: 14px;
  text-align: center;
}
</style>
