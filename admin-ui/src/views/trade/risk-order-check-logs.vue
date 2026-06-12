<template>
  <div class="module-page">
    <CrudQueryCard
      :model="riskQuery"
      label-width="90px"
      @search="loadRiskLogs"
      @reset="resetRiskLogQuery"
    >
      <el-form-item :label="t('trade.tenantId')">
        <TenantSelect v-model="riskQuery.tenantId" class="tenant-select-filter" />
      </el-form-item>
      <el-form-item :label="t('trade.userId')">
        <el-input-number v-model="riskQuery.userId" :min="0" :precision="0" />
      </el-form-item>
      <el-form-item :label="t('trade.symbolId')">
        <el-input-number v-model="riskQuery.symbolId" :min="0" :precision="0" />
      </el-form-item>
      <el-form-item :label="t('trade.marketType')">
        <el-input-number v-model="riskQuery.marketType" :min="0" :precision="0" />
      </el-form-item>
    </CrudQueryCard>
    <el-card shadow="never" class="table-card">
      <el-table v-loading="loading" :data="rows" stripe>
        <el-table-column
          prop="orderNo"
          :label="t('trade.orderNo')"
          min-width="180"
          show-overflow-tooltip
        />
        <el-table-column prop="userId" :label="t('trade.userId')" width="100" />
        <el-table-column prop="symbolId" :label="t('trade.symbolId')" width="100" />
        <el-table-column prop="checkType" :label="t('trade.checkType')" width="100" />
        <el-table-column prop="checkResult" :label="t('trade.checkResult')" width="100" />
        <el-table-column
          prop="rejectMsg"
          :label="t('trade.rejectMsg')"
          min-width="220"
          show-overflow-tooltip
        />
        <el-table-column :label="t('common.actions')" width="100" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="((detailData = row), (detailVisible = true))">
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
    </el-dialog>
  </div>
</template>
<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { usePagination } from '@/composables'
import { tradeService, type GetRiskOrderCheckLogListReq, type RiskOrderCheckLog } from '@/services'
import TenantSelect from '@/components/TenantSelect.vue'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'

const { t } = useI18n()
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)

const loading = ref(false)
interface RiskQuery {
  tenantId: number
  userId: number
  symbolId: number
  marketType: number
}

const rows = ref<RiskOrderCheckLog[]>([])
const detailVisible = ref(false)
const detailData = ref<RiskOrderCheckLog | null>(null)
const riskQuery = reactive<RiskQuery>({
  tenantId: 0,
  userId: 0,
  symbolId: 0,
  marketType: 0,
})
const riskLogQuery = reactive<GetRiskOrderCheckLogListReq>({
  tenantId: 0,
  userId: 0,
  symbolId: 0,
  orderNo: '',
  limit: 20,
})

const loadRiskLogs = async () => {
  loading.value = true
  try {
    const res = await tradeService.listRiskLogs({
      ...riskLogQuery,
      cursor: pagination.cursor,
      limit: pagination.limit,
    })
    rows.value = res?.data || []
    updateFromResponse(res)
  } finally {
    loading.value = false
  }
}

function resetRiskLogQuery() {
  riskLogQuery.tenantId = 0
  riskLogQuery.userId = 0
  riskLogQuery.symbolId = 0
  riskLogQuery.orderNo = ''
  resetAndLoad(loadRiskLogs)
}

function handleLimitChange() {
  resetAndLoad(loadRiskLogs)
}

function handlePrevPage() {
  prevAndLoad(loadRiskLogs)
}

function handleNextPage() {
  nextAndLoad(loadRiskLogs)
}

onMounted(loadRiskLogs)
</script>
