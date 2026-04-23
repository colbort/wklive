<template>
  <div class="payment-page">
    <div class="page-header">
      <h2>{{ t('payment.tenantAccounts') }}</h2>
      <div>
        <el-button type="primary" @click="openDialog()">
          {{ t('payment.addAccount') }}
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
        <el-form-item :label="t('common.keyword')">
          <el-input v-model="query.keyword" clearable />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadList">
            {{ t('common.search') }}
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="never">
      <el-table v-loading="loading" :data="list" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="tenantId" :label="t('common.tenantId')" width="100" />
        <el-table-column prop="platformId" :label="t('payment.platformId')" width="100" />
        <el-table-column prop="accountCode" :label="t('payment.accountCode')" min-width="140" />
        <el-table-column prop="accountName" :label="t('payment.accountName')" min-width="160" />
        <el-table-column prop="merchantId" :label="t('payment.merchantId')" min-width="140" />
        <el-table-column :label="t('common.status')" width="100">
          <template #default="{ row }">
            <el-tag :class="getStatusTagClass(row.status)" disable-transitions>
              {{ getOptionValueLabel(optionGroups, 'status', row.status, t) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="t('common.default')" width="90">
          <template #default="{ row }">
            {{ getOptionValueLabel(optionGroups, 'yesNo', row.isDefault, t) }}
          </template>
        </el-table-column>
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
      :title="form.id ? t('payment.editAccount') : t('payment.addAccount')"
      width="760px"
    >
      <el-form label-width="120px">
        <el-form-item :label="t('common.tenantId')">
          <div class="verify-row">
            <el-input-number
              v-model="form.tenantId"
              :min="1"
              :precision="0"
              :disabled="!!form.id"
              @change="handleTenantChange"
            />
            <el-button v-if="!form.id" :loading="tenantChecking" @click="checkTenant">
              {{ t('payment.verifyTenant') }}
            </el-button>
            <span v-if="!form.id && tenantVerified" class="verified-text">
              {{ t('payment.verified') }}
            </span>
          </div>
        </el-form-item>

        <el-form-item v-if="!form.id" :label="t('payment.tenantPayPlatformId')">
          <div class="verify-row">
            <el-input-number
              v-model="form.tenantPayPlatformId"
              :min="1"
              :precision="0"
              @change="handleTenantPlatformChange"
            />
            <el-button :loading="tenantPlatformChecking" @click="checkTenantPlatform">
              {{ t('payment.verifyTenantPlatform') }}
            </el-button>
            <span v-if="tenantPlatformVerified" class="verified-text">
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
              @change="handlePlatformChange"
            />
            <el-button :loading="platformChecking" @click="checkPlatform">
              {{ t('payment.verifyPlatform') }}
            </el-button>
            <span v-if="platformVerified" class="verified-text"> {{ t('payment.verified') }} </span>
          </div>
        </el-form-item>

        <el-form-item v-if="!form.id" :label="t('payment.accountCode')">
          <el-input v-model="form.accountCode" />
        </el-form-item>
        <el-form-item :label="t('payment.accountName')">
          <el-input v-model="form.accountName" />
        </el-form-item>
        <el-form-item label="APP ID">
          <el-input v-model="form.appId" />
        </el-form-item>
        <el-form-item :label="t('payment.merchantId')">
          <el-input v-model="form.merchantId" />
        </el-form-item>
        <el-form-item :label="t('payment.merchantName')">
          <el-input v-model="form.merchantName" />
        </el-form-item>
        <el-form-item :label="t('payment.apiKeyCipher')">
          <el-input v-model="form.apiKeyCipher" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item :label="t('payment.apiSecretCipher')">
          <el-input v-model="form.apiSecretCipher" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item :label="t('payment.privateKeyCipher')">
          <el-input v-model="form.privateKeyCipher" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item :label="t('payment.publicKey')">
          <el-input v-model="form.publicKey" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item :label="t('payment.certCipher')">
          <el-input v-model="form.certCipher" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item :label="t('payment.extConfig')">
          <el-input v-model="form.extConfig" type="textarea" :rows="3" />
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
        <el-form-item :label="t('common.default')">
          <el-select v-model="form.isDefault" style="width: 100%">
            <el-option
              v-for="item in yesNoOptions"
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
          {{ t('common.cancel') }}
        </el-button>
        <el-button type="primary" :disabled="submitDisabled" @click="submit">
          {{ t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="detailVisible" :title="t('payment.detailTitle')" width="760px">
      <pre class="detail-pre">{{ JSON.stringify(detailData, null, 2) }}</pre>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import {
  catalogService,
  tenantService,
  tenantsService,
  type OptionGroup,
  type TenantPayAccount,
} from '@/services'
import { findOptionGroup, getOptionLabel, getOptionValueLabel } from '@/utils/options'

const { t } = useI18n()

const loading = ref(false)
const list = ref<TenantPayAccount[]>([])
const dialogVisible = ref(false)
const detailVisible = ref(false)
const detailData = ref<Record<string, unknown>>({})

const optionGroups = ref<OptionGroup[]>([])
const statusOptions = computed(() => findOptionGroup(optionGroups.value, 'status'))
const yesNoOptions = computed(() => findOptionGroup(optionGroups.value, 'yesNo'))

const tenantChecking = ref(false)
const tenantPlatformChecking = ref(false)
const platformChecking = ref(false)
const tenantVerified = ref(false)
const tenantPlatformVerified = ref(false)
const platformVerified = ref(false)
const verifiedTenantId = ref(0)
const verifiedTenantPlatformId = ref(0)
const verifiedPlatformId = ref(0)

const query = reactive({
  tenantId: 0,
  platformId: 0,
  keyword: '',
})

const createEmptyForm = () => ({
  id: 0,
  tenantId: 0,
  tenantPayPlatformId: 0,
  platformId: 0,
  accountCode: '',
  accountName: '',
  appId: '',
  merchantId: '',
  merchantName: '',
  apiKeyCipher: '',
  apiSecretCipher: '',
  privateKeyCipher: '',
  publicKey: '',
  certCipher: '',
  extConfig: '',
  status: 1,
  isDefault: 1,
  remark: '',
})

const form = reactive(createEmptyForm())

const submitDisabled = computed(
  () =>
    !form.id &&
    (!tenantVerified.value ||
      !tenantPlatformVerified.value ||
      !platformVerified.value ||
      verifiedTenantId.value !== form.tenantId ||
      verifiedTenantPlatformId.value !== form.tenantPayPlatformId ||
      verifiedPlatformId.value !== form.platformId),
)

const loadOptions = async () => {
  const res = await tenantService.getOptions()
  optionGroups.value = res.data || []
}

const loadList = async () => {
  loading.value = true
  try {
    const res = await tenantService.getTenantAccountList({
      ...query,
      tenantId: query.tenantId || undefined,
      platformId: query.platformId || undefined,
      keyword: query.keyword || undefined,
      limit: 100,
    })
    list.value = res.data || []
  } finally {
    loading.value = false
  }
}

const resetVerifyState = () => {
  tenantVerified.value = false
  tenantPlatformVerified.value = false
  platformVerified.value = false
  verifiedTenantId.value = 0
  verifiedTenantPlatformId.value = 0
  verifiedPlatformId.value = 0
}

const openDialog = (row?: TenantPayAccount) => {
  Object.assign(form, createEmptyForm(), row || {})
  if (row?.id) {
    tenantVerified.value = true
    tenantPlatformVerified.value = true
    platformVerified.value = true
    verifiedTenantId.value = row.tenantId
    verifiedTenantPlatformId.value = row.tenantPayPlatformId
    verifiedPlatformId.value = row.platformId
  } else {
    resetVerifyState()
  }
  dialogVisible.value = true
}

const handleTenantChange = () => {
  tenantVerified.value = false
  verifiedTenantId.value = 0
  tenantPlatformVerified.value = false
  verifiedTenantPlatformId.value = 0
}

const handleTenantPlatformChange = () => {
  tenantPlatformVerified.value = false
  verifiedTenantPlatformId.value = 0
}

const handlePlatformChange = () => {
  platformVerified.value = false
  verifiedPlatformId.value = 0
  tenantPlatformVerified.value = false
  verifiedTenantPlatformId.value = 0
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

const validateTenantPlatformExists = async (
  tenantPayPlatformId: number,
  tenantId: number,
  platformId: number,
) => {
  if (!tenantPayPlatformId) {
    ElMessage.warning(t('payment.pleaseInputTenantPlatformId'))
    return false
  }
  if (!tenantId) {
    ElMessage.warning(t('payment.pleaseInputTenantFirst'))
    return false
  }
  if (!platformId) {
    ElMessage.warning(t('payment.pleaseInputPlatformFirst'))
    return false
  }

  tenantPlatformChecking.value = true
  try {
    const res = await tenantService.getTenantPlatformDetail(tenantPayPlatformId, tenantId)
    if (!res.data?.id) {
      ElMessage.error(t('payment.tenantPlatformNotFound'))
      return false
    }
    if (res.data.platformId !== platformId) {
      ElMessage.error(t('payment.tenantPlatformPlatformMismatch'))
      return false
    }
    return true
  } catch {
    ElMessage.error(t('payment.tenantPlatformNotFound'))
    return false
  } finally {
    tenantPlatformChecking.value = false
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

const checkTenantPlatform = async () => {
  const exists = await validateTenantPlatformExists(
    form.tenantPayPlatformId,
    form.tenantId,
    form.platformId,
  )
  tenantPlatformVerified.value = exists
  verifiedTenantPlatformId.value = exists ? form.tenantPayPlatformId : 0
  if (exists) {
    ElMessage.success(t('payment.tenantPlatformVerifiedSuccess'))
  }
}

const submit = async () => {
  if (!form.id && submitDisabled.value) {
    ElMessage.warning(t('payment.pleaseCompleteAccountValidation'))
    return
  }

  if (form.id) {
    await tenantService.updateTenantAccount({ ...form })
  } else {
    await tenantService.createTenantAccount({ ...form })
  }
  ElMessage.success(t('common.operationSuccess'))
  dialogVisible.value = false
  loadList()
}

const showDetail = async (row: TenantPayAccount) => {
  const res = await tenantService.getTenantAccountDetail(row.id, row.tenantId)
  detailData.value = res.data || row
  detailVisible.value = true
}

function getStatusTagClass(value?: number) {
  const num = Number(value ?? 0)
  if (num === 1) return 'option-tag option-tag--green'
  if (num === 2) return 'option-tag option-tag--red'
  return 'option-tag option-tag--slate'
}

onMounted(async () => {
  await Promise.all([loadOptions(), loadList()])
})
</script>

<style scoped>
.query-card :deep(.el-form-item) {
  margin-bottom: 12px;
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
</style>
