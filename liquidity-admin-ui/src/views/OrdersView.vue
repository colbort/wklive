<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from "vue";

import { liquidityApi } from "@/api/liquidity";
import ListPager from "@/components/ListPager.vue";
import type { OptionGroup, OptionItem } from "@/types/liquidity";

const tab = ref<"quotes" | "external">("quotes");
const loading = ref(false);
const rows = ref<Record<string, unknown>[]>([]);
const optionGroups = ref<OptionGroup[]>([]);
const query = reactive({ keyword: "", status: "", limit: 20, cursor: 0 });
const page = reactive({ total: 0, nextCursor: 0, hasMore: false });
const cursorHistory = ref<number[]>([]);

const statusGroupKey = computed(() =>
  tab.value === "quotes" ? "quoteOrderStatus" : "externalOrderStatus",
);
const statusOptions = computed(() => groupOptions(statusGroupKey.value).filter((item) => item.value));

async function load(reset = false) {
  if (reset) {
    query.cursor = 0;
    cursorHistory.value = [];
    page.total = 0;
  }
  loading.value = true;
  try {
    const request = { ...query, status: query.status || undefined, count: page.total };
    const response =
      tab.value === "quotes"
        ? await liquidityApi.quoteOrders(request)
        : await liquidityApi.externalOrders(request);
    rows.value = response.data || [];
    page.total = Number(response.total || 0);
    page.nextCursor = Number(response.nextCursor || 0);
    page.hasMore = Boolean(response.hasNext);
  } finally {
    loading.value = false;
  }
}

async function loadOptions() {
  const response = await liquidityApi.options();
  optionGroups.value = response.data || [];
}

function groupOptions(key: string) {
  return optionGroups.value.find((group) => group.key === key)?.options || [];
}

function optionCode(groupKey: string, value: unknown) {
  return groupOptions(groupKey).find((item) => item.value === Number(value))?.code || "";
}

function optionText(groupKey: string, value: unknown) {
  const code = optionCode(groupKey, value);
  const labels: Record<string, string> = {
    SIDE_BUY: "买入",
    SIDE_SELL: "卖出",
    QUOTE_ORDER_STATUS_UNKNOWN: "未知",
    QUOTE_ORDER_STATUS_PENDING_SUBMIT: "待提交",
    QUOTE_ORDER_STATUS_OPEN: "挂单中",
    QUOTE_ORDER_STATUS_PART_FILLED: "部分成交",
    QUOTE_ORDER_STATUS_FILLED: "已成交",
    QUOTE_ORDER_STATUS_CANCELING: "撤单中",
    QUOTE_ORDER_STATUS_CANCELED: "已撤销",
    QUOTE_ORDER_STATUS_FAILED: "失败",
    QUOTE_ORDER_STATUS_UNCERTAIN: "状态不确定",
    EXTERNAL_ORDER_STATUS_UNKNOWN: "未知",
    EXTERNAL_ORDER_STATUS_PENDING_SUBMIT: "待提交",
    EXTERNAL_ORDER_STATUS_SUBMITTED: "已提交",
    EXTERNAL_ORDER_STATUS_PART_FILLED: "部分成交",
    EXTERNAL_ORDER_STATUS_FILLED: "已成交",
    EXTERNAL_ORDER_STATUS_CANCELING: "撤单中",
    EXTERNAL_ORDER_STATUS_CANCELED: "已撤销",
    EXTERNAL_ORDER_STATUS_REJECTED: "已拒绝",
    EXTERNAL_ORDER_STATUS_FAILED: "失败",
    EXTERNAL_ORDER_STATUS_UNCERTAIN: "状态不确定",
  };
  return labels[code] || code || String(value ?? "-");
}

function text(row: Record<string, unknown>, ...keys: string[]) {
  return String(keys.map((key) => row[key]).find((value) => value !== undefined && value !== "") ?? "-");
}

function formatDecimal(value: unknown) {
  const raw = String(value ?? "");
  if (!raw) return "-";
  if (!raw.includes(".")) return raw;
  return raw.replace(/(\.\d*?[1-9])0+$|\.0+$/, "$1");
}

