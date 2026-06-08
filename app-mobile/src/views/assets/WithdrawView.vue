<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { apiGetAssetOptions, apiGetMyAssetSummary, apiListAssetCoinConfigs } from '@/api/asset'
import { apiCreateWithdrawOrder } from '@/api/payment'
import AssetCoinSelectSheet from '@/components/assets/AssetCoinSelectSheet.vue'
import AssetCoinPicker from '@/components/assets/AssetCoinPicker.vue'
import AssetPrimaryButton from '@/components/assets/AssetPrimaryButton.vue'
import CommonPage from '@/components/common/CommonPage.vue'
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
const router = useRouter()
const assetOptions = useOptions(apiGetAssetOptions)
const { t } = useI18n()
const coinConfigs = ref<AssetCoinConfig[]>([])
const assets = ref<AssetUserAsset[]>([])
const amount = ref('')
const address = ref('')
const bankName = ref('')
const bankAccountName = ref('')
const bankAccountNo = ref('')
const bankBranchName = ref('')
const submitLoading = ref(false)
const pageError = ref('')
const pageTip = ref('')
const selectedConfig = ref<AssetCoinConfig | null>(null)
const coinSheetVisible = ref(false)

const walletType = computed(() => Number(route.query.walletType || 1))
const isBankWithdraw = computed(() => route.query.method === 'bank')
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
const withdrawTargetCurrency = computed(() => (isBankWithdraw.value ? 'USD' : coin.value))

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
  const withdrawAddress = isBankWithdraw.value ? bankAccountNo.value.trim() : address.value.trim()
  if (!withdrawAddress) {
    pageError.value = t('assetFlow.inputWithdrawAddress')
    return
  }
  if (!withdrawAmountText || withdrawAmount <= 0) {
    pageError.value = t('assetFlow.invalidWithdrawAmount', {
      places: selectedInputDecimalPlaces.value,
    })
    return
  }

  if (isBankWithdraw.value && !bankName.value.trim()) {
    pageError.value = t('assetFlow.inputBankName')
    return
  }

  submitLoading.value = true
  try {
    const resp = await apiCreateWithdrawOrder({
      amount: withdrawAmount,
      currency: coin.value,
      address: withdrawAddress,
      bankId: 0,
      remark: isBankWithdraw.value
        ? `bank:${bankName.value.trim()};accountName:${bankAccountName.value.trim()};branch:${bankBranchName.value.trim()};target:${withdrawTargetCurrency.value}`
        : selectedChain.value
          ? `chain:${selectedChain.value}`
          : '',
    })
    if (isSuccessCode(resp.code)) {
      pageTip.value = resp.data
        ? t('assetFlow.withdrawSubmittedWithId', { id: resp.data })
        : t('assetFlow.withdrawSubmitted')
      amount.value = ''
      address.value = ''
      bankName.value = ''
      bankAccountName.value = ''
      bankAccountNo.value = ''
      bankBranchName.value = ''
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
  <CommonPage
    :title="t('assetFlow.withdraw')"
    :right-text="t('assetFlow.records')"
    :nav-height="76"
    @back="router.back()"
    @right-click="router.push({ name: 'asset-flows', query: { tab: 'withdraw' } })"
  >
    <section class="asset-flow-content">
      <button type="button" class="asset-type-pill">
        {{ isBankWithdraw ? t('assetFlow.bankCard') : t('assetFlow.crypto') }}
      </button>

      <template v-if="isBankWithdraw">
        <label class="field-block">
          <span class="field-block__row">
            <span>{{ t('assetFlow.withdrawAmount') }}</span>
            <small>{{ t('assetFlow.withdrawable') }} <b>{{ availableAmount }}</b> {{ coin }}</small>
          </span>
          <span class="asset-input asset-input--unit">
            <input
              v-model="amount"
              :placeholder="t('assetFlow.amountPlaceholder')"
              inputmode="decimal"
            >
            <i>{{ coin }}</i>
          </span>
        </label>

        <div class="bank-currency-grid">
          <label class="field-block">
            <span>{{ t('assetFlow.withdrawCoin') }}</span>
            <span class="bank-select-box">
              <span class="bank-select-box__coin">₮</span>
              <strong>{{ coin }}</strong>
              <i>›</i>
            </span>
          </label>
          <label class="field-block">
            <span>{{ t('assetFlow.arrivalCurrency') }}</span>
            <span class="bank-select-box">
              <span class="bank-select-box__flag">🇺🇸</span>
              <strong>{{ withdrawTargetCurrency }}</strong>
              <i>›</i>
            </span>
          </label>
        </div>

        <div class="bank-withdraw-divider" />

        <label class="field-block">
          <span>{{ t('assetFlow.bankName') }}</span>
          <input v-model="bankName" class="asset-input" :placeholder="t('assetFlow.inputEllipsis')">
        </label>

        <label class="field-block">
          <span>{{ t('assetFlow.payeeName') }}</span>
          <input
            v-model="bankAccountName"
            class="asset-input"
            :placeholder="t('assetFlow.identityUnverified')"
          >
        </label>

        <label class="field-block">
          <span>{{ t('assetFlow.bankCardNo') }}</span>
          <input
            v-model="bankAccountNo"
            class="asset-input"
            :placeholder="t('assetFlow.inputEllipsis')"
          >
        </label>

        <label class="field-block">
          <span>{{ t('assetFlow.openingBank') }}</span>
          <input
            v-model="bankBranchName"
            class="asset-input"
            :placeholder="t('assetFlow.inputEllipsis')"
          >
        </label>
      </template>

      <template v-else>
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
            <input v-model="address" :placeholder="t('assetFlow.addressPlaceholder')">
            <i>▣</i>
          </span>
        </label>

        <label class="field-block">
          <span class="field-block__row">
            <span>{{ t('assetFlow.withdrawAmount') }}</span>
            <small>{{ t('assetFlow.withdrawable') }} <b>{{ availableAmount }}</b> {{ coin }}</small>
          </span>
          <input v-model="amount" class="asset-input" inputmode="decimal">
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
      </template>

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
    </section>

    <AssetCoinSelectSheet
      v-model="coinSheetVisible"
      :configs="coinConfigs"
      :selected-config="selectedConfig"
      :operation-type="2"
      @select="selectConfig"
    />
  </CommonPage>
</template>

<style scoped>
button,
input {
  border: 0;
  background: transparent;
  color: inherit;
  font: inherit;
}

.asset-flow-content {
  min-height: calc(100dvh - 76px);
  padding: 20px 18px 56px;
  background: #0b0c15;
  color: #f8f8fb;
}

:deep(.header-bar) {
  background: #0b0c15;
}

:deep(.header-left) {
  left: 18px;
  justify-content: center;
  width: 36px;
  height: 36px;
  margin-top: 20px;
  border-radius: 50%;
  background: #242633;
}

:deep(.header-title) {
  font-size: 18px;
  font-weight: 800;
}

:deep(.header-right) {
  right: 18px;
  color: #fff;
  font-size: 14px;
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

.asset-input--unit {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  align-items: center;
  gap: 12px;
}

.asset-input--unit input {
  min-width: 0;
  outline: 0;
}

.asset-input--unit i {
  color: #c9cbd3;
  font-style: normal;
  font-weight: 800;
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

.bank-currency-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
  margin-bottom: 18px;
}

.bank-currency-grid .field-block {
  margin-bottom: 0;
}

.bank-select-box {
  display: flex;
  align-items: center;
  gap: 10px;
  min-height: 58px;
  padding: 0 12px;
  border-radius: 14px;
  background: #292b36;
}

.bank-select-box strong {
  min-width: 0;
  flex: 1;
  overflow: hidden;
  font-size: 20px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.bank-select-box i {
  color: #9699a2;
  font-size: 24px;
  font-style: normal;
}

.bank-select-box__coin,
.bank-select-box__flag {
  display: inline-grid;
  flex: none;
  width: 28px;
  height: 28px;
  place-items: center;
  border-radius: 50%;
  overflow: hidden;
}

.bank-select-box__coin {
  background: linear-gradient(135deg, #21b78c, #129f75);
  color: #fff;
  font-size: 18px;
  font-weight: 900;
}

.bank-select-box__flag {
  font-size: 27px;
  line-height: 1;
}

.bank-withdraw-divider {
  height: 1px;
  margin: 8px 0 24px;
  background: rgba(255, 255, 255, 0.08);
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
