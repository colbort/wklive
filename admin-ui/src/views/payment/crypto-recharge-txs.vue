<template>
  <div class="payment-page module-page">
    <CrudQueryCard :model="query" @search="loadList" @reset="resetQuery">
      <el-form-item :label="t('common.tenantId')">
        <TenantSelect v-model="query.tenantId" class="tenant-select-filter" />
      </el-form-item>
      <el-form-item :label="t('common.userId')">
        <UserSelect v-model="query.userId" :tenant-id="query.tenantId || undefined" />
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
    </CrudQueryCard>

    <el-card shadow="never" class="table-card">
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
        <el-table-column
          :label="t('common.actions')"
          align="center"
          width="140"
          fixed="right"
        >
          <template #default="{ row }">
            <el-button
              v-perm="'payment:crypto-recharge-tx:detail'"
              link
              type="primary"
              @click="showDetail(row)"
            >
              {{ t('common.detail') }}
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

    <el-dialog v-model="detailVisible" :title="t('payment.cryptoRechargeTxDetail')" width="780px">
      <PaymentDetailDescriptions :data="detailData" />
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { usePagination } from '@/composables'
import { catalogService, cryptoService, type CryptoRechargeTx, type OptionGroup } from '@/services'
import { findOptionGroup, getOptionLabel } from '@/utils/options'
import TenantSelect from '@/components/TenantSelect.vue'
import UserSelect from '@/components/UserSelect.vue'
import PaymentDetailDescriptions from '@/components/payment/PaymentDetailDescriptions.vue'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'

const { t } = useI18n()
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)
const loading = ref(false)
const detailVisible = ref(false)
const list = ref<CryptoRechargeTx[]>([])
const detailData = ref<CryptoRechargeTx | null>(null)
const optionGroups = ref<OptionGroup[]>([])
const chainCodeOptions = computed(() => findOptionGroup(optionGroups.value, 'chainCode'))
const query = reactive({
  tenantId: undefined as number | undefined,
  userId: undefined as number | undefined,
  orderNo: '',
  coin: '',
  chainCode: undefined as number | undefined,
  txHash: '',
})

function params() {
  return Object.fromEntries(Object.entries(query).filter(([, v]) => v !== '' && v !== 0))
}
async function loadList() {
  loading.value = true
  try {
    const res = await cryptoService.listRechargeTxs({
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
function resetQuery() {
  Object.assign(query, {
    tenantId: undefined as number | undefined,
    userId: undefined as number | undefined,
    orderNo: '',
    coin: '',
    chainCode: undefined,
    txHash: '',
  })
  void loadList()
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
function handleLimitChange() {
  resetAndLoad(loadList)
}

function handlePrevPage() {
  prevAndLoad(loadList)
}

function handleNextPage() {
  nextAndLoad(loadList)
}

onMounted(() => {
  void loadOptions()
  void loadList()
})
</script>
