<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";

import { liquidityApi } from "@/api/liquidity";
import ListPager from "@/components/ListPager.vue";

const props = defineProps<{ resource: "hedges" | "risks" | "reconcile" }>();
const loading = ref(false);
const rows = ref<Record<string, unknown>[]>([]);
const query = reactive({ status: "", limit: 20, cursor: 0 });
const page = reactive({ total: 0, nextCursor: 0, hasMore: false });
const cursorHistory = ref<number[]>([]);
const detailVisible = ref(false);
const detailLoading = ref(false);
const selectedRow = ref<Record<string, unknown>>({});
const reconcileDetails = ref<Record<string, unknown>[]>([]);
const manualHedgeVisible = ref(false);
const runReconcileVisible = ref(false);
const submitting = ref(false);
const hedgeForm = reactive({
  configId: 0,
  providerId: 0,
  side: 1,
  qty: "",
  targetExposure: "0",
  remark: "",
});
const reconcileForm = reactive({
  providerId: 0,
  reconcileType: 1,
  windowStart: 0,
  windowEnd: 0,
});
const config = computed(
  () =>
    ({
      hedges: {
        title: "对冲任务",
        desc: "监控库存敞口触发的自动与人工对冲",
        api: liquidityApi.hedgeTasks,
        columns: [
          ["hedgeNo", "对冲编号"],
          ["configId", "策略ID"],
          ["providerId", "提供方"],
          ["targetQty", "目标数量"],
          ["executedQty", "已执行"],
          ["status", "状态"],
          ["lastErrorMsg", "最后错误"],
        ],
      },
      risks: {
        title: "风险事件",
        desc: "集中处理点差、库存、通道和结算风险",
        api: liquidityApi.riskEvents,
        columns: [
          ["eventNo", "事件编号"],
          ["riskType", "风险类型"],
          ["riskLevel", "等级"],
          ["metricValue", "当前值"],
          ["thresholdValue", "阈值"],
          ["status", "状态"],
          ["message", "事件说明"],
        ],
      },
      reconcile: {
        title: "外部对账",
        desc: "核对本地与外部场所的订单、成交、余额和持仓",
        api: liquidityApi.reconcileBatches,
        columns: [
          ["batchNo", "对账批次"],
          ["providerId", "提供方"],
          ["reconcileType", "类型"],
          ["localCount", "本地数量"],
          ["externalCount", "外部数量"],
          ["differenceCount", "差异"],
          ["status", "状态"],
        ],
      },
    })[props.resource],
);

async function load(reset = false) {
  if (reset) {
    query.cursor = 0;
    cursorHistory.value = [];
    page.total = 0;
  }
  loading.value = true;
  try {
    const response = await config.value.api({ ...query, count: page.total });
    rows.value = response.data || [];
    page.total = Number(response.total || 0);
    page.nextCursor = Number(response.nextCursor || 0);
    page.hasMore = Boolean(response.hasNext);
  } finally {
    loading.value = false;
  }
}

async function nextPage() {
  if (!page.hasMore || !page.nextCursor) return;
  cursorHistory.value.push(query.cursor);
  query.cursor = page.nextCursor;
  await load();
}

async function previousPage() {
  const cursor = cursorHistory.value.pop();
  if (cursor === undefined) return;
  query.cursor = cursor;
  await load();
}

function openPrimaryAction() {
  if (props.resource === "hedges") manualHedgeVisible.value = true;
  if (props.resource === "reconcile") {
    reconcileForm.windowEnd = Date.now();
    reconcileForm.windowStart = reconcileForm.windowEnd - 24 * 60 * 60 * 1000;
    runReconcileVisible.value = true;
  }
}

async function submitManualHedge() {
  if (hedgeForm.configId <= 0 || hedgeForm.providerId <= 0 || Number(hedgeForm.qty) <= 0) {
    ElMessage.warning("请填写有效的策略、提供方和对冲数量");
    return;
  }
  submitting.value = true;
  try {
    await liquidityApi.createManualHedge({ ...hedgeForm });
    ElMessage.success("人工对冲任务已创建");
    manualHedgeVisible.value = false;
    await load(true);
  } finally {
    submitting.value = false;
  }
}

