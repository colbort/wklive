<template>
  <div class="payment-page">
    <div class="page-header">
      <h2>{{ t('payment.rechargeNotifyLogs') }}</h2>
      <el-button @click="loadList">
        {{ t('common.refresh') }}
      </el-button>
    </div>
    <el-card shadow="never" class="query-card">
      <el-form :model="query" inline label-width="100px">
        <el-form-item :label="t('common.tenantId')">
          <el-input-number v-model="query.tenantId" :min="0" :precision="0" />
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
      </el-form>
    </el-card>
    <el-card shadow="never">
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
            <el-button link type="primary" @click="showDetail(row)">
              {{ t('common.detail') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
    <el-dialog v-model="detailVisible" :title="t('payment.logDetail')" width="760px">
      <pre class="detail-pre">{{ JSON.stringify(detailData, null, 2) }}</pre>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { rechargeService, type PayNotifyLog } from '@/services'

const { t } = useI18n()

const loading = ref(false)
const list = ref<PayNotifyLog[]>([])
const detailVisible = ref(false)
const detailData = ref<PayNotifyLog | null>(null)
const query = reactive({ tenantId: 0, orderNo: '', notifyStatus: 0 })

const loadList = async () => {
  loading.value = true
  try {
    const res = await rechargeService.getRechargeNotifyLogList({
      ...query,
      tenantId: query.tenantId || undefined,
      orderNo: query.orderNo || undefined,
      notifyStatus: query.notifyStatus || undefined,
      limit: 100,
    })
    list.value = res.data || []
  } finally {
    loading.value = false
  }
}

const showDetail = async (row: PayNotifyLog) => {
  const res = await rechargeService.getRechargeNotifyLogDetail(row.id, row.tenantId)
  detailData.value = res.data || row
  detailVisible.value = true
}

onMounted(loadList)
</script>

<style scoped></style>
