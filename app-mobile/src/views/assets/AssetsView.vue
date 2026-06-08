<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'

import { apiGetAssetOptions, apiGetMyAssetSummary, apiListAssetCoinConfigs } from '@/api/asset'
import { getAccessToken } from '@/api/http'
import BottomDrawer from '@/components/common/BottomDrawer.vue'
import CommonPage from '@/components/common/CommonPage.vue'
import LoginPrompt from '@/components/common/LoginPrompt.vue'
import { optionText, useOptions } from '@/composables/useOptions'
import { useI18n } from '@/i18n'
import type { AssetCoinConfig, AssetUserAsset } from '@/types/asset'
import { DEFAULT_ASSET_DECIMAL_PLACES, formatAssetMinorAmount } from '@/utils/assetAmount'

type AssetTopTab = 'assets' | 'orders' | 'profile'
type AssetActionKey =
  | 'cryptoRecharge'
  | 'bankRecharge'
  | 'cryptoWithdraw'
  | 'bankWithdraw'
  | 'transfer'
  | 'flows'

const ASSET_OPERATION_TYPES: Partial<Record<AssetActionKey, number>> = {
  cryptoRecharge: 1,
  cryptoWithdraw: 2,
  transfer: 3,
}

const router = useRouter()
const assetOptions = useOptions(apiGetAssetOptions)
const { t } = useI18n()
const activeTopTab = ref<AssetTopTab>('assets')
const activeAssetAccount = ref('cash')
const activeAssetAction = ref<AssetActionKey>('cryptoRecharge')
const activeOrderScope = ref('stock')
const coinConfigs = ref<AssetCoinConfig[]>([])
const coinConfigsLoading = ref(false)
const summaryAssets = ref<AssetUserAsset[]>([])
const summaryLoading = ref(false)
const selectedCoinConfig = ref<AssetCoinConfig | null>(null)

const assetActions: Array<{
  key: AssetActionKey
  labelKey: string
  icon: string
}> = [
  { key: 'cryptoRecharge', labelKey: 'assets.cryptoRecharge', icon: '$' },
  { key: 'bankRecharge', labelKey: 'assets.bankRecharge', icon: '▣' },
  { key: 'cryptoWithdraw', labelKey: 'assets.cryptoWithdraw', icon: '$' },
  { key: 'bankWithdraw', labelKey: 'assets.bankWithdraw', icon: '▭' },
  { key: 'transfer', labelKey: 'assets.transfer', icon: '⇄' },
  { key: 'flows', labelKey: 'assets.flows', icon: '▤' },
]

const fallbackAssetAccounts = [
  { key: 'cash', code: 'WALLET_TYPE_SPOT', walletType: 1 },
  { key: 'stock', code: 'WALLET_TYPE_FUNDING', walletType: 2 },
  { key: 'contract', code: 'WALLET_TYPE_CONTRACT', walletType: 3 },
]

const orderScopes = [
  { key: 'stock', labelKey: 'assets.stock' },
  { key: 'contract', labelKey: 'assets.contractOrders' },
  { key: 'option', labelKey: 'assets.optionContract' },
]

const orderTabs = ['assets.currentPosition', 'assets.history', 'assets.premarketOrders']

const pageTabs: Array<{ key: AssetTopTab; labelKey: string; shortLabelKey: string }> = [
  { key: 'assets', labelKey: 'assets.assetCenter', shortLabelKey: 'assets.assets' },
  { key: 'orders', labelKey: 'assets.orderCenter', shortLabelKey: 'assets.orders' },
  { key: 'profile', labelKey: 'assets.profileCenter', shortLabelKey: 'assets.profile' },
]

const assetAccounts = computed(() => {
  const walletTypeOptions = assetOptions
    .getGroup('walletType')
    .filter((option) => option.value > 0)
    .map((option) => ({
      key: String(option.value),
      label: optionText(option),
      walletType: option.value,
    }))

  return walletTypeOptions.length
    ? walletTypeOptions
    : fallbackAssetAccounts.map((account) => ({
        key: account.key,
        label: t(`options.${account.code}`),
        walletType: account.walletType,
      }))
})

const pageMenus = computed(() => {
  if (activeTopTab.value !== 'assets') return []

  return assetAccounts.value.map((account) => ({
    label: account.label,
    value: account.key,
  }))
})

const activeWalletType = computed(() => {
  return (
    assetAccounts.value.find((account) => account.key === activeAssetAccount.value)?.walletType || 1
  )
})

