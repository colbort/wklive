<template>
  <div class="module-page">
    <CrudQueryCard :model="query" @search="search" @reset="resetQuery">
      <el-form-item v-if="authStore.isSystemAdmin" :label="t('option.tenantId')">
        <TenantSelect v-model="query.tenantId" />
      </el-form-item>
      <el-form-item :label="t('option.userId')">
        <UserSelect v-model="query.userId" :tenant-id="query.tenantId" />
      </el-form-item>
      <el-form-item :label="t('option.contractId')">
        <ContractSelect v-model="query.contractId" :tenant-id="query.tenantId" />
      </el-form-item>
      <template #actions>
        <el-button v-perm="'option:mmp:config'" type="primary" @click="openDialog()">
          {{ t('option.configureMMP') }}
        </el-button>
      </template>
    </CrudQueryCard>

    <el-card v-loading="loading" shadow="never" class="table-card">
      <el-table :data="items" stripe>
        <el-table-column
          v-if="authStore.isSystemAdmin"
          prop="tenantId"
          :label="t('option.tenantId')"
          width="110"
        />
        <el-table-column prop="userId" :label="t('option.userId')" width="100" />
        <el-table-column prop="contractId" :label="t('option.contractId')" width="110" />
        <el-table-column prop="groupCode" :label="t('option.mmpGroup')" min-width="130" />
        <el-table-column :label="t('option.status')" width="110">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)">
              {{ statusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="qtyThreshold" :label="t('option.mmpQtyThreshold')" min-width="140" />
        <el-table-column
          prop="tradeCountThreshold"
          :label="t('option.mmpTradeThreshold')"
          min-width="140"
        />
        <el-table-column
          prop="lossThreshold"
          :label="t('option.mmpLossThreshold')"
          min-width="150"
        />
        <el-table-column :label="t('option.mmpWindowState')" min-width="210">
          <template #default="{ row }">
            {{ row.accumulatedQty }} / {{ row.tradeCount }} / {{ row.accumulatedLoss }}
          </template>
        </el-table-column>
        <el-table-column
          prop="triggerReason"
          :label="t('option.triggerReason')"
          min-width="180"
          show-overflow-tooltip
        />
        <el-table-column
          prop="lastErrorMsg"
          :label="t('option.lastError')"
          min-width="180"
          show-overflow-tooltip
        />
        <el-table-column :label="t('common.actions')" width="150" fixed="right">
          <template #default="{ row }">
            <el-button
              v-perm="'option:mmp:config'"
              link
              type="primary"
              @click="openDialog(row)"
            >
              {{ t('common.edit') }}
            </el-button>
            <el-button
              v-if="row.status === 2"
              v-perm="'option:mmp:reset'"
              link
              type="warning"
              @click="reset(row)"
            >
              {{ t('option.resetMMP') }}
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
      :title="t('option.configureMMP')"
      width="680px"
      destroy-on-close
    >
      <el-alert
        :title="t('option.mmpConfigWarning')"
        type="warning"
        :closable="false"
        class="dialog-alert"
      />
      <el-form :model="form" label-width="180px">
        <el-form-item :label="t('option.mmpGroup')" required>
          <el-input v-model="form.groupCode" maxlength="32" :disabled="editing" />
        </el-form-item>
        <el-form-item :label="t('common.status')" required>
          <el-switch
            v-model="form.enabled"
            :active-value="1"
            :inactive-value="2"
            :active-text="t('common.enabled')"
            :inactive-text="t('common.disabled')"
          />
        </el-form-item>
        <el-form-item :label="t('option.mmpQtyThreshold')" required>
          <el-input v-model="form.qtyThreshold" />
        </el-form-item>
        <el-form-item :label="t('option.mmpTradeThreshold')" required>
          <el-input-number v-model="form.tradeCountThreshold" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('option.mmpLossThreshold')" required>
          <el-input v-model="form.lossThreshold" />
        </el-form-item>
        <el-form-item :label="t('option.mmpWindowSeconds')" required>
          <el-input-number
            v-model="form.windowSeconds"
            :min="1"
            :max="3600"
            :precision="0"
          />
        </el-form-item>
        <el-form-item :label="t('option.mmpCooldownSeconds')" required>
          <el-input-number
            v-model="form.cooldownSeconds"
            :min="0"
            :max="86400"
            :precision="0"
          />
        </el-form-item>
        <el-form-item :label="t('option.controlReason')" required>
          <el-input
            v-model="form.reason"
            type="textarea"
            :rows="2"
            maxlength="500"
            show-word-limit
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">
          {{ t('common.cancel') }}
        </el-button>
        <el-button type="primary" :loading="submitting" @click="submit">
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
import { optionService, type OptionMMPConfig } from '@/services'
import { usePagination } from '@/composables'
import { useAuthStore } from '@/stores/auth'
import TenantSelect from '@/components/TenantSelect.vue'
import UserSelect from '@/components/UserSelect.vue'
import ContractSelect from '@/components/ContractSelect.vue'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'

const { t } = useI18n()
const authStore = useAuthStore()
const loading = ref(false)
const submitting = ref(false)
const dialogVisible = ref(false)
const editing = ref(false)
const items = ref<OptionMMPConfig[]>([])
const pagination = usePagination<number>(20)
const query = reactive({
  tenantId: (authStore.isTenantUser ? authStore.profileTenantId : undefined) as number | undefined,
  userId: undefined as number | undefined,
  contractId: undefined as number | undefined,
})
const form = reactive({
  groupCode: '',
  enabled: 1,
  qtyThreshold: '0',
  tradeCountThreshold: 0,
  lossThreshold: '0',
  windowSeconds: 5,
  cooldownSeconds: 30,
  reason: '',
})

async function loadData() {
  loading.value = true
  try {
    const response = await optionService.listMMPConfigs({
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
  query.userId = undefined
  query.contractId = undefined
  pagination.resetAndLoad(loadData)
}

function openDialog(row?: OptionMMPConfig) {
  if (row) {
    query.tenantId = row.tenantId
    query.userId = row.userId
    query.contractId = row.contractId
  }
  if (!query.tenantId || !query.userId || !query.contractId) {
    ElMessage.error(t('option.selectMMPContext'))
    return
  }
  editing.value = Boolean(row)
  form.groupCode = row?.groupCode || ''
  form.enabled = row?.enabled || 1
  form.qtyThreshold = row?.qtyThreshold || '0'
  form.tradeCountThreshold = row?.tradeCountThreshold || 0
  form.lossThreshold = row?.lossThreshold || '0'
  form.windowSeconds = row?.windowSeconds || 5
  form.cooldownSeconds = row?.cooldownSeconds ?? 30
  form.reason = ''
  dialogVisible.value = true
}

async function submit() {
  if (
    !query.tenantId ||
    !query.userId ||
    !query.contractId ||
    !/^[A-Za-z0-9_-]{1,32}$/.test(form.groupCode) ||
    !form.reason.trim()
  ) {
    ElMessage.error(t('option.completeMMPForm'))
    return
  }
  if (
    form.enabled === 1 &&
    Number(form.qtyThreshold) <= 0 &&
    form.tradeCountThreshold <= 0 &&
    Number(form.lossThreshold) <= 0
  ) {
    ElMessage.error(t('option.mmpThresholdRequired'))
    return
  }
  submitting.value = true
  try {
    await optionService.upsertMMPConfig({
      tenantId: query.tenantId,
      userId: query.userId,
      contractId: query.contractId,
      ...form,
      reason: form.reason.trim(),
    })
    dialogVisible.value = false
    ElMessage.success(t('common.success'))
    await loadData()
  } finally {
    submitting.value = false
  }
}

async function reset(row: OptionMMPConfig) {
  const { value } = await ElMessageBox.prompt(t('option.resetMMPReason'), t('option.resetMMP'), {
    inputValidator: (input) => Boolean(input?.trim()) || t('option.reviewReasonRequired'),
  })
  await optionService.resetMMPConfig({
    tenantId: row.tenantId,
    userId: row.userId,
    contractId: row.contractId,
    groupCode: row.groupCode,
    reason: value.trim(),
  })
  ElMessage.success(t('common.success'))
  await loadData()
}

function statusLabel(status: number) {
  return (
    {
      1: t('option.mmpActive'),
      2: t('option.mmpTriggered'),
      3: t('option.mmpDisabled'),
    }[status] || String(status)
  )
}

function statusType(status: number) {
  return status === 1 ? 'success' : status === 2 ? 'danger' : 'info'
}

onMounted(loadData)
</script>
