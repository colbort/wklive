<template>
  <div class="payment-page">
    <div class="page-header">
      <h2>{{ t('payment.rechargeOrders') }}</h2>
      <el-button @click="loadList">
        {{ t('common.refresh') }}
      </el-button>
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
        <el-form-item :label="t('payment.bizOrderNo')">
          <el-input v-model="query.bizOrderNo" clearable />
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
        <el-table-column prop="orderNo" :label="t('payment.orderNo')" min-width="180" />
        <el-table-column prop="tenantId" :label="t('common.tenantId')" width="100" />
        <el-table-column prop="userId" :label="t('common.userId')" width="100" />
        <el-table-column prop="currency" :label="t('payment.currency')" width="80" />
        <el-table-column prop="orderAmount" :label="t('payment.orderAmount')" min-width="120" />
        <el-table-column prop="payAmount" :label="t('payment.payAmount')" min-width="120" />
        <el-table-column prop="status" :label="t('common.status')" width="90" />
        <el-table-column :label="t('common.actions')" width="260" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="showDetail(row)">
              {{ t('common.detail') }}
            </el-button>
            <el-button link type="warning" @click="closeOrder(row)">
              {{ t('payment.closeOrder') }}
            </el-button>
            <el-button link type="success" @click="openManualSuccess(row)">
              {{ t('payment.manualMarkSuccess') }}
            </el-button>
            <el-button link type="primary" @click="retryNotify(row)">
              {{ t('payment.retryNotify') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
    <el-dialog v-model="detailVisible" :title="t('payment.orderDetail')" width="780px">
      <pre class="detail-pre">{{ JSON.stringify(detailData, null, 2) }}</pre>
    </el-dialog>
    <el-dialog v-model="manualVisible" :title="t('payment.manualMarkSuccess')" width="520px">
      <el-form label-width="110px">
        <el-form-item :label="t('payment.thirdTradeNo')">
          <el-input v-model="manualForm.thirdTradeNo" />
        </el-form-item>
        <el-form-item :label="t('payment.payAmount')">
          <el-input-number
            v-model="manualForm.payAmount"
            :min="0"
            :precision="0"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item :label="t('common.remark')">
          <el-input v-model="manualForm.remark" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="manualVisible = false">
          {{ t('common.cancel') }}
        </el-button>
        <el-button type="primary" @click="submitManual">
          {{ t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { rechargeService, type RechargeOrder } from '@/services'

const { t } = useI18n()

const loading = ref(false)
const list = ref<RechargeOrder[]>([])
const detailVisible = ref(false)
const detailData = ref<RechargeOrder | null>(null)
const manualVisible = ref(false)
const currentOrder = ref<RechargeOrder | null>(null)
const manualForm = reactive({ thirdTradeNo: '', payAmount: 0, remark: '' })
const query = reactive({ tenantId: 0, userId: 0, orderNo: '', bizOrderNo: '' })

const loadList = async () => {
  loading.value = true
  try {
    const res = await rechargeService.getRechargeOrderList({
      ...query,
      tenantId: query.tenantId || undefined,
      userId: query.userId || undefined,
      orderNo: query.orderNo || undefined,
      bizOrderNo: query.bizOrderNo || undefined,
      limit: 100,
    })
    list.value = res.data || []
  } finally {
    loading.value = false
  }
}

const showDetail = async (row: RechargeOrder) => {
  const res = await rechargeService.getRechargeOrderDetail(row.orderNo, row.tenantId)
  detailData.value = res.data || row
  detailVisible.value = true
}

const closeOrder = async (row: RechargeOrder) => {
  await ElMessageBox.confirm(t('payment.confirmCloseRechargeOrder'), t('common.warning'))
  await rechargeService.closeRechargeOrder(row.orderNo, row.tenantId)
  ElMessage.success(t('common.operationSuccess'))
  loadList()
}

const openManualSuccess = (row: RechargeOrder) => {
  currentOrder.value = row
  Object.assign(manualForm, {
    thirdTradeNo: '',
    payAmount: row.payAmount || row.orderAmount || 0,
    remark: '',
  })
  manualVisible.value = true
}

const submitManual = async () => {
  if (!currentOrder.value) return
  await rechargeService.manualSuccessRechargeOrder(currentOrder.value.orderNo, {
    tenantId: currentOrder.value.tenantId,
    thirdTradeNo: manualForm.thirdTradeNo,
    payAmount: manualForm.payAmount,
    remark: manualForm.remark,
  })
  ElMessage.success(t('common.operationSuccess'))
  manualVisible.value = false
  loadList()
}

const retryNotify = async (row: RechargeOrder) => {
  await rechargeService.retryRechargeNotify(row.orderNo, row.tenantId)
  ElMessage.success(t('payment.retryNotifySubmitted'))
}

onMounted(loadList)
</script>

<style scoped></style>
