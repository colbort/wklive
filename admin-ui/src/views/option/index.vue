<template>
  <div class="module-page">
    <div class="page-header">
      <h2>期权管理</h2>
      <div class="header-actions">
        <el-button @click="loadCurrent">
          刷新
        </el-button>
        <el-button v-if="activeTab === 'contracts'" type="primary" @click="openContractDialog()">
          新增合约
        </el-button>
        <el-button
          v-if="activeTab === 'markets'"
          type="primary"
          plain
          @click="openMarketDialog()"
        >
          更新行情
        </el-button>
      </div>
    </div>

    <el-tabs v-model="activeTab" @tab-change="loadCurrent">
      <el-tab-pane label="合约" name="contracts" />
      <el-tab-pane label="市场快照" name="markets" />
      <el-tab-pane label="订单" name="orders" />
      <el-tab-pane label="成交" name="trades" />
      <el-tab-pane label="持仓" name="positions" />
      <el-tab-pane label="行权" name="exercises" />
      <el-tab-pane label="结算" name="settlements" />
      <el-tab-pane label="账户" name="accounts" />
      <el-tab-pane label="账单" name="bills" />
    </el-tabs>

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
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="showDetail(row)">
              详情
            </el-button>
            <el-button
              v-if="activeTab === 'contracts'"
              link
              type="primary"
              @click="openContractDialog(row)"
            >
              编辑
            </el-button>
            <el-button
              v-if="activeTab === 'contracts'"
              link
              type="primary"
              @click="openMarketDialog(row)"
            >
              行情
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="contractVisible" :title="contractForm.id ? '编辑合约' : '新增合约'" width="760px">
      <el-form label-width="110px">
        <el-form-item label="租户ID">
          <el-input-number v-model="contractForm.tenantId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item label="合约编码">
          <el-input v-model="contractForm.contractCode" />
        </el-form-item>
        <el-form-item label="标的">
          <el-input v-model="contractForm.underlyingSymbol" />
        </el-form-item>
        <el-form-item label="结算币">
          <el-input v-model="contractForm.settleCoin" />
        </el-form-item>
        <el-form-item label="计价币">
          <el-input v-model="contractForm.quoteCoin" />
        </el-form-item>
        <el-form-item label="期权类型">
          <el-input-number v-model="contractForm.optionType" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item label="行权方式">
          <el-input-number v-model="contractForm.exerciseStyle" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item label="结算方式">
          <el-input-number v-model="contractForm.settlementType" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item label="行权价">
          <el-input v-model="contractForm.strikePrice" />
        </el-form-item>
        <el-form-item label="合约单位">
          <el-input v-model="contractForm.contractUnit" />
        </el-form-item>
        <el-form-item label="最小下单量">
          <el-input v-model="contractForm.minOrderQty" />
        </el-form-item>
        <el-form-item label="最大下单量">
          <el-input v-model="contractForm.maxOrderQty" />
        </el-form-item>
        <el-form-item label="价格精度">
          <el-input v-model="contractForm.priceTick" />
        </el-form-item>
        <el-form-item label="数量步长">
          <el-input v-model="contractForm.qtyStep" />
        </el-form-item>
        <el-form-item label="乘数">
          <el-input v-model="contractForm.multiplier" />
        </el-form-item>
        <el-form-item label="上市时间">
          <el-input-number v-model="contractForm.listTime" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item label="到期时间">
          <el-input-number v-model="contractForm.expireTime" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item label="交割时间">
          <el-input-number v-model="contractForm.deliverTime" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item label="自动行权">
          <el-input-number v-model="contractForm.isAutoExercise" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item label="状态">
          <el-input-number v-model="contractForm.status" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="contractForm.sort" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="contractForm.remark" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="contractVisible = false">
          取消
        </el-button>
        <el-button type="primary" :loading="submitLoading" @click="submitContract">
          确定
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="marketVisible" title="更新行情" width="720px">
      <el-form label-width="110px">
        <el-form-item label="租户ID">
          <el-input-number v-model="marketForm.tenantId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item label="合约ID">
          <el-input-number v-model="marketForm.contractId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item label="标的价格">
          <el-input v-model="marketForm.underlyingPrice" />
        </el-form-item>
        <el-form-item label="标记价格">
          <el-input v-model="marketForm.markPrice" />
        </el-form-item>
        <el-form-item label="最新价">
          <el-input v-model="marketForm.lastPrice" />
        </el-form-item>
        <el-form-item label="买一价">
          <el-input v-model="marketForm.bidPrice" />
        </el-form-item>
        <el-form-item label="卖一价">
          <el-input v-model="marketForm.askPrice" />
        </el-form-item>
        <el-form-item label="理论价">
          <el-input v-model="marketForm.theoreticalPrice" />
        </el-form-item>
        <el-form-item label="内在价值">
          <el-input v-model="marketForm.intrinsicValue" />
        </el-form-item>
        <el-form-item label="时间价值">
          <el-input v-model="marketForm.timeValue" />
        </el-form-item>
        <el-form-item label="IV">
          <el-input v-model="marketForm.iv" />
        </el-form-item>
        <el-form-item label="Delta">
          <el-input v-model="marketForm.delta" />
        </el-form-item>
        <el-form-item label="Gamma">
          <el-input v-model="marketForm.gamma" />
        </el-form-item>
        <el-form-item label="Theta">
          <el-input v-model="marketForm.theta" />
        </el-form-item>
        <el-form-item label="Vega">
          <el-input v-model="marketForm.vega" />
        </el-form-item>
        <el-form-item label="Rho">
          <el-input v-model="marketForm.rho" />
        </el-form-item>
        <el-form-item label="无风险利率">
          <el-input v-model="marketForm.riskFreeRate" />
        </el-form-item>
        <el-form-item label="定价模型">
          <el-input v-model="marketForm.pricingModel" />
        </el-form-item>
        <el-form-item label="快照时间">
          <el-input-number v-model="marketForm.snapshotTime" :min="0" :precision="0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="marketVisible = false">
          取消
        </el-button>
        <el-button type="primary" :loading="submitLoading" @click="submitMarket">
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
import { optionService } from '@/services'

