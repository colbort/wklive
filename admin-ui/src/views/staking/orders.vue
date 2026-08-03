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
      <el-form-item :label="t('common.status')">
        <el-select v-model="query.status" clearable>
          <el-option
            v-for="item in orderStatusOptions"
            :key="item.value"
            :label="getOptionLabel(t, item.code, item.value)"
            :value="item.value"
          />
        </el-select>
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
      <el-form-item :label="t('staking.source')">
        <el-select v-model="query.source" clearable>
          <el-option
            v-for="item in sourceTypeOptions"
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
          prop="productName"
          :label="t('staking.productName')"
          min-width="180"
          show-overflow-tooltip
        />
        <el-table-column
          prop="stakeAmount"
          :label="t('staking.stakeAmount')"
          min-width="120"
          show-overflow-tooltip
        />
        <el-table-column
          prop="totalReward"
          :label="t('staking.totalRewardAmount')"
          min-width="120"
          show-overflow-tooltip
        />
        <el-table-column :label="t('common.status')" prop="status" width="120">
          <template #default="{ row }">
            {{ optionLabel('stakingOrderStatus', row.status) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('common.actions')" align="center" width="220" fixed="right">
          <template #default="{ row }">
            <el-button v-perm="'staking:order:detail'" link type="primary" @click="showDetail(row)">
              {{ t('market.detail') }}
            </el-button>
            <el-button
              v-perm="'staking:reward-log:manual'"
              link
              type="success"
              :disabled="!canReward(row)"
              @click="openRewardDialog(row)"
            >
              {{ t('staking.manualReward') }}
            </el-button>
            <el-button
              v-perm="'staking:redeem-log:manual'"
              link
              type="danger"
              :disabled="!canRedeem(row)"
              @click="openRedeemDialog(row)"
            >
              {{ t('staking.manualRedeem') }}
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

    <el-dialog v-model="rewardVisible" :title="t('staking.manualReward')" width="640px">
      <el-form label-width="100px">
        <el-form-item :label="t('staking.tenantId')">
          <span>{{ rewardForm.tenantId }}</span>
        </el-form-item>
        <el-form-item :label="t('staking.orderId')">
          <span>{{ rewardForm.orderId }}</span>
        </el-form-item>
        <el-form-item :label="t('staking.rewardAmount')">
          <el-input v-model="rewardForm.rewardAmount" />
        </el-form-item>
        <el-form-item :label="t('staking.rewardType')">
          <span>{{ optionLabel('stakingRewardType', rewardForm.rewardType) }}</span>
        </el-form-item>
        <el-form-item :label="t('common.remark')">
          <el-input v-model="rewardForm.remark" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="rewardVisible = false">
          {{ t('common.cancel') }}
        </el-button>
        <el-button
          v-perm="'staking:reward-log:manual'"
          type="primary"
          :loading="submitLoading"
          @click="submitReward"
        >
          {{ t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="redeemVisible" :title="t('staking.manualRedeem')" width="680px">
      <el-form label-width="100px">
        <el-form-item :label="t('staking.tenantId')">
          <span>{{ redeemForm.tenantId }}</span>
        </el-form-item>
        <el-form-item :label="t('staking.orderId')">
          <span>{{ redeemForm.orderId }}</span>
        </el-form-item>
        <el-form-item :label="t('staking.redeemType')">
          <span>{{ optionLabel('stakingRedeemType', redeemForm.redeemType) }}</span>
        </el-form-item>
        <el-form-item :label="t('staking.redeemAmount')">
          <el-input v-model="redeemForm.redeemAmount" />
        </el-form-item>
        <el-form-item :label="t('staking.rewardAmount')">
          <el-input v-model="redeemForm.rewardAmount" />
        </el-form-item>
        <el-form-item :label="t('staking.feeRate')">
          <el-input v-model="redeemForm.feeRate" />
        </el-form-item>
        <el-form-item :label="t('staking.feeAmount')">
          <el-input v-model="redeemForm.feeAmount" />
        </el-form-item>
        <el-form-item :label="t('common.remark')">
          <el-input v-model="redeemForm.remark" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="redeemVisible = false">
          {{ t('common.cancel') }}
        </el-button>
        <el-button
          v-perm="'staking:redeem-log:manual'"
          type="primary"
          :loading="submitLoading"
          @click="submitRedeem"
        >
          {{ t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="detailVisible" :title="t('market.detail')" width="760px">
      <el-descriptions v-if="detailData" :column="2" border>
        <el-descriptions-item :label="t('staking.orderNo')">{{
          detailData.orderNo
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('staking.requestNo')">{{
          detailData.requestNo || '-'
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('staking.tenantId')">{{
          detailData.tenantId
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('staking.userId')">{{
          detailData.userId
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('staking.productName')">{{
          detailData.productName
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('staking.productType')">{{
          optionLabel('stakingProductType', detailData.productType)
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('staking.coinSymbol')">{{
          detailData.coinSymbol
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('staking.stakeAmount')">{{
          detailData.stakeAmount
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('staking.totalRewardAmount')">{{
          detailData.totalReward
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('staking.interestMode')">{{
          optionLabel('interestMode', detailData.interestMode)
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('staking.rewardMode')">{{
          optionLabel('rewardMode', detailData.rewardMode)
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('staking.redeemType')">{{
          optionLabel('stakingRedeemType', detailData.redeemType)
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('staking.source')">{{
          optionLabel('stakingSourceType', detailData.source)
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('common.status')">{{
          optionLabel('stakingOrderStatus', detailData.status)
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('staking.operationNo')">{{
          detailData.activeOperationNo || '-'
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('staking.startTimes')">{{
          formatTime(detailData.startTimes)
        }}</el-descriptions-item>
        <el-descriptions-item :label="t('staking.endTimes')">{{
          formatTime(detailData.endTimes)
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
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  ManualRedeemReq,
  ManualRewardReq,
  stakingService,
  type OptionGroup,
  type StakeOrder,
} from '@/services'
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
const submitLoading = ref(false)
const rows = ref<StakeOrder[]>([])
const detailVisible = ref(false)
const detailData = ref<StakeOrder | null>(null)
const rewardVisible = ref(false)
const redeemVisible = ref(false)
const optionGroups = ref<OptionGroup[]>([])
const orderStatusOptions = computed(() =>
  findFormOptionGroup(optionGroups.value, 'stakingOrderStatus'),
)
const redeemTypeOptions = computed(() =>
  findFormOptionGroup(optionGroups.value, 'stakingRedeemType'),
)
const sourceTypeOptions = computed(() =>
  findFormOptionGroup(optionGroups.value, 'stakingSourceType'),
)
const optionLabel = (key: string, value: number) =>
  getOptionValueLabel(optionGroups.value, key, value, t)

const query = reactive({
  tenantId: undefined as number | undefined,
  orderNo: '',
  userId: undefined as number | undefined,
  productId: undefined as number | undefined,
  status: undefined as number | undefined,
  redeemType: undefined as number | undefined,
  source: undefined as number | undefined,
  limit: 20,
})
const rewardForm = reactive<ManualRewardReq>({
  tenantId: 0,
  orderId: 0,
  rewardAmount: '',
  rewardType: 4,
  requestNo: '',
  remark: '',
})
const redeemForm = reactive<ManualRedeemReq>({
  tenantId: 0,
  orderId: 0,
  redeemType: 4,
  redeemAmount: '',
  rewardAmount: '',
  feeRate: '',
  feeAmount: '',
  requestNo: '',
  remark: '',
})

const loadList = async () => {
  loading.value = true
  try {
    const res = await stakingService.listOrders({
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
  query.status = undefined
  query.redeemType = undefined
  query.source = undefined
  query.limit = 20
  loadList()
}

const showDetail = async (row: StakeOrder) => {
  detailData.value =
    (await stakingService.getOrder({ tenantId: row.tenantId, id: row.id })).data || row
  detailVisible.value = true
}

const openRewardDialog = (row: StakeOrder) => {
  Object.assign(rewardForm, {
    tenantId: row.tenantId || 0,
    orderId: row.id || 0,
    rewardAmount: '',
    rewardType: 4,
    requestNo: crypto.randomUUID(),
    remark: '',
  })
  rewardVisible.value = true
}

const canReward = (row: StakeOrder) => row.status === 1 || row.status === 2
const canRedeem = (row: StakeOrder) =>
  (row.status === 1 || row.status === 2) && (row.status === 2 || row.allowEarlyRedeem === 1)

const submitReward = async () => {
  await ElMessageBox.confirm(
    t('staking.manualRewardConfirm', {
      amount: rewardForm.rewardAmount || '0',
      tenant: rewardForm.tenantId,
      order: rewardForm.orderId,
    }),
    t('staking.manualReward'),
    { type: 'warning' },
  )
  submitLoading.value = true
  try {
    await stakingService.manualReward(rewardForm)
    ElMessage.success(t('staking.rewardSuccess'))
    rewardVisible.value = false
    loadList()
  } finally {
    submitLoading.value = false
  }
}

const openRedeemDialog = (row: StakeOrder) => {
  Object.assign(redeemForm, {
    tenantId: row.tenantId || 0,
    orderId: row.id || 0,
    redeemType: 4,
    redeemAmount: row.stakeAmount || '0',
    rewardAmount: row.pendingReward || '0',
    feeRate: '0',
    feeAmount: '0',
    requestNo: crypto.randomUUID(),
    remark: '',
  })
  redeemVisible.value = true
}

const submitRedeem = async () => {
  await ElMessageBox.confirm(
    t('staking.manualRedeemConfirm', {
      principal: redeemForm.redeemAmount,
      reward: redeemForm.rewardAmount,
      fee: redeemForm.feeAmount,
      tenant: redeemForm.tenantId,
      order: redeemForm.orderId,
    }),
    t('staking.manualRedeem'),
    { type: 'warning' },
  )
  submitLoading.value = true
  try {
    await stakingService.manualRedeem(redeemForm)
    ElMessage.success(t('staking.redeemSuccess'))
    redeemVisible.value = false
    loadList()
  } finally {
    submitLoading.value = false
  }
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
