<script setup lang="ts">
import { ref } from 'vue'

import { getAccessToken } from '@/api/http'
import { apiCreateRechargeOrder, apiListAvailableRechargeChannels } from '@/api/payment'

// 资产页：展示资产中心和订单中心的移动端占位结构。
type AssetTopTab = 'assets' | 'orders'

const RECHARGE_AMOUNT = 1000
const RECHARGE_CURRENCY = 'USDT'
const CLIENT_TYPE_H5 = 2

const activeTopTab = ref<AssetTopTab>('assets')
const activeAssetAccount = ref('cash')
const activeOrderScope = ref('stock')
const rechargeLoading = ref(false)
const rechargeMessage = ref('')

const assetActions = [
  { key: 'recharge', label: '加密货币充值', icon: '$' },
  { key: 'withdraw', label: '加密货币提现', icon: '$' },
  { key: 'transfer', label: '账户划转', icon: '▭' },
  { key: 'flows', label: '资金记录', icon: '▤' },
]

const assetAccounts = [
  { key: 'cash', label: '现金账户' },
  { key: 'stock', label: '股票账户' },
  { key: 'contract', label: '合约账户' },
]

const orderScopes = [
  { key: 'stock', label: '股票' },
  { key: 'contract', label: '合约订单' },
  { key: 'option', label: '期权合约' },
]

const orderTabs = ['当前持仓', '历史查询', '盘前订单']

function isSuccessCode(code: number) {
  return code === 0 || code === 200
}

async function handleAssetAction(key: string) {
  if (key !== 'recharge') return
  if (rechargeLoading.value) return

  rechargeMessage.value = ''
  if (!getAccessToken()) {
    rechargeMessage.value = '请先登录'
    return
  }

  rechargeLoading.value = true
  try {
    const channelsResp = await apiListAvailableRechargeChannels({
      rechargeAmount: RECHARGE_AMOUNT,
      currency: RECHARGE_CURRENCY,
      clientType: CLIENT_TYPE_H5,
    })

    if (!isSuccessCode(channelsResp.code)) {
      rechargeMessage.value = channelsResp.msg || '充值通道获取失败'
      return
    }

    const channel = channelsResp.data?.[0]
    if (!channel) {
      rechargeMessage.value = '暂无可用充值通道'
      return
    }

    const orderResp = await apiCreateRechargeOrder({
      channelId: channel.channelId,
      rechargeAmount: RECHARGE_AMOUNT,
      currency: channel.currency || RECHARGE_CURRENCY,
      subject: `充值 ${RECHARGE_AMOUNT}`,
      body: '加密货币充值',
      clientType: CLIENT_TYPE_H5,
    })

    if (!isSuccessCode(orderResp.code)) {
      rechargeMessage.value = orderResp.msg || '充值订单创建失败'
      return
    }

    const payUrl = orderResp.data?.payUrl
    if (payUrl) {
      window.location.href = payUrl
      return
    }

    rechargeMessage.value = orderResp.data?.orderNo ? `充值订单已创建：${orderResp.data.orderNo}` : '充值订单已创建'
  } catch (error) {
    console.warn('create recharge order failed', error)
    rechargeMessage.value = '充值失败'
  } finally {
    rechargeLoading.value = false
  }
}
</script>

<template>
  <section class="assets-page">
    <nav class="assets-top-tabs" aria-label="资产页面">
      <button type="button" :class="{ active: activeTopTab === 'assets' }" @click="activeTopTab = 'assets'">资产</button>
      <button type="button" :class="{ active: activeTopTab === 'orders' }" @click="activeTopTab = 'orders'">订单中心</button>
    </nav>

    <template v-if="activeTopTab === 'assets'">
      <section class="asset-actions">
        <button
          v-for="action in assetActions"
          :key="action.key"
          type="button"
          :disabled="action.key === 'recharge' && rechargeLoading"
          :aria-busy="action.key === 'recharge' && rechargeLoading"
          @click="handleAssetAction(action.key)"
        >
          <span>{{ action.icon }}</span>
          <strong>{{ action.key === 'recharge' && rechargeLoading ? '充值中' : action.label }}</strong>
        </button>
      </section>

      <p v-if="rechargeMessage" class="asset-recharge-message">{{ rechargeMessage }}</p>

      <nav class="assets-sub-tabs" aria-label="账户类型">
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

      <div class="asset-login-tip"><span>登录</span>或<span>注册</span>查看资产</div>
    </template>

    <template v-else>
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
  </section>
</template>

<style scoped>
.assets-page {
  min-height: calc(100dvh - 72px);
  padding: 24px 16px 112px;
  background: #0b0c15;
  color: #f6f7fb;
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
  gap: 28px;
  overflow-x: auto;
}

.assets-top-tabs {
  margin-bottom: 42px;
}

.assets-top-tabs button {
  flex: 0 0 auto;
  color: #8f929d;
  font-size: 20px;
  font-weight: 500;
  white-space: nowrap;
}

.assets-top-tabs button.active {
  color: #fff;
  font-size: 24px;
  font-weight: 600;
}

.asset-actions {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  margin: 0 -16px 82px;
  border-top: 1px solid #242633;
  border-bottom: 1px solid #242633;
}

.asset-actions button {
  display: grid;
  min-height: 150px;
  place-items: center;
  align-content: center;
  gap: 18px;
  border-right: 1px solid #242633;
}

.asset-actions button:disabled {
  cursor: wait;
  opacity: 0.72;
}

.asset-actions button:nth-child(3n) {
  border-right: 0;
}

.asset-actions button:nth-child(n + 4) {
  border-top: 1px solid #242633;
}

.asset-actions span {
  display: grid;
  width: 42px;
  height: 42px;
  place-items: center;
  border: 2px solid #02b904;
  border-radius: 12px;
  color: #02b904;
  font-size: 24px;
}

.asset-actions strong {
  font-size: 17px;
  font-weight: 500;
}

.asset-recharge-message {
  margin: -58px 0 54px;
  color: #02b904;
  font-size: 15px;
  text-align: center;
}

.assets-sub-tabs {
  margin: 0 -16px;
  padding: 0 16px;
  border-bottom: 1px solid #242633;
}

.assets-sub-tabs--orders {
  margin-top: 48px;
}

.assets-sub-tabs button,
.assets-order-tabs button {
  position: relative;
  flex: 0 0 auto;
  padding: 0 0 16px;
  color: #8f929d;
  font-size: 19px;
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
  font-size: 17px;
}

.asset-login-tip--orders {
  min-height: 160px;
}

.asset-login-tip span {
  color: #02b904;
}

@media (max-width: 520px) {
  .asset-actions button {
    min-height: 122px;
  }

  .assets-top-tabs button.active {
    font-size: 22px;
  }
}
</style>
