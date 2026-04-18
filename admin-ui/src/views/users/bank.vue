<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  memberUserService,
  tenantsService,
  type AddUserBankReq,
  type OptionGroup,
  type UserBankItem,
  type UpdateMemberUserBankReq,
} from '@/services'
import { formatDate } from '@/utils'
import { findOptionGroup, getOptionLabel, getOptionValueLabel } from '@/utils/options'

const { t } = useI18n()
const loading = ref(false)
const submitLoading = ref(false)
const list = ref<UserBankItem[]>([])
const editVisible = ref(false)
const statusVisible = ref(false)
const detailVisible = ref(false)
const detailData = ref<UserBankItem>()
const tenantChecking = ref(false)
const tenantChecked = ref(false)
const tenantExists = ref(false)
const tenantCheckName = ref('')
const userChecking = ref(false)
const userChecked = ref(false)
const userExists = ref(false)
const userCheckName = ref('')
const userCheckUserNo = ref('')
const optionGroups = ref<OptionGroup[]>([])
const bankStatusOptions = computed(() => findOptionGroup(optionGroups.value, 'bankStatus'))

const query = reactive({
  tenantId: undefined as number | undefined,
  userId: undefined as number | undefined,
  keyword: '',
  status: undefined as number | undefined,
  limit: 100,
})

const form = reactive<any>({
  id: 0,
  tenantId: 0,
  userId: 0,
  bankName: '',
  bankCode: '',
  accountName: '',
  accountNo: '',
  branchName: '',
  countryCode: '',
  isDefault: 0,
  status: 1,
})

const statusForm = reactive({
  id: 0,
  tenantId: 0,
  status: 1,
})

const isCreate = computed(() => !form.id)
const canSubmitCreate = computed(
  () =>
    !isCreate.value ||
    (tenantChecked.value && tenantExists.value && userChecked.value && userExists.value),
)

function checkCode(code: number) {
  return code === 0 || code === 200
}

function getBooleanLabel(value?: number) {
  return Number(value) === 1 ? t('users.yes') : t('users.no')
}

function getBooleanTagClass(value?: number) {
  return Number(value) === 1 ? 'option-tag option-tag--green' : 'option-tag option-tag--red'
}

function getBankStatusTagClass(value?: number) {
  const bankStatusMap: Record<number, string> = {
    1: 'option-tag option-tag--green',
    2: 'option-tag option-tag--red',
  }
  return bankStatusMap[Number(value ?? 0)] || 'option-tag'
}

function getBankStatusLabel(value?: number) {
  return getOptionValueLabel(optionGroups.value, 'bankStatus', value, t)
}

async function fetchOptions() {
  try {
    const res = await memberUserService.getOptions()
    if (!checkCode(res.code)) throw new Error(res.msg || t('users.loadOptionsFailed'))
    optionGroups.value = res.data || []
  } catch (error: unknown) {
    ElMessage.error(error instanceof Error ? error.message : t('users.loadOptionsFailed'))
  }
}

function resetTenantCheck() {
  tenantChecked.value = false
  tenantExists.value = false
  tenantCheckName.value = ''
}

function resetUserCheck() {
  userChecked.value = false
  userExists.value = false
  userCheckName.value = ''
  userCheckUserNo.value = ''
}

function onTenantChange() {
  resetTenantCheck()
  resetUserCheck()
}

function onUserChange() {
  resetUserCheck()
}

async function verifyTenant() {
  const tenantId = Number(form.tenantId || 0)
  if (!tenantId) {
    resetTenantCheck()
    ElMessage.warning(t('users.queryTenantPrompt'))
    return false
  }

  tenantChecking.value = true
  try {
    const res = await tenantsService.detail({ tenantId })
    if (!checkCode(res.code)) throw new Error(res.msg || t('users.queryTenantFailed'))

    const tenant = res.data
    tenantChecked.value = true
    tenantExists.value = Boolean(tenant)
    tenantCheckName.value = tenant?.tenantName || ''

    if (!tenant) {
      ElMessage.warning(t('users.tenantNotFoundPrompt'))
      return false
    }
    ElMessage.success(t('users.tenantFound', { name: tenant.tenantName }))
    return true
  } catch (error: unknown) {
    resetTenantCheck()
    ElMessage.error(error instanceof Error ? error.message : t('users.queryTenantFailed'))
    return false
  } finally {
    tenantChecking.value = false
  }
}

