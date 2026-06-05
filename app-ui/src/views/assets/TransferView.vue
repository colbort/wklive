<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'

import {
  apiGetAssetOptions,
  apiGetMyAssetSummary,
  apiListAssetCoinConfigs,
  apiTransferMyAsset,
} from '@/api/asset'
import AssetCoinIcon from '@/components/assets/AssetCoinIcon.vue'
import AssetFlowLayout from '@/components/assets/AssetFlowLayout.vue'
import AssetPrimaryButton from '@/components/assets/AssetPrimaryButton.vue'
import AssetTransferSelectSheet from '@/components/assets/AssetTransferSelectSheet.vue'
import { optionText, useOptions } from '@/composables/useOptions'
import { useI18n } from '@/i18n'
import type { AssetCoinConfig, AssetUserAsset } from '@/types/asset'
import {
  compareDecimalText,
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
const submitLoading = ref(false)
const pageError = ref('')
const pageTip = ref('')
const pickerTarget = ref<'from' | 'to' | ''>('')
const fromWalletType = ref(Number(route.query.walletType || 1))
const toWalletType = ref(Number(route.query.walletType || 1))
const fromSelectedCoin = ref('')
const toSelectedCoin = ref('')

const direction = computed(() => String(route.query.direction || 'out'))
const queryCoin = computed(() => String(route.query.coin || 'USDT'))
const fromCoin = computed(() => fromSelectedCoin.value)
const toCoin = computed(() => toSelectedCoin.value)
const fromConfig = computed(() =>
  coinConfigs.value.find(
    (config) => config.walletType === fromWalletType.value && config.coin === fromCoin.value,
  ),
)
const toConfig = computed(() =>
  coinConfigs.value.find(
    (config) => config.walletType === toWalletType.value && config.coin === toCoin.value,
  ),
)
const fromDecimalPlaces = computed(() =>
  normalizeAssetDecimalPlaces(fromConfig.value?.decimalPlaces),
)
const fromInputDecimalPlaces = computed(() =>
  normalizeAssetInputDecimalPlaces(fromConfig.value?.decimalPlaces),
)
const pickerWalletType = computed(() =>
  pickerTarget.value === 'to' ? toWalletType.value : fromWalletType.value,
)
const pickerCoin = computed(() => (pickerTarget.value === 'to' ? toCoin.value : fromCoin.value))
const pickerTitle = computed(() =>
  pickerTarget.value === 'to' ? t('assetFlow.chooseToAccount') : t('assetFlow.chooseFromAccount'),
)
const pickerVisible = computed(() => Boolean(pickerTarget.value))
const rawAvailableAmount = computed(() => {
  if (!fromCoin.value) return '0'
  return (
    assets.value.find(
      (asset) => asset.walletType === fromWalletType.value && asset.coin === fromCoin.value,
    )?.availableAmount || '0'
  )
})
const availableAmount = computed(() =>
  formatAssetMinorAmount(rawAvailableAmount.value, fromDecimalPlaces.value),
)
const walletTypes = computed(() => {
  const options = assetOptions
    .getGroup('walletType')
    .filter((option) => option.value > 0)
    .map((option) => ({ value: option.value, label: optionText(option) }))

  return options.length
    ? options
    : [
        { value: 1, label: t('options.WALLET_TYPE_SPOT') },
        { value: 2, label: t('options.WALLET_TYPE_FUNDING') },
        { value: 3, label: t('options.WALLET_TYPE_CONTRACT') },
        { value: 4, label: t('options.WALLET_TYPE_EARN') },
        { value: 5, label: t('options.WALLET_TYPE_OPTION') },
      ]
})
const fromAccountLabel = computed(() => accountLabel(fromWalletType.value))
const toAccountLabel = computed(() => accountLabel(toWalletType.value))
const fromPlaceholder = computed(() => placeholderAccountLabel(fromWalletType.value))
const toPlaceholder = computed(() => placeholderAccountLabel(toWalletType.value))

function accountLabel(walletType: number) {
  return (
    walletTypes.value.find((account) => account.value === walletType)?.label ||
    t('options.WALLET_TYPE_SPOT')
  )
}

function placeholderAccountLabel(walletType: number) {
  return (
    walletTypes.value.find((account) => account.value !== walletType)?.label ||
    t('assetFlow.chooseAccount')
  )
}

function firstOtherWalletType(walletType: number) {
  return walletTypes.value.find((account) => account.value !== walletType)?.value || walletType
}

function isSuccessCode(code: number) {
  return code === 0 || code === 200
}

function coinKey(coin: string) {
  return String(coin || '').toUpperCase()
}

function isSameTransferCoin(
  walletType: number,
  coin: string,
  targetWalletType: number,
  targetCoin: string,
) {
  return walletType === targetWalletType && Boolean(coin) && coinKey(coin) === coinKey(targetCoin)
}

function openPicker(target: 'from' | 'to') {
  pickerTarget.value = target
  pageError.value = ''
  pageTip.value = ''
}

function selectTransferAsset(payload: { walletType: number; coin: string }) {
  const coin = payload.coin
  if (pickerTarget.value === 'to') {
    if (isSameTransferCoin(payload.walletType, coin, fromWalletType.value, fromCoin.value)) {
      pageError.value = t('assetFlow.sameCoinNoNeedTransfer')
      return
    }
    toWalletType.value = payload.walletType
    toSelectedCoin.value = coin
  } else {
    if (isSameTransferCoin(payload.walletType, coin, toWalletType.value, toCoin.value)) {
      pageError.value = t('assetFlow.sameCoinNoNeedTransfer')
      return
    }
    fromWalletType.value = payload.walletType
    fromSelectedCoin.value = coin
  }
  pageError.value = ''
  pageTip.value = ''
}

function updatePickerVisible(visible: boolean) {
  if (!visible) pickerTarget.value = ''
}

async function loadPageData() {
  try {
    const configRequests = walletTypes.value.map((account) =>
      apiListAssetCoinConfigs({ walletType: account.value, operationType: 3 }),
    )
    const [summaryResp, ...configResponses] = await Promise.all([
      apiGetMyAssetSummary({}),
      ...configRequests,
    ])
    if (isSuccessCode(summaryResp.code)) {
      assets.value = summaryResp.data?.assets || []
    }
    coinConfigs.value = configResponses.flatMap((resp) =>
      isSuccessCode(resp.code) ? resp.data || [] : [],
    )
  } catch (error) {
    console.warn('load transfer data failed', error)
  }
}

async function submitTransfer() {
  if (submitLoading.value) return

  pageError.value = ''
  pageTip.value = ''

  const transferAmount = parseAssetDecimalToMinorText(amount.value, fromDecimalPlaces.value)
  if (!fromCoin.value || !toCoin.value) {
    pageError.value = t('assetFlow.selectTransferCoin')
    return
  }
  if (isSameTransferCoin(fromWalletType.value, fromCoin.value, toWalletType.value, toCoin.value)) {
    pageError.value = t('assetFlow.sameCoinNoNeedTransfer')
    return
  }
  if (!transferAmount || transferAmount === '0') {
    pageError.value = t('assetFlow.invalidTransferAmount', {
      places: fromInputDecimalPlaces.value,
    })
    return
  }
  if (compareDecimalText(transferAmount, rawAvailableAmount.value) > 0) {
    pageError.value = t('assetFlow.insufficientBalance')
    return
  }

  submitLoading.value = true
  try {
    const resp = await apiTransferMyAsset({
      fromWalletType: fromWalletType.value,
      toWalletType: toWalletType.value,
      fromCoin: fromCoin.value,
      toCoin: toCoin.value,
      amount: transferAmount,
    })
    if (isSuccessCode(resp.code)) {
      pageTip.value = t('assetFlow.transferSuccess')
      amount.value = ''
      await loadPageData()
    } else {
      pageError.value = resp.msg || t('assetFlow.transferFailedLater')
    }
  } catch (error) {
    console.warn('transfer asset failed', error)
    pageError.value = t('assetFlow.transferFailedLater')
  } finally {
    submitLoading.value = false
  }
}

onMounted(() => {
  const walletType = Number(route.query.walletType || 1)
  if (direction.value === 'in') {
    fromWalletType.value = firstOtherWalletType(walletType)
    toWalletType.value = walletType
    fromSelectedCoin.value = queryCoin.value
    toSelectedCoin.value = queryCoin.value
  } else {
    fromWalletType.value = walletType
    toWalletType.value = firstOtherWalletType(walletType)
    fromSelectedCoin.value = queryCoin.value
    toSelectedCoin.value = queryCoin.value
  }
  void loadPageData()
})
</script>

<template>
  <AssetFlowLayout :title="t('assetFlow.transfer')" narrow>
    <section class="transfer-card">
      <label class="transfer-field">
        <span class="transfer-field__head">
          <strong>{{ t('assetFlow.transferFrom') }}</strong>
          <small v-if="fromCoin">{{ t('assetFlow.available') }} <b>{{ availableAmount }}</b> {{ fromCoin }}</small>
        </span>
        <button type="button" class="transfer-picker" @click="openPicker('from')">
          <span>{{ fromAccountLabel }}</span>
          <i />
          <span v-if="fromCoin" class="transfer-picker__coin">
            <AssetCoinIcon :coin="fromCoin" :config="fromConfig" />
            <strong>{{ fromCoin }}</strong>
          </span>
          <em v-else>{{ fromPlaceholder }}</em>
          <b>⌄</b>
        </button>
      </label>

      <label class="transfer-field">
        <span class="transfer-field__head"><strong>{{ t('assetFlow.transferTo') }}</strong></span>
        <button type="button" class="transfer-picker" @click="openPicker('to')">
          <span>{{ toAccountLabel }}</span>
          <i />
          <span v-if="toCoin" class="transfer-picker__coin">
            <AssetCoinIcon :coin="toCoin" :config="toConfig" />
            <strong>{{ toCoin }}</strong>
          </span>
          <em v-else>{{ toPlaceholder }}</em>
          <b>⌄</b>
        </button>
      </label>
    </section>

    <label class="amount-field">
      <span>{{ t('assetFlow.transferAmount') }}</span>
      <span class="amount-input">
        <input v-model="amount" inputmode="decimal">
        <strong v-if="fromCoin">{{ fromCoin }}</strong>
      </span>
    </label>

    <p v-if="pageError" class="state-text state-text--error">
      {{ pageError }}
    </p>
    <p v-if="pageTip" class="state-text state-text--success">
      {{ pageTip }}
    </p>

    <AssetPrimaryButton
      class="transfer-button"
      :label="submitLoading ? t('assetFlow.transfering') : t('assetFlow.transfer')"
      @click="submitTransfer"
    />

    <AssetTransferSelectSheet
      :model-value="pickerVisible"
      :title="pickerTitle"
      :source="pickerTarget === 'to' ? 'configs' : 'assets'"
      :excluded-wallet-type="pickerTarget === 'to' ? fromWalletType : toWalletType"
      :excluded-coin="pickerTarget === 'to' ? fromCoin : toCoin"
      :wallet-types="walletTypes"
      :selected-wallet-type="pickerWalletType"
      :selected-coin="pickerCoin"
      :assets="assets"
      :configs="coinConfigs"
      @update:model-value="updatePickerVisible"
      @select="selectTransferAsset"
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

.transfer-card {
  display: grid;
  gap: 18px;
  padding: 18px 18px;
  border-radius: 18px;
  background: #292b36;
}

.transfer-field {
  display: grid;
  gap: 9px;
}

.transfer-field__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.transfer-field__head strong,
.amount-field > span:first-child {
  font-size: 15px;
  font-weight: 800;
}

.transfer-field__head small {
  color: #9b9da6;
  font-size: 12px;
  font-weight: 700;
}

.transfer-field__head b {
  color: #02b904;
}

.transfer-picker {
  display: grid;
  grid-template-columns: auto 1px minmax(0, 1fr) 18px;
  align-items: center;
  gap: 14px;
  width: 100%;
  min-height: 56px;
  padding: 0 16px;
  border-radius: 14px;
  background: #454751;
  text-align: left;
}

.transfer-picker > span:first-child {
  font-size: 15px;
  font-weight: 800;
}

.transfer-picker i {
  width: 1px;
  height: 32px;
  background: #666975;
}

.transfer-picker em {
  overflow: hidden;
  color: #9b9da6;
  font-size: 14px;
  font-style: normal;
  font-weight: 700;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.transfer-picker b {
  color: #c1c3ca;
  font-size: 15px;
}

.transfer-picker__coin {
  display: flex;
  align-items: center;
  min-width: 0;
  gap: 9px;
}

.transfer-picker__coin strong {
  overflow: hidden;
  font-size: 15px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.amount-field {
  display: grid;
  gap: 9px;
  margin-top: 24px;
}

.amount-input {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  align-items: center;
  min-height: 56px;
  padding: 0 16px;
  border-radius: 14px;
  background: #292b36;
}

.amount-input input {
  min-width: 0;
  outline: 0;
}

.amount-input strong {
  font-size: 15px;
}

.transfer-button {
  margin-top: 30px;
}

.state-text {
  margin: 14px 0 0;
  font-size: 13px;
  font-weight: 700;
}

.state-text--error {
  color: #ff6b6b;
}

.state-text--success {
  color: #02b904;
}
</style>
