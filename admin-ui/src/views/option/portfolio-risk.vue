<template>
  <div class="module-page">
    <CrudQueryCard :model="query" @search="search" @reset="resetQuery">
      <el-form-item v-if="authStore.isSystemAdmin" :label="t('option.tenantId')">
        <TenantSelect v-model="query.tenantId" />
      </el-form-item>
      <el-form-item :label="t('option.settleCoin')">
        <el-input v-model="query.settleCoin" clearable style="width: 140px" />
      </el-form-item>
      <template #actions>
        <el-button v-perm="'option:portfolio-risk:create'" type="primary" @click="openDialog()">
          {{ t('option.createPortfolioConfig') }}
        </el-button>
      </template>
    </CrudQueryCard>

    <el-card v-loading="loading" shadow="never" class="table-card">
      <el-alert
        :title="t('option.portfolioGovernanceWarning')"
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
        <el-table-column prop="settleCoin" :label="t('option.settleCoin')" width="100" />
        <el-table-column prop="version" :label="t('option.portfolioConfigVersion')" width="100" />
        <el-table-column :label="t('option.status')" width="110">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)">
              {{ statusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          prop="initialShockRate"
          :label="t('option.initialShockRate')"
          width="140"
        />
        <el-table-column
          prop="maintenanceShockRate"
          :label="t('option.maintenanceShockRate')"
          width="150"
        />
        <el-table-column
          prop="scenarioShocks"
          :label="t('option.scenarioShocks')"
          min-width="220"
          show-overflow-tooltip
        />
        <el-table-column
          prop="concentrationThreshold"
          :label="t('option.concentrationThreshold')"
          min-width="170"
        />
        <el-table-column :label="t('option.portfolioAddonRates')" min-width="150">
          <template #default="{ row }">
            {{ row.concentrationAddonRate }} / {{ row.liquidityAddonRate }}
          </template>
        </el-table-column>
        <el-table-column :label="t('option.effectivePeriod')" min-width="220">
          <template #default="{ row }">
            {{ formatTime(row.effectiveFrom) }} –
            {{ row.effectiveUntil ? formatTime(row.effectiveUntil) : '∞' }}
          </template>
        </el-table-column>
        <el-table-column :label="t('option.portfolioConfigLineage')" min-width="180">
          <template #default="{ row }">
            {{ t('option.portfolioConfigSource') }}: {{ row.sourceConfigId || '—' }} /
            {{ t('option.portfolioConfigSupersedes') }}: {{ row.supersedesId || '—' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="evidenceRef"
          :label="t('option.evidenceRef')"
          min-width="200"
          show-overflow-tooltip
        />
        <el-table-column :label="t('common.actions')" width="210" fixed="right">
          <template #default="{ row }">
            <template v-if="row.status === 1">
              <el-button
                v-perm="'option:portfolio-risk:review'"
                link
                type="success"
                @click="review(row, true)"
              >
                {{ t('option.approve') }}
              </el-button>
              <el-button
                v-perm="'option:portfolio-risk:review'"
                link
                type="danger"
                @click="review(row, false)"
              >
                {{ t('option.reject') }}
              </el-button>
            </template>
            <el-button
              v-if="[2, 4].includes(row.status)"
              v-perm="'option:portfolio-risk:create'"
              link
              type="warning"
              @click="openDialog(row)"
            >
              {{ t('option.createRollbackVersion') }}
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
      :title="
        form.sourceConfigId ? t('option.createRollbackVersion') : t('option.createPortfolioConfig')
      "
      width="720px"
      destroy-on-close
    >
      <el-alert
        :title="t('option.portfolioApprovalWarning')"
        type="warning"
        :closable="false"
        class="dialog-alert"
      />
      <el-form :model="form" label-width="210px">
        <el-form-item :label="t('option.settleCoin')" required>
          <el-input v-model="form.settleCoin" maxlength="16" />
        </el-form-item>
        <el-form-item :label="t('option.initialShockRate')" required>
          <el-input v-model="form.initialShockRate" :disabled="Boolean(form.sourceConfigId)" />
        </el-form-item>
        <el-form-item :label="t('option.maintenanceShockRate')" required>
          <el-input v-model="form.maintenanceShockRate" :disabled="Boolean(form.sourceConfigId)" />
        </el-form-item>
        <el-form-item :label="t('option.scenarioShocks')" required>
          <el-input v-model="form.scenarioShocks" :disabled="Boolean(form.sourceConfigId)" />
        </el-form-item>
        <el-form-item :label="t('option.concentrationThreshold')" required>
          <el-input
            v-model="form.concentrationThreshold"
            :disabled="Boolean(form.sourceConfigId)"
          />
        </el-form-item>
        <el-form-item :label="t('option.concentrationAddonRate')" required>
          <el-input
            v-model="form.concentrationAddonRate"
            :disabled="Boolean(form.sourceConfigId)"
          />
        </el-form-item>
        <el-form-item :label="t('option.liquidityAddonRate')" required>
          <el-input v-model="form.liquidityAddonRate" :disabled="Boolean(form.sourceConfigId)" />
        </el-form-item>
        <el-form-item :label="t('option.effectiveFromUnix')" required>
          <el-input-number v-model="form.effectiveFrom" :min="1" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('option.changeReason')" required>
          <el-input
            v-model="form.changeReason"
            type="textarea"
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
import { optionService, type OptionPortfolioRiskConfig } from '@/services'
import { usePagination } from '@/composables'
import { useAuthStore } from '@/stores/auth'
import TenantSelect from '@/components/TenantSelect.vue'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'

const { t } = useI18n()
const authStore = useAuthStore()
const loading = ref(false)
const submitting = ref(false)
const dialogVisible = ref(false)
const items = ref<OptionPortfolioRiskConfig[]>([])
const pagination = usePagination<number>(20)
const query = reactive({
  tenantId: (authStore.isTenantUser ? authStore.profileTenantId : undefined) as number | undefined,
  settleCoin: '',
})
const form = reactive({
  settleCoin: 'USDT',
  modelMethod: 1,
  initialShockRate: '0.3',
  maintenanceShockRate: '0.2',
  scenarioShocks: '-1,-0.5,-0.3,-0.2,0.2,0.3,0.5,1,4',
  concentrationThreshold: '100000',
  concentrationAddonRate: '0.1',
  liquidityAddonRate: '0.02',
  effectiveFrom: Math.floor(Date.now() / 1000),
  changeReason: '',
  evidenceRef: '',
  sourceConfigId: 0,
})

async function loadData() {
  loading.value = true
  try {
    const response = await optionService.listPortfolioRiskConfigs({
      tenantId: query.tenantId,
      settleCoin: query.settleCoin || undefined,
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
  query.settleCoin = ''
  pagination.resetAndLoad(loadData)
}

function openDialog(source?: OptionPortfolioRiskConfig) {
  if (source?.tenantId && !query.tenantId) query.tenantId = source.tenantId
  if (!query.tenantId) {
    ElMessage.error(t('option.selectTenantForControlEvents'))
    return
  }
  form.settleCoin = source?.settleCoin || query.settleCoin || 'USDT'
  form.modelMethod = 1
  form.initialShockRate = source?.initialShockRate || '0.3'
  form.maintenanceShockRate = source?.maintenanceShockRate || '0.2'
  form.scenarioShocks = source?.scenarioShocks || '-1,-0.5,-0.3,-0.2,0.2,0.3,0.5,1,4'
  form.concentrationThreshold = source?.concentrationThreshold || '100000'
  form.concentrationAddonRate = source?.concentrationAddonRate || '0.1'
  form.liquidityAddonRate = source?.liquidityAddonRate || '0.02'
  form.effectiveFrom = Math.floor(Date.now() / 1000) + 300
  form.changeReason = ''
  form.evidenceRef = ''
  form.sourceConfigId = source?.id || 0
  dialogVisible.value = true
}

async function submit() {
  if (
    !query.tenantId ||
    !form.settleCoin.trim() ||
    !form.changeReason.trim() ||
    !form.evidenceRef.trim() ||
    form.effectiveFrom <= 0
  ) {
    ElMessage.error(t('option.completePortfolioConfig'))
    return
  }
  submitting.value = true
  try {
    await optionService.createPortfolioRiskConfig({
      tenantId: query.tenantId,
      ...form,
      settleCoin: form.settleCoin.trim().toUpperCase(),
      changeReason: form.changeReason.trim(),
      evidenceRef: form.evidenceRef.trim(),
      sourceConfigId: form.sourceConfigId || undefined,
    })
    dialogVisible.value = false
    ElMessage.success(t('common.success'))
    await loadData()
  } finally {
    submitting.value = false
  }
}

async function review(row: OptionPortfolioRiskConfig, approve: boolean) {
  const { value } = await ElMessageBox.prompt(
    approve ? t('option.approvePortfolioReason') : t('option.rejectPortfolioReason'),
    approve ? t('option.approve') : t('option.reject'),
    { inputValidator: (input) => Boolean(input?.trim()) || t('option.reviewReasonRequired') },
  )
  await optionService.reviewPortfolioRiskConfig({
    tenantId: row.tenantId,
    configId: row.id,
    approve,
    reason: value.trim(),
  })
  ElMessage.success(t('common.success'))
  await loadData()
}

function statusLabel(status: number) {
  return (
    {
      1: t('option.pendingReview'),
      2: t('option.approved'),
      3: t('option.rejected'),
      4: t('option.superseded'),
    }[status] || String(status)
  )
}

function statusType(status: number) {
  return status === 2 ? 'success' : status === 1 ? 'warning' : status === 3 ? 'danger' : 'info'
}

function formatTime(timestamp: number) {
  return timestamp ? new Date(timestamp * 1000).toLocaleString() : '-'
}

onMounted(loadData)
</script>