async function verifyUser() {
  const tenantId = Number(form.tenantId || 0)
  const userId = Number(form.userId || 0)

  if (!tenantId) {
    ElMessage.warning(t('users.inputTenantAndConfirm'))
    return false
  }
  if (!tenantChecked.value || !tenantExists.value) {
    const verified = await verifyTenant()
    if (!verified) return false
  }
  if (!userId) {
    resetUserCheck()
    ElMessage.warning(t('users.queryUserPrompt'))
    return false
  }

  userChecking.value = true
  try {
    const res = await memberUserService.getDetail(userId, tenantId)
    if (!checkCode(res.code)) throw new Error(res.msg || t('users.queryUserFailed'))

    const data = res.detail || res.data
    const base = data?.base
    userChecked.value = true
    userExists.value = Boolean(base?.id)
    userCheckName.value = base?.username || ''
    userCheckUserNo.value = base?.userNo || ''

    if (!base?.id) {
      ElMessage.warning(t('users.userNotFoundPrompt'))
      return false
    }
    ElMessage.success(t('users.userFound', { name: base.username }))
    return true
  } catch (error: unknown) {
    resetUserCheck()
    ElMessage.error(error instanceof Error ? error.message : t('users.queryUserFailed'))
    return false
  } finally {
    userChecking.value = false
  }
}

async function fetchList() {
  loading.value = true
  try {
    const res = await memberUserService.listBanks(query)
    if (!checkCode(res.code)) throw new Error(res.msg || t('users.loadFailed'))
    list.value = res.data || []
  } catch (error: unknown) {
    ElMessage.error(error instanceof Error ? error.message : t('users.loadFailed'))
  } finally {
    loading.value = false
  }
}

function resetQuery() {
  Object.assign(query, {
    tenantId: undefined,
    userId: undefined,
    keyword: '',
    status: undefined,
    limit: 100,
  })
  fetchList()
}

function openCreate() {
  Object.assign(form, {
    id: 0,
    tenantId: Number(query.tenantId || 0),
    userId: Number(query.userId || 0),
    bankName: '',
    bankCode: '',
    accountName: '',
    accountNo: '',
    branchName: '',
    countryCode: '',
    isDefault: 0,
    status: 1,
  })
  resetTenantCheck()
  resetUserCheck()
  editVisible.value = true
}

async function openEdit(row: UserBankItem) {
  const tenantId = Number(row.tenantId || query.tenantId || 0)
  if (!tenantId) {
    ElMessage.warning(t('users.queryTenantPrompt'))
    return
  }
  const res = await memberUserService.getBank(row.id, tenantId)
  if (!checkCode(res.code)) {
    ElMessage.error(res.msg || t('users.loadDetailFailed'))
    return
  }
  Object.assign(form, res.bank || res.data, { id: row.id })
  resetTenantCheck()
  resetUserCheck()
  editVisible.value = true
}

async function showDetail(row: UserBankItem) {
  const res = await memberUserService.getBank(row.id, row.tenantId)
  if (!checkCode(res.code)) {
    ElMessage.error(res.msg)
    return
  }
  detailData.value = res.bank || res.data
  detailVisible.value = true
}

async function submitEdit() {
  submitLoading.value = true
  try {
    if (form.id) {
      const payload: UpdateMemberUserBankReq = {
        tenantId: Number(form.tenantId),
        userId: Number(form.userId),
        bankName: form.bankName,
        bankCode: form.bankCode || undefined,
        accountName: form.accountName,
        accountNo: form.accountNo,
        branchName: form.branchName || undefined,
        countryCode: form.countryCode || undefined,
        isDefault: Number(form.isDefault),
        status: Number(form.status),
      }
      const res = await memberUserService.updateBank(Number(form.id), payload)
      if (!checkCode(res.code)) throw new Error(res.msg || t('users.updateFailed'))
    } else {
      if (!tenantChecked.value || !tenantExists.value) {
        const verifiedTenant = await verifyTenant()
        if (!verifiedTenant) return
      }
      if (!userChecked.value || !userExists.value) {
        const verifiedUser = await verifyUser()
        if (!verifiedUser) return
      }
      const payload: AddUserBankReq = {
        tenantId: Number(form.tenantId),
        userId: Number(form.userId),
        bankName: form.bankName,
        bankCode: form.bankCode || undefined,
        accountName: form.accountName,
        accountNo: form.accountNo,
        branchName: form.branchName || undefined,
        countryCode: form.countryCode || undefined,
        isDefault: Number(form.isDefault),
        status: Number(form.status),
      }
      const res = await memberUserService.addBank(payload)
      if (!checkCode(res.code)) throw new Error(res.msg || t('users.createFailed'))
    }
    ElMessage.success(t('users.saveSuccess'))
    editVisible.value = false
    fetchList()
  } catch (error: unknown) {
    ElMessage.error(error instanceof Error ? error.message : t('users.saveFailed'))
  } finally {
    submitLoading.value = false
  }
}

function openStatus(row: UserBankItem) {
  const tenantId = Number(row.tenantId || query.tenantId || 0)
  if (!tenantId) {
    ElMessage.warning(t('users.queryTenantPrompt'))
    return
  }
  Object.assign(statusForm, {
    id: row.id,
    tenantId,
    status: Number(row.status || 1),
  })
  statusVisible.value = true
}

