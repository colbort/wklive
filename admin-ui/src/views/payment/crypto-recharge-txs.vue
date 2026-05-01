<template>
  <div class="payment-page module-page">
    <div class="page-header">
      <h2>{{ t('payment.cryptoRechargeTxs') }}</h2>
      <div class="header-actions">
        <el-button @click="loadList">
          {{ t('common.refresh') }}
        </el-button>
        <el-button type="primary" @click="openDialog()">
          {{ t('payment.addCryptoRechargeTx') }}
        </el-button>
      </div>
    </div>

    <el-card shadow="never" class="query-card">
      <el-form :model="query" inline label-width="90px">
        <el-form-item :label="t('common.tenantId')">
          <el-input-number v-model="query.tenantId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('common.userId')">
          <el-input-number v-model="query.userId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('payment.orderNo')">
          <el-input v-model="query.orderNo" clearable />
        </el-form-item>
        <el-form-item :label="t('payment.currency')">
          <el-input v-model="query.coin" clearable />
        </el-form-item>
        <el-form-item :label="t('payment.chain')">
          <el-select v-model="query.chainCode" clearable style="width: 150px">
            <el-option
              v-for="item in chainCodeOptions"
              :key="item.value"
              :label="getOptionLabel(t, item.code, item.value)"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="Hash">
          <el-input v-model="query.txHash" clearable />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadList">
            {{ t('common.search') }}
          </el-button>
          <el-button @click="resetQuery">
            {{ t('common.reset') }}
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="never">
      <el-table v-loading="loading" :data="list" stripe>
        <el-table-column prop="id" :label="t('common.id')" width="80" />
        <el-table-column prop="tenantId" :label="t('common.tenantId')" width="90" />
        <el-table-column prop="userId" :label="t('common.userId')" width="100" />
        <el-table-column prop="orderNo" :label="t('payment.orderNo')" min-width="160" />
        <el-table-column prop="coin" :label="t('payment.currency')" width="90" />
        <el-table-column :label="t('payment.chain')" width="100">
          <template #default="{ row }">
            {{ formatChainCode(row.chainCode) }}
          </template>
        </el-table-column>
        <el-table-column
          prop="txHash"
          label="TxHash"
          min-width="240"
          show-overflow-tooltip
        />
        <el-table-column prop="amount" :label="t('payment.quantity')" width="120" />
        <el-table-column prop="confirmCount" :label="t('payment.confirmCount')" width="90" />
        <el-table-column prop="status" :label="t('common.status')" width="90" />
        <el-table-column :label="t('common.actions')" width="140" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="showDetail(row)">
              {{ t('common.detail') }}
            </el-button>
            <el-button link type="primary" @click="openDialog(row)">
              {{ t('common.edit') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog
      v-model="dialogVisible"
      :title="form.id ? t('payment.editCryptoRechargeTx') : t('payment.addCryptoRechargeTx')"
      width="720px"
    >
      <el-form label-width="130px">
        <el-form-item :label="t('common.tenantId')">
          <el-input-number v-model="form.tenantId" :min="0" :precision="0" />
        </el-form-item>
        <template v-if="!form.id">
          <el-form-item :label="t('common.userId')">
            <el-input-number v-model="form.userId" :min="0" :precision="0" />
          </el-form-item>
          <el-form-item :label="t('payment.currency')">
            <el-input v-model="form.coin" />
          </el-form-item>
          <el-form-item :label="t('payment.chain')">
            <el-select v-model="form.chainCode" style="width: 100%">
              <el-option
                v-for="item in chainCodeOptions"
                :key="item.value"
                :label="getOptionLabel(t, item.code, item.value)"
                :value="item.value"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="TxHash">
            <el-input v-model="form.txHash" />
          </el-form-item>
          <el-form-item :label="t('payment.fromAddress')">
            <el-input v-model="form.fromAddress" />
          </el-form-item>
          <el-form-item :label="t('payment.toAddress')">
            <el-input v-model="form.toAddress" />
          </el-form-item>
          <el-form-item :label="t('payment.cryptoAmount')">
            <el-input v-model="form.amount" />
          </el-form-item>
        </template>
        <el-form-item :label="t('payment.orderId')">
          <el-input-number v-model="form.orderId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('payment.orderNo')">
          <el-input v-model="form.orderNo" />
        </el-form-item>
        <el-form-item :label="t('payment.confirmCount')">
          <el-input-number v-model="form.confirmCount" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('payment.requiredConfirmCount')">
          <el-input-number v-model="form.requiredConfirmCount" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('common.status')">
          <el-select v-model="form.status" style="width: 100%">
            <el-option :label="t('payment.pendingConfirm')" :value="1" />
            <el-option :label="t('payment.confirming')" :value="2" />
            <el-option :label="t('payment.confirmed')" :value="3" />
            <el-option :label="t('common.failed')" :value="4" />
            <el-option :label="t('payment.credited')" :value="5" />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('payment.rawData')">
          <el-input v-model="form.rawData" type="textarea" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">
          {{ t('common.cancel') }}
        </el-button>
        <el-button type="primary" @click="submit">
          {{ t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="detailVisible" :title="t('payment.cryptoRechargeTxDetail')" width="780px">
      <pre class="detail-pre">{{ JSON.stringify(detailData, null, 2) }}</pre>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { catalogService, cryptoService, type CryptoRechargeTx, type OptionGroup } from '@/services'
import { findOptionGroup, getOptionLabel } from '@/utils/options'

const { t } = useI18n()
const loading = ref(false)
const dialogVisible = ref(false)
const detailVisible = ref(false)
const list = ref<CryptoRechargeTx[]>([])
const detailData = ref<CryptoRechargeTx | null>(null)
const optionGroups = ref<OptionGroup[]>([])
const chainCodeOptions = computed(() => findOptionGroup(optionGroups.value, 'chainCode'))
const query = reactive({
  tenantId: 0,
  userId: 0,
  orderNo: '',
  coin: '',
  chainCode: undefined as number | undefined,
  txHash: '',
})
const form = reactive({
  id: 0,
  tenantId: 0,
  userId: 0,
  orderId: 0,
  orderNo: '',
  coin: 'USDT',
  chainCode: 20,
  txHash: '',
  fromAddress: '',
  toAddress: '',
  memo: '',
  amount: '0',
  blockHeight: 0,
  confirmCount: 0,
  requiredConfirmCount: 0,
  status: 1,
  rawData: '',
  createTimes: 0,
  updateTimes: 0,
})

function params() {
  return Object.fromEntries(Object.entries(query).filter(([, v]) => v !== '' && v !== 0))
}
async function loadList() {
  loading.value = true
  try {
    list.value = (await cryptoService.listRechargeTxs({ ...params(), limit: 100 })).data || []
  } finally {
    loading.value = false
  }
}
function resetQuery() {
  Object.assign(query, {
    tenantId: 0,
    userId: 0,
    orderNo: '',
    coin: '',
    chainCode: undefined,
    txHash: '',
  })
  void loadList()
}
function openDialog(row?: CryptoRechargeTx) {
  Object.assign(
    form,
    row || {
      id: 0,
      tenantId: 0,
      userId: 0,
      orderId: 0,
      orderNo: '',
      coin: 'USDT',
      chainCode: 20,
      txHash: '',
      fromAddress: '',
      toAddress: '',
      memo: '',
      amount: '0',
      blockHeight: 0,
      confirmCount: 0,
      requiredConfirmCount: 0,
      status: 1,
      rawData: '',
      createTimes: 0,
      updateTimes: 0,
    },
  )
  dialogVisible.value = true
}
function showDetail(row: CryptoRechargeTx) {
  detailData.value = row
  detailVisible.value = true
}
function formatChainCode(value: number) {
  const item = chainCodeOptions.value.find((option) => String(option.value) === String(value))
  return item ? getOptionLabel(t, item.code, item.value) : value
}
async function loadOptions() {
  optionGroups.value = (await catalogService.getOptions()).data || []
}
async function submit() {
  const res = form.id
    ? await cryptoService.updateRechargeTx(form)
    : await cryptoService.createRechargeTx(form)
  if (res.code === 200 || res.code === 0) {
    ElMessage.success(t('common.success'))
    dialogVisible.value = false
    await loadList()
  }
}
onMounted(() => {
  void loadOptions()
  void loadList()
})
</script>
