<script setup lang="ts">
/**
 * 租户提现订单：固定当前租户，查看订单详情并做审核通过/拒绝。
 */
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { withdrawService, type WithdrawOrder } from '@/services'
import { useTenantStore } from '@/stores/tenant'
import { formatDate } from '@/utils'

const tenant = useTenantStore()
const loading = ref(false)
const list = ref<WithdrawOrder[]>([])
const detailVisible = ref(false)
const detailData = ref<WithdrawOrder | null>(null)
const auditVisible = ref(false)
const currentOrder = ref<WithdrawOrder | null>(null)

const auditForm = reactive({
  approve: 1,
  remark: '',
})

const query = reactive({
  userId: 0,
  orderNo: '',
})

const loadList = async () => {
  await tenant.ensureLoaded()
  loading.value = true
  try {
    const res = await withdrawService.getWithdrawOrderList({
      tenantId: tenant.tenantId,
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
  const res = await withdrawService.getWithdrawOrderDetail(row.orderNo, tenant.tenantId)
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
    tenantId: tenant.tenantId,
    orderNo: currentOrder.value.orderNo,
    approve: auditForm.approve,
    remark: auditForm.remark,
  })
  ElMessage.success('提现审核已提交')
  auditVisible.value = false
  loadList()
}

onMounted(loadList)
</script>

<template>
  <div class="page">
    <div class="page-header">
      <div>
        <h2>提现订单</h2>
        <p>当前租户：{{ tenant.tenantName || tenant.tenantCode }}</p>
      </div>
      <el-button @click="loadList">刷新</el-button>
    </div>

    <el-card shadow="never">
      <el-form :model="query" inline>
        <el-form-item label="用户ID">
          <el-input-number v-model="query.userId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item label="平台单号">
          <el-input v-model="query.orderNo" clearable />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadList">查询</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="never">
      <el-table v-loading="loading" :data="list" stripe>
        <el-table-column prop="orderNo" label="平台单号" min-width="180" />
        <el-table-column prop="bizOrderNo" label="业务单号" min-width="180" />
        <el-table-column prop="userId" label="用户ID" width="100" />
        <el-table-column prop="currency" label="币种" width="90" />
        <el-table-column prop="amount" label="提现金额" min-width="120" />
        <el-table-column prop="feeAmount" label="手续费" min-width="100" />
        <el-table-column prop="status" label="状态" width="100" />
        <el-table-column label="创建时间" min-width="170">
          <template #default="{ row }">{{ formatDate(row.createTimes) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="showDetail(row)">详情</el-button>
            <el-button link type="warning" @click="openAudit(row)">审核</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="detailVisible" title="提现订单详情" width="760px">
      <pre class="detail-pre">{{ JSON.stringify(detailData, null, 2) }}</pre>
    </el-dialog>

    <el-dialog v-model="auditVisible" title="审核提现订单" width="520px">
      <el-form label-width="100px">
        <el-form-item label="审核结果">
          <el-radio-group v-model="auditForm.approve">
            <el-radio :value="1">通过</el-radio>
            <el-radio :value="2">拒绝</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="auditForm.remark" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="auditVisible = false">取消</el-button>
        <el-button type="primary" @click="submitAudit">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.page { display: grid; gap: 16px; }
.page-header { display: flex; align-items: center; justify-content: space-between; }
.page-header h2 { margin: 0 0 6px; }
.page-header p { margin: 0; color: #909399; }
.detail-pre { margin: 0; white-space: pre-wrap; word-break: break-all; }
</style>