const activeOperationType = computed(() => ASSET_OPERATION_TYPES[activeAssetAction.value])
const activeAccountLabel = computed(() => {
  return (
    assetAccounts.value.find((account) => account.key === activeAssetAccount.value)?.label ||
    t('options.WALLET_TYPE_SPOT')
  )
})

const displayedAssets = computed(() => {
  const amountMap = new Map(
    summaryAssets.value
      .filter((asset) => asset.walletType === activeWalletType.value)
      .map((asset) => [asset.coin, asset]),
  )

  const configs = coinConfigs.value.length
    ? coinConfigs.value
    : summaryAssets.value
        .filter((asset) => asset.walletType === activeWalletType.value)
        .map((asset) => ({
          id: asset.id,
          tenantId: asset.tenantId,
          walletType: asset.walletType,
          coin: asset.coin,
          symbol: asset.coin,
          coinName: asset.coin,
          coinType: 0,
          chainCode: 0,
          iconUrl: '',
          iconText: asset.coin,
          iconBgColor: '',
          decimalPlaces: DEFAULT_ASSET_DECIMAL_PLACES,
          appVisible: 1,
          rechargeEnabled: 1,
          withdrawEnabled: 1,
          transferEnabled: 1,
          status: asset.status || 1,
          sort: 0,
          remark: '',
          createTimes: asset.createTimes,
          updateTimes: asset.updateTimes,
        }))

  return configs.map((config) => {
    const asset = amountMap.get(config.coin)
    return {
      config,
      amount: formatAssetMinorAmount(
        asset?.availableAmount || asset?.totalAmount || '0',
        config.decimalPlaces,
      ),
    }
  })
})

function isSuccessCode(code: number) {
  return code === 0 || code === 200
}

function coinConfigName(config: AssetCoinConfig) {
  return config.coinName || config.symbol || config.coin
}

function coinConfigIconText(config: AssetCoinConfig) {
  return (config.iconText || config.symbol || config.coin || '?').slice(0, 3).toUpperCase()
}

function selectAssetAction(key: AssetActionKey) {
  activeAssetAction.value = key
}

function activeCoinConfig() {
  return displayedAssets.value[0]?.config
}

function openCoinActions(config: AssetCoinConfig) {
  selectedCoinConfig.value = config
}

function closeCoinActions() {
  selectedCoinConfig.value = null
}

function openAssetFlow(action: 'recharge' | 'withdraw' | 'transfer', direction?: 'in' | 'out') {
  const config = selectedCoinConfig.value || activeCoinConfig()
  if (!config) return

  router.push({
    path: `/assets/${action}`,
    query: {
      coin: config.coin,
      walletType: String(activeWalletType.value),
      direction,
    },
  })
}

async function loadAssetSummary() {
  if (!getAccessToken()) {
    summaryAssets.value = []
    return
  }

  summaryLoading.value = true
  try {
    const resp = await apiGetMyAssetSummary({})
    if (isSuccessCode(resp.code)) {
      summaryAssets.value = resp.data?.assets || []
    }
  } catch (error) {
    console.warn('get my asset summary failed', error)
  } finally {
    summaryLoading.value = false
  }
}

async function loadCoinConfigs() {
  coinConfigs.value = []

  if (!getAccessToken()) {
    return
  }

  coinConfigsLoading.value = true
  try {
    const resp = await apiListAssetCoinConfigs({
      walletType: activeWalletType.value,
      operationType: 0,
    })

    if (!isSuccessCode(resp.code)) {
      return
    }

    coinConfigs.value = resp.data || []
  } catch (error) {
    console.warn('list asset coin configs failed', error)
  } finally {
    coinConfigsLoading.value = false
  }
}

function handleAssetAction(key: AssetActionKey) {
  if (key === 'flows') {
    router.push('/assets/flows')
    return
  }
  if (key === 'bankWithdraw') {
    router.push({
      path: '/assets/withdraw',
      query: {
        method: 'bank',
        coin: 'USDT',
        walletType: String(activeWalletType.value),
      },
    })
    return
  }

  selectAssetAction(key)
  if (key === 'cryptoRecharge') openAssetFlow('recharge')
  if (key === 'cryptoWithdraw') openAssetFlow('withdraw')
  if (key === 'transfer') openAssetFlow('transfer')
}

watch([activeWalletType, activeOperationType], () => {
  void loadCoinConfigs()
})

watch(assetAccounts, (accounts) => {
  if (!accounts.some((account) => account.key === activeAssetAccount.value)) {
    activeAssetAccount.value = accounts[0]?.key || 'cash'
  }
})

onMounted(() => {
  void loadCoinConfigs()
  void loadAssetSummary()
})
</script>

