<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'

import { apiGetAssetOptions, apiGetMyAssetSummary, apiListAssetCoinConfigs } from '@/api/asset'
import { getAccessToken } from '@/api/http'
import { useDevice } from '@/composables/useDevice'
import { optionText, useOptions } from '@/composables/useOptions'
import type { AssetCoinConfig, AssetUserAsset } from '@/types/asset'

// 资产页：展示资产中心和订单中心的移动端占位结构。
type AssetTopTab = 'assets' | 'orders' | 'profile'
type AssetActionKey = 'cryptoRecharge' | 'bankRecharge' | 'cryptoWithdraw' | 'bankWithdraw' | 'transfer' | 'flows'

const ASSET_OPERATION_TYPES: Partial<Record<AssetActionKey, number>> = {
  cryptoRecharge: 1,
  cryptoWithdraw: 2,
  transfer: 3,
}

const { isDesktop } = useDevice()
const router = useRouter()
const assetOptions = useOptions(apiGetAssetOptions)
const activeTopTab = ref<AssetTopTab>('assets')
const activeAssetAccount = ref('cash')
const activeAssetAction = ref<AssetActionKey>('cryptoRecharge')
const activeOrderScope = ref('stock')
const coinConfigs = ref<AssetCoinConfig[]>([])
const coinConfigsLoading = ref(false)
const coinConfigsMessage = ref('')
const summaryAssets = ref<AssetUserAsset[]>([])
const summaryLoading = ref(false)
const selectedCoinConfig = ref<AssetCoinConfig | null>(null)

const assetActions: Array<{ key: AssetActionKey; label: string; icon: string }> = [
  { key: 'cryptoRecharge', label: '加密货币充值', icon: '$' },
  { key: 'bankRecharge', label: '银行卡充值', icon: '▣' },
  { key: 'cryptoWithdraw', label: '加密货币提现', icon: '$' },
  { key: 'bankWithdraw', label: '银行卡提现', icon: '▭' },
  { key: 'transfer', label: '账户划转', icon: '⇄' },
  { key: 'flows', label: '资金记录', icon: '▤' },
]

const fallbackAssetAccounts = [
  { key: 'cash', label: '现金账户', walletType: 1 },
  { key: 'stock', label: '股票账户', walletType: 2 },
  { key: 'contract', label: '合约账户', walletType: 3 },
]

const orderScopes = [
  { key: 'stock', label: '股票' },
  { key: 'contract', label: '合约订单' },
  { key: 'option', label: '期权合约' },
]

const orderTabs = ['当前持仓', '历史查询', '盘前订单']

const pageTabs: Array<{ key: AssetTopTab; label: string }> = [
  { key: 'assets', label: '资产中心' },
  { key: 'orders', label: '订单中心' },
  { key: 'profile', label: '个人中心' },
]

const assetAccounts = computed(() => {
  const walletTypeOptions = assetOptions.getGroup('walletType')
    .filter((option) => option.value > 0)
    .map((option) => ({
      key: String(option.value),
      label: optionText(option),
      walletType: option.value,
    }))

  return walletTypeOptions.length ? walletTypeOptions : fallbackAssetAccounts
})

const activeWalletType = computed(() => {
  return assetAccounts.value.find((account) => account.key === activeAssetAccount.value)?.walletType || 1
})