const props = defineProps<{ initialTab?: string }>()
const activeTab = ref(props.initialTab || 'contracts')
const loading = ref(false)
const submitLoading = ref(false)
const rows = ref<Record<string, any>[]>([])
const detailVisible = ref(false)
const detailData = ref<Record<string, any>>({})
const contractVisible = ref(false)
const marketVisible = ref(false)

const queries = reactive({
  contracts: { tenantId: undefined, contractCode: '', underlyingSymbol: '', status: undefined, limit: 100 },
  markets: { tenantId: undefined, contractId: undefined, limit: 100 },
  orders: { tenantId: undefined, uid: undefined, orderNo: '', contractId: undefined, limit: 100 },
  trades: { tenantId: undefined, contractId: undefined, tradeNo: '', limit: 100 },
  positions: { tenantId: undefined, uid: undefined, contractId: undefined, limit: 100 },
  exercises: { tenantId: undefined, uid: undefined, exerciseNo: '', contractId: undefined, limit: 100 },
  settlements: { tenantId: undefined, contractId: undefined, settlementNo: '', status: undefined, limit: 100 },
  accounts: { tenantId: undefined, uid: undefined, accountId: undefined, marginCoin: '', limit: 100 },
  bills: { tenantId: undefined, uid: undefined, accountId: undefined, bizNo: '', limit: 100 },
})

const fieldMap: Record<string, Array<{ key: string; label: string; type?: string }>> = {
  contracts: [
    { key: 'tenantId', label: '租户ID', type: 'number' },
    { key: 'contractCode', label: '合约编码' },
    { key: 'underlyingSymbol', label: '标的' },
  ],
  markets: [
    { key: 'tenantId', label: '租户ID', type: 'number' },
    { key: 'contractId', label: '合约ID', type: 'number' },
  ],
  orders: [
    { key: 'tenantId', label: '租户ID', type: 'number' },
    { key: 'uid', label: '用户ID', type: 'number' },
    { key: 'orderNo', label: '订单号' },
  ],
  trades: [
    { key: 'tenantId', label: '租户ID', type: 'number' },
    { key: 'contractId', label: '合约ID', type: 'number' },
    { key: 'tradeNo', label: '成交号' },
  ],
  positions: [
    { key: 'tenantId', label: '租户ID', type: 'number' },
    { key: 'uid', label: '用户ID', type: 'number' },
    { key: 'contractId', label: '合约ID', type: 'number' },
  ],
  exercises: [
    { key: 'tenantId', label: '租户ID', type: 'number' },
    { key: 'uid', label: '用户ID', type: 'number' },
    { key: 'exerciseNo', label: '行权号' },
  ],
  settlements: [
    { key: 'tenantId', label: '租户ID', type: 'number' },
    { key: 'contractId', label: '合约ID', type: 'number' },
    { key: 'settlementNo', label: '结算号' },
  ],
  accounts: [
    { key: 'tenantId', label: '租户ID', type: 'number' },
    { key: 'uid', label: '用户ID', type: 'number' },
    { key: 'accountId', label: '账户ID', type: 'number' },
  ],
  bills: [
    { key: 'tenantId', label: '租户ID', type: 'number' },
    { key: 'uid', label: '用户ID', type: 'number' },
    { key: 'bizNo', label: '业务单号' },
  ],
}

