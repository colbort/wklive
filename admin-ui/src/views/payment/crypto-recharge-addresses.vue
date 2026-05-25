<template>
  <div class="payment-page module-page">
    <div class="page-header">
      <h2>{{ t('payment.cryptoRechargeAddresses') }}</h2>
      <div class="header-actions">
        <el-button @click="loadList">
          {{ t('common.refresh') }}
        </el-button>
        <el-button type="primary" @click="openDialog()">
          {{ t('payment.addCryptoRechargeAddress') }}
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
        <el-form-item :label="t('payment.walletType')">
          <el-input-number v-model="query.walletType" :min="0" :precision="0" />
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
        <el-form-item :label="t('payment.address')">
          <el-input v-model="query.address" clearable />
        </el-form-item>
        <el-form-item :label="t('common.status')">
          <el-select v-model="query.status" clearable style="width: 150px">
            <el-option :label="t('common.disabled')" :value="1" />
            <el-option :label="t('payment.available')" :value="2" />
            <el-option :label="t('payment.frozen')" :value="3" />
          </el-select>
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
        <el-table-column prop="walletType" :label="t('payment.walletType')" width="80" />
        <el-table-column prop="coin" :label="t('payment.currency')" width="90" />
        <el-table-column :label="t('payment.chain')" width="110">
          <template #default="{ row }">
            {{ formatChainCode(row.chainCode) }}
          </template>
        </el-table-column>
        <el-table-column
          prop="address"
          :label="t('payment.address')"
          min-width="260"
          show-overflow-tooltip
        />
        <el-table-column
          prop="memo"
          :label="t('payment.memo')"
          min-width="120"
          show-overflow-tooltip
        />
        <el-table-column :label="t('payment.type')" width="130">
          <template #default="{ row }">
            {{
              row.addressType === 2
                ? t('payment.sharedAddressWithMemo')
                : t('payment.exclusiveAddress')
            }}
          </template>
        </el-table-column>
        <el-table-column :label="t('common.status')" width="90">
          <template #default="{ row }">
            <el-tag>{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
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
      :title="
        form.id ? t('payment.editCryptoRechargeAddress') : t('payment.addCryptoRechargeAddress')
      "
      width="680px"
    >
      <el-form label-width="110px">
        <el-form-item :label="t('common.tenantId')">
          <TenantSelect v-model="form.tenantId" include-system />
        </el-form-item>
        <template v-if="!form.id">
          <el-form-item :label="t('common.userId')">
            <el-input-number v-model="form.userId" :min="0" :precision="0" />
          </el-form-item>
          <el-form-item :label="t('payment.walletType')">
            <el-input-number v-model="form.walletType" :min="1" :precision="0" />
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
        </template>
        <el-form-item :label="t('payment.address')">
          <el-input v-model="form.address" />
        </el-form-item>
        <el-form-item :label="t('payment.memo')">
          <el-input v-model="form.memo" />
        </el-form-item>
        <el-form-item :label="t('payment.source')">
          <el-select v-model="form.addressSource" style="width: 100%">
            <el-option :label="t('payment.systemGenerated')" :value="1" />
            <el-option :label="t('payment.thirdPartyAssigned')" :value="2" />
            <el-option :label="t('payment.manualImport')" :value="3" />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('payment.type')">
          <el-select v-model="form.addressType" style="width: 100%">
            <el-option :label="t('payment.exclusiveAddress')" :value="1" />
            <el-option :label="t('payment.sharedAddressWithMemo')" :value="2" />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('common.status')">
          <el-select v-model="form.status" style="width: 100%">
            <el-option :label="t('common.disabled')" :value="1" />
            <el-option :label="t('payment.available')" :value="2" />
            <el-option :label="t('payment.frozen')" :value="3" />
          </el-select>
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

    <el-dialog
      v-model="detailVisible"
      :title="t('payment.cryptoRechargeAddressDetail')"
      width="760px"
    >
      <PaymentDetailDescriptions :data="detailData" />
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import {
  catalogService,
  cryptoService,
  type CryptoRechargeAddress,
  type OptionGroup,
} from '@/services'
import { findOptionGroup, getOptionLabel } from '@/utils/options'
import TenantSelect from '@/components/TenantSelect.vue'
import PaymentDetailDescriptions from '@/components/payment/PaymentDetailDescriptions.vue'

const { t } = useI18n()

const loading = ref(false)
const list = ref<CryptoRechargeAddress[]>([])
const dialogVisible = ref(false)
const detailVisible = ref(false)
const detailData = ref<CryptoRechargeAddress | null>(null)
const optionGroups = ref<OptionGroup[]>([])
const chainCodeOptions = computed(() => findOptionGroup(optionGroups.value, 'chainCode'))
const query = reactive({
  tenantId: 0,
  userId: 0,
  walletType: 0,
  coin: '',
  chainCode: undefined as number | undefined,
  address: '',
  status: undefined as number | undefined,
})
const form = reactive({
  id: 0,
  tenantId: 0,
  userId: 0,
  walletType: 1,
  coin: 'USDT',
  chainCode: 20,
  address: '',
  memo: '',
  addressSource: 3,
  addressType: 1,
  status: 2,
})

function params() {
  return Object.fromEntries(
    Object.entries(query).filter(([, v]) => v !== '' && v !== 0 && v !== undefined),
  )
}
async function loadList() {
  loading.value = true
  try {
    list.value = (await cryptoService.listRechargeAddresses({ ...params(), limit: 100 })).data || []
  } finally {
    loading.value = false
  }
}
async function loadOptions() {
  optionGroups.value = (await catalogService.getOptions()).data || []
}
function resetQuery() {
  Object.assign(query, {
    tenantId: 0,
    userId: 0,
    walletType: 0,
    coin: '',
    chainCode: undefined,
    address: '',
    status: undefined,
  })
  void loadList()
}
function openDialog(row?: CryptoRechargeAddress) {
  Object.assign(
    form,
    row
      ? { ...row, chainCode: Number(row.chainCode || 0) }
      : {
          id: 0,
          tenantId: 0,
          userId: 0,
          walletType: 1,
          coin: 'USDT',
          chainCode: 20,
          address: '',
          memo: '',
          addressSource: 3,
          addressType: 1,
          status: 2,
        },
  )
  dialogVisible.value = true
}
function showDetail(row: CryptoRechargeAddress) {
  detailData.value = row
  detailVisible.value = true
}
function statusText(status: number) {
  return status === 1
    ? t('common.disabled')
    : status === 3
      ? t('payment.frozen')
      : t('payment.available')
}
function formatChainCode(value: number) {
  const item = chainCodeOptions.value.find((option) => String(option.value) === String(value))
  return item ? getOptionLabel(t, item.code, item.value) : value
}
async function submit() {
  const res = form.id
    ? await cryptoService.updateRechargeAddress(form)
    : await cryptoService.createRechargeAddress(form)
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
