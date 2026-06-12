<template>
  <div class="module-page">
    <div class="page-header">
      <h2>{{ t('asset.freezes') }}</h2>
      <div class="header-actions">
        <el-button @click="loadList">
          {{ t('common.refresh') }}
        </el-button>
        <el-button
          v-perm="'asset:freeze:unfreeze'"
          type="success"
          plain
          @click="openChangeDialog"
        >
          {{ t('asset.unfreezeAsset') }}
        </el-button>
      </div>
    </div>

    <CrudQueryCard :model="query" label-width="88px" :show-actions="false">
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
      <el-form-item :label="t('asset.bizType')">
        <el-select v-model="query.bizType" clearable style="width: 160px">
          <el-option
            v-for="item in bizTypeOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
      </el-form-item>
      <el-form-item :label="t('common.status')">
        <el-select v-model="query.status" clearable style="width: 160px">
          <el-option
            v-for="item in freezeStatusOptions"
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
    </CrudQueryCard>

    <el-card shadow="never" class="table-card">
      <el-table v-loading="loading" :data="rows" stripe>
        <el-table-column
          prop="freezeNo"
          :label="t('asset.freezeNo')"
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
        <el-table-column
          prop="amount"
          :label="t('asset.freezeAmount')"
          min-width="140"
          show-overflow-tooltip
        />
        <el-table-column
          prop="usedAmount"
          :label="t('asset.usedAmount')"
          min-width="140"
          show-overflow-tooltip
        />
        <el-table-column
          prop="unfreezeAmount"
          :label="t('asset.unfreezeAmount')"
          min-width="140"
          show-overflow-tooltip
        />
        <el-table-column
          prop="remainAmount"
          :label="t('asset.remainAmount')"
          min-width="140"
          show-overflow-tooltip
        />
        <el-table-column prop="bizType" :label="t('asset.bizType')" min-width="120">
          <template #default="{ row }">
            {{ optionLabel('bizType', row.bizType) }}
          </template>
        </el-table-column>
        <el-table-column prop="status" :label="t('common.status')" min-width="110">
          <template #default="{ row }">
            {{ optionLabel('freezeStatus', row.status) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('common.actions')" width="180" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="showDetail(row)">
              {{ t('asset.detail') }}
            </el-button>
            <el-button
              v-perm="'asset:freeze:unfreeze'"
              link
              type="success"
              @click="prefillUnfreeze(row)"
            >
              {{ t('asset.unfreezeAsset') }}
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
        <el-form-item :label="t('asset.freezeNo')">
          <el-input v-model="changeForm.freezeNo" />
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
        <el-button
          v-perm="'asset:freeze:unfreeze'"
          type="primary"
          :loading="submitLoading"
          @click="submitChange"
        >
          {{ t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>

    <el-drawer v-model="detailVisible" :title="detailTitle" size="760px">
      <el-descriptions v-if="detailData" :column="2" border>
        <el-descriptions-item :label="t('common.id')">
          {{ detailData.id }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('asset.freezeNo')">
          {{ detailData.freezeNo }}
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
        <el-descriptions-item :label="t('asset.freezeAmount')">
          {{ detailData.amount }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('asset.usedAmount')">
          {{ detailData.usedAmount }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('asset.unfreezeAmount')">
          {{ detailData.unfreezeAmount }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('asset.remainAmount')">
          {{ detailData.remainAmount }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('common.status')">
          {{ optionLabel('freezeStatus', detailData.status) }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('asset.expireTime')">
          {{ formatDate(detailData.expireTime) }}
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
import { assetService, type AssetFreeze, type OptionGroup } from '@/services'
import { formatDate } from '@/utils'
import TenantSelect from '@/components/TenantSelect.vue'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'

const { t } = useI18n()
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)

const loading = ref(false)
const submitLoading = ref(false)
const rows = ref<AssetFreeze[]>([])
const optionGroups = ref<OptionGroup[]>([])
const detailVisible = ref(false)
const detailData = ref<AssetFreeze | null>(null)
const changeVisible = ref(false)
const { optionItems, optionLabel } = useOptions(optionGroups)

const query = reactive({
  tenantId: undefined as number | undefined,
  userId: undefined as number | undefined,
  walletType: undefined as number | undefined,
  coin: '',
  bizNo: '',
  bizType: undefined as number | undefined,
  status: undefined as number | undefined,
  limit: 20,
})

const changeForm = reactive({
  tenantId: 0,
  freezeNo: '',
  amount: '',
  bizNo: '',
  operatorId: 0,
  remark: '',
})

const walletTypeOptions = optionItems('walletType')
const bizTypeOptions = optionItems('bizType')
const freezeStatusOptions = optionItems('freezeStatus')
const detailTitle = computed(() => `${t('asset.freezes')}${t('asset.detail')}`)
const changeTitle = computed(() => t('asset.unfreezeAsset'))

async function loadOptions() {
  const res = await assetService.getOptions()
  optionGroups.value = res.data || []
}

async function loadList() {
  loading.value = true
  try {
    const resp = await assetService.getFreezes({
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
  query.bizNo = ''
  query.bizType = undefined
  query.status = undefined
  query.limit = 100
  loadList()
}

function showDetail(row: AssetFreeze) {
  detailData.value = row
  detailVisible.value = true
}

function openChangeDialog() {
  changeForm.tenantId = 0
  changeForm.freezeNo = ''
  changeForm.amount = ''
  changeForm.bizNo = ''
  changeForm.operatorId = 0
  changeForm.remark = ''
  changeVisible.value = true
}

function prefillUnfreeze(row: AssetFreeze) {
  changeForm.tenantId = Number(row.tenantId || 0)
  changeForm.freezeNo = String(row.freezeNo || '')
  changeForm.amount = String(row.remainAmount || '')
  changeForm.bizNo = ''
  changeForm.operatorId = 0
  changeForm.remark = ''
  changeVisible.value = true
}

async function submitChange() {
  submitLoading.value = true
  try {
    await assetService.unfreezeAsset(changeForm)
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
