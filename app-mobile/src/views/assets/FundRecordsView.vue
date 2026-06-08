<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { apiGetAssetOptions, apiListAssetCoinConfigs } from '@/api/asset'
import {
  apiGetPaymentOptions,
  apiListMyRechargeOrders,
  apiListMyWithdrawOrders,
} from '@/api/payment'
import CommonPage from '@/components/common/CommonPage.vue'
import { useOptions } from '@/composables/useOptions'
import { useI18n } from '@/i18n'
import type { AssetCoinConfig } from '@/types/asset'
import type { RechargeOrder, WithdrawOrder } from '@/types/payment'
import { formatAssetMinorAmount } from '@/utils/assetAmount'

type MainTab = 'recharge' | 'withdraw'
type WithdrawTab = 'crypto' | 'bank'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const assetOptions = useOptions(apiGetAssetOptions)
const paymentOptions = useOptions(apiGetPaymentOptions)
const activeTab = ref<MainTab>(route.query.tab === 'withdraw' ? 'withdraw' : 'recharge')
const withdrawTab = ref<WithdrawTab>('crypto')
const coinConfigs = ref<AssetCoinConfig[]>([])
const rechargeOrders = ref<RechargeOrder[]>([])
const withdrawOrders = ref<WithdrawOrder[]>([])
const loading = ref(false)
const pageError = ref('')

const recordMenus = computed(() => [
  { label: t('assetFlow.rechargeRecords'), value: 'recharge' },
  { label: t('assetFlow.withdrawRecords'), value: 'withdraw' },
])

const visibleRecords = computed(() => {
  if (activeTab.value === 'recharge') {
    return rechargeOrders.value.map((order) => ({
      id: order.id,
      title: t('assetFlow.rechargeAccount'),
      orderNo: order.orderNo,
      currency: order.currency,
      amount: order.orderAmount,
      statusValue: order.status,
      status: rechargeStatusLabel(order.status),
      time: order.createTimes,
      accountLabel: walletTypeLabel(order.walletType),
      positive: true,
    }))
  }

  if (withdrawTab.value === 'bank') return []

  return withdrawOrders.value.map((order) => ({
    id: order.id,
    title: t('assetFlow.withdrawAccount'),
    orderNo: order.orderNo,
    currency: order.currency,
    amount: order.amount,
    statusValue: order.status,
    status: withdrawStatusLabel(order.status),
    time: order.createTimes,
    accountLabel: walletTypeLabel(1),
    positive: false,
  }))
})

function isSuccessCode(code: number) {
  return code === 0 || code === 200
}

function selectMainTab(tab: MainTab) {
  activeTab.value = tab
  pageError.value = ''
  void loadActiveRecords()
}

function selectMenu(value: string | number) {
  selectMainTab(value === 'withdraw' ? 'withdraw' : 'recharge')
}

function selectWithdrawTab(tab: WithdrawTab) {
  withdrawTab.value = tab
  pageError.value = ''
  if (tab === 'crypto') void loadWithdrawOrders()
}

async function loadActiveRecords() {
  if (activeTab.value === 'recharge') {
    await loadRechargeOrders()
    return
  }
  if (withdrawTab.value === 'crypto') {
    await loadWithdrawOrders()
  }
}

async function loadCoinConfigs() {
  try {
    const resp = await apiListAssetCoinConfigs({ walletType: 1, operationType: 1 })
    if (isSuccessCode(resp.code)) {
      coinConfigs.value = resp.data || []
    }
  } catch (error) {
    console.warn('load fund record coin configs failed', error)
  }
}

async function loadRechargeOrders() {
  loading.value = true
  pageError.value = ''
  try {
    const resp = await apiListMyRechargeOrders({ limit: 20 })
    if (isSuccessCode(resp.code)) {
      rechargeOrders.value = resp.data || []
    } else {
      pageError.value = resp.msg || t('assetFlow.recordsLoadFailed')
    }
  } catch (error) {
    console.warn('load recharge records failed', error)
    pageError.value = t('assetFlow.recordsLoadFailed')
  } finally {
    loading.value = false
  }
}

