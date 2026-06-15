<template>
  <div class="module-page">
    <CrudQueryCard
      :model="query"
      label-width="auto"
      @search="loadCurrent"
      @reset="resetCurrent"
    >
      <el-form-item :label="t('option.tenantId')">
        <TenantSelect v-model="query.tenantId" class="tenant-select-filter" />
      </el-form-item>
      <el-form-item :label="t('option.contractId')">
        <el-input-number v-model="query.contractId" :min="0" :precision="0" />
      </el-form-item>
      <el-form-item :label="t('option.settlementNo')">
        <el-input v-model="query.settlementNo" clearable />
      </el-form-item>
    </CrudQueryCard>

    <el-card shadow="never" class="table-card">
      <el-table v-loading="loading" :data="rows" stripe>
        <el-table-column
          :label="t('option.settlementNo')"
          prop="settlementNo"
          min-width="180"
          show-overflow-tooltip
        />
        <el-table-column :label="t('option.contractId')" prop="contractId" width="100" />
        <el-table-column
          :label="t('option.deliveryPrice')"
          prop="deliveryPrice"
          min-width="120"
          show-overflow-tooltip
        />
        <el-table-column :label="t('common.status')" prop="status" width="100" />
        <el-table-column :label="t('common.actions')" width="100" fixed="right">
          <template #default="{ row }">
            <el-button
              v-perm="'option:settlement:detail'"
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
import { optionService, type OptionSettlement, type OptionSettlementDetail } from '@/services'
import TenantSelect from '@/components/TenantSelect.vue'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'

const { t } = useI18n()
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)

const loading = ref(false)
const rows = ref<OptionSettlement[]>([])
const detailVisible = ref(false)
const detailData = ref<OptionSettlementDetail | OptionSettlement | null>(null)
const query = reactive({
  tenantId: undefined as number | undefined,
  contractId: undefined as number | undefined,
  settlementNo: '',
  status: undefined as number | undefined,
  limit: 20,
})

const loadCurrent = async () => {
  loading.value = true
  try {
    const res = await optionService.listSettlements({
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

const resetCurrent = () => {
  query.tenantId = undefined
  query.contractId = undefined
  query.settlementNo = ''
  query.status = undefined
  query.limit = 100
  loadCurrent()
}

const showDetail = async (row: OptionSettlement) => {
  detailData.value =
    (
      await optionService.getSettlement({
        tenantId: row.tenantId,
        id: row.id,
        settlementNo: row.settlementNo,
      })
    ).data || row
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

<style scoped></style>
