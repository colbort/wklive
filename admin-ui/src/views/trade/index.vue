<template>
  <div class="module-page">
    <div class="page-header">
      <h2>交易管理</h2>
      <div class="header-actions">
        <el-button @click="loadCurrent">
          刷新
        </el-button>
        <el-button v-if="activeTab === 'symbols'" type="primary" @click="openSymbolDialog()">
          新增交易对
        </el-button>
      </div>
    </div>

    <el-tabs v-model="activeTab" @tab-change="loadCurrent">
      <el-tab-pane label="交易对" name="symbols" />
      <el-tab-pane label="订单" name="orders" />
      <el-tab-pane label="成交" name="fills" />
      <el-tab-pane label="持仓" name="positions" />
      <el-tab-pane label="持仓历史" name="histories" />
      <el-tab-pane label="保证金账户" name="margin-accounts" />
      <el-tab-pane label="撤单日志" name="cancel-logs" />
      <el-tab-pane label="风控" name="risk" />
      <el-tab-pane label="事件" name="events" />
    </el-tabs>

    <template v-if="activeTab !== 'risk'">
      <el-card shadow="never" class="query-card">
        <el-form :model="currentQuery" inline label-width="90px">
          <el-form-item v-for="field in currentFields" :key="field.key" :label="field.label">
            <el-input v-if="field.type !== 'number'" v-model="currentQuery[field.key]" clearable />
            <el-input-number
              v-else
              v-model="currentQuery[field.key]"
              :min="0"
              :precision="0"
            />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="loadCurrent">
              查询
            </el-button>
            <el-button @click="resetCurrent">
              重置
            </el-button>
          </el-form-item>
        </el-form>
      </el-card>

      <el-card shadow="never" class="table-card">
        <el-table v-loading="loading" :data="rows" stripe>
          <el-table-column
            v-for="column in currentColumns"
            :key="column.prop"
            :prop="column.prop"
            :label="column.label"
            :min-width="column.width || 140"
            show-overflow-tooltip
          />
          <el-table-column label="操作" width="240" fixed="right">
            <template #default="{ row }">
              <el-button link type="primary" @click="showDetail(row)">
                详情
              </el-button>
              <el-button
                v-if="activeTab === 'symbols'"
                link
                type="primary"
                @click="openSymbolDialog(row)"
              >
                编辑
              </el-button>
              <el-button
                v-if="activeTab === 'symbols'"
                link
                type="primary"
                @click="openSpotDialog(row)"
              >
                现货配置
              </el-button>
              <el-button
                v-if="activeTab === 'symbols'"
                link
                type="primary"
                @click="openContractDialog(row)"
              >
                合约配置
              </el-button>
              <el-button
                v-if="activeTab === 'events'"
                link
                type="warning"
                @click="retryEvent(row)"
              >
                重试
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </template>

    <template v-else>
      <el-card shadow="never" class="query-card">
        <template #header>
          风险配置查询
        </template>
        <el-form :model="riskQuery" inline label-width="90px">
          <el-form-item label="租户ID">
            <el-input-number v-model="riskQuery.tenantId" :min="0" :precision="0" />
          </el-form-item>
          <el-form-item label="用户ID">
            <el-input-number v-model="riskQuery.userId" :min="0" :precision="0" />
          </el-form-item>
          <el-form-item label="交易对ID">
            <el-input-number v-model="riskQuery.symbolId" :min="0" :precision="0" />
          </el-form-item>
          <el-form-item label="市场类型">
            <el-input-number v-model="riskQuery.marketType" :min="0" :precision="0" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="loadRiskData">
              加载配置
            </el-button>
          </el-form-item>
        </el-form>
      </el-card>

      <div class="risk-grid">
        <el-card shadow="never">
          <template #header>
            用户交易限制
          </template>
          <el-form label-width="120px">
            <el-form-item label="可开仓">
              <el-switch v-model="tradeLimitForm.canOpen" />
            </el-form-item>
            <el-form-item label="可平仓">
              <el-switch v-model="tradeLimitForm.canClose" />
            </el-form-item>
            <el-form-item label="可撤单">
              <el-switch v-model="tradeLimitForm.canCancel" />
            </el-form-item>
            <el-form-item label="仅减仓">
              <el-switch v-model="tradeLimitForm.onlyReduceOnly" />
            </el-form-item>
            <el-form-item label="交易开关">
              <el-switch v-model="tradeLimitForm.tradeEnabled" />
            </el-form-item>
            <el-form-item label="最大开单数">
              <el-input-number v-model="tradeLimitForm.maxOpenOrderCount" :min="0" :precision="0" />
            </el-form-item>
            <el-form-item label="最大持仓名义">
              <el-input v-model="tradeLimitForm.maxPositionNotional" />
            </el-form-item>
            <el-form-item label="操作人ID">
              <el-input-number v-model="tradeLimitForm.operatorId" :min="0" :precision="0" />
            </el-form-item>
            <el-button type="primary" :loading="submitLoading" @click="submitTradeLimit">
              保存
            </el-button>
          </el-form>
        </el-card>

        <el-card shadow="never">
          <template #header>
            用户交易对限制
          </template>
          <el-form label-width="120px">
            <el-form-item label="最大持仓量">
              <el-input v-model="symbolLimitForm.maxPositionQty" />
            </el-form-item>
            <el-form-item label="最大订单量">
              <el-input v-model="symbolLimitForm.maxOrderQty" />
            </el-form-item>
            <el-form-item label="最小订单量">
              <el-input v-model="symbolLimitForm.minOrderQty" />
            </el-form-item>
            <el-form-item label="价格偏离率">
              <el-input v-model="symbolLimitForm.priceDeviationRate" />
            </el-form-item>
            <el-form-item label="操作人ID">
              <el-input-number v-model="symbolLimitForm.operatorId" :min="0" :precision="0" />
            </el-form-item>
            <el-button type="primary" :loading="submitLoading" @click="submitSymbolLimit">
              保存
            </el-button>
          </el-form>
        </el-card>

        <el-card shadow="never">
          <template #header>
            用户交易配置
          </template>
          <el-form label-width="120px">
            <el-form-item label="仓位模式">
              <el-input-number v-model="tradeConfigForm.positionMode" :min="0" :precision="0" />
            </el-form-item>
            <el-form-item label="保证金模式">
              <el-input-number v-model="tradeConfigForm.marginMode" :min="0" :precision="0" />
            </el-form-item>
            <el-form-item label="默认杠杆">
              <el-input-number v-model="tradeConfigForm.defaultLeverage" :min="0" :precision="0" />
            </el-form-item>
            <el-form-item label="允许交易">
              <el-switch v-model="tradeConfigForm.tradeEnabled" />
            </el-form-item>
            <el-form-item label="只减仓">
              <el-switch v-model="tradeConfigForm.reduceOnlyEnabled" />
            </el-form-item>
            <el-button type="primary" :loading="submitLoading" @click="submitTradeConfig">
              保存
            </el-button>
          </el-form>
        </el-card>

        <el-card shadow="never">
          <template #header>
            杠杆配置
          </template>
          <el-form label-width="120px">
            <el-form-item label="保证金模式">
              <el-input-number v-model="leverageForm.marginMode" :min="0" :precision="0" />
            </el-form-item>
            <el-form-item label="仓位模式">
              <el-input-number v-model="leverageForm.positionMode" :min="0" :precision="0" />
            </el-form-item>
            <el-form-item label="多头杠杆">
              <el-input-number v-model="leverageForm.longLeverage" :min="0" :precision="0" />
            </el-form-item>
            <el-form-item label="空头杠杆">
              <el-input-number v-model="leverageForm.shortLeverage" :min="0" :precision="0" />
            </el-form-item>
            <el-form-item label="最大杠杆">
              <el-input-number v-model="leverageForm.maxLeverage" :min="0" :precision="0" />
            </el-form-item>
            <el-form-item label="操作人ID">
              <el-input-number v-model="leverageForm.operatorId" :min="0" :precision="0" />
            </el-form-item>
            <el-button type="primary" :loading="submitLoading" @click="submitLeverage">
              保存
            </el-button>
          </el-form>
        </el-card>
      </div>

      <el-card shadow="never" class="table-card">
        <template #header>
          风控校验日志
        </template>
        <el-form
          :model="riskLogQuery"
          inline
          label-width="90px"
          class="query-card-inner"
        >
          <el-form-item label="租户ID">
            <el-input-number v-model="riskLogQuery.tenantId" :min="0" :precision="0" />
          </el-form-item>
          <el-form-item label="用户ID">
            <el-input-number v-model="riskLogQuery.userId" :min="0" :precision="0" />
          </el-form-item>
          <el-form-item label="交易对ID">
            <el-input-number v-model="riskLogQuery.symbolId" :min="0" :precision="0" />
          </el-form-item>
          <el-form-item label="订单号">
            <el-input v-model="riskLogQuery.orderNo" clearable />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="loadRiskLogs">
              查询日志
            </el-button>
          </el-form-item>
        </el-form>
        <el-table v-loading="loading" :data="rows" stripe>
          <el-table-column
            prop="orderNo"
            label="订单号"
            min-width="180"
            show-overflow-tooltip
          />
          <el-table-column prop="userId" label="用户ID" width="100" />
          <el-table-column prop="symbolId" label="交易对ID" width="100" />
          <el-table-column prop="checkType" label="校验类型" width="100" />
          <el-table-column prop="checkResult" label="校验结果" width="100" />
          <el-table-column
            prop="rejectMsg"
            label="拒绝原因"
            min-width="220"
            show-overflow-tooltip
          />
          <el-table-column label="操作" width="100" fixed="right">
            <template #default="{ row }">
              <el-button link type="primary" @click="detailData = row; detailVisible = true">
                详情
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </template>

    <el-dialog v-model="symbolVisible" :title="symbolForm.id ? '编辑交易对' : '新增交易对'" width="760px">
      <el-form label-width="110px">
        <el-form-item label="租户ID">
          <el-input-number v-model="symbolForm.tenantId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item label="交易对">
          <el-input v-model="symbolForm.symbol" />
        </el-form-item>
        <el-form-item label="展示名称">
          <el-input v-model="symbolForm.displaySymbol" />
        </el-form-item>
        <el-form-item label="市场类型">
          <el-input-number v-model="symbolForm.marketType" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item label="基础资产">
          <el-input v-model="symbolForm.baseAsset" />
        </el-form-item>
        <el-form-item label="计价资产">
          <el-input v-model="symbolForm.quoteAsset" />
        </el-form-item>
        <el-form-item label="结算资产">
          <el-input v-model="symbolForm.settleAsset" />
        </el-form-item>
        <el-form-item label="合约类型">
          <el-input-number v-model="symbolForm.contractType" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item label="状态">
          <el-input-number v-model="symbolForm.status" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item label="价格位数">
          <el-input-number v-model="symbolForm.priceScale" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item label="数量位数">
          <el-input-number v-model="symbolForm.qtyScale" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item label="最小价格">
          <el-input v-model="symbolForm.minPrice" />
        </el-form-item>
        <el-form-item label="最大价格">
          <el-input v-model="symbolForm.maxPrice" />
        </el-form-item>
        <el-form-item label="价格步长">
          <el-input v-model="symbolForm.priceTick" />
        </el-form-item>
        <el-form-item label="最小数量">
          <el-input v-model="symbolForm.minQty" />
        </el-form-item>
        <el-form-item label="最大数量">
          <el-input v-model="symbolForm.maxQty" />
        </el-form-item>
        <el-form-item label="数量步长">
          <el-input v-model="symbolForm.qtyStep" />
        </el-form-item>
        <el-form-item label="最小名义">
          <el-input v-model="symbolForm.minNotional" />
        </el-form-item>
        <el-form-item label="最大杠杆">
          <el-input-number v-model="symbolForm.maxLeverage" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item label="开放时间">
          <el-input-number v-model="symbolForm.openTime" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item label="关闭时间">
          <el-input-number v-model="symbolForm.closeTime" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="symbolForm.sort" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="symbolForm.remark" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="symbolVisible = false">
          取消
        </el-button>
        <el-button type="primary" :loading="submitLoading" @click="submitSymbol">
          确定
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="spotVisible" title="现货配置" width="640px">
      <el-form label-width="110px">
        <el-form-item label="租户ID">
          <el-input-number v-model="spotForm.tenantId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item label="交易对ID">
          <el-input-number v-model="spotForm.symbolId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item label="Maker费率">
          <el-input v-model="spotForm.makerFeeRate" />
        </el-form-item>
        <el-form-item label="Taker费率">
          <el-input v-model="spotForm.takerFeeRate" />
        </el-form-item>
        <el-form-item label="允许买入">
          <el-switch v-model="spotForm.buyEnabled" />
        </el-form-item>
        <el-form-item label="允许卖出">
          <el-switch v-model="spotForm.sellEnabled" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="spotVisible = false">
          取消
        </el-button>
        <el-button type="primary" :loading="submitLoading" @click="submitSpotConfig">
          确定
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="contractVisible" title="合约配置" width="700px">
      <el-form label-width="120px">
        <el-form-item label="租户ID">
          <el-input-number v-model="contractForm.tenantId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item label="交易对ID">
          <el-input-number v-model="contractForm.symbolId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item label="合约面值">
          <el-input v-model="contractForm.contractSize" />
        </el-form-item>
        <el-form-item label="乘数">
          <el-input v-model="contractForm.multiplier" />
        </el-form-item>
        <el-form-item label="维持保证金率">
          <el-input v-model="contractForm.maintenanceMarginRate" />
        </el-form-item>
        <el-form-item label="初始保证金率">
          <el-input v-model="contractForm.initialMarginRate" />
        </el-form-item>
        <el-form-item label="Maker费率">
          <el-input v-model="contractForm.makerFeeRate" />
        </el-form-item>
        <el-form-item label="Taker费率">
          <el-input v-model="contractForm.takerFeeRate" />
        </el-form-item>
        <el-form-item label="资金费间隔">
          <el-input-number v-model="contractForm.fundingIntervalMinutes" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item label="交割时间">
          <el-input-number v-model="contractForm.deliveryTime" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item label="支持全仓">
          <el-switch v-model="contractForm.supportCross" />
        </el-form-item>
        <el-form-item label="支持逐仓">
          <el-switch v-model="contractForm.supportIsolated" />
        </el-form-item>
        <el-form-item label="允许买入">
          <el-switch v-model="contractForm.buyEnabled" />
        </el-form-item>
        <el-form-item label="允许卖出">
          <el-switch v-model="contractForm.sellEnabled" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="contractVisible = false">
          取消
        </el-button>
        <el-button type="primary" :loading="submitLoading" @click="submitContractConfig">
          确定
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="detailVisible" title="详情" width="760px">
      <pre class="detail-pre">{{ JSON.stringify(detailData, null, 2) }}</pre>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { tradeService } from '@/services'

