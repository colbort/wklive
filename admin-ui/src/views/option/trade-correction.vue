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
        <el-button v-perm="'option:trade-correction:create'" type="danger" @click="openDialog">
          {{ t('option.createTradeCorrection') }}
        </el-button>
      </template>
    </CrudQueryCard>

    <el-card v-loading="loading" shadow="never" class="table-card">
      <el-table :data="items" stripe>
        <el-table-column type="expand">
          <template #default="{ row }">
            <el-table :data="row.legs" size="small" border>
              <el-table-column prop="legNo" :label="t('option.legNo')" width="80" />
              <el-table-column prop="userId" :label="t('option.userId')" width="110" />
              <el-table-column prop="accountId" :label="t('option.accountId')" width="120" />
              <el-table-column prop="coin" :label="t('option.coin')" width="90" />
              <el-table-column :label="t('option.correctionDirection')" width="110">
                <template #default="{ row: leg }">
                  {{ leg.direction === 1 ? t('option.debit') : t('option.credit') }}
                </template>
              </el-table-column>
              <el-table-column prop="amount" :label="t('option.amount')" min-width="130" />
              <el-table-column
                prop="instructionNo"
                :label="t('option.instructionNo')"
                min-width="240"
              />
            </el-table>
          </template>
        </el-table-column>
        <el-table-column
          v-if="authStore.isSystemAdmin"
          prop="tenantId"
          :label="t('option.tenantId')"
          width="110"
        />
        <el-table-column prop="caseNo" :label="t('option.caseNo')" min-width="220" />
        <el-table-column prop="tradeId" :label="t('option.tradeId')" width="110" />
        <el-table-column prop="contractId" :label="t('option.contractId')" width="110" />
        <el-table-column :label="t('option.status')" width="130">
          <template #default="{ row }">
            {{ statusLabel(row.status) }}
          </template>
        </el-table-column>
        <el-table-column
          prop="reason"
          :label="t('option.correctionReason')"
          min-width="220"
          show-overflow-tooltip
        />
        <el-table-column
          prop="evidenceRef"
          :label="t('option.evidenceRef')"
          min-width="200"
          show-overflow-tooltip
        />
        <el-table-column
          prop="lastErrorMsg"
          :label="t('option.lastError')"
          min-width="180"
          show-overflow-tooltip
        />
        <el-table-column :label="t('common.actions')" width="170" fixed="right">
          <template #default="{ row }">
            <template v-if="row.status === 1">
              <el-button
                v-perm="'option:trade-correction:review'"
                link
                type="success"
                @click="review(row, true)"
              >
                {{ t('option.approve') }}
              </el-button>
              <el-button
                v-perm="'option:trade-correction:review'"
                link
                type="danger"
                @click="review(row, false)"
              >
                {{ t('option.reject') }}
              </el-button>
            </template>
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
      :title="t('option.createTradeCorrection')"
      width="900px"
      destroy-on-close
    >
      <el-alert
        :title="t('option.tradeCorrectionWarning')"
        type="error"
        :closable="false"
        class="dialog-alert"
      />
      <el-form :model="form" label-width="130px">
        <el-form-item :label="t('option.tradeId')" required>
          <el-input-number v-model="form.tradeId" :min="1" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('option.correctionReason')" required>
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
        <el-form-item :label="t('option.correctionLegs')" required>
          <div class="legs-editor">
            <div v-for="(leg, index) in form.legs" :key="index" class="leg-row">
              <el-input-number
                v-model="leg.userId"
                :min="1"
                :precision="0"
                :placeholder="t('option.userId')"
              />
              <el-input-number
                v-model="leg.accountId"
                :min="0"
                :precision="0"
                :placeholder="t('option.accountId')"
              />
              <el-input v-model="leg.coin" :placeholder="t('option.coin')" />
              <el-select v-model="leg.direction">
                <el-option :label="t('option.debit')" :value="1" />
                <el-option :label="t('option.credit')" :value="2" />
              </el-select>
              <el-input v-model="leg.amount" :placeholder="t('option.amount')" />
              <el-button
                :disabled="form.legs.length <= 2"
                type="danger"
                link
                @click="removeLeg(index)"
              >
                {{ t('common.delete') }}
              </el-button>
            </div>
            <el-button link type="primary" @click="addLeg">
              {{ t('option.addCorrectionLeg') }}
            </el-button>
          </div>
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
import { optionService, type OptionTradeCorrection, type TradeCorrectionLegInput } from '@/services'
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
const items = ref<OptionTradeCorrection[]>([])
const pagination = usePagination<number>(20)
const query = reactive({
  tenantId: (authStore.isTenantUser ? authStore.profileTenantId : undefined) as number | undefined,
  contractId: undefined as number | undefined,
})
const form = reactive({
  tradeId: 0,
  reason: '',
  evidenceRef: '',
  legs: [] as TradeCorrectionLegInput[],
})

async function loadData() {
  loading.value = true
  try {
    const response = await optionService.listTradeCorrections({
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

function newLeg(direction: number): TradeCorrectionLegInput {
  return { userId: 0, accountId: 0, coin: '', direction, amount: '' }
}

function openDialog() {
  if (!query.tenantId) {
    ElMessage.error(t('option.selectTenantForControlEvents'))
    return
  }
  form.tradeId = 0
  form.reason = ''
  form.evidenceRef = ''
  form.legs = [newLeg(1), newLeg(2)]
  dialogVisible.value = true
}

function addLeg() {
  form.legs.push(newLeg(2))
}

function removeLeg(index: number) {
  form.legs.splice(index, 1)
}

async function submit() {
  if (
    !query.tenantId ||
    form.tradeId <= 0 ||
    !form.reason.trim() ||
    !form.evidenceRef.trim() ||
    form.legs.length < 2 ||
    form.legs.some((leg) => leg.userId <= 0 || !leg.coin.trim() || !leg.amount)
  ) {
    ElMessage.error(t('option.completeCorrectionForm'))
    return
  }
  submitting.value = true
  try {
    await optionService.createTradeCorrection({
      tenantId: query.tenantId,
      tradeId: form.tradeId,
      action: 1,
      reason: form.reason.trim(),
      evidenceRef: form.evidenceRef.trim(),
      legs: form.legs.map((leg) => ({ ...leg, coin: leg.coin.trim() })),
    })
    dialogVisible.value = false
    ElMessage.success(t('common.success'))
    await loadData()
  } finally {
    submitting.value = false
  }
}

async function review(row: OptionTradeCorrection, approve: boolean) {
  const { value } = await ElMessageBox.prompt(
    approve ? t('option.approveCorrectionReason') : t('option.rejectCorrectionReason'),
    approve ? t('option.approve') : t('option.reject'),
    { inputValidator: (input) => Boolean(input?.trim()) || t('option.reviewReasonRequired') },
  )
  await optionService.reviewTradeCorrection({
    tenantId: row.tenantId,
    correctionId: row.id,
    approve,
    reason: value.trim(),
  })
  ElMessage.success(t('common.success'))
  await loadData()
}

function statusLabel(status: number) {
  const key = {
    1: 'pendingReview',
    2: 'rejected',
    3: 'executing',
    4: 'completed',
    5: 'manualReview',
  }[status]
  return key ? t(`option.${key}`) : String(status)
}

onMounted(loadData)
</script>

<style scoped>
.legs-editor {
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: 12px;
}

.leg-row {
  display: grid;
  grid-template-columns: 150px 150px 100px 110px 140px auto;
  gap: 8px;
  align-items: center;
}
</style>
