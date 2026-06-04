<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'

import { apiGetAssetOptions, apiGetMyAssetSummary, apiListAssetCoinConfigs } from '@/api/asset'
import { apiCreateWithdrawOrder } from '@/api/payment'
import AssetCoinSelectSheet from '@/components/assets/AssetCoinSelectSheet.vue'
import AssetCoinPicker from '@/components/assets/AssetCoinPicker.vue'
import AssetFlowLayout from '@/components/assets/AssetFlowLayout.vue'
import AssetPrimaryButton from '@/components/assets/AssetPrimaryButton.vue'
import { useOptions } from '@/composables/useOptions'
import { useI18n } from '@/i18n'
import type { AssetCoinConfig, AssetUserAsset } from '@/types/asset'
import {
  formatAssetDecimalAmount,
  formatAssetMinorAmount,
  normalizeAssetDecimalPlaces,
  normalizeAssetInputDecimalPlaces,
  parseAssetDecimalToMinorText,
} from '@/utils/assetAmount'

const route = useRoute()
const assetOptions = useOptions(apiGetAssetOptions)
const { t } = useI18n()
const coinConfigs = ref<AssetCoinConfig[]>([])
const assets = ref<AssetUserAsset[]>([])
const amount = ref('')
const address = ref('')
const submitLoading = ref(false)
const pageError = ref('')
const pageTip = ref('')
const selectedConfig = ref<AssetCoinConfig | null>(null)
const coinSheetVisible = ref(false)

const walletType = computed(() => Number(route.query.walletType || 1))
const routeCoin = computed(() => String(route.query.coin || 'USDT'))
const coin = computed(() => selectedConfig.value?.coin || routeCoin.value)
const selectedDecimalPlaces = computed(() =>
  normalizeAssetDecimalPlaces(selectedConfig.value?.decimalPlaces),
)
const selectedInputDecimalPlaces = computed(() =>
  normalizeAssetInputDecimalPlaces(selectedConfig.value?.decimalPlaces),
)
const selectedChain = computed(() => {
  const config = selectedConfig.value
  if (!config) return ''
  return getChainLabel(config)
})
const rawAvailableAmount = computed(() => {
  return (
    assets.value.find((asset) => asset.walletType === walletType.value && asset.coin === coin.value)
      ?.availableAmount || '0'
  )
})
const availableAmount = computed(() =>
  formatAssetMinorAmount(rawAvailableAmount.value, selectedDecimalPlaces.value),
)
const feeAmount = computed(() => formatAssetDecimalAmount('0', selectedDecimalPlaces.value))
const receivedAmount = computed(() => {
  if (!amount.value.trim()) return formatAssetDecimalAmount('0', selectedDecimalPlaces.value)
  return formatAssetDecimalAmount(amount.value, selectedDecimalPlaces.value)
})

function isSuccessCode(code: number) {
  return code === 0 || code === 200
}

function getChainLabel(config: AssetCoinConfig) {
  if (String(config.coin).toLocaleUpperCase() != 'USDT' || !config.chainCode) return ''
  const option = assetOptions.getGroup('chainCode').find((item) => item.value === config.chainCode)
  return option ? formatChainCode(option.code) : String(config.chainCode)
}

function formatChainCode(code: string) {
  return code.replace(/^CHAIN_CODE_/, '')
}

function syncSelectedConfig(configs: AssetCoinConfig[]) {
  selectedConfig.value =
    configs.find(
      (config) =>
        config.coin === routeCoin.value && config.id === Number(route.query.coinConfigId || 0),
    ) ||
    configs.find((config) => config.coin === routeCoin.value) ||
    configs[0] ||
    null
}

