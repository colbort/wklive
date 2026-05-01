<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'

import { apiListAssetCoinConfigs } from '@/api/asset'
import { apiCreateCryptoRechargeOrder, apiGetMyCryptoRechargeAddress } from '@/api/payment'
import AssetCoinSelectSheet from '@/components/assets/AssetCoinSelectSheet.vue'
import AssetCoinPicker from '@/components/assets/AssetCoinPicker.vue'
import AssetFlowLayout from '@/components/assets/AssetFlowLayout.vue'
import AssetPrimaryButton from '@/components/assets/AssetPrimaryButton.vue'
import type { AssetCoinConfig } from '@/types/asset'
import type { CryptoRechargeAddress } from '@/types/payment'

const route = useRoute()
const coinConfigs = ref<AssetCoinConfig[]>([])
const selectedConfig = ref<AssetCoinConfig | null>(null)
const coinSheetVisible = ref(false)
const addressLoading = ref(false)
const rechargeAddress = ref<CryptoRechargeAddress | null>(null)
const pageError = ref('')
const copyTip = ref('')
const amount = ref('')
const voucherName = ref('')
const submitLoading = ref(false)
const addressSecondsLeft = ref(0)
let addressTimer: ReturnType<typeof setInterval> | undefined
const fileInputRef = ref<HTMLInputElement | null>(null)
const step = ref<'select' | 'detail'>('select')

const walletType = computed(() => Number(route.query.walletType || 1))
const routeCoin = computed(() => String(route.query.coin || 'USDT'))
const selectedCoin = computed(() => selectedConfig.value?.coin || routeCoin.value)
const selectedChainCode = computed(() => selectedConfig.value?.chainCode || (selectedCoin.value === 'USDT' ? 20 : 0))
const selectedChain = computed(() => {
  const config = selectedConfig.value
  if (!config) return routeCoin.value === 'USDT' ? 'TRC20' : ''
  return getChainLabel(config)
})
const qrImageUrl = computed(() => {
  const address = rechargeAddress.value?.address || ''
  if (!address) return ''
  return `https://api.qrserver.com/v1/create-qr-code/?size=220x220&margin=10&data=${encodeURIComponent(address)}`
})
const addressCountdownText = computed(() => {
  const seconds = addressSecondsLeft.value
  const minutes = Math.floor(seconds / 60)
  const remainSeconds = String(seconds % 60).padStart(2, '0')
  return `${minutes}:${remainSeconds}`
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
  rechargeAddress.value = null
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
    copyTip.value = ''
    pageError.value = '充值地址已过期，请重新获取地址'
    step.value = 'select'
  }, 1000)
}

async function startRecharge() {
  pageError.value = ''
  copyTip.value = ''
  if (!selectedCoin.value || !selectedChainCode.value) {
    pageError.value = '请选择支持充值的币种和网络'
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
      startAddressCountdown()
      step.value = 'detail'
    } else {
      pageError.value = resp.msg || '暂未获取到充值地址'
    }
  } catch (error) {
    console.warn('load crypto recharge address failed', error)
    pageError.value = '充值地址获取失败，请稍后重试'
  } finally {
    addressLoading.value = false
  }
}

async function copyText(text: string) {
  if (!text) return
  try {
    await navigator.clipboard.writeText(text)
    copyTip.value = '复制成功'
  } catch (error) {
    console.warn('copy failed', error)
    copyTip.value = '复制失败，请手动复制'
  }
}

function handleVoucherChange(event: Event) {
  const input = event.target as HTMLInputElement
  voucherName.value = input.files?.[0]?.name || ''
}

function parseAmountToMinor(value: string) {
  const normalized = value.trim()
  if (!/^\d+(\.\d{1,2})?$/.test(normalized)) return 0
  const [integerPart, decimalPart = ''] = normalized.split('.')
  return Number(integerPart) * 100 + Number(decimalPart.padEnd(2, '0'))
}

async function completeRecharge() {
  if (submitLoading.value) return

  pageError.value = ''
  copyTip.value = ''
  const rechargeAmount = parseAmountToMinor(amount.value)
  if (rechargeAmount <= 0) {
    pageError.value = '请输入充值金额'
    return
  }

  submitLoading.value = true
  try {
    const resp = await apiCreateCryptoRechargeOrder({
      walletType: walletType.value,
      coin: selectedCoin.value,
      chainCode: selectedChainCode.value,
      rechargeAmount,
      clientType: 2,
    })
    if (isSuccessCode(resp.code)) {
      if (resp.address) {
        rechargeAddress.value = resp.address
      }
      stopAddressCountdown()
      copyTip.value = resp.data?.orderNo ? `已提交：${resp.data.orderNo}` : '已提交，请等待链上确认'
    } else {
      pageError.value = resp.msg || '提交失败，请稍后重试'
    }
  } catch (error) {
    console.warn('create crypto recharge order failed', error)
    pageError.value = '提交失败，请稍后重试'
  } finally {
    submitLoading.value = false
  }
}

