<template>
  <div class="payment-page">
    <div class="page-header">
      <h2>{{ t('payment.withdrawOrders') }}</h2>
      <el-button @click="loadList"> {{ t('common.refresh') }} </el-button>
    </div>

    <el-card shadow="never" class="query-card">
      <el-form :model="query" inline label-width="90px">
        <el-form-item :label="t('common.tenantId')">
          <el-input-number v-model="query.tenantId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('common.userId')">
          <el-input-number v-model="query.userId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('payment.orderNo')">
          <el-input v-model="query.orderNo" clearable />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadList"> {{ t('common.search') }} </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="never">
      <el-table v-loading="loading" :data="list" stripe>
        <el-table-column prop="orderNo" :label="t('payment.orderNo')" min-width="180" />
        <el-table-column prop="tenantId" :label="t('common.tenantId')" width="100" />
        <el-table-column prop="userId" :label="t('common.userId')" width="100" />
        <el-table-column prop="platformId" :label="t('payment.platformId')" width="100" />
        <el-table-column prop="channelId" :label="t('payment.channelId')" width="100" />
        <el-table-column prop="currency" :label="t('payment.currency')" width="80" />
        <el-table-column prop="amount" :label="t('payment.amount')" min-width="120" />
        <el-table-column prop="feeAmount" :label="t('payment.feeAmount')" min-width="100" />
        <el-table-column prop="status" :label="t('common.status')" width="90" />
        <el-table-column :label="t('common.actions')" width="180">
          <template #default="{ row }">
            <el-button link type="primary" @click="showDetail(row)"> {{ t('common.detail') }} </el-button>
            <el-button link type="warning" @click="openAudit(row)"> {{ t('payment.auditWithdraw') }} </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="detailVisible" :title="t('payment.orderDetail')" width="760px">
      <pre class="detail-pre">{{ JSON.stringify(detailData, null, 2) }}</pre>
    </el-dialog>

    <el-dialog v-model="auditVisible" :title="t('payment.auditWithdraw')" width="520px">
      <el-form label-width="110px">
        <el-form-item :label="t('payment.auditResult')">
          <el-radio-group v-model="auditForm.approve">
            <el-radio :value="1"> {{ t('payment.approved') }} </el-radio>
            <el-radio :value="2"> {{ t('payment.rejected') }} </el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="t('common.remark')">
          <el-input v-model="auditForm.remark" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="auditVisible = false"> {{ t('common.cancel') }} </el-button>
        <el-button type="primary" @click="submitAudit"> {{ t('common.confirm') }} </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { withdrawService, type WithdrawOrder } from '@/services'

const { t } = useI18n()

const loading = ref(false)
const list = ref<WithdrawOrder[]>([])
const detailVisible = ref(false)
const detailData = ref<Record<string, any>>({})
const auditVisible = ref(false)
const currentOrder = ref<WithdrawOrder | null>(null)

const auditForm = reactive({
  approve: 1,
  remark: '',
})

const query = reactive({
  tenantId: 0,
  userId: 0,
  orderNo: '',
})

const loadList = async () => {
  loading.value = true
  try {
    const res = await withdrawService.getWithdrawOrderList({
      ...query,
      tenantId: query.tenantId || undefined,
      userId: query.userId || undefined,
      orderNo: query.orderNo || undefined,
      limit: 100,
    })
    list.value = res.data || []
  } finally {
    loading.value = false
  }
}

const showDetail = async (row: WithdrawOrder) => {
  const res = await withdrawService.getWithdrawOrderDetail(row.orderNo, row.tenantId)
  detailData.value = res.data || row
  detailVisible.value = true
}

const openAudit = (row: WithdrawOrder) => {
  currentOrder.value = row
  Object.assign(auditForm, { approve: 1, remark: '' })
  auditVisible.value = true
}

const submitAudit = async () => {
  if (!currentOrder.value) return
  await withdrawService.auditWithdrawOrder({
    tenantId: currentOrder.value.tenantId,
    orderNo: currentOrder.value.orderNo,
    approve: auditForm.approve,
    remark: auditForm.remark,
  })
  ElMessage.success(t('payment.auditSuccess'))
  auditVisible.value = false
  loadList()
}

onMounted(loadList)
</script>

<style scoped></style>
