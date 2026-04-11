<template>
  <div class="payment-page">
    <div class="page-header">
      <h2>提现通知日志</h2><el-button @click="loadList">
        刷新
      </el-button>
    </div><el-card shadow="never" class="query-card">
      <el-form :model="query" inline label-width="100px">
        <el-form-item label="租户ID">
          <el-input-number v-model="query.tenantId" :min="0" :precision="0" />
        </el-form-item><el-form-item label="订单号">
          <el-input v-model="query.orderNo" clearable />
        </el-form-item><el-form-item label="通知状态">
          <el-input-number v-model="query.notifyStatus" :min="0" :precision="0" />
        </el-form-item><el-form-item>
          <el-button type="primary" @click="loadList">
            查询
          </el-button>
        </el-form-item>
      </el-form>
    </el-card><el-card shadow="never">
      <el-table v-loading="loading" :data="list" stripe>
        <el-table-column prop="id" label="ID" width="80" /><el-table-column prop="tenantId" label="租户ID" width="100" /><el-table-column prop="orderNo" label="订单号" min-width="180" /><el-table-column prop="notifyStatus" label="通知状态" width="100" /><el-table-column prop="signResult" label="验签结果" width="100" /><el-table-column
          prop="errorMessage"
          label="错误信息"
          min-width="200"
          show-overflow-tooltip
        /><el-table-column label="操作" width="100">
          <template #default="{ row }">
            <el-button link type="primary" @click="showDetail(row)">
              详情
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card><el-dialog v-model="detailVisible" title="日志详情" width="760px">
      <pre class="detail-pre">{{ JSON.stringify(detailData, null, 2) }}</pre>
    </el-dialog>
  </div>
</template>
<script setup lang="ts">import { onMounted, reactive, ref } from 'vue';import { withdrawService, type PayNotifyLog } from '@/services';const loading=ref(false);const list=ref<PayNotifyLog[]>([]);const detailVisible=ref(false);const detailData=ref<Record<string,any>>({});const query=reactive({tenantId:0,orderNo:'',notifyStatus:0});const loadList=async()=>{loading.value=true;try{const res=await withdrawService.getWithdrawNotifyLogList({...query,tenantId:query.tenantId||undefined,orderNo:query.orderNo||undefined,notifyStatus:query.notifyStatus||undefined,limit:100});list.value=res.data||[]}finally{loading.value=false}};const showDetail=async(row:PayNotifyLog)=>{const res=await withdrawService.getWithdrawNotifyLogDetail(row.id,row.tenantId);detailData.value=res.data||row;detailVisible.value=true};onMounted(loadList)</script>
<style scoped></style>
