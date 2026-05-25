<template>
  <div class="module-page">
    <div class="page-header">
      <h2>{{ t('asset.flows') }}</h2>
      <div class="header-actions">
        <el-button @click="loadList">
          {{ t('common.refresh') }}
        </el-button>
      </div>
    </div>

    <el-card shadow="never" class="query-card">
      <el-form :model="query" inline label-width="88px">
        <el-form-item :label="t('asset.tenantId')">
          <el-input-number v-model="query.tenantId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('asset.userId')">
          <el-input-number v-model="query.userId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('asset.walletType')">
          <el-select v-model="query.walletType" clearable style="width: 160px">
            <el-option
              v-for="item in walletTypeOptions"
              :key="item.value"
              :label="getOptionLabel(t, item.code, item.value)"
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
        <el-form-item>
          <el-button type="primary" @click="loadList">
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
            {{ formatOptionValue('walletType', row.walletType) }}
          </template>
        </el-table-column>
        <el-table-column
          prop="coin"
          :label="t('asset.coin')"
          min-width="100"
          show-overflow-tooltip
        />
        <el-table-column
          prop="changeAmount"
          :label="t('asset.changeAmount')"
          min-width="140"
          show-overflow-tooltip
        />
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
        <el-table-column :label="t('common.actions')" width="120" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="showDetail(row)">
              {{ t('asset.detail') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
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
          {{ formatOptionValue('walletType', detailData.walletType) }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('asset.coin')">
          {{ detailData.coin }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('asset.changeAmount')">
          {{ detailData.changeAmount }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('asset.opType')">
          {{ detailData.opType }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('asset.bizType')">
          {{ detailData.bizType }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('asset.sceneType')">
          {{ detailData.sceneType }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('asset.bizId')">
          {{ detailData.bizId }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('asset.bizNo')">
          {{ detailData.bizNo || '-' }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('asset.beforeTotalAmount')">
          {{ detailData.beforeTotalAmount }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('asset.afterTotalAmount')">
          {{ detailData.afterTotalAmount }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('asset.beforeAvailableAmount')">
          {{ detailData.beforeAvailableAmount }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('asset.afterAvailableAmount')">
          {{ detailData.afterAvailableAmount }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('asset.beforeFrozenAmount')">
          {{ detailData.beforeFrozenAmount }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('asset.afterFrozenAmount')">
          {{ detailData.afterFrozenAmount }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('asset.beforeLockedAmount')">
          {{ detailData.beforeLockedAmount }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('asset.afterLockedAmount')">
          {{ detailData.afterLockedAmount }}
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
import { assetService, type AssetFlow, type OptionGroup } from '@/services'
import { formatDate } from '@/utils'
import { findOptionGroup, getOptionLabel, getOptionValueLabel } from '@/utils/options'

const { t } = useI18n()

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
  limit: 100,
})

const walletTypeOptions = computed(() => findOptionGroup(optionGroups.value, 'walletType'))
const detailTitle = computed(() => `${t('asset.flows')}${t('asset.detail')}`)

function pickList(res: any) {
  return res?.data || res?.list || []
}

function formatOptionValue(key: string, value: number | string | undefined) {
  return getOptionValueLabel(optionGroups.value, key, value, t)
}

async function loadOptions() {
  const res = await assetService.getOptions()
  optionGroups.value = res.data || []
}

async function loadList() {
  loading.value = true
  try {
    rows.value = pickList(await assetService.getFlows(query))
  } finally {
    loading.value = false
  }
}

function resetQuery() {
  query.tenantId = undefined
  query.userId = undefined
  query.walletType = undefined
  query.coin = ''
  query.bizNo = ''
  query.limit = 100
  loadList()
}

function showDetail(row: AssetFlow) {
  detailData.value = row
  detailVisible.value = true
}

onMounted(loadList)
onMounted(loadOptions)
</script>

<style scoped>
.query-card :deep(.el-form-item) {
  margin-bottom: 12px;
}
</style>
