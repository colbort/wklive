<template>
  <div class="module-page">
    <div class="page-header">
      <h2>质押管理</h2>
      <div class="header-actions">
        <el-button @click="loadCurrent">
          刷新
        </el-button>
        <el-button v-if="activeTab === 'products'" type="primary" @click="openProductDialog()">
          新增产品
        </el-button>
      </div>
    </div>

    <el-tabs v-model="activeTab" @tab-change="loadCurrent">
      <el-tab-pane label="产品" name="products" />
      <el-tab-pane label="订单" name="orders" />
      <el-tab-pane label="奖励日志" name="reward-logs" />
      <el-tab-pane label="赎回日志" name="redeem-logs" />
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
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="showDetail(row)">
              详情
            </el-button>
            <el-button
              v-if="activeTab === 'products'"
              link
              type="primary"
              @click="openProductDialog(row)"
            >
              编辑
            </el-button>
            <el-button
              v-if="activeTab === 'products'"
              link
              type="warning"
              @click="changeStatus(row)"
            >
              状态
            </el-button>
            <el-button
              v-if="activeTab === 'orders'"
              link
              type="success"
              @click="openRewardDialog(row)"
            >
              手动发奖
            </el-button>
            <el-button
              v-if="activeTab === 'orders'"
              link
              type="danger"
              @click="openRedeemDialog(row)"
            >
              手动赎回
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="productVisible" :title="productForm.id ? '编辑产品' : '新增产品'" width="760px">
      <el-form label-width="110px">
        <el-form-item label="租户ID">
          <el-input-number v-model="productForm.tenantId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item label="产品编号">
          <el-input v-model="productForm.productNo" />
        </el-form-item>
        <el-form-item label="产品名称">
          <el-input v-model="productForm.productName" />
        </el-form-item>
        <el-form-item label="产品类型">
          <el-select v-model="productForm.productType" style="width: 100%">
            <el-option
              v-for="item in productTypeOptions"
              :key="item.value"
              :label="getOptionLabel(t, item.code, item.value)"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="币种名称">
          <el-input v-model="productForm.coinName" />
        </el-form-item>
        <el-form-item label="币种符号">
          <el-input v-model="productForm.coinSymbol" />
        </el-form-item>
        <el-form-item label="奖励币名称">
          <el-input v-model="productForm.rewardCoinName" />
        </el-form-item>
        <el-form-item label="奖励币符号">
          <el-input v-model="productForm.rewardCoinSymbol" />
        </el-form-item>
        <el-form-item label="APR">
          <el-input v-model="productForm.apr" />
        </el-form-item>
        <el-form-item label="锁仓天数">
          <el-input-number v-model="productForm.lockDays" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item label="最小数量">
          <el-input v-model="productForm.minAmount" />
        </el-form-item>
        <el-form-item label="最大数量">
          <el-input v-model="productForm.maxAmount" />
        </el-form-item>
        <el-form-item label="步长">
          <el-input v-model="productForm.stepAmount" />
        </el-form-item>
        <el-form-item label="总量">
          <el-input v-model="productForm.totalAmount" />
        </el-form-item>
        <el-form-item label="个人限额">
          <el-input v-model="productForm.userLimitAmount" />
        </el-form-item>
        <el-form-item label="计息模式">
          <el-select v-model="productForm.interestMode" style="width: 100%">
            <el-option
              v-for="item in interestModeOptions"
              :key="item.value"
              :label="getOptionLabel(t, item.code, item.value)"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="奖励模式">
          <el-select v-model="productForm.rewardMode" style="width: 100%">
            <el-option
              v-for="item in rewardModeOptions"
              :key="item.value"
              :label="getOptionLabel(t, item.code, item.value)"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="允许提前赎回">
          <el-select v-model="productForm.allowEarlyRedeem" style="width: 100%">
            <el-option
              v-for="item in yesNoOptions"
              :key="item.value"
              :label="getOptionLabel(t, item.code, item.value)"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="提前赎回费率">
          <el-input v-model="productForm.earlyRedeemRate" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="productForm.status" style="width: 100%">
            <el-option
              v-for="item in productStatusOptions"
              :key="item.value"
              :label="getOptionLabel(t, item.code, item.value)"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="productForm.sort" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item label="操作人UID">
          <el-input-number v-model="productForm.operatorUid" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="productForm.remark" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="productVisible = false">
          取消
        </el-button>
        <el-button type="primary" :loading="submitLoading" @click="submitProduct">
          确定
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="rewardVisible" title="手动发奖" width="640px">
      <el-form label-width="100px">
        <el-form-item label="租户ID">
          <el-input-number v-model="rewardForm.tenantId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item label="订单ID">
          <el-input-number v-model="rewardForm.orderId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item label="奖励数量">
          <el-input v-model="rewardForm.rewardAmount" />
        </el-form-item>
        <el-form-item label="奖励类型">
          <el-input-number v-model="rewardForm.rewardType" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item label="操作人UID">
          <el-input-number v-model="rewardForm.operatorUid" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="rewardForm.remark" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="rewardVisible = false">
          取消
        </el-button>
        <el-button type="primary" :loading="submitLoading" @click="submitReward">
          确定
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="redeemVisible" title="手动赎回" width="680px">
      <el-form label-width="100px">
        <el-form-item label="租户ID">
          <el-input-number v-model="redeemForm.tenantId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item label="订单ID">
          <el-input-number v-model="redeemForm.orderId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item label="赎回类型">
          <el-input-number v-model="redeemForm.redeemType" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item label="赎回数量">
          <el-input v-model="redeemForm.redeemAmount" />
        </el-form-item>
        <el-form-item label="奖励数量">
          <el-input v-model="redeemForm.rewardAmount" />
        </el-form-item>
        <el-form-item label="费率">
          <el-input v-model="redeemForm.feeRate" />
        </el-form-item>
        <el-form-item label="手续费">
          <el-input v-model="redeemForm.feeAmount" />
        </el-form-item>
        <el-form-item label="操作人UID">
          <el-input-number v-model="redeemForm.operatorUid" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="redeemForm.remark" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="redeemVisible = false">
          取消
        </el-button>
        <el-button type="primary" :loading="submitLoading" @click="submitRedeem">
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
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { stakingService, type OptionGroup } from '@/services'
import { findOptionGroup, getOptionLabel } from '@/utils/options'

