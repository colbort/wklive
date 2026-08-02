<template>
  <div class="module-page">
    <CrudQueryCard :model="query" @search="search" @reset="resetQuery">
      <el-form-item v-if="authStore.isSystemAdmin" :label="t('option.tenantId')">
        <TenantSelect v-model="query.tenantId" class="tenant-select-filter" />
      </el-form-item>
      <el-form-item :label="t('option.underlyingSymbol')">
        <el-input v-model="query.underlyingSymbol" clearable />
      </el-form-item>
      <el-form-item :label="t('option.corporateActionType')">
        <el-select v-model="query.actionType" clearable style="width: 180px">
          <el-option
            v-for="item in actionTypes"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
      </el-form-item>
      <el-form-item :label="t('common.status')">
        <el-select v-model="query.status" clearable style="width: 160px">
          <el-option
            v-for="item in statuses"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
      </el-form-item>
      <template #actions>
        <el-button v-perm="'option:corporate-action:create'" type="primary" @click="openCreate">
          {{ t('option.createCorporateAction') }}
        </el-button>
      </template>
    </CrudQueryCard>

    <el-card shadow="never" class="table-card">
      <el-table v-loading="loading" :data="actions" stripe>
        <el-table-column
          v-if="authStore.isSystemAdmin"
          prop="tenantId"
          :label="t('option.tenantId')"
          width="110"
        />
        <el-table-column prop="eventNo" :label="t('option.eventNo')" min-width="170" />
        <el-table-column
          prop="externalEventRef"
          :label="t('option.externalEventRef')"
          min-width="170"
        />
        <el-table-column prop="version" :label="t('option.version')" width="80" />
        <el-table-column
          prop="underlyingSymbol"
          :label="t('option.underlyingSymbol')"
          width="130"
        />
        <el-table-column :label="t('option.corporateActionType')" min-width="150">
          <template #default="{ row }">
            {{ actionTypeName(row.actionType) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('common.status')" width="130">
          <template #default="{ row }">
            {{ statusName(row.status) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('option.effectiveTime')" min-width="170">
          <template #default="{ row }">
            {{ formatTime(row.effectiveTime) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('option.migrationProgress')" min-width="190">
          <template #default="{ row }">
            <div v-for="mapping in row.contracts" :key="mapping.id">
              {{ mapping.sourceContractId }} → {{ mapping.successorContractId || 'MANUAL' }}:
              {{ mapping.positionCompleted }}/{{ mapping.positionTotal }}
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="evidenceRef" :label="t('option.evidenceRef')" min-width="180" />
        <el-table-column prop="lastErrorMsg" :label="t('option.lastError')" min-width="180" />
        <el-table-column :label="t('common.actions')" width="210" fixed="right">
          <template #default="{ row }">
            <template v-if="row.status === 1">
              <el-button
                v-perm="'option:corporate-action:review'"
                link
                type="success"
                @click="review(row, true)"
              >
                {{ t('option.approve') }}
              </el-button>
              <el-button
                v-perm="'option:corporate-action:review'"
                link
                type="danger"
                @click="review(row, false)"
              >
                {{ t('option.reject') }}
              </el-button>
            </template>
            <el-button
              v-perm="'option:corporate-action:position:list'"
              link
              @click="showPositions(row)"
            >
              {{ t('option.migrationDetails') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
      <CursorPagination
        v-model:limit="pagination.pagination.limit"
        :total="pagination.pagination.total"
        :has-prev="pagination.pagination.hasPrev"
        :has-next="pagination.pagination.hasNext"
        @prev="pagination.prevAndLoad(refresh)"
        @next="pagination.nextAndLoad(refresh)"
        @limit-change="pagination.resetAndLoad(refresh)"
      />
    </el-card>

    <el-dialog v-model="createVisible" :title="t('option.createCorporateAction')" width="760px">
      <el-form :model="form" label-width="160px">
        <el-form-item :label="t('option.eventNo')">
          <el-input v-model="form.eventNo" />
        </el-form-item>
        <el-form-item :label="t('option.externalEventRef')">
          <el-input v-model="form.externalEventRef" />
        </el-form-item>
        <el-form-item :label="t('option.underlyingSymbol')">
          <el-input v-model="form.underlyingSymbol" />
        </el-form-item>
        <el-form-item :label="t('option.corporateActionType')">
          <el-select v-model="form.actionType">
            <el-option
              v-for="item in actionTypes"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('option.announcementTime')">
          <el-date-picker v-model="form.announcementTime" type="datetime" />
        </el-form-item>
        <el-form-item :label="t('option.effectiveTime')">
          <el-date-picker v-model="form.effectiveTime" type="datetime" />
        </el-form-item>
        <el-form-item :label="t('option.contractMappings')">
          <el-input
            v-model="form.contractMappings"
            type="textarea"
            :rows="5"
            :placeholder="t('option.contractMappingFormat')"
          />
        </el-form-item>
        <el-alert
          :title="t('option.corporateActionAutoScope')"
          type="warning"
          :closable="false"
          class="form-alert"
        />
        <el-form-item :label="t('option.evidenceRef')">
          <el-input v-model="form.evidenceRef" />
        </el-form-item>
        <el-form-item :label="t('option.evidenceHash')">
          <el-input v-model="form.evidenceHash" />
        </el-form-item>
        <el-form-item :label="t('option.description')">
          <el-input v-model="form.description" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createVisible = false">
          {{ t('common.cancel') }}
        </el-button>
        <el-button type="primary" :loading="saving" @click="createAction">
          {{ t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="positionVisible" :title="t('option.migrationDetails')" width="1050px">
      <el-table v-loading="positionLoading" :data="positions" stripe>
        <el-table-column prop="sourcePositionId" :label="t('option.sourcePosition')" width="125" />
        <el-table-column
          prop="successorPositionId"
          :label="t('option.successorPosition')"
          width="135"
        />
        <el-table-column prop="userId" :label="t('option.userId')" width="105" />
        <el-table-column :label="t('option.quantityConversion')" min-width="180">
          <template #default="{ row }">
            {{ row.sourceQuantity }} → {{ row.successorQuantity }}
          </template>
        </el-table-column>
        <el-table-column :label="t('option.costBasisInvariant')" min-width="210">
          <template #default="{ row }">
            {{ row.costBasisBefore }} = {{ row.costBasisAfter }}
          </template>
        </el-table-column>
        <el-table-column prop="cashDifference" :label="t('option.cashDifference')" width="130" />
        <el-table-column prop="lastErrorMsg" :label="t('option.lastError')" min-width="170" />
      </el-table>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  optionService,
  type CorporateActionContractInput,
  type OptionCorporateAction,
  type OptionCorporateActionPosition,
} from '@/services'
import TenantSelect from '@/components/TenantSelect.vue'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'
import { useAuthStore } from '@/stores/auth'
import { usePagination } from '@/composables'

const { t } = useI18n()
const authStore = useAuthStore()
const loading = ref(false)
const saving = ref(false)
const positionLoading = ref(false)
const createVisible = ref(false)
const positionVisible = ref(false)
const actions = ref<OptionCorporateAction[]>([])
const positions = ref<OptionCorporateActionPosition[]>([])
const pagination = usePagination<number>(20)
const query = reactive({
  tenantId: (authStore.isTenantUser ? authStore.profileTenantId : undefined) as number | undefined,
  underlyingSymbol: '',
  actionType: undefined as number | undefined,
  status: undefined as number | undefined,
})
const form = reactive({
  eventNo: '',
  externalEventRef: '',
  underlyingSymbol: '',
  actionType: 1,
  announcementTime: new Date(),
  effectiveTime: new Date(Date.now() + 3600_000),
  contractMappings: '1001,1002,AUTO,2,1',
  evidenceRef: '',
  evidenceHash: '',
  description: '',
})

const actionTypes = computed(() =>
  [
    'split',
    'reverseSplit',
    'ordinaryDividend',
    'specialDividend',
    'merger',
    'spinOff',
    'fork',
    'airdrop',
    'delisting',
    'other',
  ].map((key, index) => ({ value: index + 1, label: t(`option.${key}`) })),
)
const statuses = computed(() =>
  ['draft', 'approved', 'rejected', 'executing', 'completed', 'manualReview', 'failed'].map(
    (key, index) => ({
      value: index + 1,
      label: t(`option.corporateAction${key[0].toUpperCase()}${key.slice(1)}`),
    }),
  ),
)

function resetQuery() {
  query.tenantId = authStore.isTenantUser ? authStore.profileTenantId || undefined : undefined
  query.underlyingSymbol = ''
  query.actionType = undefined
  query.status = undefined
  pagination.resetAndLoad(refresh)
}

function search() {
  pagination.resetAndLoad(refresh)
}

async function refresh() {
  loading.value = true
  try {
    const response = await optionService.listCorporateActions({
      tenantId: query.tenantId,
      underlyingSymbol: query.underlyingSymbol || undefined,
      actionType: query.actionType,
      status: query.status,
      cursor: pagination.pagination.cursor,
      limit: pagination.pagination.limit,
    })
    actions.value = response.data || []
    pagination.updateFromResponse(response)
  } finally {
    loading.value = false
  }
}

function openCreate() {
  form.announcementTime = new Date()
  form.effectiveTime = new Date(Date.now() + 3600_000)
  createVisible.value = true
}

function parseMappings(): CorporateActionContractInput[] {
  return form.contractMappings
    .split('\n')
    .filter((line) => line.trim())
    .map((line) => {
      const [source, successor, mode, numerator, denominator] = line
        .split(',')
        .map((item) => item.trim())
      if (!source || !mode || !numerator || !denominator) {
        throw new Error(t('option.invalidContractMapping'))
      }
      const executionMode =
        mode.toUpperCase() === 'AUTO' ? 1 : mode.toUpperCase() === 'MANUAL' ? 2 : 0
      if (!executionMode || (executionMode === 1 && !successor)) {
        throw new Error(t('option.invalidContractMapping'))
      }
      return {
        sourceContractId: Number(source),
        successorContractId: successor ? Number(successor) : undefined,
        executionMode,
        quantityNumerator: numerator,
        quantityDenominator: denominator,
      }
    })
}

async function createAction() {
  if (!query.tenantId) return
  saving.value = true
  try {
    await optionService.createCorporateAction({
      tenantId: query.tenantId,
      eventNo: form.eventNo,
      externalEventRef: form.externalEventRef,
      underlyingSymbol: form.underlyingSymbol,
      actionType: form.actionType,
      announcementTime: Math.floor(form.announcementTime.getTime() / 1000),
      effectiveTime: Math.floor(form.effectiveTime.getTime() / 1000),
      evidenceRef: form.evidenceRef,
      evidenceHash: form.evidenceHash,
      description: form.description,
      contracts: parseMappings(),
    })
    ElMessage.success(t('common.success'))
    createVisible.value = false
    await refresh()
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : String(error))
  } finally {
    saving.value = false
  }
}

async function review(row: OptionCorporateAction, approve: boolean) {
  const { value } = await ElMessageBox.prompt(
    t('option.reviewReason'),
    t('option.corporateActionReview'),
    { inputValidator: (text) => Boolean(text?.trim()) },
  )
  await optionService.reviewCorporateAction({
    tenantId: row.tenantId,
    actionId: row.id,
    approve,
    reason: value,
  })
  ElMessage.success(t('common.success'))
  await refresh()
}

async function showPositions(row: OptionCorporateAction) {
  positionVisible.value = true
  positionLoading.value = true
  try {
    const response = await optionService.listCorporateActionPositions({
      tenantId: row.tenantId,
      actionId: row.id,
      limit: 100,
    })
    positions.value = response.data || []
  } finally {
    positionLoading.value = false
  }
}

function actionTypeName(value: number) {
  return actionTypes.value.find((item) => item.value === value)?.label || String(value)
}

function statusName(value: number) {
  return statuses.value.find((item) => item.value === value)?.label || String(value)
}

function formatTime(value?: number) {
  return value ? new Date(value * 1000).toLocaleString() : '-'
}

onMounted(refresh)
</script>

<style scoped>
.table-card {
  margin-top: 16px;
}
.form-alert {
  margin-bottom: 18px;
}
</style>