const columnMap: Record<string, Array<{ prop: string; label: string; width?: number }>> = {
  contracts: [
    { prop: 'id', label: 'ID', width: 80 },
    { prop: 'contractCode', label: '合约编码', width: 180 },
    { prop: 'underlyingSymbol', label: '标的', width: 120 },
    { prop: 'settleCoin', label: '结算币', width: 120 },
    { prop: 'strikePrice', label: '行权价' },
    { prop: 'status', label: '状态', width: 100 },
  ],
  markets: [
    { prop: 'contractId', label: '合约ID', width: 100 },
    { prop: 'underlyingPrice', label: '标的价格' },
    { prop: 'markPrice', label: '标记价格' },
    { prop: 'lastPrice', label: '最新价' },
    { prop: 'snapshotTime', label: '快照时间', width: 160 },
  ],
  orders: [
    { prop: 'orderNo', label: '订单号', width: 180 },
    { prop: 'uid', label: '用户ID', width: 100 },
    { prop: 'contractId', label: '合约ID', width: 100 },
    { prop: 'price', label: '价格' },
    { prop: 'qty', label: '数量' },
    { prop: 'status', label: '状态', width: 100 },
  ],
  trades: [
    { prop: 'tradeNo', label: '成交号', width: 180 },
    { prop: 'contractId', label: '合约ID', width: 100 },
    { prop: 'price', label: '价格' },
    { prop: 'qty', label: '数量' },
    { prop: 'tradeTime', label: '成交时间', width: 160 },
  ],
  positions: [
    { prop: 'uid', label: '用户ID', width: 100 },
    { prop: 'contractId', label: '合约ID', width: 100 },
    { prop: 'side', label: '方向', width: 100 },
    { prop: 'positionQty', label: '持仓量' },
    { prop: 'unrealizedPnl', label: '未实现盈亏' },
  ],
  exercises: [
    { prop: 'exerciseNo', label: '行权号', width: 180 },
    { prop: 'uid', label: '用户ID', width: 100 },
    { prop: 'exerciseQty', label: '行权数量' },
    { prop: 'profitAmount', label: '收益' },
    { prop: 'status', label: '状态', width: 100 },
  ],
  settlements: [
    { prop: 'settlementNo', label: '结算号', width: 180 },
    { prop: 'contractId', label: '合约ID', width: 100 },
    { prop: 'deliveryPrice', label: '交割价' },
    { prop: 'status', label: '状态', width: 100 },
  ],
  accounts: [
    { prop: 'uid', label: '用户ID', width: 100 },
    { prop: 'accountId', label: '账户ID', width: 120 },
    { prop: 'marginCoin', label: '保证金币种', width: 120 },
    { prop: 'balance', label: '余额' },
    { prop: 'availableBalance', label: '可用余额' },
  ],
  bills: [
    { prop: 'bizNo', label: '业务单号', width: 180 },
    { prop: 'uid', label: '用户ID', width: 100 },
    { prop: 'coin', label: '币种', width: 100 },
    { prop: 'changeAmount', label: '变动金额' },
    { prop: 'createTimes', label: '创建时间', width: 160 },
  ],
}

const contractForm = reactive<Record<string, any>>({
  id: 0,
  tenantId: 0,
  contractCode: '',
  underlyingSymbol: '',
  settleCoin: '',
  quoteCoin: '',
  optionType: 1,
  exerciseStyle: 1,
  settlementType: 1,
  strikePrice: '',
  contractUnit: '1',
  minOrderQty: '1',
  maxOrderQty: '1000',
  priceTick: '0.01',
  qtyStep: '1',
  multiplier: '1',
  listTime: 0,
  expireTime: 0,
  deliverTime: 0,
  isAutoExercise: 0,
  status: 1,
  sort: 0,
  remark: '',
  isDeleted: 0,
})

const marketForm = reactive<Record<string, any>>({
  tenantId: 0,
  contractId: 0,
  underlyingPrice: '',
  markPrice: '',
  lastPrice: '',
  bidPrice: '',
  askPrice: '',
  theoreticalPrice: '',
  intrinsicValue: '',
  timeValue: '',
  iv: '',
  delta: '',
  gamma: '',
  theta: '',
  vega: '',
  rho: '',
  riskFreeRate: '',
  pricingModel: '',
  snapshotTime: 0,
})

const currentQuery = computed(() => queries[activeTab.value as keyof typeof queries])
const currentFields = computed(() => fieldMap[activeTab.value] || [])
const currentColumns = computed(() => columnMap[activeTab.value] || [])
const pickList = (res: any) => res?.data || res?.list || []

