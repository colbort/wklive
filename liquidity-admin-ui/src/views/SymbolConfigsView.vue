<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from "vue";
import { ElMessage } from "element-plus";
import { liquidityApi } from "@/api/liquidity";
import type {
  ConfigOptions,
  ConfigProviderOption,
  SymbolConfig,
} from "@/types/liquidity";

const loading = ref(false);
const dialog = ref(false);
const rows = ref<SymbolConfig[]>([]);
const configOptions = ref<ConfigOptions>({
  symbols: [],
  providers: [],
  tradingUsers: [],
});
const query = reactive({ keyword: "", status: "", limit: 20 });
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

async function load() {
  loading.value = true;
  try {
    const response = await liquidityApi.symbolConfigs(query);
    rows.value = (response.data || []) as unknown as SymbolConfig[];
  } finally {
    loading.value = false;
  }
}

async function loadConfigOptions() {
  const response = await liquidityApi.configOptions();
  configOptions.value = response as unknown as ConfigOptions;
}

async function openCreate() {
  await loadConfigOptions();
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
  await liquidityApi.createSymbolConfig({ ...form, symbolId });
  ElMessage.success("策略配置已创建");
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

const mode = (value: number) =>
  ["未知", "内部做市", "外部路由", "内部做市 + 外部对冲"][value] ||
  "未知";
const status = (value: number) =>
  ["未知", "运行中", "已停用", "已暂停", "已熔断"][value] || "未知";

onMounted(load);
</script>

<template>
  <div class="page-head">
    <div>
      <h1>交易对策略</h1>
      <p>配置现货、交割与永续合约的报价和对冲参数</p>
    </div>
    <el-button type="primary" @click="openCreate">＋ 新建策略</el-button>
  </div>

  <section class="panel">
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
      <el-button class="search-button" type="primary" @click="load">
        查询
      </el-button>
    </div>
    <div class="table-wrap">
      <el-table :data="rows" v-loading="loading">
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
        <el-table-column label="操作" fixed="right" width="210">
          <template #default="{ row }">
            <el-button link type="success" @click="action(row, 'start')">
              启动
            </el-button>
            <el-button link type="warning" @click="action(row, 'pause')">
              暂停
            </el-button>
            <el-button link type="danger" @click="action(row, 'stop')">
              停止
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>
  </section>

  <el-dialog
    v-model="dialog"
    title="新建交易对流动性策略"
    width="820px"
  >
    <el-form :model="form" label-width="135px">
      <el-row :gutter="16">
        <el-col :span="12">
          <el-form-item label="交易对">
            <el-select
              v-model="form.symbolId"
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
      </el-row>
    </el-form>
    <template #footer>
      <el-button @click="dialog = false">取消</el-button>
      <el-button type="primary" @click="create">创建策略</el-button>
    </template>
  </el-dialog>
</template>
