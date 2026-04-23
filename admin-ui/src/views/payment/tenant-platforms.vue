<template>
  <div class="payment-page module-page">
    <div class="page-header">
      <h2>{{ t('payment.tenantPlatforms') }}</h2>
      <div class="header-actions">
        <el-button type="primary" @click="openDialog()">
          {{ t('payment.addTenantPlatform') }}
        </el-button>
        <el-button @click="loadList">
          {{ t('common.refresh') }}
        </el-button>
      </div>
    </div>
    <el-card shadow="never" class="query-card">
      <el-form :model="query" inline label-width="90px">
        <el-form-item :label="t('common.tenantId')">
          <el-input-number v-model="query.tenantId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('payment.platformId')">
          <el-input-number v-model="query.platformId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('common.status')">
          <el-select v-model="query.status" clearable style="width: 160px">
            <el-option :label="t('payment.all')" :value="0" />
            <el-option
              v-for="item in statusOptions"
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
        </el-form-item>
      </el-form>
    </el-card>
    <el-card shadow="never" class="table-card">
      <el-table v-loading="loading" :data="list" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="tenantId" :label="t('common.tenantId')" width="100" />
        <el-table-column prop="platformId" :label="t('payment.platformId')" width="100" />
        <el-table-column :label="t('common.status')" width="100">
          <template #default="{ row }">
            <el-tag :class="getOptionTagClass(row.status)" disable-transitions>
              {{ getStatusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="t('payment.openStatus')" width="120">
          <template #default="{ row }">
            <el-tag :class="getOpenStatusTagClass(row.openStatus)" disable-transitions>
              {{ getOpenStatusLabel(row.openStatus) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="remark" :label="t('common.remark')" min-width="180" />
        <el-table-column :label="t('common.actions')" width="160">
          <template #default="{ row }">
            <el-button link type="primary" @click="showDetail(row)">
              {{ t('common.detail') }}
            </el-button>
            <el-button link type="primary" @click="openDialog(row)">
              {{ t('common.edit') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
    <el-dialog
      v-model="dialogVisible"
      :title="form.id ? t('payment.editTenantPlatform') : t('payment.openTenantPlatform')"
      width="620px"
    >
      <el-form label-width="100px">
        <el-form-item :label="t('common.tenantId')">
          <div class="verify-row">
            <el-input-number
              v-model="form.tenantId"
              :min="1"
              :precision="0"
              :disabled="!!form.id"
              @change="handleTenantIdChange"
            />
            <el-button v-if="!form.id" :loading="tenantChecking" @click="checkTenant">
              {{ t('payment.verifyTenant') }}
            </el-button>
            <span v-if="!form.id && tenantVerified" class="verified-text">
              {{ t('payment.verified') }}
            </span>
          </div>
        </el-form-item>
        <el-form-item v-if="!form.id" :label="t('payment.platformId')">
          <div class="verify-row">
            <el-input-number
              v-model="form.platformId"
              :min="1"
              :precision="0"
              @change="handlePlatformIdChange"
            />
            <el-button :loading="platformChecking" @click="checkPlatform">
              {{ t('payment.verifyPlatform') }}
            </el-button>
            <span v-if="platformVerified" class="verified-text"> {{ t('payment.verified') }} </span>
          </div>
        </el-form-item>
        <el-form-item :label="t('common.status')">
          <el-select v-model="form.status" style="width: 100%">
            <el-option
              v-for="item in statusOptions"
              :key="item.value"
              :label="getOptionLabel(t, item.code, item.value)"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('payment.openStatus')">
          <el-select v-model="form.openStatus" style="width: 100%">
            <el-option
              v-for="item in openStatusOptions"
              :key="item.value"
              :label="getOptionLabel(t, item.code, item.value)"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('common.remark')">
          <el-input v-model="form.remark" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">
          {{ t('common.cancel') }} </el-button
        ><el-button type="primary" :disabled="submitDisabled" @click="submit">
          {{ t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>
    <el-dialog v-model="detailVisible" :title="t('payment.detailTitle')" width="680px">
      <pre class="detail-pre">{{ JSON.stringify(detailData, null, 2) }}</pre>
    </el-dialog>
  </div>
</template>
<script setup lang="ts">
import { onMounted, reactive, ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import {
  catalogService,
  tenantService,
  tenantsService,
  type TenantPayPlatform,
  type OptionGroup,
} from '@/services'
import { findOptionGroup, getOptionLabel, getOptionValueLabel } from '@/utils/options'

const { t } = useI18n()

// reactive data
const loading = ref(false)
const list = ref<TenantPayPlatform[]>([])
const dialogVisible = ref(false)
const detailVisible = ref(false)
const detailData = ref<TenantPayPlatform | null>(null)
const tenantChecking = ref(false)
const platformChecking = ref(false)
const tenantVerified = ref(false)
const platformVerified = ref(false)
const verifiedTenantId = ref(0)
const verifiedPlatformId = ref(0)

// query parameters for the list
const query = reactive({ tenantId: 0, platformId: 0, status: 0 })

// form data for create / edit
const form = reactive({ id: 0, tenantId: 0, platformId: 0, status: 1, openStatus: 1, remark: '' })

// option groups fetched from backend
const optionGroups = ref<OptionGroup[]>([])
const statusOptions = computed(() => findOptionGroup(optionGroups.value, 'status'))
const openStatusOptions = computed(() => findOptionGroup(optionGroups.value, 'openStatus'))
const submitDisabled = computed(
  () =>
    !form.id &&
    (!tenantVerified.value ||
      !platformVerified.value ||
      verifiedTenantId.value !== form.tenantId ||
      verifiedPlatformId.value !== form.platformId),
)

// fetch all option groups once
const fetchOptions = async () => {
  try {
    const res = await tenantService.getOptions()
    optionGroups.value = res.data || []
  } catch (e) {
    console.error('load options failed', e)
  }
}

// load the list data
const loadList = async () => {
  loading.value = true
  try {
    const res = await tenantService.getTenantPlatformList({
      ...query,
      tenantId: query.tenantId || undefined,
      platformId: query.platformId || undefined,
      limit: 100,
    })
    list.value = res.data || []
  } finally {
    loading.value = false
  }
}

// open dialog for create or edit
const openDialog = (row?: TenantPayPlatform) => {
  Object.assign(
    form,
    row || { id: 0, tenantId: 0, platformId: 0, status: 1, openStatus: 1, remark: '' },
  )
  tenantVerified.value = !!row?.id
  platformVerified.value = !!row?.id
  verifiedTenantId.value = row?.id ? row.tenantId : 0
  verifiedPlatformId.value = row?.id ? row.platformId : 0
  dialogVisible.value = true
}

const handleTenantIdChange = () => {
  tenantVerified.value = false
  verifiedTenantId.value = 0
}

const handlePlatformIdChange = () => {
  platformVerified.value = false
  verifiedPlatformId.value = 0
}

const validateTenantExists = async (tenantId: number) => {
  if (!tenantId) {
    ElMessage.warning(t('payment.pleaseInputTenantId'))
    return false
  }

  tenantChecking.value = true
  try {
    const res = await tenantsService.detail({ tenantId })
    if (!res.data?.id) {
      ElMessage.error(t('payment.tenantNotFound'))
      return false
    }
    return true
  } catch {
    ElMessage.error(t('payment.tenantNotFound'))
    return false
  } finally {
    tenantChecking.value = false
  }
}

const validatePlatformExists = async (platformId: number) => {
  if (!platformId) {
    ElMessage.warning(t('payment.pleaseInputPlatformId'))
    return false
  }

  platformChecking.value = true
  try {
    const res = await catalogService.getPlatformDetail(platformId)
    if (!res.data?.id) {
      ElMessage.error(t('payment.platformNotFound'))
      return false
    }
    return true
  } catch {
    ElMessage.error(t('payment.platformNotFound'))
    return false
  } finally {
    platformChecking.value = false
  }
}

const checkTenant = async () => {
  const exists = await validateTenantExists(form.tenantId)
  tenantVerified.value = exists
  verifiedTenantId.value = exists ? form.tenantId : 0
  if (exists) {
    ElMessage.success(t('payment.tenantVerifiedSuccess'))
  }
}

const checkPlatform = async () => {
  const exists = await validatePlatformExists(form.platformId)
  platformVerified.value = exists
  verifiedPlatformId.value = exists ? form.platformId : 0
  if (exists) {
    ElMessage.success(t('payment.platformVerifiedSuccess'))
  }
}

// submit create / edit
const submit = async () => {
  if (!form.id && submitDisabled.value) {
    ElMessage.warning(t('payment.pleaseCompleteTenantPlatformValidation'))
    return
  }

  if (form.id) {
    await tenantService.updateTenantPlatform({ ...form })
  } else {
    await tenantService.openTenantPlatform({
      tenantId: form.tenantId,
      platformId: form.platformId,
      status: form.status,
      openStatus: form.openStatus,
      remark: form.remark,
    })
  }
  ElMessage.success(t('common.operationSuccess'))
  dialogVisible.value = false
  loadList()
}

// show detail (kept as raw JSON for now)
const showDetail = async (row: TenantPayPlatform) => {
  const res = await tenantService.getTenantPlatformDetail(row.id, row.tenantId)
  detailData.value = res.data || row
  detailVisible.value = true
}

const getStatusLabel = (value?: number) =>
  getOptionValueLabel(optionGroups.value, 'status', value, t)

const getOpenStatusLabel = (value?: number) =>
  getOptionValueLabel(optionGroups.value, 'openStatus', value, t)

// tag class helper – green for enabled (1), red for disabled (2), slate for others
function getOptionTagClass(value?: number) {
  const num = Number(value ?? 0)
  if (num === 1) return 'option-tag option-tag--green'
  if (num === 2) return 'option-tag option-tag--red'
  return 'option-tag option-tag--slate'
}

function getOpenStatusTagClass(value?: number) {
  const num = Number(value ?? 0)
  if (num === 2) return 'option-tag option-tag--green'
  if (num === 3) return 'option-tag option-tag--warning'
  if (num === 4) return 'option-tag option-tag--red'
  return 'option-tag option-tag--slate'
}

onMounted(() => {
  fetchOptions()
  loadList()
})
</script>
<style scoped>
.option-tag {
  border: none;
}

.option-tag--green {
  color: var(--el-color-success);
  background: var(--el-color-success-light-9);
}

.option-tag--red {
  color: var(--el-color-danger);
  background: var(--el-color-danger-light-9);
}

.option-tag--slate {
  color: var(--el-text-color-regular);
  background: var(--el-fill-color-light);
}

.option-tag--warning {
  color: var(--el-color-warning);
  background: var(--el-color-warning-light-9);
}

.verify-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.verified-text {
  color: var(--el-color-success);
  font-size: 14px;
}
</style>
