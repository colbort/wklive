<template>
  <div class="module-page">
    <div class="page-header">
      <h2>{{ pageTitle }}</h2>
      <div class="header-actions">
        <el-button @click="loadCurrent">
          刷新
        </el-button>
        <el-button v-if="activeTab === 'user-assets'" type="primary" @click="openChangeDialog('add')">
          人工加币
        </el-button>
        <el-button v-if="activeTab === 'user-assets'" type="warning" @click="openChangeDialog('sub')">
          人工扣币
        </el-button>
        <el-button
          v-if="activeTab === 'user-assets'"
          type="primary"
          @click="openChangeDialog('freeze')"
        >
          冻结资产
        </el-button>
        <el-button
          v-if="activeTab === 'user-assets'"
          type="primary"
          plain
          @click="openChangeDialog('lock')"
        >
          锁仓资产
        </el-button>
        <el-button
          v-if="activeTab === 'freezes'"
          type="success"
          plain
          @click="openChangeDialog('unfreeze')"
        >
          解冻资产
        </el-button>
        <el-button
          v-if="activeTab === 'locks'"
          type="success"
          plain
          @click="openChangeDialog('unlock')"
        >
          解锁资产
        </el-button>
      </div>
    </div>

    <el-tabs v-if="showTabs" v-model="activeTab" @tab-change="loadCurrent">
      <el-tab-pane label="用户资产" name="user-assets" />
      <el-tab-pane label="资产流水" name="flows" />
      <el-tab-pane label="资产冻结明细" name="freezes" />
      <el-tab-pane label="资产锁仓明细" name="locks" />
    </el-tabs>

    <el-card shadow="never" class="query-card">
      <el-form :model="currentQuery" inline label-width="88px">
        <el-form-item v-for="field in currentFields" :key="field.key" :label="field.label">
          <el-select
            v-if="field.type === 'select'"
            v-model="currentQuery[field.key]"
            clearable
            style="width: 160px"
          >
            <el-option
              v-for="item in getFieldOptions(field.optionsKey)"
              :key="item.value"
              :label="getOptionLabel(t, item.code, item.value)"
              :value="item.value"
            />
          </el-select>
          <el-input v-else-if="field.type !== 'number'" v-model="currentQuery[field.key]" clearable />
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
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="showDetail(row)">
              详情
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="changeVisible" :title="changeTitle" width="680px">
      <el-form label-width="100px">
        <el-form-item label="租户ID">
          <el-input-number v-model="changeForm.tenantId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item v-if="!['unfreeze', 'unlock'].includes(changeMode)" label="用户ID">
          <el-input-number v-model="changeForm.userId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item v-if="!['unfreeze', 'unlock'].includes(changeMode)" label="钱包类型">
          <el-select v-model="changeForm.walletType" style="width: 100%">
            <el-option
              v-for="item in walletTypeOptions"
              :key="item.value"
              :label="getOptionLabel(t, item.code, item.value)"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item v-if="!['unfreeze', 'unlock'].includes(changeMode)" label="币种">
          <el-input v-model="changeForm.coin" />
        </el-form-item>
        <el-form-item v-if="changeMode === 'unfreeze'" label="冻结单号">
          <el-input v-model="changeForm.freezeNo" />
        </el-form-item>
        <el-form-item v-if="changeMode === 'unlock'" label="锁仓单号">
          <el-input v-model="changeForm.lockNo" />
        </el-form-item>
        <el-form-item label="数量">
          <el-input v-model="changeForm.amount" />
        </el-form-item>
        <el-form-item label="业务单号">
          <el-input v-model="changeForm.bizNo" />
        </el-form-item>
        <el-form-item label="操作人ID">
          <el-input-number v-model="changeForm.operatorId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="changeForm.remark" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="changeVisible = false">
          取消
        </el-button>
        <el-button type="primary" :loading="submitLoading" @click="submitChange">
          确定
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="detailVisible" :title="detailTitle" width="760px">
      <pre class="detail-pre">{{ JSON.stringify(detailData, null, 2) }}</pre>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { assetService, type OptionGroup } from '@/services'
import { findOptionGroup, getOptionLabel } from '@/utils/options'

const { t } = useI18n()

const props = withDefaults(defineProps<{
  mode?: 'user-assets' | 'flows' | 'freezes' | 'locks'
  title?: string
  showTabs?: boolean
}>(), {
  mode: 'user-assets',
  title: '',
  showTabs: false,
})

const titleMap = {
  'user-assets': '用户资产',
  flows: '资产流水',
  freezes: '资产冻结明细',
  locks: '资产锁仓明细',
}