const loadCurrent = async () => {
  loading.value = true
  try {
    if (activeTab.value === 'contracts') rows.value = pickList(await optionService.listContracts(currentQuery.value))
    if (activeTab.value === 'markets') rows.value = pickList(await optionService.listMarketSnapshots(currentQuery.value))
    if (activeTab.value === 'orders') rows.value = pickList(await optionService.listOrders(currentQuery.value))
    if (activeTab.value === 'trades') rows.value = pickList(await optionService.listTrades(currentQuery.value))
    if (activeTab.value === 'positions') rows.value = pickList(await optionService.listPositions(currentQuery.value))
    if (activeTab.value === 'exercises') rows.value = pickList(await optionService.listExercises(currentQuery.value))
    if (activeTab.value === 'settlements') rows.value = pickList(await optionService.listSettlements(currentQuery.value))
    if (activeTab.value === 'accounts') rows.value = pickList(await optionService.listAccounts(currentQuery.value))
    if (activeTab.value === 'bills') rows.value = pickList(await optionService.listBills(currentQuery.value))
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
  if (activeTab.value === 'contracts') detailData.value = (await optionService.getContract({ tenantId: row.tenantId, id: row.id })).data || row
  if (activeTab.value === 'markets') detailData.value = (await optionService.getMarket({ tenantId: row.tenantId, contractId: row.contractId })).data || row
  if (activeTab.value === 'orders') detailData.value = (await optionService.getOrder({ tenantId: row.tenantId, id: row.id, orderNo: row.orderNo })).data || row
  if (activeTab.value === 'trades') detailData.value = (await optionService.getTrade({ tenantId: row.tenantId, id: row.id, tradeNo: row.tradeNo })).data || row
  if (activeTab.value === 'positions') detailData.value = (await optionService.getPosition({ tenantId: row.tenantId, id: row.id })).data || row
  if (activeTab.value === 'exercises') detailData.value = (await optionService.getExercise({ tenantId: row.tenantId, id: row.id, exerciseNo: row.exerciseNo })).data || row
  if (activeTab.value === 'settlements') detailData.value = (await optionService.getSettlement({ tenantId: row.tenantId, id: row.id, settlementNo: row.settlementNo })).data || row
  if (activeTab.value === 'accounts') detailData.value = (await optionService.getAccount({ tenantId: row.tenantId, id: row.id, accountId: row.accountId })).data || row
  if (activeTab.value === 'bills') detailData.value = (await optionService.getBill({ tenantId: row.tenantId, id: row.id, bizNo: row.bizNo })).data || row
  detailVisible.value = true
}

const openContractDialog = (row?: Record<string, any>) => {
  Object.assign(contractForm, {
    id: 0,
    tenantId: 0,
    contractCode: '',
    underlyingSymbol: '',
    settleCoin: '',
    quoteCoin: '',
    optionType: 1,
    exerciseStyle: 1,
    settlementType: 1,
    strikePrice: '',
    contractUnit: '1',
    minOrderQty: '1',
    maxOrderQty: '1000',
    priceTick: '0.01',
    qtyStep: '1',
    multiplier: '1',
    listTime: 0,
    expireTime: 0,
    deliverTime: 0,
    isAutoExercise: 0,
    status: 1,
    sort: 0,
    remark: '',
    isDeleted: 0,
  }, row || {})
  contractVisible.value = true
}

const submitContract = async () => {
  submitLoading.value = true
  try {
    if (contractForm.id) await optionService.updateContract(contractForm)
    else await optionService.createContract(contractForm)
    ElMessage.success('保存成功')
    contractVisible.value = false
    loadCurrent()
  } finally {
    submitLoading.value = false
  }
}

const openMarketDialog = (row?: Record<string, any>) => {
  Object.assign(marketForm, {
    tenantId: row?.tenantId || 0,
    contractId: row?.contractId || row?.id || 0,
    underlyingPrice: '',
    markPrice: '',
    lastPrice: '',
    bidPrice: '',
    askPrice: '',
    theoreticalPrice: '',
    intrinsicValue: '',
    timeValue: '',
    iv: '',
    delta: '',
    gamma: '',
    theta: '',
    vega: '',
    rho: '',
    riskFreeRate: '',
    pricingModel: '',
    snapshotTime: 0,
  })
  marketVisible.value = true
}

const submitMarket = async () => {
  submitLoading.value = true
  try {
    await optionService.updateMarket(marketForm)
    ElMessage.success('更新成功')
    marketVisible.value = false
    if (activeTab.value !== 'markets') activeTab.value = 'markets'
    loadCurrent()
  } finally {
    submitLoading.value = false
  }
}

onMounted(loadCurrent)
</script>

<style scoped></style>
