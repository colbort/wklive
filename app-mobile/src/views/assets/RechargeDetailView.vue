<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { apiGetAssetOptions, apiListAssetCoinConfigs } from '@/api/asset'
import { apiGetMyRechargeOrder } from '@/api/payment'
import CommonPage from '@/components/common/CommonPage.vue'
import { useOptions } from '@/composables/useOptions'
import { useI18n } from '@/i18n'
import { useSystemStore } from '@/stores/system'
import type { AssetCoinConfig } from '@/types/asset'
import type { RechargeOrder } from '@/types/payment'
import { formatAssetMinorAmount } from '@/utils/assetAmount'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const assetOptions = useOptions(apiGetAssetOptions)
const systemStore = useSystemStore()
const order = ref<RechargeOrder | null>(null)
const coinConfigs = ref<AssetCoinConfig[]>([])
const loading = ref(false)
const pageError = ref('')

const orderNo = computed(() => String(route.params.orderNo || ''))
const requestSnapshot = computed<Record<string, unknown>>(() =>
  parseJsonObject(order.value?.requestData),
)
const chainCode = computed(() => Number(requestSnapshot.value.chainCode || 0))
const coinConfig = computed(() => {
  return (
    coinConfigs.value.find(
      (config) =>
        config.coin === order.value?.currency &&
        (!chainCode.value || Number(config.chainCode) === chainCode.value),
    ) ||
    coinConfigs.value.find((config) => config.coin === order.value?.currency) ||
    null
  )
})
const chainLabel = computed(() => {
  const code = chainCode.value || coinConfig.value?.chainCode || 0
  if (!code) return ''
  const option = assetOptions.getGroup('chainCode').find((item) => item.value === code)
  return option ? option.code.replace(/^CHAIN_CODE_/, '') : String(code)
})
const rechargeAddress = computed(() => {
  return String(requestSnapshot.value.address || order.value?.qrContent || order.value?.body || '')
})
const voucherImageUrl = computed(() => resolveAssetUrl(order.value?.voucherImage))
const rechargeAccountLabel = computed(() => walletTypeLabel(order.value?.walletType))

function isSuccessCode(code: number) {
  return code === 0 || code === 200
}

async function loadDetail() {
  if (!orderNo.value) {
    pageError.value = t('assetFlow.rechargeDetailLoadFailed')
    return
  }

  loading.value = true
  pageError.value = ''
  try {
    const resp = await apiGetMyRechargeOrder({ orderNo: orderNo.value })
    if (isSuccessCode(resp.code)) {
      order.value = resp.data || null
      if (order.value) {
        await loadCoinConfigs(order.value.walletType)
      }
    } else {
      pageError.value = resp.msg || t('assetFlow.rechargeDetailLoadFailed')
    }
  } catch (error) {
    console.warn('load recharge detail failed', error)
    pageError.value = t('assetFlow.rechargeDetailLoadFailed')
  } finally {
    loading.value = false
  }
}

async function loadCoinConfigs(walletType = 1) {
  try {
    const resp = await apiListAssetCoinConfigs({ walletType, operationType: 1 })
    if (isSuccessCode(resp.code)) {
      coinConfigs.value = resp.data || []
    }
  } catch (error) {
    console.warn('load recharge detail coin configs failed', error)
  }
}

function parseJsonObject(value?: string) {
  if (!value) return {}
  try {
    const parsed = JSON.parse(value)
    return parsed && typeof parsed === 'object' && !Array.isArray(parsed)
      ? (parsed as Record<string, unknown>)
      : {}
  } catch {
    return {}
  }
}

function coinIconText() {
  return (coinConfig.value?.iconText || coinConfig.value?.symbol || order.value?.currency || '?')
    .slice(0, 3)
    .toUpperCase()
}