const props = defineProps<{ initialTab?: string }>()
const activeTab = ref(props.initialTab || 'symbols')
const loading = ref(false)
const submitLoading = ref(false)
const rows = ref<Record<string, any>[]>([])
const detailVisible = ref(false)
const detailData = ref<Record<string, any>>({})
const symbolVisible = ref(false)
const spotVisible = ref(false)
const contractVisible = ref(false)

const queries = reactive({
  symbols: { tenantId: undefined, marketType: undefined, keyword: '', status: undefined, limit: 100 },
  orders: { tenantId: undefined, userId: undefined, symbolId: undefined, keyword: '', limit: 100 },
  fills: { tenantId: undefined, userId: undefined, symbolId: undefined, keyword: '', limit: 100 },
  positions: { tenantId: undefined, userId: undefined, symbolId: undefined, limit: 100 },
  histories: { tenantId: undefined, userId: undefined, symbolId: undefined, limit: 100 },
  'margin-accounts': { tenantId: undefined, userId: undefined, marketType: undefined, limit: 100 },
  'cancel-logs': { tenantId: undefined, userId: undefined, orderNo: '', limit: 100 },
  events: { tenantId: undefined, userId: undefined, symbolId: undefined, eventNo: '', limit: 100 },
})

const fieldMap: Record<string, Array<{ key: string; label: string; type?: string }>> = {
  symbols: [
    { key: 'tenantId', label: '租户ID', type: 'number' },
    { key: 'marketType', label: '市场类型', type: 'number' },
    { key: 'keyword', label: '关键字' },
  ],
  orders: [
    { key: 'tenantId', label: '租户ID', type: 'number' },
    { key: 'userId', label: '用户ID', type: 'number' },
    { key: 'symbolId', label: '交易对ID', type: 'number' },
  ],
  fills: [
    { key: 'tenantId', label: '租户ID', type: 'number' },
    { key: 'userId', label: '用户ID', type: 'number' },
    { key: 'symbolId', label: '交易对ID', type: 'number' },
  ],
  positions: [
    { key: 'tenantId', label: '租户ID', type: 'number' },
    { key: 'userId', label: '用户ID', type: 'number' },
    { key: 'symbolId', label: '交易对ID', type: 'number' },
  ],
  histories: [
    { key: 'tenantId', label: '租户ID', type: 'number' },
    { key: 'userId', label: '用户ID', type: 'number' },
    { key: 'symbolId', label: '交易对ID', type: 'number' },
  ],
  'margin-accounts': [
    { key: 'tenantId', label: '租户ID', type: 'number' },
    { key: 'userId', label: '用户ID', type: 'number' },
    { key: 'marketType', label: '市场类型', type: 'number' },
  ],
  'cancel-logs': [
    { key: 'tenantId', label: '租户ID', type: 'number' },
    { key: 'userId', label: '用户ID', type: 'number' },
    { key: 'orderNo', label: '订单号' },
  ],
  events: [
    { key: 'tenantId', label: '租户ID', type: 'number' },
    { key: 'userId', label: '用户ID', type: 'number' },
    { key: 'eventNo', label: '事件号' },
  ],
}

