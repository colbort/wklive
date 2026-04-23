<template>
  <div class="module-page">
    <div class="page-header">
      <h2>{{ t('asset.freezes') }}</h2>
      <div class="header-actions">
        <el-button @click="loadList">
          {{ t('common.refresh') }}
        </el-button>
        <el-button type="success" plain @click="openChangeDialog">
          {{ t('asset.unfreezeAsset') }}
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
        <el-form-item :label="t('common.status')">
          <el-select v-model="query.status" clearable style="width: 160px">
            <el-option
              v-for="item in freezeStatusOptions"
              :key="item.value"
              :label="getOptionLabel(t, item.code, item.value)"
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
        <el-table-column
          prop="status"
          :label="t('common.status')"
          min-width="100"
          show-overflow-tooltip
        />
        <el-table-column :label="t('common.actions')" width="180" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="showDetail(row)">
              {{ t('asset.detail') }}
            </el-button>
            <el-button link type="success" @click="prefillUnfreeze(row)">
              {{ t('asset.unfreezeAsset') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="changeVisible" :title="changeTitle" width="680px">
      <el-form label-width="100px">
        <el-form-item :label="t('asset.tenantId')">
          <el-input-number v-model="changeForm.tenantId" :min="0" :precision="0" />
        </el-form-item>
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
        <el-button type="primary" :loading="submitLoading" @click="submitChange">
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
import { assetService, type AssetFreeze, type OptionGroup } from '@/services'
import { findOptionGroup, getOptionLabel } from '@/utils/options'

const { t } = useI18n()

const loading = ref(false)
const submitLoading = ref(false)
const rows = ref<AssetFreeze[]>([])
const optionGroups = ref<OptionGroup[]>([])
const detailVisible = ref(false)
const detailData = ref<AssetFreeze | null>(null)
const changeVisible = ref(false)

const query = reactive({
  tenantId: undefined as number | undefined,
  userId: undefined as number | undefined,
  walletType: undefined as number | undefined,
  coin: '',
  bizNo: '',
  status: undefined as number | undefined,
  limit: 100,
})

const changeForm = reactive({
  tenantId: 0,
  freezeNo: '',
  amount: '',
  bizNo: '',
  operatorId: 0,
  remark: '',
})

const walletTypeOptions = computed(() => findOptionGroup(optionGroups.value, 'walletType'))
const freezeStatusOptions = computed(() => findOptionGroup(optionGroups.value, 'freezeStatus'))
const detailTitle = computed(() => `${t('asset.freezes')}${t('asset.detail')}`)
const changeTitle = computed(() => t('asset.unfreezeAsset'))

function pickList(res: any) {
  return res?.data || res?.list || []
}

async function loadOptions() {
  const res = await assetService.getOptions()
  optionGroups.value = res.data || []
}

async function loadList() {
  loading.value = true
  try {
    rows.value = pickList(await assetService.getFreezes(query))
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

onMounted(loadList)
onMounted(loadOptions)
</script>

<style scoped>
.query-card :deep(.el-form-item) {
  margin-bottom: 12px;
}
</style>