function formatTime(value: unknown) {
  const raw = Number(value);
  if (!Number.isFinite(raw) || raw <= 0) return "-";
  const milliseconds = raw > 9_999_999_999 ? raw : raw * 1000;
  const date = new Date(milliseconds);
  if (Number.isNaN(date.getTime())) return "-";
  const pad = (part: number) => String(part).padStart(2, "0");
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`;
}

function statusTagType(value: unknown) {
  const code = optionCode(statusGroupKey.value, value);
  if (code.endsWith("_FILLED") || code.endsWith("_OPEN")) return "success";
  if (code.endsWith("_FAILED") || code.endsWith("_REJECTED")) return "danger";
  if (code.endsWith("_PART_FILLED") || code.endsWith("_UNCERTAIN")) return "warning";
  return "info";
}

async function nextPage() {
  if (!page.hasMore || !page.nextCursor) return;
  cursorHistory.value.push(query.cursor);
  query.cursor = page.nextCursor;
  await load();
}

async function previousPage() {
  const previous = cursorHistory.value.pop();
  if (previous === undefined) return;
  query.cursor = previous;
  await load();
}

watch(tab, async () => {
  query.status = "";
  await load(true);
});

onMounted(async () => {
  await Promise.all([loadOptions(), load()]);
});
</script>

<template>
  <div class="list-page">
    <div class="page-head">
      <div><h1>订单与成交</h1><p>追踪内部报价单与路由至外部场所的订单</p></div>
    </div>
    <section class="panel list-panel">
    <el-tabs v-model="tab" class="tabs">
      <el-tab-pane label="内部报价订单" name="quotes" />
      <el-tab-pane label="外部路由订单" name="external" />
    </el-tabs>
    <div class="toolbar">
      <el-input v-model="query.keyword" placeholder="订单号 / Client Order ID" clearable style="width: 280px" />
      <el-select v-model="query.status" placeholder="全部状态" clearable style="width: 160px">
        <el-option
          v-for="item in statusOptions"
          :key="item.value"
          :label="optionText(statusGroupKey, item.value)"
          :value="item.value"
        />
      </el-select>
      <el-button class="search-button" type="primary" @click="load(true)">查询</el-button>
    </div>
    <div class="table-wrap list-table-wrap">
      <el-table
        class="list-table"
        :data="rows"
        height="100%"
        v-loading="loading"
      >
        <el-table-column label="订单号" min-width="190">
          <template #default="{ row }"><span class="mono">{{ text(row, "quoteNo", "orderNo") }}</span></template>
        </el-table-column>
        <el-table-column label="交易对ID" width="105">
          <template #default="{ row }">{{ text(row, "symbolId") }}</template>
        </el-table-column>
        <el-table-column label="方向" width="90">
          <template #default="{ row }">
            <el-tag :type="Number(row.side) === 1 ? 'success' : 'danger'" effect="plain">
              {{ optionText("tradeSide", row.side) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="价格 / 数量" min-width="180">
          <template #default="{ row }"><span class="mono">{{ formatDecimal(row.price) }} / {{ formatDecimal(row.qty) }}</span></template>
        </el-table-column>
        <el-table-column label="已成交" width="120">
          <template #default="{ row }">{{ formatDecimal(row.filledQty) }}</template>
        </el-table-column>
        <el-table-column v-if="tab === 'external'" label="外部订单ID" min-width="170">
          <template #default="{ row }"><span class="mono">{{ text(row, "externalOrderId") }}</span></template>
        </el-table-column>
        <el-table-column label="状态" width="120">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" effect="plain">
              {{ optionText(statusGroupKey, row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="更新时间" width="180">
          <template #default="{ row }">{{ formatTime(row.updateTimes || row.finishedAt) }}</template>
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
  </div>
</template>

<style scoped>
.tabs { padding: 5px 17px 0; }
.tabs:deep(.el-tabs__header) { margin: 0; }
.tabs:deep(.el-tabs__nav-wrap::after) { background: rgba(111, 140, 176, 0.13); }
</style>
