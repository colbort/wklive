<template>
  <div class="payment-page">
    <div class="page-header">
      <h2>充值订单</h2><el-button @click="loadList">
        刷新
      </el-button>
    </div><el-card shadow="never" class="query-card">
      <el-form :model="query" inline label-width="90px">
        <el-form-item label="租户ID">
          <el-input-number v-model="query.tenantId" :min="0" :precision="0" />
        </el-form-item><el-form-item label="用户ID">
          <el-input-number v-model="query.userId" :min="0" :precision="0" />
        </el-form-item><el-form-item label="订单号">
          <el-input v-model="query.orderNo" clearable />
        </el-form-item><el-form-item label="业务单号">
          <el-input v-model="query.bizOrderNo" clearable />
        </el-form-item><el-form-item>
          <el-button type="primary" @click="loadList">
            查询
          </el-button>
        </el-form-item>
      </el-form>
    </el-card><el-card shadow="never">
      <el-table v-loading="loading" :data="list" stripe>
        <el-table-column prop="orderNo" label="订单号" min-width="180" /><el-table-column prop="tenantId" label="租户ID" width="100" /><el-table-column prop="userId" label="用户ID" width="100" /><el-table-column prop="currency" label="币种" width="80" /><el-table-column prop="orderAmount" label="订单金额" min-width="120" /><el-table-column prop="payAmount" label="支付金额" min-width="120" /><el-table-column prop="status" label="状态" width="90" /><el-table-column label="操作" width="260" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="showDetail(row)">
              详情
            </el-button><el-button link type="warning" @click="closeOrder(row)">
              关闭
            </el-button><el-button link type="success" @click="openManualSuccess(row)">
              手动成功
            </el-button><el-button link type="primary" @click="retryNotify(row)">
              重试通知
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card><el-dialog v-model="detailVisible" title="订单详情" width="780px">
      <pre class="detail-pre">{{ JSON.stringify(detailData, null, 2) }}</pre>
    </el-dialog><el-dialog v-model="manualVisible" title="手动标记成功" width="520px">
      <el-form label-width="110px">
        <el-form-item label="第三方流水">
          <el-input v-model="manualForm.thirdTradeNo" />
        </el-form-item><el-form-item label="支付金额">
          <el-input-number
            v-model="manualForm.payAmount"
            :min="0"
            :precision="0"
            style="width:100%"
          />
        </el-form-item><el-form-item label="备注">
          <el-input v-model="manualForm.remark" type="textarea" :rows="3" />
        </el-form-item>
      </el-form><template #footer>
        <el-button @click="manualVisible=false">
          取消
        </el-button><el-button type="primary" @click="submitManual">
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>
<script setup lang="ts">import { onMounted, reactive, ref } from 'vue';import { ElMessage, ElMessageBox } from 'element-plus';import { rechargeService, type RechargeOrder } from '@/services';const loading=ref(false);const list=ref<RechargeOrder[]>([]);const detailVisible=ref(false);const detailData=ref<Record<string,any>>({});const manualVisible=ref(false);const currentOrder=ref<RechargeOrder|null>(null);const manualForm=reactive({thirdTradeNo:'',payAmount:0,remark:''});const query=reactive({tenantId:0,userId:0,orderNo:'',bizOrderNo:''});const loadList=async()=>{loading.value=true;try{const res=await rechargeService.getRechargeOrderList({...query,tenantId:query.tenantId||undefined,userId:query.userId||undefined,orderNo:query.orderNo||undefined,bizOrderNo:query.bizOrderNo||undefined,limit:100});list.value=res.data||[]}finally{loading.value=false}};const showDetail=async(row:RechargeOrder)=>{const res=await rechargeService.getRechargeOrderDetail(row.orderNo,row.tenantId);detailData.value=res.data||row;detailVisible.value=true};const closeOrder=async(row:RechargeOrder)=>{await ElMessageBox.confirm('确认关闭该订单？','提示');await rechargeService.closeRechargeOrder(row.orderNo,row.tenantId);ElMessage.success('操作成功');loadList()};const openManualSuccess=(row:RechargeOrder)=>{currentOrder.value=row;Object.assign(manualForm,{thirdTradeNo:'',payAmount:row.payAmount||row.orderAmount||0,remark:''});manualVisible.value=true};const submitManual=async()=>{if(!currentOrder.value)return;await rechargeService.manualSuccessRechargeOrder(currentOrder.value.orderNo,{tenantId:currentOrder.value.tenantId,thirdTradeNo:manualForm.thirdTradeNo,payAmount:manualForm.payAmount,remark:manualForm.remark});ElMessage.success('操作成功');manualVisible.value=false;loadList()};const retryNotify=async(row:RechargeOrder)=>{await rechargeService.retryRechargeNotify(row.orderNo,row.tenantId);ElMessage.success('已提交重试通知')};onMounted(loadList)</script>
<style scoped></style>
