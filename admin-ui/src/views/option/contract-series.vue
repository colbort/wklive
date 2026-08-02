<template>
  <div class="module-page">
    <CrudQueryCard :model="query" @search="search" @reset="resetQuery">
      <el-form-item v-if="authStore.isSystemAdmin" :label="t('option.tenantId')">
        <TenantSelect v-model="query.tenantId" class="tenant-select-filter" />
      </el-form-item>
      <el-form-item :label="t('option.seriesCode')">
        <el-input v-model="query.seriesCode" clearable />
      </el-form-item>
      <el-form-item :label="t('common.status')">
        <el-select v-model="query.status" clearable style="width: 180px">
          <el-option
            v-for="item in statuses"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
      </el-form-item>
      <template #actions>
        <el-button v-perm="'option:contract-series:create'" type="primary" @click="openCreate">
          {{ t('option.createContractSeries') }}
        </el-button>
      </template>
    </CrudQueryCard>

    <el-card shadow="never" class="table-card">
      <el-alert
        :title="t('option.contractSeriesPendingOnlyNotice')"
        type="warning"
        :closable="false"
        class="table-alert"
      />
      <el-table v-loading="loading" :data="series" stripe>
        <el-table-column
          v-if="authStore.isSystemAdmin"
          prop="tenantId"
          :label="t('option.tenantId')"
          width="110"
        />
        <el-table-column prop="seriesCode" :label="t('option.seriesCode')" min-width="130" />
        <el-table-column prop="version" :label="t('option.version')" width="80" />
        <el-table-column
          prop="underlyingSymbol"
          :label="t('option.underlyingSymbol')"
          width="130"
        />
        <el-table-column :label="t('common.status')" width="130">
          <template #default="{ row }">
            {{ statusName(row.status) }}
          </template>
        </el-table-column>
        <el-table-column prop="referencePrice" :label="t('option.referencePrice')" width="125" />
        <el-table-column
          prop="referenceSource"
          :label="t('option.referenceSource')"
          min-width="150"
        />
        <el-table-column :label="t('option.expiryCount')" width="105">
          <template #default="{ row }">
            {{ row.expiries.length }}
          </template>
        </el-table-column>
        <el-table-column :label="t('option.generationProgress')" width="140">
          <template #default="{ row }">
            {{ row.generatedContractCount }}/{{ row.expectedContractCount }}
          </template>
        </el-table-column>
        <el-table-column :label="t('option.launchReviewStatus')" width="135">
          <template #default="{ row }">
            {{ launchStatusName(row.launchStatus) }}
          </template>
        </el-table-column>
        <el-table-column prop="evidenceRef" :label="t('option.evidenceRef')" min-width="180" />
        <el-table-column :label="t('common.actions')" width="230" fixed="right">
          <template #default="{ row }">
            <template v-if="row.status === 1">
              <el-button
                v-perm="'option:contract-series:review'"
                link
                type="success"
                @click="review(row, true)"
              >
                {{ t('option.approveAndGenerate') }}
              </el-button>
              <el-button
                v-perm="'option:contract-series:review'"
                link
                type="danger"
                @click="review(row, false)"
              >
                {{ t('option.reject') }}
              </el-button>
            </template>
            <el-button
              v-if="row.status === 2"
              v-perm="'option:contract-series:detail:list'"
              link
              @click="showDetails(row)"
            >
              {{ t('option.generatedContracts') }}
            </el-button>
            <template v-if="row.status === 2 && row.launchStatus === 1">
              <el-button
                v-perm="'option:contract-series:launch-review'"
                link
                type="success"
                @click="reviewLaunch(row, true)"
              >
                {{ t('option.approveLaunch') }}
              </el-button>
              <el-button
                v-perm="'option:contract-series:launch-review'"
                link
                type="danger"
                @click="reviewLaunch(row, false)"
              >
                {{ t('option.rejectLaunch') }}
              </el-button>
            </template>
          </template>
        </el-table-column>
      </el-table>
      <CursorPagination
        v-model:limit="pagination.pagination.limit"
        :total="pagination.pagination.total"
        :has-prev="pagination.pagination.hasPrev"
        :has-next="pagination.pagination.hasNext"
        @prev="pagination.prevAndLoad(refresh)"
        @next="pagination.nextAndLoad(refresh)"
        @limit-change="pagination.resetAndLoad(refresh)"
      />
    </el-card>

    <el-dialog v-model="createVisible" :title="t('option.createContractSeries')" width="860px">
      <el-form :model="form" label-width="170px">
        <el-form-item :label="t('option.seriesCode')">
          <el-input v-model="form.seriesCode" placeholder="BTCUSD" />
        </el-form-item>
        <el-form-item :label="t('option.parameterSourceContract')">
          <el-select
            v-model="form.sourceContractId"
            filterable
            style="width: 100%"
            :loading="contractLoading"
          >
            <el-option
              v-for="item in contracts"
              :key="item.contract.id"
              :label="`${item.contract.contractCode} · ${item.contract.underlyingSymbol}`"
              :value="item.contract.id"
            />
          </el-select>
        </el-form-item>
        <el-alert
          :title="t('option.contractSeriesTemplateNotice')"
          type="info"
          :closable="false"
          class="form-alert"
        />
        <el-form-item :label="t('option.referencePrice')">
          <el-input v-model="form.referencePrice" />
        </el-form-item>
        <el-form-item :label="t('option.referenceSource')">
          <el-input v-model="form.referenceSource" />
        </el-form-item>
        <el-form-item :label="t('option.referenceTime')">
          <el-date-picker v-model="form.referenceTime" type="datetime" />
        </el-form-item>
        <el-form-item :label="t('option.expirySpecifications')">
          <el-input
            v-model="form.expiryLines"
            type="textarea"
            :rows="5"
            :placeholder="t('option.expirySpecificationFormat')"
          />
        </el-form-item>
        <el-form-item :label="t('option.strikeBands')">
          <el-input
            v-model="form.bandLines"
            type="textarea"
            :rows="4"
            :placeholder="t('option.strikeBandFormat')"
          />
        </el-form-item>
        <el-form-item :label="t('option.expectedContractCount')">
          <el-tag>{{ expectedCount }}</el-tag>
          <span class="count-note">{{ t('option.callPutSymmetryNotice') }}</span>
        </el-form-item>
        <el-form-item :label="t('option.evidenceRef')">
          <el-input v-model="form.evidenceRef" />
        </el-form-item>
        <el-form-item :label="t('option.changeReason')">
          <el-input v-model="form.changeReason" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createVisible = false">
          {{ t('common.cancel') }}
        </el-button>
        <el-button type="primary" :loading="saving" @click="createSeries">
          {{ t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="detailVisible" :title="t('option.generatedContracts')" width="980px">
      <el-table v-loading="detailLoading" :data="details" stripe>
        <el-table-column prop="contractCode" :label="t('option.contractCode')" min-width="210" />
        <el-table-column :label="t('option.optionType')" width="100">
          <template #default="{ row }">
            {{ row.optionType === 1 ? 'CALL' : 'PUT' }}
          </template>
        </el-table-column>
        <el-table-column prop="strikePrice" :label="t('option.strikePrice')" width="140" />
        <el-table-column prop="expiryId" :label="t('option.expirySpecId')" width="120" />
        <el-table-column prop="contractId" :label="t('option.contractId')" width="120" />
      </el-table>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  optionService,
  type ContractSeriesExpiryInput,
  type ContractSeriesStrikeBandInput,
  type CreateContractReq,
  type OptionContractDetail,
  type OptionContractSeries,
  type OptionContractSeriesDetail,
} from '@/services'
import TenantSelect from '@/components/TenantSelect.vue'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'
import { useAuthStore } from '@/stores/auth'
import { usePagination } from '@/composables'

const { t } = useI18n()
const authStore = useAuthStore()
const loading = ref(false)
const saving = ref(false)
const contractLoading = ref(false)
const detailLoading = ref(false)
const createVisible = ref(false)
const detailVisible = ref(false)
const series = ref<OptionContractSeries[]>([])
const contracts = ref<OptionContractDetail[]>([])
const details = ref<OptionContractSeriesDetail[]>([])
const pagination = usePagination<number>(20)
const query = reactive({
  tenantId: (authStore.isTenantUser ? authStore.profileTenantId : undefined) as number | undefined,
  seriesCode: '',
  status: undefined as number | undefined,
})
const nowSeconds = () => Math.floor(Date.now() / 1000)
const defaultExpiryLine = () => {
  const now = nowSeconds()
  return `WEEKLY,${now + 3600},${now + 7 * 86400 - 7200},${now + 7 * 86400 - 3600},${now + 7 * 86400},${now + 7 * 86400 + 60}`
}
const form = reactive({
  seriesCode: '',
  sourceContractId: undefined as number | undefined,
  referencePrice: '',
  referenceSource: 'authoritative-index',
  referenceTime: new Date(),
  expiryLines: defaultExpiryLine(),
  bandLines: '80,120,10',
  evidenceRef: '',
  changeReason: '',
})

const statuses = computed(() => [
  { value: 1, label: t('option.contractSeriesPendingReview') },
  { value: 2, label: t('option.contractSeriesGenerated') },
  { value: 3, label: t('option.contractSeriesRejected') },
])
const expectedCount = computed(() => {
  try {
    const expiries = parseExpiries()
    const strikes = parseBands().reduce((count, band) => {
      const lower = Number(band.lowerStrike)
      const upper = Number(band.upperStrike)
      const step = Number(band.strikeStep)
      return count + Math.floor((upper - lower) / step) + 1
    }, 0)
    return expiries.length * strikes * 2
  } catch {
    return 0
  }
})

function resetQuery() {
  query.tenantId = authStore.isTenantUser ? authStore.profileTenantId || undefined : undefined
  query.seriesCode = ''
  query.status = undefined
  pagination.resetAndLoad(refresh)
}

function search() {
  pagination.resetAndLoad(refresh)
}

async function refresh() {
  loading.value = true
  try {
    const response = await optionService.listContractSeries({
      tenantId: query.tenantId,
      seriesCode: query.seriesCode || undefined,
      status: query.status,
      cursor: pagination.pagination.cursor,
      limit: pagination.pagination.limit,
    })
    series.value = response.data || []
    pagination.updateFromResponse(response)
  } finally {
    loading.value = false
  }
}

async function openCreate() {
  if (!query.tenantId) return
  form.referenceTime = new Date()
  form.expiryLines = defaultExpiryLine()
  form.sourceContractId = undefined
  createVisible.value = true
  contractLoading.value = true
  try {
    const response = await optionService.listContracts({
      tenantId: query.tenantId,
      limit: 100,
    })
    contracts.value = response.data || []
  } finally {
    contractLoading.value = false
  }
}

function parsePositiveInteger(value: string) {
  const parsed = Number(value.trim())
  if (!Number.isSafeInteger(parsed) || parsed <= 0) throw new Error(t('option.invalidSeriesInput'))
  return parsed
}

function parseExpiries(): ContractSeriesExpiryInput[] {
  return form.expiryLines
    .split('\n')
    .filter((line) => line.trim())
    .map((line, index) => {
      const [cycleCode, listTime, lastTradeTime, cutoff, expireTime, deliverTime] = line
        .split(',')
        .map((item) => item.trim())
      if (!cycleCode || !deliverTime) throw new Error(t('option.invalidSeriesInput'))
      return {
        sequenceNo: index + 1,
        cycleCode,
        listTime: parsePositiveInteger(listTime),
        lastTradeTime: parsePositiveInteger(lastTradeTime),
        exerciseCutoffTime: parsePositiveInteger(cutoff),
        expireTime: parsePositiveInteger(expireTime),
        deliverTime: parsePositiveInteger(deliverTime),
      }
    })
}

function parseBands(): ContractSeriesStrikeBandInput[] {
  return form.bandLines
    .split('\n')
    .filter((line) => line.trim())
    .map((line, index) => {
      const [lowerStrike, upperStrike, strikeStep] = line.split(',').map((item) => item.trim())
      if (!lowerStrike || !upperStrike || !strikeStep) {
        throw new Error(t('option.invalidSeriesInput'))
      }
      return { sequenceNo: index + 1, lowerStrike, upperStrike, strikeStep }
    })
}

function contractTemplate(source: OptionContractDetail): CreateContractReq {
  const item = source.contract
  return {
    tenantId: query.tenantId || 0,
    contractCode: '',
    underlyingSymbol: item.underlyingSymbol,
    underlyingCoin: item.underlyingCoin,
    settleCoin: item.settleCoin,
    quoteCoin: item.quoteCoin,
    optionType: 1,
    exerciseStyle: item.exerciseStyle,
    settlementType: item.settlementType,
    strikePrice: form.referencePrice,
    contractUnit: item.contractUnit,
    minOrderQty: item.minOrderQty,
    maxOrderQty: item.maxOrderQty,
    priceTick: item.priceTick,
    qtyStep: item.qtyStep,
    multiplier: item.multiplier,
    listTime: 0,
    lastTradeTime: 0,
    expireTime: 0,
    deliverTime: 0,
    exerciseCutoffTime: 0,
    autoExerciseThreshold: item.autoExerciseThreshold,
    maxUserLongQty: item.maxUserLongQty,
    maxUserShortQty: item.maxUserShortQty,
    maxOpenInterest: item.maxOpenInterest,
    orderPriceBandRatio: item.orderPriceBandRatio,
    circuitBreakerRatio: item.circuitBreakerRatio,
    greeksMaxAgeSeconds: item.greeksMaxAgeSeconds,
    settlementPriceSource: item.settlementPriceSource,
    settlementPriceMethod: item.settlementPriceMethod,
    settlementWindowSeconds: item.settlementWindowSeconds,
    settlementMinSamples: item.settlementMinSamples,
    isAutoExercise: item.isAutoExercise,
    status: 1,
    sort: item.sort,
    remark: item.remark,
    makerFeeRate: item.makerFeeRate,
    takerFeeRate: item.takerFeeRate,
    exerciseFeeRate: item.exerciseFeeRate,
    feeUserId: item.feeUserId,
    feeAccountId: item.feeAccountId,
    sellerMarginMode: item.sellerMarginMode,
    initialMarginRate: item.initialMarginRate,
    maintenanceMarginRate: item.maintenanceMarginRate,
    minMarginRate: item.minMarginRate,
    liquidationFeeRate: item.liquidationFeeRate,
    insuranceUserId: item.insuranceUserId,
    insuranceAccountId: item.insuranceAccountId,
    liquidationDeficitPolicy: item.liquidationDeficitPolicy,
    physicalDeliveryPolicy: item.physicalDeliveryPolicy,
    physicalDeliveryCureSeconds: item.physicalDeliveryCureSeconds,
    tradingCalendarCode: item.tradingCalendarCode,
  }
}

async function createSeries() {
  if (!query.tenantId || !form.sourceContractId) return
  const source = contracts.value.find((item) => item.contract.id === form.sourceContractId)
  if (!source) return
  saving.value = true
  try {
    await optionService.createContractSeries({
      tenantId: query.tenantId,
      requestKey: globalThis.crypto.randomUUID(),
      seriesCode: form.seriesCode,
      contractTemplate: contractTemplate(source),
      referencePrice: form.referencePrice,
      referenceSource: form.referenceSource,
      referenceTime: Math.floor(form.referenceTime.getTime() / 1000),
      evidenceRef: form.evidenceRef,
      changeReason: form.changeReason,
      expiries: parseExpiries(),
      strikeBands: parseBands(),
    })
    ElMessage.success(t('common.success'))
    createVisible.value = false
    await refresh()
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : String(error))
  } finally {
    saving.value = false
  }
}

