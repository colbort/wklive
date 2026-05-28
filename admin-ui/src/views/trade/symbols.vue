<template>
  <div class="module-page">
    <div class="page-header">
      <h2>{{ t('trade.symbols') }}</h2>
      <div class="header-actions">
        <el-button @click="loadCurrent">
          <el-icon><Refresh /></el-icon>
          {{ t('common.refresh') }}
        </el-button>
        <el-button type="primary" @click="openSymbolDialog()">
          <el-icon><Plus /></el-icon>
          {{ t('trade.addSymbol') }}
        </el-button>
      </div>
    </div>

    <el-card shadow="never" class="query-card">
      <el-form
        :model="query"
        inline
        label-width="90px"
        class="query-form"
      >
        <el-form-item :label="t('trade.tenantId')">
          <div class="query-field">
            <TenantSelect v-model="query.tenantId" include-system />
          </div>
        </el-form-item>

        <el-form-item :label="t('trade.marketType')">
          <el-select v-model="query.marketType" clearable class="query-field">
            <el-option
              v-for="item in marketTypeOptions"
              :key="item.value"
              :label="optionItemLabel(item)"
              :value="item.value"
            />
          </el-select>
        </el-form-item>

        <el-form-item :label="t('trade.status')">
          <el-select v-model="query.status" clearable class="query-field">
            <el-option
              v-for="item in symbolStatusOptions"
              :key="item.value"
              :label="optionItemLabel(item)"
              :value="item.value"
            />
          </el-select>
        </el-form-item>

        <el-form-item :label="t('common.keyword')">
          <el-input v-model="query.keyword" clearable class="query-keyword" />
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
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="tenantId" :label="t('trade.tenantId')" width="100" />

        <el-table-column :label="t('trade.symbol')" min-width="170" show-overflow-tooltip>
          <template #default="{ row }">
            <div class="symbol-cell">
              <span class="symbol-code">{{ row.symbol || '-' }}</span>
              <span class="muted">{{ row.displaySymbol || '-' }}</span>
            </div>
          </template>
        </el-table-column>

        <el-table-column :label="t('trade.marketType')" min-width="130">
          <template #default="{ row }">
            <el-tag size="small" effect="light">
              {{ optionLabel('marketType', row.marketType) }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column :label="t('trade.contractType')" min-width="130">
          <template #default="{ row }">
            {{ optionLabel('contractType', row.contractType) }}
          </template>
        </el-table-column>

        <el-table-column :label="t('trade.status')" width="120">
          <template #default="{ row }">
            <el-tag size="small" :type="symbolStatusTagType(row.status)" effect="light">
              {{ optionLabel('symbolStatus', row.status) }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column :label="t('trade.baseAsset')" min-width="160" show-overflow-tooltip>
          <template #default="{ row }">
            <div class="asset-pair">
              <el-tag size="small">
                {{ row.baseAsset || '-' }}
              </el-tag>
              <span>/</span>
              <el-tag size="small" type="info">
                {{ row.quoteAsset || '-' }}
              </el-tag>
              <span v-if="row.settleAsset" class="muted">{{ row.settleAsset }}</span>
            </div>
          </template>
        </el-table-column>

        <el-table-column :label="t('trade.priceTick')" min-width="120">
          <template #default="{ row }">
            {{ row.priceTick || '-' }}
          </template>
        </el-table-column>

        <el-table-column :label="t('trade.qtyStep')" min-width="120">
          <template #default="{ row }">
            {{ row.qtyStep || '-' }}
          </template>
        </el-table-column>

        <el-table-column :label="t('trade.maxLeverage')" width="110">
          <template #default="{ row }">
            {{ row.maxLeverage || '-' }}
          </template>
        </el-table-column>

        <el-table-column :label="t('common.sort')" width="90">
          <template #default="{ row }">
            {{ row.sort || 0 }}
          </template>
        </el-table-column>

        <el-table-column :label="t('common.actions')" width="260" fixed="right">
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
      v-model="symbolVisible"
      :title="symbolForm.id ? t('trade.editSymbol') : t('trade.addSymbol')"
      width="920px"
    >
      <el-form label-width="108px" class="dialog-form">
        <div class="form-grid">
          <el-form-item :label="t('trade.tenantId')">
            <TenantSelect v-model="symbolForm.tenantId" include-system />
          </el-form-item>

          <el-form-item :label="t('trade.symbol')">
            <el-input v-model="symbolForm.symbol" :disabled="Boolean(symbolForm.id)" />
          </el-form-item>

          <el-form-item :label="t('trade.displaySymbol')">
            <el-input v-model="symbolForm.displaySymbol" />
          </el-form-item>

          <el-form-item :label="t('trade.marketType')">
            <el-select
              v-model="symbolForm.marketType"
              class="full-width"
              :disabled="Boolean(symbolForm.id)"
            >
              <el-option
                v-for="item in marketTypeFormOptions"
                :key="item.value"
                :label="optionItemLabel(item)"
                :value="item.value"
              />
            </el-select>
          </el-form-item>

          <el-form-item :label="t('trade.contractType')">
            <el-select
              v-model="symbolForm.contractType"
              class="full-width"
              :disabled="Boolean(symbolForm.id)"
            >
              <el-option
                v-for="item in contractTypeFormOptions"
                :key="item.value"
                :label="optionItemLabel(item)"
                :value="item.value"
              />
            </el-select>
          </el-form-item>

          <el-form-item :label="t('trade.status')">
            <el-select v-model="symbolForm.status" class="full-width">
              <el-option
                v-for="item in symbolStatusFormOptions"
                :key="item.value"
                :label="optionItemLabel(item)"
                :value="item.value"
              />
            </el-select>
          </el-form-item>

          <el-form-item :label="t('trade.baseAsset')">
            <el-input v-model="symbolForm.baseAsset" :disabled="Boolean(symbolForm.id)" />
          </el-form-item>

          <el-form-item :label="t('trade.quoteAsset')">
            <el-input v-model="symbolForm.quoteAsset" :disabled="Boolean(symbolForm.id)" />
          </el-form-item>

          <el-form-item :label="t('trade.settleAsset')">
            <el-input v-model="symbolForm.settleAsset" :disabled="Boolean(symbolForm.id)" />
          </el-form-item>

          <el-form-item :label="t('trade.priceScale')">
            <el-input-number
              v-model="symbolForm.priceScale"
              :min="0"
              :precision="0"
              class="full-width"
            />
          </el-form-item>

          <el-form-item :label="t('trade.qtyScale')">
            <el-input-number
              v-model="symbolForm.qtyScale"
              :min="0"
              :precision="0"
              class="full-width"
            />
          </el-form-item>

          <el-form-item :label="t('trade.maxLeverage')">
            <el-input-number
              v-model="symbolForm.maxLeverage"
              :min="0"
              :precision="0"
              class="full-width"
            />
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

          <el-form-item :label="t('trade.openTime')">
            <el-date-picker
              v-model="symbolOpenTime"
              type="datetime"
              clearable
              class="full-width"
            />
          </el-form-item>

          <el-form-item :label="t('trade.closeTime')">
            <el-date-picker
              v-model="symbolCloseTime"
              type="datetime"
              clearable
              class="full-width"
            />
          </el-form-item>

          <el-form-item :label="t('common.sort')">
            <el-input-number
              v-model="symbolForm.sort"
              :min="0"
              :precision="0"
              class="full-width"
            />
          </el-form-item>

          <el-form-item :label="t('common.remark')" class="wide">
            <el-input v-model="symbolForm.remark" type="textarea" :rows="3" />
          </el-form-item>
        </div>
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

    <el-dialog v-model="spotVisible" :title="t('trade.spotConfig')" width="620px">
      <el-form label-width="116px" class="dialog-form">
        <div class="form-grid two">
          <el-form-item :label="t('trade.tenantId')">
            <TenantSelect v-model="spotForm.tenantId" include-system />
          </el-form-item>

          <el-form-item :label="t('trade.symbolId')">
            <el-input-number
              v-model="spotForm.symbolId"
              :min="0"
              :precision="0"
              class="full-width"
            />
          </el-form-item>

          <el-form-item :label="t('trade.makerFeeRate')">
            <el-input v-model="spotForm.makerFeeRate" />
          </el-form-item>

          <el-form-item :label="t('trade.takerFeeRate')">
            <el-input v-model="spotForm.takerFeeRate" />
          </el-form-item>

          <el-form-item :label="t('trade.buyEnabled')">
            <el-select v-model="spotForm.buyEnabled" class="full-width">
              <el-option
                v-for="item in enableStatusOptions"
                :key="item.value"
                :label="optionItemLabel(item)"
                :value="item.value"
              />
            </el-select>
          </el-form-item>

          <el-form-item :label="t('trade.sellEnabled')">
            <el-select v-model="spotForm.sellEnabled" class="full-width">
              <el-option
                v-for="item in enableStatusOptions"
                :key="item.value"
                :label="optionItemLabel(item)"
                :value="item.value"
              />
            </el-select>
          </el-form-item>
        </div>
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

    <el-dialog v-model="contractVisible" :title="t('trade.contractConfig')" width="760px">
      <el-form label-width="126px" class="dialog-form">
        <div class="form-grid two">
          <el-form-item :label="t('trade.tenantId')">
            <TenantSelect v-model="contractForm.tenantId" include-system />
          </el-form-item>

          <el-form-item :label="t('trade.symbolId')">
            <el-input-number
              v-model="contractForm.symbolId"
              :min="0"
              :precision="0"
              class="full-width"
            />
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
            <el-input-number
              v-model="contractForm.fundingIntervalMinutes"
              :min="0"
              :precision="0"
              class="full-width"
            />
          </el-form-item>

          <el-form-item :label="t('option.deliverTime')">
            <el-date-picker
              v-model="contractDeliveryTime"
              type="datetime"
              clearable
              class="full-width"
            />
          </el-form-item>

          <el-form-item :label="t('trade.supportCross')">
            <el-select v-model="contractForm.supportCross" class="full-width">
              <el-option
                v-for="item in enableStatusOptions"
                :key="item.value"
                :label="optionItemLabel(item)"
                :value="item.value"
              />
            </el-select>
          </el-form-item>

          <el-form-item :label="t('trade.supportIsolated')">
            <el-select v-model="contractForm.supportIsolated" class="full-width">
              <el-option
                v-for="item in enableStatusOptions"
                :key="item.value"
                :label="optionItemLabel(item)"
                :value="item.value"
              />
            </el-select>
          </el-form-item>

          <el-form-item :label="t('trade.buyEnabled')">
            <el-select v-model="contractForm.buyEnabled" class="full-width">
              <el-option
                v-for="item in enableStatusOptions"
                :key="item.value"
                :label="optionItemLabel(item)"
                :value="item.value"
              />
            </el-select>
          </el-form-item>

          <el-form-item :label="t('trade.sellEnabled')">
            <el-select v-model="contractForm.sellEnabled" class="full-width">
              <el-option
                v-for="item in enableStatusOptions"
                :key="item.value"
                :label="optionItemLabel(item)"
                :value="item.value"
              />
            </el-select>
          </el-form-item>
        </div>
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

    <el-dialog v-model="detailVisible" :title="t('option.detail')" width="860px">
      <el-descriptions v-if="detailData" :column="3" border>
        <el-descriptions-item label="ID">
          {{ detailData.id }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('trade.tenantId')">
          {{ detailData.tenantId }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('trade.status')">
          <el-tag size="small" :type="symbolStatusTagType(detailData.status)" effect="light">
            {{ optionLabel('symbolStatus', detailData.status) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item :label="t('trade.symbol')">
          {{ detailData.symbol || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('trade.displaySymbol')">
          {{ detailData.displaySymbol || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('trade.marketType')">
          {{ optionLabel('marketType', detailData.marketType) }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('trade.baseAsset')">
          {{ detailData.baseAsset || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('trade.quoteAsset')">
          {{ detailData.quoteAsset || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('trade.settleAsset')">
          {{ detailData.settleAsset || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('trade.contractType')">
          {{ optionLabel('contractType', detailData.contractType) }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('trade.priceScale')">
          {{ detailData.priceScale }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('trade.qtyScale')">
          {{ detailData.qtyScale }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('trade.minPrice')">
          {{ detailData.minPrice || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('trade.maxPrice')">
          {{ detailData.maxPrice || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('trade.priceTick')">
          {{ detailData.priceTick || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('trade.minQty')">
          {{ detailData.minQty || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('trade.maxQty')">
          {{ detailData.maxQty || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('trade.qtyStep')">
          {{ detailData.qtyStep || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('trade.minNotional')">
          {{ detailData.minNotional || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('trade.maxLeverage')">
          {{ detailData.maxLeverage || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('common.sort')">
          {{ detailData.sort || 0 }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('trade.openTime')">
          {{ formatDate(detailData.openTime || 0) }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('trade.closeTime')">
          {{ formatDate(detailData.closeTime || 0) }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('common.createTimes')">
          {{ formatDate(detailData.createTimes || 0) }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('common.updateTimes')">
          {{ formatDate(detailData.updateTimes || 0) }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('common.remark')" :span="2">
          {{ detailData.remark || '-' }}
        </el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus, Refresh } from '@element-plus/icons-vue'
import { useI18n } from 'vue-i18n'
import { usePagination } from '@/composables'
import { tradeService, type OptionGroup, type OptionItem, type TradeSymbol } from '@/services'
import { formatDate } from '@/utils'
import { findOptionGroup, getOptionLabel } from '@/utils/options'
import TenantSelect from '@/components/TenantSelect.vue'

type TagType = '' | 'success' | 'warning' | 'info' | 'danger'
type DatePickerValue = Date | string | number | null | undefined

interface SymbolQuery {
  tenantId: number | undefined
  marketType: number | undefined
  keyword: string
  status: number | undefined
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

const { t } = useI18n()
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)

const loading = ref(false)
const submitLoading = ref(false)
const rows = ref<TradeSymbol[]>([])
const detailVisible = ref(false)
const detailData = ref<TradeSymbol | null>(null)
const symbolVisible = ref(false)
const spotVisible = ref(false)
const contractVisible = ref(false)
const optionGroups = ref<OptionGroup[]>([])

const marketTypeFallbackOptions: OptionItem[] = [
  { value: 0, code: 'MARKET_TYPE_UNKNOWN' },
  { value: 1, code: 'MARKET_TYPE_SPOT' },
  { value: 2, code: 'MARKET_TYPE_SECONDS_CONTRACT' },
  { value: 3, code: 'MARKET_TYPE_USDT_CONTRACT' },
  { value: 4, code: 'MARKET_TYPE_COIN_CONTRACT' },
]

const contractTypeFallbackOptions: OptionItem[] = [
  { value: 0, code: 'CONTRACT_TYPE_UNKNOWN' },
  { value: 1, code: 'CONTRACT_TYPE_NONE' },
  { value: 2, code: 'CONTRACT_TYPE_PERPETUAL' },
  { value: 3, code: 'CONTRACT_TYPE_DELIVERY' },
  { value: 4, code: 'CONTRACT_TYPE_SECONDS' },
]

const symbolStatusFallbackOptions: OptionItem[] = [
  { value: 0, code: 'SYMBOL_STATUS_UNKNOWN' },
  { value: 1, code: 'SYMBOL_STATUS_ENABLED' },
  { value: 2, code: 'SYMBOL_STATUS_DISABLED' },
  { value: 3, code: 'SYMBOL_STATUS_CLOSE_ONLY' },
]

const enableStatusFallbackOptions: OptionItem[] = [
  { value: 0, code: 'ENABLE_STATUS_DISABLED' },
  { value: 1, code: 'ENABLE_STATUS_ENABLED' },
]

const query = reactive<SymbolQuery>({
  tenantId: undefined,
  marketType: undefined,
  keyword: '',
  status: undefined,
})

const getDefaultSymbolForm = (): SymbolForm => ({
  id: 0,
  tenantId: 0,
  symbol: '',
  displaySymbol: '',
  marketType: 1,
  baseAsset: '',
  quoteAsset: '',
  settleAsset: '',
  contractType: 1,
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

const optionGroupWithFallback = (key: string, fallback: OptionItem[]) =>
  computed(() => {
    const options = findOptionGroup(optionGroups.value, key)
    return options.length ? options : fallback
  })

const withoutUnknown = (options: OptionItem[]) => options.filter((item) => item.value !== 0)

const marketTypeOptions = optionGroupWithFallback('marketType', marketTypeFallbackOptions)
const contractTypeOptions = optionGroupWithFallback('contractType', contractTypeFallbackOptions)
const symbolStatusOptions = optionGroupWithFallback('symbolStatus', symbolStatusFallbackOptions)
const enableStatusOptions = optionGroupWithFallback('enableStatus', enableStatusFallbackOptions)
const marketTypeFormOptions = computed(() => withoutUnknown(marketTypeOptions.value))
const contractTypeFormOptions = computed(() => withoutUnknown(contractTypeOptions.value))
const symbolStatusFormOptions = computed(() => withoutUnknown(symbolStatusOptions.value))

const timestampToDate = (timestamp?: number) => {
  if (!timestamp) return null
  return new Date(timestamp < 1e12 ? timestamp * 1000 : timestamp)
}

const dateToUnixSeconds = (value: DatePickerValue) => {
  if (!value) return 0
  const time =
    typeof value === 'number'
      ? value < 1e12
        ? value * 1000
        : value
      : value instanceof Date
        ? value.getTime()
        : new Date(value).getTime()
  return Number.isNaN(time) ? 0 : Math.floor(time / 1000)
}

const symbolOpenTime = computed({
  get: () => timestampToDate(symbolForm.openTime),
  set: (value: DatePickerValue) => {
    symbolForm.openTime = dateToUnixSeconds(value)
  },
})

const symbolCloseTime = computed({
  get: () => timestampToDate(symbolForm.closeTime),
  set: (value: DatePickerValue) => {
    symbolForm.closeTime = dateToUnixSeconds(value)
  },
})

const contractDeliveryTime = computed({
  get: () => timestampToDate(contractForm.deliveryTime),
  set: (value: DatePickerValue) => {
    contractForm.deliveryTime = dateToUnixSeconds(value)
  },
})

const optionItemLabel = (item: OptionItem) => getOptionLabel(t, item.code, item.value)

const optionLabel = (key: string, value?: number | string) => {
  const fallbackMap: Record<string, OptionItem[]> = {
    marketType: marketTypeFallbackOptions,
    contractType: contractTypeFallbackOptions,
    symbolStatus: symbolStatusFallbackOptions,
    enableStatus: enableStatusFallbackOptions,
  }
  const option =
    findOptionGroup(optionGroups.value, key).find((item) => String(item.value) === String(value)) ||
    fallbackMap[key]?.find((item) => String(item.value) === String(value))
  return option ? optionItemLabel(option) : '-'
}

const symbolStatusTagType = (status?: number): TagType => {
  switch (status) {
    case 1:
      return 'success'
    case 2:
      return 'danger'
    case 3:
      return 'warning'
    default:
      return 'info'
  }
}

const loadOptions = async () => {
  try {
    optionGroups.value = (await tradeService.getOptions()).data || []
  } catch (error) {
    console.error('load trade options failed', error)
  }
}

const loadCurrent = async () => {
  loading.value = true
  try {
    const res = await tradeService.listSymbols({
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

const resetCurrent = () => {
  query.tenantId = undefined
  query.marketType = undefined
  query.keyword = ''
  query.status = undefined
  resetAndLoad(loadCurrent)
}

const showDetail = async (row: TradeSymbol) => {
  detailData.value =
    (await tradeService.getSymbol({ tenantId: row.tenantId, id: row.id })).data || row
  detailVisible.value = true
}

const openSymbolDialog = (row?: TradeSymbol) => {
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

function handleLimitChange() {
  resetAndLoad(loadCurrent)
}

function handlePrevPage() {
  prevAndLoad(loadCurrent)
}

function handleNextPage() {
  nextAndLoad(loadCurrent)
}

onMounted(() => {
  loadOptions()
  loadCurrent()
})
</script>

<style scoped>
.query-form :deep(.el-form-item) {
  margin-bottom: 12px;
}

.query-field {
  width: 220px;
}

.query-keyword {
  width: 220px;
}

.symbol-cell {
  display: flex;
  flex-direction: column;
  gap: 2px;
  line-height: 1.35;
}

.symbol-code {
  font-weight: 600;
}

.muted {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.asset-pair {
  display: flex;
  align-items: center;
  gap: 6px;
  min-width: 0;
}

.dialog-form :deep(.el-form-item) {
  margin-bottom: 14px;
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  column-gap: 16px;
}

.form-grid.two {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.wide {
  grid-column: 1 / -1;
}

.full-width {
  width: 100%;
}

@media (max-width: 768px) {
  .query-field,
  .query-keyword {
    width: 100%;
  }

  .form-grid,
  .form-grid.two {
    grid-template-columns: 1fr;
  }
}
</style>