<template>
  <CommonPage
    :show-back="false"
    :nav-height="76"
    :menus="pageMenus"
    :model-value="activeAssetAccount"
    @update:model-value="activeAssetAccount = String($event)"
  >
    <template #tabbar>
      <nav class="assets-page-tabbar" :aria-label="t('assets.pageLabel')">
        <button
          v-for="tab in pageTabs.slice(0, 2)"
          :key="tab.key"
          type="button"
          :class="{ active: activeTopTab === tab.key }"
          @click="activeTopTab = tab.key"
        >
          {{ t(tab.shortLabelKey) }}
        </button>
      </nav>
    </template>

    <template #custom>
      <section v-if="activeTopTab === 'assets'" class="asset-actions">
        <button
          v-for="action in assetActions"
          :key="action.key"
          type="button"
          :class="{ active: activeAssetAction === action.key }"
          @click="handleAssetAction(action.key)"
        >
          <span>{{ action.icon }}</span>
          <strong>{{ t(action.labelKey) }}</strong>
        </button>
      </section>
    </template>

    <section class="assets-page">
      <template v-if="activeTopTab === 'assets'">
        <div class="assets-center">
          <section class="assets-main-panel">
            <header class="asset-list-head">
              <h2>{{ activeAccountLabel }}</h2>
              <span>◎</span>
            </header>

            <section class="asset-coin-configs" aria-live="polite">
              <div v-if="coinConfigsLoading || summaryLoading" class="asset-coin-configs__state">
                {{ t('common.loading') }}
              </div>
              <div v-else class="asset-coin-list">
                <button
                  v-for="item in displayedAssets"
                  :key="item.config.id || item.config.coin"
                  type="button"
                  class="asset-coin-row"
                  @click="openCoinActions(item.config)"
                >
                  <span
                    class="asset-coin-row__icon"
                    :style="{
                      backgroundColor: item.config.iconBgColor || '#17391f',
                    }"
                  >
                    <img
                      v-if="item.config.iconUrl"
                      :src="item.config.iconUrl"
                      :alt="coinConfigName(item.config)"
                    >
                    <span v-else>{{ coinConfigIconText(item.config) }}</span>
                  </span>
                  <span class="asset-coin-row__main">
                    <strong>{{ item.config.coin }}</strong>
                  </span>
                  <span class="asset-coin-row__meta">{{ item.amount }}</span>
                </button>
              </div>
            </section>

            <LoginPrompt v-if="!getAccessToken()" :action-text="t('assets.viewAssets')" />
          </section>
        </div>

        <BottomDrawer
          :model-value="Boolean(selectedCoinConfig)"
          :title="t('assets.coinAction')"
          :close-label="t('common.close')"
          max-height="68dvh"
          :z-index="90"
          @update:model-value="
            (value) => {
              if (!value) closeCoinActions()
            }
          "
        >
          <div v-if="selectedCoinConfig" class="coin-action-sheet">
            <span
              class="coin-action-sheet__coin"
              :style="{
                backgroundColor: selectedCoinConfig.iconBgColor || '#16ad77',
              }"
            >
              <img
                v-if="selectedCoinConfig.iconUrl"
                :src="selectedCoinConfig.iconUrl"
                :alt="coinConfigName(selectedCoinConfig)"
              >
              <span v-else>{{ coinConfigIconText(selectedCoinConfig) }}</span>
            </span>
            <strong>{{ selectedCoinConfig.coin }}</strong>

            <div class="coin-action-sheet__grid">
              <button
                type="button"
                class="coin-action coin-action--recharge"
                @click="openAssetFlow('recharge')"
              >
                <span>⇣</span>
                <strong>{{ t('userMenu.recharge') }}</strong>
              </button>
              <button
                type="button"
                class="coin-action coin-action--withdraw"
                @click="openAssetFlow('withdraw')"
              >
                <span>⇡</span>
                <strong>{{ t('userMenu.withdraw') }}</strong>
              </button>
              <button
                type="button"
                class="coin-action coin-action--transfer-in"
                @click="openAssetFlow('transfer', 'in')"
              >
                <span>↯</span>
                <strong>{{ t('assets.transferIn') }}</strong>
              </button>
              <button
                type="button"
                class="coin-action coin-action--transfer-out"
                @click="openAssetFlow('transfer', 'out')"
              >
                <span>↟</span>
                <strong>{{ t('assets.transferOut') }}</strong>
              </button>
            </div>
          </div>
        </BottomDrawer>
      </template>

      <template v-else-if="activeTopTab === 'orders'">
        <nav class="assets-sub-tabs assets-sub-tabs--orders" :aria-label="t('assets.orderCategory')">
          <button
            v-for="scope in orderScopes"
            :key="scope.key"
            type="button"
            :class="{ active: activeOrderScope === scope.key }"
            @click="activeOrderScope = scope.key"
          >
            {{ t(scope.labelKey) }}
          </button>
        </nav>

        <nav class="assets-order-tabs" :aria-label="t('assets.orderStatus')">
          <button
            v-for="(tab, index) in orderTabs"
            :key="tab"
            type="button"
            :class="{ active: index === 0 }"
          >
            {{ t(tab) }}
          </button>
        </nav>

        <LoginPrompt :action-text="t('assets.viewData')" compact />
      </template>

      <template v-else>
        <LoginPrompt :action-text="t('assets.viewProfile')" compact />
      </template>
    </section>
  </CommonPage>
