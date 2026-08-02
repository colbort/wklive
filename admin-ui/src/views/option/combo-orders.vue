<template>
  <div class="module-page">
    <CrudQueryCard :model="query" @search="search" @reset="resetQuery">
      <el-form-item v-if="authStore.isSystemAdmin" :label="t('option.tenantId')">
        <TenantSelect v-model="query.tenantId" class="tenant-select-filter" />
      </el-form-item>
      <el-form-item :label="t('option.comboNo')">
        <el-input v-model="query.comboNo" clearable />
      </el-form-item>
      <el-form-item :label="t('option.comboStatus')">
        <el-select v-model="query.status" clearable>
          <el-option
            v-for="status in comboStatuses"
            :key="status.value"
            :label="status.label"
            :value="status.value"
          />
        </el-select>
      </el-form-item>
    </CrudQueryCard>

    <el-card v-loading="loading" shadow="never" class="table-card">
      <el-table :data="rows" stripe>
        <el-table-column
          v-if="authStore.isSystemAdmin"
          prop="comboOrder.tenantId"
          :label="t('option.tenantId')"
          width="110"
        />
        <el-table-column prop="comboOrder.comboNo" :label="t('option.comboNo')" min-width="210" />
        <el-table-column prop="comboOrder.userId" :label="t('option.userId')" width="100" />
        <el-table-column prop="comboOrder.accountId" :label="t('option.accountId')" width="110" />
        <el-table-column
          prop="comboOrder.underlyingSymbol"
          :label="t('option.underlyingSymbol')"
          min-width="140"
        />
        <el-table-column prop="comboOrder.netPrice" :label="t('option.netPrice')" width="120" />
        <el-table-column prop="comboOrder.qty" :label="t('option.quantity')" width="110" />
        <el-table-column prop="comboOrder.filledQty" :label="t('option.filledQty')" width="120" />
        <el-table-column :label="t('option.comboLegs')" min-width="300">
          <template #default="{ row }">{{ formatComboLegs(row) }}</template>
        </el-table-column>
        <el-table-column :label="t('common.status')" width="120">
          <template #default="{ row }">
            <el-tag :type="comboStatusType(row.comboOrder.status)">
              {{ comboStatusLabel(row.comboOrder.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="t('common.actions')" width="160" fixed="right">
          <template #default="{ row }">
            <el-button
              v-perm="'option:operations:combo-view'"
              link
              type="primary"
              @click="openDetail(row)"
            >
              {{ t('option.detail') }}
            </el-button>
            <el-button
              v-if="[1, 2, 3].includes(row.comboOrder.status)"
              v-perm="'option:operations:combo-cancel'"
              link
              type="danger"
              @click="forceCancel(row)"
            >
              {{ t('option.forceCancelCombo') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
      <CursorPagination
        v-model:limit="page.limit"
        :total="page.total"
        :has-prev="page.hasPrev"
        :has-next="page.hasNext"
        @prev="pagination.prevAndLoad(loadRows)"
        @next="pagination.nextAndLoad(loadRows)"
        @limit-change="pagination.resetAndLoad(loadRows)"
      />
    </el-card>

    <el-drawer
      v-model="drawerVisible"
      :title="t('option.comboOrderDetail')"
      size="80%"
      destroy-on-close
    >
      <template v-if="detail">
        <el-alert
          v-if="detail.dataTruncated"
          :title="t('option.comboDetailTruncated')"
          type="warning"
          :closable="false"
          show-icon
        />
        <el-descriptions :column="3" border class="detail-section">
          <el-descriptions-item :label="t('option.comboNo')">
            {{ detail.comboOrder.comboNo }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.userId')">
            {{ detail.comboOrder.userId }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.accountId')">
            {{ detail.comboOrder.accountId }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.netPrice')">
            {{ detail.comboOrder.netPrice }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.quantity')">
            {{ detail.comboOrder.qty }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('common.status')">
            {{ comboStatusLabel(detail.comboOrder.status) }}
          </el-descriptions-item>
          <el-descriptions-item :label="t('option.cancelReason')" :span="3">
            {{ detail.comboOrder.cancelReason || '-' }}
          </el-descriptions-item>
        </el-descriptions>

        <el-divider>{{ t('option.comboLegs') }}</el-divider>
        <el-table :data="detail.legs" stripe>
          <el-table-column prop="legNo" :label="t('option.legNo')" width="80" />
          <el-table-column prop="contractId" :label="t('option.contractId')" width="120" />
          <el-table-column :label="t('option.side')" width="90">
            <template #default="{ row }">{{ row.side === 1 ? 'BUY' : 'SELL' }}</template>
          </el-table-column>
          <el-table-column prop="ratio" :label="t('option.ratio')" width="80" />
          <el-table-column prop="price" :label="t('option.price')" />
          <el-table-column prop="qty" :label="t('option.quantity')" />
          <el-table-column prop="filledQty" :label="t('option.filledQty')" />
          <el-table-column prop="childOrderId" :label="t('option.childOrderId')" min-width="120" />
        </el-table>

        <el-divider>{{ t('option.shadowOrders') }}</el-divider>
        <el-table :data="detail.childOrders" stripe>
          <el-table-column prop="order.orderNo" :label="t('option.orderNo')" min-width="210" />
          <el-table-column prop="order.contractId" :label="t('option.contractId')" width="120" />
          <el-table-column
            prop="order.marginAmount"
            :label="t('option.marginAmount')"
            width="140"
          />
          <el-table-column prop="order.marginCoin" :label="t('option.coin')" width="100" />
          <el-table-column prop="order.status" :label="t('common.status')" width="90" />
          <el-table-column
            prop="order.cancelReason"
            :label="t('option.cancelReason')"
            min-width="180"
          />
        </el-table>

        <el-divider>{{ t('option.comboTrades') }} ({{ detail.tradeTotal }})</el-divider>
        <el-table :data="detail.trades" stripe>
          <el-table-column
            prop="trade.comboMatchNo"
            :label="t('option.comboMatchNo')"
            min-width="210"
          />
          <el-table-column prop="trade.comboLegNo" :label="t('option.legNo')" width="80" />
          <el-table-column prop="trade.contractId" :label="t('option.contractId')" width="120" />
          <el-table-column prop="trade.price" :label="t('option.price')" width="120" />
          <el-table-column prop="trade.qty" :label="t('option.quantity')" width="120" />
          <el-table-column prop="trade.tradeTime" :label="t('option.tradeTime')" min-width="170">
            <template #default="{ row }">{{ formatTime(row.trade.tradeTime) }}</template>
          </el-table-column>
        </el-table>

        <el-divider>
          {{ t('option.assetInstructions') }} ({{ detail.assetInstructionTotal }})
        </el-divider>
        <el-table :data="detail.assetInstructions" stripe>
          <el-table-column
            prop="instructionNo"
            :label="t('option.instructionNo')"
            min-width="220"
          />
          <el-table-column prop="orderId" :label="t('option.childOrderId')" width="120" />
          <el-table-column prop="action" :label="t('option.action')" width="90" />
          <el-table-column prop="coin" :label="t('option.coin')" width="90" />
          <el-table-column prop="amount" :label="t('option.amount')" width="130" />
          <el-table-column prop="status" :label="t('common.status')" width="90" />
          <el-table-column prop="lastErrorMsg" :label="t('option.lastErrorMsg')" min-width="220" />
        </el-table>
      </template>
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'
import TenantSelect from '@/components/TenantSelect.vue'
import { usePagination } from '@/composables'
import {
  optionService,
  type OptionAdminComboOrderDetail,
  type OptionComboOrderDetail,
} from '@/services'
import { useAuthStore } from '@/stores/auth'

const { t } = useI18n()
const authStore = useAuthStore()
const loading = ref(false)
const rows = ref<OptionComboOrderDetail[]>([])
const detail = ref<OptionAdminComboOrderDetail>()
const drawerVisible = ref(false)
const pagination = usePagination<number>(20)
const page = pagination.pagination
const query = reactive({
  tenantId: (authStore.isTenantUser ? authStore.profileTenantId : undefined) as number | undefined,
  comboNo: '',
  status: undefined as number | undefined,
})

const comboStatuses = computed(() =>
  [1, 2, 3, 4, 5, 6, 7, 8].map((value) => ({ value, label: comboStatusLabel(value) })),
)

async function loadRows() {
  loading.value = true
  try {
    const response = await optionService.listAdminComboOrders({
      tenantId: query.tenantId,
      comboNo: query.comboNo || undefined,
      status: query.status,
      cursor: page.cursor,
      limit: page.limit,
    })
    rows.value = response.data || []
    pagination.updateFromResponse(response)
  } finally {
    loading.value = false
  }
}

function search() {
  pagination.reset()
  void loadRows()
}

function resetQuery() {
  query.tenantId = authStore.isTenantUser ? authStore.profileTenantId || undefined : undefined
  query.comboNo = ''
  query.status = undefined
  detail.value = undefined
  drawerVisible.value = false
  pagination.reset()
  void loadRows()
}

async function openDetail(row: OptionComboOrderDetail) {
  detail.value = (
    await optionService.getAdminComboOrder({
      tenantId: row.comboOrder.tenantId,
      id: row.comboOrder.id,
    })
  ).data
  drawerVisible.value = true
}

async function forceCancel(row: OptionComboOrderDetail) {
  const { value } = await ElMessageBox.prompt(
    t('option.forceCancelComboPrompt'),
    t('option.forceCancelCombo'),
    {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      inputPattern: /\S+/,
      inputErrorMessage: t('option.forceCancelReasonRequired'),
      inputValidator: (input) => input.trim().length <= 200 || t('option.forceCancelReasonTooLong'),
    },
  )
  await optionService.forceCancelComboOrder({
    tenantId: row.comboOrder.tenantId,
    id: row.comboOrder.id,
    reason: value.trim(),
  })
  ElMessage.success(t('common.success'))
  await loadRows()
  if (drawerVisible.value) await openDetail(row)
}

const comboStatusLabel = (status: number) =>
  ({
    1: t('option.comboFunding'),
    2: t('option.comboActive'),
    3: t('option.comboPartFilled'),
    4: t('option.comboFilled'),
    5: t('option.comboCanceling'),
    6: t('option.comboCanceled'),
    7: t('option.comboRejected'),
    8: t('option.manualReview'),
  })[status] || String(status)

const comboStatusType = (status: number) => {
  if (status === 4) return 'success'
  if ([6, 7].includes(status)) return 'info'
  if ([5, 8].includes(status)) return 'danger'
  if ([1, 3].includes(status)) return 'warning'
  return 'primary'
}

const formatComboLegs = (row: OptionComboOrderDetail) =>
  row.legs
    .map(
      (leg) =>
        `#${leg.legNo} ${leg.contractId} ${leg.side === 1 ? 'BUY' : 'SELL'} ${leg.ratio}@${leg.price}`,
    )
    .join(' | ')
const formatTime = (value?: number) => (value ? new Date(value * 1000).toLocaleString() : '-')

onMounted(loadRows)
</script>

<style scoped>
.detail-section {
  margin-top: 12px;
}
</style>