const columnMap: Record<string, Array<{ prop: string; label: string; width?: number }>> = {
  symbols: [
    { prop: 'id', label: 'ID', width: 80 },
    { prop: 'symbol', label: '交易对', width: 160 },
    { prop: 'displaySymbol', label: '展示名称', width: 160 },
    { prop: 'marketType', label: '市场类型', width: 100 },
    { prop: 'status', label: '状态', width: 100 },
    { prop: 'maxLeverage', label: '最大杠杆', width: 100 },
  ],
  orders: [
    { prop: 'orderNo', label: '订单号', width: 180 },
    { prop: 'userId', label: '用户ID', width: 100 },
    { prop: 'symbolId', label: '交易对ID', width: 100 },
    { prop: 'price', label: '价格' },
    { prop: 'qty', label: '数量' },
    { prop: 'status', label: '状态', width: 100 },
  ],
  fills: [
    { prop: 'fillNo', label: '成交号', width: 180 },
    { prop: 'orderNo', label: '订单号', width: 180 },
    { prop: 'userId', label: '用户ID', width: 100 },
    { prop: 'price', label: '价格' },
    { prop: 'qty', label: '数量' },
  ],
  positions: [
    { prop: 'userId', label: '用户ID', width: 100 },
    { prop: 'symbolId', label: '交易对ID', width: 100 },
    { prop: 'positionSide', label: '持仓方向', width: 100 },
    { prop: 'qty', label: '持仓量' },
    { prop: 'unrealizedPnl', label: '未实现盈亏' },
  ],
  histories: [
    { prop: 'positionId', label: '持仓ID', width: 120 },
    { prop: 'userId', label: '用户ID', width: 100 },
    { prop: 'symbolId', label: '交易对ID', width: 100 },
    { prop: 'actionType', label: '动作类型', width: 100 },
    { prop: 'realizedPnlDelta', label: '已实现盈亏' },
  ],
  'margin-accounts': [
    { prop: 'userId', label: '用户ID', width: 100 },
    { prop: 'marketType', label: '市场类型', width: 100 },
    { prop: 'marginAsset', label: '保证金币种', width: 120 },
    { prop: 'balance', label: '余额' },
    { prop: 'availableBalance', label: '可用余额' },
  ],
  'cancel-logs': [
    { prop: 'orderNo', label: '订单号', width: 180 },
    { prop: 'userId', label: '用户ID', width: 100 },
    { prop: 'cancelSource', label: '撤单来源', width: 100 },
    { prop: 'cancelReason', label: '撤单原因', width: 200 },
  ],
  events: [
    { prop: 'eventNo', label: '事件号', width: 180 },
    { prop: 'eventType', label: '事件类型', width: 140 },
    { prop: 'userId', label: '用户ID', width: 100 },
    { prop: 'eventStatus', label: '状态', width: 100 },
    { prop: 'retryCount', label: '重试次数', width: 100 },
  ],
}

