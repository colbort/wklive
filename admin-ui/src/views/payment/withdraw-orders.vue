<template>
  <div class="payment-page">
    <div class="page-header">
      <h2>{{ t('payment.withdrawOrders') }}</h2>
      <el-button @click="loadList">
        {{ t('common.refresh') }}
      </el-button>
    </div>

    <CrudQueryCard :model="query" label-width="90px" :show-actions="false">
      <el-form-item :label="t('common.tenantId')">
        <TenantSelect v-model="query.tenantId" class="tenant-select-filter" />
      </el-form-item>
      <el-form-item :label="t('common.userId')">
        <el-input-number v-model="query.userId" :min="0" :precision="0" />
      </el-form-item>
      <el-form-item :label="t('payment.orderNo')">
        <el-input v-model="query.orderNo" clearable />
      </el-form-item>
      <el-form-item :label="t('common.status')">
        <el-select v-model="query.status" clearable style="width: 150px">
          <el-option
            v-for="item in payOrderStatusOptions"
            :key="item.value"
            :label="getOptionLabel(t, item.code, item.value)"
            :value="item.value"
          />
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
    </CrudQueryCard>

    <el-card shadow="never" class="table-card">
      <el-table v-loading="loading" :data="list" stripe>
        <el-table-column prop="orderNo" :label="t('payment.orderNo')" min-width="180" />
        <el-table-column prop="tenantId" :label="t('common.tenantId')" width="100" />
        <el-table-column prop="userId" :label="t('common.userId')" width="100" />
        <el-table-column prop="platformId" :label="t('payment.platformId')" width="100" />
        <el-table-column prop="channelId" :label="t('payment.channelId')" width="100" />
        <el-table-column prop="currency" :label="t('payment.currency')" width="80" />
        <el-table-column :label="t('payment.amount')" min-width="120">
          <template #default="{ row }">
            {{ formatCentAmount(row.amount) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('payment.feeAmount')" min-width="100">
          <template #default="{ row }">
            {{ formatCentAmount(row.feeAmount) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('common.status')" width="110">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" disable-transitions>
              {{ getOptionValueLabel(optionGroups, 'payOrderStatus', row.status, t) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="t('common.actions')" width="180">
          <template #default="{ row }">
            <el-button
              v-perm="'payment:withdraw-order:detail'"
              link
              type="primary"
              @click="showDetail(row)"
            >
              {{ t('common.detail') }}
            </el-button>
            <el-button
              v-if="canAudit(row)"
              v-perm="'payment:withdraw-order:audit'"
              link
              type="warning"
              @click="openAudit(row)"
            >
              {{ t('payment.auditWithdraw') }}
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

    <el-drawer v-model="detailVisible" :title="t('payment.orderDetail')" size="720px">
      <PaymentDetailDescriptions
        :data="detailDisplayData"
        :option-groups="optionGroups"
        :columns="1"
      />
    </el-drawer>

    <el-dialog v-model="auditVisible" :title="t('payment.auditWithdraw')" width="520px">
      <el-form label-width="110px">
        <el-form-item :label="t('payment.auditResult')">
          <el-radio-group v-model="auditForm.approve">
            <el-radio :value="1">
              {{ t('payment.approved') }}
            </el-radio>
            <el-radio :value="2">
              {{ t('payment.rejected') }}
            </el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="t('common.remark')">
          <el-input v-model="auditForm.remark" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="auditVisible = false">
          {{ t('common.cancel') }}
        </el-button>
        <el-button v-perm="'payment:withdraw-order:audit'" type="primary" @click="submitAudit">
          {{ t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { usePagination } from '@/composables'
import { ElMessage } from 'element-plus'
import { catalogService, withdrawService, type OptionGroup, type WithdrawOrder } from '@/services'
import PaymentDetailDescriptions from '@/components/payment/PaymentDetailDescriptions.vue'
import { findOptionGroup, getOptionLabel, getOptionValueLabel } from '@/utils/options'
import { formatCentAmount, formatCentFields } from '@/utils/amount'
import TenantSelect from '@/components/TenantSelect.vue'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'

const { t } = useI18n()
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)

const loading = ref(false)
const list = ref<WithdrawOrder[]>([])
const detailVisible = ref(false)
const detailData = ref<WithdrawOrder | null>(null)
const auditVisible = ref(false)
const currentOrder = ref<WithdrawOrder | null>(null)
const optionGroups = ref<OptionGroup[]>([])
const payOrderStatusOptions = computed(() => findOptionGroup(optionGroups.value, 'payOrderStatus'))

const PAY_ORDER_STATUS_PENDING = 1
const CENT_AMOUNT_KEYS = new Set(['amount', 'feeAmount', 'actualAmount'])

const detailDisplayData = computed(() => {
  return formatCentFields(detailData.value, CENT_AMOUNT_KEYS)
})

const auditForm = reactive({
  approve: 1,
  remark: '',
})

const query = reactive({
  tenantId: 0,
  userId: 0,
  orderNo: '',
  status: undefined as number | undefined,
})

const loadList = async () => {
  loading.value = true
  try {
    const res = await withdrawService.getWithdrawOrderList({
      ...query,
      tenantId: query.tenantId || undefined,
      userId: query.userId || undefined,
      orderNo: query.orderNo || undefined,
      status: query.status || undefined,
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
  query.tenantId = 0
  query.userId = 0
  query.orderNo = ''
  query.status = undefined
  resetAndLoad(loadList)
}

const showDetail = async (row: WithdrawOrder) => {
  const res = await withdrawService.getWithdrawOrderDetail(row.orderNo, row.tenantId)
  detailData.value = res.data || row
  detailVisible.value = true
}

const statusTagType = (status: number) => {
  if (status === 3) return 'success'
  if (status === 4) return 'danger'
  if (status === 5 || status === 6) return 'info'
  if (status === 2) return 'warning'
  return ''
}

const canAudit = (row: WithdrawOrder) => row.status === PAY_ORDER_STATUS_PENDING

const openAudit = (row: WithdrawOrder) => {
  if (!canAudit(row)) {
    ElMessage.warning(t('payment.onlyPendingWithdrawCanAudit'))
    return
  }
  currentOrder.value = row
  Object.assign(auditForm, { approve: 1, remark: '' })
  auditVisible.value = true
}

const submitAudit = async () => {
  if (!currentOrder.value) return
  if (!canAudit(currentOrder.value)) {
    ElMessage.warning(t('payment.onlyPendingWithdrawCanAudit'))
    auditVisible.value = false
    return
  }
  const result = await withdrawService.auditWithdrawOrder({
    tenantId: currentOrder.value.tenantId,
    orderNo: currentOrder.value.orderNo,
    approve: auditForm.approve,
    remark: auditForm.remark,
  })
  if (result.code !== 200) {
    ElMessage.error(result.msg || t('payment.auditFailed'))
    return
  } else {
    ElMessage.success(t('payment.auditSuccess'))
    auditVisible.value = false
  }
  loadList()
}

const loadOptions = async () => {
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

<style scoped></style>