onMounted(() => {
  void loadCoinConfigs()
})

onBeforeUnmount(() => {
  stopAddressCountdown()
})
</script>

<template>
  <AssetFlowLayout title="充值" :right-text="step === 'select' ? '资金记录' : undefined" narrow>
    <template v-if="step === 'select'">
      <h2>支付方式</h2>
      <AssetCoinPicker
        :coin="selectedCoin"
        :config="selectedConfig || undefined"
        :chain="selectedChain"
        @click="coinSheetVisible = true"
      />
      <p v-if="pageError" class="state-text state-text--error">{{ pageError }}</p>
      <AssetPrimaryButton class="recharge-button" :label="addressLoading ? '获取中' : '充值'" @click="startRecharge" />
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
        <img v-if="qrImageUrl" :src="qrImageUrl" alt="充值地址二维码" />
      </div>

      <div class="address-row">
        <strong>{{ rechargeAddress?.address }}</strong>
        <button type="button" @click="copyText(rechargeAddress?.address || '')">复制</button>
      </div>
      <p v-if="addressSecondsLeft > 0" class="address-countdown">
        地址有效期 {{ addressCountdownText }}，超时后需重新获取
      </p>
      <div v-if="rechargeAddress?.memo" class="memo-row">
        <span>Memo / Tag</span>
        <strong>{{ rechargeAddress.memo }}</strong>
        <button type="button" @click="copyText(rechargeAddress.memo)">复制</button>
      </div>

      <div class="divider" />

      <section class="field-block">
        <h2>付款凭证</h2>
        <button type="button" class="voucher-upload" @click="fileInputRef?.click()">
          <span>+</span>
          <b>{{ voucherName || '上传付款凭证' }}</b>
        </button>
        <input ref="fileInputRef" class="file-input" type="file" accept="image/*,.pdf" @change="handleVoucherChange" />
      </section>

      <section class="field-block">
        <h2>充值金额</h2>
        <label class="amount-input">
          <input v-model="amount" inputmode="decimal" />
          <span>{{ selectedCoin }}</span>
        </label>
      </section>

      <p v-if="pageError" class="state-text state-text--error">{{ pageError }}</p>
      <p v-if="copyTip" class="copy-tip">{{ copyTip }}</p>
      <AssetPrimaryButton
        class="complete-button"
        :label="submitLoading ? '提交中' : '完成'"
        @click="completeRecharge"
      />
    </template>

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
button,
input {
  border: 0;
  background: transparent;
  color: inherit;
  font: inherit;
}

h2 {
  margin: 0 0 12px;
  font-size: 14px;
  font-weight: 700;
}

.recharge-button {
  margin-top: 36px;
}

.recharge-button,
.complete-button {
  min-height: 48px;
  border-radius: 14px;
  font-size: 16px;
}

.state-text {
  margin: 14px 0 0;
  color: #a8abb6;
  font-size: 13px;
  line-height: 1.6;
}

.state-text--error {
  color: #ff7676;
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
  font-size: 15px;
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
  background: #20222d;
}

.address-row strong,
.memo-row strong {
  flex: 1;
  overflow-wrap: anywhere;
  color: #fff;
  font-size: 14px;
  line-height: 1.35;
}

.address-row button,
.memo-row button {
  flex: 0 0 auto;
  min-width: 56px;
  min-height: 34px;
  border: 1px solid #02b904;
  border-radius: 999px;
  color: #02d107;
  font-size: 14px;
  font-weight: 800;
}

.address-countdown {
  margin: 10px 0 0;
  color: #ffce6a;
  font-size: 12px;
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
  color: #ffce6a;
  font-size: 12px;
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
  color: #8f929e;
  font-size: 14px;
  font-weight: 800;
  text-align: left;
}

.voucher-upload span {
  display: grid;
  width: 78px;
  height: 78px;
  place-items: center;
  border-radius: 18px;
  background: #292b36;
  color: #9b9da6;
  font-size: 40px;
  font-weight: 300;
  line-height: 1;
}

.voucher-upload b {
  max-width: 180px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
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
  font-size: 20px;
  font-weight: 800;
}

.amount-input span {
  font-size: 15px;
  font-weight: 800;
}

.copy-tip {
  margin: -8px 0 18px;
  color: #02d107;
  font-size: 13px;
  font-weight: 800;
  text-align: center;
}

.complete-button {
  margin-top: 18px;
}
</style>
