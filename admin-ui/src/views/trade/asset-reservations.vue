<template>
  <div class="module-page">
    <CrudQueryCard :model="query" @search="load" @reset="resetQuery">
      <el-form-item :label="t('trade.tenantId')">
        <TenantSelect v-model="query.tenantId" />
      </el-form-item>
      <el-form-item :label="t('trade.orderId')">
        <el-input-number
          v-model="query.orderId"
          :min="0"
          :precision="0"
        />
      </el-form-item>
      <el-form-item :label="t('trade.status')">
        <el-input-number
          v-model="query.status"
          :min="1"
          :precision="0"
        />
      </el-form-item>
    </CrudQueryCard>
    <el-card shadow="never" class="table-card">
      <el-table v-loading="loading" :data="rows" stripe>
        <el-table-column
          prop="reservationNo"
          :label="t('trade.reservationNo')"
          min-width="210"
          show-overflow-tooltip
        />
        <el-table-column prop="orderId" :label="t('trade.orderId')" min-width="120" />
        <el-table-column prop="asset" :label="t('trade.asset')" min-width="100" />
        <el-table-column prop="reservedAmount" :label="t('trade.reservedAmount')" min-width="140" />
        <el-table-column prop="consumedAmount" :label="t('trade.consumedAmount')" min-width="140" />
        <el-table-column prop="releasedAmount" :label="t('trade.releasedAmount')" min-width="140" />
        <el-table-column prop="status" :label="t('trade.status')" min-width="100" />
        <el-table-column prop="retryCount" :label="t('trade.retryCount')" min-width="110" />
        <el-table-column
          prop="lastErrorMsg"
          :label="t('trade.lastErrorMsg')"
          min-width="220"
          show-overflow-tooltip
        />
        <el-table-column
          prop="updateTimes"
          :label="t('trade.updateTimes')"
          min-width="190"
        >
          <template #default="{ row }">
            {{
              formatTime(row.updateTimes)
            }}
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
import TenantSelect from '@/components/TenantSelect.vue'
import { usePagination } from '@/composables'
import { formatDate } from '@/utils'
import { apiTradeListAssetReservations } from '@/api/trade'
import type { TradeAssetReservation } from '@/services'

const { t } = useI18n()
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)
const query = reactive<{ tenantId?: number; orderId?: number; status?: number }>({})
const rows = ref<TradeAssetReservation[]>([])
const loading = ref(false)
async function load() {
  loading.value = true
  try {
    const res = await apiTradeListAssetReservations({
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
  Object.assign(query, { tenantId: undefined, orderId: undefined, status: undefined })
  resetAndLoad(load)
}
function formatTime(value: number) {
  return value > 0 ? formatDate(value) : '-'
}
onMounted(load)
</script>