async function review(row: OptionContractSeries, approve: boolean) {
  const { value } = await ElMessageBox.prompt(
    t('option.reviewReason'),
    t('option.contractSeriesReview'),
    { inputValidator: (text) => Boolean(text?.trim()) },
  )
  await optionService.reviewContractSeries({
    tenantId: row.tenantId,
    seriesId: row.id,
    approve,
    reason: value,
  })
  ElMessage.success(t('common.success'))
  await refresh()
}

async function showDetails(row: OptionContractSeries) {
  detailVisible.value = true
  detailLoading.value = true
  try {
    const response = await optionService.listContractSeriesDetails({
      tenantId: row.tenantId,
      seriesId: row.id,
      limit: 500,
    })
    details.value = response.data || []
  } finally {
    detailLoading.value = false
  }
}

async function reviewLaunch(row: OptionContractSeries, approve: boolean) {
  const { value } = await ElMessageBox.prompt(
    t('option.reviewReason'),
    t('option.contractSeriesLaunchReview'),
    { inputValidator: (text) => Boolean(text?.trim()) },
  )
  await optionService.reviewContractSeriesLaunch({
    tenantId: row.tenantId,
    seriesId: row.id,
    approve,
    reason: value,
  })
  ElMessage.success(t('common.success'))
  await refresh()
}

function statusName(value: number) {
  return statuses.value.find((item) => item.value === value)?.label || String(value)
}

function launchStatusName(value: number) {
  const names: Record<number, string> = {
    0: '-',
    1: t('option.contractSeriesLaunchPending'),
    2: t('option.contractSeriesLaunchApproved'),
    3: t('option.contractSeriesLaunchRejected'),
  }
  return names[value] || String(value)
}

onMounted(refresh)
</script>

<style scoped>
.table-card {
  margin-top: 16px;
}
.table-alert,
.form-alert {
  margin-bottom: 16px;
}
.count-note {
  margin-left: 12px;
  color: var(--el-text-color-secondary);
}
</style>
