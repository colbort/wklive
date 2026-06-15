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
        <UserSelect v-model="query.userId" :tenant-id="query.tenantId || undefined" />
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
      <el-form-item :label="t('common.status')">
        <el-select v-model="query.status" clearable style="width: 160px">
          <el-option
            v-for="item in lockStatusOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
      </el-form-item>
      <template #actions>
        <el-button
          v-perm="'asset:lock:unlock'"
          type="success"
          plain
          @click="openChangeDialog"
        >
          {{ t('asset.unlockAsset') }}
        </el-button>
      </template>
    </CrudQueryCard>

    <el-card shadow="never" class="table-card">
      <el-table v-loading="loading" :data="rows" stripe>
        <el-table-column
          prop="lockNo"
          :label="t('asset.lockNo')"
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
        <el-table-column
          prop="walletType"
          :label="t('asset.walletType')"
          min-width="120"
          show-overflow-tooltip
        />
        <el-table-column
          prop="coin"
          :label="t('asset.coin')"
          min-width="100"
          show-overflow-tooltip
        />
        <el-table-column
          prop="amount"
          :label="t('asset.lockAmount')"
          min-width="140"
          show-overflow-tooltip
        />
        <el-table-column
          prop="unlockAmount"
          :label="t('asset.unlockAmount')"
          min-width="140"
          show-overflow-tooltip
        />
        <el-table-column
          prop="remainAmount"
          :label="t('asset.remainAmount')"
          min-width="140"
          show-overflow-tooltip
        />
        <el-table-column
          prop="status"
          :label="t('common.status')"
          min-width="100"
          show-overflow-tooltip
        />
        <el-table-column
          :label="t('common.actions')"
          align="center"
          width="180"
          fixed="right"
        >
          <template #default="{ row }">
            <el-button link type="primary" @click="showDetail(row)">
              {{ t('asset.detail') }}
            </el-button>
            <el-button
              v-perm="'asset:lock:unlock'"
              link
              type="success"
              @click="prefillUnlock(row)"
            >
              {{ t('asset.unlockAsset') }}
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
        <el-form-item :label="t('asset.lockNo')">
          <el-input v-model="changeForm.lockNo" />
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
          v-perm="'asset:lock:unlock'"
          type="primary"
          :loading="submitLoading"
          @click="submitChange"
        >
          {{ t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="detailVisible" :title="detailTitle" width="760px">
      <pre class="detail-pre">{{ JSON.stringify(detailData, null, 2) }}</pre>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { useOptions, usePagination } from '@/composables'
import { assetService, type AssetLock, type OptionGroup } from '@/services'
import TenantSelect from '@/components/TenantSelect.vue'
import UserSelect from '@/components/UserSelect.vue'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'

const { t } = useI18n()
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)

const loading = ref(false)
const submitLoading = ref(false)
const rows = ref<AssetLock[]>([])
const optionGroups = ref<OptionGroup[]>([])
const detailVisible = ref(false)
const detailData = ref<AssetLock | null>(null)
const changeVisible = ref(false)
const { optionItems } = useOptions(optionGroups)

const query = reactive({
  tenantId: undefined as number | undefined,
  userId: undefined as number | undefined,
  walletType: undefined as number | undefined,
  coin: '',
  bizNo: '',
  status: undefined as number | undefined,
  limit: 20,
})

const changeForm = reactive({
  tenantId: 0,
  lockNo: '',
  amount: '',
  bizNo: '',
  operatorId: 0,
  remark: '',
})

const walletTypeOptions = optionItems('walletType')
const lockStatusOptions = optionItems('lockStatus')
const detailTitle = computed(() => `${t('asset.locks')}${t('asset.detail')}`)
const changeTitle = computed(() => t('asset.unlockAsset'))

async function loadOptions() {
  const res = await assetService.getOptions()
  optionGroups.value = res.data || []
}

async function loadList() {
  loading.value = true
  try {
    const resp = await assetService.getLocks({
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
  query.status = undefined
  query.limit = 100
  loadList()
}

function showDetail(row: AssetLock) {
  detailData.value = row
  detailVisible.value = true
}

function openChangeDialog() {
  changeForm.tenantId = 0
  changeForm.lockNo = ''
  changeForm.amount = ''
  changeForm.bizNo = ''
  changeForm.operatorId = 0
  changeForm.remark = ''
  changeVisible.value = true
}

function prefillUnlock(row: AssetLock) {
  changeForm.tenantId = Number(row.tenantId || 0)
  changeForm.lockNo = String(row.lockNo || '')
  changeForm.amount = String(row.remainAmount || '')
  changeForm.bizNo = ''
  changeForm.operatorId = 0
  changeForm.remark = ''
  changeVisible.value = true
}

async function submitChange() {
  submitLoading.value = true
  try {
    await assetService.unlockAsset(changeForm)
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