const activeTab = ref<'user-assets' | 'flows' | 'freezes' | 'locks'>(props.mode)
const loading = ref(false)
const submitLoading = ref(false)
const rows = ref<Record<string, any>[]>([])
const optionGroups = ref<OptionGroup[]>([])
const detailVisible = ref(false)
const detailData = ref<Record<string, any>>({})
const detailTitle = computed(() => `${titleMap[activeTab.value]}详情`)
const changeVisible = ref(false)
const changeMode = ref<'add' | 'sub' | 'freeze' | 'unfreeze' | 'lock' | 'unlock'>('add')

const queries = reactive({
  'user-assets': { tenantId: undefined, userId: undefined, walletType: undefined, coin: '', status: undefined, limit: 100 },
  flows: { tenantId: undefined, userId: undefined, walletType: undefined, coin: '', bizNo: '', limit: 100 },
  freezes: { tenantId: undefined, userId: undefined, walletType: undefined, coin: '', bizNo: '', status: undefined, limit: 100 },
  locks: { tenantId: undefined, userId: undefined, walletType: undefined, coin: '', bizNo: '', status: undefined, limit: 100 },
})

const changeForm = reactive<Record<string, any>>({
  tenantId: 0,
  userId: 0,
  walletType: 1,
  coin: '',
  amount: '',
  bizNo: '',
  operatorId: 0,
  remark: '',
  freezeNo: '',
  lockNo: '',
})

const walletTypeOptions = computed(() => findOptionGroup(optionGroups.value, 'walletType'))
const assetStatusOptions = computed(() => findOptionGroup(optionGroups.value, 'assetStatus'))
const freezeStatusOptions = computed(() => findOptionGroup(optionGroups.value, 'freezeStatus'))
const lockStatusOptions = computed(() => findOptionGroup(optionGroups.value, 'lockStatus'))

const fieldMap: Record<string, Array<{ key: string; label: string; type?: string; optionsKey?: string }>> = {
  'user-assets': [
    { key: 'tenantId', label: '租户ID', type: 'number' },
    { key: 'userId', label: '用户ID', type: 'number' },
    { key: 'walletType', label: '钱包类型', type: 'select', optionsKey: 'walletType' },
    { key: 'coin', label: '币种' },
    { key: 'status', label: '状态', type: 'select', optionsKey: 'assetStatus' },
  ],
  flows: [
    { key: 'tenantId', label: '租户ID', type: 'number' },
    { key: 'userId', label: '用户ID', type: 'number' },
    { key: 'walletType', label: '钱包类型', type: 'select', optionsKey: 'walletType' },
    { key: 'coin', label: '币种' },
    { key: 'bizNo', label: '业务单号' },
  ],
  freezes: [
    { key: 'tenantId', label: '租户ID', type: 'number' },
    { key: 'userId', label: '用户ID', type: 'number' },
    { key: 'walletType', label: '钱包类型', type: 'select', optionsKey: 'walletType' },
    { key: 'coin', label: '币种' },
    { key: 'bizNo', label: '业务单号' },
    { key: 'status', label: '状态', type: 'select', optionsKey: 'freezeStatus' },
  ],
  locks: [
    { key: 'tenantId', label: '租户ID', type: 'number' },
    { key: 'userId', label: '用户ID', type: 'number' },
    { key: 'walletType', label: '钱包类型', type: 'select', optionsKey: 'walletType' },
    { key: 'coin', label: '币种' },
    { key: 'bizNo', label: '业务单号' },
    { key: 'status', label: '状态', type: 'select', optionsKey: 'lockStatus' },
  ],
}

