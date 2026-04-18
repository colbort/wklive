<template>
  <div class="module-page">
    <div class="page-header">
      <h2>{{ t('option.contracts') }}</h2>
      <div class="header-actions">
        <el-button @click="loadCurrent">{{ t('common.refresh') }}</el-button>
        <el-button type="primary" @click="openContractDialog()">{{ t('option.createContract') }}</el-button>
      </div>
    </div>

    <el-card shadow="never" class="query-card">
      <el-form :model="query" inline label-width="90px">
        <el-form-item :label="t('option.tenantId')">
          <el-input-number v-model="query.tenantId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('option.contractCode')">
          <el-input v-model="query.contractCode" clearable />
        </el-form-item>
        <el-form-item :label="t('option.underlying')">
          <el-input v-model="query.underlyingSymbol" clearable />
        </el-form-item>
        <el-form-item :label="t('option.optionType')">
          <el-select v-model="query.optionType" clearable style="width: 180px">
            <el-option
              v-for="item in optionTypeOptions"
              :key="item.value"
              :label="getOptionLabel(t, item.code, item.value)"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('common.status')">
          <el-select v-model="query.status" clearable style="width: 180px">
            <el-option
              v-for="item in statusOptions"
              :key="item.value"
              :label="getOptionLabel(t, item.code, item.value)"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadCurrent">{{ t('common.search') }}</el-button>
          <el-button @click="resetCurrent">{{ t('common.reset') }}</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="never" class="table-card">
      <el-table v-loading="loading" :data="rows" stripe>
        <el-table-column prop="id" label="ID" width="90" />
        <el-table-column :label="t('option.tenantId')" prop="tenantId" width="100" />
        <el-table-column :label="t('option.contractCode')" prop="contractCode" min-width="180" show-overflow-tooltip />
        <el-table-column
          prop="underlyingSymbol"
          :label="t('option.underlying')"
          min-width="140"
          show-overflow-tooltip
        />
        <el-table-column :label="t('option.settleCoin')" prop="settleCoin" width="100" />
        <el-table-column :label="t('option.quoteCoin')" prop="quoteCoin" width="100" />
        <el-table-column :label="t('option.strikePrice')" prop="strikePrice" min-width="120" show-overflow-tooltip />
        <el-table-column :label="t('common.status')" prop="status" width="100" />
        <el-table-column :label="t('common.actions')" width="160" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="showDetail(row)">{{ t('option.detail') }}</el-button>
            <el-button link type="primary" @click="openContractDialog(row)">{{ t('common.edit') }}</el-button>
            <el-button link type="primary" @click="openMarketDialog(row)">{{ t('option.market') }}</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog
      v-model="contractVisible"
      :title="contractForm.id ? t('option.editContract') : t('option.createContract')"
      width="760px"
    >
      <el-form label-width="110px">
        <el-form-item :label="t('option.tenantId')">
          <el-input-number v-model="contractForm.tenantId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('option.contractCode')">
          <el-input v-model="contractForm.contractCode" />
        </el-form-item>
        <el-form-item :label="t('option.underlying')">
          <el-input v-model="contractForm.underlyingSymbol" />
        </el-form-item>
        <el-form-item :label="t('option.settleCoin')">
          <el-input v-model="contractForm.settleCoin" />
        </el-form-item>
        <el-form-item :label="t('option.quoteCoin')">
          <el-input v-model="contractForm.quoteCoin" />
        </el-form-item>
        <el-form-item :label="t('option.optionType')">
          <el-select v-model="contractForm.optionType" style="width: 100%">
            <el-option
              v-for="item in optionTypeOptions"
              :key="item.value"
              :label="getOptionLabel(t, item.code, item.value)"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('option.exerciseStyle')">
          <el-select v-model="contractForm.exerciseStyle" style="width: 100%">
            <el-option
              v-for="item in exerciseStyleOptions"
              :key="item.value"
              :label="getOptionLabel(t, item.code, item.value)"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('option.settlementType')">
          <el-input-number v-model="contractForm.settlementType" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('option.strikePrice')">
          <el-input v-model="contractForm.strikePrice" />
        </el-form-item>
        <el-form-item :label="t('option.contractUnit')">
          <el-input v-model="contractForm.contractUnit" />
        </el-form-item>
        <el-form-item :label="t('option.minOrderQty')">
          <el-input v-model="contractForm.minOrderQty" />
        </el-form-item>
        <el-form-item :label="t('option.maxOrderQty')">
          <el-input v-model="contractForm.maxOrderQty" />
        </el-form-item>
        <el-form-item :label="t('option.priceTick')">
          <el-input v-model="contractForm.priceTick" />
        </el-form-item>
        <el-form-item :label="t('option.qtyStep')">
          <el-input v-model="contractForm.qtyStep" />
        </el-form-item>
        <el-form-item :label="t('option.multiplier')">
          <el-input v-model="contractForm.multiplier" />
        </el-form-item>
        <el-form-item :label="t('option.listTime')">
          <el-input-number v-model="contractForm.listTime" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('option.expireTime')">
          <el-input-number v-model="contractForm.expireTime" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('option.deliverTime')">
          <el-input-number v-model="contractForm.deliverTime" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('option.autoExercise')">
          <el-input-number v-model="contractForm.isAutoExercise" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('common.status')">
          <el-select v-model="contractForm.status" style="width: 100%">
            <el-option
              v-for="item in statusOptions"
              :key="item.value"
              :label="getOptionLabel(t, item.code, item.value)"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('common.sort')">
          <el-input-number v-model="contractForm.sort" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('common.remark')">
          <el-input v-model="contractForm.remark" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="contractVisible = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="submitLoading" @click="submitContract">{{ t('common.confirm') }}</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="marketVisible" :title="t('option.updateMarket')" width="720px">
      <el-form label-width="110px">
        <el-form-item :label="t('option.tenantId')">
          <el-input-number v-model="marketForm.tenantId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('option.contractId')">
          <el-input-number v-model="marketForm.contractId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('option.underlyingPrice')">
          <el-input v-model="marketForm.underlyingPrice" />
        </el-form-item>
        <el-form-item :label="t('option.markPrice')">
          <el-input v-model="marketForm.markPrice" />
        </el-form-item>
        <el-form-item :label="t('option.lastPrice')">
          <el-input v-model="marketForm.lastPrice" />
        </el-form-item>
        <el-form-item :label="t('option.bidPrice')">
          <el-input v-model="marketForm.bidPrice" />
        </el-form-item>
        <el-form-item :label="t('option.askPrice')">
          <el-input v-model="marketForm.askPrice" />
        </el-form-item>
        <el-form-item :label="t('option.theoreticalPrice')">
          <el-input v-model="marketForm.theoreticalPrice" />
        </el-form-item>
        <el-form-item :label="t('option.intrinsicValue')">
          <el-input v-model="marketForm.intrinsicValue" />
        </el-form-item>
        <el-form-item :label="t('option.timeValue')">
          <el-input v-model="marketForm.timeValue" />
        </el-form-item>
        <el-form-item :label="t('option.iv')">
          <el-input v-model="marketForm.iv" />
        </el-form-item>
        <el-form-item :label="t('option.delta')">
          <el-input v-model="marketForm.delta" />
        </el-form-item>
        <el-form-item :label="t('option.gamma')">
          <el-input v-model="marketForm.gamma" />
        </el-form-item>
        <el-form-item :label="t('option.theta')">
          <el-input v-model="marketForm.theta" />
        </el-form-item>
        <el-form-item :label="t('option.vega')">
          <el-input v-model="marketForm.vega" />
        </el-form-item>
        <el-form-item :label="t('option.rho')">
          <el-input v-model="marketForm.rho" />
        </el-form-item>
        <el-form-item :label="t('option.riskFreeRate')">
          <el-input v-model="marketForm.riskFreeRate" />
        </el-form-item>
        <el-form-item :label="t('option.pricingModel')">
          <el-input v-model="marketForm.pricingModel" />
        </el-form-item>
        <el-form-item :label="t('option.snapshotTime')">
          <el-input-number v-model="marketForm.snapshotTime" :min="0" :precision="0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="marketVisible = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="submitLoading" @click="submitMarket">{{ t('common.confirm') }}</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="detailVisible" :title="t('option.detail')" width="760px">
      <pre class="detail-pre">{{ JSON.stringify(detailData, null, 2) }}</pre>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { optionService, type OptionGroup, type UpdateContractReq, type UpdateMarketReq } from '@/services'
