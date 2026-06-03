<template>
  <div class="module-page">
    <div class="page-header">
      <h2>{{ t('option.contracts') }}</h2>
      <div class="header-actions">
        <el-button @click="loadCurrent">
          {{ t('common.refresh') }}
        </el-button>
        <el-button v-perm="'option:contract:add'" type="primary" @click="openContractDialog()">
          {{ t('option.createContract') }}
        </el-button>
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
              v-for="item in contractStatusOptions"
              :key="item.value"
              :label="getOptionLabel(t, item.code, item.value)"
              :value="item.value"
            />
          </el-select>
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
        <el-table-column label="ID" width="90">
          <template #default="{ row }">
            {{ row.contract?.id ?? '-' }}
          </template>
        </el-table-column>
        <el-table-column :label="t('option.tenantId')" width="100">
          <template #default="{ row }">
            {{ row.contract?.tenantId ?? '-' }}
          </template>
        </el-table-column>
        <el-table-column :label="t('option.contractCode')" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.contract?.contractCode || '-' }}
          </template>
        </el-table-column>
        <el-table-column :label="t('option.underlying')" min-width="140" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.contract?.underlyingSymbol || '-' }}
          </template>
        </el-table-column>
        <el-table-column :label="t('option.optionType')" width="110">
          <template #default="{ row }">
            {{ optionLabel('optionType', row.contract?.optionType) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('option.exerciseStyle')" width="110">
          <template #default="{ row }">
            {{ optionLabel('exerciseStyle', row.contract?.exerciseStyle) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('option.settlementType')" width="120">
          <template #default="{ row }">
            {{ optionLabel('settlementType', row.contract?.settlementType) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('option.settleCoin')" width="100">
          <template #default="{ row }">
            {{ row.contract?.settleCoin || '-' }}
          </template>
        </el-table-column>
        <el-table-column :label="t('option.quoteCoin')" width="100">
          <template #default="{ row }">
            {{ row.contract?.quoteCoin || '-' }}
          </template>
        </el-table-column>
        <el-table-column :label="t('option.strikePrice')" min-width="120" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.contract?.strikePrice || '-' }}
          </template>
        </el-table-column>
        <el-table-column :label="t('option.markPrice')" min-width="120" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.market?.markPrice || '-' }}
          </template>
        </el-table-column>
        <el-table-column :label="t('option.expireTime')" min-width="170">
          <template #default="{ row }">
            {{ formatDate(row.contract?.expireTime || 0) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('common.status')" width="120">
          <template #default="{ row }">
            <el-tag :type="contractStatusTagType(row.contract?.status)" disable-transitions>
              {{ optionLabel('contractStatus', row.contract?.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="t('common.actions')" width="220" fixed="right">
          <template #default="{ row }">
            <el-button
              v-perm="'option:contract:detail'"
              link
              type="primary"
              @click="showDetail(row)"
            >
              {{ t('option.detail') }}
            </el-button>
            <el-button
              v-perm="'option:contract:update'"
              link
              type="primary"
              @click="openContractDialog(row)"
            >
              {{ t('option.editContract') }}
            </el-button>
            <el-button
              v-perm="'option:market:update'"
              link
              type="primary"
              @click="openMarketDialog(row)"
            >
              {{ t('option.editMarket') }}
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
      v-model="contractVisible"
      :title="contractForm.id ? t('option.editContract') : t('option.createContract')"
      width="960px"
    >
      <el-form class="compact-contract-form" label-width="92px">
        <div class="contract-form-grid">
          <el-form-item :label="t('option.tenantId')">
            <TenantSelect v-model="contractForm.tenantId" include-system />
          </el-form-item>
          <el-form-item :label="t('option.contractCode')" class="span-2">
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
            <el-select v-model="contractForm.settlementType" style="width: 100%">
              <el-option
                v-for="item in settlementTypeOptions"
                :key="item.value"
                :label="getOptionLabel(t, item.code, item.value)"
                :value="item.value"
              />
            </el-select>
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
            <el-date-picker
              v-model="contractListTime"
              type="datetime"
              :placeholder="t('common.pleaseSelect')"
              format="YYYY-MM-DD HH:mm:ss"
              clearable
              style="width: 100%"
            />
          </el-form-item>
          <el-form-item :label="t('option.expireTime')">
            <el-date-picker
              v-model="contractExpireTime"
              type="datetime"
              :placeholder="t('common.pleaseSelect')"
              format="YYYY-MM-DD HH:mm:ss"
              clearable
              style="width: 100%"
            />
          </el-form-item>
          <el-form-item :label="t('option.deliverTime')">
            <el-date-picker
              v-model="contractDeliverTime"
              type="datetime"
              :placeholder="t('common.pleaseSelect')"
              format="YYYY-MM-DD HH:mm:ss"
              clearable
              style="width: 100%"
            />
          </el-form-item>
          <el-form-item :label="t('option.autoExercise')">
            <el-select v-model="contractForm.isAutoExercise" style="width: 100%">
              <el-option
                v-for="item in yesNoOptions"
                :key="item.value"
                :label="getOptionLabel(t, item.code, item.value)"
                :value="item.value"
              />
            </el-select>
          </el-form-item>
          <el-form-item :label="t('common.status')">
            <el-select v-model="contractForm.status" style="width: 100%">
              <el-option
                v-for="item in contractStatusOptions"
                :key="item.value"
                :label="getOptionLabel(t, item.code, item.value)"
                :value="item.value"
              />
            </el-select>
          </el-form-item>
          <el-form-item :label="t('common.sort')">
            <el-input-number v-model="contractForm.sort" :min="0" :precision="0" />
          </el-form-item>
          <el-form-item :label="t('common.remark')" class="span-3">
            <el-input v-model="contractForm.remark" type="textarea" :rows="2" />
          </el-form-item>
        </div>
      </el-form>
      <template #footer>
        <el-button @click="contractVisible = false">
          {{ t('common.cancel') }}
        </el-button>
        <el-button
          v-perm="contractForm.id ? 'option:contract:update' : 'option:contract:add'"
          type="primary"
          :loading="submitLoading"
          @click="submitContract"
        >
          {{ t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="marketVisible" :title="t('option.editMarket')" width="920px">
      <el-form class="compact-market-form" label-width="92px">
        <div class="market-form-grid">
          <el-form-item :label="t('option.tenantId')">
            <TenantSelect v-model="marketForm.tenantId" include-system />
          </el-form-item>
          <el-form-item :label="t('option.contractId')">
            <el-input-number v-model="marketForm.contractId" :min="0" :precision="0" />
          </el-form-item>
          <el-form-item :label="t('option.snapshotTime')">
            <el-date-picker
              v-model="marketSnapshotTime"
              type="datetime"
              :placeholder="t('common.pleaseSelect')"
              format="YYYY-MM-DD HH:mm:ss"
              clearable
              style="width: 100%"
            />
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
        </div>
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

    <el-dialog v-model="detailVisible" :title="t('option.detail')" width="900px">
      <div v-loading="detailLoading" class="contract-detail">
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
          <el-descriptions-item :label="t('option.optionType')">
            {{ optionLabel('optionType', detailData.contract.optionType) }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.exerciseStyle')">
            {{ optionLabel('exerciseStyle', detailData.contract.exerciseStyle) }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.settlementType')">
            {{ optionLabel('settlementType', detailData.contract.settlementType) }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.autoExercise')">
            {{ optionLabel('yesNo', detailData.contract.isAutoExercise) }}
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
          <el-descriptions-item :label="t('option.contractUnit')">
            {{ detailData.contract.contractUnit || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.minOrderQty')">
            {{ detailData.contract.minOrderQty || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.maxOrderQty')">
            {{ detailData.contract.maxOrderQty || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.priceTick')">
            {{ detailData.contract.priceTick || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.qtyStep')">
            {{ detailData.contract.qtyStep || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.multiplier')">
            {{ detailData.contract.multiplier || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('common.status')">
            <el-tag :type="contractStatusTagType(detailData.contract.status)" disable-transitions>
              {{ optionLabel('contractStatus', detailData.contract.status) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.listTime')">
            {{ formatDate(detailData.contract.listTime || 0) }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.expireTime')">
            {{ formatDate(detailData.contract.expireTime || 0) }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.deliverTime')">
            {{ formatDate(detailData.contract.deliverTime || 0) }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('common.remark')">
            {{ detailData.contract.remark || '-' }}
          </el-descriptions-item>
        </el-descriptions>

        <el-descriptions :title="t('option.marketInfo')" :column="2" border>
          <el-descriptions-item :label="t('option.underlyingPrice')">
            {{ detailData?.market?.underlyingPrice || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.markPrice')">
            {{ detailData?.market?.markPrice || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.lastPrice')">
            {{ detailData?.market?.lastPrice || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.bidPrice')">
            {{ detailData?.market?.bidPrice || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.askPrice')">
            {{ detailData?.market?.askPrice || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.theoreticalPrice')">
            {{ detailData?.market?.theoreticalPrice || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.intrinsicValue')">
            {{ detailData?.market?.intrinsicValue || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.timeValue')">
            {{ detailData?.market?.timeValue || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.iv')">
            {{ detailData?.market?.iv || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.delta')">
            {{ detailData?.market?.delta || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.gamma')">
            {{ detailData?.market?.gamma || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.theta')">
            {{ detailData?.market?.theta || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.vega')">
            {{ detailData?.market?.vega || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.rho')">
            {{ detailData?.market?.rho || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.riskFreeRate')">
            {{ detailData?.market?.riskFreeRate || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.pricingModel')">
            {{ detailData?.market?.pricingModel || '-' }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.snapshotTime')">
            {{ formatDate(detailData?.market?.snapshotTime || 0) }}
          </el-descriptions-item>
        </el-descriptions>
      </div>
      <template #footer>
        <el-button @click="detailVisible = false">
          {{ t('common.close') }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { usePagination } from '@/composables'
import { ElMessage } from 'element-plus'
import {
  optionService,
  type OptionContractDetail,
  type OptionGroup,
  type OptionItem,
  type UpdateContractReq,
  type UpdateMarketReq,
} from '@/services'
import { formatDate } from '@/utils'
import { findOptionGroup, getOptionLabel, getOptionValueLabel } from '@/utils/options'
import TenantSelect from '@/components/TenantSelect.vue'

const { t } = useI18n()
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)

const loading = ref(false)
const detailLoading = ref(false)
const submitLoading = ref(false)
const rows = ref<OptionContractDetail[]>([])
const detailVisible = ref(false)
const detailData = ref<OptionContractDetail | null>(null)
const contractVisible = ref(false)
const marketVisible = ref(false)
const optionGroups = ref<OptionGroup[]>([])

const query = reactive({
  tenantId: undefined as number | undefined,
  contractCode: '',
  underlyingSymbol: '',
  optionType: undefined as number | undefined,
  status: undefined as number | undefined,
  limit: 20,
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
const settlementTypeOptions = computed(() => findOptionGroup(optionGroups.value, 'settlementType'))
const yesNoFallbackOptions: OptionItem[] = [
  { value: 1, code: 'YES_NO_NO' },
  { value: 2, code: 'YES_NO_YES' },
]
const yesNoOptions = computed(() => {
  const options = findOptionGroup(optionGroups.value, 'yesNo').filter((item) => item.value !== 0)
  return options.length ? options : yesNoFallbackOptions
})
const contractStatusOptions = computed(() => findOptionGroup(optionGroups.value, 'contractStatus'))

const loadOptions = async () => {
  optionGroups.value = (await optionService.getOptions()).data || []
}

const firstBusinessOptionValue = (options: OptionItem[], fallback = 0) =>
  options.find((item) => item.value !== 0)?.value ?? options[0]?.value ?? fallback

const optionLabel = (key: string, value?: number | string) => {
  if (key === 'yesNo') {
    const option = yesNoOptions.value.find((item) => String(item.value) === String(value))
    return option ? getOptionLabel(t, option.code, option.value) : '-'
  }
  const label = getOptionValueLabel(optionGroups.value, key, value, t)
  return label === '' ? '-' : label
}

type DatePickerValue = Date | string | number | null | undefined
type TagType = '' | 'success' | 'warning' | 'info' | 'danger'

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

const contractListTime = computed({
  get: () => timestampToDate(contractForm.listTime),
  set: (value: DatePickerValue) => {
    contractForm.listTime = dateToUnixSeconds(value)
  },
})

const contractExpireTime = computed({
  get: () => timestampToDate(contractForm.expireTime),
  set: (value: DatePickerValue) => {
    contractForm.expireTime = dateToUnixSeconds(value)
  },
})

const contractDeliverTime = computed({
  get: () => timestampToDate(contractForm.deliverTime),
  set: (value: DatePickerValue) => {
    contractForm.deliverTime = dateToUnixSeconds(value)
  },
})

const marketSnapshotTime = computed({
  get: () => timestampToDate(marketForm.snapshotTime),
  set: (value: DatePickerValue) => {
    marketForm.snapshotTime = dateToUnixSeconds(value)
  },
})

const contractStatusTagType = (status?: number): TagType => {
  switch (status) {
    case 2:
    case 5:
      return 'success'
    case 1:
    case 3:
      return 'warning'
    case 4:
      return 'info'
    case 6:
      return 'danger'
    default:
      return ''
  }
}

const loadCurrent = async () => {
  loading.value = true
  try {
    const res = await optionService.listContracts({
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

const resetContractForm = () => {
  Object.assign(contractForm, {
    id: 0,
    tenantId: 0,
    contractCode: '',
    underlyingSymbol: '',
    settleCoin: '',
    quoteCoin: '',
    optionType: firstBusinessOptionValue(optionTypeOptions.value),
    exerciseStyle: firstBusinessOptionValue(exerciseStyleOptions.value),
    settlementType: firstBusinessOptionValue(settlementTypeOptions.value),
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
    isAutoExercise: firstBusinessOptionValue(yesNoOptions.value, 1),
    status: firstBusinessOptionValue(contractStatusOptions.value),
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

const showDetail = async (row: OptionContractDetail) => {
  detailVisible.value = true
  detailLoading.value = true
  detailData.value = row
  try {
    detailData.value =
      (
        await optionService.getContract({
          tenantId: row.contract.tenantId,
          id: row.contract.id,
          contractCode: row.contract.contractCode,
        })
      ).data || row
  } finally {
    detailLoading.value = false
  }
}

const openContractDialog = (row?: OptionContractDetail) => {
  resetContractForm()
  if (row) {
    Object.assign(contractForm, row.contract)
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

const openMarketDialog = (row?: OptionContractDetail | null) => {
  resetMarketForm()
  if (row?.contract) {
    Object.assign(marketForm, {
      tenantId: row.contract.tenantId || 0,
      contractId: row.contract.id || 0,
    })
    if (row.market) {
      Object.assign(marketForm, {
        underlyingPrice: row.market.underlyingPrice || '',
        markPrice: row.market.markPrice || '',
        lastPrice: row.market.lastPrice || '',
        bidPrice: row.market.bidPrice || '',
        askPrice: row.market.askPrice || '',
        theoreticalPrice: row.market.theoreticalPrice || '',
        intrinsicValue: row.market.intrinsicValue || '',
        timeValue: row.market.timeValue || '',
        iv: row.market.iv || '',
        delta: row.market.delta || '',
        gamma: row.market.gamma || '',
        theta: row.market.theta || '',
        vega: row.market.vega || '',
        rho: row.market.rho || '',
        riskFreeRate: row.market.riskFreeRate || '',
        pricingModel: row.market.pricingModel || '',
        snapshotTime: row.market.snapshotTime || 0,
      })
    }
  }
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

onMounted(async () => {
  await loadOptions()
  resetContractForm()
  await loadCurrent()
})
</script>

<style scoped>
.contract-detail {
  display: grid;
  gap: 18px;
}

.contract-form-grid,
.market-form-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  column-gap: 12px;
}

.compact-contract-form :deep(.el-form-item),
.compact-market-form :deep(.el-form-item) {
  margin-bottom: 10px;
}

.compact-contract-form :deep(.el-form-item__label),
.compact-market-form :deep(.el-form-item__label) {
  padding-right: 6px;
}

.compact-contract-form :deep(.el-input-number),
.compact-market-form :deep(.el-input-number) {
  width: 100%;
}

.span-2 {
  grid-column: span 2;
}

.span-3 {
  grid-column: 1 / -1;
}

@media (max-width: 900px) {
  .contract-form-grid,
  .market-form-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .span-2,
  .span-3 {
    grid-column: 1 / -1;
  }
}

@media (max-width: 640px) {
  .contract-form-grid,
  .market-form-grid {
    grid-template-columns: 1fr;
  }
}
</style>
