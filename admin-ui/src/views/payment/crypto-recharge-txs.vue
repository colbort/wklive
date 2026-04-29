<template>
  <div class="payment-page module-page">
    <div class="page-header"><h2>链上充值交易</h2><div class="header-actions"><el-button @click="loadList">刷新</el-button><el-button type="primary" @click="openDialog()">新增交易</el-button></div></div>
    <el-card shadow="never" class="query-card">
      <el-form :model="query" inline label-width="90px">
        <el-form-item label="租户ID"><el-input-number v-model="query.tenantId" :min="0" :precision="0" /></el-form-item><el-form-item label="用户ID"><el-input-number v-model="query.userId" :min="0" :precision="0" /></el-form-item><el-form-item label="订单号"><el-input v-model="query.orderNo" clearable /></el-form-item><el-form-item label="币种"><el-input v-model="query.coin" clearable /></el-form-item><el-form-item label="链"><el-input v-model="query.chainCode" clearable /></el-form-item><el-form-item label="Hash"><el-input v-model="query.txHash" clearable /></el-form-item><el-form-item><el-button type="primary" @click="loadList">搜索</el-button><el-button @click="resetQuery">重置</el-button></el-form-item>
      </el-form>
    </el-card>
    <el-card shadow="never"><el-table v-loading="loading" :data="list" stripe><el-table-column prop="id" label="ID" width="80" /><el-table-column prop="tenantId" label="租户" width="90" /><el-table-column prop="userId" label="用户" width="100" /><el-table-column prop="orderNo" label="订单号" min-width="160" /><el-table-column prop="coin" label="币种" width="90" /><el-table-column prop="chainCode" label="链" width="100" /><el-table-column prop="txHash" label="TxHash" min-width="240" show-overflow-tooltip /><el-table-column prop="amount" label="数量" width="120" /><el-table-column prop="confirmCount" label="确认数" width="90" /><el-table-column prop="status" label="状态" width="90" /><el-table-column label="操作" width="140" fixed="right"><template #default="{ row }"><el-button link type="primary" @click="showDetail(row)">详情</el-button><el-button link type="primary" @click="openDialog(row)">编辑</el-button></template></el-table-column></el-table></el-card>
    <el-dialog v-model="dialogVisible" :title="form.id ? '编辑充值交易' : '新增充值交易'" width="720px">
      <el-form label-width="130px">
        <el-form-item label="租户ID"><el-input-number v-model="form.tenantId" :min="0" :precision="0" /></el-form-item>
        <template v-if="!form.id"><el-form-item label="用户ID"><el-input-number v-model="form.userId" :min="0" :precision="0" /></el-form-item><el-form-item label="币种"><el-input v-model="form.coin" /></el-form-item><el-form-item label="链"><el-input v-model="form.chainCode" /></el-form-item><el-form-item label="TxHash"><el-input v-model="form.txHash" /></el-form-item><el-form-item label="付款地址"><el-input v-model="form.fromAddress" /></el-form-item><el-form-item label="收款地址"><el-input v-model="form.toAddress" /></el-form-item><el-form-item label="金额"><el-input v-model="form.amount" /></el-form-item></template>
        <el-form-item label="订单ID"><el-input-number v-model="form.orderId" :min="0" :precision="0" /></el-form-item>
        <el-form-item label="订单号"><el-input v-model="form.orderNo" /></el-form-item>
        <el-form-item label="确认数"><el-input-number v-model="form.confirmCount" :min="0" :precision="0" /></el-form-item>
        <el-form-item label="要求确认数"><el-input-number v-model="form.requiredConfirmCount" :min="0" :precision="0" /></el-form-item>
        <el-form-item label="状态"><el-select v-model="form.status" style="width: 100%"><el-option label="待确认" :value="1" /><el-option label="确认中" :value="2" /><el-option label="已确认" :value="3" /><el-option label="失败" :value="4" /><el-option label="已入账" :value="5" /></el-select></el-form-item>
        <el-form-item label="原始数据"><el-input v-model="form.rawData" type="textarea" /></el-form-item>
      </el-form>
      <template #footer><el-button @click="dialogVisible = false">取消</el-button><el-button type="primary" @click="submit">确认</el-button></template>
    </el-dialog>
    <el-dialog v-model="detailVisible" title="交易详情" width="780px"><pre class="detail-pre">{{ JSON.stringify(detailData, null, 2) }}</pre></el-dialog>
  </div>
</template>
<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { cryptoService, type CryptoRechargeTx } from '@/services'
const loading = ref(false), dialogVisible = ref(false), detailVisible = ref(false)
const list = ref<CryptoRechargeTx[]>([]), detailData = ref<CryptoRechargeTx | null>(null)
const query = reactive({ tenantId: 0, userId: 0, orderNo: '', coin: '', chainCode: '', txHash: '' })
const form = reactive({ id: 0, tenantId: 0, userId: 0, orderId: 0, orderNo: '', coin: 'USDT', chainCode: 'TRC20', txHash: '', fromAddress: '', toAddress: '', memo: '', amount: '0', blockHeight: 0, confirmCount: 0, requiredConfirmCount: 0, status: 1, rawData: '', createTimes: 0, updateTimes: 0 })
function params() { return Object.fromEntries(Object.entries(query).filter(([, v]) => v !== '' && v !== 0)) }
async function loadList() { loading.value = true; try { list.value = (await cryptoService.listRechargeTxs({ ...params(), limit: 100 })).data || [] } finally { loading.value = false } }
function resetQuery() { Object.assign(query, { tenantId: 0, userId: 0, orderNo: '', coin: '', chainCode: '', txHash: '' }); void loadList() }
function openDialog(row?: CryptoRechargeTx) { Object.assign(form, row || { id: 0, tenantId: 0, userId: 0, orderId: 0, orderNo: '', coin: 'USDT', chainCode: 'TRC20', txHash: '', fromAddress: '', toAddress: '', memo: '', amount: '0', blockHeight: 0, confirmCount: 0, requiredConfirmCount: 0, status: 1, rawData: '', createTimes: 0, updateTimes: 0 }); dialogVisible.value = true }
function showDetail(row: CryptoRechargeTx) { detailData.value = row; detailVisible.value = true }
async function submit() { const res = form.id ? await cryptoService.updateRechargeTx(form) : await cryptoService.createRechargeTx(form); if (res.code === 200 || res.code === 0) { ElMessage.success('操作成功'); dialogVisible.value = false; await loadList() } }
onMounted(loadList)
</script>
