<template>
  <div class="module-page">
    <div class="page-header">
      <h2>{{ t('staking.orders') }}</h2>
      <div class="header-actions">
        <el-button @click="loadOrders">
          {{ t('common.refresh') }}
        </el-button>
      </div>
    </div>

    <el-card shadow="never" class="query-card">
      <el-form :model="query" inline label-width="90px">
        <el-form-item :label="t('staking.tenantId')">
          <el-input-number v-model="query.tenantId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('staking.orderNo')">
          <el-input v-model="query.orderNo" clearable />
        </el-form-item>
        <el-form-item :label="t('staking.userId')">
          <el-input-number v-model="query.userId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadOrders">
            {{ t('common.search') }}
          </el-button>
          <el-button @click="resetQuery">
            {{ t('common.reset') }}
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

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
        <el-table-column :label="t('common.status')" prop="status" width="100" />
        <el-table-column :label="t('common.actions')" width="220" fixed="right">
          <template #default="{ row }">
            <el-button
              v-perm="'staking:order:detail'"
              link
              type="primary"
              @click="showDetail(row)"
            >
              {{ t('itick.detail') }}
            </el-button>
            <el-button
              v-perm="'staking:reward-log:manual'"
              link
              type="success"
              @click="openRewardDialog(row)"
            >
              {{ t('staking.manualReward') }}
            </el-button>
            <el-button
              v-perm="'staking:redeem-log:manual'"
              link
              type="danger"
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
          <TenantSelect v-model="rewardForm.tenantId" include-system />
        </el-form-item>
        <el-form-item :label="t('staking.orderId')">
          <el-input-number v-model="rewardForm.orderId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('staking.rewardAmount')">
          <el-input v-model="rewardForm.rewardAmount" />
        </el-form-item>
        <el-form-item :label="t('staking.rewardType')">
          <el-input-number v-model="rewardForm.rewardType" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('staking.operatorUid')">
          <el-input-number v-model="rewardForm.operatorUid" :min="0" :precision="0" />
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
          <TenantSelect v-model="redeemForm.tenantId" include-system />
        </el-form-item>
        <el-form-item :label="t('staking.orderId')">
          <el-input-number v-model="redeemForm.orderId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('staking.redeemType')">
          <el-input-number v-model="redeemForm.redeemType" :min="0" :precision="0" />
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
        <el-form-item :label="t('staking.operatorUid')">
          <el-input-number v-model="redeemForm.operatorUid" :min="0" :precision="0" />
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

    <el-dialog v-model="detailVisible" :title="t('itick.detail')" width="760px">
      <pre class="detail-pre">{{ JSON.stringify(detailData, null, 2) }}</pre>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import {
  AdminManualRedeemReq,
  AdminManualRewardReq,
  stakingService,
  type StakeOrder,
} from '@/services'
import { useI18n } from 'vue-i18n'
import { usePagination } from '@/composables'
import TenantSelect from '@/components/TenantSelect.vue'

const { t } = useI18n()
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)

const loading = ref(false)
const submitLoading = ref(false)
const rows = ref<StakeOrder[]>([])
const detailVisible = ref(false)
const detailData = ref<StakeOrder | null>(null)
const rewardVisible = ref(false)
const redeemVisible = ref(false)

const query = reactive({
  tenantId: undefined as number | undefined,
  orderNo: '',
  userId: undefined as number | undefined,
  productId: undefined as number | undefined,
  limit: 20,
})
const rewardForm = reactive<AdminManualRewardReq>({
  tenantId: 0,
  orderId: 0,
  rewardAmount: '',
  rewardType: 1,
  operatorUid: 0,
  remark: '',
})
const redeemForm = reactive<AdminManualRedeemReq>({
  tenantId: 0,
  orderId: 0,
  redeemType: 1,
  redeemAmount: '',
  rewardAmount: '',
  feeRate: '',
  feeAmount: '',
  operatorUid: 0,
  remark: '',
})

const loadOrders = async () => {
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

const resetQuery = () => {
  query.tenantId = undefined
  query.orderNo = ''
  query.userId = undefined
  query.productId = undefined
  query.limit = 100
  loadOrders()
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
    rewardType: 1,
    operatorUid: 0,
    remark: '',
  })
  rewardVisible.value = true
}

const submitReward = async () => {
  submitLoading.value = true
  try {
    await stakingService.manualReward(rewardForm)
    ElMessage.success(t('staking.rewardSuccess'))
    rewardVisible.value = false
    loadOrders()
  } finally {
    submitLoading.value = false
  }
}

const openRedeemDialog = (row: StakeOrder) => {
  Object.assign(redeemForm, {
    tenantId: row.tenantId || 0,
    orderId: row.id || 0,
    redeemType: 1,
    redeemAmount: '',
    rewardAmount: '',
    feeRate: '',
    feeAmount: '',
    operatorUid: 0,
    remark: '',
  })
  redeemVisible.value = true
}

const submitRedeem = async () => {
  submitLoading.value = true
  try {
    await stakingService.manualRedeem(redeemForm)
    ElMessage.success(t('staking.redeemSuccess'))
    redeemVisible.value = false
    loadOrders()
  } finally {
    submitLoading.value = false
  }
}

function handleLimitChange() {
  resetAndLoad(loadOrders)
}

function handlePrevPage() {
  prevAndLoad(loadOrders)
}

function handleNextPage() {
  nextAndLoad(loadOrders)
}

onMounted(loadOrders)
</script>

<style scoped></style>
