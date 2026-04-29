<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'

import { apiListAssetCoinConfigs } from '@/api/asset'
import AssetCoinSelectSheet from '@/components/assets/AssetCoinSelectSheet.vue'
import AssetCoinPicker from '@/components/assets/AssetCoinPicker.vue'
import AssetFlowLayout from '@/components/assets/AssetFlowLayout.vue'
import AssetPrimaryButton from '@/components/assets/AssetPrimaryButton.vue'
import type { AssetCoinConfig } from '@/types/asset'

const route = useRoute()
const coinConfigs = ref<AssetCoinConfig[]>([])
const selectedConfig = ref<AssetCoinConfig | null>(null)
const coinSheetVisible = ref(false)

const walletType = computed(() => Number(route.query.walletType || 1))
const routeCoin = computed(() => String(route.query.coin || 'USDT'))
const selectedCoin = computed(() => selectedConfig.value?.coin || routeCoin.value)
const selectedChain = computed(() => {
  const config = selectedConfig.value
  if (!config) return routeCoin.value === 'USDT' ? 'TRC20' : ''
  return getChainLabel(config)
})

const chainLabels: Record<number, string> = {
  1: 'BTC',
  2: 'ETH',
  3: 'TRX',
  4: 'BSC',
  5: 'SOL',
  6: 'POLYGON',
  20: 'TRC20',
  21: 'ERC20',
  22: 'BEP20',
}

function isSuccessCode(code: number) {
  return code === 0 || code === 200
}

function getChainLabel(config: AssetCoinConfig) {
  if (config.chainCode) return chainLabels[config.chainCode] || String(config.chainCode)
  return config.coin === 'USDT' ? 'TRC20' : ''
}

function syncSelectedConfig(configs: AssetCoinConfig[]) {
  selectedConfig.value =
    configs.find((config) => config.coin === routeCoin.value && config.id === Number(route.query.coinConfigId || 0)) ||
    configs.find((config) => config.coin === routeCoin.value) ||
    configs[0] ||
    null
}

async function loadCoinConfigs() {
  try {
    const resp = await apiListAssetCoinConfigs({
      walletType: walletType.value,
      operationType: 1,
    })
    if (isSuccessCode(resp.code)) {
      coinConfigs.value = resp.data || []
      syncSelectedConfig(coinConfigs.value)
    }
  } catch (error) {
    console.warn('load recharge coin configs failed', error)
  }
}

function selectConfig(config: AssetCoinConfig) {
  selectedConfig.value = config
}

onMounted(() => {
  void loadCoinConfigs()
})
</script>

<template>
  <AssetFlowLayout title="充值" right-text="资金记录" narrow>
    <h2>支付方式</h2>
    <AssetCoinPicker
      :coin="selectedCoin"
      :config="selectedConfig || undefined"
      :chain="selectedChain"
      @click="coinSheetVisible = true"
    />
    <AssetPrimaryButton class="recharge-button" label="充值" />

    <AssetCoinSelectSheet
      v-model="coinSheetVisible"
      :configs="coinConfigs"
      :selected-config="selectedConfig"
      :operation-type="1"
      @select="selectConfig"
    />
  </AssetFlowLayout>
</template>

<style scoped>
h2 {
  margin: 0 0 14px;
  font-size: 15px;
  font-weight: 700;
}

.recharge-button {
  margin-top: 30px;
}
</style>
