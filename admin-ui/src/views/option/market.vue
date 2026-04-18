<template>
  <div class="module-page">
    <div class="page-header">
      <h2>{{ t('option.marketSnapshots') }}</h2>
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

    <el-card shadow="never" class="table-card">
      <el-table v-loading="loading" :data="rows" stripe>
        <el-table-column :label="t('option.contractId')" prop="contractId" width="100" />
        <el-table-column :label="t('option.underlyingPrice')" prop="underlyingPrice" min-width="120" show-overflow-tooltip />
        <el-table-column :label="t('option.markPrice')" prop="markPrice" min-width="120" show-overflow-tooltip />
        <el-table-column :label="t('option.lastPrice')" prop="lastPrice" min-width="120" show-overflow-tooltip />
        <el-table-column :label="t('option.snapshotTime')" prop="snapshotTime" min-width="160" show-overflow-tooltip />
        <el-table-column :label="t('common.actions')" width="100" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="showDetail(row)">{{ t('option.detail') }}</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="marketVisible" :title="t('option.updateMarket')" width="720px">
      <el-form label-width="110px">
        <el-form-item :label="t('option.tenantId')">
          <el-input-number v-model="marketForm.tenantId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('option.contractId')">
          <el-input-number v-model="marketForm.contractId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('option.underlyingPrice')"><el-input v-model="marketForm.underlyingPrice" /></el-form-item>
        <el-form-item :label="t('option.markPrice')"><el-input v-model="marketForm.markPrice" /></el-form-item>
        <el-form-item :label="t('option.lastPrice')"><el-input v-model="marketForm.lastPrice" /></el-form-item>
        <el-form-item :label="t('option.bidPrice')"><el-input v-model="marketForm.bidPrice" /></el-form-item>
        <el-form-item :label="t('option.askPrice')"><el-input v-model="marketForm.askPrice" /></el-form-item>
        <el-form-item :label="t('option.theoreticalPrice')"><el-input v-model="marketForm.theoreticalPrice" /></el-form-item>
        <el-form-item :label="t('option.intrinsicValue')"><el-input v-model="marketForm.intrinsicValue" /></el-form-item>
        <el-form-item :label="t('option.timeValue')"><el-input v-model="marketForm.timeValue" /></el-form-item>
        <el-form-item :label="t('option.iv')"><el-input v-model="marketForm.iv" /></el-form-item>
        <el-form-item :label="t('option.delta')"><el-input v-model="marketForm.delta" /></el-form-item>
        <el-form-item :label="t('option.gamma')"><el-input v-model="marketForm.gamma" /></el-form-item>
        <el-form-item :label="t('option.theta')"><el-input v-model="marketForm.theta" /></el-form-item>
        <el-form-item :label="t('option.vega')"><el-input v-model="marketForm.vega" /></el-form-item>
        <el-form-item :label="t('option.rho')"><el-input v-model="marketForm.rho" /></el-form-item>
        <el-form-item :label="t('option.riskFreeRate')"><el-input v-model="marketForm.riskFreeRate" /></el-form-item>
        <el-form-item :label="t('option.pricingModel')"><el-input v-model="marketForm.pricingModel" /></el-form-item>
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
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { optionService, UpdateMarketReq } from '@/services'

const { t } = useI18n()

const loading = ref(false)
const submitLoading = ref(false)
const rows = ref<Record<string, any>[]>([])
const detailVisible = ref(false)
const detailData = ref<Record<string, any>>({})
const marketVisible = ref(false)
const query = reactive({
  tenantId: undefined as number | undefined,
  contractId: undefined as number | undefined,
  limit: 100,
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

const pickList = (res: any) => res?.data || res?.list || []

const loadCurrent = async () => {
  loading.value = true
  try {
    rows.value = pickList(await optionService.listMarketSnapshots(query))
  } finally {
    loading.value = false
  }
}

const resetCurrent = () => {
  query.tenantId = undefined
  query.contractId = undefined
  query.limit = 100
  loadCurrent()
}

const showDetail = async (row: Record<string, any>) => {
  detailData.value = (await optionService.getMarket({ tenantId: row.tenantId, contractId: row.contractId })).data || row
  detailVisible.value = true
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
    ElMessage.success(t('common.updateSuccess'))
    marketVisible.value = false
    loadCurrent()
  } finally {
    submitLoading.value = false
  }
}

onMounted(loadCurrent)
</script>

<style scoped></style>