const { t } = useI18n()

const props = defineProps<{ initialTab?: string }>()
const activeTab = ref(props.initialTab || 'products')
const optionGroups = ref<OptionGroup[]>([])
const productTypeOptions = computed(() => findOptionGroup(optionGroups.value, 'productType'))
const productStatusOptions = computed(() => findOptionGroup(optionGroups.value, 'productStatus'))
const interestModeOptions = computed(() => findOptionGroup(optionGroups.value, 'interestMode'))
const rewardModeOptions = computed(() => findOptionGroup(optionGroups.value, 'rewardMode'))
const yesNoOptions = computed(() => findOptionGroup(optionGroups.value, 'yesNo'))
const loading = ref(false)
const submitLoading = ref(false)
const rows = ref<Record<string, any>[]>([])
const detailVisible = ref(false)
const detailData = ref<Record<string, any>>({})
const productVisible = ref(false)
const rewardVisible = ref(false)
const redeemVisible = ref(false)

const queries = reactive({
  products: { tenantId: undefined, productNo: '', productName: '', coinSymbol: '', limit: 100 },
  orders: { tenantId: undefined, orderNo: '', uid: undefined, productId: undefined, limit: 100 },
  'reward-logs': { tenantId: undefined, orderNo: '', uid: undefined, productId: undefined, limit: 100 },
  'redeem-logs': { tenantId: undefined, orderNo: '', uid: undefined, redeemNo: '', limit: 100 },
})

