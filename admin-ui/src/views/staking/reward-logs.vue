<template>
  <div class="module-page">
    <CrudQueryCard :model="query" @search="loadList" @reset="resetQuery">
      <el-form-item :label="t('staking.tenantId')">
        <TenantSelect v-model="query.tenantId" class="tenant-select-filter" />
      </el-form-item>
      <el-form-item :label="t('staking.orderNo')">
        <el-input v-model="query.orderNo" clearable />
      </el-form-item>
      <el-form-item :label="t('staking.userId')">
        <UserSelect v-model="query.userId" :tenant-id="query.tenantId || undefined" />
      </el-form-item>
      <el-form-item :label="t('staking.rewardType')">
        <el-select v-model="query.rewardType" clearable>
          <el-option
            v-for="item in rewardTypeOptions"
            :key="item.value"
            :label="getOptionLabel(t, item.code, item.value)"
            :value="item.value"
          />
        </el-select>
      </el-form-item>
      <el-form-item :label="t('staking.rewardStatus')">
        <el-select v-model="query.rewardStatus" clearable>
          <el-option
            v-for="item in rewardStatusOptions"
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
          :label="t('staking.orderNo')"
          prop="orderNo"
          min-width="180"
          show-overflow-tooltip
        />
        <el-table-column :label="t('staking.userId')" prop="userId" width="100" />
        <el-table-column
          prop="rewardAmount"
          :label="t('staking.rewardAmount')"
          min-width="120"
          show-overflow-tooltip
        />
        <el-table-column :label="t('staking.rewardType')" prop="rewardType" width="120">
          <template #default="{ row }">
            {{ optionLabel('stakingRewardType', row.rewardType) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('staking.rewardStatus')" prop="rewardStatus" width="110">
          <template #default="{ row }">
            {{ optionLabel('stakingRewardStatus', row.rewardStatus) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('common.actions')" align="center" width="100" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="showDetail(row)">
              {{ t('market.detail') }}
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

    <el-dialog v-model="detailVisible" :title="t('market.detail')" width="760px">
      <el-descriptions v-if="detailData" :column="2" border>
        <el-descriptions-item :label="t('staking.operationNo')">{{
          detailData.operationNo || '-'
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('staking.orderNo')">{{
          detailData.orderNo
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('staking.tenantId')">{{
          detailData.tenantId
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('staking.userId')">{{
          detailData.userId
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('staking.rewardAmount')"
          >{{ detailData.rewardAmount }} {{ detailData.rewardCoinSymbol }}</el-descriptions-item
        >
        <el-descriptions-item :label="t('staking.rewardType')">{{
          optionLabel('stakingRewardType', detailData.rewardType)
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('staking.beforeReward')">{{
          detailData.beforeReward
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('staking.afterReward')">{{
          detailData.afterReward
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('staking.rewardStatus')">{{
          optionLabel('stakingRewardStatus', detailData.rewardStatus)
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('staking.rewardTimes')">{{
          formatTime(detailData.rewardTimes)
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('common.remark')" :span="2">{{
          detailData.remark || '-'
        }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { stakingService, type OptionGroup, type StakeRewardLog } from '@/services'
import { useI18n } from 'vue-i18n'
import { usePagination } from '@/composables'
import TenantSelect from '@/components/TenantSelect.vue'
import UserSelect from '@/components/UserSelect.vue'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'
import { findFormOptionGroup, getOptionLabel, getOptionValueLabel } from '@/utils/options'

const { t } = useI18n()
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)
const formatTime = (value: number) => (value > 0 ? new Date(value).toLocaleString() : '-')

const loading = ref(false)
const rows = ref<StakeRewardLog[]>([])
const detailVisible = ref(false)
const detailData = ref<StakeRewardLog | null>(null)
const optionGroups = ref<OptionGroup[]>([])
const rewardTypeOptions = computed(() =>
  findFormOptionGroup(optionGroups.value, 'stakingRewardType'),
)
const rewardStatusOptions = computed(() =>
  findFormOptionGroup(optionGroups.value, 'stakingRewardStatus'),
)
const optionLabel = (key: string, value: number) =>
  getOptionValueLabel(optionGroups.value, key, value, t)
const query = reactive({
  tenantId: undefined as number | undefined,
  orderNo: '',
  userId: undefined as number | undefined,
  productId: undefined as number | undefined,
  rewardType: undefined as number | undefined,
  rewardStatus: undefined as number | undefined,
  limit: 20,
})

const loadList = async () => {
  loading.value = true
  try {
    const res = await stakingService.listRewardLogs({
      ...query,
      cursor: pagination.cursor,
      limit: pagination.limit,
    })
    rows.value = res?.data || []
    updateFromResponse(res)
  } finally {
    loading.value = false
  }
}

const loadOptions = async () => {
  optionGroups.value = (await stakingService.getOptions()).data || []
}

const resetQuery = () => {
  query.tenantId = undefined
  query.orderNo = ''
  query.userId = undefined
  query.productId = undefined
  query.rewardType = undefined
  query.rewardStatus = undefined
  query.limit = 100
  loadList()
}

const showDetail = (row: StakeRewardLog) => {
  detailData.value = row
  detailVisible.value = true
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

onMounted(() => Promise.all([loadList(), loadOptions()]))
</script>
