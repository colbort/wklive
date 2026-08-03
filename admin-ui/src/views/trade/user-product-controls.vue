<template>
  <div class="module-page">
    <CrudQueryCard :model="query" @search="load" @reset="resetQuery">
      <el-form-item :label="t('trade.tenantId')">
        <TenantSelect v-model="query.tenantId" />
      </el-form-item>
      <el-form-item :label="t('trade.userId')">
        <UserSelect v-model="query.userId" :tenant-id="query.tenantId || undefined" />
      </el-form-item>
      <el-form-item :label="t('trade.productType')">
        <el-select v-model="query.productType" clearable class="query-control">
          <el-option :label="t('trade.spot')" :value="1" />
          <el-option :label="t('trade.derivative')" :value="2" />
          <el-option :label="t('trade.seconds')" :value="3" />
        </el-select>
      </el-form-item>
      <el-form-item v-if="query.productType === 2" :label="t('trade.contractType')">
        <el-select v-model="query.contractType" class="query-control">
          <el-option :label="t('trade.allContracts')" :value="0" />
          <el-option :label="t('trade.perpetual')" :value="1" />
          <el-option :label="t('trade.delivery')" :value="2" />
        </el-select>
      </el-form-item>
      <el-form-item :label="t('common.status')">
        <el-select v-model="query.enabled" clearable class="query-control">
          <el-option :label="t('common.enabled')" :value="1" />
          <el-option :label="t('common.disabled')" :value="2" />
        </el-select>
      </el-form-item>
      <template #actions>
        <el-button v-perm="'trade:user-trade-limit:update'" type="primary" @click="openCreate">
          {{ t('common.add') }}
        </el-button>
      </template>
    </CrudQueryCard>

    <el-card shadow="never" class="table-card">
      <el-table v-loading="loading" :data="rows" stripe>
        <el-table-column prop="id" :label="t('trade.controlId')" width="100" />
        <el-table-column prop="tenantId" :label="t('trade.tenantId')" width="100" />
        <el-table-column prop="userId" :label="t('trade.userId')" width="110" />
        <el-table-column :label="t('trade.controlScope')" min-width="200">
          <template #default="{ row }">{{ productTarget(row) }}</template>
        </el-table-column>
        <el-table-column :label="t('trade.controlMode')" width="150">
          <template #default="{ row }">{{ controlModeText(row.controlMode) }}</template>
        </el-table-column>
        <el-table-column :label="t('trade.enabled')" width="100">
          <template #default="{ row }">
            <el-tag :type="row.enabled === 1 ? 'success' : 'info'">
              {{ row.enabled === 1 ? t('common.enabled') : t('common.disabled') }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="version" :label="t('trade.version')" width="90" />
        <el-table-column :label="t('trade.effectiveEndTime')" min-width="180">
          <template #default="{ row }">{{ formatTime(row.effectiveEndTime) }}</template>
        </el-table-column>
        <el-table-column :label="t('common.actions')" align="center" width="220" fixed="right">
          <template #default="{ row }">
            <el-button
              v-perm="'trade:user-trade-limit:update'"
              link
              type="primary"
              @click="openEdit(row)"
            >
              {{ t('common.edit') }}
            </el-button>
            <el-button link type="primary" @click="openAudit(row)">
              {{ t('trade.auditLog') }}
            </el-button>
            <el-button
              v-if="row.enabled === 1"
              v-perm="'trade:user-trade-control:disable'"
              link
              type="danger"
              @click="deleteControl(row)"
            >
              {{ t('common.delete') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
      <CursorPagination
        v-model:limit="pagination.limit"
        :total="pagination.total"
        :has-prev="pagination.hasPrev"
        :has-next="pagination.hasNext"
        @prev="prevAndLoad(load)"
        @next="nextAndLoad(load)"
        @limit-change="resetAndLoad(load)"
      />
    </el-card>

    <el-dialog
      v-model="editorVisible"
      :title="editorEditing ? t('trade.editProductControl') : t('trade.addProductControl')"
      width="920px"
      destroy-on-close
    >
      <el-form :model="form" label-width="150px">
        <div class="form-grid">
          <el-form-item :label="t('trade.tenantId')" required>
            <TenantSelect v-model="form.tenantId" :disabled="editorEditing" />
          </el-form-item>
          <el-form-item :label="t('trade.userId')" required>
            <UserSelect
              v-model="form.userId"
              :tenant-id="form.tenantId || undefined"
              :disabled="editorEditing"
            />
          </el-form-item>
          <el-form-item :label="t('trade.productType')" required>
            <el-select v-model="form.productType" :disabled="editorEditing" class="full-width">
              <el-option :label="t('trade.spot')" :value="1" />
              <el-option :label="t('trade.derivative')" :value="2" />
              <el-option :label="t('trade.seconds')" :value="3" />
            </el-select>
          </el-form-item>
          <el-form-item v-if="form.productType === 2" :label="t('trade.contractType')">
            <el-select v-model="form.contractType" :disabled="editorEditing" class="full-width">
              <el-option :label="t('trade.allContracts')" :value="0" />
              <el-option :label="t('trade.perpetual')" :value="1" />
              <el-option :label="t('trade.delivery')" :value="2" />
            </el-select>
          </el-form-item>
          <el-form-item :label="t('trade.controlMode')">
            <el-select v-model="form.controlMode" class="full-width">
              <el-option :label="t('trade.controlNormal')" :value="1" />
              <el-option :label="t('trade.controlCloseOnly')" :value="2" />
              <el-option :label="t('trade.controlReduceOnly')" :value="3" />
              <el-option :label="t('trade.controlDisabled')" :value="4" />
            </el-select>
          </el-form-item>
          <el-form-item :label="t('trade.enabled')">
            <el-switch v-model="form.enabled" :active-value="1" :inactive-value="2" />
          </el-form-item>
          <el-form-item :label="t('trade.tradeEnabled')">
            <el-switch v-model="form.tradeEnabled" :active-value="1" :inactive-value="2" />
          </el-form-item>
          <el-form-item :label="t('trade.onlyReduceOnly')">
            <el-switch v-model="form.onlyReduceOnly" :active-value="1" :inactive-value="2" />
          </el-form-item>
          <el-form-item :label="t('trade.canOpen')">
            <el-switch v-model="form.canOpen" :active-value="1" :inactive-value="0" />
          </el-form-item>
          <el-form-item :label="t('trade.canClose')">
            <el-switch v-model="form.canClose" :active-value="1" :inactive-value="0" />
          </el-form-item>
          <el-form-item :label="t('trade.canCancel')">
            <el-switch v-model="form.canCancel" :active-value="1" :inactive-value="0" />
          </el-form-item>
          <el-form-item :label="t('trade.canTriggerOrder')">
            <el-switch v-model="form.canTriggerOrder" :active-value="1" :inactive-value="0" />
          </el-form-item>
          <el-form-item :label="t('trade.canApiTrade')">
            <el-switch v-model="form.canApiTrade" :active-value="1" :inactive-value="0" />
          </el-form-item>
          <el-form-item :label="t('trade.maxOpenOrderCount')">
            <el-input-number v-model="form.maxOpenOrderCount" :min="0" :precision="0" />
          </el-form-item>
          <el-form-item :label="t('trade.maxOrderCountPerDay')">
            <el-input-number v-model="form.maxOrderCountPerDay" :min="0" :precision="0" />
          </el-form-item>
          <el-form-item :label="t('trade.maxCancelCountPerDay')">
            <el-input-number v-model="form.maxCancelCountPerDay" :min="0" :precision="0" />
          </el-form-item>
          <el-form-item :label="t('trade.maxOpenNotional')">
            <el-input v-model="form.maxOpenNotional" />
          </el-form-item>
          <el-form-item :label="t('trade.maxPositionNotional')">
            <el-input v-model="form.maxPositionNotional" />
          </el-form-item>
          <el-form-item :label="t('trade.effectiveStartTime')">
            <el-date-picker v-model="startAt" type="datetime" class="full-width" />
          </el-form-item>
          <el-form-item :label="t('trade.effectiveEndTime')">
            <el-date-picker v-model="endAt" type="datetime" class="full-width" />
          </el-form-item>
          <el-form-item :label="t('trade.remark')" class="wide">
            <el-input v-model="form.remark" type="textarea" :rows="2" />
          </el-form-item>
        </div>
      </el-form>
      <template #footer>
        <el-button @click="editorVisible = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="saving" @click="save">
          {{ t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="auditVisible" :title="t('trade.auditLog')" width="960px">
      <el-table v-loading="auditLoading" :data="auditRows" stripe>
        <el-table-column prop="id" :label="t('trade.auditId')" width="90" />
        <el-table-column prop="changeType" :label="t('trade.changeType')" width="110" />
        <el-table-column prop="operatorId" :label="t('trade.operatorId')" width="120" />
        <el-table-column prop="source" :label="t('trade.source')" width="100" />
        <el-table-column
          prop="reason"
          :label="t('trade.reason')"
          min-width="180"
          show-overflow-tooltip
        />
        <el-table-column :label="t('trade.createTimes')" min-width="180">
          <template #default="{ row }">{{ formatTime(row.createTimes) }}</template>
        </el-table-column>
        <el-table-column :label="t('common.detail')" width="90">
          <template #default="{ row }">
            <el-button link type="primary" @click="showDetail(row)">{{
              t('common.detail')
            }}</el-button>
          </template>
        </el-table-column>
      </el-table>
      <CursorPagination
        v-model:limit="auditPagination.limit"
        :total="auditPagination.total"
        :has-prev="auditPagination.hasPrev"
        :has-next="auditPagination.hasNext"
        @prev="auditPrevAndLoad(loadAudits)"
        @next="auditNextAndLoad(loadAudits)"
        @limit-change="auditResetAndLoad(loadAudits)"
      />
    </el-dialog>

    <el-dialog v-model="detailVisible" :title="t('common.detail')" width="760px">
      <pre class="detail-pre">{{ JSON.stringify(detailData, null, 2) }}</pre>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { usePagination } from '@/composables'
import {
  tradeService,
  type RiskUserTradeLimit,
  type SetUserTradeLimitReq,
  type TradeUserControlAudit,
} from '@/services'
import TenantSelect from '@/components/TenantSelect.vue'
import UserSelect from '@/components/UserSelect.vue'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'

const { t } = useI18n()
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)
const {
  pagination: auditPagination,
  updateFromResponse: updateAuditPagination,
  resetAndLoad: auditResetAndLoad,
  prevAndLoad: auditPrevAndLoad,
  nextAndLoad: auditNextAndLoad,
} = usePagination<number>(20)

const query = reactive({ tenantId: 0, userId: 0, productType: 0, contractType: 0, enabled: 0 })
const rows = ref<RiskUserTradeLimit[]>([])
const loading = ref(false)
const editorVisible = ref(false)
const editorEditing = ref(false)
const saving = ref(false)
const startAt = ref<Date | null>(null)
const endAt = ref<Date | null>(null)
const auditVisible = ref(false)
const auditLoading = ref(false)
const auditRows = ref<TradeUserControlAudit[]>([])
const auditTarget = ref<RiskUserTradeLimit | null>(null)
const detailVisible = ref(false)
const detailData = ref<unknown>(null)

function emptyForm(): SetUserTradeLimitReq {
  return {
    tenantId: query.tenantId,
    userId: query.userId,
    productType: query.productType || 1,
    contractType: query.productType === 2 ? query.contractType : 0,
    controlMode: 1,
    canOpen: 1,
    canClose: 1,
    canCancel: 1,
    canTriggerOrder: 1,
    canApiTrade: 1,
    tradeEnabled: 1,
    onlyReduceOnly: 2,
    maxOpenOrderCount: 0,
    maxOrderCountPerDay: 0,
    maxCancelCountPerDay: 0,
    maxOpenNotional: '0',
    maxPositionNotional: '0',
    riskLevel: 0,
    operatorId: 0,
    source: 3,
    enabled: 1,
    effectiveStartTime: 0,
    effectiveEndTime: 0,
    remark: '',
    expectedVersion: 0,
  }
}

const form = reactive<SetUserTradeLimitReq>(emptyForm())

function applyForm(data?: RiskUserTradeLimit) {
  Object.assign(form, emptyForm(), data || {}, { expectedVersion: data?.version || 0 })
  startAt.value = data?.effectiveStartTime ? new Date(data.effectiveStartTime) : null
  endAt.value = data?.effectiveEndTime ? new Date(data.effectiveEndTime) : null
}

async function load() {
  loading.value = true
  try {
    const res = await tradeService.listUserTradeControls({
      tenantId: query.tenantId || undefined,
      userId: query.userId || undefined,
      productType: query.productType || undefined,
      contractType: query.productType === 2 ? query.contractType || undefined : undefined,
      enabled: query.enabled || undefined,
      scopeType: 1,
      cursor: pagination.cursor,
      limit: pagination.limit,
    })
    rows.value = (res.data || []).flatMap((entry) =>
      entry.productLimit ? [entry.productLimit] : [],
    )
    updateFromResponse(res)
  } finally {
    loading.value = false
  }
}

function resetQuery() {
  Object.assign(query, { tenantId: 0, userId: 0, productType: 0, contractType: 0, enabled: 0 })
  resetAndLoad(load)
}

function openCreate() {
  editorEditing.value = false
  applyForm()
  editorVisible.value = true
}

function openEdit(row: RiskUserTradeLimit) {
  editorEditing.value = true
  applyForm(row)
  editorVisible.value = true
}

async function save() {
  if (!form.tenantId || !form.userId || !form.productType) {
    ElMessage.warning(t('trade.selectTenantAndUser'))
    return
  }
  if (form.productType !== 2) form.contractType = 0
  form.effectiveStartTime = startAt.value?.getTime() || 0
  form.effectiveEndTime = endAt.value?.getTime() || 0
  saving.value = true
  try {
    await tradeService.setUserTradeLimit({ ...form })
    ElMessage.success(t('common.success'))
    editorVisible.value = false
    await resetAndLoad(load)
  } finally {
    saving.value = false
  }
}

async function deleteControl(row: RiskUserTradeLimit) {
  let result
  try {
    result = await ElMessageBox.prompt(t('trade.deleteReasonPrompt'), t('common.delete'), {
      inputValidator: (text) =>
        Boolean(String(text || '').trim()) || t('trade.deleteReasonRequired'),
    })
  } catch {
    return
  }
  await tradeService.disableUserTradeControl({
    tenantId: row.tenantId,
    controlId: row.id,
    scopeType: 1,
    expectedVersion: row.version,
    reason: result.value.trim(),
  })
  ElMessage.success(t('trade.deleteSuccess'))
  await resetAndLoad(load)
}

async function openAudit(row: RiskUserTradeLimit) {
  auditTarget.value = row
  auditVisible.value = true
  await auditResetAndLoad(loadAudits)
}

async function loadAudits() {
  if (!auditTarget.value) return
  auditLoading.value = true
  try {
    const res = await tradeService.listUserTradeControlAudits({
      tenantId: auditTarget.value.tenantId,
      controlId: auditTarget.value.id,
      userId: auditTarget.value.userId,
      scopeType: 1,
      cursor: auditPagination.cursor,
      limit: auditPagination.limit,
    })
    auditRows.value = res.data || []
    updateAuditPagination(res)
  } finally {
    auditLoading.value = false
  }
}

function productTarget(row: RiskUserTradeLimit) {
  const product =
    [t('trade.spot'), t('trade.derivative'), t('trade.seconds')][row.productType - 1] ||
    row.productType
  if (row.productType !== 2) return product
  const contract =
    [t('trade.allContracts'), t('trade.perpetual'), t('trade.delivery')][row.contractType] ||
    row.contractType
  return `${product} / ${contract}`
}

function controlModeText(mode: number) {
  return (
    [
      t('trade.controlNormal'),
      t('trade.controlCloseOnly'),
      t('trade.controlReduceOnly'),
      t('trade.controlDisabled'),
    ][mode - 1] || '-'
  )
}

function formatTime(value?: number) {
  return value ? new Date(value).toLocaleString() : '-'
}

function showDetail(row: unknown) {
  detailData.value = row
  detailVisible.value = true
}

onMounted(load)
</script>

<style scoped>
.query-control,
.full-width {
  width: 100%;
}
.form-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0 24px;
}
.wide {
  grid-column: 1 / -1;
}
.detail-pre {
  white-space: pre-wrap;
  word-break: break-all;
}
@media (max-width: 1100px) {
  .form-grid {
    grid-template-columns: 1fr;
  }
  .wide {
    grid-column: auto;
  }
}
</style>
