<template>
  <div class="module-page">
    <CrudQueryCard :model="query" @search="loadList" @reset="resetQuery">
      <el-form-item :label="t('staking.tenantId')">
        <TenantSelect v-model="query.tenantId" class="tenant-select-filter" />
      </el-form-item>
      <el-form-item :label="t('staking.orderNo')">
        <el-input v-model="query.orderNo" clearable />
      </el-form-item>
      <el-form-item :label="t('staking.redeemNo')">
        <el-input v-model="query.redeemNo" clearable />
      </el-form-item>
      <el-form-item :label="t('staking.redeemType')">
        <el-select v-model="query.redeemType" clearable>
          <el-option
            v-for="item in redeemTypeOptions"
            :key="item.value"
            :label="getOptionLabel(t, item.code, item.value)"
            :value="item.value"
          />
        </el-select>
      </el-form-item>
      <el-form-item :label="t('staking.redeemStatus')">
        <el-select v-model="query.redeemStatus" clearable>
          <el-option
            v-for="item in redeemStatusOptions"
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
          :label="t('staking.redeemNo')"
          prop="redeemNo"
          min-width="180"
          show-overflow-tooltip
        />
        <el-table-column
          :label="t('staking.orderNo')"
          prop="orderNo"
          min-width="180"
          show-overflow-tooltip
        />
        <el-table-column :label="t('staking.userId')" prop="userId" width="100" />
        <el-table-column
          prop="redeemAmount"
          :label="t('staking.redeemAmount')"
          min-width="120"
          show-overflow-tooltip
        />
        <el-table-column
          :label="t('staking.feeAmount')"
          prop="feeAmount"
          min-width="120"
          show-overflow-tooltip
        />
        <el-table-column :label="t('staking.redeemType')" prop="redeemType" width="120">
          <template #default="{ row }">
            {{ optionLabel('stakingRedeemType', row.redeemType) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('staking.redeemStatus')" prop="redeemStatus" width="110">
          <template #default="{ row }">
            {{ optionLabel('stakingRedeemStatus', row.redeemStatus) }}
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
        <el-descriptions-item :label="t('staking.redeemNo')">{{
          detailData.redeemNo
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
        <el-descriptions-item :label="t('staking.stakeAmount')">{{
          detailData.stakeAmount
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('staking.redeemAmount')">{{
          detailData.redeemAmount
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('staking.rewardAmount')">{{
          detailData.rewardAmount
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('staking.feeAmount')">{{
          detailData.feeAmount
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('staking.redeemType')">{{
          optionLabel('stakingRedeemType', detailData.redeemType)
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('staking.redeemStatus')">{{
          optionLabel('stakingRedeemStatus', detailData.redeemStatus)
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('staking.redeemTimes')">{{
          formatTime(detailData.redeemTimes)
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('common.remark')">{{
          detailData.remark || '-'
        }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { stakingService, type OptionGroup, type StakeRedeemLog } from '@/services'
import { useI18n } from 'vue-i18n'
import { usePagination } from '@/composables'
import TenantSelect from '@/components/TenantSelect.vue'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'
import { findFormOptionGroup, getOptionLabel, getOptionValueLabel } from '@/utils/options'

const { t } = useI18n()
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)
const formatTime = (value: number) => (value > 0 ? new Date(value).toLocaleString() : '-')

const loading = ref(false)
const rows = ref<StakeRedeemLog[]>([])
const detailVisible = ref(false)
const detailData = ref<StakeRedeemLog | null>(null)
const optionGroups = ref<OptionGroup[]>([])
const redeemTypeOptions = computed(() =>
  findFormOptionGroup(optionGroups.value, 'stakingRedeemType'),
)
const redeemStatusOptions = computed(() =>
  findFormOptionGroup(optionGroups.value, 'stakingRedeemStatus'),
)
const optionLabel = (key: string, value: number) =>
  getOptionValueLabel(optionGroups.value, key, value, t)
const query = reactive({
  tenantId: undefined as number | undefined,
  orderNo: '',
  userId: undefined as number | undefined,
  redeemNo: '',
  redeemType: undefined as number | undefined,
  redeemStatus: undefined as number | undefined,
  limit: 20,
})

const loadList = async () => {
  loading.value = true
  try {
    const res = await stakingService.listRedeemLogs({
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
  query.redeemNo = ''
  query.redeemType = undefined
  query.redeemStatus = undefined
  query.limit = 100
  loadList()
}

const showDetail = (row: StakeRedeemLog) => {
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