async function submitReconcile() {
  if (reconcileForm.providerId <= 0 || reconcileForm.windowStart <= 0 || reconcileForm.windowEnd <= reconcileForm.windowStart) {
    ElMessage.warning("请填写有效的提供方和对账时间窗口");
    return;
  }
  submitting.value = true;
  try {
    await liquidityApi.runReconcile({ ...reconcileForm });
    ElMessage.success("对账任务已发起");
    runReconcileVisible.value = false;
    await load(true);
  } finally {
    submitting.value = false;
  }
}

async function showDetail(row: Record<string, unknown>) {
  selectedRow.value = row;
  reconcileDetails.value = [];
  detailVisible.value = true;
  if (props.resource !== "reconcile") return;
  detailLoading.value = true;
  try {
    const response = await liquidityApi.reconcileDetails(Number(row.id), { limit: 100 });
    reconcileDetails.value = response.data || [];
  } finally {
    detailLoading.value = false;
  }
}

async function cancelHedge() {
  const reason = await ElMessageBox.prompt("请输入取消原因", "取消对冲任务", {
    inputValidator: (value) => Boolean(value.trim()) || "取消原因不能为空",
  });
  await liquidityApi.cancelHedgeTask(Number(selectedRow.value.id), {
    version: Number(selectedRow.value.version || 0),
    reason: reason.value,
  });
  ElMessage.success("对冲任务已取消");
  detailVisible.value = false;
  await load();
}

async function retryHedge() {
  await ElMessageBox.confirm("确认重试该对冲任务？", "重试对冲");
  await liquidityApi.retryHedgeTask(Number(selectedRow.value.id), {
    version: Number(selectedRow.value.version || 0),
  });
  ElMessage.success("已提交重试");
  detailVisible.value = false;
  await load();
}

async function resolveRisk() {
  const result = await ElMessageBox.prompt("请输入处置说明", "处置风险事件", {
    inputValidator: (value) => Boolean(value.trim()) || "处置说明不能为空",
  });
  await liquidityApi.resolveRiskEvent(Number(selectedRow.value.id), {
    status: 3,
    resolution: result.value,
  });
  ElMessage.success("风险事件已处置");
  detailVisible.value = false;
  await load();
}

async function resolveDifference(row: Record<string, unknown>) {
  const result = await ElMessageBox.prompt("请输入差异处置说明", "解决对账差异", {
    inputValidator: (value) => Boolean(value.trim()) || "处置说明不能为空",
  });
  await liquidityApi.resolveReconcileDifference(Number(row.id), {
    status: 3,
    resolution: result.value,
  });
  ElMessage.success("对账差异已解决");
  const response = await liquidityApi.reconcileDetails(Number(selectedRow.value.id), { limit: 100 });
  reconcileDetails.value = response.data || [];
}

watch(() => props.resource, () => load(true));
onMounted(load);
</script>

