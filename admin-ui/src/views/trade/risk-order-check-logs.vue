<template>
  <div class="module-page">
    <div class="page-header">
      <h2>{{ t('trade.riskOrderCheckLogs') }}</h2>
      <div class="header-actions">
        <el-button @click="loadCurrent">
          {{ t('common.refresh') }}
        </el-button>
      </div>
    </div>
    <el-card shadow="never" class="query-card">
      <template #header>
        {{ t('trade.riskQuery') }}
      </template>
      <el-form :model="riskQuery" inline label-width="90px">
        <el-form-item :label="t('trade.tenantId')">
          <el-input-number v-model="riskQuery.tenantId" :min="0" :precision="0" />
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
        <el-form-item>
          <el-button type="primary" @click="loadRiskData">
            {{ t('trade.loadConfig') }}
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>
    <el-card shadow="never" class="table-card">
      <template #header>
        {{ t('trade.riskLogs') }}
      </template>
      <el-form
        :model="riskLogQuery"
        inline
        label-width="90px"
        class="query-card-inner"
      >
        <el-form-item :label="t('trade.tenantId')">
          <el-input-number v-model="riskLogQuery.tenantId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('trade.userId')">
          <el-input-number v-model="riskLogQuery.userId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('trade.symbolId')">
          <el-input-number v-model="riskLogQuery.symbolId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('trade.orderNo')">
          <el-input v-model="riskLogQuery.orderNo" clearable />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadRiskLogs">
            {{ t('common.search') }}
          </el-button>
        </el-form-item>
      </el-form>
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

const loadCurrent = async () => {
  await loadRiskLogs()
}

const loadRiskData = async () => {
  await loadRiskLogs()
}

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
<style scoped></style>
