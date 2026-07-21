<template>
  <div class="module-page">
    <CrudQueryCard :model="query" @search="load" @reset="reset">
      <el-form-item :label="t('itick.authority')">
        <el-input
          v-model="query.authority"
          clearable
        />
      </el-form-item>
      <el-form-item :label="t('itick.snapshotKind')">
        <el-input
          v-model="query.snapshotKind"
          clearable
        />
      </el-form-item>
      <el-form-item :label="t('itick.symbol')">
        <el-input
          v-model="query.symbol"
          clearable
        />
      </el-form-item>
      <el-form-item :label="t('common.status')">
        <el-select v-model="query.status" clearable style="width: 140px">
          <el-option
            v-for="item in formulaStatuses"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
      </el-form-item>
      <template #actions>
        <el-button v-perm="'itick:price-formula:create'" type="primary" @click="openCreate">
          {{
            t('itick.createFormula')
          }}
        </el-button>
      </template>
    </CrudQueryCard>

    <el-card shadow="never" class="table-card">
      <el-table v-loading="loading" :data="rows" stripe>
        <el-table-column prop="formulaNo" :label="t('itick.formulaNo')" min-width="150" />
        <el-table-column prop="authority" :label="t('itick.authority')" min-width="130" />
        <el-table-column prop="snapshotKind" :label="t('itick.snapshotKind')" min-width="140" />
        <el-table-column prop="symbol" :label="t('itick.symbol')" min-width="120" />
        <el-table-column prop="algorithm" :label="t('itick.algorithm')" min-width="140" />
        <el-table-column prop="formulaVersion" :label="t('itick.formulaVersion')" min-width="130" />
        <el-table-column :label="t('common.status')" width="110">
          <template #default="{ row }">
            <el-tag :type="formulaStatusType(row.status)">
              {{
                formulaStatusLabel(row.status)
              }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="lastTargetTime" :label="t('itick.lastTargetTime')" min-width="170">
          <template #default="{ row }">
            {{ formatTime(row.lastTargetTime) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('common.actions')" width="180" fixed="right">
          <template #default="{ row }">
            <el-button
              v-if="row.status !== 1 && row.status !== 3"
              v-perm="'itick:price-formula:status'"
              link
              type="success"
              @click="changeStatus(row, 1)"
            >
              {{ t('itick.activate') }}
            </el-button>
            <el-button
              v-if="row.status !== 3"
              v-perm="'itick:price-formula:status'"
              link
              type="danger"
              @click="changeStatus(row, 3)"
            >
              {{ t('itick.revoke') }}
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

    <el-dialog v-model="dialogVisible" :title="t('itick.createFormula')" width="900px">
      <el-form :model="form" label-width="150px">
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item
              :label="t('itick.formulaNo')"
              required
            >
              <el-input v-model="form.formulaNo" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item
              :label="t('itick.formulaVersion')"
              required
            >
              <el-input v-model="form.formulaVersion" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item
              :label="t('itick.authority')"
              required
            >
              <el-input v-model="form.authority" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item
              :label="t('itick.snapshotKind')"
              required
            >
              <el-input v-model="form.snapshotKind" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item :label="t('itick.categoryCode')">
              <el-input v-model="form.categoryCode" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item :label="t('itick.market')">
              <el-input v-model="form.market" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item
              :label="t('itick.symbol')"
              required
            >
              <el-input v-model="form.symbol" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item
              :label="t('itick.algorithm')"
              required
            >
              <el-select
                v-model="form.algorithm"
                style="width: 100%"
              >
                <el-option
                  v-for="algorithm in algorithms"
                  :key="algorithm"
                  :value="algorithm"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('itick.activateOnCreate')">
              <el-switch v-model="form.activate" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item :label="t('itick.maxLookbackMs')">
              <el-input-number v-model="form.maxLookbackMs" :min="1" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item :label="t('itick.maxDeviationBps')">
              <el-input-number v-model="form.maxDeviationBps" :min="0" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item :label="t('itick.intervalMs')">
              <el-input-number v-model="form.intervalMs" :min="1" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-divider>{{ t('itick.components') }}</el-divider>
        <el-table :data="form.components" border>
          <el-table-column :label="t('itick.authority')">
            <template #default="{ row }">
              <el-input v-model="row.authority" />
            </template>
          </el-table-column>
          <el-table-column :label="t('itick.snapshotKind')">
            <template #default="{ row }">
              <el-input v-model="row.snapshotKind" />
            </template>
          </el-table-column>
          <el-table-column :label="t('itick.categoryCode')">
            <template #default="{ row }">
              <el-input v-model="row.categoryCode" />
            </template>
          </el-table-column>
          <el-table-column :label="t('itick.market')">
            <template #default="{ row }">
              <el-input v-model="row.market" />
            </template>
          </el-table-column>
          <el-table-column :label="t('itick.symbol')">
            <template #default="{ row }">
              <el-input v-model="row.symbol" />
            </template>
          </el-table-column>
          <el-table-column
            :label="t('itick.weight')"
            width="130"
          >
            <template #default="{ row }">
              <el-input v-model="row.weight" />
            </template>
          </el-table-column>
          <el-table-column width="70">
            <template #default="{ $index }">
              <el-button
                link
                type="danger"
                @click="form.components.splice($index, 1)"
              >
                ×
              </el-button>
            </template>
          </el-table-column>
        </el-table>
        <el-button class="component-add" @click="addComponent">
          {{
            t('itick.addComponent')
          }}
        </el-button>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">
          {{ t('common.cancel') }}
        </el-button><el-button type="primary" :loading="saving" @click="save">
          {{
            t('common.confirm')
          }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { usePagination } from '@/composables'
import {
  apiChangePriceFormulaStatus,
  apiCreatePriceFormula,
  apiListPriceFormulas,
  type CreatePriceFormulaReq,
  type PriceFormula,
} from '@/api/itick/price-engine'

const { t } = useI18n()
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)
const loading = ref(false),
  saving = ref(false),
  dialogVisible = ref(false),
  rows = ref<PriceFormula[]>([])
const query = reactive({
  authority: '',
  snapshotKind: '',
  symbol: '',
  status: undefined as number | undefined,
})
const algorithms = ['WEIGHTED_MEAN', 'MEDIAN', 'PREMIUM_RATE']
const formulaStatuses = computed(() => [
  { value: 1, label: t('itick.active') },
  { value: 2, label: t('itick.inactive') },
  { value: 3, label: t('itick.revoked') },
])
const emptyForm = (): CreatePriceFormulaReq => ({
  formulaNo: '',
  authority: '',
  snapshotKind: '',
  categoryCode: '',
  market: '',
  symbol: '',
  algorithm: 'WEIGHTED_MEAN',
  formulaVersion: '',
  components: [],
  maxLookbackMs: 60000,
  maxDeviationBps: 500,
  intervalMs: 1000,
  activate: false,
})
const form = reactive<CreatePriceFormulaReq>(emptyForm())
function formulaStatusLabel(status: number) {
  return formulaStatuses.value.find((item) => item.value === status)?.label || String(status)
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
  addComponent()
  dialogVisible.value = true
}
function addComponent() {
  form.components.push({
    authority: '',
    snapshotKind: '',
    categoryCode: '',
    market: '',
    symbol: '',
    weight: '1',
  })
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
    ElMessage.warning(t('itick.formulaRequired'))
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
    status === 1 ? t('itick.activateFormulaConfirm') : t('itick.revokeFormulaConfirm'),
  )
  await apiChangePriceFormulaStatus(row.id, status)
  ElMessage.success(t('common.success'))
  load()
}
onMounted(load)
</script>

<style scoped>
.component-add {
  margin-top: 12px;
}
</style>
