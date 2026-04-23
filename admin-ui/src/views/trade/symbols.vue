<template>
  <div class="module-page">
    <div class="page-header">
      <h2>{{ t('trade.symbols') }}</h2>
      <div class="header-actions">
        <el-button @click="loadCurrent">
          {{ t('common.refresh') }}
        </el-button>
        <el-button type="primary" @click="openSymbolDialog()">
          {{ t('trade.addSymbol') }}
        </el-button>
      </div>
    </div>

    <el-card shadow="never" class="query-card">
      <el-form :model="currentQuery" inline label-width="90px">
        <el-form-item v-for="field in currentFields" :key="field.key" :label="field.label">
          <el-input v-if="field.type !== 'number'" v-model="currentQuery[field.key]" clearable />

          <el-input-number v-else v-model="currentQuery[field.key]" :min="0" :precision="0" />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="loadCurrent">
            {{ t('common.search') }}
          </el-button>
          <el-button @click="resetCurrent">
            {{ t('common.reset') }}
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

        <el-table-column :label="t('common.actions')" width="240" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="showDetail(row)">
              {{ t('option.detail') }}
            </el-button>

            <el-button link type="primary" @click="openSymbolDialog(row)">
              {{ t('common.edit') }}
            </el-button>

            <el-button link type="primary" @click="openSpotDialog(row)">
              {{ t('trade.spotConfig') }}
            </el-button>

            <el-button link type="primary" @click="openContractDialog(row)">
              {{ t('trade.contractConfig') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog
      v-model="symbolVisible"
      :title="symbolForm.id ? t('trade.editSymbol') : t('trade.addSymbol')"
      width="760px"
    >
      <el-form label-width="110px">
        <el-form-item :label="t('trade.tenantId')">
          <el-input-number v-model="symbolForm.tenantId" :min="0" :precision="0" />
        </el-form-item>

        <el-form-item :label="t('trade.symbol')">
          <el-input v-model="symbolForm.symbol" />
        </el-form-item>

        <el-form-item :label="t('trade.displaySymbol')">
          <el-input v-model="symbolForm.displaySymbol" />
        </el-form-item>

        <el-form-item :label="t('trade.marketType')">
          <el-input-number v-model="symbolForm.marketType" :min="0" :precision="0" />
        </el-form-item>

        <el-form-item :label="t('trade.baseAsset')">
          <el-input v-model="symbolForm.baseAsset" />
        </el-form-item>

        <el-form-item :label="t('trade.quoteAsset')">
          <el-input v-model="symbolForm.quoteAsset" />
        </el-form-item>

        <el-form-item :label="t('trade.settleAsset')">
          <el-input v-model="symbolForm.settleAsset" />
        </el-form-item>

        <el-form-item :label="t('trade.contractType')">
          <el-input-number v-model="symbolForm.contractType" :min="0" :precision="0" />
        </el-form-item>

        <el-form-item :label="t('trade.status')">
          <el-input-number v-model="symbolForm.status" :min="0" :precision="0" />
        </el-form-item>

        <el-form-item :label="t('trade.priceScale')">
          <el-input-number v-model="symbolForm.priceScale" :min="0" :precision="0" />
        </el-form-item>

        <el-form-item :label="t('trade.qtyScale')">
          <el-input-number v-model="symbolForm.qtyScale" :min="0" :precision="0" />
        </el-form-item>

        <el-form-item :label="t('trade.minPrice')">
          <el-input v-model="symbolForm.minPrice" />
        </el-form-item>

        <el-form-item :label="t('trade.maxPrice')">
          <el-input v-model="symbolForm.maxPrice" />
        </el-form-item>

        <el-form-item :label="t('trade.priceTick')">
          <el-input v-model="symbolForm.priceTick" />
        </el-form-item>

        <el-form-item :label="t('trade.minQty')">
          <el-input v-model="symbolForm.minQty" />
        </el-form-item>

        <el-form-item :label="t('trade.maxQty')">
          <el-input v-model="symbolForm.maxQty" />
        </el-form-item>

        <el-form-item :label="t('trade.qtyStep')">
          <el-input v-model="symbolForm.qtyStep" />
        </el-form-item>

        <el-form-item :label="t('trade.minNotional')">
          <el-input v-model="symbolForm.minNotional" />
        </el-form-item>

        <el-form-item :label="t('trade.maxLeverage')">
          <el-input-number v-model="symbolForm.maxLeverage" :min="0" :precision="0" />
        </el-form-item>

        <el-form-item :label="t('trade.openTime')">
          <el-input-number v-model="symbolForm.openTime" :min="0" :precision="0" />
        </el-form-item>

        <el-form-item :label="t('trade.closeTime')">
          <el-input-number v-model="symbolForm.closeTime" :min="0" :precision="0" />
        </el-form-item>

        <el-form-item :label="t('common.sort')">
          <el-input-number v-model="symbolForm.sort" :min="0" :precision="0" />
        </el-form-item>

        <el-form-item :label="t('common.remark')">
          <el-input v-model="symbolForm.remark" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="symbolVisible = false">
          {{ t('common.cancel') }}
        </el-button>
        <el-button type="primary" :loading="submitLoading" @click="submitSymbol">
          {{ t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="spotVisible" :title="t('trade.spotConfig')" width="640px">
      <el-form label-width="110px">
        <el-form-item :label="t('trade.tenantId')">
          <el-input-number v-model="spotForm.tenantId" :min="0" :precision="0" />
        </el-form-item>

        <el-form-item :label="t('trade.symbolId')">
          <el-input-number v-model="spotForm.symbolId" :min="0" :precision="0" />
        </el-form-item>

        <el-form-item :label="t('trade.makerFeeRate')">
          <el-input v-model="spotForm.makerFeeRate" />
        </el-form-item>

        <el-form-item :label="t('trade.takerFeeRate')">
          <el-input v-model="spotForm.takerFeeRate" />
        </el-form-item>

        <el-form-item :label="t('trade.buyEnabled')">
          <el-switch v-model="spotForm.buyEnabled" :active-value="1" :inactive-value="0" />
        </el-form-item>

        <el-form-item :label="t('trade.sellEnabled')">
          <el-switch v-model="spotForm.sellEnabled" :active-value="1" :inactive-value="0" />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="spotVisible = false">
          {{ t('common.cancel') }}
        </el-button>
        <el-button type="primary" :loading="submitLoading" @click="submitSpotConfig">
          {{ t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="contractVisible" :title="t('trade.contractConfig')" width="700px">
      <el-form label-width="120px">
        <el-form-item :label="t('trade.tenantId')">
          <el-input-number v-model="contractForm.tenantId" :min="0" :precision="0" />
        </el-form-item>

        <el-form-item :label="t('trade.symbolId')">
          <el-input-number v-model="contractForm.symbolId" :min="0" :precision="0" />
        </el-form-item>

        <el-form-item :label="t('trade.contractSize')">
          <el-input v-model="contractForm.contractSize" />
        </el-form-item>

        <el-form-item :label="t('option.multiplier')">
          <el-input v-model="contractForm.multiplier" />
        </el-form-item>

        <el-form-item :label="t('trade.maintenanceMarginRate')">
          <el-input v-model="contractForm.maintenanceMarginRate" />
        </el-form-item>

        <el-form-item :label="t('trade.initialMarginRate')">
          <el-input v-model="contractForm.initialMarginRate" />
        </el-form-item>

        <el-form-item :label="t('trade.makerFeeRate')">
          <el-input v-model="contractForm.makerFeeRate" />
        </el-form-item>

        <el-form-item :label="t('trade.takerFeeRate')">
          <el-input v-model="contractForm.takerFeeRate" />
        </el-form-item>

        <el-form-item :label="t('trade.fundingIntervalMinutes')">
          <el-input-number v-model="contractForm.fundingIntervalMinutes" :min="0" :precision="0" />
        </el-form-item>

        <el-form-item :label="t('option.deliverTime')">
          <el-input-number v-model="contractForm.deliveryTime" :min="0" :precision="0" />
        </el-form-item>

        <el-form-item :label="t('trade.supportCross')">
          <el-switch v-model="contractForm.supportCross" :active-value="1" :inactive-value="0" />
        </el-form-item>

        <el-form-item :label="t('trade.supportIsolated')">
          <el-switch v-model="contractForm.supportIsolated" :active-value="1" :inactive-value="0" />
        </el-form-item>

        <el-form-item :label="t('trade.buyEnabled')">
          <el-switch v-model="contractForm.buyEnabled" :active-value="1" :inactive-value="0" />
        </el-form-item>

        <el-form-item :label="t('trade.sellEnabled')">
          <el-switch v-model="contractForm.sellEnabled" :active-value="1" :inactive-value="0" />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="contractVisible = false">
          {{ t('common.cancel') }}
        </el-button>
        <el-button type="primary" :loading="submitLoading" @click="submitContractConfig">
          {{ t('common.confirm') }}
        </el-button>
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
import { tradeService, type TradeSymbol } from '@/services'

const { t } = useI18n()

interface CurrentQuery {
  tenantId: number | undefined
  marketType: number | undefined
  keyword: string
  status: number | undefined
  limit: number
}

type CurrentField =
  | {
      key: 'tenantId' | 'marketType' | 'status'
      label: string
      type: 'number'
    }
  | {
      key: 'keyword'
      label: string
      type?: 'text'
    }

interface CurrentColumn {
  prop: string
  label: string
  width?: number
}

interface SymbolForm {
  id: number
  tenantId: number
  symbol: string
  displaySymbol: string
  marketType: number
  baseAsset: string
  quoteAsset: string
  settleAsset: string
  contractType: number
  status: number
  priceScale: number
  qtyScale: number
  minPrice: string
  maxPrice: string
  priceTick: string
  minQty: string
  maxQty: string
  qtyStep: string
  minNotional: string
  maxLeverage: number
  openTime: number
  closeTime: number
  sort: number
  remark: string
}

interface SpotForm {
  tenantId: number
  symbolId: number
  makerFeeRate: string
  takerFeeRate: string
  buyEnabled: number
  sellEnabled: number
}

interface ContractForm {
  tenantId: number
  symbolId: number
  contractSize: string
  multiplier: string
  maintenanceMarginRate: string
  initialMarginRate: string
  makerFeeRate: string
  takerFeeRate: string
  fundingIntervalMinutes: number
  deliveryTime: number
  supportCross: number
  supportIsolated: number
  buyEnabled: number
  sellEnabled: number
}

const loading = ref(false)
const submitLoading = ref(false)
const rows = ref<TradeSymbol[]>([])
const detailVisible = ref(false)
const detailData = ref<TradeSymbol | null>(null)
const symbolVisible = ref(false)
const spotVisible = ref(false)
const contractVisible = ref(false)

const currentQuery = reactive<CurrentQuery>({
  tenantId: undefined,
  marketType: undefined,
  keyword: '',
  status: undefined,
  limit: 100,
})

const currentFields: CurrentField[] = [
  { key: 'tenantId', label: t('trade.tenantId'), type: 'number' },
  { key: 'marketType', label: t('trade.marketType'), type: 'number' },
  { key: 'status', label: t('trade.status'), type: 'number' },
  { key: 'keyword', label: t('common.keyword'), type: 'text' },
]

const currentColumns: CurrentColumn[] = [
  { prop: 'id', label: 'ID', width: 80 },
  { prop: 'symbol', label: t('trade.symbol'), width: 160 },
  { prop: 'displaySymbol', label: t('trade.displaySymbol'), width: 160 },
  { prop: 'marketType', label: t('trade.marketType'), width: 100 },
  { prop: 'status', label: t('trade.status'), width: 100 },
  { prop: 'maxLeverage', label: t('trade.maxLeverage'), width: 100 },
]

const getDefaultSymbolForm = (): SymbolForm => ({
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

const getDefaultSpotForm = (): SpotForm => ({
  tenantId: 0,
  symbolId: 0,
  makerFeeRate: '',
  takerFeeRate: '',
  buyEnabled: 1,
  sellEnabled: 1,
})

const getDefaultContractForm = (): ContractForm => ({
  tenantId: 0,
  symbolId: 0,
  contractSize: '',
  multiplier: '',
  maintenanceMarginRate: '',
  initialMarginRate: '',
  makerFeeRate: '',
  takerFeeRate: '',
  fundingIntervalMinutes: 0,
  deliveryTime: 0,
  supportCross: 1,
  supportIsolated: 1,
  buyEnabled: 1,
  sellEnabled: 1,
})

const symbolForm = reactive<SymbolForm>(getDefaultSymbolForm())
const spotForm = reactive<SpotForm>(getDefaultSpotForm())
const contractForm = reactive<ContractForm>(getDefaultContractForm())

const pickList = (res: any) => res?.data || res?.list || []

const loadCurrent = async () => {
  loading.value = true
  try {
    rows.value = pickList(await tradeService.listSymbols(currentQuery))
  } finally {
    loading.value = false
  }
}

const resetCurrent = () => {
  currentQuery.tenantId = undefined
  currentQuery.marketType = undefined
  currentQuery.keyword = ''
  currentQuery.status = undefined
  currentQuery.limit = 100
  loadCurrent()
}

const showDetail = async (row: TradeSymbol) => {
  detailData.value =
    (await tradeService.getSymbol({ tenantId: row.tenantId, id: row.id })).data || row
  detailVisible.value = true
}

const openSymbolDialog = (row?: Partial<SymbolForm>) => {
  Object.assign(symbolForm, getDefaultSymbolForm(), row || {})
  symbolVisible.value = true
}

const submitSymbol = async () => {
  submitLoading.value = true
  try {
    if (symbolForm.id) {
      await tradeService.updateSymbol(symbolForm)
    } else {
      await tradeService.createSymbol(symbolForm)
    }
    ElMessage.success(t('trade.saveSuccess'))
    symbolVisible.value = false
    loadCurrent()
  } finally {
    submitLoading.value = false
  }
}

const openSpotDialog = (row: TradeSymbol) => {
  Object.assign(spotForm, getDefaultSpotForm(), {
    tenantId: row.tenantId || 0,
    symbolId: row.id || 0,
  })
  spotVisible.value = true
}

const submitSpotConfig = async () => {
  submitLoading.value = true
  try {
    await tradeService.setSpotConfig(spotForm)
    ElMessage.success(t('trade.saveSuccessSpotConfig'))
    spotVisible.value = false
  } finally {
    submitLoading.value = false
  }
}

const openContractDialog = (row: TradeSymbol) => {
  Object.assign(contractForm, getDefaultContractForm(), {
    tenantId: row.tenantId || 0,
    symbolId: row.id || 0,
  })
  contractVisible.value = true
}

const submitContractConfig = async () => {
  submitLoading.value = true
  try {
    await tradeService.setContractConfig(contractForm)
    ElMessage.success(t('trade.saveSuccessContractConfig'))
    contractVisible.value = false
  } finally {
    submitLoading.value = false
  }
}

onMounted(loadCurrent)
</script>

<style scoped>
.detail-pre {
  white-space: pre-wrap;
  word-break: break-all;
}
</style>
