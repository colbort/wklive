<template>
  <div class="module-page">
    <div class="page-header">
      <h2>{{ pageTitle }}</h2>
      <div class="header-actions">
        <el-button @click="loadCurrent"> {{ t('common.refresh') }} </el-button>
        <el-button
          v-if="activeTab === 'user-assets'"
          type="primary"
          @click="openChangeDialog('add')"
        >
          {{ t('asset.addAsset') }}
        </el-button>
        <el-button
          v-if="activeTab === 'user-assets'"
          type="warning"
          @click="openChangeDialog('sub')"
        >
          {{ t('asset.subAsset') }}
        </el-button>
        <el-button
          v-if="activeTab === 'user-assets'"
          type="primary"
          @click="openChangeDialog('freeze')"
        >
          {{ t('asset.freezeAsset') }}
        </el-button>
        <el-button
          v-if="activeTab === 'user-assets'"
          type="primary"
          plain
          @click="openChangeDialog('lock')"
        >
          {{ t('asset.lockAsset') }}
        </el-button>
        <el-button
          v-if="activeTab === 'freezes'"
          type="success"
          plain
          @click="openChangeDialog('unfreeze')"
        >
          {{ t('asset.unfreezeAsset') }}
        </el-button>
        <el-button
          v-if="activeTab === 'locks'"
          type="success"
          plain
          @click="openChangeDialog('unlock')"
        >
          {{ t('asset.unlockAsset') }}
        </el-button>
      </div>
    </div>

    <el-tabs v-if="showTabs" v-model="activeTab" @tab-change="loadCurrent">
      <el-tab-pane :label="t('asset.userAssets')" name="user-assets" />
      <el-tab-pane :label="t('asset.flows')" name="flows" />
      <el-tab-pane :label="t('asset.freezes')" name="freezes" />
      <el-tab-pane :label="t('asset.locks')" name="locks" />
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
          <el-input
            v-else-if="field.type !== 'number'"
            v-model="currentQuery[field.key]"
            clearable
          />
          <el-input-number v-else v-model="currentQuery[field.key]" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadCurrent"> {{ t('common.search') }} </el-button>
          <el-button @click="resetCurrent"> {{ t('common.reset') }} </el-button>
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
        <el-table-column :label="t('common.actions')" width="120" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="showDetail(row)"> {{ t('asset.detail') }} </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="changeVisible" :title="changeTitle" width="680px">
      <el-form label-width="100px">
        <el-form-item :label="t('asset.tenantId')">
          <el-input-number v-model="changeForm.tenantId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item v-if="!['unfreeze', 'unlock'].includes(changeMode)" :label="t('asset.userId')">
          <el-input-number v-model="changeForm.userId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item v-if="!['unfreeze', 'unlock'].includes(changeMode)" :label="t('asset.walletType')">
          <el-select v-model="changeForm.walletType" style="width: 100%">
            <el-option
              v-for="item in walletTypeOptions"
              :key="item.value"
              :label="getOptionLabel(t, item.code, item.value)"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item v-if="!['unfreeze', 'unlock'].includes(changeMode)" :label="t('asset.coin')">
          <el-input v-model="changeForm.coin" />
        </el-form-item>
        <el-form-item v-if="changeMode === 'unfreeze'" :label="t('asset.freezeNo')">
          <el-input v-model="changeForm.freezeNo" />
        </el-form-item>
        <el-form-item v-if="changeMode === 'unlock'" :label="t('asset.lockNo')">
          <el-input v-model="changeForm.lockNo" />
        </el-form-item>
        <el-form-item :label="t('asset.amount')">
          <el-input v-model="changeForm.amount" />
        </el-form-item>
        <el-form-item :label="t('asset.bizNo')">
          <el-input v-model="changeForm.bizNo" />
        </el-form-item>
        <el-form-item :label="t('asset.operatorId')">
          <el-input-number v-model="changeForm.operatorId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('common.remark')">
          <el-input v-model="changeForm.remark" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="changeVisible = false"> {{ t('common.cancel') }} </el-button>
        <el-button type="primary" :loading="submitLoading" @click="submitChange"> {{ t('common.confirm') }} </el-button>
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

const props = withDefaults(
  defineProps<{
    mode?: 'user-assets' | 'flows' | 'freezes' | 'locks'
    title?: string
    showTabs?: boolean
  }>(),
  {
    mode: 'user-assets',
    title: '',
    showTabs: false,
  },
)

const titleMap = {
  'user-assets': 'asset.userAssets',
  flows: 'asset.flows',
  freezes: 'asset.freezes',
  locks: 'asset.locks',
}

const activeTab = ref<'user-assets' | 'flows' | 'freezes' | 'locks'>(props.mode)
const loading = ref(false)
const submitLoading = ref(false)
const rows = ref<Record<string, any>[]>([])
const optionGroups = ref<OptionGroup[]>([])
const detailVisible = ref(false)
const detailData = ref<Record<string, any>>({})
const detailTitle = computed(() => `${t(titleMap[activeTab.value])}${t('asset.detail')}`)
const changeVisible = ref(false)
const changeMode = ref<'add' | 'sub' | 'freeze' | 'unfreeze' | 'lock' | 'unlock'>('add')

