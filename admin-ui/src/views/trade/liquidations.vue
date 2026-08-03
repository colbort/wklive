<template>
  <div class="module-page">
    <CrudQueryCard :model="query" @search="load" @reset="resetQuery">
      <el-form-item :label="t('trade.tenantId')">
        <TenantSelect v-model="query.tenantId" />
      </el-form-item>
      <el-form-item :label="t('trade.userId')">
        <UserSelect v-model="query.userId" :tenant-id="query.tenantId" />
      </el-form-item>
      <el-form-item :label="t('trade.symbolId')">
        <SymbolSelect v-model="query.symbolId" :tenant-id="query.tenantId" />
      </el-form-item>
      <el-form-item :label="t('trade.positionId')">
        <el-input-number v-model="query.positionId" :min="0" :precision="0" />
      </el-form-item>
      <el-form-item :label="t('trade.status')">
        <el-input-number v-model="query.status" :min="1" :precision="0" />
      </el-form-item>
    </CrudQueryCard>
    <el-card shadow="never" class="table-card">
      <el-table v-loading="loading" :data="rows" stripe>
        <el-table-column
          prop="liquidationNo"
          :label="t('trade.liquidationNo')"
          min-width="210"
          show-overflow-tooltip
        />
        <el-table-column prop="userId" :label="t('trade.userId')" min-width="100" />
        <el-table-column prop="symbolId" :label="t('trade.symbolId')" min-width="110" />
        <el-table-column prop="positionId" :label="t('trade.positionId')" min-width="110" />
        <el-table-column prop="positionSide" :label="t('trade.positionSide')" min-width="120" />
        <el-table-column
          prop="triggerSnapshotId"
          :label="t('trade.triggerSnapshotId')"
          min-width="160"
        />
        <el-table-column prop="triggerQty" :label="t('trade.triggerQty')" min-width="130" />
        <el-table-column prop="liquidatedQty" :label="t('trade.liquidatedQty')" min-width="130" />
        <el-table-column prop="liquidationFee" :label="t('trade.liquidationFee')" min-width="140" />
        <el-table-column prop="status" :label="t('trade.status')" min-width="100" />
        <el-table-column
          prop="reason"
          :label="t('trade.reason')"
          min-width="180"
          show-overflow-tooltip
        />
        <el-table-column prop="completedAt" :label="t('trade.completedAt')" min-width="190">
          <template #default="{ row }">
            {{ formatTime(row.completedAt) }}
          </template>
        </el-table-column>
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
import UserSelect from '@/components/UserSelect.vue'
import { usePagination } from '@/composables'
import { formatDate } from '@/utils'
import { apiTradeListLiquidations } from '@/api/trade'
import type { ContractLiquidation } from '@/services'

const { t } = useI18n()
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)
const query = reactive<{
  tenantId?: number
  userId?: number
  symbolId?: number
  positionId?: number
  status?: number
}>({})
const rows = ref<ContractLiquidation[]>([])
const loading = ref(false)
async function load() {
  loading.value = true
  try {
    const res = await apiTradeListLiquidations({
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
  Object.assign(query, {
    tenantId: undefined,
    userId: undefined,
    symbolId: undefined,
    positionId: undefined,
    status: undefined,
  })
  resetAndLoad(load)
}
function formatTime(value: number) {
  return value > 0 ? formatDate(value) : '-'
}
onMounted(load)
</script>