async function submitStatus() {
  try {
    const res = await memberUserService.updateBankStatus(statusForm.id, {
      tenantId: Number(statusForm.tenantId),
      status: Number(statusForm.status),
    })
    if (!checkCode(res.code)) throw new Error(res.msg || t('users.updateFailed'))
    ElMessage.success(t('users.updateSuccess'))
    statusVisible.value = false
    fetchList()
  } catch (error: unknown) {
    ElMessage.error(error instanceof Error ? error.message : t('users.updateFailed'))
  }
}

async function setDefault(row: UserBankItem) {
  const tenantId = Number(row.tenantId || query.tenantId || 0)
  try {
    const res = await memberUserService.setDefaultBank(row.id, { tenantId, userId: row.userId })
    if (!checkCode(res.code)) throw new Error(res.msg || t('users.setFailed'))
    ElMessage.success(t('users.setSuccess'))
    fetchList()
  } catch (error: unknown) {
    ElMessage.error(error instanceof Error ? error.message : t('users.setFailed'))
  }
}

async function remove(row: UserBankItem) {
  const tenantId = Number(row.tenantId || query.tenantId || 0)
  try {
    await ElMessageBox.confirm(
      t('users.deleteBankConfirm', { name: row.bankName }),
      t('common.warning'),
      { type: 'warning' },
    )
    const res = await memberUserService.deleteBank(row.id, tenantId)
    if (!checkCode(res.code)) throw new Error(res.msg || t('users.deleteFailed'))
    ElMessage.success(t('users.deleteSuccess'))
    fetchList()
  } catch (error: unknown) {
    if (error === 'cancel') return
    ElMessage.error(error instanceof Error ? error.message : t('users.deleteFailed'))
  }
}

onMounted(fetchList)
onMounted(fetchOptions)
</script>