<template>
  <div class="list-page">
    <div class="page-head">
      <div><h1>{{ config.title }}</h1><p>{{ config.desc }}</p></div>
      <el-button v-if="resource === 'reconcile'" type="primary" @click="openPrimaryAction">发起对账</el-button>
      <el-button v-if="resource === 'hedges'" type="primary" @click="openPrimaryAction">人工对冲</el-button>
    </div>
    <section class="panel list-panel">
      <div class="toolbar">
        <el-input v-model="query.status" placeholder="状态值" clearable style="width: 150px" />
        <el-button class="search-button" type="primary" @click="load(true)">查询</el-button>
      </div>
      <div class="table-wrap list-table-wrap">
        <el-table class="list-table" :data="rows" height="100%" v-loading="loading">
          <el-table-column
            v-for="[prop, label] in config.columns"
            :key="prop"
            :prop="prop"
            :label="label"
            :min-width="prop.toLowerCase().includes('msg') || prop === 'message' ? 220 : 120"
          />
          <el-table-column label="操作" width="100" fixed="right">
            <template #default="{ row }"><el-button link type="primary" @click="showDetail(row)">详情</el-button></template>
          </el-table-column>
        </el-table>
      </div>
      <ListPager
        v-model:limit="query.limit"
        :total="page.total"
        :can-previous="Boolean(cursorHistory.length)"
        :can-next="page.hasMore && Boolean(page.nextCursor)"
        :loading="loading"
        @previous="previousPage"
        @next="nextPage"
        @limit-change="load(true)"
      />
    </section>

    <el-dialog v-model="manualHedgeVisible" title="创建人工对冲" width="520px">
      <el-form :model="hedgeForm" label-width="110px">
        <el-form-item label="策略 ID"><el-input-number v-model="hedgeForm.configId" :min="1" /></el-form-item>
        <el-form-item label="提供方 ID"><el-input-number v-model="hedgeForm.providerId" :min="1" /></el-form-item>
        <el-form-item label="方向"><el-select v-model="hedgeForm.side"><el-option label="买入" :value="1" /><el-option label="卖出" :value="2" /></el-select></el-form-item>
        <el-form-item label="数量"><el-input v-model="hedgeForm.qty" /></el-form-item>
        <el-form-item label="目标敞口"><el-input v-model="hedgeForm.targetExposure" /></el-form-item>
        <el-form-item label="备注"><el-input v-model="hedgeForm.remark" type="textarea" /></el-form-item>
      </el-form>
      <template #footer><el-button @click="manualHedgeVisible = false">取消</el-button><el-button type="primary" :loading="submitting" @click="submitManualHedge">创建</el-button></template>
    </el-dialog>

    <el-dialog v-model="runReconcileVisible" title="发起外部对账" width="520px">
      <el-form :model="reconcileForm" label-width="110px">
        <el-form-item label="提供方 ID"><el-input-number v-model="reconcileForm.providerId" :min="1" /></el-form-item>
        <el-form-item label="对账类型"><el-input-number v-model="reconcileForm.reconcileType" :min="1" /></el-form-item>
        <el-form-item label="开始时间(ms)"><el-input-number v-model="reconcileForm.windowStart" :min="1" :controls="false" /></el-form-item>
        <el-form-item label="结束时间(ms)"><el-input-number v-model="reconcileForm.windowEnd" :min="1" :controls="false" /></el-form-item>
      </el-form>
      <template #footer><el-button @click="runReconcileVisible = false">取消</el-button><el-button type="primary" :loading="submitting" @click="submitReconcile">发起</el-button></template>
    </el-dialog>

    <el-dialog v-model="detailVisible" :title="`${config.title}详情`" width="760px">
      <el-descriptions :column="2" border>
        <el-descriptions-item v-for="(value, key) in selectedRow" :key="key" :label="String(key)">{{ value }}</el-descriptions-item>
      </el-descriptions>
      <el-table v-if="resource === 'reconcile'" v-loading="detailLoading" :data="reconcileDetails" style="margin-top: 16px">
        <el-table-column prop="differenceNo" label="差异编号" min-width="150" />
        <el-table-column prop="businessType" label="业务类型" />
        <el-table-column prop="localValue" label="本地值" />
        <el-table-column prop="externalValue" label="外部值" />
        <el-table-column prop="status" label="状态" width="80" />
		<el-table-column label="操作" width="100">
		  <template #default="{ row }">
			<el-button v-if="Number(row.status) !== 3 && Number(row.status) !== 4" link type="primary" @click="resolveDifference(row)">解决</el-button>
		  </template>
		</el-table-column>
      </el-table>
      <template #footer>
        <el-button v-if="resource === 'hedges'" type="danger" @click="cancelHedge">取消任务</el-button>
        <el-button v-if="resource === 'hedges'" type="warning" @click="retryHedge">重试任务</el-button>
        <el-button v-if="resource === 'risks'" type="primary" @click="resolveRisk">处置事件</el-button>
        <el-button @click="detailVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>
