<template>
  <div class="module-page">
    <CrudQueryCard :model="query" @search="loadList" @reset="resetQuery">
      <el-form-item :label="t('staking.tenantId')">
        <TenantSelect v-model="query.tenantId" class="tenant-select-filter" />
      </el-form-item>
      <el-form-item :label="t('staking.reconciliationDate')">
        <el-date-picker
          v-model="query.reconciliationDate"
          type="date"
          value-format="YYYYMMDD"
          clearable
        />
      </el-form-item>
      <el-form-item :label="t('staking.coinSymbol')">
        <el-input v-model="query.coinSymbol" clearable />
      </el-form-item>
      <el-form-item :label="t('common.status')">
        <el-select v-model="query.status" clearable>
          <el-option
            v-for="item in reconciliationStatusOptions"
            :key="item.value"
            :label="getOptionLabel(t, item.code, item.value)"
            :value="item.value"
          />
        </el-select>
      </el-form-item>
    </CrudQueryCard>

    <el-card shadow="never" class="table-card">
      <el-table v-loading="loading" :data="rows" stripe>
        <el-table-column
          prop="reconciliationDate"
          :label="t('staking.reconciliationDate')"
          width="130"
        />
        <el-table-column prop="tenantId" :label="t('staking.tenantId')" width="110" />
        <el-table-column prop="coinSymbol" :label="t('staking.coinSymbol')" width="100" />
        <el-table-column
          prop="activePrincipal"
          :label="t('staking.activePrincipal')"
          min-width="140"
        >
          <template #default="{ row }">{{ formatAmount(row.activePrincipal) }}</template>
        </el-table-column>
        <el-table-column prop="productDiff" :label="t('staking.productDiff')" min-width="120">
          <template #default="{ row }"><DiffAmount :value="row.productDiff" /></template>
        </el-table-column>
        <el-table-column prop="positionDiff" :label="t('staking.positionDiff')" min-width="120">
          <template #default="{ row }"><DiffAmount :value="row.positionDiff" /></template>
        </el-table-column>
        <el-table-column prop="lockDiff" :label="t('staking.lockDiff')" min-width="120">
          <template #default="{ row }"><DiffAmount :value="row.lockDiff" /></template>
        </el-table-column>
        <el-table-column prop="rewardDiff" :label="t('staking.rewardDiff')" min-width="120">
          <template #default="{ row }"><DiffAmount :value="row.rewardDiff" /></template>
        </el-table-column>
        <el-table-column prop="feeDiff" :label="t('staking.feeDiff')" min-width="120">
          <template #default="{ row }"><DiffAmount :value="row.feeDiff" /></template>
        </el-table-column>
        <el-table-column prop="status" :label="t('common.status')" width="110">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'">{{
              statusLabel(row.status)
            }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column
          prop="detail"
          :label="t('staking.reconciliationDetail')"
          min-width="220"
          show-overflow-tooltip
        />
        <el-table-column :label="t('common.actions')" fixed="right" width="90" align="center">
          <template #default="{ row }">
            <el-button link type="primary" @click="showDetail(row)">{{
              t('market.detail')
            }}</el-button>
          </template>
        </el-table-column>
      </el-table>
      <CursorPagination
        v-model:limit="pagination.limit"
        :total="pagination.total"
        :has-prev="pagination.hasPrev"
        :has-next="pagination.hasNext"
        @prev="prevAndLoad(loadList)"
        @next="nextAndLoad(loadList)"
        @limit-change="resetAndLoad(loadList)"
      />
    </el-card>

    <el-dialog v-model="detailVisible" :title="t('staking.reconciliationDetail')" width="860px">
      <el-descriptions v-if="detail" :column="2" border>
        <el-descriptions-item :label="t('staking.activePrincipal')">{{
          formatAmount(detail.activePrincipal)
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('staking.productStaked')">{{
          formatAmount(detail.productStaked)
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('staking.positionStaked')">{{
          formatAmount(detail.positionStaked)
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('staking.assetLocked')">{{
          formatAmount(detail.assetLocked)
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('staking.rewardLogAmount')">{{
          formatAmount(detail.rewardLogAmount)
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('staking.rewardPlatformAmount')">{{
          formatAmount(detail.rewardPlatformAmount)
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('staking.feeLogAmount')">{{
          formatAmount(detail.feeLogAmount)
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('staking.feePlatformAmount')">{{
          formatAmount(detail.feePlatformAmount)
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('staking.reconciliationDetail')" :span="2">{{
          detail.detail || '-'
        }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElText } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { usePagination } from '@/composables'
import { stakingService, type OptionGroup, type StakeReconciliation } from '@/services'
import TenantSelect from '@/components/TenantSelect.vue'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'
import { findFormOptionGroup, getOptionLabel, getOptionValueLabel } from '@/utils/options'

const { t } = useI18n()
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)
const loading = ref(false)
const rows = ref<StakeReconciliation[]>([])
const optionGroups = ref<OptionGroup[]>([])
const reconciliationStatusOptions = computed(() =>
  findFormOptionGroup(optionGroups.value, 'stakingReconciliationStatus'),
)
const detail = ref<StakeReconciliation | null>(null)
const detailVisible = computed({
  get: () => detail.value !== null,
  set: (value) => {
    if (!value) detail.value = null
  },
})
const query = reactive({
  tenantId: undefined as number | undefined,
  reconciliationDate: '' as string,
  coinSymbol: '',
  status: undefined as number | undefined,
})

const formatAmount = (value?: string) => {
  if (!value) return '0'
  const normalized = value.replace(/(\.\d*?[1-9])0+$|\.0+$/, '$1')
  return normalized || '0'
}
const DiffAmount = defineComponent({
  props: { value: { type: String, required: true } },
  setup(props) {
    return () =>
      h(ElText, { type: Number(props.value) === 0 ? 'success' : 'danger' }, () =>
        formatAmount(props.value),
      )
  },
})
const statusLabel = (status: number) =>
  getOptionValueLabel(optionGroups.value, 'stakingReconciliationStatus', status, t)

const loadList = async () => {
  loading.value = true
  try {
    const response = await stakingService.listReconciliations({
      ...query,
      reconciliationDate: query.reconciliationDate ? Number(query.reconciliationDate) : undefined,
      cursor: pagination.cursor,
      limit: pagination.limit,
    })
    rows.value = response.data || []
    updateFromResponse(response)
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : String(error))
  } finally {
    loading.value = false
  }
}
const loadOptions = async () => {
  optionGroups.value = (await stakingService.getOptions()).data || []
}
const resetQuery = () => {
  query.tenantId = undefined
  query.reconciliationDate = ''
  query.coinSymbol = ''
  query.status = undefined
  resetAndLoad(loadList)
}
const showDetail = (row: StakeReconciliation) => {
  detail.value = row
}

onMounted(() => Promise.all([loadList(), loadOptions()]))
</script>