const queries = reactive({
  'user-assets': {
    tenantId: undefined,
    userId: undefined,
    walletType: undefined,
    coin: '',
    status: undefined,
    limit: 100,
  },
  flows: {
    tenantId: undefined,
    userId: undefined,
    walletType: undefined,
    coin: '',
    bizNo: '',
    limit: 100,
  },
  freezes: {
    tenantId: undefined,
    userId: undefined,
    walletType: undefined,
    coin: '',
    bizNo: '',
    status: undefined,
    limit: 100,
  },
  locks: {
    tenantId: undefined,
    userId: undefined,
    walletType: undefined,
    coin: '',
    bizNo: '',
    status: undefined,
    limit: 100,
  },
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

const fieldMap: Record<
  string,
  Array<{ key: string; label: string; type?: string; optionsKey?: string }>
> = {
  'user-assets': [
    { key: 'tenantId', label: t('asset.tenantId'), type: 'number' },
    { key: 'userId', label: t('asset.userId'), type: 'number' },
    { key: 'walletType', label: t('asset.walletType'), type: 'select', optionsKey: 'walletType' },
    { key: 'coin', label: t('asset.coin') },
    { key: 'status', label: t('common.status'), type: 'select', optionsKey: 'assetStatus' },
  ],
  flows: [
    { key: 'tenantId', label: t('asset.tenantId'), type: 'number' },
    { key: 'userId', label: t('asset.userId'), type: 'number' },
    { key: 'walletType', label: t('asset.walletType'), type: 'select', optionsKey: 'walletType' },
    { key: 'coin', label: t('asset.coin') },
    { key: 'bizNo', label: t('asset.bizNo') },
  ],
  freezes: [
    { key: 'tenantId', label: t('asset.tenantId'), type: 'number' },
    { key: 'userId', label: t('asset.userId'), type: 'number' },
    { key: 'walletType', label: t('asset.walletType'), type: 'select', optionsKey: 'walletType' },
    { key: 'coin', label: t('asset.coin') },
    { key: 'bizNo', label: t('asset.bizNo') },
    { key: 'status', label: t('common.status'), type: 'select', optionsKey: 'freezeStatus' },
  ],
  locks: [
    { key: 'tenantId', label: t('asset.tenantId'), type: 'number' },
    { key: 'userId', label: t('asset.userId'), type: 'number' },
    { key: 'walletType', label: t('asset.walletType'), type: 'select', optionsKey: 'walletType' },
    { key: 'coin', label: t('asset.coin') },
    { key: 'bizNo', label: t('asset.bizNo') },
    { key: 'status', label: t('common.status'), type: 'select', optionsKey: 'lockStatus' },
  ],
}

const columnMap: Record<string, Array<{ prop: string; label: string; width?: number }>> = {
  'user-assets': [
    { prop: 'tenantId', label: t('asset.tenantId'), width: 100 },
    { prop: 'userId', label: t('asset.userId'), width: 100 },
    { prop: 'walletType', label: t('asset.walletType'), width: 120 },
    { prop: 'coin', label: t('asset.coin'), width: 100 },
    { prop: 'totalAmount', label: t('asset.totalAmount') },
    { prop: 'availableAmount', label: t('asset.availableAmount') },
    { prop: 'frozenAmount', label: t('asset.frozenAmount') },
    { prop: 'lockedAmount', label: t('asset.lockedAmount') },
  ],
  flows: [
    { prop: 'flowNo', label: t('asset.flowNo'), width: 180 },
    { prop: 'tenantId', label: t('asset.tenantId'), width: 100 },
    { prop: 'userId', label: t('asset.userId'), width: 100 },
    { prop: 'coin', label: t('asset.coin'), width: 100 },
    { prop: 'changeAmount', label: t('asset.changeAmount') },
    { prop: 'bizNo', label: t('asset.bizNo'), width: 180 },
  ],
  freezes: [
    { prop: 'freezeNo', label: t('asset.freezeNo'), width: 180 },
    { prop: 'tenantId', label: t('asset.tenantId'), width: 100 },
    { prop: 'userId', label: t('asset.userId'), width: 100 },
    { prop: 'coin', label: t('asset.coin'), width: 100 },
    { prop: 'amount', label: t('asset.freezeAmount') },
    { prop: 'usedAmount', label: t('asset.usedAmount') },
    { prop: 'unfreezeAmount', label: t('asset.unfreezeAmount') },
    { prop: 'remainAmount', label: t('asset.remainAmount') },
    { prop: 'status', label: t('common.status'), width: 100 },
  ],
  locks: [
    { prop: 'lockNo', label: t('asset.lockNo'), width: 180 },
    { prop: 'tenantId', label: t('asset.tenantId'), width: 100 },
    { prop: 'userId', label: t('asset.userId'), width: 100 },
    { prop: 'coin', label: t('asset.coin'), width: 100 },
    { prop: 'amount', label: t('asset.lockAmount') },
    { prop: 'unlockAmount', label: t('asset.unlockAmount') },
    { prop: 'remainAmount', label: t('asset.remainAmount') },
    { prop: 'status', label: t('common.status'), width: 100 },
  ],
}

const currentQuery = computed(() => queries[activeTab.value])
const currentFields = computed(() => fieldMap[activeTab.value] || [])
const currentColumns = computed(() => columnMap[activeTab.value] || [])
const pageTitle = computed(
  () => props.title || (props.showTabs ? t('asset.management') : t(titleMap[activeTab.value])),
)
const changeTitle = computed(
  () =>
    ({
      add: t('asset.addAsset'),
      sub: t('asset.subAsset'),
      freeze: t('asset.freezeAsset'),
      unfreeze: t('asset.unfreezeAsset'),
      lock: t('asset.lockAsset'),
      unlock: t('asset.unlockAsset'),
    })[changeMode.value],
)

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
    if (activeTab.value === 'user-assets')
      rows.value = pickList(await assetService.getUserAssets(currentQuery.value))
    if (activeTab.value === 'flows')
      rows.value = pickList(await assetService.getFlows(currentQuery.value))
    if (activeTab.value === 'freezes')
      rows.value = pickList(await assetService.getFreezes(currentQuery.value))
    if (activeTab.value === 'locks')
      rows.value = pickList(await assetService.getLocks(currentQuery.value))
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
    ElMessage.success(t('common.success'))
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
