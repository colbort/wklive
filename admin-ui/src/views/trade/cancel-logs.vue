<template>
  <div class="module-page">
    <CrudQueryCard
      :model="currentQuery"
      label-width="auto"
      @search="loadCurrent"
      @reset="resetCurrent"
    >
      <el-form-item v-for="field in currentFields" :key="field.key" :label="field.label">
        <el-input v-if="field.type !== 'number'" v-model="currentQuery[field.key]" clearable />

        <el-input-number
          v-else
          v-model="currentQuery[field.key]"
          :min="0"
          :precision="0"
        />
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

        <el-table-column :label="t('common.actions')" width="100" fixed="right">
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

const { t } = useI18n()
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)

interface CurrentQuery {
  tenantId: number | undefined
  userId: number | undefined
  orderNo: string
  limit: number
}

type CurrentField =
  | {
      key: 'tenantId' | 'userId'
      label: string
      type: 'number'
    }
  | {
      key: 'orderNo'
      label: string
      type?: 'text'
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

const currentFields: CurrentField[] = [
  { key: 'tenantId', label: t('trade.tenantId'), type: 'number' },
  { key: 'userId', label: t('trade.userId'), type: 'number' },
  { key: 'orderNo', label: t('trade.orderNo'), type: 'text' },
]

const currentColumns: CurrentColumn[] = [
  { prop: 'orderNo', label: t('trade.orderNo'), width: 180 },
  { prop: 'userId', label: t('trade.userId'), width: 100 },
  { prop: 'cancelSource', label: t('trade.cancelSource'), width: 100 },
  { prop: 'cancelReason', label: t('trade.cancelReason'), width: 200 },
]

const loadCurrent = async () => {
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

const resetCurrent = () => {
  currentQuery.tenantId = undefined
  currentQuery.userId = undefined
  currentQuery.orderNo = ''
  currentQuery.limit = 100
  loadCurrent()
}

const showDetail = (row: TradeCancelLog) => {
  detailData.value = row
  detailVisible.value = true
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

<style scoped>
.detail-pre {
  white-space: pre-wrap;
  word-break: break-all;
}
</style>
