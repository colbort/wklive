<template>
  <div class="payment-page">
    <div class="page-header">
      <h2>通道与规则</h2><el-button @click="refreshCurrentPage">
        刷新
      </el-button>
    </div>
    <el-tabs v-model="activeTab">
      <el-tab-pane label="租户通道" name="channels">
        <el-card shadow="never" class="query-card">
          <el-form :model="channelQuery" inline label-width="90px">
            <el-form-item label="租户ID">
              <el-input-number v-model="channelQuery.tenantId" :min="0" :precision="0" />
            </el-form-item><el-form-item label="平台ID">
              <el-input-number v-model="channelQuery.platformId" :min="0" :precision="0" />
            </el-form-item><el-form-item label="关键字">
              <el-input v-model="channelQuery.keyword" clearable />
            </el-form-item><el-form-item>
              <el-button type="primary" @click="loadChannels">
                查询
              </el-button><el-button type="primary" @click="openChannelDialog()">
                新增通道
              </el-button>
            </el-form-item>
          </el-form>
        </el-card>
        <el-card shadow="never">
          <el-table v-loading="channelLoading" :data="channels" stripe>
            <el-table-column prop="id" label="ID" width="80" /><el-table-column prop="tenantId" label="租户ID" width="100" /><el-table-column prop="channelCode" label="通道编码" min-width="120" /><el-table-column prop="channelName" label="通道名称" min-width="140" /><el-table-column prop="displayName" label="展示名称" min-width="140" /><el-table-column prop="currency" label="币种" width="90" /><el-table-column prop="status" label="状态" width="80" /><el-table-column prop="visible" label="显示" width="80" /><el-table-column label="操作" width="160">
              <template #default="{ row }">
                <el-button link type="primary" @click="showChannelDetail(row)">
                  详情
                </el-button><el-button link type="primary" @click="openChannelDialog(row)">
                  编辑
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>
      <el-tab-pane label="通道规则" name="rules">
        <el-card shadow="never" class="query-card">
          <el-form :model="ruleQuery" inline label-width="90px">
            <el-form-item label="租户ID">
              <el-input-number v-model="ruleQuery.tenantId" :min="0" :precision="0" />
            </el-form-item><el-form-item label="通道ID">
              <el-input-number v-model="ruleQuery.channelId" :min="0" :precision="0" />
            </el-form-item><el-form-item>
              <el-button type="primary" @click="loadRules">
                查询
              </el-button><el-button type="primary" @click="openRuleDialog()">
                新增规则
              </el-button>
            </el-form-item>
          </el-form>
        </el-card>
        <el-card shadow="never">
          <el-table v-loading="ruleLoading" :data="rules" stripe>
            <el-table-column prop="id" label="ID" width="80" /><el-table-column prop="tenantId" label="租户ID" width="100" /><el-table-column prop="channelId" label="通道ID" width="100" /><el-table-column prop="ruleName" label="规则名称" min-width="140" /><el-table-column prop="priority" label="优先级" width="90" /><el-table-column prop="status" label="状态" width="80" /><el-table-column label="操作" width="160">
              <template #default="{ row }">
                <el-button link type="primary" @click="showRuleDetail(row)">
                  详情
                </el-button><el-button link type="primary" @click="openRuleDialog(row)">
                  编辑
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>
    </el-tabs>
    <el-dialog v-model="channelDialogVisible" :title="channelForm.id?'编辑通道':'新增通道'" width="760px">
      <el-form label-width="110px">
        <el-form-item label="租户ID">
          <el-input-number
            v-model="channelForm.tenantId"
            :min="1"
            :precision="0"
            :disabled="!!channelForm.id"
          />
        </el-form-item><el-form-item v-if="!channelForm.id" label="平台ID">
          <el-input-number v-model="channelForm.platformId" :min="1" :precision="0" />
        </el-form-item><el-form-item v-if="!channelForm.id" label="产品ID">
          <el-input-number v-model="channelForm.productId" :min="1" :precision="0" />
        </el-form-item><el-form-item v-if="!channelForm.id" label="账号ID">
          <el-input-number v-model="channelForm.accountId" :min="1" :precision="0" />
        </el-form-item><el-form-item v-if="!channelForm.id" label="通道编码">
          <el-input v-model="channelForm.channelCode" />
        </el-form-item><el-form-item label="通道名称">
          <el-input v-model="channelForm.channelName" />
        </el-form-item><el-form-item label="展示名称">
          <el-input v-model="channelForm.displayName" />
        </el-form-item><el-form-item label="币种">
          <el-input v-model="channelForm.currency" />
        </el-form-item><el-form-item label="状态">
          <el-input-number v-model="channelForm.status" :min="1" :precision="0" />
        </el-form-item><el-form-item label="显示">
          <el-input-number v-model="channelForm.visible" :min="1" :precision="0" />
        </el-form-item><el-form-item label="费率">
          <el-input v-model="channelForm.feeRate" />
        </el-form-item><el-form-item label="固定费用">
          <el-input-number v-model="channelForm.feeFixedAmount" :min="0" :precision="0" />
        </el-form-item><el-form-item label="扩展配置">
          <el-input v-model="channelForm.extConfig" type="textarea" :rows="3" />
        </el-form-item><el-form-item label="备注">
          <el-input v-model="channelForm.remark" type="textarea" :rows="3" />
        </el-form-item>
      </el-form><template #footer>
        <el-button @click="channelDialogVisible=false">
          取消
        </el-button><el-button type="primary" @click="submitChannel">
          确定
        </el-button>
      </template>
    </el-dialog>
    <el-dialog v-model="ruleDialogVisible" :title="ruleForm.id?'编辑规则':'新增规则'" width="760px">
      <el-form label-width="120px">
        <el-form-item label="租户ID">
          <el-input-number
            v-model="ruleForm.tenantId"
            :min="1"
            :precision="0"
            :disabled="!!ruleForm.id"
          />
        </el-form-item><el-form-item label="通道ID">
          <el-input-number
            v-model="ruleForm.channelId"
            :min="1"
            :precision="0"
            :disabled="!!ruleForm.id"
          />
        </el-form-item><el-form-item label="规则名称">
          <el-input v-model="ruleForm.ruleName" />
        </el-form-item><el-form-item label="优先级">
          <el-input-number v-model="ruleForm.priority" :min="0" :precision="0" />
        </el-form-item><el-form-item label="状态">
          <el-input-number v-model="ruleForm.status" :min="1" :precision="0" />
        </el-form-item><el-form-item label="单笔最小">
          <el-input-number v-model="ruleForm.singleAmountMin" :min="0" :precision="0" />
        </el-form-item><el-form-item label="单笔最大">
          <el-input-number v-model="ruleForm.singleAmountMax" :min="0" :precision="0" />
        </el-form-item><el-form-item label="新用户">
          <el-input-number v-model="ruleForm.allowNewUser" :min="1" :precision="0" />
        </el-form-item><el-form-item label="老用户">
          <el-input-number v-model="ruleForm.allowOldUser" :min="1" :precision="0" />
        </el-form-item><el-form-item label="备注">
          <el-input v-model="ruleForm.remark" type="textarea" :rows="3" />
        </el-form-item>
      </el-form><template #footer>
        <el-button @click="ruleDialogVisible=false">
          取消
        </el-button><el-button type="primary" @click="submitRule">
          确定
        </el-button>
      </template>
    </el-dialog>
    <el-dialog v-model="detailVisible" title="详情" width="700px">
      <pre class="detail-pre">{{ JSON.stringify(detailData, null, 2) }}</pre>
    </el-dialog>
  </div>
