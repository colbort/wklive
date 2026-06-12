<template>
  <div class="payment-page">
    <div class="page-header">
      <h2>{{ t('payment.withdrawNotifyLogs') }}</h2>
      <el-button @click="loadList">
        {{ t('common.refresh') }}
      </el-button>
    </div>
    <CrudQueryCard :model="query" label-width="100px" :show-actions="false">
      <el-form-item :label="t('common.tenantId')">
        <TenantSelect v-model="query.tenantId" class="tenant-select-filter" />
      </el-form-item>
      <el-form-item :label="t('payment.orderNo')">
        <el-input v-model="query.orderNo" clearable />
      </el-form-item>
      <el-form-item :label="t('payment.notifyStatus')">
        <el-input-number v-model="query.notifyStatus" :min="0" :precision="0" />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="loadList">
          {{ t('common.search') }}
        </el-button>
      </el-form-item>
    </CrudQueryCard>
    <el-card shadow="never" class="table-card">
      <el-table v-loading="loading" :data="list" stripe>
        <el-table-column prop="id" :label="t('common.id')" width="80" />
        <el-table-column prop="tenantId" :label="t('common.tenantId')" width="100" />
        <el-table-column prop="orderNo" :label="t('payment.orderNo')" min-width="180" />
        <el-table-column prop="notifyStatus" :label="t('payment.notifyStatus')" width="100" />
        <el-table-column prop="signResult" :label="t('payment.signResult')" width="100" />
        <el-table-column
          prop="errorMessage"
          :label="t('payment.errorMessage')"
          min-width="200"
          show-overflow-tooltip
        />
        <el-table-column :label="t('common.actions')" width="100">
          <template #default="{ row }">
            <el-button
              v-perm="'payment:withdraw-notify-log:detail'"
              link
              type="primary"
              @click="showDetail(row)"
            >
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
    <el-dialog v-model="detailVisible" :title="t('payment.logDetail')" width="760px">
      <PaymentDetailDescriptions :data="detailData" />
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { usePagination } from '@/composables'
import { withdrawService, type PayNotifyLog } from '@/services'
import PaymentDetailDescriptions from '@/components/payment/PaymentDetailDescriptions.vue'
import TenantSelect from '@/components/TenantSelect.vue'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'

const { t } = useI18n()
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)

const loading = ref(false)
const list = ref<PayNotifyLog[]>([])
const detailVisible = ref(false)
const detailData = ref<PayNotifyLog | null>(null)
const query = reactive({ tenantId: 0, orderNo: '', notifyStatus: 0 })

const loadList = async () => {
  loading.value = true
  try {
    const res = await withdrawService.getWithdrawNotifyLogList({
      ...query,
      tenantId: query.tenantId || undefined,
      orderNo: query.orderNo || undefined,
      notifyStatus: query.notifyStatus || undefined,
      cursor: pagination.cursor,
      limit: pagination.limit,
    })
    list.value = res.data || []
    updateFromResponse(res)
  } finally {
    loading.value = false
  }
}

const showDetail = async (row: PayNotifyLog) => {
  const res = await withdrawService.getWithdrawNotifyLogDetail(row.id, row.tenantId)
  detailData.value = res.data || row
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