const fieldMap: Record<string, Array<{ key: string; label: string; type?: string }>> = {
  products: [
    { key: 'tenantId', label: '租户ID', type: 'number' },
    { key: 'productNo', label: '产品编号' },
    { key: 'productName', label: '产品名称' },
  ],
  orders: [
    { key: 'tenantId', label: '租户ID', type: 'number' },
    { key: 'orderNo', label: '订单号' },
    { key: 'uid', label: '用户ID', type: 'number' },
  ],
  'reward-logs': [
    { key: 'tenantId', label: '租户ID', type: 'number' },
    { key: 'orderNo', label: '订单号' },
    { key: 'uid', label: '用户ID', type: 'number' },
  ],
  'redeem-logs': [
    { key: 'tenantId', label: '租户ID', type: 'number' },
    { key: 'orderNo', label: '订单号' },
    { key: 'redeemNo', label: '赎回单号' },
  ],
}

const columnMap: Record<string, Array<{ prop: string; label: string; width?: number }>> = {
  products: [
    { prop: 'productNo', label: '产品编号', width: 180 },
    { prop: 'productName', label: '产品名称', width: 180 },
    { prop: 'coinSymbol', label: '币种', width: 120 },
    { prop: 'apr', label: 'APR' },
    { prop: 'lockDays', label: '锁仓天数', width: 120 },
    { prop: 'status', label: '状态', width: 100 },
  ],
  orders: [
    { prop: 'orderNo', label: '订单号', width: 180 },
    { prop: 'uid', label: '用户ID', width: 100 },
    { prop: 'productName', label: '产品名称', width: 180 },
    { prop: 'stakeAmount', label: '质押数量' },
    { prop: 'totalReward', label: '累计奖励' },
    { prop: 'status', label: '状态', width: 100 },
  ],
  'reward-logs': [
    { prop: 'orderNo', label: '订单号', width: 180 },
    { prop: 'uid', label: '用户ID', width: 100 },
    { prop: 'rewardAmount', label: '奖励数量' },
    { prop: 'rewardType', label: '奖励类型', width: 100 },
    { prop: 'rewardStatus', label: '奖励状态', width: 100 },
  ],
  'redeem-logs': [
    { prop: 'redeemNo', label: '赎回单号', width: 180 },
    { prop: 'orderNo', label: '订单号', width: 180 },
    { prop: 'uid', label: '用户ID', width: 100 },
    { prop: 'redeemAmount', label: '赎回数量' },
    { prop: 'feeAmount', label: '手续费' },
    { prop: 'redeemStatus', label: '状态', width: 100 },
  ],
}

const productForm = reactive<Record<string, any>>({
  id: 0,
  tenantId: 0,
  productNo: '',
  productName: '',
  productType: 1,
  coinName: '',
  coinSymbol: '',
  rewardCoinName: '',
  rewardCoinSymbol: '',
  apr: '',
  lockDays: 0,
  minAmount: '',
  maxAmount: '',
  stepAmount: '',
  totalAmount: '',
  userLimitAmount: '',
  interestMode: 1,
  rewardMode: 1,
  allowEarlyRedeem: 0,
  earlyRedeemRate: '',
  status: 1,
  sort: 0,
  operatorUid: 0,
  remark: '',
})

const rewardForm = reactive<Record<string, any>>({ tenantId: 0, orderId: 0, rewardAmount: '', rewardType: 1, operatorUid: 0, remark: '' })
const redeemForm = reactive<Record<string, any>>({ tenantId: 0, orderId: 0, redeemType: 1, redeemAmount: '', rewardAmount: '', feeRate: '', feeAmount: '', operatorUid: 0, remark: '' })

const currentQuery = computed(() => queries[activeTab.value as keyof typeof queries])
const currentFields = computed(() => fieldMap[activeTab.value] || [])
const currentColumns = computed(() => columnMap[activeTab.value] || [])
const pickList = (res: any) => res?.data || res?.list || []

