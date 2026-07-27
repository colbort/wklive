<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from "vue";
import { ElMessage } from "element-plus";
import { liquidityApi } from "@/api/liquidity";
import ListPager from "@/components/ListPager.vue";
import { useAuthStore } from "@/stores/auth";
import type {
  ConfigOptions,
  ConfigProviderOption,
  StrategyLevel,
  SymbolConfig,
  SymbolConfigDetail,
  SymbolConfigDetailResponse,
} from "@/types/liquidity";

const auth = useAuthStore();
const loading = ref(false);
const dialog = ref(false);
const editingId = ref<number | null>(null);
const editingVersion = ref(0);
const detailDialog = ref(false);
const detailLoading = ref(false);
const detail = ref<SymbolConfigDetail | null>(null);
const detailLevels = ref<StrategyLevel[]>([]);
const rows = ref<SymbolConfig[]>([]);
const configOptions = ref<ConfigOptions>({
  symbols: [],
  providers: [],
  tradingUsers: [],
});
const query = reactive({ keyword: "", status: "", limit: 20, cursor: 0 });
const page = reactive({ total: 0, nextCursor: 0, hasMore: false });
const cursorHistory = ref<number[]>([]);
const form = reactive({
  symbolId: undefined as number | undefined,
  liquidityMode: 1,
  internalProviderId: undefined as number | undefined,
  externalProviderId: undefined as number | undefined,
  externalSymbol: "",
  referencePriceSource: "",
  referencePriceKind: "MARK",
  quoteValidityMs: 3000,
  refreshIntervalMs: 1000,
  quoteTtlMs: 5000,
  repriceThresholdBps: "2",
  baseSpreadBps: "5",
  maxSpreadBps: "20",
  maxPriceDeviationBps: "50",
  minQuoteQty: "0.001",
  maxQuoteQty: "1",
  maxQuoteNotional: "100000",
  targetBaseInventory: "0",
  minBaseInventory: "0",
  maxBaseInventory: "10",
  maxNetExposure: "10",
  maxDailyNotional: "1000000",
  inventorySkewBps: "2",
  hedgeThreshold: "1",
  hedgeRatio: "1",
  selfTradePrevention: 1,
});

const selectedSymbol = computed(() =>
  configOptions.value.symbols.find((item) => item.symbolId === form.symbolId),
);
const internalProviders = computed(() =>
  configOptions.value.providers.filter((item) => item.providerType === 1),
);
const externalProviders = computed(() =>
  configOptions.value.providers.filter((item) => item.providerType === 2),
);
const selectedInternalProvider = computed(() =>
  internalProviders.value.find(
    (item) => item.providerId === form.internalProviderId,
  ),
);
const quoteLimitHint = computed(() => {
  const qty = Number(form.minQuoteQty);
  const maxNotional = Number(form.maxQuoteNotional);
  if (!Number.isFinite(qty) || qty <= 0) {
    return "填写单档报价数量后，将自动计算名义金额限制对应的最高参考价格。";
  }
  if (!Number.isFinite(maxNotional) || maxNotional < 0) {
    return "单档最大名义金额不能小于 0；填写 0 表示不限制。";
  }
  if (maxNotional === 0) {
    return "当前单档最大名义金额不设上限，实际报价仍受单档最大数量和账户余额限制。";
  }
  const maxPrice = maxNotional / qty;
  return `当前配置仅支持参考价格不高于 ${maxPrice.toLocaleString("zh-CN", {
    maximumFractionDigits: 8,
  })}；计算公式：${maxNotional} ÷ ${qty}。若市场价格高于该值，系统将无法生成报价。`;
});
const quoteLimitHintType = computed(() => {
  const qty = Number(form.minQuoteQty);
  const maxNotional = Number(form.maxQuoteNotional);
  if (
    Number.isFinite(qty) &&
    qty > 0 &&
    Number.isFinite(maxNotional) &&
    maxNotional > 0
  ) {
    return "warning";
  }
  return "info";
});

