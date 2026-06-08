<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import QRCode from 'qrcode'

import { apiGetAssetOptions, apiListAssetCoinConfigs } from '@/api/asset'
import { apiCreateCryptoRechargeOrder, apiGetMyCryptoRechargeAddress } from '@/api/payment'
import { apiUploadFile } from '@/api/upload'
import AssetCoinSelectSheet from '@/components/assets/AssetCoinSelectSheet.vue'
import AssetCoinPicker from '@/components/assets/AssetCoinPicker.vue'
import AssetPrimaryButton from '@/components/assets/AssetPrimaryButton.vue'
import CommonPage from '@/components/common/CommonPage.vue'
import { useOptions } from '@/composables/useOptions'
import { useI18n } from '@/i18n'
import type { AssetCoinConfig } from '@/types/asset'
import type { CryptoRechargeAddress } from '@/types/payment'
import {
  normalizeAssetDecimalPlaces,
  normalizeAssetInputDecimalPlaces,
  parseAssetDecimalToMinorText,
} from '@/utils/assetAmount'

const route = useRoute()
const router = useRouter()
const assetOptions = useOptions(apiGetAssetOptions)
const { t } = useI18n()
const coinConfigs = ref<AssetCoinConfig[]>([])
const selectedConfig = ref<AssetCoinConfig | null>(null)
const coinSheetVisible = ref(false)
const addressLoading = ref(false)
const rechargeAddress = ref<CryptoRechargeAddress | null>(null)
const pageError = ref('')
const copyTip = ref('')
const amount = ref('')
const voucherName = ref('')
const voucherFile = ref<File | null>(null)
const voucherPreviewUrl = ref('')
const submitLoading = ref(false)
const addressSecondsLeft = ref(0)
const qrImageUrl = ref('')
let addressTimer: ReturnType<typeof setInterval> | undefined
const fileInputRef = ref<HTMLInputElement | null>(null)
const step = ref<'select' | 'detail'>('select')

const walletType = computed(() => Number(route.query.walletType || 1))
const routeCoin = computed(() => String(route.query.coin || 'USDT'))
const selectedCoin = computed(() => selectedConfig.value?.coin || routeCoin.value)
const selectedDecimalPlaces = computed(() =>
  normalizeAssetDecimalPlaces(selectedConfig.value?.decimalPlaces),
)
const selectedInputDecimalPlaces = computed(() =>
  normalizeAssetInputDecimalPlaces(selectedConfig.value?.decimalPlaces),
)
const selectedChainCode = computed(
  () => selectedConfig.value?.chainCode || (selectedCoin.value === 'USDT' ? 20 : 0),
)
const selectedChain = computed(() => {
  const config = selectedConfig.value
  if (!config) return routeCoin.value === 'USDT' ? 'TRC20' : ''
  return getChainLabel(config)
})
const addressCountdownText = computed(() => {
  const seconds = addressSecondsLeft.value
  const minutes = Math.floor(seconds / 60)
  const remainSeconds = String(seconds % 60).padStart(2, '0')
  return `${minutes}:${remainSeconds}`
})

function isSuccessCode(code: number) {
  return code === 0 || code === 200
}

function getChainLabel(config: AssetCoinConfig) {
  if (config.chainCode) {
    const option = assetOptions
      .getGroup('chainCode')
      .find((item) => item.value === config.chainCode)
    return option ? formatChainCode(option.code) : String(config.chainCode)
  }
  return config.coin === 'USDT' ? 'TRC20' : ''
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
  rechargeAddress.value = null
  qrImageUrl.value = ''
  stopAddressCountdown()
  pageError.value = ''
  copyTip.value = ''
  step.value = 'select'
  coinSheetVisible.value = false
}

function stopAddressCountdown() {
  if (addressTimer) {
    clearInterval(addressTimer)
    addressTimer = undefined
  }
  addressSecondsLeft.value = 0
}

function startAddressCountdown() {
  stopAddressCountdown()
  addressSecondsLeft.value = 180
  addressTimer = setInterval(() => {
    addressSecondsLeft.value -= 1
    if (addressSecondsLeft.value > 0) return

    stopAddressCountdown()
    rechargeAddress.value = null
    qrImageUrl.value = ''
    copyTip.value = ''
    pageError.value = t('assetFlow.expiredAddress')
    step.value = 'select'
  }, 1000)
}

