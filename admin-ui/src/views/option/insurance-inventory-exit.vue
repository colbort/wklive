<template>
  <div class="module-page">
    <CrudQueryCard :model="query" @search="search" @reset="resetQuery">
      <el-form-item v-if="authStore.isSystemAdmin" :label="t('option.tenantId')">
        <TenantSelect v-model="query.tenantId" />
      </el-form-item>
      <el-form-item :label="t('option.contractId')">
        <ContractSelect v-model="query.contractId" :tenant-id="query.tenantId" />
      </el-form-item>
      <template #actions>
        <el-button
          v-perm="'option:insurance-inventory-exit:create'"
          type="primary"
          @click="openCreateDialog"
        >
          {{ t('option.createInsuranceInventoryExit') }}
        </el-button>
      </template>
    </CrudQueryCard>

    <el-card v-loading="loading" shadow="never" class="table-card">
      <el-alert
        :title="t('option.insuranceInventoryExitWarning')"
        type="warning"
        :closable="false"
        class="dialog-alert"
      />
      <el-table :data="items" stripe>
        <el-table-column
          v-if="authStore.isSystemAdmin"
          prop="tenantId"
          :label="t('option.tenantId')"
          width="110"
        />
        <el-table-column prop="requestNo" :label="t('option.requestNo')" min-width="190" />
        <el-table-column prop="positionId" :label="t('option.positionId')" width="110" />
        <el-table-column prop="contractId" :label="t('option.contractId')" width="110" />
        <el-table-column :label="t('option.exitQuantityAndPrice')" min-width="150">
          <template #default="{ row }">
            {{ row.quantity }} @ {{ row.limitPrice }}
          </template>
        </el-table-column>
        <el-table-column :label="t('option.status')" width="110">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)">
              {{ statusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="orderId" :label="t('option.orderId')" width="110">
          <template #default="{ row }">
            {{ row.orderId || '—' }}
          </template>
        </el-table-column>
        <el-table-column :label="t('option.exitOrderStatus')" width="120">
          <template #default="{ row }">
            {{ row.orderId ? orderStatusLabel(row.orderStatus) : '—' }}
          </template>
        </el-table-column>
        <el-table-column :label="t('option.exitFillProgress')" min-width="150">
          <template #default="{ row }">
            {{ row.orderId ? `${row.filledQty} / ${row.unfilledQty}` : '—' }}
          </template>
        </el-table-column>
        <el-table-column :label="t('option.insuranceExitReview')" min-width="180">
          <template #default="{ row }">
            {{ row.requestedBy }} → {{ row.reviewedBy || '—' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="evidenceRef"
          :label="t('option.evidenceRef')"
          min-width="190"
          show-overflow-tooltip
        />
        <el-table-column
          prop="lastErrorMsg"
          :label="t('option.lastError')"
          min-width="180"
          show-overflow-tooltip
        />
        <el-table-column :label="t('common.actions')" width="220" fixed="right">
          <template #default="{ row }">
            <template v-if="row.status === 1">
              <el-button
                v-perm="'option:insurance-inventory-exit:review'"
                link
                type="success"
                @click="review(row, true)"
              >
                {{ t('option.approve') }}
              </el-button>
              <el-button
                v-perm="'option:insurance-inventory-exit:review'"
                link
                type="danger"
                @click="review(row, false)"
              >
                {{ t('option.reject') }}
              </el-button>
            </template>
            <el-button
              v-if="row.status === 2"
              v-perm="'option:insurance-inventory-exit:execute'"
              link
              type="danger"
              @click="execute(row)"
            >
              {{ t('option.executeReduceOnlyIOC') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
      <CursorPagination
        v-model:limit="pagination.pagination.limit"
        :total="pagination.pagination.total"
        :has-prev="pagination.pagination.hasPrev"
        :has-next="pagination.pagination.hasNext"
        @prev="pagination.prevAndLoad(loadData)"
        @next="pagination.nextAndLoad(loadData)"
        @limit-change="pagination.resetAndLoad(loadData)"
      />
    </el-card>

    <el-dialog
      v-model="dialogVisible"
      :title="t('option.createInsuranceInventoryExit')"
      width="680px"
      destroy-on-close
    >
      <el-alert
        :title="t('option.insuranceInventoryExitApprovalWarning')"
        type="error"
        :closable="false"
        class="dialog-alert"
      />
      <el-form :model="form" label-width="180px">
        <el-form-item :label="t('option.positionId')" required>
          <el-input-number v-model="form.positionId" :min="1" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('option.exitQuantity')" required>
          <el-input v-model="form.quantity" />
        </el-form-item>
        <el-form-item :label="t('option.iocLimitPrice')" required>
          <el-input v-model="form.limitPrice" />
        </el-form-item>
        <el-form-item :label="t('option.exitReason')" required>
          <el-input
            v-model="form.reason"
            type="textarea"
            :rows="2"
            maxlength="500"
            show-word-limit
          />
        </el-form-item>
        <el-form-item :label="t('option.evidenceRef')" required>
          <el-input v-model="form.evidenceRef" maxlength="500" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">
          {{ t('common.cancel') }}
        </el-button>
        <el-button type="danger" :loading="submitting" @click="submit">
          {{ t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { optionService, type OptionInsuranceInventoryExit } from '@/services'
import { usePagination } from '@/composables'
import { useAuthStore } from '@/stores/auth'
import TenantSelect from '@/components/TenantSelect.vue'
import ContractSelect from '@/components/ContractSelect.vue'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'

const { t } = useI18n()
const authStore = useAuthStore()
const loading = ref(false)
const submitting = ref(false)
const dialogVisible = ref(false)
const items = ref<OptionInsuranceInventoryExit[]>([])
const pagination = usePagination<number>(20)
const query = reactive({
  tenantId: (authStore.isTenantUser ? authStore.profileTenantId : undefined) as number | undefined,
  contractId: undefined as number | undefined,
})
const form = reactive({ positionId: 0, quantity: '', limitPrice: '', reason: '', evidenceRef: '' })

async function loadData() {
  loading.value = true
  try {
    const response = await optionService.listInsuranceInventoryExits({
      ...query,
      cursor: pagination.pagination.cursor,
      limit: pagination.pagination.limit,
    })
    items.value = response.data || []
    pagination.updateFromResponse(response)
  } finally {
    loading.value = false
  }
}

function search() {
  pagination.resetAndLoad(loadData)
}

function resetQuery() {
  query.tenantId = authStore.isTenantUser ? authStore.profileTenantId || undefined : undefined
  query.contractId = undefined
  pagination.resetAndLoad(loadData)
}

function openCreateDialog() {
  if (!query.tenantId) {
    ElMessage.error(t('option.selectTenantForControlEvents'))
    return
  }
  Object.assign(form, { positionId: 0, quantity: '', limitPrice: '', reason: '', evidenceRef: '' })
  dialogVisible.value = true
}

async function submit() {
  if (
    !query.tenantId ||
    form.positionId <= 0 ||
    Number(form.quantity) <= 0 ||
    Number(form.limitPrice) <= 0 ||
    !form.reason.trim() ||
    !form.evidenceRef.trim()
  ) {
    ElMessage.error(t('option.completeInsuranceInventoryExit'))
    return
  }
  submitting.value = true
  try {
    await optionService.createInsuranceInventoryExit({
      tenantId: query.tenantId,
      ...form,
      quantity: form.quantity.trim(),
      limitPrice: form.limitPrice.trim(),
      reason: form.reason.trim(),
      evidenceRef: form.evidenceRef.trim(),
    })
    dialogVisible.value = false
    ElMessage.success(t('common.success'))
    await loadData()
  } finally {
    submitting.value = false
  }
}

async function review(row: OptionInsuranceInventoryExit, approve: boolean) {
  const { value } = await ElMessageBox.prompt(
    approve ? t('option.approveInsuranceExitReason') : t('option.rejectInsuranceExitReason'),
    approve ? t('option.approve') : t('option.reject'),
    { inputValidator: (input) => Boolean(input?.trim()) || t('option.reviewReasonRequired') },
  )
  await optionService.reviewInsuranceInventoryExit({
    tenantId: row.tenantId,
    exitId: row.id,
    approve,
    reason: value.trim(),
  })
  ElMessage.success(t('common.success'))
  await loadData()
}

async function execute(row: OptionInsuranceInventoryExit) {
  await ElMessageBox.confirm(
    t('option.executeInsuranceExitConfirm', {
      quantity: row.quantity,
      price: row.limitPrice,
      positionId: row.positionId,
    }),
    t('option.executeReduceOnlyIOC'),
    { type: 'error', confirmButtonText: t('option.executeReduceOnlyIOC') },
  )
  await optionService.executeInsuranceInventoryExit({ tenantId: row.tenantId, exitId: row.id })
  ElMessage.success(t('common.success'))
  await loadData()
}

function statusLabel(status: number) {
  return (
    {
      1: t('option.pendingReview'),
      2: t('option.approved'),
      3: t('option.rejected'),
      4: t('option.submitted'),
    }[status] || String(status)
  )
}

function statusType(status: number) {
  return status === 2 ? 'success' : status === 1 ? 'warning' : status === 3 ? 'danger' : 'info'
}

function orderStatusLabel(status: number) {
  return (
    {
      1: t('option.orderPendingMatch'),
      2: t('option.orderPartFilled'),
      3: t('option.orderFilled'),
      4: t('option.orderCanceled'),
      5: t('option.orderRejected'),
      6: t('option.orderExpired'),
    }[status] || '—'
  )
}

onMounted(loadData)
</script>
