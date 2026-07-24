<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from "vue";
import { liquidityApi } from "@/api/liquidity";
const props=defineProps<{resource:"hedges"|"risks"|"reconcile"}>();
const loading=ref(false),rows=ref<Record<string,unknown>[]>([]);
const query=reactive({tenantId:1,status:"",limit:20});
const config=computed(()=>({
  hedges:{title:"对冲任务",desc:"监控库存敞口触发的自动与人工对冲",api:liquidityApi.hedgeTasks,columns:[["hedgeNo","对冲编号"],["configId","策略ID"],["providerId","提供方"],["targetQty","目标数量"],["executedQty","已执行"],["status","状态"],["lastErrorMsg","最后错误"]]},
  risks:{title:"风险事件",desc:"集中处理点差、库存、通道和结算风险",api:liquidityApi.riskEvents,columns:[["eventNo","事件编号"],["riskType","风险类型"],["riskLevel","等级"],["metricValue","当前值"],["thresholdValue","阈值"],["status","状态"],["message","事件说明"]]},
  reconcile:{title:"外部对账",desc:"核对本地与外部场所的订单、成交、余额和持仓",api:liquidityApi.reconcileBatches,columns:[["batchNo","对账批次"],["providerId","提供方"],["reconcileType","类型"],["localCount","本地数量"],["externalCount","外部数量"],["differenceCount","差异"],["status","状态"]]},
})[props.resource]);
async function load(){loading.value=true;try{const r=await config.value.api(query);rows.value=r.data||[]}finally{loading.value=false}}
watch(()=>props.resource,load);onMounted(load);
</script>
<template>
  <div class="page-head"><div><h1>{{ config.title }}</h1><p>{{ config.desc }}</p></div><el-button v-if="resource==='reconcile'" type="primary">发起对账</el-button><el-button v-if="resource==='hedges'" type="primary">人工对冲</el-button></div>
  <section class="panel"><div class="toolbar"><el-input v-model="query.status" placeholder="状态值" clearable style="width:150px"/><el-button @click="load">查询</el-button></div><div class="table-wrap"><el-table :data="rows" v-loading="loading"><el-table-column v-for="[prop,label] in config.columns" :key="prop" :prop="prop" :label="label" :min-width="prop.toLowerCase().includes('msg')||prop==='message'?220:120"/><el-table-column label="操作" width="100" fixed="right"><template #default><el-button link type="primary">详情</el-button></template></el-table-column></el-table></div></section>
</template>
