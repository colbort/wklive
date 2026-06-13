<template>
  <div class="module-page">
    <div class="page-header">
      <h2>{{ t('option.marketSnapshots') }}</h2>
      <div class="header-actions">
        <el-button @click="loadCurrent">
          {{ t('common.refresh') }}
        </el-button>
        <el-button
          v-perm="'option:market:update'"
          type="primary"
          plain
          @click="openMarketDialog()"
        >
          {{ t('option.updateMarket') }}
        </el-button>
      </div>
    </div>

    <CrudQueryCard :model="query" label-width="auto" :show-actions="false">
      <el-form-item :label="t('option.tenantId')">
        <TenantSelect v-model="query.tenantId" class="tenant-select-filter" />
      </el-form-item>
      <el-form-item :label="t('option.contractId')">
        <el-input-number v-model="query.contractId" :min="0" :precision="0" />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="loadCurrent">
          {{ t('common.search') }}
        </el-button>
        <el-button @click="resetCurrent">
          {{ t('common.reset') }}
        </el-button>
      </el-form-item>
    </CrudQueryCard>

    <el-card shadow="never" class="table-card">
      <el-table v-loading="loading" :data="rows" stripe>
        <el-table-column :label="t('option.contractId')" prop="contractId" width="100" />
        <el-table-column
          :label="t('option.underlyingPrice')"
          prop="underlyingPrice"
          min-width="120"
          show-overflow-tooltip
        />
        <el-table-column
          :label="t('option.markPrice')"
          prop="markPrice"
          min-width="120"
          show-overflow-tooltip
        />
        <el-table-column
          :label="t('option.lastPrice')"
          prop="lastPrice"
          min-width="120"
          show-overflow-tooltip
        />
        <el-table-column
          :label="t('option.snapshotTime')"
          prop="snapshotTime"
          min-width="160"
          show-overflow-tooltip
        />
        <el-table-column :label="t('common.actions')" width="100" fixed="right">
          <template #default="{ row }">
            <el-button
              v-perm="'option:market:detail'"
              link
              type="primary"
              @click="showDetail(row)"
            >
              {{ t('option.detail') }}
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

    <el-dialog v-model="marketVisible" :title="t('option.updateMarket')" width="720px">
      <el-form label-width="110px">
        <el-form-item :label="t('option.tenantId')">
          <TenantSelect v-model="marketForm.tenantId" include-system />
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
        <el-button @click="marketVisible = false">
          {{ t('common.cancel') }}
        </el-button>
        <el-button
          v-perm="'option:market:update'"
          type="primary"
          :loading="submitLoading"
          @click="submitMarket"
        >
          {{ t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="detailVisible" :title="t('option.detail')" width="860px">
      <div v-loading="detailLoading" class="market-detail">
        <el-descriptions
          v-if="detailData?.contract"
          :title="t('option.contractInfo')"
          :column="2"
          border
        >
          <el-descriptions-item label="ID">
            {{ detailData.contract.id ?? '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.tenantId')">
            {{ detailData.contract.tenantId ?? '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.contractCode')">
            {{ detailData.contract.contractCode || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.underlying')">
            {{ detailData.contract.underlyingSymbol || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.settleCoin')">
            {{ detailData.contract.settleCoin || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.quoteCoin')">
            {{ detailData.contract.quoteCoin || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.strikePrice')">
            {{ detailData.contract.strikePrice || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('common.status')">
            {{ detailData.contract.status ?? '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.listTime')">
            {{ formatDate(detailData.contract.listTime || 0) }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.expireTime')">
            {{ formatDate(detailData.contract.expireTime || 0) }}
          </el-descriptions-item>
        </el-descriptions>

        <el-descriptions :title="t('option.marketInfo')" :column="2" border>
          <el-descriptions-item label="ID">
            {{ detailData?.id ?? '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.tenantId')">
            {{ detailData?.tenantId ?? '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.contractId')">
            {{ detailData?.contractId ?? '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.snapshotTime')">
            {{ formatDate(detailData?.snapshotTime || 0) }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.underlyingPrice')">
            {{ detailData?.underlyingPrice || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.markPrice')">
            {{ detailData?.markPrice || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.lastPrice')">
            {{ detailData?.lastPrice || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.bidPrice')">
            {{ detailData?.bidPrice || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.askPrice')">
            {{ detailData?.askPrice || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.theoreticalPrice')">
            {{ detailData?.theoreticalPrice || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.intrinsicValue')">
            {{ detailData?.intrinsicValue || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.timeValue')">
            {{ detailData?.timeValue || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.iv')">
            {{ detailData?.iv || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.delta')">
            {{ detailData?.delta || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.gamma')">
            {{ detailData?.gamma || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.theta')">
            {{ detailData?.theta || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.vega')">
            {{ detailData?.vega || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.rho')">
            {{ detailData?.rho || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.riskFreeRate')">
            {{ detailData?.riskFreeRate || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.pricingModel')">
            {{ detailData?.pricingModel || '-' }}
          </el-descriptions-item>
        </el-descriptions>
      </div>

      <template #footer>
        <el-button @click="detailVisible = false">
          {{ t('common.close') }}
        </el-button>
        <el-button
          v-perm="'option:market:update'"
          type="primary"
          @click="openMarketDialog(detailData || undefined)"
        >
          {{ t('option.updateMarket') }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { usePagination } from '@/composables'
import TenantSelect from '@/components/TenantSelect.vue'
import { formatDate } from '@/utils'
import {
  ListMarketSnapshotsReq,
  optionService,
  type OptionContract,
  type OptionMarket,
  type OptionMarketSnapshot,
  UpdateMarketReq,
} from '@/services'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'

type MarketDetail = Partial<OptionMarket> &
  Partial<OptionMarketSnapshot> & {
    contract?: OptionContract
  }

const { t } = useI18n()
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)

const loading = ref(false)
const detailLoading = ref(false)
const submitLoading = ref(false)
const rows = ref<OptionMarketSnapshot[]>([])
const detailVisible = ref(false)
const detailData = ref<MarketDetail | null>(null)
const marketVisible = ref(false)
const query = reactive<ListMarketSnapshotsReq>({
  tenantId: undefined as number | undefined,
  contractId: 0,
  limit: 20,
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

const loadCurrent = async () => {
  loading.value = true
  try {
    const resp = await optionService.listMarketSnapshots({
      ...query,
      cursor: pagination.cursor,
      limit: pagination.limit,
    })
    if (resp.code !== 200) {
      ElMessage.error(resp.msg || t('common.loadFailed'))
      return
    }
    rows.value = resp?.data || []
    updateFromResponse(resp)
  } finally {
    loading.value = false
  }
}

const resetCurrent = () => {
  query.tenantId = undefined
  query.contractId = 0
  query.limit = 100
  loadCurrent()
}

const showDetail = async (row: OptionMarketSnapshot) => {
  detailVisible.value = true
  detailLoading.value = true
  detailData.value = row
  try {
    const market: MarketDetail =
      ((await optionService.getMarket({ tenantId: row.tenantId, contractId: row.contractId }))
        .data as MarketDetail | undefined) || row
    if (!market.contract && row.contractId) {
      const contractRes = await optionService.getContract({
        tenantId: row.tenantId,
        id: row.contractId,
      })
      market.contract = contractRes.data?.contract
    }
    detailData.value = market
  } finally {
    detailLoading.value = false
  }
}

const openMarketDialog = (row?: { tenantId?: number; contractId?: number } | null) => {
  Object.assign(marketForm, {
    tenantId: row?.tenantId || 0,
    contractId: row?.contractId || 0,
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

function handleLimitChange() {
  resetAndLoad(loadCurrent)
}

function handlePrevPage() {
  prevAndLoad(loadCurrent)
}

function handleNextPage() {
  nextAndLoad(loadCurrent)
}

onMounted(loadCurrent)
</script>

<style scoped>
.market-detail {
  display: grid;
  gap: 18px;
}
</style>
