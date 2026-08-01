<template>
  <div class="module-page">
    <CrudQueryCard :model="query" @search="loadList" @reset="resetQuery">
      <el-form-item :label="t('option.tenantId')">
        <TenantSelect v-model="query.tenantId" class="tenant-select-filter" />
      </el-form-item>
      <el-form-item :label="t('option.contractId')">
        <ContractSelect v-model="query.contractId" :tenant-id="query.tenantId || undefined" />
      </el-form-item>
      <el-form-item :label="t('option.settlementNo')">
        <el-input v-model="query.settlementNo" clearable />
      </el-form-item>
    </CrudQueryCard>

    <el-card shadow="never" class="table-card">
      <el-table v-loading="loading" :data="rows" stripe>
        <el-table-column
          :label="t('option.settlementNo')"
          prop="settlementNo"
          min-width="180"
          show-overflow-tooltip
        />
        <el-table-column :label="t('option.contractId')" prop="contractId" width="100" />
        <el-table-column
          :label="t('option.deliveryPrice')"
          prop="deliveryPrice"
          min-width="120"
          show-overflow-tooltip
        />
        <el-table-column :label="t('common.status')" prop="status" width="100" />
        <el-table-column
          :label="t('common.actions')"
          align="center"
          width="100"
          fixed="right"
        >
          <template #default="{ row }">
            <el-button
              v-perm="'option:settlement:detail'"
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

    <el-dialog v-model="detailVisible" :title="t('option.detail')" width="760px">
      <pre class="detail-pre">{{ JSON.stringify(detailData, null, 2) }}</pre>
      <el-table
        v-if="detailData && 'assetInstructions' in detailData"
        :data="detailData.assetInstructions"
        size="small"
      >
        <el-table-column prop="instructionNo" :label="t('option.instructionNo')" min-width="220" />
        <el-table-column prop="coin" :label="t('option.coin')" width="90" />
        <el-table-column prop="amount" :label="t('option.amount')" min-width="120" />
        <el-table-column prop="status" :label="t('common.status')" width="80" />
        <el-table-column :label="t('common.actions')" width="100">
          <template #default="{ row }">
            <el-button
              v-if="[4, 5].includes(row.status) && !row.deliveryUnitId"
              v-perm="'option:settlement-instruction:retry'"
              link
              type="warning"
              @click="retryInstruction(row.id)"
            >
              {{ t('option.retry') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-divider v-if="deliveryUnits.length">
        {{ t('option.physicalDeliveryUnits') }}
      </el-divider>
      <el-table v-if="deliveryUnits.length" :data="deliveryUnits" size="small">
        <el-table-column
          prop="deliveryUnitNo"
          :label="t('option.deliveryUnitNo')"
          min-width="210"
          show-overflow-tooltip
        />
        <el-table-column :label="t('option.longUser')" min-width="120">
          <template #default="{ row }">
            {{ row.longUserId }}/{{ row.longAccountId }}
          </template>
        </el-table-column>
        <el-table-column :label="t('option.shortUser')" min-width="120">
          <template #default="{ row }">
            {{ row.shortUserId }}/{{ row.shortAccountId }}
          </template>
        </el-table-column>
        <el-table-column prop="quantity" :label="t('option.quantity')" min-width="90" />
        <el-table-column :label="t('option.deliveryAsset')" min-width="140">
          <template #default="{ row }">
            {{ row.deliveryQuantity }} {{ row.deliveryCoin }}
          </template>
        </el-table-column>
        <el-table-column :label="t('option.paymentAsset')" min-width="140">
          <template #default="{ row }">
            {{ row.paymentAmount }} {{ row.paymentCoin }}
          </template>
        </el-table-column>
        <el-table-column prop="status" :label="t('common.status')" width="80" />
        <el-table-column prop="cureDeadline" :label="t('option.cureDeadline')" width="170">
          <template #default="{ row }">
            {{ row.cureDeadline ? new Date(row.cureDeadline * 1000).toLocaleString() : '-' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="lastErrorMsg"
          :label="t('option.lastErrorMsg')"
          min-width="180"
          show-overflow-tooltip
        />
        <el-table-column :label="t('common.actions')" width="100" fixed="right">
          <template #default="{ row }">
            <el-button
              v-if="[4, 6].includes(row.status)"
              v-perm="'option:physical-delivery:retry'"
              link
              type="warning"
              @click="retryDeliveryUnit(row)"
            >
              {{ t('option.retry') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { usePagination } from '@/composables'
import {
  optionService,
  type OptionPhysicalDeliveryUnit,
  type OptionSettlement,
  type OptionSettlementDetail,
} from '@/services'
import TenantSelect from '@/components/TenantSelect.vue'
import ContractSelect from '@/components/ContractSelect.vue'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'

const { t } = useI18n()
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)

const loading = ref(false)
const rows = ref<OptionSettlement[]>([])
const detailVisible = ref(false)
const detailData = ref<OptionSettlementDetail | OptionSettlement | null>(null)
const deliveryUnits = ref<OptionPhysicalDeliveryUnit[]>([])
const query = reactive({
  tenantId: undefined as number | undefined,
  contractId: undefined as number | undefined,
  settlementNo: '',
  status: undefined as number | undefined,
  limit: 20,
})

const loadList = async () => {
  loading.value = true
  try {
    const res = await optionService.listSettlements({
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

const resetQuery = () => {
  query.tenantId = undefined
  query.contractId = undefined
  query.settlementNo = ''
  query.status = undefined
  query.limit = 100
  loadList()
}

const showDetail = async (row: OptionSettlement) => {
  detailData.value =
    (
      await optionService.getSettlement({
        tenantId: row.tenantId,
        id: row.id,
        settlementNo: row.settlementNo,
      })
    ).data || row
  deliveryUnits.value = []
  if (detailData.value && 'batch' in detailData.value && detailData.value.batch?.id) {
    const units = await optionService.listPhysicalDeliveryUnits({
      tenantId: row.tenantId,
      batchId: detailData.value.batch.id,
      limit: 200,
    })
    deliveryUnits.value = units?.data || []
  }
  detailVisible.value = true
}

const retryInstruction = async (instructionId: number) => {
  if (!detailData.value || !('settlement' in detailData.value)) return
  const { value } = await ElMessageBox.prompt(t('option.retryReason'), t('option.retry'), {
    inputType: 'textarea',
    inputValidator: (input) => {
      const reason = input?.trim() || ''
      if (!reason) return t('option.retryReasonRequired')
      return Array.from(reason).length <= 64 || t('option.retryReasonTooLong')
    },
  })
  await optionService.retrySettlementInstruction({
    tenantId: detailData.value.settlement.tenantId,
    settlementId: detailData.value.settlement.id,
    instructionId,
    reason: value.trim(),
  })
  ElMessage.success(t('common.success'))
  await showDetail(detailData.value.settlement)
}

const retryDeliveryUnit = async (row: OptionPhysicalDeliveryUnit) => {
  const { value } = await ElMessageBox.prompt(
    t('option.physicalDeliveryRetryReason'),
    t('option.retry'),
    {
      inputType: 'textarea',
      inputValidator: (input) => Boolean(input?.trim()) || t('option.reasonRequired'),
    },
  )
  await optionService.retryPhysicalDeliveryUnit({
    tenantId: row.tenantId,
    deliveryUnitId: row.id,
    reason: value.trim(),
  })
  ElMessage.success(t('common.success'))
  if (detailData.value && 'settlement' in detailData.value) {
    await showDetail(detailData.value.settlement)
  }
}

function handleLimitChange() {
  resetAndLoad(loadList)
}

function handlePrevPage() {
  prevAndLoad(loadList)
}

function handleNextPage() {
  nextAndLoad(loadList)
}

onMounted(loadList)
</script>

<style scoped></style>
