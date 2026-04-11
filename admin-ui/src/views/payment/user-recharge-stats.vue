<template>
  <div class="payment-page module-page">
    <div class="page-header">
      <h2>用户充值统计</h2><div class="header-actions">
        <el-button @click="loadList">
          刷新
        </el-button>
      </div>
    </div><el-card shadow="never" class="query-card">
      <el-form :model="query" inline label-width="120px">
        <el-form-item label="租户ID">
          <el-input-number v-model="query.tenantId" :min="0" :precision="0" />
        </el-form-item><el-form-item label="用户ID">
          <el-input-number v-model="query.userId" :min="0" :precision="0" />
        </el-form-item><el-form-item label="累计充值最小">
          <el-input-number v-model="query.successTotalAmountMin" :min="0" :precision="0" />
        </el-form-item><el-form-item label="累计充值最大">
          <el-input-number v-model="query.successTotalAmountMax" :min="0" :precision="0" />
        </el-form-item><el-form-item>
          <el-button type="primary" @click="loadList">
            查询
          </el-button>
        </el-form-item>
      </el-form>
    </el-card><el-card shadow="never" class="table-card">
      <el-table v-loading="loading" :data="list" stripe>
        <el-table-column prop="id" label="ID" width="80" /><el-table-column prop="tenantId" label="租户ID" width="100" /><el-table-column prop="userId" label="用户ID" width="100" /><el-table-column prop="successOrderCount" label="成功订单数" width="120" /><el-table-column prop="successTotalAmount" label="累计充值" min-width="120" /><el-table-column prop="todaySuccessAmount" label="今日充值" min-width="120" /><el-table-column prop="todaySuccessCount" label="今日笔数" width="100" /><el-table-column label="操作" width="100">
          <template #default="{ row }">
            <el-button link type="primary" @click="showDetail(row)">
              详情
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card><el-dialog v-model="detailVisible" title="详情" width="680px">
      <pre class="detail-pre">{{ JSON.stringify(detailData, null, 2) }}</pre>
    </el-dialog>
  </div>
</template>
<script setup lang="ts">import { onMounted, reactive, ref } from 'vue';import { rechargeService, type UserRechargeStat } from '@/services';const loading=ref(false);const list=ref<UserRechargeStat[]>([]);const detailVisible=ref(false);const detailData=ref<Record<string,any>>({});const query=reactive({tenantId:0,userId:0,successTotalAmountMin:0,successTotalAmountMax:0});const loadList=async()=>{loading.value=true;try{const res=await rechargeService.getUserRechargeStatList({...query,tenantId:query.tenantId||undefined,userId:query.userId||undefined,successTotalAmountMin:query.successTotalAmountMin||undefined,successTotalAmountMax:query.successTotalAmountMax||undefined,limit:100});list.value=res.data||[]}finally{loading.value=false}};const showDetail=async(row:UserRechargeStat)=>{const res=await rechargeService.getUserRechargeStat({tenantId:row.tenantId,userId:row.userId});detailData.value=res.data||row;detailVisible.value=true};onMounted(loadList)</script>
<style scoped></style>
