<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'

import { apiGetMyAssetSummary, apiListAssetCoinConfigs } from '@/api/asset'
import AssetCoinSelectSheet from '@/components/assets/AssetCoinSelectSheet.vue'
import AssetCoinPicker from '@/components/assets/AssetCoinPicker.vue'
import AssetFlowLayout from '@/components/assets/AssetFlowLayout.vue'
import AssetPrimaryButton from '@/components/assets/AssetPrimaryButton.vue'
import type { AssetCoinConfig, AssetUserAsset } from '@/types/asset'

const route = useRoute()
const coinConfigs = ref<AssetCoinConfig[]>([])
const assets = ref<AssetUserAsset[]>([])
const amount = ref('')
const selectedConfig = ref<AssetCoinConfig | null>(null)
const coinSheetVisible = ref(false)

const walletType = computed(() => Number(route.query.walletType || 1))
const routeCoin = computed(() => String(route.query.coin || 'USDT'))
const coin = computed(() => selectedConfig.value?.coin || routeCoin.value)
const selectedChain = computed(() => {
  const config = selectedConfig.value
  if (!config) return routeCoin.value === 'USDT' ? 'TRC20' : ''
  return getChainLabel(config)
})
const availableAmount = computed(() => {
  return assets.value.find((asset) => asset.walletType === walletType.value && asset.coin === coin.value)?.availableAmount || '0'
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

async function loadPageData() {
  try {
    const [configsResp, summaryResp] = await Promise.all([
      apiListAssetCoinConfigs({ walletType: walletType.value, operationType: 2 }),
      apiGetMyAssetSummary({}),
    ])
    if (isSuccessCode(configsResp.code)) {
      coinConfigs.value = configsResp.data || []
      syncSelectedConfig(coinConfigs.value)
    }
    if (isSuccessCode(summaryResp.code)) assets.value = summaryResp.data?.assets || []
  } catch (error) {
    console.warn('load withdraw data failed', error)
  }
}

function selectConfig(config: AssetCoinConfig) {
  selectedConfig.value = config
}

onMounted(() => {
  void loadPageData()
})
</script>

<template>
  <AssetFlowLayout title="提现" right-text="资金记录" narrow>
    <button type="button" class="asset-type-pill">加密货币</button>

    <label class="field-block">
      <span>币种</span>
      <AssetCoinPicker
        :coin="coin"
        :config="selectedConfig || undefined"
        :chain="selectedChain"
        @click="coinSheetVisible = true"
      />
    </label>

    <label class="field-block">
      <span>提现地址</span>
      <span class="asset-input asset-input--address">
        <input placeholder="选择或输入地址" />
        <i>▣</i>
      </span>
    </label>

    <label class="field-block">
      <span class="field-block__row">
        <span>提现金额</span>
        <small>可提现 <b>{{ availableAmount }}</b> {{ coin }}</small>
      </span>
      <input v-model="amount" class="asset-input" inputmode="decimal" />
    </label>

    <dl class="withdraw-summary">
      <div>
        <dt>手续费</dt>
        <dd>0 {{ coin }}</dd>
      </div>
      <div>
        <dt>到账金额</dt>
        <dd>{{ amount || 0 }} {{ coin }}</dd>
      </div>
    </dl>

    <AssetPrimaryButton label="提现" />

    <AssetCoinSelectSheet
      v-model="coinSheetVisible"
      :configs="coinConfigs"
      :selected-config="selectedConfig"
      :operation-type="2"
      @select="selectConfig"
    />
  </AssetFlowLayout>
</template>

<style scoped>
button,
input {
  border: 0;
  background: transparent;
  color: inherit;
  font: inherit;
}

.asset-type-pill {
  min-width: 104px;
  min-height: 36px;
  margin-bottom: 22px;
  border-radius: 999px;
  background: #02b904;
  font-size: 15px;
  font-weight: 800;
}

.field-block {
  display: grid;
  gap: 9px;
  margin-bottom: 18px;
  font-size: 15px;
  font-weight: 700;
}

.field-block__row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.field-block small {
  color: #9b9da6;
  font-size: 12px;
  font-weight: 700;
}

.field-block b {
  color: #fff;
}

.asset-input {
  width: 100%;
  min-height: 56px;
  padding: 0 16px;
  border-radius: 14px;
  outline: 0;
  background: #292b36;
}

.asset-input--address {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 28px;
  align-items: center;
}

.asset-input--address input {
  min-width: 0;
  outline: 0;
}

.asset-input--address i {
  color: #9699a2;
  font-style: normal;
}

.asset-input::placeholder,
.asset-input--address input::placeholder {
  color: #8f929d;
}

.withdraw-summary {
  display: grid;
  gap: 14px;
  margin: 0 0 28px;
}

.withdraw-summary div {
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 16px;
}

.withdraw-summary dt {
  color: #9b9da6;
  font-size: 14px;
  font-weight: 700;
}

.withdraw-summary dd {
  margin: 0;
  font-size: 14px;
  font-weight: 800;
}
</style>