const symbolForm = reactive<Record<string, any>>({
  id: 0,
  tenantId: 0,
  symbol: '',
  displaySymbol: '',
  marketType: 1,
  baseAsset: '',
  quoteAsset: '',
  settleAsset: '',
  contractType: 0,
  status: 1,
  priceScale: 2,
  qtyScale: 4,
  minPrice: '',
  maxPrice: '',
  priceTick: '',
  minQty: '',
  maxQty: '',
  qtyStep: '',
  minNotional: '',
  maxLeverage: 1,
  openTime: 0,
  closeTime: 0,
  sort: 0,
  remark: '',
})

const spotForm = reactive<Record<string, any>>({ tenantId: 0, symbolId: 0, makerFeeRate: '', takerFeeRate: '', buyEnabled: true, sellEnabled: true })
const contractForm = reactive<Record<string, any>>({ tenantId: 0, symbolId: 0, contractSize: '', multiplier: '', maintenanceMarginRate: '', initialMarginRate: '', makerFeeRate: '', takerFeeRate: '', fundingIntervalMinutes: 0, deliveryTime: 0, supportCross: true, supportIsolated: true, buyEnabled: true, sellEnabled: true })
const riskQuery = reactive<Record<string, any>>({ tenantId: 0, userId: 0, symbolId: 0, marketType: 0 })
const riskLogQuery = reactive<Record<string, any>>({ tenantId: 0, userId: 0, symbolId: 0, orderNo: '', limit: 100 })
const tradeLimitForm = reactive<Record<string, any>>({ tenantId: 0, userId: 0, marketType: 0, canOpen: true, canClose: true, canCancel: true, onlyReduceOnly: false, tradeEnabled: true, maxOpenOrderCount: 0, maxPositionNotional: '', operatorId: 0 })
const symbolLimitForm = reactive<Record<string, any>>({ tenantId: 0, userId: 0, symbolId: 0, marketType: 0, maxPositionQty: '', maxOrderQty: '', minOrderQty: '', priceDeviationRate: '', operatorId: 0 })
const tradeConfigForm = reactive<Record<string, any>>({ tenantId: 0, userId: 0, marketType: 0, symbolId: 0, positionMode: 0, marginMode: 0, defaultLeverage: 1, tradeEnabled: true, reduceOnlyEnabled: false })
const leverageForm = reactive<Record<string, any>>({ tenantId: 0, userId: 0, symbolId: 0, marketType: 0, marginMode: 0, positionMode: 0, longLeverage: 1, shortLeverage: 1, maxLeverage: 1, operatorId: 0 })