watch(selectedSymbol, (symbol) => {
  if (!symbol) return;
  form.externalSymbol = symbol.symbol;
  form.referencePriceSource = `crypto:BA:${symbol.symbol}`;
});

function providerLabel(item: ConfigProviderOption) {
  const user = item.tradeUserId ? ` / 用户 ${item.tradeUserId}` : "";
  return `${item.providerName}（${item.providerCode}${user}）`;
}

function walletLabel(value?: number) {
  return value === 1 ? "现货钱包" : value === 3 ? "合约钱包" : "—";
}

function symbolTypeLabel(
  productType: number,
  contractType: number,
  contractValueType: number,
) {
  if (productType === 1) return "现货";
  const valueType =
    contractValueType === 1
      ? "线性"
      : contractValueType === 2
        ? "反向"
        : "价值类型未知";
  if (productType === 2 && contractType === 1) return `永续·${valueType}`;
  if (productType === 2 && contractType === 2) return `交割·${valueType}`;
  return "不支持";
}

async function load(reset = false) {
  if (reset) {
    query.cursor = 0;
    cursorHistory.value = [];
  }
  loading.value = true;
  try {
    const response = await liquidityApi.symbolConfigs(query);
    rows.value = (response.data || []) as unknown as SymbolConfig[];
    page.total = Number(response.page?.total || 0);
    page.nextCursor = Number(response.page?.nextCursor || 0);
    page.hasMore = Boolean(response.page?.hasMore);
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

async function loadConfigOptions() {
  const response = await liquidityApi.configOptions();
  configOptions.value = response as unknown as ConfigOptions;
}

async function openCreate() {
  await loadConfigOptions();
  editingId.value = null;
  editingVersion.value = 0;
  dialog.value = true;
}

async function openEdit(row: SymbolConfig) {
  if (row.status === 1) {
    ElMessage.warning("请先暂停流动性，再修改策略");
    return;
  }
  await loadConfigOptions();
  const response = (await liquidityApi.symbolConfigDetail(
    row.id,
  )) as unknown as SymbolConfigDetailResponse;
  const data = response.data;
  Object.assign(form, {
    symbolId: data.symbolId,
    liquidityMode: data.liquidityMode,
    internalProviderId: data.internalProviderId || undefined,
    externalProviderId: data.externalProviderId || undefined,
    externalSymbol: data.externalSymbol,
    referencePriceSource: data.referencePriceSource,
    referencePriceKind: data.referencePriceKind,
    quoteValidityMs: data.quoteValidityMs,
    refreshIntervalMs: data.refreshIntervalMs,
    quoteTtlMs: data.quoteTtlMs,
    repriceThresholdBps: data.repriceThresholdBps,
    baseSpreadBps: data.baseSpreadBps,
    maxSpreadBps: data.maxSpreadBps,
    maxPriceDeviationBps: data.maxPriceDeviationBps,
    minQuoteQty: data.minQuoteQty,
    maxQuoteQty: data.maxQuoteQty,
    maxQuoteNotional: data.maxQuoteNotional,
    targetBaseInventory: data.targetBaseInventory,
    minBaseInventory: data.minBaseInventory,
    maxBaseInventory: data.maxBaseInventory,
    maxNetExposure: data.maxNetExposure,
    maxDailyNotional: data.maxDailyNotional,
    inventorySkewBps: data.inventorySkewBps,
    hedgeThreshold: data.hedgeThreshold,
    hedgeRatio: data.hedgeRatio,
    selfTradePrevention: data.selfTradePrevention,
  });
  editingId.value = row.id;
  editingVersion.value = data.version;
  dialog.value = true;
}

async function create() {
  const symbolId = form.symbolId;
  if (!symbolId) {
    ElMessage.warning("请选择交易对");
    return;
  }
  if (
    (form.liquidityMode === 1 || form.liquidityMode === 3) &&
    !form.internalProviderId
  ) {
    ElMessage.warning("请选择内部做市提供方");
    return;
  }
  if (
    (form.liquidityMode === 2 || form.liquidityMode === 3) &&
    !form.externalProviderId
  ) {
    ElMessage.warning("请选择外部流动性提供方");
    return;
  }
  const minQuoteQty = Number(form.minQuoteQty);
  const maxQuoteQty = Number(form.maxQuoteQty);
  const maxQuoteNotional = Number(form.maxQuoteNotional);
  if (!Number.isFinite(minQuoteQty) || minQuoteQty <= 0) {
    ElMessage.warning("单档报价数量必须大于 0");
    return;
  }
  if (
    !Number.isFinite(maxQuoteQty) ||
    maxQuoteQty <= 0 ||
    maxQuoteQty < minQuoteQty
  ) {
    ElMessage.warning("单档最大数量不能小于报价数量");
    return;
  }
  if (!Number.isFinite(maxQuoteNotional) || maxQuoteNotional < 0) {
    ElMessage.warning("单档最大名义金额不能小于 0");
    return;
  }
  if (editingId.value) {
    await liquidityApi.updateSymbolConfig(editingId.value, {
      ...form,
      symbolId,
      version: editingVersion.value,
    });
    ElMessage.success("策略配置已更新");
  } else {
    await liquidityApi.createSymbolConfig({ ...form, symbolId });
    ElMessage.success("策略配置已创建");
  }
  dialog.value = false;
  await load();
}

async function action(
  row: SymbolConfig,
  type: "start" | "pause" | "stop",
) {
  await liquidityApi.symbolAction(row.id, type, row.version);
  ElMessage.success("操作已提交");
  await load();
}

async function showDetail(row: SymbolConfig) {
  detailDialog.value = true;
  detailLoading.value = true;
  try {
    const response = (await liquidityApi.symbolConfigDetail(
      row.id,
    )) as unknown as SymbolConfigDetailResponse;
    detail.value = response.data;
    detailLevels.value = response.levels || [];
  } finally {
    detailLoading.value = false;
  }
}

const mode = (value: number) =>
  ["未知", "内部做市", "外部路由", "内部做市 + 外部对冲"][value] ||
  "未知";
const status = (value: number) =>
  ["未知", "运行中", "已停用", "已暂停", "已熔断"][value] || "未知";

onMounted(load);
</script>

<template>
  <div class="list-page">
  <div class="page-head">
    <div>
      <h1>交易对策略</h1>
      <p>配置现货、交割与永续合约的报价和对冲参数</p>
    </div>
    <el-button
      v-if="auth.hasPerm('liquidity:strategy:add')"
      type="primary"
      @click="openCreate"
    >
      ＋ 新建策略
    </el-button>
  </div>

  <section class="panel list-panel">
    <div class="toolbar">
      <el-input
        v-model="query.keyword"
        placeholder="交易对 / 外部交易对"
        clearable
        style="width: 260px"
      />
      <el-select
        v-model="query.status"
        clearable
        placeholder="全部状态"
        style="width: 150px"
      >
        <el-option label="运行中" :value="1" />
        <el-option label="已停用" :value="2" />
        <el-option label="已暂停" :value="3" />
        <el-option label="已熔断" :value="4" />
      </el-select>
      <el-button class="search-button" type="primary" @click="load(true)">
        查询
      </el-button>
    </div>
    <div class="table-wrap list-table-wrap">
      <el-table class="list-table" :data="rows" height="100%" v-loading="loading">
        <el-table-column label="交易对" min-width="140">
          <template #default="{ row }">
            <b>{{ row.symbol || `#${row.symbolId}` }}</b>
          </template>
        </el-table-column>
        <el-table-column label="产品" width="110">
          <template #default="{ row }">
            {{ ["未知", "现货", "合约", "秒合约"][row.productType] }}
          </template>
        </el-table-column>
        <el-table-column label="流动性模式" min-width="190">
          <template #default="{ row }">{{ mode(row.liquidityMode) }}</template>
        </el-table-column>
        <el-table-column
          prop="referencePriceSource"
          label="参考价格源"
          width="150"
        />
        <el-table-column label="点差(BPS)" width="120">
          <template #default="{ row }">
            {{ row.baseSpreadBps }} / {{ row.maxSpreadBps }}
          </template>
        </el-table-column>
        <el-table-column
          prop="refreshIntervalMs"
          label="刷新间隔(ms)"
          width="125"
        />
        <el-table-column label="状态" width="110">
          <template #default="{ row }">
            <el-tag
              :type="
                row.status === 1
                  ? 'success'
                  : row.status === 4
                    ? 'danger'
                    : 'warning'
              "
            >
              {{ status(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" fixed="right" width="250">
          <template #default="{ row }">
            <el-button
              v-if="auth.hasPerm('liquidity:strategy:detail')"
              link
              type="primary"
              @click="showDetail(row)"
            >
              详情
            </el-button>
            <el-button
              v-if="auth.hasPerm('liquidity:strategy:update')"
              link
              type="primary"
              @click="openEdit(row)"
            >
              编辑
            </el-button>
            <el-button
              v-if="auth.hasPerm('liquidity:strategy:start')"
              link
              type="success"
              @click="action(row, 'start')"
            >
              启动
            </el-button>
            <el-button
              v-if="auth.hasPerm('liquidity:strategy:pause')"
              link
              type="warning"
              @click="action(row, 'pause')"
            >
              暂停
            </el-button>
            <el-button
              v-if="auth.hasPerm('liquidity:strategy:stop')"
              link
              type="danger"
              @click="action(row, 'stop')"
            >
              停止
            </el-button>
          </template>
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

  <el-dialog
    v-model="dialog"
    :title="editingId ? '编辑交易对流动性策略' : '新建交易对流动性策略'"
    width="820px"
  >
    <el-form :model="form" label-width="135px">
      <el-row :gutter="16">
        <el-col :span="12">
          <el-form-item label="交易对">
            <el-select
              v-model="form.symbolId"
              :disabled="Boolean(editingId)"
              filterable
              placeholder="请选择交易对"
              style="width: 100%"
            >
              <el-option
                v-for="item in configOptions.symbols"
                :key="item.symbolId"
                :value="item.symbolId"
                :label="
                  `[${symbolTypeLabel(item.productType, item.contractType, item.contractValueType)}] ${item.displaySymbol || item.symbol}（${item.symbol} · ID ${item.symbolId}）`
                "
              />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="钱包类型">
            <el-input :model-value="walletLabel(selectedSymbol?.walletType)" disabled />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="流动性模式">
            <el-select v-model="form.liquidityMode" style="width: 100%">
              <el-option label="内部做市" :value="1" />
              <el-option label="外部路由" :value="2" />
              <el-option label="内部做市 + 外部对冲" :value="3" />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="内部做市提供方">
            <el-select
              v-model="form.internalProviderId"
              filterable
              clearable
              placeholder="请选择"
              style="width: 100%"
            >
              <el-option
                v-for="item in internalProviders"
                :key="item.providerId"
                :value="item.providerId"
                :label="providerLabel(item)"
              />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="交易用户">
            <el-input
              :model-value="
                selectedInternalProvider?.tradeUserId
                  ? String(selectedInternalProvider.tradeUserId)
                  : '—'
              "
              disabled
            />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="外部流动性提供方">
            <el-select
              v-model="form.externalProviderId"
              filterable
              clearable
              placeholder="请选择"
              style="width: 100%"
            >
              <el-option
                v-for="item in externalProviders"
                :key="item.providerId"
                :value="item.providerId"
                :label="providerLabel(item)"
              />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="外部交易对">
            <el-input v-model="form.externalSymbol" placeholder="BTCUSDT" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="参考价格源">
            <el-input
              v-model="form.referencePriceSource"
              placeholder="crypto:BA:BTCUSDT"
            />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="参考价格类型">
            <el-input v-model="form.referencePriceKind" />
          </el-form-item>
        </el-col>
        <el-col :span="8">
          <el-form-item label="基础点差(BPS)">
            <el-input v-model="form.baseSpreadBps" />
          </el-form-item>
        </el-col>
        <el-col :span="8">
          <el-form-item label="最大点差(BPS)">
            <el-input v-model="form.maxSpreadBps" />
          </el-form-item>
        </el-col>
        <el-col :span="8">
          <el-form-item label="刷新间隔(ms)">
            <el-input-number v-model="form.refreshIntervalMs" :min="100" />
          </el-form-item>
        </el-col>
        <el-col :span="8">
          <el-form-item label="单档报价数量">
            <el-input
              v-model="form.minQuoteQty"
              placeholder="例如 0.5"
            />
          </el-form-item>
        </el-col>
        <el-col :span="8">
          <el-form-item label="单档最大数量">
            <el-input
              v-model="form.maxQuoteQty"
              placeholder="例如 1"
            />
          </el-form-item>
        </el-col>
        <el-col :span="8">
          <el-form-item label="单档最大名义金额">
            <el-input
              v-model="form.maxQuoteNotional"
              placeholder="例如 200000，0 表示不限"
            />
          </el-form-item>
        </el-col>
        <el-col :span="24">
          <el-alert
            class="quote-limit-hint"
            :title="quoteLimitHint"
            :type="quoteLimitHintType"
            :closable="false"
            show-icon
          />
        </el-col>
      </el-row>
    </el-form>
    <template #footer>
      <el-button @click="dialog = false">取消</el-button>
      <el-button type="primary" @click="create">
        {{ editingId ? "保存修改" : "创建策略" }}
      </el-button>
    </template>
  </el-dialog>

  <el-dialog
    v-model="detailDialog"
    class="strategy-detail-dialog"
    title="流动性策略详情"
    width="980px"
  >
    <div class="strategy-detail" v-loading="detailLoading">
      <template v-if="detail">
        <h3 class="detail-title">基本配置</h3>
        <el-descriptions :column="3" border>
          <el-descriptions-item label="交易对">{{ detail.symbol }}</el-descriptions-item>
          <el-descriptions-item label="流动性模式">
            {{ mode(detail.liquidityMode) }}
          </el-descriptions-item>
          <el-descriptions-item label="状态">
            {{ status(detail.status) }}
          </el-descriptions-item>
          <el-descriptions-item label="内部提供方ID">
            {{ detail.internalProviderId || "—" }}
          </el-descriptions-item>
          <el-descriptions-item label="外部提供方ID">
            {{ detail.externalProviderId || "—" }}
          </el-descriptions-item>
          <el-descriptions-item label="外部交易对">
            {{ detail.externalSymbol || "—" }}
          </el-descriptions-item>
          <el-descriptions-item label="参考价格源" :span="2">
            {{ detail.referencePriceSource }}
          </el-descriptions-item>
          <el-descriptions-item label="价格类型">
            {{ detail.referencePriceKind }}
          </el-descriptions-item>
          <el-descriptions-item label="基础/最大点差">
            {{ detail.baseSpreadBps }} / {{ detail.maxSpreadBps }} BPS
          </el-descriptions-item>
          <el-descriptions-item label="刷新间隔">
            {{ detail.refreshIntervalMs }} ms
          </el-descriptions-item>
          <el-descriptions-item label="报价有效期">
            {{ detail.quoteValidityMs }} ms
          </el-descriptions-item>
        </el-descriptions>

        <h3 class="detail-title">报价与风险限制</h3>
        <el-descriptions :column="3" border>
          <el-descriptions-item label="最小报价数量">
            {{ detail.minQuoteQty }}
          </el-descriptions-item>
          <el-descriptions-item label="单档最大数量">
            {{ detail.maxQuoteQty }}
          </el-descriptions-item>
          <el-descriptions-item label="单档最大名义金额">
            {{ detail.maxQuoteNotional }}
          </el-descriptions-item>
          <el-descriptions-item label="数量步长">{{ detail.qtyStep }}</el-descriptions-item>
          <el-descriptions-item label="价格步长">{{ detail.priceTick }}</el-descriptions-item>
          <el-descriptions-item label="最大净敞口">
            {{ detail.maxNetExposure }}
          </el-descriptions-item>
          <el-descriptions-item label="基础库存范围" :span="2">
            {{ detail.minBaseInventory }} ～ {{ detail.maxBaseInventory }}
          </el-descriptions-item>
          <el-descriptions-item label="每日最大名义金额">
            {{ detail.maxDailyNotional }}
          </el-descriptions-item>
        </el-descriptions>

        <h3 class="detail-title">策略档位</h3>
        <el-table :data="detailLevels" border>
          <el-table-column prop="levelNo" label="档位" width="80" />
          <el-table-column prop="bidSpreadBps" label="买价差(BPS)" />
          <el-table-column prop="askSpreadBps" label="卖价差(BPS)" />
          <el-table-column prop="bidQty" label="买单数量" />
          <el-table-column prop="askQty" label="卖单数量" />
          <el-table-column label="启用" width="90">
            <template #default="{ row }">
              <el-tag :type="row.enabled === 1 ? 'success' : 'info'">
                {{ row.enabled === 1 ? "是" : "否" }}
              </el-tag>
            </template>
          </el-table-column>
        </el-table>
      </template>
    </div>
  </el-dialog>
</template>

<style scoped>
.detail-title {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 26px 0 12px;
  color: #dce7f7;
  font-size: 15px;
  font-weight: 650;
}

.detail-title:first-child {
  margin-top: 0;
}

.detail-title::before {
  width: 3px;
  height: 15px;
  border-radius: 2px;
  background: #25d3a2;
  content: "";
}

.strategy-detail {
  min-height: 180px;
}

.strategy-detail :deep(.el-descriptions) {
  --el-descriptions-table-border: rgba(111, 140, 176, 0.18);
}

.strategy-detail :deep(.el-descriptions__body) {
  overflow: hidden;
  border-radius: 8px;
  background: #091727;
}

.strategy-detail
  :deep(.el-descriptions__table.is-bordered .el-descriptions__cell) {
  padding: 13px 15px;
  border-color: rgba(111, 140, 176, 0.18);
}

.strategy-detail :deep(.el-descriptions__label.el-descriptions__cell) {
  width: 138px;
  color: #8194ad;
  font-weight: 600;
  background: rgba(22, 43, 65, 0.82);
}

.strategy-detail :deep(.el-descriptions__content.el-descriptions__cell) {
  color: #dce7f7;
  background: rgba(8, 22, 37, 0.72);
}

.strategy-detail :deep(.el-table) {
  overflow: hidden;
  border-radius: 8px;
  --el-table-bg-color: #091727;
  --el-table-tr-bg-color: #091727;
  --el-table-header-bg-color: #102238;
  --el-table-row-hover-bg-color: #102a3d;
  --el-table-border-color: rgba(111, 140, 176, 0.18);
  --el-table-text-color: #d3deec;
  --el-table-header-text-color: #8194ad;
}

.strategy-detail :deep(.el-table th.el-table__cell) {
  height: 46px;
  font-weight: 600;
}

.strategy-detail :deep(.el-table td.el-table__cell) {
  height: 46px;
}

.strategy-detail :deep(.el-tag--success) {
  border-color: rgba(37, 211, 162, 0.28);
  color: #45dfb4;
  background: rgba(37, 211, 162, 0.12);
}

.quote-limit-hint {
  margin-top: 2px;
  line-height: 1.6;
}
</style>
