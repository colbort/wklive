<template>
  <div class="module-page">
    <CrudQueryCard :model="query" @search="load" @reset="resetQuery">
      <el-form-item :label="t('trade.tenantId')">
        <TenantSelect v-model="query.tenantId" />
      </el-form-item>
      <el-form-item :label="t('trade.symbolId')">
        <SymbolSelect v-model="query.symbolId" :tenant-id="query.tenantId" />
      </el-form-item>
      <el-form-item :label="t('trade.status')">
        <el-input-number v-model="query.status" :min="1" :precision="0" />
      </el-form-item>
    </CrudQueryCard>
    <el-card shadow="never" class="table-card">
      <el-table v-loading="loading" :data="rows" stripe>
        <el-table-column
          prop="batchNo"
          :label="t('trade.batchNo')"
          min-width="210"
          show-overflow-tooltip
        />
        <el-table-column prop="symbolId" :label="t('trade.symbolId')" min-width="110" />
        <el-table-column
          prop="settlementPrice"
          :label="t('trade.settlementPrice')"
          min-width="140"
        />
        <el-table-column prop="priceSource" :label="t('trade.priceSource')" min-width="150" />
        <el-table-column prop="priceAlgorithm" :label="t('trade.priceAlgorithm')" min-width="150" />
        <el-table-column prop="formulaVersion" :label="t('trade.formulaVersion')" min-width="140" />
        <el-table-column
          prop="sampleSnapshot"
          :label="t('trade.sampleSnapshot')"
          min-width="220"
          show-overflow-tooltip
        />
        <el-table-column prop="deliveryTime" :label="t('trade.deliveryTime')" min-width="190">
          <template #default="{ row }">
            {{ formatTime(row.deliveryTime) }}
          </template>
        </el-table-column>
        <el-table-column prop="status" :label="t('trade.status')" min-width="100" />
        <el-table-column
          prop="settledPositions"
          :label="t('trade.settledPositions')"
          min-width="140"
        />
        <el-table-column prop="totalPositions" :label="t('trade.totalPositions')" min-width="130" />
        <el-table-column
          prop="lastErrorMsg"
          :label="t('trade.lastErrorMsg')"
          min-width="220"
          show-overflow-tooltip
        />
      </el-table>
      <CursorPagination
        v-model:limit="pagination.limit"
        :total="pagination.total"
        :has-prev="pagination.hasPrev"
        :has-next="pagination.hasNext"
        @prev="prevAndLoad(load)"
        @next="nextAndLoad(load)"
        @limit-change="resetAndLoad(load)"
      />
    </el-card>
  </div>
</template>
<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'
import SymbolSelect from '@/components/SymbolSelect.vue'
import TenantSelect from '@/components/TenantSelect.vue'
import { usePagination } from '@/composables'
import { formatDate } from '@/utils'
import { apiTradeListDeliveryBatches } from '@/api/trade'
import type { ContractDeliveryBatch } from '@/services'

const { t } = useI18n()
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)
const query = reactive<{ tenantId?: number; symbolId?: number; status?: number }>({})
const rows = ref<ContractDeliveryBatch[]>([])
const loading = ref(false)
async function load() {
  loading.value = true
  try {
    const res = await apiTradeListDeliveryBatches({
      ...query,
      cursor: pagination.cursor,
      limit: pagination.limit,
    })
    rows.value = res.data || []
    updateFromResponse(res)
  } finally {
    loading.value = false
  }
}
function resetQuery() {
  Object.assign(query, { tenantId: undefined, symbolId: undefined, status: undefined })
  resetAndLoad(load)
}
function formatTime(value: number) {
  return value > 0 ? formatDate(value) : '-'
}
onMounted(load)
</script>