const currentQuery = computed(() => queries[activeTab.value as keyof typeof queries])
const currentFields = computed(() => fieldMap[activeTab.value] || [])
const currentColumns = computed(() => columnMap[activeTab.value] || [])
const pickList = (res: any) => res?.data || res?.list || []

const loadCurrent = async () => {
  if (activeTab.value === 'risk') {
    await loadRiskLogs()
    return
  }
  loading.value = true
  try {
    if (activeTab.value === 'symbols') rows.value = pickList(await tradeService.listSymbols(currentQuery.value))
    if (activeTab.value === 'orders') rows.value = pickList(await tradeService.listOrders(currentQuery.value))
    if (activeTab.value === 'fills') rows.value = pickList(await tradeService.listFills(currentQuery.value))
    if (activeTab.value === 'positions') rows.value = pickList(await tradeService.listPositions(currentQuery.value))
    if (activeTab.value === 'histories') rows.value = pickList(await tradeService.listPositionHistories(currentQuery.value))
    if (activeTab.value === 'margin-accounts') rows.value = pickList(await tradeService.listMarginAccounts(currentQuery.value))
    if (activeTab.value === 'cancel-logs') rows.value = pickList(await tradeService.listCancelLogs(currentQuery.value))
    if (activeTab.value === 'events') rows.value = pickList(await tradeService.listEvents(currentQuery.value))
  } finally {
    loading.value = false
  }
}

