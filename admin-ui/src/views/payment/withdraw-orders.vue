<template>
  <div class="payment-page">
    <div class="page-header">
      <h2>提现订单</h2>
      <el-button @click="loadList">
        刷新
      </el-button>
    </div>

    <el-card shadow="never" class="query-card">
      <el-form :model="query" inline label-width="90px">
        <el-form-item label="租户ID">
          <el-input-number v-model="query.tenantId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item label="用户ID">
          <el-input-number v-model="query.userId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item label="订单号">
          <el-input v-model="query.orderNo" clearable />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadList">
            查询
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="never">
      <el-table v-loading="loading" :data="list" stripe>
        <el-table-column prop="orderNo" label="订单号" min-width="180" />
        <el-table-column prop="tenantId" label="租户ID" width="100" />
        <el-table-column prop="userId" label="用户ID" width="100" />
        <el-table-column prop="platformId" label="平台ID" width="100" />
        <el-table-column prop="channelId" label="通道ID" width="100" />
        <el-table-column prop="currency" label="币种" width="80" />
        <el-table-column prop="amount" label="提现金额" min-width="120" />
        <el-table-column prop="feeAmount" label="手续费" min-width="100" />
        <el-table-column prop="status" label="状态" width="90" />
        <el-table-column label="操作" width="180">
          <template #default="{ row }">
            <el-button link type="primary" @click="showDetail(row)">
              详情
            </el-button>
            <el-button link type="warning" @click="openAudit(row)">
              审核
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="detailVisible" title="订单详情" width="760px">
      <pre class="detail-pre">{{ JSON.stringify(detailData, null, 2) }}</pre>
    </el-dialog>

    <el-dialog v-model="auditVisible" title="审核提现" width="520px">
      <el-form label-width="110px">
        <el-form-item label="审核结果">
          <el-radio-group v-model="auditForm.approve">
            <el-radio :value="1">
              通过
            </el-radio>
            <el-radio :value="2">
              拒绝
            </el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="auditForm.remark" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="auditVisible = false">
          取消
        </el-button>
        <el-button type="primary" @click="submitAudit">
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { withdrawService, type WithdrawOrder } from '@/services'

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
  ElMessage.success('审核成功')
  auditVisible.value = false
  loadList()
}

onMounted(loadList)
</script>

<style scoped></style>
