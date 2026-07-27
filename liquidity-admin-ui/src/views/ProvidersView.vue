<script setup lang="ts">
import { onMounted, reactive, ref } from "vue";
import { ElMessage } from "element-plus";
import { liquidityApi } from "@/api/liquidity";
import ListPager from "@/components/ListPager.vue";
import type { ConfigOptions, Provider } from "@/types/liquidity";

const loading = ref(false);
const dialog = ref(false);
const provisionDialog = ref(false);
const provisioning = ref(false);
const rows = ref<Provider[]>([]);
const configOptions = ref<ConfigOptions>({
  symbols: [],
  providers: [],
  tradingUsers: [],
  unavailableSections: [],
  warnings: [],
});
const query = reactive({ keyword: "", status: "", limit: 20, cursor: 0 });
const page = reactive({ total: 0, nextCursor: 0, hasMore: false });
const cursorHistory = ref<number[]>([]);
const form = reactive({
  providerCode: "",
  providerName: "",
  providerType: 1,
  tradeUserId: 0,
  venueCode: "",
  environment: 1,
  credentialRef: "",
  accountRef: "",
  rateLimitPerSecond: 10,
  status: 2,
  remark: "",
});
const provisionForm = reactive({
  symbolId: undefined as number | undefined,
  providerCode: "",
  providerName: "",
  baseAmount: "",
  quoteAmount: "",
  remark: "",
});