async function startRecharge() {
  pageError.value = ''
  copyTip.value = ''
  if (!selectedCoin.value || !selectedChainCode.value) {
    pageError.value = t('assetFlow.selectRechargeCoinNetwork')
    return
  }

  addressLoading.value = true
  try {
    const resp = await apiGetMyCryptoRechargeAddress({
      walletType: walletType.value,
      coin: selectedCoin.value,
      chainCode: selectedChainCode.value,
    })
    if (isSuccessCode(resp.code) && resp.data?.address) {
      rechargeAddress.value = resp.data
      qrImageUrl.value = await createQrDataUrl(resp.data.address)
      startAddressCountdown()
      step.value = 'detail'
    } else {
      pageError.value = resp.msg || t('assetFlow.noRechargeAddress')
    }
  } catch (error) {
    console.warn('load crypto recharge address failed', error)
    pageError.value = t('assetFlow.rechargeAddressLoadFailed')
  } finally {
    addressLoading.value = false
  }
}

async function createQrDataUrl(text: string) {
  return QRCode.toDataURL(text, {
    errorCorrectionLevel: 'M',
    margin: 1,
    width: 220,
    color: {
      dark: '#000000',
      light: '#ffffff',
    },
  })
}

async function copyText(text: string) {
  if (!text) return
  try {
    await navigator.clipboard.writeText(text)
    copyTip.value = t('assetFlow.copySuccess')
  } catch (error) {
    console.warn('copy failed', error)
    copyTip.value = t('assetFlow.copyFailed')
  }
}

function handleVoucherChange(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0] || null
  if (!file) {
    clearVoucherPreview()
    voucherFile.value = null
    voucherName.value = ''
    return
  }
  if (!file.type.startsWith('image/')) {
    clearVoucherPreview()
    voucherFile.value = null
    voucherName.value = ''
    input.value = ''
    pageError.value = t('assetFlow.selectVoucherImage')
    return
  }
  clearVoucherPreview()
  voucherFile.value = file
  voucherName.value = file.name
  voucherPreviewUrl.value = URL.createObjectURL(file)
  pageError.value = ''
}

function clearVoucherPreview() {
  if (!voucherPreviewUrl.value) return
  URL.revokeObjectURL(voucherPreviewUrl.value)
  voucherPreviewUrl.value = ''
}

async function uploadVoucherImage() {
  if (!voucherFile.value) return ''

  const resp = await apiUploadFile(voucherFile.value)
  if (!isSuccessCode(resp.code) || !resp.data?.url) {
    throw new Error(resp.msg || t('assetFlow.uploadVoucherFailed'))
  }
  return resp.data.url
}

async function completeRecharge() {
  if (submitLoading.value) return

  pageError.value = ''
  copyTip.value = ''
  const rechargeAmountText = parseAssetDecimalToMinorText(amount.value, selectedDecimalPlaces.value)
  const rechargeAmount = Number(rechargeAmountText)
  if (!rechargeAmountText || rechargeAmount <= 0) {
    pageError.value = t('assetFlow.invalidRechargeAmount', {
      places: selectedInputDecimalPlaces.value,
    })
    return
  }
  if (!voucherFile.value) {
    pageError.value = t('assetFlow.selectVoucherImage')
    return
  }

  submitLoading.value = true
  try {
    const voucherImage = await uploadVoucherImage()
    const resp = await apiCreateCryptoRechargeOrder({
      walletType: walletType.value,
      coin: selectedCoin.value,
      chainCode: selectedChainCode.value,
      rechargeAmount,
      clientType: 2,
      voucherImage,
    })
    if (isSuccessCode(resp.code)) {
      if (resp.data?.address) {
        rechargeAddress.value = resp.data.address
        qrImageUrl.value = await createQrDataUrl(resp.data.address.address)
      }
      stopAddressCountdown()
      copyTip.value = resp.data?.order?.orderNo
        ? t('assetFlow.submittedWithNo', { orderNo: resp.data.order.orderNo })
        : t('assetFlow.submittedWaitConfirm')
    } else {
      pageError.value = resp.msg || t('assetFlow.submitFailedLater')
    }
  } catch (error) {
    console.warn('create crypto recharge order failed', error)
    pageError.value = t('assetFlow.submitFailedLater')
  } finally {
    submitLoading.value = false
  }
}

