<template>
  <div class="module-page">
    <CrudQueryCard :model="query" @search="loadList" @reset="resetQuery">
      <el-form-item :label="t('option.tenantId')">
        <TenantSelect v-model="query.tenantId" class="tenant-select-filter" />
      </el-form-item>
      <el-form-item :label="t('option.contractId')">
        <ContractSelect v-model="query.contractId" :tenant-id="query.tenantId || undefined" />
      </el-form-item>
      <el-form-item :label="t('option.tradeNo')">
        <el-input v-model="query.tradeNo" clearable />
      </el-form-item>
    </CrudQueryCard>

    <el-card shadow="never" class="table-card">
      <el-table v-loading="loading" :data="rows" stripe>
        <el-table-column
          :label="t('option.tradeNo')"
          prop="tradeNo"
          min-width="180"
          show-overflow-tooltip
        />
        <el-table-column :label="t('option.contractId')" prop="contractId" width="100" />
        <el-table-column
          :label="t('option.price')"
          prop="price"
          min-width="120"
          show-overflow-tooltip
        />
        <el-table-column
          :label="t('option.qty')"
          prop="qty"
          min-width="120"
          show-overflow-tooltip
        />
        <el-table-column
          :label="t('option.tradeTime')"
          prop="tradeTime"
          min-width="160"
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
              v-perm="'option:trade:detail'"
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
import { optionService, type OptionTrade, type OptionTradeDetail } from '@/services'
import TenantSelect from '@/components/TenantSelect.vue'
import ContractSelect from '@/components/ContractSelect.vue'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'

const { t } = useI18n()
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)

const loading = ref(false)
const rows = ref<OptionTrade[]>([])
const detailVisible = ref(false)
const detailData = ref<OptionTradeDetail | OptionTrade | null>(null)
const query = reactive({
  tenantId: undefined as number | undefined,
  contractId: undefined as number | undefined,
  tradeNo: '',
  limit: 20,
})

const loadList = async () => {
  loading.value = true
  try {
    const res = await optionService.listTrades({
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
  query.contractId = undefined
  query.tradeNo = ''
  query.limit = 100
  loadList()
}

const showDetail = async (row: OptionTrade) => {
  detailData.value =
    (await optionService.getTrade({ tenantId: row.tenantId, id: row.id, tradeNo: row.tradeNo }))
      .data || row
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
