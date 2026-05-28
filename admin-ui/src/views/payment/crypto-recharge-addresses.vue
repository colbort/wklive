<template>
  <div class="payment-page module-page crypto-address-page">
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
          <el-select v-model="query.walletType" clearable style="width: 150px">
            <el-option
              v-for="item in walletTypeOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('payment.currency')">
          <el-input v-model="query.coin" clearable />
        </el-form-item>
        <el-form-item :label="t('payment.chain')">
          <el-select v-model="query.chainCode" clearable style="width: 150px">
            <el-option
              v-for="item in chainCodeOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('payment.address')">
          <el-input v-model="query.address" clearable />
        </el-form-item>
        <el-form-item :label="t('common.status')">
          <el-select v-model="query.status" clearable style="width: 150px">
            <el-option
              v-for="item in addressStatusOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleQuery">
            {{ t('common.search') }}
          </el-button>
          <el-button @click="resetQuery">
            {{ t('common.reset') }}
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="never" class="table-card crypto-address-table-card">
      <el-table
        v-loading="loading"
        :data="list"
        stripe
        height="100%"
        class="crypto-address-table"
      >
        <el-table-column prop="id" :label="t('common.id')" width="80" />
        <el-table-column prop="tenantId" :label="t('common.tenantId')" width="90" />
        <el-table-column prop="userId" :label="t('common.userId')" width="100" />
        <el-table-column :label="t('payment.walletType')" width="100">
          <template #default="{ row }">
            {{ optionLabel('walletType', row.walletType) }}
          </template>
        </el-table-column>
        <el-table-column prop="coin" :label="t('payment.currency')" width="90" />
        <el-table-column :label="t('payment.chain')" width="110">
          <template #default="{ row }">
            {{ optionLabel('chainCode', row.chainCode) }}
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
            {{ optionLabel('cryptoRechargeAddressType', row.addressType) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('common.status')" width="90">
          <template #default="{ row }">
            <el-tag>{{ optionLabel('cryptoRechargeAddressStatus', row.status) }}</el-tag>
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

      <CursorPagination
        v-model:limit="pagination.limit"
        :total="pagination.total"
        :has-prev="pagination.hasPrev"
        :has-next="pagination.hasNext"
        @prev="handlePrevPage"
        @next="handleNextPage"
        @limit-change="handleLimitChange"
      />
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
            <el-select v-model="form.walletType" style="width: 100%">
              <el-option
                v-for="item in walletTypeOptions"
                :key="item.value"
                :label="item.label"
                :value="item.value"
              />
            </el-select>
          </el-form-item>
          <el-form-item :label="t('payment.currency')">
            <el-input v-model="form.coin" />
          </el-form-item>
          <el-form-item :label="t('payment.chain')">
            <el-select v-model="form.chainCode" style="width: 100%">
              <el-option
                v-for="item in chainCodeOptions"
                :key="item.value"
                :label="item.label"
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
            <el-option
              v-for="item in addressSourceOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('payment.type')">
          <el-select v-model="form.addressType" style="width: 100%">
            <el-option
              v-for="item in addressTypeOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('common.status')">
          <el-select v-model="form.status" style="width: 100%">
            <el-option
              v-for="item in addressStatusOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
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
      <PaymentDetailDescriptions :data="detailDisplayData" />
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { useOptions, usePagination } from '@/composables'
import {
  assetService,
  catalogService,
  cryptoService,
  type CryptoRechargeAddress,
  type OptionGroup,
} from '@/services'
import TenantSelect from '@/components/TenantSelect.vue'
import PaymentDetailDescriptions from '@/components/payment/PaymentDetailDescriptions.vue'

const { t } = useI18n()

const loading = ref(false)
const list = ref<CryptoRechargeAddress[]>([])
const dialogVisible = ref(false)
const detailVisible = ref(false)
const detailData = ref<CryptoRechargeAddress | null>(null)
const optionGroups = ref<OptionGroup[]>([])
const { pagination, updateFromResponse, reset, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)
const { optionItems, optionLabel } = useOptions(optionGroups)
const walletTypeOptions = optionItems('walletType')
const chainCodeOptions = optionItems('chainCode')
const addressSourceOptions = optionItems('cryptoRechargeAddressSource')
const addressTypeOptions = optionItems('cryptoRechargeAddressType')
const addressStatusOptions = optionItems('cryptoRechargeAddressStatus')
const detailOptionKeys: Record<string, string> = {
  walletType: 'walletType',
  chainCode: 'chainCode',
  addressSource: 'cryptoRechargeAddressSource',
  addressType: 'cryptoRechargeAddressType',
  status: 'cryptoRechargeAddressStatus',
}
const detailDisplayData = computed(() => {
  if (!detailData.value) return null
  return Object.fromEntries(
    Object.entries(detailData.value).map(([key, value]) => [
      key,
      detailOptionKeys[key] ? optionLabel(detailOptionKeys[key], value as number) : value,
    ]),
  )
})
const query = reactive({
  tenantId: undefined as number | undefined,
  userId: undefined as number | undefined,
  walletType: undefined as number | undefined,
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
    const res = await cryptoService.listRechargeAddresses({
      ...params(),
      cursor: pagination.cursor,
      limit: pagination.limit,
    })
    list.value = res.data || []
    updateFromResponse(res)
  } finally {
    loading.value = false
  }
}
async function loadOptions() {
  const [paymentOptions, assetOptions] = await Promise.allSettled([
    catalogService.getOptions(),
    assetService.getOptions(),
  ])
  optionGroups.value = [
    ...(paymentOptions.status === 'fulfilled' ? paymentOptions.value.data || [] : []),
    ...(assetOptions.status === 'fulfilled' ? assetOptions.value.data || [] : []),
  ]
}
function resetQuery() {
  Object.assign(query, {
    tenantId: undefined,
    userId: undefined,
    walletType: undefined,
    coin: '',
    chainCode: undefined,
    address: '',
    status: undefined,
  })
  reset()
  void loadList()
}
function handleQuery() {
  resetAndLoad(loadList)
}
function handleLimitChange() {
  resetAndLoad(loadList)
}
function handlePrevPage() {
  prevAndLoad(loadList)
}
function handleNextPage() {
  nextAndLoad(loadList)
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

<style scoped>
.crypto-address-page {
  height: 100%;
  min-height: 0;
  overflow: hidden;
}

.crypto-address-table-card {
  flex: 1;
  min-height: 0;
}

.crypto-address-table-card :deep(.el-card__body) {
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 0;
}

.crypto-address-table {
  flex: 1;
  min-height: 0;
}
</style>
