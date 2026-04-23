<template>
  <div class="module-page">
    <div class="page-header">
      <h2>{{ t('staking.rewardLogs') }}</h2>
      <div class="header-actions">
        <el-button @click="loadRows">
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
          <el-input-number v-model="query.uid" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadRows">
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
        <el-table-column :label="t('staking.userId')" prop="uid" width="100" />
        <el-table-column
          prop="rewardAmount"
          :label="t('staking.rewardAmount')"
          min-width="120"
          show-overflow-tooltip
        />
        <el-table-column :label="t('staking.rewardType')" prop="rewardType" width="100" />
        <el-table-column :label="t('staking.rewardStatus')" prop="rewardStatus" width="100" />
        <el-table-column :label="t('common.actions')" width="100" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="showDetail(row)">
              {{ t('itick.detail') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="detailVisible" :title="t('itick.detail')" width="760px">
      <pre class="detail-pre">{{ JSON.stringify(detailData, null, 2) }}</pre>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { stakingService, type StakeRewardLog } from '@/services'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const loading = ref(false)
const rows = ref<StakeRewardLog[]>([])
const detailVisible = ref(false)
const detailData = ref<StakeRewardLog | null>(null)
const query = reactive({
  tenantId: undefined as number | undefined,
  orderNo: '',
  uid: undefined as number | undefined,
  productId: undefined as number | undefined,
  limit: 100,
})

const pickList = (res: any) => res?.data || res?.list || []

const loadRows = async () => {
  loading.value = true
  try {
    rows.value = pickList(await stakingService.listRewardLogs(query))
  } finally {
    loading.value = false
  }
}

const resetQuery = () => {
  query.tenantId = undefined
  query.orderNo = ''
  query.uid = undefined
  query.productId = undefined
  query.limit = 100
  loadRows()
}

const showDetail = (row: StakeRewardLog) => {
  detailData.value = row
  detailVisible.value = true
}

onMounted(loadRows)
</script>

<style scoped></style>