const columnMap: Record<string, Array<{ prop: string; label: string; width?: number }>> = {
  'user-assets': [
    { prop: 'tenantId', label: '租户ID', width: 100 },
    { prop: 'userId', label: '用户ID', width: 100 },
    { prop: 'walletType', label: '钱包类型', width: 120 },
    { prop: 'coin', label: '币种', width: 100 },
    { prop: 'totalAmount', label: '总额' },
    { prop: 'availableAmount', label: '可用' },
    { prop: 'frozenAmount', label: '冻结' },
    { prop: 'lockedAmount', label: '锁仓' },
  ],
  flows: [
    { prop: 'flowNo', label: '流水号', width: 180 },
    { prop: 'tenantId', label: '租户ID', width: 100 },
    { prop: 'userId', label: '用户ID', width: 100 },
    { prop: 'coin', label: '币种', width: 100 },
    { prop: 'changeAmount', label: '变动数量' },
    { prop: 'bizNo', label: '业务单号', width: 180 },
  ],
  freezes: [
    { prop: 'freezeNo', label: '冻结单号', width: 180 },
    { prop: 'tenantId', label: '租户ID', width: 100 },
    { prop: 'userId', label: '用户ID', width: 100 },
    { prop: 'coin', label: '币种', width: 100 },
    { prop: 'amount', label: '冻结数量' },
    { prop: 'usedAmount', label: '已使用数量' },
    { prop: 'unfreezeAmount', label: '已解冻数量' },
    { prop: 'remainAmount', label: '剩余数量' },
    { prop: 'status', label: '状态', width: 100 },
  ],
  locks: [
    { prop: 'lockNo', label: '锁仓单号', width: 180 },
    { prop: 'tenantId', label: '租户ID', width: 100 },
    { prop: 'userId', label: '用户ID', width: 100 },
    { prop: 'coin', label: '币种', width: 100 },
    { prop: 'amount', label: '锁仓数量' },
    { prop: 'unlockAmount', label: '已解锁数量' },
    { prop: 'remainAmount', label: '剩余数量' },
    { prop: 'status', label: '状态', width: 100 },
  ],
}

const currentQuery = computed(() => queries[activeTab.value])
const currentFields = computed(() => fieldMap[activeTab.value] || [])
const currentColumns = computed(() => columnMap[activeTab.value] || [])
const pageTitle = computed(() => props.title || (props.showTabs ? '资产管理' : titleMap[activeTab.value]))
const changeTitle = computed(() => ({
  add: '人工加币',
  sub: '人工扣币',
  freeze: '冻结资产',
  unfreeze: '解冻资产',
  lock: '锁仓资产',
  unlock: '解锁资产',
}[changeMode.value]))

const pickList = (res: any) => res?.data || res?.list || []
const getFieldOptions = (key?: string) => {
  if (key === 'walletType') return walletTypeOptions.value
  if (key === 'assetStatus') return assetStatusOptions.value
  if (key === 'freezeStatus') return freezeStatusOptions.value
  if (key === 'lockStatus') return lockStatusOptions.value
  return []
}

const loadCurrent = async () => {
  loading.value = true
  try {
    if (activeTab.value === 'user-assets') rows.value = pickList(await assetService.getUserAssets(currentQuery.value))
    if (activeTab.value === 'flows') rows.value = pickList(await assetService.getFlows(currentQuery.value))
    if (activeTab.value === 'freezes') rows.value = pickList(await assetService.getFreezes(currentQuery.value))
    if (activeTab.value === 'locks') rows.value = pickList(await assetService.getLocks(currentQuery.value))
  } finally {
    loading.value = false
  }
}

const loadOptions = async () => {
  const res = await assetService.getOptions()
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
  if (activeTab.value === 'user-assets') {
    const res = await assetService.getUserAssetDetail({
      tenantId: Number(row.tenantId),
      userId: Number(row.userId),
      walletType: Number(row.walletType),
      coin: String(row.coin || ''),
    })
    detailData.value = res.data || row
  } else {
    detailData.value = row
  }
  detailVisible.value = true
}

const openChangeDialog = (mode: typeof changeMode.value) => {
  changeMode.value = mode
  Object.assign(changeForm, {
    tenantId: 0,
    userId: 0,
    walletType: 1,
    coin: '',
    amount: '',
    bizNo: '',
    operatorId: 0,
    remark: '',
    freezeNo: '',
    lockNo: '',
  })
  changeVisible.value = true
}

const submitChange = async () => {
  submitLoading.value = true
  try {
    if (changeMode.value === 'add') await assetService.addAsset(changeForm)
    if (changeMode.value === 'sub') await assetService.subAsset(changeForm)
    if (changeMode.value === 'freeze') await assetService.freezeAsset(changeForm)
    if (changeMode.value === 'unfreeze') await assetService.unfreezeAsset(changeForm)
    if (changeMode.value === 'lock') await assetService.lockAsset(changeForm)
    if (changeMode.value === 'unlock') await assetService.unlockAsset(changeForm)
    ElMessage.success('操作成功')
    changeVisible.value = false
    loadCurrent()
  } finally {
    submitLoading.value = false
  }
}

watch(
  () => props.mode,
  (mode) => {
    activeTab.value = mode
    loadCurrent()
  },
)

onMounted(loadCurrent)
onMounted(loadOptions)
</script>

<style scoped>
.query-card :deep(.el-form-item) {
  margin-bottom: 12px;
}
</style>
