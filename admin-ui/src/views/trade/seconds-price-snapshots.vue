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
      <el-form-item :label="t('trade.snapshotType')">
        <el-input-number
          v-model="query.snapshotType"
          :min="1"
          :precision="0"
        />
      </el-form-item>
    </CrudQueryCard>
    <el-card shadow="never" class="table-card">
      <el-table v-loading="loading" :data="rows" stripe>
        <el-table-column prop="id" :label="t('trade.id')" min-width="100" />
        <el-table-column prop="orderId" :label="t('trade.orderId')" min-width="120" />
        <el-table-column prop="snapshotType" :label="t('trade.snapshotType')" min-width="140" />
        <el-table-column
          prop="source"
          :label="t('trade.source')"
          min-width="180"
          show-overflow-tooltip
        />
        <el-table-column prop="price" :label="t('trade.price')" min-width="130" />
        <el-table-column
          prop="quoteTime"
          :label="t('trade.quoteTime')"
          min-width="190"
        >
          <template #default="{ row }">
            {{ formatTime(row.quoteTime) }}
          </template>
        </el-table-column>
        <el-table-column
          prop="receivedAt"
          :label="t('trade.receivedAt')"
          min-width="190"
        >
          <template #default="{ row }">
            {{ formatTime(row.receivedAt) }}
          </template>
        </el-table-column>
        <el-table-column prop="algorithm" :label="t('trade.algorithm')" min-width="150" />
        <el-table-column prop="isSelected" :label="t('trade.isSelected')" min-width="110">
          <template #default="{ row }">
            <el-tag :type="row.isSelected === 1 ? 'success' : 'info'">
              {{
                row.isSelected === 1 ? t('common.yes') : t('common.no')
              }}
            </el-tag>
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
import { apiTradeListSecondsPriceSnapshots } from '@/api/trade'
import type { TradeSecondsPriceSnapshot } from '@/services'

const { t } = useI18n()
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)
const query = reactive<{ tenantId?: number; orderId?: number; snapshotType?: number }>({})
const rows = ref<TradeSecondsPriceSnapshot[]>([])
const loading = ref(false)
async function load() {
  loading.value = true
  try {
    const res = await apiTradeListSecondsPriceSnapshots({
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
  Object.assign(query, { tenantId: undefined, orderId: undefined, snapshotType: undefined })
  resetAndLoad(load)
}
function formatTime(value: number) {
  return value > 0 ? formatDate(value) : '-'
}
onMounted(load)
</script>