<template>
  <div class="module-page">
    <div class="page-header">
      <h2>{{ t('users.banks') }}</h2>
      <div class="header-actions">
        <el-button @click="fetchList"> {{ t('common.refresh') }} </el-button>
        <el-button type="primary" @click="openCreate"> {{ t('users.addBank') }} </el-button>
      </div>
    </div>

    <el-card shadow="never" class="query-card">
      <el-form :model="query" inline label-width="90px">
        <el-form-item :label="t('common.tenantId')">
          <el-input-number v-model="query.tenantId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('users.userId')">
          <el-input-number v-model="query.userId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('common.keyword')">
          <el-input v-model="query.keyword" clearable />
        </el-form-item>
        <el-form-item :label="t('users.status')">
          <el-select v-model="query.status" clearable style="width: 140px">
            <el-option
              v-for="item in bankStatusOptions"
              :key="item.value"
              :label="getOptionLabel(t, item.code, item.value)"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="fetchList"> {{ t('common.search') }} </el-button>
          <el-button @click="resetQuery"> {{ t('common.reset') }} </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="never" class="table-card">
      <el-table v-loading="loading" :data="list" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="tenantId" :label="t('common.tenantId')" width="100" />
        <el-table-column prop="userId" :label="t('users.userId')" width="100" />
        <el-table-column prop="bankName" :label="t('users.bankName')" min-width="140" />
        <el-table-column prop="accountName" :label="t('users.accountName')" min-width="120" />
        <el-table-column prop="accountNo" :label="t('users.accountNo')" min-width="160" show-overflow-tooltip />
        <el-table-column prop="isDefault" :label="t('common.default')" width="80">
          <template #default="{ row }">
            <span :class="getBooleanTagClass(row.isDefault)">
              {{ getBooleanLabel(row.isDefault) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column :label="t('users.status')" width="90">
          <template #default="{ row }">
            <span :class="getBankStatusTagClass(row.status)">
              {{ getBankStatusLabel(row.status) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column :label="t('common.createTimes')" min-width="170">
          <template #default="{ row }">
            {{ formatDate(row.createTimes) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('common.actions')" width="260" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="showDetail(row)"> {{ t('common.detail') }} </el-button>
            <el-button link type="primary" @click="openEdit(row)"> {{ t('common.edit') }} </el-button>
            <el-button link type="warning" @click="openStatus(row)"> {{ t('users.status') }} </el-button>
            <el-button link type="success" @click="setDefault(row)"> {{ t('users.setDefault') }} </el-button>
            <el-button link type="danger" @click="remove(row)"> {{ t('common.delete') }} </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="editVisible" :title="form.id ? t('users.editBank') : t('users.addBank')" width="620px">
      <el-form label-width="100px">
        <el-form-item :label="t('common.tenantId')">
          <div class="verify-row">
            <el-input-number
              v-model="form.tenantId"
              :min="0"
              :precision="0"
              :disabled="!isCreate"
              @change="onTenantChange"
            />
            <el-button
              v-if="isCreate"
              type="primary"
              plain
              :loading="tenantChecking"
              @click="verifyTenant"
            >
              {{ t('users.confirmTenant') }}
            </el-button>
          </div>
          <div v-if="isCreate" class="verify-tip">
            <span v-if="tenantChecked && tenantExists" class="verify-tip verify-tip--success">
              {{ t('users.tenantConfirmed', { name: tenantCheckName || form.tenantId }) }}
            </span>
            <span v-else-if="tenantChecked" class="verify-tip verify-tip--error">
              {{ t('users.tenantMissingRetry') }}
            </span>
            <span v-else class="verify-tip verify-tip--muted"> {{ t('users.confirmTenantBeforeCreate') }} </span>
          </div>
        </el-form-item>
        <el-form-item :label="t('users.userId')">
          <div class="verify-row">
            <el-input-number
              v-model="form.userId"
              :min="0"
              :precision="0"
              :disabled="!isCreate"
              @change="onUserChange"
            />
            <el-button
              v-if="isCreate"
              type="primary"
              plain
              :loading="userChecking"
              @click="verifyUser"
            >
              {{ t('users.confirmUser') }}
            </el-button>
          </div>
          <div v-if="isCreate" class="verify-tip">
            <span v-if="userChecked && userExists" class="verify-tip verify-tip--success">
              {{ t('users.userConfirmed', { name: userCheckName }) }}
              <template v-if="userCheckUserNo"> ({{ userCheckUserNo }}) </template>
            </span>
            <span v-else-if="userChecked" class="verify-tip verify-tip--error">
              {{ t('users.userMissingRetry') }}
            </span>
            <span v-else class="verify-tip verify-tip--muted"> {{ t('users.confirmUserBeforeCreate') }} </span>
          </div>
        </el-form-item>
        <el-form-item :label="t('users.bankName')">
          <el-input v-model="form.bankName" />
        </el-form-item>
        <el-form-item :label="t('users.bankCode')">
          <el-input v-model="form.bankCode" />
        </el-form-item>
        <el-form-item :label="t('users.accountName')">
          <el-input v-model="form.accountName" />
        </el-form-item>
        <el-form-item :label="t('users.accountNo')">
          <el-input v-model="form.accountNo" />
        </el-form-item>
        <el-form-item :label="t('users.branchName')">
          <el-input v-model="form.branchName" />
        </el-form-item>
        <el-form-item :label="t('users.countryCode')">
          <el-input v-model="form.countryCode" />
        </el-form-item>
        <el-form-item :label="t('common.default')">
          <el-switch v-model="form.isDefault" :active-value="1" :inactive-value="0" />
        </el-form-item>
        <el-form-item :label="t('users.status')">
          <el-select v-model="form.status" style="width: 100%">
            <el-option
              v-for="item in bankStatusOptions"
              :key="item.value"
              :label="getOptionLabel(t, item.code, item.value)"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editVisible = false"> {{ t('common.cancel') }} </el-button>
        <el-button
          type="primary"
          :loading="submitLoading"
          :disabled="!canSubmitCreate"
          @click="submitEdit"
        >
          {{ t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="statusVisible" :title="t('users.changeBankStatus')" width="420px">
      <el-form label-width="90px">
        <el-form-item :label="t('users.status')">
          <el-select v-model="statusForm.status" style="width: 100%">
            <el-option
              v-for="item in bankStatusOptions"
              :key="item.value"
              :label="getOptionLabel(t, item.code, item.value)"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="statusVisible = false"> {{ t('common.cancel') }} </el-button>
        <el-button type="primary" @click="submitStatus"> {{ t('common.confirm') }} </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="detailVisible" :title="t('users.bankDetail')" width="760px">
      <pre class="detail-pre">{{ JSON.stringify(detailData, null, 2) }}</pre>
    </el-dialog>
  </div>
</template>

<style scoped>
.option-tag {
  display: inline-flex;
  align-items: center;
  padding: 3px 10px;
  border-radius: 999px;
  font-size: 12px;
  line-height: 1.2;
  font-weight: 600;
  white-space: nowrap;
  background: #f3f4f6;
  color: #475467;
}

.option-tag--green {
  background: #dcfce7;
  color: #166534;
}

.option-tag--red {
  background: #fee2e2;
  color: #b91c1c;
}

.verify-row {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
}

.verify-tip {
  margin-top: 8px;
  font-size: 12px;
  line-height: 1.4;
}

.verify-tip--success {
  color: #16a34a;
}

.verify-tip--error {
  color: #dc2626;
}

.verify-tip--muted {
  color: #909399;
}
</style>
