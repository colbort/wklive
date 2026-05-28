<template>
  <div class="module-page">
    <div class="page-header">
      <h2>{{ t('asset.userAssets') }}</h2>
      <div class="header-actions">
        <el-button @click="loadList">
          {{ t('common.refresh') }}
        </el-button>
        <el-button type="primary" @click="openChangeDialog('add')">
          {{ t('asset.addAsset') }}
        </el-button>
        <el-button type="warning" @click="openChangeDialog('sub')">
          {{ t('asset.subAsset') }}
        </el-button>
        <el-button type="primary" @click="openChangeDialog('freeze')">
          {{ t('asset.freezeAsset') }}
        </el-button>
        <el-button type="primary" plain @click="openChangeDialog('lock')">
          {{ t('asset.lockAsset') }}
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
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('asset.coin')">
          <el-input v-model="query.coin" clearable />
        </el-form-item>
        <el-form-item :label="t('common.status')">
          <el-select v-model="query.status" clearable style="width: 160px">
            <el-option
              v-for="item in assetStatusOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
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
        <el-table-column
          prop="totalAmount"
          :label="t('asset.totalAmount')"
          min-width="140"
          show-overflow-tooltip
        />
        <el-table-column
          prop="availableAmount"
          :label="t('asset.availableAmount')"
          min-width="140"
          show-overflow-tooltip
        />
        <el-table-column
          prop="frozenAmount"
          :label="t('asset.frozenAmount')"
          min-width="140"
          show-overflow-tooltip
        />
        <el-table-column
          prop="lockedAmount"
          :label="t('asset.lockedAmount')"
          min-width="140"
          show-overflow-tooltip
        />
        <el-table-column :label="t('common.actions')" width="120" fixed="right">
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

    <el-dialog v-model="changeVisible" :title="changeTitle" width="680px">
      <el-form label-width="100px">
        <el-form-item :label="t('asset.userId')">
          <el-input-number v-model="changeForm.userId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('asset.walletType')">
          <el-select v-model="changeForm.walletType" style="width: 100%">
            <el-option
              v-for="item in walletTypeOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('asset.coin')">
          <el-input v-model="changeForm.coin" />
        </el-form-item>
        <el-form-item :label="t('asset.amount')">
          <el-input v-model="changeForm.amount" />
        </el-form-item>
        <el-form-item :label="t('asset.bizNo')">
          <el-input v-model="changeForm.bizNo" />
        </el-form-item>
        <el-form-item :label="t('asset.operatorId')">
          <el-input-number v-model="changeForm.operatorId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('common.remark')">
          <el-input v-model="changeForm.remark" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="changeVisible = false">
          {{ t('common.cancel') }}
        </el-button>
        <el-button type="primary" :loading="submitLoading" @click="submitChange">
          {{ t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>

    <el-drawer v-model="detailVisible" :title="detailTitle" size="720px">
      <el-descriptions v-if="detailData" :column="2" border>
        <el-descriptions-item :label="t('common.id')">
          {{ detailData.id }}
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
        <el-descriptions-item :label="t('common.status')">
          {{ optionLabel('assetStatus', detailData.status) }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('asset.totalAmount')">
          {{ detailData.totalAmount }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('asset.availableAmount')">
          {{ detailData.availableAmount }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('asset.frozenAmount')">
          {{ detailData.frozenAmount }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('asset.lockedAmount')">
          {{ detailData.lockedAmount }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('asset.version')">
          {{ detailData.version }}
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
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { useOptions, usePagination } from '@/composables'
import { assetService, type AssetUserAsset, type OptionGroup } from '@/services'
import { formatDate } from '@/utils'

const { t } = useI18n()
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)

const loading = ref(false)
const submitLoading = ref(false)
const rows = ref<AssetUserAsset[]>([])
const optionGroups = ref<OptionGroup[]>([])
const detailVisible = ref(false)
const detailData = ref<AssetUserAsset | null>(null)
const changeVisible = ref(false)
const changeMode = ref<'add' | 'sub' | 'freeze' | 'lock'>('add')
const { optionItems, optionLabel } = useOptions(optionGroups)

const query = reactive({
  tenantId: undefined as number | undefined,
  userId: undefined as number | undefined,
  walletType: undefined as number | undefined,
  coin: '',
  status: undefined as number | undefined,
  limit: 20,
})

const changeForm = reactive({
  tenantId: 0,
  userId: 0,
  walletType: 1,
  coin: '',
  amount: '',
  bizNo: '',
  operatorId: 0,
  remark: '',
})

const walletTypeOptions = optionItems('walletType')
const assetStatusOptions = optionItems('assetStatus')
const detailTitle = computed(() => `${t('asset.userAssets')}${t('asset.detail')}`)
const changeTitle = computed(
  () =>
    ({
      add: t('asset.addAsset'),
      sub: t('asset.subAsset'),
      freeze: t('asset.freezeAsset'),
      lock: t('asset.lockAsset'),
    })[changeMode.value],
)

async function loadOptions() {
  const res = await assetService.getOptions()
  optionGroups.value = res.data || []
}

async function loadList() {
  loading.value = true
  try {
    const resp = await assetService.getUserAssets({
      ...query,
      cursor: pagination.cursor,
      limit: pagination.limit,
    })
    if (resp.code !== 200) {
      ElMessage.error(resp.msg || t('common.loadFailed'))
      return
    }
    rows.value = resp?.data || []
    updateFromResponse(resp)
  } finally {
    loading.value = false
  }
}

function resetQuery() {
  query.tenantId = undefined
  query.userId = undefined
  query.walletType = undefined
  query.coin = ''
  query.status = undefined
  query.limit = 100
  loadList()
}

async function showDetail(row: AssetUserAsset) {
  const res = await assetService.getUserAssetDetail({
    tenantId: Number(row.tenantId),
    userId: Number(row.userId),
    walletType: Number(row.walletType),
    coin: String(row.coin || ''),
  })
  detailData.value = res.data || row
  detailVisible.value = true
}

function openChangeDialog(mode: typeof changeMode.value) {
  changeMode.value = mode
  changeForm.tenantId = 0
  changeForm.userId = 0
  changeForm.walletType = 1
  changeForm.coin = ''
  changeForm.amount = ''
  changeForm.bizNo = ''
  changeForm.operatorId = 0
  changeForm.remark = ''
  changeVisible.value = true
}

async function submitChange() {
  submitLoading.value = true
  try {
    if (changeMode.value === 'add') await assetService.addAsset(changeForm)
    if (changeMode.value === 'sub') await assetService.subAsset(changeForm)
    if (changeMode.value === 'freeze') await assetService.freezeAsset(changeForm)
    if (changeMode.value === 'lock') await assetService.lockAsset(changeForm)
    ElMessage.success(t('common.success'))
    changeVisible.value = false
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

onMounted(loadList)
onMounted(loadOptions)
</script>

<style scoped>
.query-card :deep(.el-form-item) {
  margin-bottom: 12px;
}
</style>
