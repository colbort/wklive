<template>
  <div class="module-page">
    <CrudQueryCard :model="currentQuery" @search="loadList" @reset="resetQuery">
      <el-form-item :label="t('trade.tenantId')">
        <TenantSelect v-model="currentQuery.tenantId" class="tenant-select-filter" />
      </el-form-item>
      <el-form-item :label="t('trade.userId')">
        <UserSelect v-model="currentQuery.userId" :tenant-id="currentQuery.tenantId || undefined" />
      </el-form-item>
      <el-form-item :label="t('trade.orderNo')">
        <el-input v-model="currentQuery.orderNo" clearable />
      </el-form-item>
    </CrudQueryCard>

    <el-card shadow="never" class="table-card">
      <el-table v-loading="loading" :data="rows" stripe>
        <el-table-column
          v-for="column in currentColumns"
          :key="column.prop"
          :prop="column.prop"
          :label="column.label"
          :min-width="column.width || 140"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            <el-tag
              v-if="column.prop === 'cancelSource'"
              :type="cancelSourceTagType(row.cancelSource)"
            >
              {{ cancelSourceLabel(row.cancelSource) }}
            </el-tag>
            <span v-else>{{ row[column.prop] }}</span>
          </template>
        </el-table-column>

        <el-table-column
          :label="t('common.actions')"
          align="center"
          width="100"
          fixed="right"
        >
          <template #default="{ row }">
            <el-button link type="primary" @click="showDetail(row)">
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

    <el-dialog v-model="detailVisible" :title="t('trade.cancelLogDetail')" width="760px">
      <el-descriptions v-if="detailData" :column="2" border>
        <el-descriptions-item :label="t('trade.id')">
          {{ detailData.id }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('trade.tenantId')">
          {{ detailData.tenantId }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('trade.orderNo')" :span="2">
          {{ detailData.orderNo }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('trade.orderId')">
          {{ detailData.orderId }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('trade.userId')">
          {{ detailData.userId }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('trade.cancelSource')">
          <el-tag :type="cancelSourceTagType(detailData.cancelSource)">
            {{ cancelSourceLabel(detailData.cancelSource) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item :label="t('trade.createTimes')">
          {{ formatDate(detailData.createTimes) }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('trade.cancelReason')" :span="2">
          {{ detailData.cancelReason || '-' }}
        </el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { usePagination } from '@/composables'
import { tradeService, type TradeCancelLog } from '@/services'
import { formatDate } from '@/utils'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'
import TenantSelect from '@/components/TenantSelect.vue'
import UserSelect from '@/components/UserSelect.vue'

const { t } = useI18n()
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)

interface CurrentQuery {
  tenantId: number | undefined
  userId: number | undefined
  orderNo: string
  limit: number
}

interface CurrentColumn {
  prop: string
  label: string
  width?: number
}

const loading = ref(false)
const rows = ref<TradeCancelLog[]>([])
const detailVisible = ref(false)
const detailData = ref<TradeCancelLog | null>(null)

const currentQuery = reactive<CurrentQuery>({
  tenantId: undefined,
  userId: undefined,
  orderNo: '',
  limit: 20,
})

const currentColumns: CurrentColumn[] = [
  { prop: 'orderNo', label: t('trade.orderNo'), width: 180 },
  { prop: 'userId', label: t('trade.userId'), width: 100 },
  { prop: 'cancelSource', label: t('trade.cancelSource'), width: 100 },
  { prop: 'cancelReason', label: t('trade.cancelReason'), width: 200 },
]

const cancelSourceLabels: Record<number, string> = {
  0: t('options.CANCEL_SOURCE_UNKNOWN'),
  1: t('options.CANCEL_SOURCE_USER'),
  2: t('options.CANCEL_SOURCE_SYSTEM'),
  3: t('options.CANCEL_SOURCE_RISK'),
  4: t('options.CANCEL_SOURCE_TIMEOUT'),
}

function cancelSourceLabel(source: number) {
  return cancelSourceLabels[source] || String(source)
}

function cancelSourceTagType(source: number): 'primary' | 'warning' | 'danger' | 'info' {
  if (source === 1) return 'primary'
  if (source === 3) return 'danger'
  if (source === 2 || source === 4) return 'warning'
  return 'info'
}

const loadList = async () => {
  loading.value = true
  try {
    const res = await tradeService.listCancelLogs({
      ...currentQuery,
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
  currentQuery.tenantId = undefined
  currentQuery.userId = undefined
  currentQuery.orderNo = ''
  currentQuery.limit = 100
  loadList()
}

const showDetail = (row: TradeCancelLog) => {
  detailData.value = row
  detailVisible.value = true
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
