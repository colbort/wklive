<template>
  <div class="payment-page">
    <div class="page-header">
      <h2>{{ t('payment.rechargeOrders') }}</h2>
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
      <el-form-item :label="t('payment.bizOrderNo')">
        <el-input v-model="query.bizOrderNo" clearable />
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
      </el-form-item>
    </CrudQueryCard>
    <el-card shadow="never" class="table-card">
      <el-table v-loading="loading" :data="list" stripe>
        <el-table-column prop="orderNo" :label="t('payment.orderNo')" min-width="180" />
        <el-table-column prop="tenantId" :label="t('common.tenantId')" width="100" />
        <el-table-column prop="userId" :label="t('common.userId')" width="100" />
        <el-table-column prop="currency" :label="t('payment.currency')" width="80" />
        <el-table-column :label="t('payment.orderAmount')" min-width="120">
          <template #default="{ row }">
            {{ formatCentAmount(row.orderAmount) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('payment.payAmount')" min-width="120">
          <template #default="{ row }">
            {{ formatCentAmount(row.payAmount) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('common.status')" width="110">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" disable-transitions>
              {{ getOptionValueLabel(optionGroups, 'payOrderStatus', row.status, t) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="t('common.actions')" width="260" fixed="right">
          <template #default="{ row }">
            <el-button
              v-perm="'payment:recharge-order:detail'"
              link
              type="primary"
              @click="showDetail(row)"
            >
              {{ t('common.detail') }}
            </el-button>
            <el-button
              v-if="canClose(row)"
              v-perm="'payment:recharge-order:close'"
              link
              type="warning"
              @click="closeOrder(row)"
            >
              {{ t('payment.closeOrder') }}
            </el-button>
            <el-button
              v-if="canManualSuccess(row)"
              v-perm="'payment:recharge-order:manual-success'"
              link
              type="success"
              @click="openManualSuccess(row)"
            >
              {{ t('payment.manualMarkSuccess') }}
            </el-button>
            <el-button
              v-if="canRetryNotify(row)"
              v-perm="'payment:recharge-order:retry-notify'"
              link
              type="primary"
              @click="retryNotify(row)"
            >
              {{ t('payment.retryNotify') }}
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
      <div v-if="detailVoucherImageUrl" class="voucher-preview-section">
        <div class="voucher-preview-label">
          {{ t('payment.voucherImage') }}
        </div>
        <el-image
          class="voucher-preview"
          :src="detailVoucherImageUrl"
          fit="cover"
          :preview-src-list="[detailVoucherImageUrl]"
          :preview-teleported="true"
          hide-on-click-modal
        />
      </div>
      <PaymentDetailDescriptions
        :data="detailDisplayData"
        :option-groups="optionGroups"
        :columns="1"
      />
    </el-drawer>

    <el-dialog v-model="manualVisible" :title="t('payment.manualMarkSuccess')" width="520px">
      <el-form label-width="110px">
        <el-form-item v-if="currentVoucherImageUrl" :label="t('payment.voucherImage')">
          <el-image
            class="voucher-preview"
            :src="currentVoucherImageUrl"
            fit="cover"
            :preview-src-list="[currentVoucherImageUrl]"
            :preview-teleported="true"
            hide-on-click-modal
          />
        </el-form-item>
        <el-form-item :label="t('payment.thirdTradeNo')">
          <el-input v-model="manualForm.thirdTradeNo" />
        </el-form-item>
        <el-form-item :label="t('payment.payAmount')">
          <el-input-number
            v-model="manualForm.payAmount"
            :min="0"
            :precision="2"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item :label="t('common.remark')">
          <el-input v-model="manualForm.remark" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="manualVisible = false">
          {{ t('common.cancel') }}
        </el-button>
        <el-button
          v-perm="'payment:recharge-order:manual-success'"
          type="primary"
          @click="submitManual"
        >
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
import { buildSystemAssetUrl, useSystemCore } from '@/composables/useSystemCore'
import { ElMessage, ElMessageBox } from 'element-plus'
import { catalogService, rechargeService, type OptionGroup, type RechargeOrder } from '@/services'
import PaymentDetailDescriptions from '@/components/payment/PaymentDetailDescriptions.vue'
import { findOptionGroup, getOptionLabel, getOptionValueLabel } from '@/utils/options'
import { amountToCent, centToAmount, formatCentAmount, formatCentFields } from '@/utils/amount'
import TenantSelect from '@/components/TenantSelect.vue'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'

const { t } = useI18n()
const { systemCore, loadSystemCore } = useSystemCore()
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)

const loading = ref(false)
const list = ref<RechargeOrder[]>([])
const detailVisible = ref(false)
const detailData = ref<RechargeOrder | null>(null)
const manualVisible = ref(false)
const currentOrder = ref<RechargeOrder | null>(null)
const optionGroups = ref<OptionGroup[]>([])
const payOrderStatusOptions = computed(() => findOptionGroup(optionGroups.value, 'payOrderStatus'))
const manualForm = reactive({ thirdTradeNo: '', payAmount: 0, remark: '' })
const query = reactive({
  tenantId: 0,
  userId: 0,
  orderNo: '',
  bizOrderNo: '',
  status: undefined as number | undefined,
})

const PAY_ORDER_STATUS_PENDING = 1
const PAY_ORDER_STATUS_PAYING = 2
const CENT_AMOUNT_KEYS = new Set(['orderAmount', 'payAmount', 'feeAmount'])

const detailDisplayData = computed(() => {
  return formatCentFields(detailData.value, CENT_AMOUNT_KEYS)
})
const resolveAssetUrl = (url?: string) => buildSystemAssetUrl(systemCore.value.assetUrl, url)
const detailVoucherImageUrl = computed(() => resolveAssetUrl(detailData.value?.voucherImage))
const currentVoucherImageUrl = computed(() => resolveAssetUrl(currentOrder.value?.voucherImage))

const loadList = async () => {
  loading.value = true
  try {
    const res = await rechargeService.getRechargeOrderList({
      ...query,
      tenantId: query.tenantId || undefined,
      userId: query.userId || undefined,
      orderNo: query.orderNo || undefined,
      bizOrderNo: query.bizOrderNo || undefined,
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

const showDetail = async (row: RechargeOrder) => {
  const res = await rechargeService.getRechargeOrderDetail(row.orderNo, row.tenantId)
  detailData.value = res.data || row
  detailVisible.value = true
}

const canClose = (row: RechargeOrder) =>
  row.status === PAY_ORDER_STATUS_PENDING || row.status === PAY_ORDER_STATUS_PAYING
const canManualSuccess = (row: RechargeOrder) => row.status == PAY_ORDER_STATUS_PAYING
const canRetryNotify = (row: RechargeOrder) => row.status == PAY_ORDER_STATUS_PAYING

const statusTagType = (status: number) => {
  if (status === 3) return 'success'
  if (status === 4) return 'danger'
  if (status === 5 || status === 6) return 'info'
  if (status === 2) return 'warning'
  return ''
}

const closeOrder = async (row: RechargeOrder) => {
  await ElMessageBox.confirm(t('payment.confirmCloseRechargeOrder'), t('common.warning'))
  await rechargeService.closeRechargeOrder(row.orderNo, row.tenantId)
  ElMessage.success(t('common.operationSuccess'))
  loadList()
}

const openManualSuccess = (row: RechargeOrder) => {
  currentOrder.value = row
  Object.assign(manualForm, {
    thirdTradeNo: '',
    payAmount: centToAmount(row.payAmount || row.orderAmount || 0),
    remark: '',
  })
  manualVisible.value = true
}

const submitManual = async () => {
  if (!currentOrder.value) return
  await rechargeService.manualSuccessRechargeOrder(currentOrder.value.orderNo, {
    tenantId: currentOrder.value.tenantId,
    thirdTradeNo: manualForm.thirdTradeNo,
    payAmount: amountToCent(manualForm.payAmount),
    remark: manualForm.remark,
  })
  ElMessage.success(t('common.operationSuccess'))
  manualVisible.value = false
  loadList()
}

const retryNotify = async (row: RechargeOrder) => {
  await rechargeService.retryRechargeNotify(row.orderNo, row.tenantId)
  ElMessage.success(t('payment.retryNotifySubmitted'))
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
  void loadSystemCore()
  void loadOptions()
  void loadList()
})
</script>

<style scoped>
.break-text {
  word-break: break-all;
}

.voucher-preview-section {
  margin-bottom: 16px;
}

.voucher-preview-label {
  margin-bottom: 8px;
  color: #606266;
  font-size: 14px;
}

.voucher-preview {
  width: 120px;
  height: 120px;
  border: 1px solid #dcdfe6;
  border-radius: 6px;
  background: #f5f7fa;
  cursor: zoom-in;
}
</style>
