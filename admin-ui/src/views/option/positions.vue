<template>
  <div class="module-page">
    <CrudQueryCard
      :model="query"
      @search="loadList"
      @reset="resetQuery"
    >
      <el-form-item :label="t('option.tenantId')">
        <TenantSelect v-model="query.tenantId" class="tenant-select-filter" />
      </el-form-item>
      <el-form-item :label="t('option.userId')">
        <el-input-number v-model="query.userId" :min="0" :precision="0" />
      </el-form-item>
      <el-form-item :label="t('option.contractId')">
        <el-input-number v-model="query.contractId" :min="0" :precision="0" />
      </el-form-item>
    </CrudQueryCard>

    <el-card shadow="never" class="table-card">
      <el-table v-loading="loading" :data="rows" stripe>
        <el-table-column :label="t('option.userId')" prop="userId" width="100" />
        <el-table-column :label="t('option.contractId')" prop="contractId" width="100" />
        <el-table-column :label="t('option.side')" prop="side" width="100" />
        <el-table-column
          :label="t('option.positionQty')"
          prop="positionQty"
          min-width="120"
          show-overflow-tooltip
        />
        <el-table-column
          :label="t('option.unrealizedPnl')"
          prop="unrealizedPnl"
          min-width="120"
          show-overflow-tooltip
        />
        <el-table-column
          :label="t('common.actions')"
          align="center"
          width="100"
          fixed="right"
        >
          <template #default="{ row }">
            <el-button
              v-perm="'option:position:detail'"
              link
              type="primary"
              @click="showDetail(row)"
            >
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
import { optionService, type OptionPosition, type OptionPositionDetail } from '@/services'
import TenantSelect from '@/components/TenantSelect.vue'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'

const { t } = useI18n()
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)

const loading = ref(false)
const rows = ref<OptionPosition[]>([])
const detailVisible = ref(false)
const detailData = ref<OptionPositionDetail | OptionPosition | null>(null)
const query = reactive({
  tenantId: undefined as number | undefined,
  userId: undefined as number | undefined,
  contractId: undefined as number | undefined,
  limit: 20,
})

const loadList = async () => {
  loading.value = true
  try {
    const res = await optionService.listPositions({
      ...query,
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
  query.tenantId = undefined
  query.userId = undefined
  query.contractId = undefined
  query.limit = 100
  loadList()
}

const showDetail = async (row: OptionPosition) => {
  detailData.value =
    (await optionService.getPosition({ tenantId: row.tenantId, id: row.id })).data || row
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

<style scoped></style>
