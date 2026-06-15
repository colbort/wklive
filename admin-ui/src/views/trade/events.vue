<template>
  <div class="module-page">
    <CrudQueryCard
      :model="query"
      label-width="auto"
      @search="loadCurrent"
      @reset="resetCurrent"
    >
      <el-form-item :label="t('trade.tenantId')">
        <TenantSelect v-model="query.tenantId" class="tenant-select-filter" />
      </el-form-item>

      <el-form-item :label="t('trade.bizType')">
        <el-input v-model="query.bizType" clearable class="query-field" />
      </el-form-item>

      <el-form-item :label="t('trade.bizId')">
        <el-input v-model="query.bizId" clearable class="query-field" />
      </el-form-item>

      <el-form-item :label="t('trade.eventStatus')">
        <el-select v-model="query.eventStatus" clearable class="query-field">
          <el-option
            v-for="item in eventStatusOptions"
            :key="item.value"
            :label="optionItemLabel(item)"
            :value="item.value"
          />
        </el-select>
      </el-form-item>

      <el-form-item :label="t('trade.timeRange')">
        <el-date-picker
          v-model="timeRangeValue"
          type="datetimerange"
          value-format="x"
          clearable
          class="time-range"
        />
      </el-form-item>
    </CrudQueryCard>

    <el-card shadow="never" class="table-card">
      <el-table v-loading="loading" :data="rows" stripe>
        <el-table-column
          prop="eventNo"
          :label="t('trade.eventNo')"
          width="190"
          fixed="left"
          show-overflow-tooltip
        />

        <el-table-column prop="tenantId" :label="t('trade.tenantId')" width="100" />
        <el-table-column prop="userId" :label="t('trade.userId')" width="100" />
        <el-table-column prop="symbolId" :label="t('trade.symbolId')" width="100" />

        <el-table-column :label="t('trade.marketType')" min-width="130">
          <template #default="{ row }">
            {{ row.marketType ? optionLabel('marketType', row.marketType) : '-' }}
          </template>
        </el-table-column>

        <el-table-column :label="t('trade.source')" width="120">
          <template #default="{ row }">
            <el-tag size="small" effect="light">
              {{ optionLabel('sourceType', row.source) }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column :label="t('trade.eventStatus')" width="130">
          <template #default="{ row }">
            <el-tag size="small" :type="eventStatusTagType(row.eventStatus)" effect="light">
              {{ optionLabel('eventStatus', row.eventStatus) }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column :label="t('trade.retryCount')" width="120" align="right">
          <template #default="{ row }">
            {{ row.retryCount }} / {{ row.maxRetryCount }}
          </template>
        </el-table-column>

        <el-table-column :label="t('trade.nextRetryAt')" min-width="170">
          <template #default="{ row }">
            {{ formatDateOrDash(row.nextRetryAt) }}
          </template>
        </el-table-column>

        <el-table-column :label="t('trade.updateTimes')" min-width="170">
          <template #default="{ row }">
            {{ formatDateOrDash(row.updateTimes) }}
          </template>
        </el-table-column>

        <el-table-column :label="t('common.actions')" width="170" fixed="right">
          <template #default="{ row }">
            <el-button
              v-perm="'trade:event:detail'"
              link
              type="primary"
              @click="showDetail(row)"
            >
              <el-icon><View /></el-icon>
              {{ t('option.detail') }}
            </el-button>

            <el-button
              v-perm="'trade:event:retry'"
              link
              type="warning"
              @click="retryEvent(row)"
            >
              <el-icon><RefreshRight /></el-icon>
              {{ t('trade.retry') }}
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

    <el-drawer v-model="detailVisible" :title="detailTitle" size="820px">
      <div v-loading="detailLoading">
        <el-empty v-if="!detailData" :description="t('common.noData')" />

        <div v-else class="detail-layout">
          <div class="detail-header">
            <div>
              <div class="detail-title">
                {{ detailData.eventNo || '-' }}
              </div>
              <div class="detail-subtitle">
                {{ getEventType(detailData) }}
              </div>
            </div>
            <el-tag :type="eventStatusTagType(detailData.eventStatus)" effect="light">
              {{ optionLabel('eventStatus', detailData.eventStatus) }}
            </el-tag>
          </div>

          <el-descriptions :column="2" border>
            <el-descriptions-item :label="t('trade.id')">
              {{ detailData.id }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.tenantId')">
              {{ detailData.tenantId }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.userId')">
              {{ detailData.userId || '-' }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.operatorId')">
              {{ detailData.operatorId || '-' }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.symbolId')">
              {{ detailData.symbolId || '-' }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.marketType')">
              {{ detailData.marketType ? optionLabel('marketType', detailData.marketType) : '-' }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.source')">
              {{ optionLabel('sourceType', detailData.source) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.eventStatus')">
              {{ optionLabel('eventStatus', detailData.eventStatus) }}
            </el-descriptions-item>
          </el-descriptions>

          <el-descriptions :title="t('trade.eventBizInfo')" :column="2" border>
            <el-descriptions-item :label="t('trade.eventType')">
              {{ getEventType(detailData) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.eventNo')">
              {{ detailData.eventNo || '-' }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.bizType')">
              {{ getBizType(detailData) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.bizId')">
              {{ getBizId(detailData) }}
            </el-descriptions-item>
          </el-descriptions>

          <el-descriptions :title="t('trade.retryInfo')" :column="2" border>
            <el-descriptions-item :label="t('trade.retryCount')">
              {{ detailData.retryCount }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.maxRetryCount')">
              {{ detailData.maxRetryCount }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.nextRetryAt')">
              {{ formatDateOrDash(detailData.nextRetryAt) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.lastErrorMsg')" :span="2">
              <span class="error-text">{{ detailData.lastErrorMsg || '-' }}</span>
            </el-descriptions-item>
          </el-descriptions>

          <el-descriptions :title="t('trade.timeAndExt')" :column="2" border>
            <el-descriptions-item :label="t('trade.createTimes')">
              {{ formatDateOrDash(detailData.createTimes) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.updateTimes')">
              {{ formatDateOrDash(detailData.updateTimes) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('trade.payload')" :span="2">
              <pre class="detail-code">{{ formatJsonText(detailData.payload) }}</pre>
            </el-descriptions-item>
            <el-descriptions-item v-if="detailData.extData" :label="t('trade.extData')" :span="2">
              <pre class="detail-code">{{ formatJsonText(detailData.extData) }}</pre>
            </el-descriptions-item>
          </el-descriptions>
        </div>
      </div>
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { RefreshRight, View } from '@element-plus/icons-vue'
import { usePagination } from '@/composables'
import { tradeService, type BizTradeEvent, type OptionGroup, type OptionItem } from '@/services'
import { formatDate } from '@/utils'
import { findOptionGroup, getOptionLabel } from '@/utils/options'
import TenantSelect from '@/components/TenantSelect.vue'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'

type EventQuery = {
  tenantId?: number
  bizType: string
  bizId: string
  eventStatus?: number
}

const { t } = useI18n()
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)

const fallbackOptions: Record<string, OptionItem[]> = {
  marketType: [
    { value: 1, code: 'MARKET_TYPE_SPOT' },
    { value: 2, code: 'MARKET_TYPE_SECONDS_CONTRACT' },
    { value: 3, code: 'MARKET_TYPE_USDT_CONTRACT' },
    { value: 4, code: 'MARKET_TYPE_COIN_CONTRACT' },
  ],
  eventStatus: [
    { value: 1, code: 'EVENT_STATUS_PENDING' },
    { value: 2, code: 'EVENT_STATUS_SUCCESS' },
    { value: 3, code: 'EVENT_STATUS_FAILED' },
    { value: 4, code: 'EVENT_STATUS_CANCELED' },
  ],
  sourceType: [
    { value: 1, code: 'SOURCE_TYPE_SYSTEM' },
    { value: 2, code: 'SOURCE_TYPE_USER' },
    { value: 3, code: 'SOURCE_TYPE_ADMIN' },
    { value: 4, code: 'SOURCE_TYPE_TASK' },
  ],
}

const loading = ref(false)
const detailLoading = ref(false)
const rows = ref<BizTradeEvent[]>([])
const detailVisible = ref(false)
const detailData = ref<BizTradeEvent | null>(null)
const optionGroups = ref<OptionGroup[]>([])
const timeRangeValue = ref<[number, number] | []>([])

const query = reactive<EventQuery>({
  tenantId: undefined,
  bizType: '',
  bizId: '',
  eventStatus: undefined,
})

const detailTitle = computed(() => detailData.value?.eventNo || t('option.detail'))
const eventStatusOptions = computed(() =>
  getOptions('eventStatus').filter((item) => item.value !== 0),
)

function getOptions(key: string) {
  const options = findOptionGroup(optionGroups.value, key)
  return options.length ? options : fallbackOptions[key] || []
}

function optionItemLabel(item: OptionItem) {
  return getOptionLabel(t, item.code, item.value)
}

function optionLabel(key: string, value?: number | string) {
  const option = getOptions(key).find((item) => String(item.value) === String(value))
  return getOptionLabel(t, option?.code, option?.value || Number(value) || 0)
}

function eventStatusTagType(status: number) {
  if (status === 2) return 'success'
  if (status === 3) return 'danger'
  if (status === 4) return 'info'
  return 'warning'
}

function getEventType(row: BizTradeEvent) {
  const data = row as BizTradeEvent & { event_type?: string }
  return data.eventType || data.event_type || '-'
}

function getBizType(row: BizTradeEvent) {
  const data = row as BizTradeEvent & { biz_type?: string }
  return data.bizType || data.biz_type || '-'
}

function getBizId(row: BizTradeEvent) {
  const data = row as BizTradeEvent & { biz_id?: string }
  return data.bizId || data.biz_id || '-'
}

function formatDateOrDash(value?: number) {
  return value ? formatDate(value) : '-'
}

function buildTimeRange() {
  if (!Array.isArray(timeRangeValue.value) || timeRangeValue.value.length !== 2) return undefined
  const [startTime, endTime] = timeRangeValue.value
  if (!startTime || !endTime) return undefined
  return { startTime, endTime }
}

function formatJsonText(value?: string) {
  if (!value) return '-'
  try {
    return JSON.stringify(JSON.parse(value), null, 2)
  } catch {
    return value
  }
}

const loadCurrent = async () => {
  loading.value = true
  try {
    const res = await tradeService.listEvents({
      tenantId: query.tenantId,
      bizType: query.bizType || undefined,
      bizId: query.bizId || undefined,
      eventStatus: query.eventStatus,
      timeRange: buildTimeRange(),
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
  query.bizType = ''
  query.bizId = ''
  query.eventStatus = undefined
  timeRangeValue.value = []
  resetAndLoad(loadCurrent)
}

const showDetail = async (row: BizTradeEvent) => {
  detailVisible.value = true
  detailLoading.value = true
  detailData.value = row
  try {
    detailData.value =
      (await tradeService.getEvent({ tenantId: row.tenantId, id: row.id })).data || row
  } finally {
    detailLoading.value = false
  }
}

const retryEvent = async (row: BizTradeEvent) => {
  await tradeService.retryEvent({
    tenantId: row.tenantId,
    id: row.id,
    eventNo: row.eventNo,
    operatorId: 0,
  })
  ElMessage.success(t('trade.eventRetrySubmitted'))
  loadCurrent()
}

async function loadOptions() {
  optionGroups.value = (await tradeService.getOptions()).data || []
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
  await loadCurrent()
})
</script>

<style scoped>
.query-field {
  width: 180px;
}

.detail-header {
  display: grid;
  gap: 4px;
}

.muted,
.detail-subtitle {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.time-range {
  width: 360px;
}

.detail-layout {
  display: grid;
  gap: 16px;
}

.detail-header {
  grid-template-columns: minmax(0, 1fr) auto;
  align-items: center;
}

.detail-title {
  color: var(--el-text-color-primary);
  font-size: 18px;
  font-weight: 700;
}

.detail-code {
  max-height: 260px;
  padding: 12px;
  margin: 0;
  overflow: auto;
  border-radius: 6px;
  background: var(--el-fill-color-light);
  white-space: pre-wrap;
  word-break: break-all;
}

.error-text {
  color: var(--el-color-danger);
  white-space: pre-wrap;
  word-break: break-word;
}
</style>
