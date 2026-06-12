<template>
  <div class="module-page">
    <div class="page-header">
      <h2>{{ t('asset.coinConfigs') }}</h2>
      <div class="header-actions">
        <el-button @click="loadList">
          {{ t('common.refresh') }}
        </el-button>
        <el-button
          v-perm="'asset:config:add'"
          class="page-create-action"
          type="primary"
          @click="openCreateDialog"
        >
          {{ t('asset.addCoinConfig') }}
        </el-button>
      </div>
    </div>

    <CrudQueryCard :model="query" label-width="104px" :show-actions="false">
      <el-form-item :label="t('asset.tenantId')">
        <TenantSelect v-model="query.tenantId" class="tenant-select-filter" />
      </el-form-item>
      <el-form-item :label="t('asset.walletType')">
        <el-select v-model="query.walletType" clearable style="width: 160px">
          <el-option
            v-for="item in walletTypeOptions"
            :key="item.value"
            :label="getOptionLabel(t, item.code, item.value)"
            :value="item.value"
          />
        </el-select>
      </el-form-item>
      <el-form-item :label="t('asset.coin')">
        <el-input v-model="query.coin" clearable />
      </el-form-item>
      <el-form-item :label="t('asset.symbol')">
        <el-input v-model="query.symbol" clearable />
      </el-form-item>
      <el-form-item :label="t('asset.coinType')">
        <el-select v-model="query.coinType" clearable style="width: 140px">
          <el-option
            v-for="item in coinTypeOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
      </el-form-item>
      <el-form-item :label="t('asset.chainCode')">
        <el-select v-model="query.chainCode" clearable style="width: 150px">
          <el-option
            v-for="item in chainCodeOptions"
            :key="item.value"
            :label="getOptionLabel(t, item.code, item.value)"
            :value="item.value"
          />
        </el-select>
      </el-form-item>
      <el-form-item :label="t('asset.appVisible')">
        <el-select v-model="query.appVisible" clearable style="width: 140px">
          <el-option :label="t('common.visible')" :value="1" />
          <el-option :label="t('common.hidden')" :value="2" />
        </el-select>
      </el-form-item>
      <el-form-item :label="t('asset.rechargeEnabled')">
        <el-select v-model="query.rechargeEnabled" clearable style="width: 140px">
          <el-option :label="t('common.enabled')" :value="1" />
          <el-option :label="t('common.disabled')" :value="2" />
        </el-select>
      </el-form-item>
      <el-form-item :label="t('asset.withdrawEnabled')">
        <el-select v-model="query.withdrawEnabled" clearable style="width: 140px">
          <el-option :label="t('common.enabled')" :value="1" />
          <el-option :label="t('common.disabled')" :value="2" />
        </el-select>
      </el-form-item>
      <el-form-item :label="t('asset.transferEnabled')">
        <el-select v-model="query.transferEnabled" clearable style="width: 140px">
          <el-option :label="t('common.enabled')" :value="1" />
          <el-option :label="t('common.disabled')" :value="2" />
        </el-select>
      </el-form-item>
      <el-form-item :label="t('common.enabled')">
        <el-select v-model="query.enabled" clearable style="width: 160px">
          <el-option
            v-for="item in assetStatusOptions"
            :key="item.value"
            :label="getOptionLabel(t, item.code, item.value)"
            :value="item.value"
          />
        </el-select>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="loadList">
          {{ t('common.search') }}
        </el-button>
        <el-button @click="resetQuery">
          {{ t('common.reset') }}
        </el-button>
      </el-form-item>
    </CrudQueryCard>

    <el-card shadow="never" class="table-card">
      <el-table
        v-loading="loading"
        :data="rows"
        stripe
        :empty-text="t('common.noData')"
      >
        <el-table-column prop="id" :label="t('common.id')" width="90" />
        <el-table-column prop="tenantId" :label="t('asset.tenantId')" min-width="100" />
        <el-table-column prop="walletType" :label="t('asset.walletType')" min-width="110">
          <template #default="{ row }">
            {{ formatOption(walletTypeOptions, row.walletType) }}
          </template>
        </el-table-column>
        <el-table-column
          prop="coin"
          :label="t('asset.coin')"
          min-width="100"
          show-overflow-tooltip
        />
        <el-table-column
          prop="symbol"
          :label="t('asset.symbol')"
          min-width="100"
          show-overflow-tooltip
        />
        <el-table-column
          prop="coinName"
          :label="t('asset.coinName')"
          min-width="130"
          show-overflow-tooltip
        />
        <el-table-column prop="coinType" :label="t('asset.coinType')" min-width="100">
          <template #default="{ row }">
            {{ formatCoinType(row.coinType) }}
          </template>
        </el-table-column>
        <el-table-column prop="chainCode" :label="t('asset.chainCode')" min-width="110">
          <template #default="{ row }">
            {{ formatOption(chainCodeOptions, row.chainCode) }}
          </template>
        </el-table-column>
        <el-table-column prop="decimalPlaces" :label="t('asset.decimalPlaces')" min-width="100" />
        <el-table-column prop="appVisible" :label="t('asset.appVisible')" min-width="100">
          <template #default="{ row }">
            <el-tag :type="row.appVisible === 1 ? 'success' : 'info'">
              {{ row.appVisible === 1 ? t('common.visible') : t('common.hidden') }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="rechargeEnabled" :label="t('asset.rechargeEnabled')" min-width="110">
          <template #default="{ row }">
            <el-tag :type="row.rechargeEnabled === 1 ? 'success' : 'info'">
              {{ formatEnabled(row.rechargeEnabled) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="withdrawEnabled" :label="t('asset.withdrawEnabled')" min-width="110">
          <template #default="{ row }">
            <el-tag :type="row.withdrawEnabled === 1 ? 'success' : 'info'">
              {{ formatEnabled(row.withdrawEnabled) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="transferEnabled" :label="t('asset.transferEnabled')" min-width="110">
          <template #default="{ row }">
            <el-tag :type="row.transferEnabled === 1 ? 'success' : 'info'">
              {{ formatEnabled(row.transferEnabled) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="enabled" :label="t('common.enabled')" min-width="100">
          <template #default="{ row }">
            {{ formatOption(assetStatusOptions, row.enabled) }}
          </template>
        </el-table-column>
        <el-table-column prop="sort" :label="t('common.sort')" min-width="90" />
        <el-table-column prop="createTimes" :label="t('common.createTimes')" min-width="160">
          <template #default="{ row }">
            {{ formatDate(row.createTimes) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('common.actions')" width="190" fixed="right">
          <template #default="{ row }">
            <el-button
              v-perm="'asset:config:detail'"
              link
              type="primary"
              @click="showDetail(row)"
            >
              {{ t('common.detail') }}
            </el-button>
            <el-button
              v-perm="'asset:config:update'"
              link
              type="primary"
              @click="openEditDialog(row)"
            >
              {{ t('common.edit') }}
            </el-button>
            <el-button
              v-perm="'asset:config:delete'"
              link
              type="danger"
              @click="deleteRow(row)"
            >
              {{ t('common.delete') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <CursorPagination
        v-model:limit="pagination.limit"
        :total="pagination.total"
        :has-prev="pagination.hasPrev"
        :has-next="pagination.hasNext"
        @prev="handlePrevPage"
        @next="handleNextPage"
        @limit-change="handleLimitChange"
      />
    </el-card>

    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="760px"
      :close-on-click-modal="false"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="118px"
      >
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item :label="t('asset.tenantId')" prop="tenantId">
              <TenantSelect v-model="form.tenantId" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('asset.walletType')" prop="walletType">
              <el-select v-model="form.walletType" style="width: 100%">
                <el-option
                  v-for="item in walletTypeOptions"
                  :key="item.value"
                  :label="getOptionLabel(t, item.code, item.value)"
                  :value="item.value"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('asset.coin')" prop="coin">
              <el-input v-model="form.coin" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('asset.symbol')" prop="symbol">
              <el-input v-model="form.symbol" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('asset.coinName')" prop="coinName">
              <el-input v-model="form.coinName" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('asset.coinType')" prop="coinType">
              <el-select v-model="form.coinType" style="width: 100%">
                <el-option
                  v-for="item in coinTypeOptions"
                  :key="item.value"
                  :label="item.label"
                  :value="item.value"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('asset.chainCode')" prop="chainCode">
              <el-select v-model="form.chainCode" style="width: 100%">
                <el-option
                  v-for="item in chainCodeOptions"
                  :key="item.value"
                  :label="getOptionLabel(t, item.code, item.value)"
                  :value="item.value"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('asset.decimalPlaces')">
              <el-input-number
                v-model="form.decimalPlaces"
                :min="0"
                :max="18"
                :precision="0"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('common.sort')">
              <el-input-number v-model="form.sort" :min="0" :precision="0" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('asset.iconUrl')">
              <el-input v-model="form.iconUrl" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('asset.iconText')">
              <el-input v-model="form.iconText" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('asset.iconBgColor')">
              <el-input v-model="form.iconBgColor" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('common.enabled')">
              <el-select v-model="form.enabled" style="width: 100%">
                <el-option
                  v-for="item in assetStatusOptions"
                  :key="item.value"
                  :label="getOptionLabel(t, item.code, item.value)"
                  :value="item.value"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('asset.appVisible')">
              <el-select v-model="form.appVisible" style="width: 100%">
                <el-option :label="t('common.visible')" :value="1" />
                <el-option :label="t('common.hidden')" :value="2" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('asset.rechargeEnabled')">
              <el-switch v-model="form.rechargeEnabled" :active-value="1" :inactive-value="2" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('asset.withdrawEnabled')">
              <el-switch v-model="form.withdrawEnabled" :active-value="1" :inactive-value="2" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('asset.transferEnabled')">
              <el-switch v-model="form.transferEnabled" :active-value="1" :inactive-value="2" />
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item :label="t('common.remark')">
              <el-input v-model="form.remark" type="textarea" :rows="3" />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">
          {{ t('common.cancel') }}
        </el-button>
        <el-button
          v-perm="isEdit ? 'asset:config:update' : 'asset:config:add'"
          type="primary"
          :loading="submitLoading"
          @click="submitForm"
        >
          {{ t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="detailVisible" :title="detailTitle" width="760px">
      <pre class="detail-pre">{{ JSON.stringify(detailData, null, 2) }}</pre>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { usePagination } from '@/composables'
import { assetService, type AssetCoinConfig, type OptionGroup, type OptionItem } from '@/services'
import { formatDate } from '@/utils'
import { findOptionGroup, getOptionLabel } from '@/utils/options'
import TenantSelect from '@/components/TenantSelect.vue'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'

const { t } = useI18n()
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)

const loading = ref(false)
const submitLoading = ref(false)
const rows = ref<AssetCoinConfig[]>([])
const optionGroups = ref<OptionGroup[]>([])
const dialogVisible = ref(false)
const detailVisible = ref(false)
const isEdit = ref(false)
const detailData = ref<AssetCoinConfig | null>(null)
const formRef = ref<FormInstance>()

const query = reactive({
  tenantId: undefined as number | undefined,
  walletType: undefined as number | undefined,
  coin: '',
  symbol: '',
  coinType: undefined as number | undefined,
  chainCode: undefined as number | undefined,
  appVisible: undefined as number | undefined,
  rechargeEnabled: undefined as number | undefined,
  withdrawEnabled: undefined as number | undefined,
  transferEnabled: undefined as number | undefined,
  enabled: undefined as number | undefined,
  limit: 20,
})

const form = reactive({
  id: 0,
  tenantId: 0,
  walletType: 1,
  coin: '',
  symbol: '',
  coinName: '',
  coinType: 2,
  chainCode: 0,
  iconUrl: '',
  iconText: '',
  iconBgColor: '',
  decimalPlaces: 8,
  appVisible: 1,
  rechargeEnabled: 2,
  withdrawEnabled: 2,
  transferEnabled: 1,
  enabled: 1,
  sort: 0,
  remark: '',
})

const requiredRule = { required: true, message: t('validation.required'), trigger: 'blur' as const }
const rules = reactive<FormRules>({
  tenantId: [requiredRule],
  walletType: [requiredRule],
  coin: [requiredRule],
  symbol: [requiredRule],
  coinName: [requiredRule],
  coinType: [requiredRule],
  chainCode: [requiredRule],
})

const walletTypeOptions = computed(() => findOptionGroup(optionGroups.value, 'walletType'))
const fallbackChainCodeOptions: OptionItem[] = [
  { value: 0, code: 'CHAIN_CODE_UNKNOWN' },
  { value: 1, code: 'CHAIN_CODE_BTC' },
  { value: 2, code: 'CHAIN_CODE_ETH' },
  { value: 3, code: 'CHAIN_CODE_TRX' },
  { value: 4, code: 'CHAIN_CODE_BSC' },
  { value: 5, code: 'CHAIN_CODE_SOL' },
  { value: 6, code: 'CHAIN_CODE_POLYGON' },
  { value: 20, code: 'CHAIN_CODE_TRC20' },
  { value: 21, code: 'CHAIN_CODE_ERC20' },
  { value: 22, code: 'CHAIN_CODE_BEP20' },
]
const chainCodeOptions = computed(() => {
  const options = findOptionGroup(optionGroups.value, 'chainCode')
  return options.length ? options : fallbackChainCodeOptions
})
const assetStatusOptions = computed(() => findOptionGroup(optionGroups.value, 'assetStatus'))
const coinTypeOptions = computed(() => [
  { label: t('asset.fiat'), value: 1 },
  { label: t('asset.crypto'), value: 2 },
])
const dialogTitle = computed(() =>
  isEdit.value ? t('asset.editCoinConfig') : t('asset.addCoinConfig'),
)
const detailTitle = computed(() => `${t('asset.coinConfigs')}${t('asset.detail')}`)

function formatEnabled(value: number) {
  return value === 1 ? t('common.enabled') : t('common.disabled')
}

function formatOption(options: OptionItem[], value: number) {
  const item = options.find((option) => option.value === value)
  return item ? getOptionLabel(t, item.code, item.value) : value
}

function formatCoinType(value: number) {
  return coinTypeOptions.value.find((item) => item.value === value)?.label || value
}

function resetFormData() {
  form.id = 0
  form.tenantId = 0
  form.walletType = 1
  form.coin = ''
  form.symbol = ''
  form.coinName = ''
  form.coinType = 2
  form.chainCode = 0
  form.iconUrl = ''
  form.iconText = ''
  form.iconBgColor = ''
  form.decimalPlaces = 8
  form.appVisible = 1
  form.rechargeEnabled = 2
  form.withdrawEnabled = 2
  form.transferEnabled = 1
  form.enabled = 1
  form.sort = 0
  form.remark = ''
  formRef.value?.clearValidate()
}

function fillForm(row: AssetCoinConfig) {
  form.id = Number(row.id)
  form.tenantId = Number(row.tenantId)
  form.walletType = Number(row.walletType)
  form.coin = row.coin || ''
  form.symbol = row.symbol || ''
  form.coinName = row.coinName || ''
  form.coinType = Number(row.coinType || 0)
  form.chainCode = Number(row.chainCode || 0)
  form.iconUrl = row.iconUrl || ''
  form.iconText = row.iconText || ''
  form.iconBgColor = row.iconBgColor || ''
  form.decimalPlaces = Number(row.decimalPlaces || 0)
  form.appVisible = Number(row.appVisible || 1)
  form.rechargeEnabled = Number(row.rechargeEnabled || 2)
  form.withdrawEnabled = Number(row.withdrawEnabled || 2)
  form.transferEnabled = Number(row.transferEnabled || 1)
  form.enabled = Number(row.enabled || 1)
  form.sort = Number(row.sort || 0)
  form.remark = row.remark || ''
}

async function loadOptions() {
  const res = await assetService.getOptions()
  optionGroups.value = res.data || []
}

async function loadList() {
  loading.value = true
  try {
    const res = await assetService.getCoinConfigs({
      ...query,
      cursor: pagination.cursor,
      limit: pagination.limit,
    })
    rows.value = res?.data || []
    updateFromResponse(res)
  } finally {
    loading.value = false
  }
}

function resetQuery() {
  query.tenantId = undefined
  query.walletType = undefined
  query.coin = ''
  query.symbol = ''
  query.coinType = undefined
  query.chainCode = undefined
  query.appVisible = undefined
  query.rechargeEnabled = undefined
  query.withdrawEnabled = undefined
  query.transferEnabled = undefined
  query.enabled = undefined
  query.limit = 100
  loadList()
}

async function showDetail(row: AssetCoinConfig) {
  const res = await assetService.getCoinConfig({
    id: Number(row.id),
    tenantId: Number(row.tenantId),
  })
  detailData.value = res.data || row
  detailVisible.value = true
}

function openCreateDialog() {
  isEdit.value = false
  resetFormData()
  dialogVisible.value = true
}

async function openEditDialog(row: AssetCoinConfig) {
  isEdit.value = true
  const res = await assetService.getCoinConfig({
    id: Number(row.id),
    tenantId: Number(row.tenantId),
  })
  fillForm(res.data || row)
  dialogVisible.value = true
}

async function submitForm() {
  if (!formRef.value) return
  await formRef.value.validate()
  submitLoading.value = true
  try {
    const payload = {
      id: form.id,
      tenantId: form.tenantId,
      walletType: form.walletType,
      coin: form.coin,
      symbol: form.symbol,
      coinName: form.coinName,
      coinType: form.coinType,
      chainCode: form.chainCode,
      iconUrl: form.iconUrl,
      iconText: form.iconText,
      iconBgColor: form.iconBgColor,
      decimalPlaces: form.decimalPlaces,
      appVisible: form.appVisible,
      rechargeEnabled: form.rechargeEnabled,
      withdrawEnabled: form.withdrawEnabled,
      transferEnabled: form.transferEnabled,
      enabled: form.enabled,
      sort: form.sort,
      remark: form.remark,
    }
    if (isEdit.value) {
      await assetService.updateCoinConfig(payload)
    } else {
      const { id: _, ...createPayload } = payload
      await assetService.createCoinConfig(createPayload)
    }
    ElMessage.success(t('common.success'))
    dialogVisible.value = false
    loadList()
  } finally {
    submitLoading.value = false
  }
}

async function deleteRow(row: AssetCoinConfig) {
  try {
    await ElMessageBox.confirm(t('common.confirmDelete'), t('common.warning'), {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      type: 'warning',
    })
    await assetService.deleteCoinConfig(Number(row.id), { tenantId: Number(row.tenantId) })
    ElMessage.success(t('common.deleteSuccess'))
    loadList()
  } catch (error: unknown) {
    if (error !== 'cancel') {
      ElMessage.error(error instanceof Error ? error.message : t('common.deleteFailed'))
    }
  }
}

function handleLimitChange() {
  resetAndLoad(loadList)
}

function handlePrevPage() {
  prevAndLoad(loadList)
}

function handleNextPage() {
  nextAndLoad(loadList)
}

onMounted(loadList)
onMounted(loadOptions)
</script>

<style scoped>
.query-card :deep(.el-form-item) {
  margin-bottom: 12px;
}

.detail-pre {
  margin: 0;
  max-height: 520px;
  overflow: auto;
  padding: 12px;
  border-radius: 6px;
  background: #f6f8fa;
  color: #24292f;
  line-height: 1.6;
}
</style>
