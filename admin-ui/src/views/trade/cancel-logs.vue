<template>
  <div class="module-page">
    <CrudQueryCard :model="currentQuery" @search="loadList" @reset="resetQuery">
      <el-form-item :label="t('trade.tenantId')">
        <TenantSelect v-model="currentQuery.tenantId" class="tenant-select-filter" />
      </el-form-item>
      <el-form-item :label="t('trade.userId')">
        <UserSelect
          v-model="currentQuery.userId"
          :tenant-id="currentQuery.tenantId || undefined"
        />
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
        />

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

    <el-dialog v-model="detailVisible" :title="t('option.detail')" width="760px">
      <pre class="detail-pre">{{ JSON.stringify(detailData, null, 2) }}</pre>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { usePagination } from '@/composables'
import { tradeService, type TradeCancelLog } from '@/services'
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

<style scoped>
.detail-pre {
  white-space: pre-wrap;
  word-break: break-all;
}
</style>
