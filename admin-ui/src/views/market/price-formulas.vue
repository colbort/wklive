<template>
  <div class="module-page price-formula-page">
    <CrudQueryCard :model="query" @search="load" @reset="reset">
      <el-form-item class="operation-query-item" :label="t('market.authority')">
        <el-input v-model="query.authority" clearable class="operation-query-control" />
      </el-form-item>
      <el-form-item class="operation-query-item" :label="t('market.snapshotKind')">
        <el-input v-model="query.snapshotKind" clearable class="operation-query-control" />
      </el-form-item>
      <el-form-item class="operation-query-item" :label="t('market.symbol')">
        <el-input v-model="query.symbol" clearable class="operation-query-control" />
      </el-form-item>
      <el-form-item class="operation-query-item" :label="t('common.status')">
        <el-select v-model="query.status" clearable class="operation-query-control">
          <el-option
            v-for="item in formulaStatuses"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
      </el-form-item>
      <template #actions>
        <el-button v-perm="'market:authority:list'" @click="openAuthorityManager">
          {{ t('market.manageAuthorities') }}
        </el-button>
        <el-button v-perm="'market:price-formula:create'" type="primary" @click="openCreate">
          {{ t('market.createFormula') }}
        </el-button>
      </template>
    </CrudQueryCard>

    <el-card shadow="never" class="table-card">
      <el-table v-loading="loading" :data="rows" stripe>
        <el-table-column prop="formulaNo" :label="t('market.formulaNo')" min-width="200" />
        <el-table-column prop="authority" :label="t('market.authority')" min-width="130" />
        <el-table-column prop="snapshotKind" :label="t('market.snapshotKind')" min-width="140" />
        <el-table-column prop="symbol" :label="t('market.symbol')" min-width="120" />
        <el-table-column prop="algorithm" :label="t('market.algorithm')" min-width="140">
          <template #default="{ row }">
            {{ algorithmLabel(row.algorithm) }}
          </template>
        </el-table-column>
        <el-table-column prop="formulaVersion" :label="t('market.formulaVersion')" min-width="130" />
        <el-table-column :label="t('common.status')" width="110">
          <template #default="{ row }">
            <el-tag :type="formulaStatusType(row.status)">
              {{ formulaStatusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="lastTargetTime" :label="t('market.lastTargetTime')" min-width="170">
          <template #default="{ row }">
            {{ formatTime(row.lastTargetTime) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('common.actions')" width="120" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openDetail(row)">
              {{ t('common.detail') }}
            </el-button>
            <el-button
              v-if="row.status !== 1 && row.status !== 3"
              v-perm="'market:price-formula:status'"
              link
              type="success"
              @click="changeStatus(row, 1)"
            >
              {{ t('market.activate') }}
            </el-button>
            <el-button
              v-if="row.status !== 3"
              v-perm="'market:price-formula:status'"
              link
              type="danger"
              @click="changeStatus(row, 3)"
            >
              {{ t('market.revoke') }}
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
      v-model="authorityManagerVisible"
      :title="t('market.manageAuthorities')"
      width="1000px"
    >
      <div class="authority-toolbar">
        <el-button v-perm="'market:authority:set'" type="primary" @click="openAuthorityCreate">
          {{ t('market.createAuthority') }}
        </el-button>
      </div>
      <el-table v-loading="authorityLoading" :data="authorityRows" stripe>
        <el-table-column prop="authority" :label="t('market.authority')" min-width="140" />
        <el-table-column prop="providerCode" :label="t('market.providerCode')" min-width="140" />
        <el-table-column prop="producerType" :label="t('market.producerType')" min-width="150" />
        <el-table-column :label="t('market.allowedKinds')" min-width="280">
          <template #default="{ row }">
            <el-tag
              v-for="kind in row.allowedKinds"
              :key="kind"
              class="authority-kind"
              effect="plain"
            >
              {{ kind }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="t('common.status')" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'">
              {{ authorityStatusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="t('common.actions')" width="90" fixed="right">
          <template #default="{ row }">
            <el-button
              v-perm="'market:authority:set'"
              link
              type="primary"
              @click="openAuthorityEdit(row)"
            >
              {{ t('common.edit') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
      <template #footer>
        <el-button @click="authorityManagerVisible = false">
          {{ t('common.close') }}
        </el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="authorityEditorVisible"
      :title="authorityForm.id ? t('market.editAuthority') : t('market.createAuthority')"
      width="620px"
      append-to-body
    >
      <el-alert
        v-if="authorityForm.id"
        :title="t('market.authorityImmutableHint')"
        type="warning"
        :closable="false"
        show-icon
        class="authority-alert"
      />
      <el-form :model="authorityForm" label-width="150px">
        <el-form-item :label="t('market.authority')" required>
          <el-input v-model="authorityForm.authority" :disabled="Boolean(authorityForm.id)" />
        </el-form-item>
        <el-form-item :label="t('market.providerCode')" required>
          <el-input v-model="authorityForm.providerCode" :disabled="Boolean(authorityForm.id)" />
        </el-form-item>
        <el-form-item :label="t('market.producerType')" required>
          <el-input v-model="authorityForm.producerType" :disabled="Boolean(authorityForm.id)" />
        </el-form-item>
        <el-form-item :label="t('market.allowedKinds')" required>
          <el-checkbox-group v-model="authorityForm.allowedKinds">
            <el-checkbox v-for="kind in authorityKinds" :key="kind" :value="kind">
              {{ kind }}
            </el-checkbox>
          </el-checkbox-group>
        </el-form-item>
        <el-form-item :label="t('common.status')" required>
          <el-radio-group v-model="authorityForm.status">
            <el-radio :value="1">{{ t('market.authorityEnabled') }}</el-radio>
            <el-radio :value="2">{{ t('market.authorityDisabled') }}</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="authorityEditorVisible = false">
          {{ t('common.cancel') }}
        </el-button>
        <el-button type="primary" :loading="authoritySaving" @click="saveAuthority">
          {{ t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="dialogVisible" :title="t('market.createFormula')" width="900px">
      <el-form :model="form" label-width="150px">
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item :label="t('market.formulaNo')" required>
              <el-input v-model="form.formulaNo" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('market.formulaVersion')" required>
              <el-input v-model="form.formulaVersion" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('market.authority')" required>
              <el-select v-model="form.authority" style="width: 100%">
                <el-option
                  v-for="item in outputAuthorities"
                  :key="item.authority"
                  :label="authorityOptionLabel(item)"
                  :value="item.authority"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('market.snapshotKind')" required>
              <el-select
                v-model="form.snapshotKind"
                style="width: 100%"
                @change="handleOutputSnapshotKindChange"
              >
                <el-option
                  v-for="item in priceEngineSnapshotKinds"
                  :key="item"
                  :label="item"
                  :value="item"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item :label="t('market.categoryCode')">
              <el-select v-model="form.categoryCode" filterable clearable style="width: 100%">
                <el-option
                  v-for="item in categoryOptions"
                  :key="item.value"
                  :label="item.label"
                  :value="item.value"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item :label="t('market.market')">
              <el-input v-model="form.market" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item :label="t('market.symbol')" required>
              <el-input v-model="form.symbol" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('market.algorithm')" required>
              <el-select v-model="form.algorithm" style="width: 100%">
                <el-option
                  v-for="algorithm in algorithms"
                  :key="algorithm.value"
                  :label="algorithm.label"
                  :value="algorithm.value"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('market.activateOnCreate')">
              <el-switch v-model="form.activate" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item :label="t('market.maxLookbackMs')">
              <el-input-number v-model="form.maxLookbackMs" :min="1" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item :label="t('market.maxDeviationBps')">
              <el-input-number v-model="form.maxDeviationBps" :min="0" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item :label="t('market.minInputCount')">
              <el-input-number
                v-model="form.minInputCount"
                :min="1"
                :max="Math.max(1, form.components.length)"
              />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item :label="t('market.intervalMs')">
              <el-input-number v-model="form.intervalMs" :min="1" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-divider>{{ t('market.components') }}</el-divider>
        <el-table :data="form.components" border>
          <el-table-column :label="t('market.authority')">
            <template #default="{ row }">
              <el-select
                v-model="row.authority"
                style="width: 100%"
                @change="handleComponentAuthorityChange(row)"
              >
                <el-option
                  v-for="item in componentAuthorities"
                  :key="item.authority"
                  :label="authorityOptionLabel(item)"
                  :value="item.authority"
                />
              </el-select>
            </template>
          </el-table-column>
          <el-table-column :label="t('market.snapshotKind')">
            <template #default="{ row }">
              <el-select v-model="row.snapshotKind" style="width: 100%">
                <el-option
                  v-for="item in componentSnapshotKinds(row.authority)"
                  :key="item"
                  :label="item"
                  :value="item"
                />
              </el-select>
            </template>
          </el-table-column>
          <el-table-column :label="t('market.categoryCode')">
            <template #default="{ row }">
              <el-select v-model="row.categoryCode" filterable clearable style="width: 100%">
                <el-option
                  v-for="item in categoryOptions"
                  :key="item.value"
                  :label="item.label"
                  :value="item.value"
                />
              </el-select>
            </template>
          </el-table-column>
          <el-table-column :label="t('market.market')">
            <template #default="{ row }">
              <el-input v-model="row.market" />
            </template>
          </el-table-column>
          <el-table-column :label="t('market.symbol')">
            <template #default="{ row }">
              <el-input v-model="row.symbol" />
            </template>
          </el-table-column>
          <el-table-column :label="t('market.weight')" width="130">
            <template #default="{ row }">
              <el-input v-model="row.weight" />
            </template>
          </el-table-column>
          <el-table-column width="70">
            <template #default="{ $index }">
              <el-button link type="danger" @click="form.components.splice($index, 1)">
                ×
              </el-button>
            </template>
          </el-table-column>
        </el-table>
        <el-button class="component-add" @click="addComponent">
          {{ t('market.addComponent') }}
        </el-button>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false"> {{ t('common.cancel') }} </el-button
        ><el-button type="primary" :loading="saving" @click="save">
          {{ t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="detailVisible" :title="t('market.formulaDetail')" width="900px">
      <el-descriptions v-if="detail" :column="2" border>
        <el-descriptions-item :label="t('market.formulaNo')">
          {{ detail.formulaNo }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('market.formulaVersion')">
          {{ detail.formulaVersion }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('market.authority')">
          {{ detail.authority }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('market.snapshotKind')">
          {{ detail.snapshotKind }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('market.categoryCode')">
          {{ detail.categoryCode || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('market.market')">
          {{ detail.market || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('market.symbol')">
          {{ detail.symbol }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('market.algorithm')">
          {{ algorithmLabel(detail.algorithm) }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('common.status')">
          <el-tag :type="formulaStatusType(detail.status)">
            {{ formulaStatusLabel(detail.status) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item :label="t('market.lastTargetTime')">
          {{ formatTime(detail.lastTargetTime) }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('market.maxLookbackMs')">
          {{ detail.maxLookbackMs }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('market.maxDeviationBps')">
          {{ detail.maxDeviationBps }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('market.minInputCount')">
          {{ detail.minInputCount }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('market.intervalMs')">
          {{ detail.intervalMs }}
        </el-descriptions-item>
      </el-descriptions>
      <el-divider>{{ t('market.components') }}</el-divider>
      <el-table :data="detail?.components || []" border>
        <el-table-column prop="authority" :label="t('market.authority')" min-width="130" />
        <el-table-column prop="snapshotKind" :label="t('market.snapshotKind')" min-width="130" />
        <el-table-column prop="categoryCode" :label="t('market.categoryCode')" min-width="120" />
        <el-table-column prop="market" :label="t('market.market')" min-width="100" />
        <el-table-column prop="symbol" :label="t('market.symbol')" min-width="120" />
        <el-table-column prop="weight" :label="t('market.weight')" width="100" />
      </el-table>
      <template #footer>
        <el-button @click="detailVisible = false">
          {{ t('common.close') }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'
import { usePagination } from '@/composables'
import { getCoreOptions } from '@/stores/core'
import type { OptionGroup } from '@/services'
import { findOptionGroup, getOptionLabel } from '@/utils/options'
import { apiMarketCategoryList } from '@/api/market/categories'
import type { MarketCategory } from '@/services/market/CategoriesService'
import {
  apiChangePriceFormulaStatus,
  apiCreatePriceFormula,
  apiListAuthorityRegistries,
  apiListPriceFormulas,
  apiSetAuthorityRegistry,
  type AuthorityRegistry,
  type CreatePriceFormulaReq,
  type PriceFormula,
  type SetAuthorityRegistryReq,
} from '@/api/market/price-engine'

const { t } = useI18n()
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)
const loading = ref(false),
  saving = ref(false),
  dialogVisible = ref(false),
  detailVisible = ref(false),
  detail = ref<PriceFormula | null>(null),
  rows = ref<PriceFormula[]>([])
const authorityManagerVisible = ref(false)
const authorityEditorVisible = ref(false)
const authorityLoading = ref(false)
const authoritySaving = ref(false)
const authorityRows = ref<AuthorityRegistry[]>([])
const authorityOriginalStatus = ref<1 | 2>(1)
const query = reactive({
  authority: '',
  snapshotKind: '',
  symbol: '',
  status: undefined as number | undefined,
})
const optionGroups = ref<OptionGroup[]>([])
const categories = ref<MarketCategory[]>([])
const authorityRegistries = ref<AuthorityRegistry[]>([])
const authorityKinds = ['FINAL_QUOTE', 'INDEX', 'MARK', 'FUNDING', 'DELIVERY']
const formulaSnapshotKindOrder = ['MARK', 'INDEX', 'FUNDING', 'DELIVERY']
const categoryOptions = computed(() =>
  categories.value.map((item) => ({
    value: item.categoryCode,
    label: `${item.categoryName} (${item.categoryCode})`,
  })),
)
const algorithms = computed(() =>
  findOptionGroup(optionGroups.value, 'priceAlgorithm').map((item) => ({
    value: item.value,
    label: getOptionLabel(t, item.code, item.value),
  })),
)
const formulaStatuses = computed(() => [
  { value: 1, label: t('market.active') },
  { value: 2, label: t('market.inactive') },
  { value: 3, label: t('market.revoked') },
])
const emptyAuthorityForm = (): SetAuthorityRegistryReq => ({
  authority: '',
  providerCode: '',
  producerType: '',
  allowedKinds: ['FINAL_QUOTE'],
  status: 1,
})
const authorityForm = reactive<SetAuthorityRegistryReq>(emptyAuthorityForm())
const emptyForm = (): CreatePriceFormulaReq => ({
  formulaNo: '',
  authority: 'price-engine',
  snapshotKind: 'MARK',
  categoryCode: '',
  market: '',
  symbol: '',
  algorithm: 1,
  formulaVersion: '',
  components: [],
  maxLookbackMs: 60000,
  maxDeviationBps: 500,
  minInputCount: 1,
  intervalMs: 1000,
  activate: false,
})
const form = reactive<CreatePriceFormulaReq>(emptyForm())
const enabledAuthorityRegistries = computed(() =>
  authorityRegistries.value.filter((item) => item.status === 1),
)
const priceEngineSnapshotKinds = computed(() =>
  formulaSnapshotKindOrder.filter((kind) =>
    enabledAuthorityRegistries.value.some((item) => item.allowedKinds.includes(kind)),
  ),
)
const outputAuthorities = computed(() =>
  enabledAuthorityRegistries.value.filter((item) => item.allowedKinds.includes(form.snapshotKind)),
)
const componentAuthorities = computed(() =>
  enabledAuthorityRegistries.value.filter((item) => item.allowedKinds.length > 0),
)
function formulaStatusLabel(status: number) {
  return formulaStatuses.value.find((item) => item.value === status)?.label || String(status)
}
function authorityStatusLabel(status: number) {
  return status === 1 ? t('market.authorityEnabled') : t('market.authorityDisabled')
}
function algorithmLabel(algorithm: number) {
  return algorithms.value.find((item) => item.value === algorithm)?.label || algorithm
}
async function loadOptions() {
  const [coreOptions, categoryList, authorityList] = await Promise.all([
    getCoreOptions(),
    apiMarketCategoryList({ enabled: 1, cursor: 0, limit: 100 }),
    apiListAuthorityRegistries({ status: 1, cursor: 0, limit: 200 }),
  ])
  optionGroups.value = coreOptions.data || []
  categories.value = categoryList.data || []
  authorityRegistries.value = authorityList.data || []
}
async function loadAuthorityRows() {
  authorityLoading.value = true
  try {
    const res = await apiListAuthorityRegistries({ cursor: 0, limit: 200 })
    authorityRows.value = res.data || []
  } finally {
    authorityLoading.value = false
  }
}
async function refreshEnabledAuthorities() {
  const res = await apiListAuthorityRegistries({
    status: 1,
    cursor: 0,
    limit: 200,
  })
  authorityRegistries.value = res.data || []
}
async function openAuthorityManager() {
  authorityManagerVisible.value = true
  await loadAuthorityRows()
}
function openAuthorityCreate() {
  Object.assign(authorityForm, emptyAuthorityForm())
  authorityOriginalStatus.value = 1
  authorityEditorVisible.value = true
}
function openAuthorityEdit(row: AuthorityRegistry) {
  Object.assign(authorityForm, {
    id: row.id,
    authority: row.authority,
    providerCode: row.providerCode,
    producerType: row.producerType,
    allowedKinds: [...row.allowedKinds],
    status: row.status as 1 | 2,
    version: row.version,
  })
  authorityOriginalStatus.value = row.status as 1 | 2
  authorityEditorVisible.value = true
}
async function saveAuthority() {
  if (
    !authorityForm.authority.trim() ||
    !authorityForm.providerCode.trim() ||
    !authorityForm.producerType.trim() ||
    authorityForm.allowedKinds.length === 0
  ) {
    ElMessage.warning(t('market.authorityRequired'))
    return
  }
  if (authorityForm.id && authorityOriginalStatus.value === 1 && authorityForm.status === 2) {
    await ElMessageBox.confirm(t('market.disableAuthorityConfirm'))
  }
  authoritySaving.value = true
  try {
    await apiSetAuthorityRegistry({
      ...authorityForm,
      authority: authorityForm.authority.trim(),
      providerCode: authorityForm.providerCode.trim(),
      producerType: authorityForm.producerType.trim(),
      allowedKinds: [...authorityForm.allowedKinds],
    })
    ElMessage.success(t('common.success'))
    authorityEditorVisible.value = false
    await Promise.all([loadAuthorityRows(), refreshEnabledAuthorities()])
  } finally {
    authoritySaving.value = false
  }
}
function formulaStatusType(status: number) {
  return status === 1 ? 'success' : status === 3 ? 'danger' : 'info'
}
function formatTime(value: number) {
  return value > 0 ? new Date(value).toLocaleString() : '-'
}
async function load() {
  loading.value = true
  try {
    const res = await apiListPriceFormulas({
      ...query,
      cursor: pagination.cursor,
      limit: pagination.limit,
    })
    rows.value = res.data || []
    updateFromResponse(res)
  } finally {
    loading.value = false
  }
}
function reset() {
  query.authority = ''
  query.snapshotKind = ''
  query.symbol = ''
  query.status = undefined
  resetAndLoad(load)
}
function openCreate() {
  Object.assign(form, emptyForm())
  handleOutputSnapshotKindChange()
  addComponent()
  dialogVisible.value = true
}
function openDetail(row: PriceFormula) {
  detail.value = row
  detailVisible.value = true
}
function addComponent() {
  const preferred = enabledAuthorityRegistries.value.find(
    (item) => item.authority === 'market-ws' && item.allowedKinds.includes('FINAL_QUOTE'),
  )
  const registry =
    preferred ||
    enabledAuthorityRegistries.value.find((item) => item.allowedKinds.includes('FINAL_QUOTE')) ||
    enabledAuthorityRegistries.value[0]
  form.components.push({
    authority: registry?.authority || '',
    snapshotKind: registry?.allowedKinds[0] || '',
    categoryCode: '',
    market: '',
    symbol: '',
    weight: '1',
  })
}
function componentSnapshotKinds(authority: string) {
  return (
    enabledAuthorityRegistries.value.find((item) => item.authority === authority)?.allowedKinds ||
    []
  )
}
function authorityOptionLabel(item: AuthorityRegistry) {
  return `${item.authority} (${item.providerCode} / ${item.producerType})`
}
function handleOutputSnapshotKindChange() {
  const authority = outputAuthorities.value.find((item) => item.authority === form.authority)
  if (!authority) {
    form.authority = outputAuthorities.value[0]?.authority || ''
  }
}
function handleComponentAuthorityChange(row: CreatePriceFormulaReq['components'][number]) {
  const kinds = componentSnapshotKinds(row.authority)
  if (!kinds.includes(row.snapshotKind)) {
    row.snapshotKind = kinds[0]
  }
}
async function save() {
  if (
    !form.formulaNo ||
    !form.authority ||
    !form.snapshotKind ||
    !form.symbol ||
    !form.formulaVersion ||
    form.components.length === 0
  ) {
    ElMessage.warning(t('market.formulaRequired'))
    return
  }
  saving.value = true
  try {
    await apiCreatePriceFormula(form)
    ElMessage.success(t('common.success'))
    dialogVisible.value = false
    load()
  } finally {
    saving.value = false
  }
}
async function changeStatus(row: PriceFormula, status: 1 | 3) {
  await ElMessageBox.confirm(
    status === 1 ? t('market.activateFormulaConfirm') : t('market.revokeFormulaConfirm'),
  )
  await apiChangePriceFormulaStatus(row.id, status)
  ElMessage.success(t('common.success'))
  load()
}
onMounted(async () => {
  await Promise.all([loadOptions(), load()])
})
</script>

<style scoped>
.authority-toolbar {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 16px;
}

.authority-kind {
  margin: 2px 6px 2px 0;
}

.authority-alert {
  margin-bottom: 18px;
}
</style>