const activeOperationType = computed(() => ASSET_OPERATION_TYPES[activeAssetAction.value])
const activeAccountLabel = computed(() => {
  return assetAccounts.value.find((account) => account.key === activeAssetAccount.value)?.label || '现金账户'
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
          decimalPlaces: 2,
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

  return configs.map((config) => ({
    config,
    amount: amountMap.get(config.coin)?.availableAmount || amountMap.get(config.coin)?.totalAmount || '0',
  }))
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
  if (isDesktop.value) return
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
  coinConfigsMessage.value = ''

  if (!getAccessToken()) {
    coinConfigsMessage.value = '请先登录'
    return
  }

  coinConfigsLoading.value = true
  try {
    const resp = await apiListAssetCoinConfigs({
      walletType: activeWalletType.value,
      operationType: 0,
    })

    if (!isSuccessCode(resp.code)) {
      coinConfigsMessage.value = resp.msg || '币种配置获取失败'
      return
    }

    coinConfigs.value = resp.data || []
    if (!coinConfigs.value.length) {
      coinConfigsMessage.value = '暂无可用币种'
    }
  } catch (error) {
    console.warn('list asset coin configs failed', error)
    coinConfigsMessage.value = '币种配置获取失败'
  } finally {
    coinConfigsLoading.value = false
  }
}

function handleAssetAction(key: AssetActionKey) {
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
  <section class="assets-page" :class="{ 'assets-page--desktop': isDesktop }">
    <nav class="assets-top-tabs" aria-label="资产页面">
      <button
        v-for="tab in isDesktop ? pageTabs : pageTabs.slice(0, 2)"
        :key="tab.key"
        type="button"
        :class="{ active: activeTopTab === tab.key }"
        @click="activeTopTab = tab.key"
      >
        {{ isDesktop ? tab.label : tab.label.replace('中心', '') }}
      </button>
    </nav>

    <template v-if="activeTopTab === 'assets'">
      <div class="assets-center">
        <aside v-if="isDesktop" class="assets-sidebar">
          <nav class="assets-account-card" aria-label="账户类型">
            <button
              v-for="account in assetAccounts"
              :key="account.key"
              type="button"
              :class="{ active: activeAssetAccount === account.key }"
              @click="activeAssetAccount = account.key"
            >
              {{ account.label }}
            </button>
          </nav>

          <nav class="assets-action-card" aria-label="资产操作">
            <button
              v-for="action in assetActions"
              :key="action.key"
              type="button"
              :class="{ active: activeAssetAction === action.key }"
              @click="handleAssetAction(action.key)"
            >
              {{ action.label }}
            </button>
          </nav>
        </aside>

        <section class="assets-main-panel">
          <section v-if="!isDesktop" class="asset-actions">
            <button
              v-for="action in assetActions"
              :key="action.key"
              type="button"
              :class="{ active: activeAssetAction === action.key }"
              @click="handleAssetAction(action.key)"
            >
              <span>{{ action.icon }}</span>
              <strong>{{ action.label }}</strong>
            </button>
          </section>

          <nav v-if="!isDesktop" class="assets-sub-tabs" aria-label="账户类型">
            <button
              v-for="account in assetAccounts"
              :key="account.key"
              type="button"
              :class="{ active: activeAssetAccount === account.key }"
              @click="activeAssetAccount = account.key"
            >
              {{ account.label }}
            </button>
          </nav>

          <header class="asset-list-head">
            <h2>{{ activeAccountLabel }}</h2>
            <span>◎</span>
          </header>

          <section class="asset-coin-configs" aria-live="polite">
            <div v-if="coinConfigsLoading || summaryLoading" class="asset-coin-configs__state">加载中</div>
            <div v-else-if="coinConfigsMessage && !displayedAssets.length" class="asset-coin-configs__state">
              {{ coinConfigsMessage }}
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
                  :style="{ backgroundColor: item.config.iconBgColor || '#17391f' }"
                >
                  <img v-if="item.config.iconUrl" :src="item.config.iconUrl" :alt="coinConfigName(item.config)" />
                  <span v-else>{{ coinConfigIconText(item.config) }}</span>
                </span>
                <span class="asset-coin-row__main">
                  <strong>{{ item.config.coin }}</strong>
                  <small v-if="isDesktop">{{ coinConfigName(item.config) }}</small>
                </span>
                <span class="asset-coin-row__meta">{{ item.amount }}</span>
              </button>
            </div>
          </section>

          <div v-if="!getAccessToken()" class="asset-login-tip"><span>登录</span>或<span>注册</span>查看资产</div>
        </section>
      </div>

      <div
        v-if="selectedCoinConfig"
        class="coin-action-overlay"
        role="presentation"
        @click.self="closeCoinActions"
      >
        <section class="coin-action-sheet" role="dialog" aria-modal="true" aria-label="币种操作">
          <button class="coin-action-sheet__close" type="button" aria-label="关闭" @click="closeCoinActions">×</button>
          <i class="coin-action-sheet__handle" />
          <h2>币种操作</h2>
          <span
            class="coin-action-sheet__coin"
            :style="{ backgroundColor: selectedCoinConfig.iconBgColor || '#16ad77' }"
          >
            <img v-if="selectedCoinConfig.iconUrl" :src="selectedCoinConfig.iconUrl" :alt="coinConfigName(selectedCoinConfig)" />
            <span v-else>{{ coinConfigIconText(selectedCoinConfig) }}</span>
          </span>
          <strong>{{ selectedCoinConfig.coin }}</strong>

          <div class="coin-action-sheet__grid">
            <button type="button" class="coin-action coin-action--recharge" @click="openAssetFlow('recharge')">
              <span>⇣</span>
              <strong>充值</strong>
            </button>
            <button type="button" class="coin-action coin-action--withdraw" @click="openAssetFlow('withdraw')">
              <span>⇡</span>
              <strong>提现</strong>
            </button>
            <button type="button" class="coin-action coin-action--transfer-in" @click="openAssetFlow('transfer', 'in')">
              <span>↯</span>
              <strong>转入</strong>
            </button>
            <button type="button" class="coin-action coin-action--transfer-out" @click="openAssetFlow('transfer', 'out')">
              <span>↟</span>
              <strong>转出</strong>
            </button>
          </div>
        </section>
      </div>
    </template>

    <template v-else-if="activeTopTab === 'orders'">
      <nav class="assets-sub-tabs assets-sub-tabs--orders" aria-label="订单分类">
        <button
          v-for="scope in orderScopes"
          :key="scope.key"
          type="button"
          :class="{ active: activeOrderScope === scope.key }"
          @click="activeOrderScope = scope.key"
        >
          {{ scope.label }}
        </button>
      </nav>

      <nav class="assets-order-tabs" aria-label="订单状态">
        <button v-for="(tab, index) in orderTabs" :key="tab" type="button" :class="{ active: index === 0 }">
          {{ tab }}
        </button>
      </nav>

      <div class="asset-login-tip asset-login-tip--orders"><span>登录</span>或<span>注册</span>查看数据</div>
    </template>

    <template v-else>
      <div class="asset-login-tip asset-login-tip--orders"><span>登录</span>或<span>注册</span>查看资料</div>
    </template>
  </section>
</template>

<style scoped>
.assets-page {
  min-height: calc(100dvh - 72px);
  padding: 18px 16px 96px;
  background: #0b0c15;
  color: #f6f7fb;
}

.assets-page--desktop {
  min-height: calc(100dvh - 76px);
  padding: 0 0 80px;
  background: #171a23;
}

button {
  border: 0;
  background: transparent;
  color: inherit;
  font: inherit;
}

.assets-top-tabs,
.assets-sub-tabs,
.assets-order-tabs {
  display: flex;
  gap: 24px;
  overflow-x: auto;
}

.assets-top-tabs {
  margin-bottom: 30px;
}

.assets-page--desktop .assets-top-tabs {
  justify-content: center;
  gap: 88px;
  margin: 0;
  padding: 28px 24px 0;
  min-height: 90px;
  border-bottom: 1px solid #2a2d38;
}

.assets-top-tabs button {
  flex: 0 0 auto;
  color: #8f929d;
  font-size: 16px;
  font-weight: 500;
  white-space: nowrap;
}

.assets-top-tabs button.active {
  color: #fff;
  font-size: 18px;
  font-weight: 600;
}

.assets-page--desktop .assets-top-tabs button {
  font-size: 20px;
  padding-bottom: 26px;
}

.assets-page--desktop .assets-top-tabs button.active {
  font-size: 20px;
}

.assets-page--desktop .assets-top-tabs button.active::after {
  position: absolute;
  right: 36px;
  bottom: 0;
  left: 36px;
  height: 3px;
  border-radius: 999px;
  background: #02b904;
  content: '';
}

.assets-center {
  width: 100%;
}

.assets-page--desktop .assets-center {
  display: grid;
  grid-template-columns: 306px minmax(0, 1fr);
  gap: 42px;
  max-width: 1640px;
  margin: 0 auto;
  padding: 56px 48px 0;
}

.assets-sidebar {
  display: grid;
  align-content: start;
  gap: 14px;
}

.assets-account-card,
.assets-action-card {
  display: grid;
  overflow: hidden;
  border-radius: 24px;
  background: #272a34;
}

.assets-account-card button,
.assets-action-card button {
  min-height: 72px;
  color: #fff;
  font-size: 16px;
  font-weight: 600;
}

.assets-account-card button {
  position: relative;
}

.assets-account-card button.active,
.assets-action-card button.active {
  background: #3d404b;
  color: #02b904;
}

.assets-account-card button.active::before {
  position: absolute;
  top: 22px;
  bottom: 22px;
  left: 0;
  width: 12px;
  border-radius: 0 999px 999px 0;
  background: #02b904;
  content: '';
}

.asset-actions {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  margin: 0 -16px 54px;
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

.asset-list-head {
  display: none;
}

.assets-page--desktop .asset-list-head {
  display: flex;
  align-items: center;
  gap: 12px;
  min-height: 56px;
  border-bottom: 1px solid #2a2d38;
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

.assets-page--desktop .asset-coin-configs {
  margin: 34px 0 0;
  border-bottom: 0;
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

.assets-page--desktop .asset-coin-list {
  gap: 14px;
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

.assets-page--desktop .asset-coin-row {
  grid-template-columns: 210px minmax(0, 1fr) auto;
  gap: 38px;
  min-height: 72px;
  overflow: hidden;
  border: 0;
  border-radius: 16px;
  background: #282b35;
  font-size: 16px;
}

.assets-page--desktop .asset-coin-row__icon {
  margin-left: 0;
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

.assets-page--desktop .asset-coin-row__main {
  display: flex;
  align-items: center;
  min-height: 72px;
  margin: -12px 0 -12px -38px;
  padding-left: 72px;
  background: #454852;
}

.asset-coin-row__main strong {
  overflow: hidden;
  color: #f6f7fb;
  font-size: 14px;
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.assets-page--desktop .asset-coin-row__main strong {
  font-size: 17px;
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
}

.assets-page--desktop .assets-sub-tabs--orders,
.assets-page--desktop .assets-order-tabs,
.assets-page--desktop .asset-login-tip--orders {
  max-width: 1200px;
  margin-right: auto;
  margin-left: auto;
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

.asset-login-tip {
  display: grid;
  min-height: 210px;
  place-items: center;
  color: #8f929d;
  font-size: 14px;
}

.asset-login-tip--orders {
  min-height: 160px;
}

.asset-login-tip span {
  color: #02b904;
}

.coin-action-overlay {
  position: fixed;
  inset: 0;
  z-index: 90;
  display: grid;
  align-items: end;
  background: rgba(0, 0, 0, 0.68);
  backdrop-filter: blur(12px);
}

.coin-action-sheet {
  position: relative;
  display: grid;
  justify-items: center;
  width: 100%;
  min-height: 356px;
  padding: 14px 22px calc(22px + env(safe-area-inset-bottom));
  border-radius: 24px 24px 0 0;
  background: #262832;
  box-shadow: 0 -18px 60px rgba(0, 0, 0, 0.38);
}

.coin-action-sheet__handle {
  width: 46px;
  height: 5px;
  border-radius: 999px;
  background: #a0a1a8;
}

.coin-action-sheet__close {
  position: absolute;
  top: 22px;
  right: 28px;
  border: 0;
  color: #fff;
  font-size: 34px;
  line-height: 1;
}

.coin-action-sheet h2 {
  margin: 22px 0 20px;
  font-size: 22px;
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
    padding: 18px 30px 96px;
  }

  .asset-actions button {
    min-height: 96px;
  }

  .asset-actions {
    margin-right: -30px;
    margin-left: -30px;
  }

  .assets-sub-tabs,
  .asset-coin-configs {
    margin-right: -30px;
    margin-left: -30px;
  }

  .assets-top-tabs button.active {
    font-size: 18px;
  }
}
</style>
