<template>
  <div class="module-page">
    <CrudQueryCard :model="query" @search="search" @reset="resetQuery">
      <el-form-item :label="t('asset.tenantId')">
        <TenantSelect v-model="query.tenantId" class="tenant-select-filter" />
      </el-form-item>
      <el-form-item :label="t('asset.coin')">
        <el-input v-model="query.coin" clearable placeholder="USDT" />
      </el-form-item>
      <el-form-item :label="t('common.status')">
        <el-select v-model="query.status" clearable style="width: 160px">
          <el-option :label="t('asset.backstopDraft')" :value="1" />
          <el-option :label="t('asset.backstopApproved')" :value="2" />
          <el-option :label="t('asset.backstopRejected')" :value="3" />
        </el-select>
      </el-form-item>
      <template #actions>
        <el-button
          v-perm="'asset:platform-backstop-policy:create'"
          type="primary"
          :disabled="!query.tenantId"
          @click="openCreate"
        >
          {{ t('asset.createBackstopPolicy') }}
        </el-button>
      </template>
    </CrudQueryCard>

    <el-alert
      v-if="!query.tenantId"
      :title="t('asset.selectTenantForBackstopPolicy')"
      type="info"
      :closable="false"
    />

    <el-card v-else shadow="never" class="table-card">
      <el-alert
        :title="t('asset.backstopFourEyesNotice')"
        type="warning"
        :closable="false"
        class="policy-alert"
      />
      <el-table v-loading="loading" :data="rows" stripe>
        <el-table-column prop="id" :label="t('common.id')" width="90" />
        <el-table-column prop="coin" :label="t('asset.coin')" width="90" />
        <el-table-column prop="version" :label="t('asset.version')" width="90" />
        <el-table-column :label="t('asset.backstopMode')" width="135">
          <template #default="{ row }">
            {{ modeName(row.mode) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('common.status')" width="110">
          <template #default="{ row }">
            <el-tag :type="statusTag(row.status)">
              {{ statusName(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          prop="perRequestLimit"
          :label="t('asset.perRequestLimit')"
          min-width="135"
        />
        <el-table-column prop="dailyLimit" :label="t('asset.dailyLimit')" min-width="130" />
        <el-table-column prop="balanceFloor" :label="t('asset.balanceFloor')" min-width="130" />
        <el-table-column :label="t('asset.effectivePeriod')" min-width="250">
          <template #default="{ row }">
            {{ formatDate(row.effectiveFrom) }} — {{ formatDate(row.effectiveUntil) }}
          </template>
        </el-table-column>
        <el-table-column
          prop="requestNo"
          :label="t('asset.requestNo')"
          min-width="210"
          show-overflow-tooltip
        />
        <el-table-column
          prop="evidenceRef"
          :label="t('asset.evidenceRef')"
          min-width="210"
          show-overflow-tooltip
        />
        <el-table-column :label="t('common.actions')" width="220" fixed="right">
          <template #default="{ row }">
            <el-button
              v-perm="'asset:platform-backstop-policy:detail'"
              link
              @click="showDetail(row)"
            >
              {{ t('common.detail') }}
            </el-button>
            <template v-if="row.status === 1">
              <el-button
                v-perm="'asset:platform-backstop-policy:review'"
                link
                type="success"
                @click="review(row, true)"
              >
                {{ t('asset.approveBackstopPolicy') }}
              </el-button>
              <el-button
                v-perm="'asset:platform-backstop-policy:review'"
                link
                type="danger"
                @click="review(row, false)"
              >
                {{ t('asset.rejectBackstopPolicy') }}
              </el-button>
            </template>
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

    <el-dialog
      v-model="createVisible"
      :title="t('asset.createBackstopPolicy')"
      width="760px"
      :close-on-click-modal="false"
    >
      <el-alert
        :title="t('asset.backstopImmutableNotice')"
        type="warning"
        :closable="false"
        class="policy-alert"
      />
      <el-form :model="form" label-width="155px">
        <el-form-item :label="t('asset.tenantId')">
          <el-input-number v-model="form.tenantId" :disabled="true" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('asset.coin')">
          <el-input v-model="form.coin" placeholder="USDT" />
        </el-form-item>
        <el-form-item :label="t('asset.requestNo')">
          <el-input v-model="form.requestNo" />
        </el-form-item>
        <el-form-item :label="t('asset.backstopMode')">
          <el-select v-model="form.mode" style="width: 100%" @change="normalizeLimits">
            <el-option :label="t('asset.backstopDisabled')" :value="1" />
            <el-option :label="t('asset.backstopPrefunded')" :value="2" />
            <el-option :label="t('asset.backstopCreditFloor')" :value="3" />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('asset.perRequestLimit')">
          <el-input v-model="form.perRequestLimit" />
        </el-form-item>
        <el-form-item :label="t('asset.dailyLimit')">
          <el-input v-model="form.dailyLimit" />
        </el-form-item>
        <el-form-item :label="t('asset.balanceFloor')">
          <el-input v-model="form.balanceFloor" />
        </el-form-item>
        <el-form-item :label="t('asset.effectiveFrom')">
          <el-date-picker v-model="form.effectiveFrom" type="datetime" style="width: 100%" />
        </el-form-item>
        <el-form-item :label="t('asset.effectiveUntil')">
          <el-date-picker v-model="form.effectiveUntil" type="datetime" style="width: 100%" />
        </el-form-item>
        <el-form-item :label="t('asset.reason')">
          <el-input v-model="form.reason" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item :label="t('asset.evidenceRef')">
          <el-input v-model="form.evidenceRef" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createVisible = false">
          {{ t('common.cancel') }}
        </el-button>
        <el-button type="primary" :loading="saving" @click="createPolicy">
          {{ t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>

    <el-drawer v-model="detailVisible" :title="t('asset.backstopPolicyDetail')" size="760px">
      <pre class="detail-pre">{{ JSON.stringify(detail, null, 2) }}</pre>
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { usePagination } from '@/composables'
import { assetService, type PlatformBackstopPolicy } from '@/services'
import { formatDate } from '@/utils'
import TenantSelect from '@/components/TenantSelect.vue'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'

const { t } = useI18n()
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)
const loading = ref(false)
const saving = ref(false)
const rows = ref<PlatformBackstopPolicy[]>([])
const createVisible = ref(false)
const detailVisible = ref(false)
const detail = ref<PlatformBackstopPolicy | null>(null)
const query = reactive({
  tenantId: undefined as number | undefined,
  coin: '',
  status: undefined as number | undefined,
})
const form = reactive({
  tenantId: 0,
  coin: 'USDT',
  requestNo: '',
  mode: 1,
  perRequestLimit: '0',
  dailyLimit: '0',
  balanceFloor: '0',
  effectiveFrom: null as Date | null,
  effectiveUntil: null as Date | null,
  reason: '',
  evidenceRef: '',
})

async function loadList() {
  if (!query.tenantId) {
    rows.value = []
    return
  }
  loading.value = true
  try {
    const res = await assetService.listPlatformBackstopPolicies({
      tenantId: query.tenantId,
      coin: query.coin || undefined,
      status: query.status,
      cursor: pagination.cursor,
      limit: pagination.limit,
    })
    rows.value = res.data || []
    updateFromResponse(res)
  } finally {
    loading.value = false
  }
}

function search() {
  resetAndLoad(loadList)
}

function resetQuery() {
  query.tenantId = undefined
  query.coin = ''
  query.status = undefined
  rows.value = []
  resetAndLoad(loadList)
}

function openCreate() {
  if (!query.tenantId) return
  Object.assign(form, {
    tenantId: query.tenantId,
    coin: query.coin || 'USDT',
    requestNo: `BACKSTOP-${query.tenantId}-${Date.now()}`,
    mode: 1,
    perRequestLimit: '0',
    dailyLimit: '0',
    balanceFloor: '0',
    effectiveFrom: new Date(Date.now() + 5 * 60_000),
    effectiveUntil: new Date(Date.now() + 24 * 60 * 60_000),
    reason: '',
    evidenceRef: '',
  })
  createVisible.value = true
}

function normalizeLimits() {
  if (form.mode === 1) {
    form.perRequestLimit = '0'
    form.dailyLimit = '0'
    form.balanceFloor = '0'
  } else if (form.mode === 2) {
    form.balanceFloor = '0'
  }
}

async function createPolicy() {
  if (!form.effectiveFrom || !form.effectiveUntil) {
    ElMessage.error(t('asset.backstopEffectiveTimeRequired'))
    return
  }
  saving.value = true
  try {
    await assetService.createPlatformBackstopPolicy({
      ...form,
      effectiveFrom: form.effectiveFrom.getTime(),
      effectiveUntil: form.effectiveUntil.getTime(),
    })
    ElMessage.success(t('asset.backstopPolicyCreated'))
    createVisible.value = false
    search()
  } finally {
    saving.value = false
  }
}

async function review(row: PlatformBackstopPolicy, approve: boolean) {
  const { value } = await ElMessageBox.prompt(
    t(approve ? 'asset.approveBackstopPrompt' : 'asset.rejectBackstopPrompt'),
    t(approve ? 'asset.approveBackstopPolicy' : 'asset.rejectBackstopPolicy'),
    { inputType: 'textarea', inputValidator: (value) => Boolean(value?.trim()) },
  )
  await assetService.reviewPlatformBackstopPolicy({
    tenantId: row.tenantId,
    policyId: row.id,
    approve,
    reason: value.trim(),
  })
  ElMessage.success(t('asset.backstopPolicyReviewed'))
  loadList()
}

async function showDetail(row: PlatformBackstopPolicy) {
  detail.value = (await assetService.getPlatformBackstopPolicy(row.tenantId, row.id)).data || row
  detailVisible.value = true
}

function modeName(mode: number) {
  return t(
    mode === 1
      ? 'asset.backstopDisabled'
      : mode === 2
        ? 'asset.backstopPrefunded'
        : 'asset.backstopCreditFloor',
  )
}

function statusName(status: number) {
  return t(
    status === 1
      ? 'asset.backstopDraft'
      : status === 2
        ? 'asset.backstopApproved'
        : 'asset.backstopRejected',
  )
}

function statusTag(status: number): 'warning' | 'success' | 'danger' {
  return status === 1 ? 'warning' : status === 2 ? 'success' : 'danger'
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

onMounted(loadList)
</script>

<style scoped>
.policy-alert {
  margin-bottom: 16px;
}
.detail-pre {
  white-space: pre-wrap;
  word-break: break-all;
}
</style>
