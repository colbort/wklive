<template>
  <div class="module-page">
    <div class="page-header">
      <h2>{{ t('option.marketDetail') }}</h2>
      <div class="header-actions">
        <el-button @click="loadCurrent">{{ t('common.refresh') }}</el-button>
        <el-button type="primary" plain @click="openMarketDialog()">{{ t('option.updateMarket') }}</el-button>
      </div>
    </div>

    <el-card shadow="never" class="query-card">
      <el-form :model="query" inline label-width="90px">
        <el-form-item :label="t('option.tenantId')">
          <el-input-number v-model="query.tenantId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('option.contractId')">
          <el-input-number v-model="query.contractId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadCurrent">{{ t('common.search') }}</el-button>
          <el-button @click="resetCurrent">{{ t('common.reset') }}</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <div v-loading="loading" class="detail-sections">
      <el-card shadow="never" class="table-card">
        <template #header>
          <div class="card-header">
            <span>{{ t('option.contractInfo') }}</span>
          </div>
        </template>
        <el-descriptions :column="2" border>
          <el-descriptions-item label="ID">
            {{ detail.contract?.id ?? '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.tenantId')">
            {{ detail.contract?.tenantId ?? '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.contractId')">
            {{ detail.market?.contractId ?? detail.contract?.id ?? '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.contractCode')">
            {{ detail.contract?.contractCode || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.underlying')">
            {{ detail.contract?.underlyingSymbol || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.optionType')">
            {{ formatOptionValue('optionType', detail.contract?.optionType) }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.exerciseStyle')">
            {{ formatOptionValue('exerciseStyle', detail.contract?.exerciseStyle) }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.settlementType')">
            {{ detail.contract?.settlementType ?? '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.settleCoin')">
            {{ detail.contract?.settleCoin || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.quoteCoin')">
            {{ detail.contract?.quoteCoin || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.strikePrice')">
            {{ detail.contract?.strikePrice || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('common.status')">
            {{ formatOptionValue('contractStatus', detail.contract?.status) }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.listTime')">
            {{ formatDate(detail.contract?.listTime??0) }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.expireTime')">
            {{ formatDate(detail.contract?.expireTime??0) }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.deliverTime')">
            {{ formatDate(detail.contract?.deliverTime??0) }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('common.remark')">
            {{ detail.contract?.remark || '-' }}
          </el-descriptions-item>
        </el-descriptions>
      </el-card>

      <el-card shadow="never" class="table-card">
        <template #header>
          <div class="card-header">
            <span>{{ t('option.marketInfo') }}</span>
            <el-button link type="primary" @click="openMarketDialog(detail.market)">
              {{ t('common.edit') }}
            </el-button>
          </div>
        </template>
        <el-descriptions :column="2" border>
          <el-descriptions-item label="ID">
            {{ detail.market?.id ?? '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.snapshotTime')">
            {{ formatDate(detail.market?.snapshotTime??0) }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.underlyingPrice')">
            {{ detail.market?.underlyingPrice || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.markPrice')">
            {{ detail.market?.markPrice || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.lastPrice')">
            {{ detail.market?.lastPrice || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.bidPrice')">
            {{ detail.market?.bidPrice || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.askPrice')">
            {{ detail.market?.askPrice || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.theoreticalPrice')">
            {{ detail.market?.theoreticalPrice || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.intrinsicValue')">
            {{ detail.market?.intrinsicValue || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.timeValue')">
            {{ detail.market?.timeValue || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.iv')">
            {{ detail.market?.iv || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.delta')">
            {{ detail.market?.delta || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.gamma')">
            {{ detail.market?.gamma || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.theta')">
            {{ detail.market?.theta || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.vega')">
            {{ detail.market?.vega || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.rho')">
            {{ detail.market?.rho || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.riskFreeRate')">
            {{ detail.market?.riskFreeRate || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.pricingModel')">
            {{ detail.market?.pricingModel || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('common.createTimes')">
            {{ formatDate(detail.market?.createTimes??0) }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('common.updateTimes')">
            {{ formatDate(detail.market?.updateTimes??0) }}
          </el-descriptions-item>
        </el-descriptions>
      </el-card>
    </div>

    <el-empty v-if="!loading && !detail.market && !detail.contract" :description="t('common.noData')" />

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
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { optionService, type OptionContract, type OptionGroup, type OptionMarket, type UpdateMarketReq } from '@/services'
import { formatDate } from '@/utils'
import { getOptionValueLabel } from '@/utils/options'

type MarketDetailState = {
  contract?: OptionContract
  market?: OptionMarket
}

const { t } = useI18n()
const route = useRoute()

const loading = ref(false)
const submitLoading = ref(false)
const marketVisible = ref(false)
const optionGroups = ref<OptionGroup[]>([])
const detail = reactive<MarketDetailState>({})
const query = reactive({
  tenantId: undefined as number | undefined,
  contractId: undefined as number | undefined,
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

const resetMarketForm = () => {
  Object.assign(marketForm, {
    tenantId: Number(query.tenantId || detail.contract?.tenantId || detail.market?.tenantId || 0),
    contractId: Number(query.contractId || detail.contract?.id || detail.market?.contractId || 0),
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

const loadOptions = async () => {
  optionGroups.value = (await optionService.getOptions()).data || []
}

const formatOptionValue = (key: string, value: number | string | undefined) => {
  if (value === undefined || value === null || value === '') return '-'
  return getOptionValueLabel(optionGroups.value, key, value, t) || value
}

const loadCurrent = async () => {
  if (!query.contractId) {
    detail.contract = undefined
    detail.market = undefined
    resetMarketForm()
    return
  }

  loading.value = true
  try {
    const res = await optionService.getMarket({
      tenantId: query.tenantId,
      contractId: query.contractId,
    })
    detail.market = res.data
    detail.contract = (res.data as unknown as { contract?: OptionContract })?.contract
    if (!detail.contract && query.contractId) {
      const contractRes = await optionService.getContract({
        tenantId: query.tenantId,
        id: query.contractId,
      })
      detail.contract = contractRes.data?.contract || (contractRes.data as unknown as OptionContract)
    }
    resetMarketForm()
  } finally {
    loading.value = false
  }
}

const resetCurrent = () => {
  query.tenantId = undefined
  query.contractId = undefined
  detail.contract = undefined
  detail.market = undefined
  resetMarketForm()
}

const openMarketDialog = (row?: OptionMarket) => {
  resetMarketForm()
  if (row) {
    Object.assign(marketForm, {
      tenantId: row.tenantId || marketForm.tenantId,
      contractId: row.contractId || marketForm.contractId,
      underlyingPrice: row.underlyingPrice || '',
      markPrice: row.markPrice || '',
      lastPrice: row.lastPrice || '',
      bidPrice: row.bidPrice || '',
      askPrice: row.askPrice || '',
      theoreticalPrice: row.theoreticalPrice || '',
      intrinsicValue: row.intrinsicValue || '',
      timeValue: row.timeValue || '',
      iv: row.iv || '',
      delta: row.delta || '',
      gamma: row.gamma || '',
      theta: row.theta || '',
      vega: row.vega || '',
      rho: row.rho || '',
      riskFreeRate: row.riskFreeRate || '',
      pricingModel: row.pricingModel || '',
      snapshotTime: row.snapshotTime || 0,
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
    query.tenantId = marketForm.tenantId
    query.contractId = marketForm.contractId
    await loadCurrent()
  } finally {
    submitLoading.value = false
  }
}

onMounted(async () => {
  query.tenantId = route.query.tenantId ? Number(route.query.tenantId) : undefined
  query.contractId = route.query.contractId ? Number(route.query.contractId) : undefined
  await loadOptions()
  resetMarketForm()
  await loadCurrent()
})
</script>

<style scoped>
.detail-sections {
  display: grid;
  gap: 16px;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
</style>