async function load(reset = false) {
  if (reset) {
    query.cursor = 0;
    cursorHistory.value = [];
  }
  loading.value = true;
  try {
    const response = await liquidityApi.providers(query);
    rows.value = (response.data || []) as unknown as Provider[];
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

async function create() {
  await liquidityApi.createProvider(form);
  ElMessage.success("提供方已创建");
  dialog.value = false;
  await load();
}

function toAssetStorageAmount(value: string): string {
  const normalized = value.trim();
  const match = normalized.match(/^(\d+)(?:\.(\d+))?$/);
  if (!match || !/[1-9]/.test(normalized)) {
    throw new Error("初始资产数量必须大于 0");
  }
  const integer = match[1].replace(/^0+(?=\d)/, "");
  const fraction = match[2] || "";
  if (fraction.length <= 2) {
    return `${integer}${fraction.padEnd(2, "0")}`.replace(/^0+(?=\d)/, "");
  }
  const whole = `${integer}${fraction.slice(0, 2)}`.replace(/^0+(?=\d)/, "");
  const decimal = fraction.slice(2).replace(/0+$/, "");
  return decimal ? `${whole}.${decimal}` : whole;
}

async function provision() {
  const symbolId = provisionForm.symbolId;
  if (
    !symbolId ||
    !provisionForm.providerCode.trim() ||
    !provisionForm.providerName.trim()
  ) {
    ElMessage.warning("请完整填写交易对和提供方信息");
    return;
  }
  provisioning.value = true;
  try {
    await liquidityApi.provisionInternalProvider({
      ...provisionForm,
      symbolId,
      providerCode: provisionForm.providerCode.trim(),
      providerName: provisionForm.providerName.trim(),
      baseAmount: toAssetStorageAmount(provisionForm.baseAmount),
      quoteAmount: toAssetStorageAmount(provisionForm.quoteAmount),
    });
    ElMessage.success("内部做市账户、初始资产和提供方已创建");
    provisionDialog.value = false;
    await load();
  } catch (error) {
    if (error instanceof Error && error.message === "初始资产数量必须大于 0") {
      ElMessage.warning(error.message);
      return;
    }
    throw error;
  } finally {
    provisioning.value = false;
  }
}

async function openProvision() {
  const response = await liquidityApi.configOptions();
  configOptions.value = response as unknown as ConfigOptions;
  if (configOptions.value.unavailableSections.includes("symbols")) {
    ElMessage.error("交易对服务暂不可用，无法安全创建内部做市账户");
    return;
  }
  configOptions.value.warnings.forEach((warning) => ElMessage.warning(warning));
  provisionDialog.value = true;
}

async function test(row: Provider) {
  await liquidityApi.testProvider(row.id);
  ElMessage.success("连接测试已提交");
  await load();
}

async function toggle(row: Provider) {
  await liquidityApi.setProviderStatus(row.id, {
    status: row.status === 1 ? 2 : 1,
    version: row.version,
  });
  ElMessage.success("状态已更新");
  await load();
}

const providerType = (value: number) =>
  value === 1 ? "内部做市" : "外部流动性";

onMounted(load);
</script>

<template>
  <div class="list-page">
  <div class="page-head">
    <div>
      <h1>流动性提供方</h1>
      <p>管理内部做市账户与外部交易所通道</p>
    </div>
    <div>
      <el-button type="success" @click="openProvision">
        一键创建内部做市账户
      </el-button>
      <el-button type="primary" @click="dialog = true">＋ 新建提供方</el-button>
    </div>
  </div>

  <section class="panel list-panel">
    <div class="toolbar">
      <el-input
        v-model="query.keyword"
        placeholder="名称 / 编码 / 场所"
        clearable
        style="width: 260px"
      />
      <el-select
        v-model="query.status"
        placeholder="全部状态"
        clearable
        style="width: 150px"
      >
        <el-option label="已启用" :value="1" />
        <el-option label="已停用" :value="2" />
      </el-select>
      <el-button class="search-button" type="primary" @click="load(true)">
        查询
      </el-button>
    </div>
    <div class="table-wrap list-table-wrap">
      <el-table class="list-table" :data="rows" height="100%" v-loading="loading">
        <el-table-column prop="providerCode" label="提供方编码" width="170">
          <template #default="{ row }">
            <span class="mono">{{ row.providerCode }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="providerName" label="名称" min-width="170" />
        <el-table-column label="类型" width="130">
          <template #default="{ row }">
            {{ providerType(row.providerType) }}
          </template>
        </el-table-column>
        <el-table-column prop="venueCode" label="交易场所" width="130" />
        <el-table-column label="环境" width="100">
          <template #default="{ row }">
            {{ row.environment === 1 ? "生产" : "沙箱" }}
          </template>
        </el-table-column>
        <el-table-column label="凭证" width="100">
          <template #default="{ row }">
            <el-tag
              :type="row.credentialConfigured ? 'success' : 'warning'"
              effect="plain"
            >
              {{ row.credentialConfigured ? "已配置" : "未配置" }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="健康状态" width="120">
          <template #default="{ row }">
            <span>
              <i
                :class="[
                  'status-dot',
                  row.lastHealthStatus === 1 ? 'ok' : 'warn',
                ]"
              />
              {{ row.lastHealthStatus === 1 ? "健康" : "待检查" }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'">
              {{ row.status === 1 ? "已启用" : "已停用" }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" fixed="right" width="170">
          <template #default="{ row }">
            <el-button link type="primary" @click="test(row)">
              测试连接
            </el-button>
            <el-button
              link
              :type="row.status === 1 ? 'warning' : 'success'"
              @click="toggle(row)"
            >
              {{ row.status === 1 ? "停用" : "启用" }}
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
    v-model="provisionDialog"
    title="一键创建内部做市账户"
    width="620px"
  >
    <el-alert
      title="将创建隔离的内部交易用户、初始化交易对两侧现货资产，并创建一个默认停用的内部流动性提供方。"
      type="info"
      :closable="false"
      show-icon
      class="dialog-alert"
    />
    <el-form :model="provisionForm" label-width="125px">
      <el-row :gutter="16">
        <el-col :span="12">
          <el-form-item label="现货交易对">
            <el-select
              v-model="provisionForm.symbolId"
              filterable
              placeholder="请选择现货交易对"
              style="width: 100%"
            >
              <el-option
                v-for="item in configOptions.symbols.filter(
                  (symbol) => symbol.productType === 1,
                )"
                :key="item.symbolId"
                :value="item.symbolId"
                :label="`[现货] ${item.displaySymbol || item.symbol}（${item.symbol} · ID ${item.symbolId}）`"
              />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="提供方编码">
            <el-input
              v-model="provisionForm.providerCode"
              placeholder="如 MM_BTC_USDT"
            />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="提供方名称">
            <el-input
              v-model="provisionForm.providerName"
              placeholder="如 BTC/USDT 内部做市"
            />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="基础资产数量">
            <el-input
              v-model="provisionForm.baseAmount"
              placeholder="如 10 BTC"
            />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="计价资产数量">
            <el-input
              v-model="provisionForm.quoteAmount"
              placeholder="如 1000000 USDT"
            />
          </el-form-item>
        </el-col>
        <el-col :span="24">
          <el-form-item label="备注">
            <el-input v-model="provisionForm.remark" type="textarea" />
          </el-form-item>
        </el-col>
      </el-row>
    </el-form>
    <template #footer>
      <el-button @click="provisionDialog = false">取消</el-button>
      <el-button type="primary" :loading="provisioning" @click="provision">
        创建账户并初始化
      </el-button>
    </template>
  </el-dialog>

  <el-dialog v-model="dialog" title="新建流动性提供方" width="620px">
    <el-form :model="form" label-width="110px">
      <el-row :gutter="16">
        <el-col :span="12">
          <el-form-item label="提供方编码">
            <el-input v-model="form.providerCode" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="名称">
            <el-input v-model="form.providerName" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="类型">
            <el-select v-model="form.providerType">
              <el-option label="内部做市" :value="1" />
              <el-option label="外部流动性" :value="2" />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="环境">
            <el-select v-model="form.environment">
              <el-option label="生产" :value="1" />
              <el-option label="沙箱" :value="2" />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="交易场所">
            <el-input v-model="form.venueCode" placeholder="如 BINANCE" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="交易用户ID">
            <el-input-number v-model="form.tradeUserId" :min="0" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="凭证引用">
            <el-input v-model="form.credentialRef" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="账户引用">
            <el-input v-model="form.accountRef" />
          </el-form-item>
        </el-col>
      </el-row>
    </el-form>
    <template #footer>
      <el-button @click="dialog = false">取消</el-button>
      <el-button type="primary" @click="create">创建</el-button>
    </template>
  </el-dialog>
</template>

<style scoped>
.dialog-alert {
  margin-bottom: 20px;
}
</style>