async function loadPageData() {
  try {
    const [configsResp, summaryResp] = await Promise.all([
      apiListAssetCoinConfigs({
        walletType: walletType.value,
        operationType: 2,
      }),
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
  pageError.value = ''
  pageTip.value = ''
}

async function submitWithdraw() {
  if (submitLoading.value) return

  pageError.value = ''
  pageTip.value = ''
  const withdrawAmountText = parseAssetDecimalToMinorText(amount.value, selectedDecimalPlaces.value)
  const withdrawAmount = Number(withdrawAmountText)
  if (!coin.value) {
    pageError.value = t('assetFlow.selectWithdrawCoin')
    return
  }
  if (!address.value.trim()) {
    pageError.value = t('assetFlow.inputWithdrawAddress')
    return
  }
  if (!withdrawAmountText || withdrawAmount <= 0) {
    pageError.value = t('assetFlow.invalidWithdrawAmount', {
      places: selectedInputDecimalPlaces.value,
    })
    return
  }

  submitLoading.value = true
  try {
    const resp = await apiCreateWithdrawOrder({
      amount: withdrawAmount,
      currency: coin.value,
      address: address.value.trim(),
      bankId: 0,
      remark: selectedChain.value ? `chain:${selectedChain.value}` : '',
    })
    if (isSuccessCode(resp.code)) {
      pageTip.value = resp.data
        ? t('assetFlow.withdrawSubmittedWithId', { id: resp.data })
        : t('assetFlow.withdrawSubmitted')
      amount.value = ''
      address.value = ''
      await loadPageData()
    } else {
      pageError.value = resp.msg || t('assetFlow.withdrawFailedLater')
    }
  } catch (error) {
    console.warn('create withdraw order failed', error)
    pageError.value = t('assetFlow.withdrawFailedLater')
  } finally {
    submitLoading.value = false
  }
}

onMounted(() => {
  void loadPageData()
})
</script>

<template>
  <AssetFlowLayout
    :title="t('assetFlow.withdraw')"
    :right-text="t('assetFlow.records')"
    :right-to="{ name: 'asset-flows', query: { tab: 'withdraw' } }"
    narrow
  >
    <button type="button" class="asset-type-pill">{{ t('assetFlow.crypto') }}</button>

    <label class="field-block">
      <span>{{ t('assetFlow.coin') }}</span>
      <AssetCoinPicker
        :coin="coin"
        :config="selectedConfig || undefined"
        :chain="selectedChain"
        @click="coinSheetVisible = true"
      />
    </label>

    <label class="field-block">
      <span>{{ t('assetFlow.withdrawAddress') }}</span>
      <span class="asset-input asset-input--address">
        <input v-model="address" :placeholder="t('assetFlow.addressPlaceholder')" />
        <i>▣</i>
      </span>
    </label>

    <label class="field-block">
      <span class="field-block__row">
        <span>{{ t('assetFlow.withdrawAmount') }}</span>
        <small
          >{{ t('assetFlow.withdrawable') }} <b>{{ availableAmount }}</b> {{ coin }}</small
        >
      </span>
      <input v-model="amount" class="asset-input" inputmode="decimal" />
    </label>

    <dl class="withdraw-summary">
      <div>
        <dt>{{ t('assetFlow.fee') }}</dt>
        <dd>{{ feeAmount }} {{ coin }}</dd>
      </div>
      <div>
        <dt>{{ t('assetFlow.receivedAmount') }}</dt>
        <dd>{{ receivedAmount }} {{ coin }}</dd>
      </div>
    </dl>

    <p v-if="pageError" class="state-text state-text--error">
      {{ pageError }}
    </p>
    <p v-if="pageTip" class="state-text state-text--success">
      {{ pageTip }}
    </p>

    <AssetPrimaryButton
      :label="submitLoading ? t('common.submitting') : t('assetFlow.withdraw')"
      @click="submitWithdraw"
    />

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

.state-text {
  margin: -10px 0 16px;
  font-size: 13px;
  font-weight: 800;
  line-height: 1.5;
  text-align: center;
}

.state-text--error {
  color: #ff7676;
}

.state-text--success {
  color: #02d107;
}
</style>
