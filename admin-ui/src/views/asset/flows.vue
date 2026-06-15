<template>
  <div class="module-page">
    <CrudQueryCard
      :model="query"
      @search="loadList"
      @reset="resetQuery"
    >
      <el-form-item :label="t('asset.tenantId')">
        <TenantSelect v-model="query.tenantId" class="tenant-select-filter" />
      </el-form-item>
      <el-form-item :label="t('asset.userId')">
        <el-input-number v-model="query.userId" :min="0" :precision="0" />
      </el-form-item>
      <el-form-item :label="t('asset.walletType')">
        <el-select v-model="query.walletType" clearable style="width: 160px">
          <el-option
            v-for="item in walletTypeOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
      </el-form-item>
      <el-form-item :label="t('asset.coin')">
        <el-input v-model="query.coin" clearable />
      </el-form-item>
      <el-form-item :label="t('asset.bizNo')">
        <el-input v-model="query.bizNo" clearable />
      </el-form-item>
    </CrudQueryCard>

    <el-card shadow="never" class="table-card">
      <!-- prettier-ignore -->
      <el-table
        v-loading="loading"
        :data="rows"
        stripe
        height="100%"
      >
        <el-table-column
          prop="flowNo"
          :label="t('asset.flowNo')"
          min-width="180"
          show-overflow-tooltip
        />
        <el-table-column
          prop="tenantId"
          :label="t('asset.tenantId')"
          min-width="100"
          show-overflow-tooltip
        />
        <el-table-column
          prop="userId"
          :label="t('asset.userId')"
          min-width="100"
          show-overflow-tooltip
        />
        <el-table-column prop="walletType" :label="t('asset.walletType')" min-width="120">
          <template #default="{ row }">
            {{ optionLabel('walletType', row.walletType) }}
          </template>
        </el-table-column>
        <el-table-column
          prop="coin"
          :label="t('asset.coin')"
          min-width="100"
          show-overflow-tooltip
        />
        <el-table-column :label="t('asset.changeAmount')" min-width="140" show-overflow-tooltip>
          <template #default="{ row }">
            {{ formatCentAmount(row.changeAmount) }}
          </template>
        </el-table-column>
        <el-table-column prop="opType" :label="t('asset.opType')" min-width="130">
          <template #default="{ row }">
            <el-tag size="small" :type="opTypeTagType(row.opType)" effect="light">
              {{ optionLabelWithFallback('opType', row.opType) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          prop="bizNo"
          :label="t('asset.bizNo')"
          min-width="180"
          show-overflow-tooltip
        />
        <el-table-column prop="createTimes" :label="t('common.createTimes')" min-width="180">
          <template #default="{ row }">
            {{ formatDate(row.createTimes) }}
          </template>
        </el-table-column>
        <el-table-column
          :label="t('common.actions')"
          align="center"
          width="120"
          fixed="right"
        >
          <template #default="{ row }">
            <el-button link type="primary" @click="showDetail(row)">
              {{ t('asset.detail') }}
            </el-button>
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

    <el-drawer v-model="detailVisible" :title="detailTitle" size="760px">
      <el-descriptions v-if="detailData" :column="2" border>
        <el-descriptions-item :label="t('common.id')">
          {{ detailData.id }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('asset.flowNo')">
          {{ detailData.flowNo }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('asset.tenantId')">
          {{ detailData.tenantId }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('asset.userId')">
          {{ detailData.userId }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('asset.walletType')">
          {{ optionLabel('walletType', detailData.walletType) }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('asset.coin')">
          {{ detailData.coin }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('asset.changeAmount')">
          {{ formatCentAmount(detailData.changeAmount) }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('asset.opType')">
          {{ optionLabelWithFallback('opType', detailData.opType) }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('asset.bizType')">
          {{ optionLabel('bizType', detailData.bizType) }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('asset.sceneType')">
          {{ optionLabel('assetSceneType', detailData.sceneType) }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('asset.bizId')">
          {{ detailData.bizId }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('asset.bizNo')">
          {{ detailData.bizNo || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('asset.beforeTotalAmount')">
          {{ formatCentAmount(detailData.beforeTotalAmount) }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('asset.afterTotalAmount')">
          {{ formatCentAmount(detailData.afterTotalAmount) }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('asset.beforeAvailableAmount')">
          {{ formatCentAmount(detailData.beforeAvailableAmount) }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('asset.afterAvailableAmount')">
          {{ formatCentAmount(detailData.afterAvailableAmount) }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('asset.beforeFrozenAmount')">
          {{ formatCentAmount(detailData.beforeFrozenAmount) }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('asset.afterFrozenAmount')">
          {{ formatCentAmount(detailData.afterFrozenAmount) }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('asset.beforeLockedAmount')">
          {{ formatCentAmount(detailData.beforeLockedAmount) }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('asset.afterLockedAmount')">
          {{ formatCentAmount(detailData.afterLockedAmount) }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('asset.balanceSnapshotVersion')">
          {{ detailData.balanceSnapshotVersion }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('asset.changeType')">
          {{ detailData.changeType || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('common.createTimes')">
          {{ formatDate(detailData.createTimes) }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('common.updateTimes')">
          {{ formatDate(detailData.updateTimes) }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('common.remark')" :span="2">
          {{ detailData.remark || '-' }}
        </el-descriptions-item>
      </el-descriptions>
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useOptions, usePagination } from '@/composables'
import { assetService, type AssetFlow, type OptionGroup, type OptionItem } from '@/services'
import { formatCentAmount, formatDate } from '@/utils'
import { findOptionGroup, getOptionLabel } from '@/utils/options'
import TenantSelect from '@/components/TenantSelect.vue'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'

const { t } = useI18n()

const fallbackOptions: Record<string, OptionItem[]> = {
  opType: [
    { value: 0, code: 'ASSET_OP_TYPE_UNKNOWN' },
    { value: 1, code: 'ASSET_OP_TYPE_ADD' },
    { value: 2, code: 'ASSET_OP_TYPE_SUB' },
    { value: 3, code: 'ASSET_OP_TYPE_FREEZE' },
    { value: 4, code: 'ASSET_OP_TYPE_UNFREEZE' },
    { value: 5, code: 'ASSET_OP_TYPE_LOCK' },
    { value: 6, code: 'ASSET_OP_TYPE_UNLOCK' },
    { value: 7, code: 'ASSET_OP_TYPE_FREEZE_DEDUCT' },
    { value: 8, code: 'ASSET_OP_TYPE_LOCK_DEDUCT' },
    { value: 9, code: 'ASSET_OP_TYPE_TRANSFER_IN' },
    { value: 10, code: 'ASSET_OP_TYPE_TRANSFER_OUT' },
  ],
}

const loading = ref(false)
const rows = ref<AssetFlow[]>([])
const optionGroups = ref<OptionGroup[]>([])
const detailVisible = ref(false)
const detailData = ref<AssetFlow | null>(null)

const query = reactive({
  tenantId: undefined as number | undefined,
  userId: undefined as number | undefined,
  walletType: undefined as number | undefined,
  coin: '',
  bizNo: '',
})

const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)
const { optionItems, optionLabel } = useOptions(optionGroups)

const walletTypeOptions = optionItems('walletType')
const detailTitle = computed(() => `${t('asset.flows')}${t('asset.detail')}`)

function optionLabelWithFallback(key: string, value?: number | string) {
  if (value === undefined || value === null || value === '') return '-'
  const option =
    findOptionGroup(optionGroups.value, key).find((item) => String(item.value) === String(value)) ||
    fallbackOptions[key]?.find((item) => String(item.value) === String(value))
  return option ? getOptionLabel(t, option.code, option.value) : String(value)
}

async function loadOptions() {
  const res = await assetService.getOptions()
  optionGroups.value = res.data || []
}

async function fetchList() {
  loading.value = true
  try {
    const res = await assetService.getFlows({
      ...query,
      tenantId: query.tenantId || undefined,
      userId: query.userId || undefined,
      walletType: query.walletType || undefined,
      coin: query.coin || undefined,
      bizNo: query.bizNo || undefined,
      cursor: pagination.cursor,
      limit: pagination.limit,
    })
    rows.value = res.data || []
    updateFromResponse(res)
  } finally {
    loading.value = false
  }
}

function loadList() {
  resetAndLoad(fetchList)
}

function resetQuery() {
  query.tenantId = undefined
  query.userId = undefined
  query.walletType = undefined
  query.coin = ''
  query.bizNo = ''
  resetAndLoad(fetchList)
}

function handleLimitChange() {
  resetAndLoad(fetchList)
}

function handlePrevPage() {
  prevAndLoad(fetchList)
}

function handleNextPage() {
  nextAndLoad(fetchList)
}

function showDetail(row: AssetFlow) {
  detailData.value = row
  detailVisible.value = true
}

function opTypeTagType(opType?: number | string) {
  switch (Number(opType)) {
    case 1:
    case 9:
      return 'success'
    case 2:
    case 10:
      return 'danger'
    case 3:
    case 5:
      return 'warning'
    case 4:
    case 6:
    case 7:
    case 8:
      return 'info'
    default:
      return ''
  }
}

onMounted(fetchList)
onMounted(loadOptions)
</script>