</template>
<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { tenantService, type TenantPayChannel, type TenantPayChannelRule } from '@/services'
const activeTab=ref('channels');const channelLoading=ref(false);const ruleLoading=ref(false);const channels=ref<TenantPayChannel[]>([]);const rules=ref<TenantPayChannelRule[]>([]);const detailVisible=ref(false);const detailData=ref<Record<string,any>>({});const channelDialogVisible=ref(false);const ruleDialogVisible=ref(false)
const channelQuery=reactive({tenantId:0,platformId:0,keyword:''});const ruleQuery=reactive({tenantId:0,channelId:0})
const channelForm=reactive({id:0,tenantId:0,platformId:0,productId:0,accountId:0,channelCode:'',channelName:'',displayName:'',icon:'',currency:'',sort:0,visible:1,status:1,singleMinAmount:0,singleMaxAmount:0,dailyMaxAmount:0,dailyMaxCount:0,feeType:1,feeRate:'',feeFixedAmount:0,extConfig:'',remark:''})
const ruleForm=reactive({id:0,tenantId:0,channelId:0,ruleName:'',priority:0,status:1,singleAmountMin:0,singleAmountMax:0,userTotalRechargeMin:0,userTotalRechargeMax:0,memberLevelMin:0,memberLevelMax:0,kycLevelMin:0,kycLevelMax:0,allowNewUser:1,allowOldUser:1,allowTags:'',denyTags:'',remark:''})
const loadChannels=async()=>{channelLoading.value=true;try{const res=await tenantService.getTenantChannelList({...channelQuery,tenantId:channelQuery.tenantId||undefined,platformId:channelQuery.platformId||undefined,keyword:channelQuery.keyword||undefined,limit:100});channels.value=res.data||[]}finally{channelLoading.value=false}}
const loadRules=async()=>{ruleLoading.value=true;try{const res=await tenantService.getTenantChannelRuleList({...ruleQuery,tenantId:ruleQuery.tenantId||undefined,channelId:ruleQuery.channelId||undefined,limit:100});rules.value=res.data||[]}finally{ruleLoading.value=false}}
const refreshCurrentPage=()=>{loadChannels();loadRules()}
const openChannelDialog=(row?:TenantPayChannel)=>{Object.assign(channelForm,row||{id:0,tenantId:0,platformId:0,productId:0,accountId:0,channelCode:'',channelName:'',displayName:'',icon:'',currency:'',sort:0,visible:1,status:1,singleMinAmount:0,singleMaxAmount:0,dailyMaxAmount:0,dailyMaxCount:0,feeType:1,feeRate:'',feeFixedAmount:0,extConfig:'',remark:''});channelDialogVisible.value=true}
const openRuleDialog=(row?:TenantPayChannelRule)=>{Object.assign(ruleForm,row||{id:0,tenantId:0,channelId:0,ruleName:'',priority:0,status:1,singleAmountMin:0,singleAmountMax:0,userTotalRechargeMin:0,userTotalRechargeMax:0,memberLevelMin:0,memberLevelMax:0,kycLevelMin:0,kycLevelMax:0,allowNewUser:1,allowOldUser:1,allowTags:'',denyTags:'',remark:''});ruleDialogVisible.value=true}
const submitChannel=async()=>{if(channelForm.id){await tenantService.updateTenantChannel({...channelForm})}else{await tenantService.createTenantChannel({...channelForm})}ElMessage.success('操作成功');channelDialogVisible.value=false;loadChannels()}
const submitRule=async()=>{if(ruleForm.id){await tenantService.updateTenantChannelRule({...ruleForm})}else{await tenantService.createTenantChannelRule({...ruleForm})}ElMessage.success('操作成功');ruleDialogVisible.value=false;loadRules()}
const showChannelDetail=async(row:TenantPayChannel)=>{const res=await tenantService.getTenantChannelDetail(row.id,row.tenantId);detailData.value=res.data||row;detailVisible.value=true}
const showRuleDetail=async(row:TenantPayChannelRule)=>{const res=await tenantService.getTenantChannelRuleDetail(row.id,row.tenantId);detailData.value=res.data||row;detailVisible.value=true}
onMounted(refreshCurrentPage)
</script>
<style scoped></style>
