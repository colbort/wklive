<script setup lang="ts">
import { onMounted, reactive, ref, watch } from "vue";
import { liquidityApi } from "@/api/liquidity";
const tab=ref("quotes"), loading=ref(false), rows=ref<Record<string,unknown>[]>([]);
const query=reactive({keyword:"",status:"",limit:20});
async function load(){loading.value=true;try{const r=tab.value==="quotes"?await liquidityApi.quoteOrders(query):await liquidityApi.externalOrders(query);rows.value=r.data||[]}finally{loading.value=false}}
watch(tab,load);onMounted(load);
const text=(row:Record<string,unknown>,...keys:string[])=>String(keys.map(k=>row[k]).find(v=>v!==undefined&&v!=="")??"-");
</script>
<template>
  <div class="page-head"><div><h1>订单与成交</h1><p>追踪内部报价单与路由至外部场所的订单</p></div></div>
  <section class="panel">
    <el-tabs v-model="tab" class="tabs"><el-tab-pane label="内部报价订单" name="quotes"/><el-tab-pane label="外部路由订单" name="external"/></el-tabs>
    <div class="toolbar"><el-input v-model="query.keyword" placeholder="订单号 / Client Order ID" clearable style="width:280px"/><el-input v-model="query.status" placeholder="状态值" style="width:130px"/><el-button class="search-button" type="primary" @click="load">查询</el-button></div>
    <div class="table-wrap"><el-table :data="rows" v-loading="loading">
      <el-table-column label="订单号" min-width="190"><template #default="{row}"><span class="mono">{{ text(row,"quoteNo","orderNo") }}</span></template></el-table-column>
      <el-table-column label="交易对ID" width="105"><template #default="{row}">{{ text(row,"symbolId") }}</template></el-table-column>
      <el-table-column label="方向" width="90"><template #default="{row}"><el-tag :type="Number(row.side)===1?'success':'danger'" effect="plain">{{ Number(row.side)===1?"买入":"卖出" }}</el-tag></template></el-table-column>
      <el-table-column label="价格 / 数量" min-width="180"><template #default="{row}"><span class="mono">{{ text(row,"price") }} / {{ text(row,"qty") }}</span></template></el-table-column>
      <el-table-column label="已成交" width="120"><template #default="{row}">{{ text(row,"filledQty") }}</template></el-table-column>
      <el-table-column v-if="tab==='external'" label="外部订单ID" min-width="170"><template #default="{row}"><span class="mono">{{ text(row,"externalOrderId") }}</span></template></el-table-column>
      <el-table-column label="状态" width="110"><template #default="{row}"><el-tag effect="dark" type="info">{{ text(row,"status") }}</el-tag></template></el-table-column>
      <el-table-column label="更新时间" width="180"><template #default="{row}">{{ text(row,"updateTimes","finishedAt") }}</template></el-table-column>
    </el-table></div>
  </section>
</template>
<style scoped>.tabs{padding:5px 17px 0}.tabs:deep(.el-tabs__header){margin:0}.tabs:deep(.el-tabs__nav-wrap:after){background:rgba(111,140,176,.13)}</style>