async function loadWithdrawOrders() {
  loading.value = true
  pageError.value = ''
  try {
    const resp = await apiListMyWithdrawOrders({ limit: 20 })
    if (isSuccessCode(resp.code)) {
      withdrawOrders.value = resp.data || []
    } else {
      pageError.value = resp.msg || t('assetFlow.recordsLoadFailed')
    }
  } catch (error) {
    console.warn('load withdraw records failed', error)
    pageError.value = t('assetFlow.recordsLoadFailed')
  } finally {
    loading.value = false
  }
}

function configForCoin(coin: string) {
  return coinConfigs.value.find((config) => config.coin === coin)
}

function coinIconText(coin: string) {
  if (coin.toUpperCase() === 'USDT') return '₮'

  const config = configForCoin(coin)
  return (config?.iconText || config?.symbol || coin || '?').slice(0, 3).toUpperCase()
}

function coinIconStyle(coin: string) {
  return { backgroundColor: configForCoin(coin)?.iconBgColor || '#16ad77' }
}

function formatRecordAmount(value: number, currency: string) {
  return `${formatAssetMinorAmount(value, 2)} ${currency}`
}

function formatRecordTime(value: number) {
  if (!value) return ''
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

function rechargeStatusLabel(status: number) {
  return paymentOptions.getLabel('payOrderStatus', status, t('assetFlow.unknown'))
}

function withdrawStatusLabel(status: number) {
  return paymentOptions.getLabel('payOrderStatus', status, t('assetFlow.unknown'))
}

function statusClass(status: number) {
  if (status === 3) return 'record-status--success'
  if (status === 4) return 'record-status--failed'
  if (status === 5 || status === 6) return 'record-status--closed'
  return 'record-status--processing'
}

function openRecord(record: { orderNo: string }) {
  if (activeTab.value !== 'recharge') return
  void router.push({ name: 'asset-recharge-detail', params: { orderNo: record.orderNo } })
}

watch(
  () => route.query.tab,
  (tab) => {
    activeTab.value = tab === 'withdraw' ? 'withdraw' : 'recharge'
    void loadActiveRecords()
  },
)

onMounted(() => {
  void loadCoinConfigs()
  void loadActiveRecords()
})
</script>

<template>
  <CommonPage
    class="fund-record-layout"
    :title="t('assetFlow.records')"
    :nav-height="76"
    :menus="recordMenus"
    :model-value="activeTab"
    @back="router.back()"
    @update:model-value="selectMenu"
  >
    <section class="fund-record-page">
      <div v-if="activeTab === 'withdraw'" class="withdraw-type-tabs">
        <button
          type="button"
          :class="{ 'withdraw-type-tabs__item--active': withdrawTab === 'crypto' }"
          @click="selectWithdrawTab('crypto')"
        >
          {{ t('assetFlow.crypto') }}
        </button>
        <button
          type="button"
          :class="{ 'withdraw-type-tabs__item--active': withdrawTab === 'bank' }"
          @click="selectWithdrawTab('bank')"
        >
          {{ t('assetFlow.bankCard') }}
        </button>
      </div>

      <p v-if="pageError" class="records-state records-state--error">
        {{ pageError }}
      </p>
      <p v-else-if="loading" class="records-state">
        {{ t('common.loading') }}
      </p>

      <section v-else-if="visibleRecords.length" class="record-list">
        <article
          v-for="record in visibleRecords"
          :key="`${activeTab}-${record.id}`"
          class="record-item"
          :class="{ 'record-item--clickable': activeTab === 'recharge' }"
          @click="openRecord(record)"
        >
          <span class="record-coin-icon" :style="coinIconStyle(record.currency)">
            <img
              v-if="configForCoin(record.currency)?.iconUrl"
              :src="configForCoin(record.currency)?.iconUrl"
              :alt="record.currency"
            >
            <span v-else>{{ coinIconText(record.currency) }}</span>
          </span>
          <div class="record-main">
            <div class="record-title-group">
              <strong>{{ record.currency }}</strong>
              <span>
                {{ record.title }}{{ t('assetFlow.accountSeparator')
                }}<b>{{ record.accountLabel }}</b>
              </span>
            </div>
            <span class="record-time">{{ formatRecordTime(record.time) }}</span>
          </div>
          <aside class="record-side">
            <b>{{ formatRecordAmount(record.amount, record.currency) }}</b>
            <span class="record-status" :class="statusClass(record.statusValue)">{{
              record.status
            }}</span>
          </aside>
        </article>
      </section>

      <section v-else class="empty-records">
        <div class="empty-records__icon" aria-hidden="true">
          <span class="empty-records__paper">
            <i />
            <i />
            <i />
          </span>
          <span class="empty-records__lens" />
        </div>
        <p>{{ t('common.none') }}</p>
      </section>
    </section>
  </CommonPage>
</template>

<style scoped>
button {
  border: 0;
  background: transparent;
  color: inherit;
  font: inherit;
}

.fund-record-layout {
  overflow-x: clip;
}

.fund-record-page {
  min-height: calc(100dvh - 152px);
  padding: 22px 18px 96px;
  background: var(--page-bg);
  color: var(--text);
}

:deep(.header-bar) {
  background: var(--page-bg);
}

:deep(.header-left) {
  left: 18px;
  justify-content: center;
  width: 46px;
  height: 46px;
  margin-top: 15px;
  border-radius: 50%;
  background: #20232e;
}

:deep(.header-title) {
  font-size: 20px;
  font-weight: 800;
}

:deep(.sub-menu) {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  height: var(--menu-bar-height);
  border-bottom: 1px solid #1f212b;
  background: var(--page-bg);
}

:deep(.sub-menu-item) {
  height: var(--menu-bar-height);
  justify-content: center;
}

:deep(.sub-menu-item.active::after) {
  bottom: -1px;
  width: 1.8rem;
}

.withdraw-type-tabs {
  display: flex;
  gap: 10px;
  margin: 22px 4px 0;
}

.withdraw-type-tabs button {
  min-width: 112px;
  min-height: 60px;
  padding: 0 24px;
  border: 1px solid #282a34;
  border-radius: 999px;
  color: var(--muted);
  font-size: 16px;
  font-weight: 900;
}

.withdraw-type-tabs__item--active {
  border-color: var(--accent) !important;
  background: var(--accent) !important;
  color: var(--text) !important;
}

.records-state {
  margin: 76px 0 0;
  color: var(--muted);
  font-size: 14px;
  font-weight: 800;
  text-align: center;
}

.records-state--error {
  color: var(--danger);
}

.record-list {
  display: grid;
  gap: 20px;
  justify-items: center;
  padding: 18px 0 0;
}

.record-item {
  display: flex;
  box-sizing: border-box;
  align-items: flex-start;
  gap: 14px;
  width: 100%;
  min-height: 104px;
  padding: 18px 16px 14px 16px;
  border-radius: 24px;
  background: #1d1f2a;
}

.record-item--clickable {
  cursor: pointer;
}

.record-coin-icon {
  flex: 0 0 44px;
  display: grid;
  width: 44px;
  height: 44px;
  place-items: center;
  overflow: hidden;
  border-radius: 50%;
  color: var(--text);
  font-size: 12px;
  font-weight: 700;
  margin-top: 1px;
}

.record-coin-icon img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.record-main {
  display: grid;
  gap: 9px;
  flex: 1 1 auto;
  min-width: 0;
}

.record-title-group {
  display: grid;
  gap: 6px;
  min-width: 0;
}

.record-main strong,
.record-side b {
  display: block;
  color: var(--text);
  font-size: 16px;
  font-weight: 500;
  line-height: 1.15;
}

.record-main span,
.record-title-group span {
  display: block;
  min-width: 0;
  overflow: hidden;
  margin-top: 0;
  color: var(--muted);
  font-size: 14px;
  font-weight: 500;
  line-height: 1.25;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.record-title-group span b {
  color: var(--text);
  font-weight: 600;
}

.record-time {
  margin-top: 0 !important;
}

.record-side {
  display: grid;
  align-content: center;
  justify-items: end;
  gap: 10px;
  flex: 0 0 112px;
  min-width: 112px;
  min-height: 68px;
  text-align: right;
}

.record-side b {
  max-width: 150px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.record-status {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 28px;
  padding: 0 10px;
  border-radius: 999px;
  font-size: 14px;
  font-weight: 700;
}

.record-status--processing {
  background: rgba(255, 122, 0, 0.13);
  color: #ff8a00;
}

.record-status--success {
  background: rgba(0, 199, 10, 0.13);
  color: var(--accent);
}

.record-status--failed {
  background: rgba(255, 118, 118, 0.13);
  color: var(--danger);
}

.record-status--closed {
  background: rgba(141, 144, 154, 0.13);
  color: var(--muted);
}

.empty-records {
  display: grid;
  place-items: center;
  padding-top: 112px;
  color: var(--muted);
}

.empty-records p {
  margin: 24px 0 0;
  font-size: 16px;
  font-weight: 900;
}

.empty-records__icon {
  position: relative;
  width: 148px;
  height: 118px;
}

.empty-records__paper {
  position: absolute;
  left: 28px;
  top: 8px;
  width: 78px;
  height: 86px;
  border-radius: 16px 16px 18px 6px;
  background: linear-gradient(90deg, #242632 0 72%, #11131d 72% 100%);
}

.empty-records__paper::before {
  position: absolute;
  bottom: 0;
  left: -18px;
  width: 72px;
  height: 20px;
  border-radius: 0 0 18px 18px;
  background: linear-gradient(90deg, #252734, #181a24);
  content: '';
}

.empty-records__paper i {
  display: block;
  width: 44px;
  height: 5px;
  margin: 18px 0 0 16px;
  border-radius: 999px;
  background: #333541;
}

.empty-records__paper i + i {
  width: 32px;
  margin-top: 13px;
}

.empty-records__lens {
  position: absolute;
  right: 26px;
  top: 42px;
  width: 52px;
  height: 52px;
  border: 9px solid var(--accent);
  border-radius: 50%;
}

.empty-records__lens::after {
  position: absolute;
  right: -25px;
  bottom: -20px;
  width: 34px;
  height: 9px;
  border-radius: 999px;
  background: var(--accent);
  content: '';
  transform: rotate(45deg);
  transform-origin: left center;
}

@media (max-width: 390px) {
  .withdraw-type-tabs button {
    min-width: 104px;
    min-height: 56px;
    font-size: 15px;
  }

  .record-list {
    gap: 16px;
    padding-top: 16px;
    padding-right: 0;
    padding-left: 0;
  }

  .record-item {
    gap: 11px;
    min-height: 98px;
    width: 100%;
    padding: 16px 12px 12px 12px;
    border-radius: 22px;
  }

  .record-coin-icon {
    flex-basis: 40px;
    width: 40px;
    height: 40px;
    font-size: 11px;
  }

  .record-main strong,
  .record-side b {
    font-size: 15px;
  }

  .record-main span {
    font-size: 12px;
  }

  .record-status {
    min-height: 26px;
    padding: 0 9px;
    font-size: 13px;
  }

  .record-side b {
    max-width: 108px;
  }

  .record-side {
    flex-basis: 92px;
    min-width: 92px;
  }
}
</style>