</template>

<style scoped>
.assets-page {
  min-height: calc(100dvh - 76px);
  padding: 18px 16px 96px;
  background: #0b0c15;
  color: #f6f7fb;
}

button {
  border: 0;
  background: transparent;
  color: inherit;
  font: inherit;
}

.assets-sub-tabs,
.assets-order-tabs {
  display: flex;
  flex-wrap: wrap;
  gap: 24px;
  overflow-x: hidden;
}

.assets-page-tabbar {
  display: flex;
  align-items: center;
  height: 76px;
  gap: 42px;
  padding: 0 24px;
  background: #0b0c15;
}

.assets-page-tabbar button {
  position: relative;
  flex: 0 0 auto;
  color: #8f929d;
  font-size: 16px;
  font-weight: 700;
  white-space: nowrap;
}

.assets-page-tabbar button.active {
  color: #fff;
  font-size: 17px;
  font-weight: 800;
}

.assets-center {
  width: 100%;
}

.asset-actions {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  margin: 0;
  border-top: 1px solid #242633;
  border-bottom: 1px solid #242633;
}

.asset-actions button {
  display: grid;
  min-height: 112px;
  place-items: center;
  align-content: center;
  gap: 10px;
  border-right: 1px solid #242633;
}

.asset-actions button:disabled {
  cursor: wait;
  opacity: 0.72;
}

.asset-actions button.active span {
  background: rgba(2, 185, 4, 0.14);
}

.asset-actions button:nth-child(3n) {
  border-right: 0;
}

.asset-actions button:nth-child(n + 4) {
  border-top: 1px solid #242633;
}

.asset-actions span {
  display: grid;
  width: 34px;
  height: 34px;
  place-items: center;
  border: 2px solid #02b904;
  border-radius: 12px;
  color: #02b904;
  font-size: 18px;
}

.asset-actions strong {
  font-size: 14px;
  font-weight: 500;
}

:deep(.sub-menu) {
  justify-content: flex-start;
  gap: 28px;
  overflow-x: auto;
  overflow-y: hidden;
  padding: 0 20px;
  border-bottom: 1px solid #242633;
  background: #0b0c15;
  -webkit-overflow-scrolling: touch;
  scrollbar-width: none;
}

:deep(.sub-menu::-webkit-scrollbar) {
  display: none;
}

:deep(.sub-menu-item) {
  flex: 0 0 auto;
  justify-content: flex-start;
  color: #8f929d;
  font-size: 16px;
  font-weight: 700;
  white-space: nowrap;
}

:deep(.sub-menu-item.active) {
  color: #fff;
  font-weight: 800;
}

:deep(.sub-menu-item.active::after) {
  right: 8px;
  bottom: 0;
  left: 8px;
  width: auto;
  background: #02b904;
}

.asset-list-head {
  display: none;
}

.asset-list-head h2 {
  margin: 0;
  font-size: 20px;
}

.asset-list-head span {
  color: #8f929d;
}

.asset-coin-configs {
  margin: 18px -16px 0;
  border-bottom: 1px solid #242633;
}

.asset-coin-configs__state {
  min-height: 76px;
  display: grid;
  place-items: center;
  color: #8f929d;
  font-size: 13px;
}

.asset-coin-list {
  display: grid;
  gap: 0;
}

.asset-coin-row {
  display: grid;
  grid-template-columns: 44px minmax(0, 1fr) auto;
  align-items: center;
  gap: 12px;
  min-height: 66px;
  padding: 10px 24px;
  border-top: 1px solid #242633;
  text-align: left;
}

.asset-coin-row__icon {
  display: grid;
  width: 34px;
  height: 34px;
  place-items: center;
  overflow: hidden;
  border-radius: 50%;
  color: #fff;
  font-size: 11px;
  font-weight: 700;
}

.asset-coin-row__icon img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.asset-coin-row__main {
  display: grid;
  min-width: 0;
  gap: 4px;
}

