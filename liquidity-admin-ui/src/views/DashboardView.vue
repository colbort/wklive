<script setup lang="ts">
import { onMounted, reactive } from "vue";
import { liquidityApi } from "@/api/liquidity";
const stats = reactive({ runningSymbols: 0, healthyProviders: 0, openQuotes: 0, pendingRisks: 0 });
const recent = reactive<Array<{ time: string; type: string; message: string; level: string }>>([]);
onMounted(async () => {
  try {
    const data = await liquidityApi.dashboard() as Partial<typeof stats> & { recentEvents?: typeof recent };
    Object.assign(stats, data); if (data.recentEvents) recent.push(...data.recentEvents);
  } catch { /* API 尚未部署时保持空状态 */ }
});
const cards = [
  ["runningSymbols", "运行中交易对", "ACTIVE SYMBOLS", "#2dd4a4"],
  ["healthyProviders", "健康提供方", "HEALTHY PROVIDERS", "#5da9ff"],
  ["openQuotes", "活动报价单", "OPEN QUOTES", "#a889ff"],
  ["pendingRisks", "待处理风险", "OPEN RISK EVENTS", "#ffb454"],
] as const;
</script>
<template>
  <div class="page-head"><div><h1>运行总览</h1><p>实时监控报价、外部通道、库存敞口与系统风险</p></div><div class="live"><i></i>实时监控中</div></div>
  <div class="cards">
    <article v-for="[key, label, en, color] in cards" :key="key" class="panel">
      <small>{{ en }}</small><strong :style="{ color }">{{ stats[key] }}</strong><span>{{ label }}</span>
    </article>
  </div>
  <div class="dashboard-grid">
    <section class="panel health">
      <header><div><h3>系统健康度</h3><p>核心执行链路</p></div><span>最近 5 分钟</span></header>
      <div v-for="item in ['参考价格源','内部撮合通道','外部交易通道','报价任务调度','资产结算链路']" :key="item" class="health-row">
        <span><i class="status-dot ok"></i>{{ item }}</span><b>正常</b><em>— ms</em>
      </div>
    </section>
    <section class="panel events">
      <header><div><h3>最新风险事件</h3><p>需要关注的异常与熔断</p></div><RouterLink to="/risks">查看全部 →</RouterLink></header>
      <div v-if="!recent.length" class="empty-note">暂无待处理风险事件</div>
      <div v-for="event in recent" :key="event.time + event.message" class="event"><time>{{ event.time }}</time><span>{{ event.type }}</span><p>{{ event.message }}</p></div>
    </section>
  </div>
</template>
<style scoped>
.live { color:#6f8299;font-size:12px}.live i{display:inline-block;width:7px;height:7px;margin-right:8px;border-radius:50%;background:#2dd4a4;box-shadow:0 0 8px #2dd4a4}
.cards{display:grid;grid-template-columns:repeat(4,1fr);gap:16px}.cards article{position:relative;padding:20px 22px;overflow:hidden}.cards small{display:block;color:#52657c;font-size:9px;letter-spacing:1.4px}.cards strong{display:block;margin:16px 0 4px;font-size:34px}.cards span{color:#8292a7;font-size:12px}
.dashboard-grid{display:grid;grid-template-columns:1.1fr 1fr;gap:16px;margin-top:16px}.panel header{display:flex;justify-content:space-between;padding:19px 21px;border-bottom:1px solid rgba(111,140,176,.13)}h3{margin:0 0 4px;font-size:15px}header p{margin:0;color:#53657d;font-size:11px}header>span,header a{color:#61748b;font-size:11px;text-decoration:none}.health-row{display:grid;grid-template-columns:1fr 70px 70px;padding:14px 21px;border-bottom:1px solid rgba(111,140,176,.08);color:#9aabc0;font-size:12px}.health-row b{color:#2dd4a4;font-weight:500}.health-row em{color:#53657d;font-style:normal;text-align:right}.event{display:grid;grid-template-columns:70px 80px 1fr;padding:13px 20px;border-bottom:1px solid rgba(111,140,176,.08);font-size:11px}.event time{color:#596c83}.event span{color:#ffb454}.event p{margin:0;color:#9bacc0}
</style>
