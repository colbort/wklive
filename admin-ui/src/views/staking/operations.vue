<template>
  <div class="module-page">
    <CrudQueryCard :model="query" @search="loadList" @reset="resetQuery">
      <el-form-item :label="t('staking.tenantId')">
        <TenantSelect v-model="query.tenantId" class="tenant-select-filter" />
      </el-form-item>
      <el-form-item :label="t('staking.operationNo')">
        <el-input v-model="query.operationNo" clearable />
      </el-form-item>
      <el-form-item :label="t('staking.orderNo')">
        <el-input v-model="query.orderNo" clearable />
      </el-form-item>
      <el-form-item :label="t('staking.operationType')">
        <el-select v-model="query.operationType" clearable>
          <el-option
            v-for="item in operationTypes"
            :key="item.value"
            :label="getOptionLabel(t, item.code, item.value)"
            :value="item.value"
          />
        </el-select>
      </el-form-item>
      <el-form-item :label="t('common.status')">
        <el-select v-model="query.status" clearable>
          <el-option
            v-for="item in operationStatuses"
            :key="item.value"
            :label="getOptionLabel(t, item.code, item.value)"
            :value="item.value"
          />
        </el-select>
      </el-form-item>
    </CrudQueryCard>

    <el-card shadow="never" class="table-card">
      <el-table v-loading="loading" :data="rows" stripe>
        <el-table-column
          prop="operationNo"
          :label="t('staking.operationNo')"
          min-width="220"
          show-overflow-tooltip
        />
        <el-table-column
          prop="orderNo"
          :label="t('staking.orderNo')"
          min-width="180"
          show-overflow-tooltip
        />
        <el-table-column prop="operationType" :label="t('staking.operationType')" width="130">
          <template #default="{ row }">{{ operationTypeLabel(row.operationType) }}</template>
        </el-table-column>
        <el-table-column
          prop="principalAmount"
          :label="t('staking.principalAmount')"
          min-width="120"
        >
          <template #default="{ row }">{{ formatAmount(row.principalAmount) }}</template>
        </el-table-column>
        <el-table-column prop="rewardAmount" :label="t('staking.rewardAmount')" min-width="120">
          <template #default="{ row }">{{ formatAmount(row.rewardAmount) }}</template>
        </el-table-column>
        <el-table-column prop="feeAmount" :label="t('staking.feeAmount')" min-width="110">
          <template #default="{ row }">{{ formatAmount(row.feeAmount) }}</template>
        </el-table-column>
        <el-table-column prop="retryCount" :label="t('staking.retryCount')" width="100" />
        <el-table-column prop="status" :label="t('common.status')" width="120">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)">{{
              operationStatusLabel(row.status)
            }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column
          prop="lastError"
          :label="t('staking.lastError')"
          min-width="220"
          show-overflow-tooltip
        />
        <el-table-column prop="updateTimes" :label="t('common.updateTimes')" min-width="170">
          <template #default="{ row }">{{ formatTime(row.updateTimes) }}</template>
        </el-table-column>
        <el-table-column :label="t('common.actions')" fixed="right" align="center" width="150">
          <template #default="{ row }">
            <el-button link type="primary" @click="showDetail(row)">{{
              t('market.detail')
            }}</el-button>
            <el-button
              v-perm="'staking:operation:retry'"
              link
              type="warning"
              :disabled="!canRetry(row)"
              @click="retryOperation(row)"
            >
              {{ t('staking.retry') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <CursorPagination
        v-model:limit="pagination.limit"
        :total="pagination.total"
        :has-prev="pagination.hasPrev"
        :has-next="pagination.hasNext"
        @prev="prevAndLoad(loadList)"
        @next="nextAndLoad(loadList)"
        @limit-change="resetAndLoad(loadList)"
      />
    </el-card>

    <el-dialog v-model="detailVisible" :title="t('staking.operationDetail')" width="820px">
      <el-descriptions v-if="detail" :column="2" border>
        <el-descriptions-item :label="t('staking.operationNo')">{{
          detail.operationNo
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('staking.requestNo')">{{
          detail.requestNo
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('staking.orderNo')">{{
          detail.orderNo
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('staking.userId')">{{
          detail.userId
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('staking.operationType')">{{
          operationTypeLabel(detail.operationType)
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('common.status')">{{
          operationStatusLabel(detail.status)
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('staking.principalStep')">{{
          stepStatusLabel(detail.principalStatus)
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('staking.rewardStep')">{{
          stepStatusLabel(detail.rewardStatus)
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('staking.feeStep')">{{
          stepStatusLabel(detail.feeStatus)
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('staking.retryCount')">{{
          detail.retryCount
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('staking.nextRetryAt')">{{
          formatTime(detail.nextRetryAt)
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('staking.operatorUid')">{{
          detail.operatorUserId || '-'
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('staking.lastError')" :span="2">{{
          detail.lastError || '-'
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('common.remark')" :span="2">{{
          detail.remark || '-'
        }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { usePagination } from '@/composables'
import { stakingService, type OptionGroup, type StakeOperation } from '@/services'
import TenantSelect from '@/components/TenantSelect.vue'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'
import { findFormOptionGroup, getOptionLabel, getOptionValueLabel } from '@/utils/options'

const { t } = useI18n()
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)
const loading = ref(false)
const rows = ref<StakeOperation[]>([])
const optionGroups = ref<OptionGroup[]>([])
const detail = ref<StakeOperation | null>(null)
const detailVisible = computed({
  get: () => detail.value !== null,
  set: (value) => {
    if (!value) detail.value = null
  },
})
const query = reactive({
  tenantId: undefined as number | undefined,
  operationNo: '',
  orderNo: '',
  operationType: undefined as number | undefined,
  status: undefined as number | undefined,
})

const operationTypes = computed(() =>
  findFormOptionGroup(optionGroups.value, 'stakingOperationType'),
)
const operationStatuses = computed(() =>
  findFormOptionGroup(optionGroups.value, 'stakingOperationStatus'),
)
const operationTypeLabel = (value: number) =>
  getOptionValueLabel(optionGroups.value, 'stakingOperationType', value, t)
const operationStatusLabel = (value: number) =>
  getOptionValueLabel(optionGroups.value, 'stakingOperationStatus', value, t)
const stepStatusLabel = (value: number) =>
  getOptionValueLabel(optionGroups.value, 'stakingOperationStepStatus', value, t)
const statusTagType = (value: number) =>
  value === 3 ? 'success' : value >= 4 ? 'danger' : 'warning'
const canRetry = (row: StakeOperation) => row.status === 4 || row.status === 5
const formatAmount = (value: string) => {
  const text = String(value || '0')
  return text.includes('.') ? text.replace(/0+$/, '').replace(/\.$/, '') || '0' : text
}
const formatTime = (value: number) => (value > 0 ? new Date(value).toLocaleString() : '-')

async function loadList() {
  loading.value = true
  try {
    const res = await stakingService.listOperations({
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

async function loadOptions() {
  optionGroups.value = (await stakingService.getOptions()).data || []
}

function resetQuery() {
  Object.assign(query, {
    tenantId: undefined,
    operationNo: '',
    orderNo: '',
    operationType: undefined,
    status: undefined,
  })
  resetAndLoad(loadList)
}

function showDetail(row: StakeOperation) {
  detail.value = row
}

async function retryOperation(row: StakeOperation) {
  if (!canRetry(row)) return
  const result = await ElMessageBox.prompt(
    t('staking.retryReasonPrompt'),
    t('staking.retryOperation'),
    {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      inputValidator: (value) => Boolean(value?.trim()) || t('staking.retryReasonRequired'),
    },
  )
  await stakingService.retryOperation({
    tenantId: row.tenantId,
    id: row.id,
    reason: result.value.trim(),
  })
  ElMessage.success(t('staking.retryQueued'))
  await loadList()
}

onMounted(() => Promise.all([loadList(), loadOptions()]))
</script>