function resolveAssetUrl(url?: string) {
  if (!url) return ''
  if (/^https?:\/\//i.test(url)) return url
  const assetUrl = systemStore.systemCore.assetUrl || ''
  if (!assetUrl) return url.startsWith('/') ? url : `/${url}`
  const base = assetUrl.replace(/\/+$/, '')
  const path = url.replace(/^\.\//, '').replace(/^\/+/, '')
  return `${base}/${path}`
}

function formatRechargeAmount() {
  if (!order.value) return '-'
  return `${formatAssetMinorAmount(order.value.orderAmount, 2)} ${order.value.currency}`
}

function formatRecordTime(value?: number) {
  if (!value) return '-'
  const date = new Date(value)
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hour = String(date.getHours()).padStart(2, '0')
  const minute = String(date.getMinutes()).padStart(2, '0')
  return `${year}-${month}-${day} ${hour}:${minute}`
}

function walletTypeLabel(walletType?: number) {
  return assetOptions.getLabel('walletType', Number(walletType), t('options.WALLET_TYPE_SPOT'))
}

function rechargeStatusLabel(status?: number) {
  const labels: Record<number, string> = {
    1: t('assetFlow.pendingPayment'),
    2: t('assetFlow.processing'),
    3: t('assetFlow.success'),
    4: t('assetFlow.failed'),
    5: t('assetFlow.closed'),
    6: t('assetFlow.refunded'),
  }
  return labels[Number(status)] || t('assetFlow.unknown')
}

function statusClass(status?: number) {
  if (status === 3) return 'detail-status--success'
  if (status === 4) return 'detail-status--failed'
  if (status === 5 || status === 6) return 'detail-status--closed'
  return 'detail-status--processing'
}

onMounted(() => {
  void loadDetail()
})
</script>

<template>
  <CommonPage
    class="recharge-detail-layout"
    :title="t('assetFlow.rechargeDetail')"
    :nav-height="76"
    @back="router.back()"
  >
    <section class="asset-flow-content">
      <p v-if="pageError" class="detail-state detail-state--error">
        {{ pageError }}
      </p>
      <p v-else-if="loading" class="detail-state">
        {{ t('common.loading') }}
      </p>

      <section v-else-if="order" class="recharge-detail-card">
        <header class="detail-head">
          <span
            class="detail-coin-icon"
            :style="{ backgroundColor: coinConfig?.iconBgColor || 'var(--coin-fallback-bg)' }"
          >
            <img v-if="coinConfig?.iconUrl" :src="coinConfig.iconUrl" :alt="order.currency">
            <span v-else>{{ coinIconText() }}</span>
          </span>
          <strong>{{ order.currency }}</strong>
          <small v-if="chainLabel">{{ chainLabel }}</small>
          <span class="detail-status" :class="statusClass(order.status)">
            {{ rechargeStatusLabel(order.status) }}
          </span>
        </header>

        <dl class="detail-list">
          <div>
            <dt>{{ t('assetFlow.rechargeChannel') }}</dt>
            <dd>{{ t('assetFlow.crypto') }}</dd>
          </div>
          <div>
            <dt>{{ t('assetFlow.rechargeAmount') }}</dt>
            <dd>{{ formatRechargeAmount() }}</dd>
          </div>
          <div>
            <dt>{{ t('assetFlow.rechargeAddress') }}</dt>
            <dd class="detail-break">
              {{ rechargeAddress || '-' }}
            </dd>
          </div>
          <div>
            <dt>{{ t('assetFlow.voucher') }}</dt>
            <dd>
              <img
                v-if="voucherImageUrl"
                class="voucher-image"
                :src="voucherImageUrl"
                :alt="t('assetFlow.voucher')"
              >
              <span v-else>-</span>
            </dd>
          </div>
          <div>
            <dt>{{ t('assetFlow.rechargeAccount') }}</dt>
            <dd>{{ rechargeAccountLabel }}</dd>
          </div>
          <div>
            <dt>{{ t('assetFlow.rechargeTime') }}</dt>
            <dd>{{ formatRecordTime(order.createTimes) }}</dd>
          </div>
        </dl>
      </section>
    </section>
  </CommonPage>
</template>

<style scoped>
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

.detail-state {
  margin: 72px 0 0;
  color: var(--muted);
  font-size: 0.7rem;
  font-weight: 800;
  text-align: center;
}

.detail-state--error {
  color: var(--danger);
}

.recharge-detail-card {
  margin: 30px 0 0;
  overflow: hidden;
  border: 1px solid var(--divider);
  border-radius: 18px;
  background: var(--page-bg-soft);
}

.detail-head {
  display: grid;
  grid-template-columns: 44px minmax(0, auto) minmax(0, auto) minmax(72px, 1fr);
  align-items: center;
  gap: 8px;
  min-height: 74px;
  padding: 14px 20px;
  border-bottom: 1px solid var(--divider);
}

.detail-coin-icon {
  display: grid;
  width: 44px;
  height: 44px;
  place-items: center;
  overflow: hidden;
  border-radius: 50%;
  color: var(--text);
  font-size: 0.6rem;
  font-weight: 700;
}

.detail-coin-icon img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.detail-head strong {
  min-width: 0;
  overflow: hidden;
  color: var(--text);
  font-size: 0.8rem;
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.detail-head small {
  min-width: 0;
  overflow: hidden;
  padding: 2px 6px;
  border-radius: 6px;
  background: var(--selection-bg);
  color: var(--text);
  font-size: 0.7rem;
  font-weight: 600;
  line-height: 1.1;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.detail-status {
  justify-self: end;
  padding: 3px 8px;
  border-radius: 999px;
  font-size: 0.65rem;
  font-weight: 700;
}

.detail-status--processing {
  background: var(--warning-bg-soft);
  color: var(--warning-strong);
}

.detail-status--success {
  background: var(--success-bg-soft);
  color: var(--accent);
}

.detail-status--failed {
  background: var(--danger-bg-soft);
  color: var(--danger);
}

.detail-status--closed {
  background: var(--neutral-bg-soft);
  color: var(--muted);
}

.detail-list {
  display: grid;
  gap: 22px;
  margin: 0;
  padding: 28px 20px 26px;
}

.detail-list div {
  display: grid;
  grid-template-columns: 86px minmax(0, 1fr);
  gap: 16px;
}

.detail-list dt,
.detail-list dd {
  margin: 0;
  font-size: 0.75rem;
  font-weight: 600;
  line-height: 1.45;
}

.detail-list dt {
  color: var(--muted);
}

.detail-list dd {
  min-width: 0;
  color: var(--text);
}

.detail-break {
  overflow-wrap: anywhere;
}

.voucher-image {
  display: block;
  width: 72px;
  height: 72px;
  border-radius: 14px;
  background: var(--text-strong);
  object-fit: cover;
}

@media (max-width: 390px) {
  .recharge-detail-card {
    margin: 24px 0 0;
    border-radius: 16px;
  }

  .detail-head {
    grid-template-columns: 40px minmax(0, auto) minmax(0, auto) minmax(64px, 1fr);
    gap: 7px;
    min-height: 68px;
    padding: 12px 16px;
  }

  .detail-coin-icon {
    width: 40px;
    height: 40px;
    font-size: 0.55rem;
  }

  .detail-head strong {
    font-size: 0.75rem;
  }

  .detail-head small,
  .detail-status {
    font-size: 0.6rem;
  }

  .detail-list {
    gap: 20px;
    padding: 24px 16px;
  }

  .detail-list div {
    grid-template-columns: 76px minmax(0, 1fr);
    gap: 14px;
  }

  .detail-list dt,
  .detail-list dd {
    font-size: 0.7rem;
  }
}
</style>
