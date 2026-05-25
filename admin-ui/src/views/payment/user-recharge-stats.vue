<template>
  <div class="payment-page module-page">
    <div class="page-header">
      <h2>{{ t('payment.userRechargeStats') }}</h2>
      <div class="header-actions">
        <el-button @click="loadList">
          {{ t('common.refresh') }}
        </el-button>
      </div>
    </div>
    <el-card shadow="never" class="query-card">
      <el-form :model="query" inline label-width="120px">
        <el-form-item :label="t('common.tenantId')">
          <el-input-number v-model="query.tenantId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('common.userId')">
          <el-input-number v-model="query.userId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('payment.successTotalAmountMin')">
          <el-input-number v-model="query.successTotalAmountMin" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('payment.successTotalAmountMax')">
          <el-input-number v-model="query.successTotalAmountMax" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadList">
            {{ t('common.search') }}
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>
    <el-card shadow="never" class="table-card">
      <el-table v-loading="loading" :data="list" stripe>
        <el-table-column prop="id" :label="t('common.id')" width="80" />
        <el-table-column prop="tenantId" :label="t('common.tenantId')" width="100" />
        <el-table-column prop="userId" :label="t('common.userId')" width="100" />
        <el-table-column
          prop="successOrderCount"
          :label="t('payment.successOrderCount')"
          width="120"
        />
        <el-table-column
          prop="successTotalAmount"
          :label="t('payment.successTotalAmount')"
          min-width="120"
        />
        <el-table-column
          prop="todaySuccessAmount"
          :label="t('payment.todaySuccessAmount')"
          min-width="120"
        />
        <el-table-column
          prop="todaySuccessCount"
          :label="t('payment.todaySuccessCount')"
          width="100"
        />
        <el-table-column :label="t('common.actions')" width="100">
          <template #default="{ row }">
            <el-button link type="primary" @click="showDetail(row)">
              {{ t('common.detail') }}
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
    <el-dialog v-model="detailVisible" :title="t('payment.userRechargeStatDetail')" width="720px">
      <el-descriptions v-if="detailData" :column="2" border>
        <el-descriptions-item :label="t('common.id')">
          {{ detailData.id }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('common.tenantId')">
          {{ detailData.tenantId }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('common.userId')">
          {{ detailData.userId }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('payment.successOrderCount')">
          {{ detailData.successOrderCount }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('payment.successTotalAmount')">
          {{ detailData.successTotalAmount }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('payment.todaySuccessAmount')">
          {{ detailData.todaySuccessAmount }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('payment.todaySuccessCount')">
          {{ detailData.todaySuccessCount }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('payment.firstSuccessTime')">
          {{ formatDate(detailData.firstSuccessTime) }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('payment.lastSuccessTime')">
          {{ formatDate(detailData.lastSuccessTime) }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('common.createTimes')">
          {{ formatDate(detailData.createTimes) }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('common.updateTimes')">
          {{ formatDate(detailData.updateTimes) }}
        </el-descriptions-item>
      </el-descriptions>
      <el-empty v-else :description="t('common.noData')" />
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { usePagination } from '@/composables'
import { ElMessage } from 'element-plus'
import { rechargeService, type UserRechargeStat } from '@/services'
import { formatDate } from '@/utils'

const { t } = useI18n()
const {
  pagination,
  updateFromResponse,
  resetAndLoad,
  prevAndLoad,
  nextAndLoad,
} = usePagination<number>(20)

const loading = ref(false)
const list = ref<UserRechargeStat[]>([])
const detailVisible = ref(false)
const detailData = ref<UserRechargeStat | null>(null)
const query = reactive({
  tenantId: 0,
  userId: 0,
  successTotalAmountMin: 0,
  successTotalAmountMax: 0,
})

const loadList = async () => {
  loading.value = true
  try {
    const res = await rechargeService.getUserRechargeStatList({
      ...query,
      tenantId: query.tenantId || undefined,
      userId: query.userId || undefined,
      successTotalAmountMin: query.successTotalAmountMin || undefined,
      successTotalAmountMax: query.successTotalAmountMax || undefined,
      cursor: pagination.cursor,
      limit: pagination.limit,
    })
    list.value = res.data || []
    updateFromResponse(res)
  } finally {
    loading.value = false
  }
}

const showDetail = async (row: UserRechargeStat) => {
  detailData.value = row
  detailVisible.value = true
  try {
    const res = await rechargeService.getUserRechargeStat({
      tenantId: row.tenantId,
      userId: row.userId,
    })
    detailData.value = res.data || row
  } catch (error: unknown) {
    ElMessage.error(error instanceof Error ? error.message : t('common.loadFailed'))
  }
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