const resetCurrent = () => {
  Object.keys(currentQuery.value).forEach((key) => {
    currentQuery.value[key] = key === 'limit' ? 100 : ''
  })
  currentQuery.value.limit = 100
  loadCurrent()
}

const showDetail = async (row: Record<string, any>) => {
  if (activeTab.value === 'symbols') detailData.value = (await tradeService.getSymbol({ tenantId: row.tenantId, id: row.id })).data || row
  else if (activeTab.value === 'orders') detailData.value = (await tradeService.getOrder({ tenantId: row.tenantId, id: row.id })).data || row
  else if (activeTab.value === 'fills') detailData.value = (await tradeService.getFill({ tenantId: row.tenantId, id: row.id })).data || row
  else if (activeTab.value === 'positions') detailData.value = (await tradeService.getPosition({ tenantId: row.tenantId, id: row.id })).data || row
  else if (activeTab.value === 'events') detailData.value = (await tradeService.getEvent({ tenantId: row.tenantId, id: row.id })).data || row
  else detailData.value = row
  detailVisible.value = true
}

const openSymbolDialog = (row?: Record<string, any>) => {
  Object.assign(symbolForm, {
    id: 0,
    tenantId: 0,
    symbol: '',
    displaySymbol: '',
    marketType: 1,
    baseAsset: '',
    quoteAsset: '',
    settleAsset: '',
    contractType: 0,
    status: 1,
    priceScale: 2,
    qtyScale: 4,
    minPrice: '',
    maxPrice: '',
    priceTick: '',
    minQty: '',
    maxQty: '',
    qtyStep: '',
    minNotional: '',
    maxLeverage: 1,
    openTime: 0,
    closeTime: 0,
    sort: 0,
    remark: '',
  }, row || {})
  symbolVisible.value = true
}

