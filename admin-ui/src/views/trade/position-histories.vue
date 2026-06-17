<template>
  <div class="module-page">
    <CrudQueryCard :model="currentQuery" @search="loadList" @reset="resetQuery">
      <el-form-item :label="t('trade.tenantId')">
        <TenantSelect v-model="currentQuery.tenantId" class="tenant-select-filter" />
      </el-form-item>
      <el-form-item :label="t('trade.userId')">
        <UserSelect v-model="currentQuery.userId" :tenant-id="currentQuery.tenantId || undefined" />
      </el-form-item>
      <el-form-item :label="t('trade.symbolId')">
        <SymbolSelect
          v-model="currentQuery.symbolId"
          :tenant-id="currentQuery.tenantId || undefined"
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
import { tradeService, type ContractPositionHistory } from '@/services'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'
import TenantSelect from '@/components/TenantSelect.vue'
import UserSelect from '@/components/UserSelect.vue'
import SymbolSelect from '@/components/SymbolSelect.vue'

const { t } = useI18n()
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)

interface CurrentQuery {
  tenantId: number | undefined
  userId: number | undefined
  symbolId: number | undefined
  limit: number
}

interface CurrentColumn {
  prop: string
  label: string
  width?: number
}

const loading = ref(false)
const rows = ref<ContractPositionHistory[]>([])
const detailVisible = ref(false)
const detailData = ref<ContractPositionHistory | null>(null)

const currentQuery = reactive<CurrentQuery>({
  tenantId: undefined,
  userId: undefined,
  symbolId: undefined,
  limit: 20,
})

const currentColumns: CurrentColumn[] = [
  { prop: 'positionId', label: t('trade.positionId'), width: 120 },
  { prop: 'userId', label: t('trade.userId'), width: 100 },
  { prop: 'symbolId', label: t('trade.symbolId'), width: 100 },
  { prop: 'actionType', label: t('trade.actionType'), width: 100 },
  { prop: 'realizedPnlDelta', label: t('trade.realizedPnl') },
]

const loadList = async () => {
  loading.value = true
  try {
    const res = await tradeService.listPositionHistories({
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
  currentQuery.symbolId = undefined
  currentQuery.limit = 100
  loadList()
}

const showDetail = (row: ContractPositionHistory) => {
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
