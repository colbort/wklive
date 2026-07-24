<script setup lang="ts">
import { onMounted, reactive, ref } from "vue";
import { ElMessage } from "element-plus";
import { liquidityApi } from "@/api/liquidity";
import type { Provider } from "@/types/liquidity";
const loading = ref(false), dialog = ref(false), rows = ref<Provider[]>([]);
const query = reactive({ tenantId: 1, keyword: "", status: "", limit: 20 });
const form = reactive({ tenantId: 1, providerCode: "", providerName: "", providerType: 1, tradeUserId: 0, venueCode: "", environment: 1, credentialRef: "", accountRef: "", rateLimitPerSecond: 10, status: 2, remark: "" });
async function load() { loading.value=true; try { const r=await liquidityApi.providers(query); rows.value=(r.data||[]) as unknown as Provider[]; } finally { loading.value=false; } }
async function create() { await liquidityApi.createProvider(form); ElMessage.success("提供方已创建"); dialog.value=false; await load(); }
async function test(row: Provider) { await liquidityApi.testProvider(row.id); ElMessage.success("连接测试已提交"); await load(); }
async function toggle(row: Provider) { await liquidityApi.setProviderStatus(row.id,{ tenantId:query.tenantId,status:row.status===1?2:1,version:row.version }); ElMessage.success("状态已更新"); await load(); }
onMounted(load);
const providerType = (v:number) => v===1?"内部做市":"外部流动性";
</script>
<template>
  <div class="page-head"><div><h1>流动性提供方</h1><p>管理内部做市账户与外部交易所通道</p></div><el-button type="primary" @click="dialog=true">＋ 新建提供方</el-button></div>
  <section class="panel">
    <div class="toolbar"><el-input v-model="query.keyword" placeholder="名称 / 编码 / 场所" clearable style="width:260px"/><el-select v-model="query.status" placeholder="全部状态" clearable style="width:150px"><el-option label="已启用" :value="1"/><el-option label="已停用" :value="2"/></el-select><el-button class="search-button" type="primary" @click="load">查询</el-button></div>
    <div class="table-wrap"><el-table :data="rows" v-loading="loading">
      <el-table-column prop="providerCode" label="提供方编码" width="170"><template #default="{row}"><span class="mono">{{ row.providerCode }}</span></template></el-table-column>
      <el-table-column prop="providerName" label="名称" min-width="170"/><el-table-column label="类型" width="130"><template #default="{row}">{{ providerType(row.providerType) }}</template></el-table-column>
      <el-table-column prop="venueCode" label="交易场所" width="130"/><el-table-column label="环境" width="100"><template #default="{row}">{{ row.environment===1?"生产":"沙箱" }}</template></el-table-column>
      <el-table-column label="凭证" width="100"><template #default="{row}"><el-tag :type="row.credentialConfigured?'success':'warning'" effect="plain">{{ row.credentialConfigured?"已配置":"未配置" }}</el-tag></template></el-table-column>
      <el-table-column label="健康状态" width="120"><template #default="{row}"><span><i :class="['status-dot',row.lastHealthStatus===1?'ok':'warn']"></i>{{ row.lastHealthStatus===1?"健康":"待检查" }}</span></template></el-table-column>
      <el-table-column label="状态" width="100"><template #default="{row}"><el-tag :type="row.status===1?'success':'info'">{{ row.status===1?"已启用":"已停用" }}</el-tag></template></el-table-column>
      <el-table-column label="操作" fixed="right" width="170"><template #default="{row}"><el-button link type="primary" @click="test(row)">测试连接</el-button><el-button link :type="row.status===1?'warning':'success'" @click="toggle(row)">{{ row.status===1?"停用":"启用" }}</el-button></template></el-table-column>
    </el-table></div>
  </section>
  <el-dialog v-model="dialog" title="新建流动性提供方" width="620px">
    <el-form :model="form" label-width="110px"><el-row :gutter="16"><el-col :span="12"><el-form-item label="提供方编码"><el-input v-model="form.providerCode"/></el-form-item></el-col><el-col :span="12"><el-form-item label="名称"><el-input v-model="form.providerName"/></el-form-item></el-col>
    <el-col :span="12"><el-form-item label="类型"><el-select v-model="form.providerType"><el-option label="内部做市" :value="1"/><el-option label="外部流动性" :value="2"/></el-select></el-form-item></el-col><el-col :span="12"><el-form-item label="环境"><el-select v-model="form.environment"><el-option label="生产" :value="1"/><el-option label="沙箱" :value="2"/></el-select></el-form-item></el-col>
    <el-col :span="12"><el-form-item label="交易场所"><el-input v-model="form.venueCode" placeholder="如 BINANCE"/></el-form-item></el-col><el-col :span="12"><el-form-item label="交易用户ID"><el-input-number v-model="form.tradeUserId" :min="0"/></el-form-item></el-col>
    <el-col :span="12"><el-form-item label="凭证引用"><el-input v-model="form.credentialRef"/></el-form-item></el-col><el-col :span="12"><el-form-item label="账户引用"><el-input v-model="form.accountRef"/></el-form-item></el-col></el-row></el-form>
    <template #footer><el-button @click="dialog=false">取消</el-button><el-button type="primary" @click="create">创建</el-button></template>
  </el-dialog>
</template>