const submitSymbol = async () => {
  submitLoading.value = true
  try {
    if (symbolForm.id) await tradeService.updateSymbol(symbolForm)
    else await tradeService.createSymbol(symbolForm)
    ElMessage.success('保存成功')
    symbolVisible.value = false
    loadCurrent()
  } finally {
    submitLoading.value = false
  }
}

const openSpotDialog = (row: Record<string, any>) => {
  Object.assign(spotForm, { tenantId: row.tenantId || 0, symbolId: row.id || row.symbolId || 0, makerFeeRate: '', takerFeeRate: '', buyEnabled: true, sellEnabled: true })
  spotVisible.value = true
}

const submitSpotConfig = async () => {
  submitLoading.value = true
  try {
    await tradeService.setSpotConfig(spotForm)
    ElMessage.success('现货配置已保存')
    spotVisible.value = false
  } finally {
    submitLoading.value = false
  }
}

const openContractDialog = (row: Record<string, any>) => {
  Object.assign(contractForm, { tenantId: row.tenantId || 0, symbolId: row.id || row.symbolId || 0, contractSize: '', multiplier: '', maintenanceMarginRate: '', initialMarginRate: '', makerFeeRate: '', takerFeeRate: '', fundingIntervalMinutes: 0, deliveryTime: 0, supportCross: true, supportIsolated: true, buyEnabled: true, sellEnabled: true })
  contractVisible.value = true
}

const submitContractConfig = async () => {
  submitLoading.value = true
  try {
    await tradeService.setContractConfig(contractForm)
    ElMessage.success('合约配置已保存')
    contractVisible.value = false
  } finally {
    submitLoading.value = false
  }
}

const loadRiskData = async () => {
  submitLoading.value = true
  try {
    Object.assign(tradeLimitForm, riskQuery, (await tradeService.getUserTradeLimit(riskQuery)).data || {})
    Object.assign(symbolLimitForm, riskQuery, (await tradeService.getUserSymbolLimit(riskQuery)).data || {})
    Object.assign(tradeConfigForm, riskQuery, (await tradeService.getUserTradeConfig(riskQuery)).data || {})
    Object.assign(leverageForm, riskQuery, (await tradeService.getUserLeverageConfig(riskQuery)).data || {})
    await loadRiskLogs()
  } finally {
    submitLoading.value = false
  }
}

const loadRiskLogs = async () => {
  loading.value = true
  try {
    rows.value = pickList(await tradeService.listRiskLogs(riskLogQuery))
  } finally {
    loading.value = false
  }
}

const submitTradeLimit = async () => {
  submitLoading.value = true
  try {
    await tradeService.setUserTradeLimit(tradeLimitForm)
    ElMessage.success('交易限制已保存')
  } finally {
    submitLoading.value = false
  }
}

const submitSymbolLimit = async () => {
  submitLoading.value = true
  try {
    await tradeService.setUserSymbolLimit(symbolLimitForm)
    ElMessage.success('交易对限制已保存')
  } finally {
    submitLoading.value = false
  }
}

const submitTradeConfig = async () => {
  submitLoading.value = true
  try {
    await tradeService.setUserTradeConfig(tradeConfigForm)
    ElMessage.success('交易配置已保存')
  } finally {
    submitLoading.value = false
  }
}

const submitLeverage = async () => {
  submitLoading.value = true
  try {
    await tradeService.setUserLeverageConfig(leverageForm)
    ElMessage.success('杠杆配置已保存')
  } finally {
    submitLoading.value = false
  }
}

const retryEvent = async (row: Record<string, any>) => {
  await tradeService.retryEvent({ tenantId: row.tenantId, id: row.id, eventNo: row.eventNo })
  ElMessage.success('事件已提交重试')
  loadCurrent()
}

onMounted(loadCurrent)
</script>

<style scoped></style>