const loadCurrent = async () => {
  loading.value = true
  try {
    if (activeTab.value === 'products') rows.value = pickList(await stakingService.listProducts(currentQuery.value))
    if (activeTab.value === 'orders') rows.value = pickList(await stakingService.listOrders(currentQuery.value))
    if (activeTab.value === 'reward-logs') rows.value = pickList(await stakingService.listRewardLogs(currentQuery.value))
    if (activeTab.value === 'redeem-logs') rows.value = pickList(await stakingService.listRedeemLogs(currentQuery.value))
  } finally {
    loading.value = false
  }
}

const loadOptions = async () => {
  const res = await stakingService.getOptions()
  optionGroups.value = res.data || []
}

const resetCurrent = () => {
  Object.keys(currentQuery.value).forEach((key) => {
    currentQuery.value[key] = key === 'limit' ? 100 : ''
  })
  currentQuery.value.limit = 100
  loadCurrent()
}

const showDetail = async (row: Record<string, any>) => {
  if (activeTab.value === 'products') detailData.value = (await stakingService.getProduct({ tenantId: row.tenantId, id: row.id })).data || row
  else if (activeTab.value === 'orders') detailData.value = (await stakingService.getOrder({ tenantId: row.tenantId, id: row.id })).data || row
  else detailData.value = row
  detailVisible.value = true
}

const openProductDialog = (row?: Record<string, any>) => {
  Object.assign(productForm, {
    id: 0,
    tenantId: 0,
    productNo: '',
    productName: '',
    productType: 1,
    coinName: '',
    coinSymbol: '',
    rewardCoinName: '',
    rewardCoinSymbol: '',
    apr: '',
    lockDays: 0,
    minAmount: '',
    maxAmount: '',
    stepAmount: '',
    totalAmount: '',
    userLimitAmount: '',
    interestMode: 1,
    rewardMode: 1,
    allowEarlyRedeem: 0,
    earlyRedeemRate: '',
    status: 1,
    sort: 0,
    operatorUid: 0,
    remark: '',
  }, row || {})
  productVisible.value = true
}

const submitProduct = async () => {
  submitLoading.value = true
  try {
    if (productForm.id) await stakingService.updateProduct(productForm)
    else await stakingService.createProduct(productForm)
    ElMessage.success('保存成功')
    productVisible.value = false
    loadCurrent()
  } finally {
    submitLoading.value = false
  }
}

const changeStatus = async (row: Record<string, any>) => {
  const { value } = await ElMessageBox.prompt('请输入新的状态值', '修改状态', { inputValue: String(row.status || 0) })
  await stakingService.changeProductStatus({ tenantId: row.tenantId, id: row.id, status: Number(value), operatorUid: row.updateUserId || 0 })
  ElMessage.success('状态已更新')
  loadCurrent()
}

const openRewardDialog = (row: Record<string, any>) => {
  Object.assign(rewardForm, { tenantId: row.tenantId || 0, orderId: row.id || 0, rewardAmount: '', rewardType: 1, operatorUid: 0, remark: '' })
  rewardVisible.value = true
}

const submitReward = async () => {
  submitLoading.value = true
  try {
    await stakingService.manualReward(rewardForm)
    ElMessage.success('发奖成功')
    rewardVisible.value = false
    loadCurrent()
  } finally {
    submitLoading.value = false
  }
}

const openRedeemDialog = (row: Record<string, any>) => {
  Object.assign(redeemForm, { tenantId: row.tenantId || 0, orderId: row.id || 0, redeemType: 1, redeemAmount: '', rewardAmount: '', feeRate: '', feeAmount: '', operatorUid: 0, remark: '' })
  redeemVisible.value = true
}

const submitRedeem = async () => {
  submitLoading.value = true
  try {
    await stakingService.manualRedeem(redeemForm)
    ElMessage.success('赎回成功')
    redeemVisible.value = false
    loadCurrent()
  } finally {
    submitLoading.value = false
  }
}

onMounted(async () => {
  await Promise.all([loadCurrent(), loadOptions()])
})
</script>

<style scoped></style>