onMounted(() => {
  void loadCoinConfigs()
})

onBeforeUnmount(() => {
  stopAddressCountdown()
  clearVoucherPreview()
})
</script>

<template>
  <CommonPage
    :title="t('assetFlow.recharge')"
    :right-text="step === 'select' ? t('assetFlow.records') : undefined"
    :nav-height="76"
    @back="router.back()"
    @right-click="router.push({ name: 'asset-flows', query: { tab: 'recharge' } })"
  >
    <section class="asset-flow-content">
      <template v-if="step === 'select'">
        <h2>{{ t('assetFlow.paymentMethod') }}</h2>
        <AssetCoinPicker
          :coin="selectedCoin"
          :config="selectedConfig || undefined"
          :chain="selectedChain"
          @click="coinSheetVisible = true"
        />
        <p v-if="pageError" class="state-text state-text--error">
          {{ pageError }}
        </p>
        <AssetPrimaryButton
          class="recharge-button"
          :label="addressLoading ? t('assetFlow.getting') : t('assetFlow.recharge')"
          @click="startRecharge"
        />
      </template>

      <template v-else>
        <div class="detail-coin">
          <AssetCoinPicker
            :coin="selectedCoin"
            :config="selectedConfig || undefined"
            :chain="selectedChain"
            @click="coinSheetVisible = true"
          />
        </div>

        <div class="qr-card">
          <img v-if="qrImageUrl" :src="qrImageUrl" :alt="t('assetFlow.qrAlt')">
        </div>

        <div class="address-row">
          <strong>{{ rechargeAddress?.address }}</strong>
          <button type="button" @click="copyText(rechargeAddress?.address || '')">
            {{ t('common.copy') }}
          </button>
        </div>
        <p v-if="addressSecondsLeft > 0" class="address-countdown">
          {{ t('assetFlow.addressExpires', { time: addressCountdownText }) }}
        </p>
        <div v-if="rechargeAddress?.memo" class="memo-row">
          <span>Memo / Tag</span>
          <strong>{{ rechargeAddress.memo }}</strong>
          <button type="button" @click="copyText(rechargeAddress.memo)">
            {{ t('common.copy') }}
          </button>
        </div>

        <div class="divider" />

        <section class="field-block">
          <h2>{{ t('assetFlow.voucher') }}</h2>
          <button type="button" class="voucher-upload" @click="fileInputRef?.click()">
            <span class="voucher-upload__thumb">
              <img
                v-if="voucherPreviewUrl"
                :src="voucherPreviewUrl"
                :alt="t('assetFlow.voucher')"
              >
              <b v-else>+</b>
            </span>
            <strong>{{
              voucherPreviewUrl ? t('assetFlow.changeVoucher') : t('assetFlow.uploadVoucher')
            }}</strong>
          </button>
          <input
            ref="fileInputRef"
            class="file-input"
            type="file"
            accept="image/*"
            @change="handleVoucherChange"
          >
        </section>

        <section class="field-block">
          <h2>{{ t('assetFlow.rechargeAmount') }}</h2>
          <label class="amount-input">
            <input v-model="amount" inputmode="decimal">
            <span>{{ selectedCoin }}</span>
          </label>
        </section>

        <p v-if="pageError" class="state-text state-text--error">
          {{ pageError }}
        </p>
        <p v-if="copyTip" class="copy-tip">
          {{ copyTip }}
        </p>
        <AssetPrimaryButton
          class="complete-button"
          :label="submitLoading ? t('common.submitting') : t('assetFlow.complete')"
          @click="completeRecharge"
        />
      </template>
    </section>

    <AssetCoinSelectSheet
      v-model="coinSheetVisible"
      :configs="coinConfigs"
      :selected-config="selectedConfig"
      :operation-type="1"
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
  background: var(--page-bg);
  color: var(--text);
}

