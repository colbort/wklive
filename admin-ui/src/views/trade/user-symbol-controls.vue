<template>
  <div class="module-page">
    <CrudQueryCard :model="query" @search="load" @reset="resetQuery">
      <el-form-item :label="t('trade.tenantId')">
        <TenantSelect v-model="query.tenantId" />
      </el-form-item>
      <el-form-item :label="t('trade.userId')">
        <UserSelect v-model="query.userId" :tenant-id="query.tenantId || undefined" />
      </el-form-item>
      <el-form-item :label="t('trade.symbolId')">
        <SymbolSelect v-model="query.symbolId" :tenant-id="query.tenantId || undefined" />
      </el-form-item>
      <el-form-item :label="t('common.status')">
        <el-select v-model="query.enabled" clearable class="query-control">
          <el-option :label="t('common.enabled')" :value="1" />
          <el-option :label="t('common.disabled')" :value="2" />
        </el-select>
      </el-form-item>
      <template #actions>
        <el-button v-perm="'trade:user-symbol-limit:update'" type="primary" @click="openCreate">
          {{ t('common.add') }}
        </el-button>
      </template>
    </CrudQueryCard>

    <el-card shadow="never" class="table-card">
      <el-table v-loading="loading" :data="rows" stripe>
        <el-table-column prop="id" :label="t('trade.controlId')" width="100" />
        <el-table-column prop="tenantId" :label="t('trade.tenantId')" width="100" />
        <el-table-column prop="userId" :label="t('trade.userId')" width="110" />
        <el-table-column prop="symbolId" :label="t('trade.symbolId')" min-width="130" />
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
              v-perm="'trade:user-symbol-limit:update'"
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
      :title="editorEditing ? t('trade.editSymbolControl') : t('trade.addSymbolControl')"
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
          <el-form-item :label="t('trade.symbolId')" required>
            <SymbolSelect
              v-model="form.symbolId"
              :tenant-id="form.tenantId || undefined"
              :disabled="editorEditing"
            />
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
          <el-form-item :label="t('trade.maxOpenOrders')">
            <el-input-number v-model="form.maxOpenOrders" :min="0" :precision="0" />
          </el-form-item>
          <el-form-item :label="t('trade.minOrderQty')">
            <el-input v-model="form.minOrderQty" />
          </el-form-item>
          <el-form-item :label="t('trade.maxOrderQty')">
            <el-input v-model="form.maxOrderQty" />
          </el-form-item>
          <el-form-item :label="t('trade.minOrderNotional')">
            <el-input v-model="form.minOrderNotional" />
          </el-form-item>
          <el-form-item :label="t('trade.maxOrderNotional')">
            <el-input v-model="form.maxOrderNotional" />
          </el-form-item>
          <el-form-item :label="t('trade.maxPositionQty')">
            <el-input v-model="form.maxPositionQty" />
          </el-form-item>
          <el-form-item :label="t('trade.maxPositionNotional')">
            <el-input v-model="form.maxPositionNotional" />
          </el-form-item>
          <el-form-item :label="t('trade.maxLongPositionQty')">
            <el-input v-model="form.maxLongPositionQty" />
          </el-form-item>
          <el-form-item :label="t('trade.maxShortPositionQty')">
            <el-input v-model="form.maxShortPositionQty" />
          </el-form-item>
          <el-form-item :label="t('trade.priceDeviationRate')">
            <el-input v-model="form.priceDeviationRate" />
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
  type RiskUserSymbolLimit,
  type SetUserSymbolLimitReq,
  type TradeUserControlAudit,
} from '@/services'
import TenantSelect from '@/components/TenantSelect.vue'
import UserSelect from '@/components/UserSelect.vue'
import SymbolSelect from '@/components/SymbolSelect.vue'
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

const query = reactive({ tenantId: 0, userId: 0, symbolId: 0, enabled: 0 })
const rows = ref<RiskUserSymbolLimit[]>([])
const loading = ref(false)
const editorVisible = ref(false)
const editorEditing = ref(false)
const saving = ref(false)
const startAt = ref<Date | null>(null)
const endAt = ref<Date | null>(null)
const auditVisible = ref(false)
const auditLoading = ref(false)
const auditRows = ref<TradeUserControlAudit[]>([])
const auditTarget = ref<RiskUserSymbolLimit | null>(null)
const detailVisible = ref(false)
const detailData = ref<unknown>(null)

function emptyForm(): SetUserSymbolLimitReq {
  return {
    tenantId: query.tenantId,
    userId: query.userId,
    symbolId: query.symbolId,
    controlMode: 1,
    maxPositionQty: '0',
    maxPositionNotional: '0',
    maxOpenOrders: 0,
    maxOrderQty: '0',
    maxOrderNotional: '0',
    minOrderQty: '0',
    minOrderNotional: '0',
    maxLongPositionQty: '0',
    maxShortPositionQty: '0',
    priceDeviationRate: '0',
    operatorId: 0,
    source: 3,
    enabled: 1,
    effectiveStartTime: 0,
    effectiveEndTime: 0,
    remark: '',
    expectedVersion: 0,
  }
}

const form = reactive<SetUserSymbolLimitReq>(emptyForm())

function applyForm(data?: RiskUserSymbolLimit) {
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
      symbolId: query.symbolId || undefined,
      enabled: query.enabled || undefined,
      scopeType: 2,
      cursor: pagination.cursor,
      limit: pagination.limit,
    })
    rows.value = (res.data || []).flatMap((entry) => (entry.symbolLimit ? [entry.symbolLimit] : []))
    updateFromResponse(res)
  } finally {
    loading.value = false
  }
}

function resetQuery() {
  Object.assign(query, { tenantId: 0, userId: 0, symbolId: 0, enabled: 0 })
  resetAndLoad(load)
}

function openCreate() {
  editorEditing.value = false
  applyForm()
  editorVisible.value = true
}

function openEdit(row: RiskUserSymbolLimit) {
  editorEditing.value = true
  applyForm(row)
  editorVisible.value = true
}

async function save() {
  if (!form.tenantId || !form.userId || !form.symbolId) {
    ElMessage.warning(t('trade.selectTenantAndUser'))
    return
  }
  form.effectiveStartTime = startAt.value?.getTime() || 0
  form.effectiveEndTime = endAt.value?.getTime() || 0
  saving.value = true
  try {
    await tradeService.setUserSymbolLimit({ ...form })
    ElMessage.success(t('common.success'))
    editorVisible.value = false
    await resetAndLoad(load)
  } finally {
    saving.value = false
  }
}

async function deleteControl(row: RiskUserSymbolLimit) {
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
    scopeType: 2,
    expectedVersion: row.version,
    reason: result.value.trim(),
  })
  ElMessage.success(t('trade.deleteSuccess'))
  await resetAndLoad(load)
}

async function openAudit(row: RiskUserSymbolLimit) {
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
      scopeType: 2,
      cursor: auditPagination.cursor,
      limit: auditPagination.limit,
    })
    auditRows.value = res.data || []
    updateAuditPagination(res)
  } finally {
    auditLoading.value = false
  }
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
