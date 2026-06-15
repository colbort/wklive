<template>
  <div class="module-page">
    <CrudQueryCard
      :model="query"
      label-width="auto"
      @search="loadRows"
      @reset="resetQuery"
    >
      <el-form-item :label="t('staking.tenantId')">
        <TenantSelect v-model="query.tenantId" class="tenant-select-filter" />
      </el-form-item>
      <el-form-item :label="t('staking.orderNo')">
        <el-input v-model="query.orderNo" clearable />
      </el-form-item>
      <el-form-item :label="t('staking.userId')">
        <el-input-number v-model="query.userId" :min="0" :precision="0" />
      </el-form-item>
    </CrudQueryCard>

    <el-card shadow="never" class="table-card">
      <el-table v-loading="loading" :data="rows" stripe>
        <el-table-column
          :label="t('staking.orderNo')"
          prop="orderNo"
          min-width="180"
          show-overflow-tooltip
        />
        <el-table-column :label="t('staking.userId')" prop="userId" width="100" />
        <el-table-column
          prop="rewardAmount"
          :label="t('staking.rewardAmount')"
          min-width="120"
          show-overflow-tooltip
        />
        <el-table-column :label="t('staking.rewardType')" prop="rewardType" width="100" />
        <el-table-column :label="t('staking.rewardStatus')" prop="rewardStatus" width="100" />
        <el-table-column :label="t('common.actions')" width="100" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="showDetail(row)">
              {{ t('itick.detail') }}
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

    <el-dialog v-model="detailVisible" :title="t('itick.detail')" width="760px">
      <pre class="detail-pre">{{ JSON.stringify(detailData, null, 2) }}</pre>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { stakingService, type StakeRewardLog } from '@/services'
import { useI18n } from 'vue-i18n'
import { usePagination } from '@/composables'
import TenantSelect from '@/components/TenantSelect.vue'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'

const { t } = useI18n()
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)

const loading = ref(false)
const rows = ref<StakeRewardLog[]>([])
const detailVisible = ref(false)
const detailData = ref<StakeRewardLog | null>(null)
const query = reactive({
  tenantId: undefined as number | undefined,
  orderNo: '',
  userId: undefined as number | undefined,
  productId: undefined as number | undefined,
  limit: 20,
})

const loadRows = async () => {
  loading.value = true
  try {
    const res = await stakingService.listRewardLogs({
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
  query.orderNo = ''
  query.userId = undefined
  query.productId = undefined
  query.limit = 100
  loadRows()
}

const showDetail = (row: StakeRewardLog) => {
  detailData.value = row
  detailVisible.value = true
}

function handleLimitChange() {
  resetAndLoad(loadRows)
}

function handlePrevPage() {
  prevAndLoad(loadRows)
}

function handleNextPage() {
  nextAndLoad(loadRows)
}

onMounted(loadRows)
</script>