.asset-coin-row__main strong {
  overflow: hidden;
  color: #f6f7fb;
  font-size: 14px;
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.asset-coin-row__main small,
.asset-coin-row__meta {
  color: #8f929d;
  font-size: 12px;
}

.asset-coin-row__meta {
  white-space: nowrap;
  color: #f6f7fb;
  font-size: 15px;
  font-weight: 700;
}

.assets-sub-tabs {
  margin: 0 -16px;
  padding: 0 16px;
  border-bottom: 1px solid #242633;
  flex-wrap: nowrap;
  overflow-x: auto;
  overflow-y: hidden;
  -webkit-overflow-scrolling: touch;
  scrollbar-width: none;
}

.assets-sub-tabs::-webkit-scrollbar {
  display: none;
}

.assets-sub-tabs--orders {
  margin-top: 48px;
}

.assets-sub-tabs button,
.assets-order-tabs button {
  position: relative;
  flex: 0 0 auto;
  padding: 0 0 12px;
  color: #8f929d;
  font-size: 16px;
  font-weight: 500;
  white-space: nowrap;
}

.assets-sub-tabs button.active,
.assets-order-tabs button.active {
  color: #fff;
  font-weight: 600;
}

.assets-sub-tabs button.active::after,
.assets-order-tabs button.active::after {
  position: absolute;
  right: 8px;
  bottom: 0;
  left: 8px;
  height: 3px;
  border-radius: 999px;
  background: #02b904;
  content: '';
}

.assets-order-tabs {
  margin-top: 24px;
  padding-bottom: 14px;
  border-bottom: 1px solid #242633;
}

.coin-action-sheet {
  display: grid;
  justify-items: center;
  width: 100%;
  min-height: 356px;
  padding-top: 10px;
}

.coin-action-sheet__coin {
  display: grid;
  width: 68px;
  height: 68px;
  place-items: center;
  overflow: hidden;
  border-radius: 50%;
  color: #fff;
  font-size: 18px;
  font-weight: 800;
}

.coin-action-sheet__coin img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.coin-action-sheet > strong {
  margin: 12px 0 24px;
  font-size: 20px;
}

.coin-action-sheet__grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
  width: 100%;
}

.coin-action {
  display: grid;
  min-height: 104px;
  place-items: center;
  align-content: center;
  gap: 10px;
  border-radius: 16px;
  background: #233329;
}

.coin-action span {
  display: grid;
  width: 44px;
  height: 44px;
  place-items: center;
  border-radius: 16px;
  background: #49c9ff;
  color: #071018;
  font-size: 26px;
  font-weight: 800;
}

.coin-action strong {
  font-size: 18px;
}

.coin-action--recharge {
  color: #49c9ff;
}

.coin-action--withdraw {
  background: #3a2a27;
  color: #ff8a22;
}

.coin-action--withdraw span {
  background: #ff9e3d;
}

.coin-action--transfer-in {
  color: #22df8d;
}

.coin-action--transfer-in span {
  background: #00b80c;
  color: #fff;
}

.coin-action--transfer-out {
  background: #3b262b;
  color: #ff4d43;
}

.coin-action--transfer-out span {
  background: #e75a5b;
  color: #071018;
}

@media (max-width: 520px) {
  .assets-page {
    padding: 18px 18px 96px;
  }

  .asset-actions button {
    min-height: 86px;
  }

  .assets-sub-tabs,
  .asset-coin-configs {
    margin-right: -18px;
    margin-left: -18px;
  }
}

@media (max-width: 390px) {
  .assets-page {
    padding: 16px 14px 92px;
  }

  .asset-actions button {
    min-height: 76px;
    gap: 8px;
  }

  .asset-actions span {
    width: 30px;
    height: 30px;
    border-radius: 10px;
    font-size: 16px;
  }

  .asset-actions strong {
    font-size: 13px;
  }

  .assets-sub-tabs,
  .asset-coin-configs {
    margin-right: -14px;
    margin-left: -14px;
  }

  .asset-coin-row {
    grid-template-columns: 36px minmax(0, 1fr) minmax(0, auto);
    gap: 10px;
    padding-right: 14px;
    padding-left: 14px;
  }

  .asset-coin-row__meta {
    max-width: 112px;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .coin-action-sheet__grid {
    gap: 8px;
  }

  .coin-action {
    min-height: 86px;
    border-radius: 14px;
  }

  .coin-action span {
    width: 36px;
    height: 36px;
    border-radius: 12px;
    font-size: 22px;
  }

  .coin-action strong {
    font-size: 14px;
  }
}
</style>
