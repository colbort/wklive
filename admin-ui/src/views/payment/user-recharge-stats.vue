<template>
  <div class="payment-page module-page">
    <div class="page-header">
      <h2>{{ t('payment.userRechargeStats') }}</h2>
      <div class="header-actions">
        <el-button @click="loadList"> {{ t('common.refresh') }} </el-button>
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
          <el-button type="primary" @click="loadList"> {{ t('common.search') }} </el-button>
        </el-form-item>
      </el-form>
    </el-card>
    <el-card shadow="never" class="table-card">
      <el-table v-loading="loading" :data="list" stripe>
        <el-table-column prop="id" :label="t('common.id')" width="80" />
        <el-table-column prop="tenantId" :label="t('common.tenantId')" width="100" />
        <el-table-column prop="userId" :label="t('common.userId')" width="100" />
        <el-table-column prop="successOrderCount" :label="t('payment.successOrderCount')" width="120" />
        <el-table-column prop="successTotalAmount" :label="t('payment.successTotalAmount')" min-width="120" />
        <el-table-column prop="todaySuccessAmount" :label="t('payment.todaySuccessAmount')" min-width="120" />
        <el-table-column prop="todaySuccessCount" :label="t('payment.todaySuccessCount')" width="100" />
        <el-table-column :label="t('common.actions')" width="100">
          <template #default="{ row }">
            <el-button link type="primary" @click="showDetail(row)"> {{ t('common.detail') }} </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
    <el-dialog v-model="detailVisible" :title="t('payment.detailTitle')" width="680px">
      <pre class="detail-pre">{{ JSON.stringify(detailData, null, 2) }}</pre>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { rechargeService, type UserRechargeStat } from '@/services'

const { t } = useI18n()

const loading = ref(false)
const list = ref<UserRechargeStat[]>([])
const detailVisible = ref(false)
const detailData = ref<Record<string, any>>({})
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
      limit: 100,
    })
    list.value = res.data || []
  } finally {
    loading.value = false
  }
}

const showDetail = async (row: UserRechargeStat) => {
  const res = await rechargeService.getUserRechargeStat({
    tenantId: row.tenantId,
    userId: row.userId,
  })
  detailData.value = res.data || row
  detailVisible.value = true
}

onMounted(loadList)
</script>

<style scoped></style>
