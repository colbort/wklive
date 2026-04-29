<template>
  <div class="payment-page module-page">
    <div class="page-header"><h2>链上钱包账号</h2><div class="header-actions"><el-button @click="loadList">刷新</el-button><el-button type="primary" @click="openDialog()">新增账号</el-button></div></div>
    <el-card shadow="never" class="query-card">
      <el-form :model="query" inline label-width="90px">
        <el-form-item label="租户ID"><el-input-number v-model="query.tenantId" :min="0" :precision="0" /></el-form-item>
        <el-form-item label="关键字"><el-input v-model="query.keyword" clearable /></el-form-item>
        <el-form-item label="服务商"><el-input v-model="query.provider" clearable /></el-form-item>
        <el-form-item label="状态"><el-select v-model="query.status" clearable style="width: 140px"><el-option label="启用" :value="1" /><el-option label="禁用" :value="2" /></el-select></el-form-item>
        <el-form-item><el-button type="primary" @click="loadList">搜索</el-button><el-button @click="resetQuery">重置</el-button></el-form-item>
      </el-form>
    </el-card>
    <el-card shadow="never">
      <el-table v-loading="loading" :data="list" stripe>
        <el-table-column prop="id" label="ID" width="80" /><el-table-column prop="tenantId" label="租户" width="90" /><el-table-column prop="accountCode" label="账号编码" min-width="150" /><el-table-column prop="accountName" label="账号名称" min-width="150" /><el-table-column prop="provider" label="服务商" width="120" /><el-table-column prop="isDefault" label="默认" width="80" /><el-table-column prop="status" label="状态" width="80" /><el-table-column label="操作" width="140" fixed="right"><template #default="{ row }"><el-button link type="primary" @click="showDetail(row)">详情</el-button><el-button link type="primary" @click="openDialog(row)">编辑</el-button></template></el-table-column>
      </el-table>
    </el-card>
    <el-dialog v-model="dialogVisible" :title="form.id ? '编辑钱包账号' : '新增钱包账号'" width="680px">
      <el-form label-width="130px">
        <el-form-item label="租户ID"><el-input-number v-model="form.tenantId" :min="0" :precision="0" /></el-form-item>
        <el-form-item label="账号编码"><el-input v-model="form.accountCode" :disabled="Boolean(form.id)" /></el-form-item>
        <el-form-item label="账号名称"><el-input v-model="form.accountName" /></el-form-item>
        <el-form-item label="服务商"><el-input v-model="form.provider" /></el-form-item>
        <el-form-item label="API Key"><el-input v-model="form.apiKeyCipher" type="textarea" /></el-form-item>
        <el-form-item label="API Secret"><el-input v-model="form.apiSecretCipher" type="textarea" /></el-form-item>
        <el-form-item label="回调密钥"><el-input v-model="form.callbackSecretCipher" type="textarea" /></el-form-item>
        <el-form-item label="扩展配置"><el-input v-model="form.extConfig" type="textarea" /></el-form-item>
        <el-form-item label="状态"><el-select v-model="form.status" style="width: 100%"><el-option label="启用" :value="1" /><el-option label="禁用" :value="2" /></el-select></el-form-item>
        <el-form-item label="默认"><el-switch v-model="form.isDefault" :active-value="1" :inactive-value="0" /></el-form-item>
      </el-form>
      <template #footer><el-button @click="dialogVisible = false">取消</el-button><el-button type="primary" @click="submit">确认</el-button></template>
    </el-dialog>
    <el-dialog v-model="detailVisible" title="账号详情" width="760px"><pre class="detail-pre">{{ JSON.stringify(detailData, null, 2) }}</pre></el-dialog>
  </div>
</template>
<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { cryptoService, type CryptoWalletAccount } from '@/services'
const loading = ref(false), dialogVisible = ref(false), detailVisible = ref(false)
const list = ref<CryptoWalletAccount[]>([]), detailData = ref<CryptoWalletAccount | null>(null)
const query = reactive({ tenantId: 0, keyword: '', provider: '', status: undefined as number | undefined })
const form = reactive({ id: 0, tenantId: 0, accountCode: '', accountName: '', provider: 'self', apiKeyCipher: '', apiSecretCipher: '', callbackSecretCipher: '', extConfig: '', status: 1, isDefault: 0, createTimes: 0, updateTimes: 0 })
function params() { return Object.fromEntries(Object.entries(query).filter(([, v]) => v !== '' && v !== 0 && v !== undefined)) }
async function loadList() { loading.value = true; try { list.value = (await cryptoService.listWalletAccounts({ ...params(), limit: 100 })).data || [] } finally { loading.value = false } }
function resetQuery() { Object.assign(query, { tenantId: 0, keyword: '', provider: '', status: undefined }); void loadList() }
function openDialog(row?: CryptoWalletAccount) { Object.assign(form, row || { id: 0, tenantId: 0, accountCode: '', accountName: '', provider: 'self', apiKeyCipher: '', apiSecretCipher: '', callbackSecretCipher: '', extConfig: '', status: 1, isDefault: 0, createTimes: 0, updateTimes: 0 }); dialogVisible.value = true }
function showDetail(row: CryptoWalletAccount) { detailData.value = row; detailVisible.value = true }
async function submit() { const res = form.id ? await cryptoService.updateWalletAccount(form) : await cryptoService.createWalletAccount(form); if (res.code === 200 || res.code === 0) { ElMessage.success('操作成功'); dialogVisible.value = false; await loadList() } }
onMounted(loadList)
</script>