import { findOptionGroup, getOptionLabel } from '@/utils/options'

const { t } = useI18n()

const loading = ref(false)
const submitLoading = ref(false)
const rows = ref<Record<string, any>[]>([])
const detailVisible = ref(false)
const detailData = ref<Record<string, any>>({})
const contractVisible = ref(false)
const marketVisible = ref(false)
const optionGroups = ref<OptionGroup[]>([])

const query = reactive({
  tenantId: undefined as number | undefined,
  contractCode: '',
  underlyingSymbol: '',
  optionType: undefined as number | undefined,
  status: undefined as number | undefined,
  limit: 100,
})

const contractForm = reactive<UpdateContractReq>({
  id: 0,
  tenantId: 0,
  contractCode: '',
  underlyingSymbol: '',
  settleCoin: '',
  quoteCoin: '',
  optionType: 0,
  exerciseStyle: 0,
  settlementType: 0,
  strikePrice: '',
  contractUnit: '',
  minOrderQty: '',
  maxOrderQty: '',
  priceTick: '',
  qtyStep: '',
  multiplier: '',
  listTime: 0,
  expireTime: 0,
  deliverTime: 0,
  isAutoExercise: 0,
  status: 0,
  sort: 0,
  remark: '',
  isDeleted: 0,
})

const marketForm = reactive<UpdateMarketReq>({
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

const optionTypeOptions = computed(() => findOptionGroup(optionGroups.value, 'optionType'))
const exerciseStyleOptions = computed(() => findOptionGroup(optionGroups.value, 'exerciseStyle'))
const statusOptions = computed(() => findOptionGroup(optionGroups.value, 'status'))

const pickList = (res: any) => res?.data || res?.list || []

const loadOptions = async () => {
  optionGroups.value = (await optionService.getOptions()).data || []
}

const loadCurrent = async () => {
  loading.value = true
  try {
    rows.value = pickList(await optionService.listContracts(query))
  } finally {
    loading.value = false
  }
}

const resetContractForm = () => {
  Object.assign(contractForm, {
    id: 0,
    tenantId: 0,
    contractCode: '',
    underlyingSymbol: '',
    settleCoin: '',
    quoteCoin: '',
    optionType: optionTypeOptions.value[0]?.value || 0,
    exerciseStyle: exerciseStyleOptions.value[0]?.value || 0,
    settlementType: 0,
    strikePrice: '',
    contractUnit: '',
    minOrderQty: '',
    maxOrderQty: '',
    priceTick: '',
    qtyStep: '',
    multiplier: '',
    listTime: 0,
    expireTime: 0,
    deliverTime: 0,
    isAutoExercise: 0,
    status: statusOptions.value[0]?.value || 0,
    sort: 0,
    remark: '',
    isDeleted: 0,
  })
}

const resetMarketForm = () => {
  Object.assign(marketForm, {
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
}

const resetCurrent = () => {
  query.tenantId = undefined
  query.contractCode = ''
  query.underlyingSymbol = ''
  query.optionType = undefined
  query.status = undefined
  query.limit = 100
  loadCurrent()
}

const showDetail = async (row: Record<string, any>) => {
  detailData.value =
    (await optionService.getContract({
      tenantId: row.tenantId,
      id: row.id,
      contractCode: row.contractCode,
    })).data || row
  detailVisible.value = true
}

const openContractDialog = (row?: Record<string, any>) => {
  resetContractForm()
  if (row) {
    Object.assign(contractForm, row)
  }
  contractVisible.value = true
}

const submitContract = async () => {
  submitLoading.value = true
  try {
    if (contractForm.id) {
      await optionService.updateContract(contractForm)
    } else {
      await optionService.createContract(contractForm)
    }
    ElMessage.success(contractForm.id ? t('common.updateSuccess') : t('common.createSuccess'))
    contractVisible.value = false
    loadCurrent()
  } finally {
    submitLoading.value = false
  }
}

const openMarketDialog = (row?: Record<string, any>) => {
  resetMarketForm()
  if (row) {
    Object.assign(marketForm, {
      tenantId: row.tenantId || 0,
      contractId: row.contractId || row.id || 0,
    })
  }
  marketVisible.value = true
}

const submitMarket = async () => {
  submitLoading.value = true
  try {
    await optionService.updateMarket(marketForm)
    ElMessage.success(t('common.updateSuccess'))
    marketVisible.value = false
  } finally {
    submitLoading.value = false
  }
}

onMounted(async () => {
  await loadOptions()
  resetContractForm()
  await loadCurrent()
})
</script>

<style scoped></style>