:deep(.header-bar) {
  background: var(--page-bg);
}

:deep(.header-left) {
  left: 18px;
  justify-content: center;
  width: 36px;
  height: 36px;
  margin-top: 20px;
  border-radius: 50%;
  background: var(--border-soft);
}

:deep(.header-title) {
  font-size: 0.9rem;
  font-weight: 800;
}

:deep(.header-right) {
  right: 18px;
  color: var(--text);
  font-size: 0.7rem;
}

h2 {
  margin: 0 0 12px;
  font-size: 0.7rem;
  font-weight: 700;
}

.recharge-button {
  margin-top: 36px;
}

.recharge-button,
.complete-button {
  min-height: 48px;
  border-radius: 14px;
  font-size: 0.8rem;
}

.state-text {
  margin: 14px 0 0;
  color: var(--muted);
  font-size: 0.65rem;
  line-height: 1.6;
}

.state-text--error {
  color: var(--danger);
}

.detail-coin {
  display: flex;
  justify-content: center;
}

.detail-coin :deep(.asset-picker) {
  width: auto;
  min-height: 38px;
  padding: 0;
  background: transparent;
}

.detail-coin :deep(.asset-picker__arrow) {
  display: none;
}

.detail-coin :deep(.asset-picker strong) {
  font-size: 0.75rem;
}

.qr-card {
  display: grid;
  width: min(248px, 68vw);
  aspect-ratio: 1;
  margin: 24px auto 22px;
  place-items: center;
  border-radius: 22px;
  background: #fff;
}

.qr-card img {
  width: 88%;
  height: 88%;
  object-fit: contain;
}

.address-row,
.memo-row,
.amount-input {
  display: flex;
  align-items: center;
  gap: 12px;
  min-height: 58px;
  padding: 0 16px;
  border-radius: 16px;
  background: var(--divider);
}

.address-row strong,
.memo-row strong {
  flex: 1;
  overflow-wrap: anywhere;
  color: var(--text);
  font-size: 0.7rem;
  line-height: 1.35;
}

.address-row button,
.memo-row button {
  flex: 0 0 auto;
  min-width: 56px;
  min-height: 34px;
  border: 1px solid var(--accent);
  border-radius: 999px;
  color: var(--accent);
  font-size: 0.7rem;
  font-weight: 800;
}

.address-countdown {
  margin: 10px 0 0;
  color: var(--warning);
  font-size: 0.6rem;
  font-weight: 800;
  text-align: center;
}

.memo-row {
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 8px 12px;
  margin-top: 10px;
  padding: 12px 16px;
}

.memo-row span {
  grid-column: 1 / -1;
  color: var(--warning);
  font-size: 0.6rem;
  font-weight: 800;
}

.divider {
  margin: 24px -18px 28px;
  border-top: 1px dashed #2c2f3a;
}

.field-block {
  margin-bottom: 24px;
}

.voucher-upload {
  display: inline-grid;
  grid-template-columns: 78px auto;
  align-items: center;
  gap: 16px;
  color: var(--muted);
  font-size: 0.7rem;
  font-weight: 800;
  text-align: left;
}

.voucher-upload__thumb {
  display: grid;
  width: 78px;
  height: 78px;
  place-items: center;
  overflow: hidden;
  border-radius: 18px;
  background: var(--field-bg);
  color: var(--muted);
}

.voucher-upload__thumb b {
  font-size: 2rem;
  font-weight: 300;
  line-height: 1;
}

.voucher-upload__thumb img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.voucher-upload strong {
  max-width: 180px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 0.7rem;
  font-weight: 800;
}

.file-input {
  display: none;
}

.amount-input {
  min-height: 72px;
  padding: 0 20px;
}

.amount-input input {
  min-width: 0;
  flex: 1;
  font-size: 1rem;
  font-weight: 800;
}

.amount-input span {
  font-size: 0.75rem;
  font-weight: 800;
}

.copy-tip {
  margin: -8px 0 18px;
  color: var(--accent);
  font-size: 0.65rem;
  font-weight: 800;
  text-align: center;
}

.complete-button {
  margin-top: 18px;
}
</style>
