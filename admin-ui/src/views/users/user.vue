<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowDown } from '@element-plus/icons-vue'
import {
  memberUserService,
  tenantsService,
  type OptionGroup,
  type OptionItem,
  type CreateMemberUserReq,
  type UserDetail,
  type UserItem,
  type UpdateMemberUserBaseReq,
} from '@/services'
import { formatDate } from '@/utils'
import { findOptionGroup, getOptionLabel, getOptionValueLabel } from '@/utils/options'

const { t } = useI18n()
const loading = ref(false)
const submitLoading = ref(false)
const list = ref<UserItem[]>([])
const detailVisible = ref(false)
const detail = ref<UserDetail | null>(null)
const detailActiveTab = ref('identity')
const editVisible = ref(false)
const pwdVisible = ref(false)
const pwdMode = ref<'login' | 'pay'>('login')
const tenantChecking = ref(false)
const tenantChecked = ref(false)
const tenantExists = ref(false)
const tenantCheckName = ref('')
const tenantCheckCode = ref('')
const optionGroups = ref<OptionGroup[]>([])
const createOptions = reactive({
  registerTypes: [] as OptionItem[],
  statuses: [] as OptionItem[],
})
const referrerChecking = ref(false)
const referrerChecked = ref(false)
const referrerExists = ref(false)
const referrerDisplay = ref('')

const query = reactive({
  tenantId: undefined as number | undefined,
  keyword: '',
  username: '',
  phone: '',
  email: '',
  status: undefined as number | undefined,
  verifyStatus: undefined as number | undefined,
  limit: 100,
})

const editForm = reactive<any>({
  userId: 0,
  tenantId: 0,
  username: '',
  nickname: '',
  avatar: '',
  phone: '',
  email: '',
  password: '',
  registerType: 1,
  status: 1,
  memberLevel: 0,
  language: '',
  timezone: '',
  inviteCode: '',
  signature: '',
  source: '',
  referrerUserId: 0,
  referrerInviteCode: '',
  remark: '',
})

const pwdForm = reactive({
  userId: 0,
  tenantId: 0,
  password: '',
})

const isCreate = computed(() => !editForm.userId)
const canSubmitCreate = computed(() => {
  if (!isCreate.value) return true
  if (!tenantChecked.value || !tenantExists.value) return false
  if (!String(editForm.referrerInviteCode || '').trim()) return true
  return referrerChecked.value && referrerExists.value
})

function checkCode(code: number) {
  return code === 0 || code === 200
}

function displayText(value: unknown) {
  if (value === null || value === undefined || value === '') return '-'
  return String(value)
}

function getGenderLabel(value?: number) {
  return getOptionValueLabel(optionGroups.value, 'gender', value, t)
}

function getIdTypeLabel(value?: number) {
  return getOptionValueLabel(optionGroups.value, 'idType', value, t)
}

function getKycLevelLabel(value?: number) {
  return getOptionValueLabel(optionGroups.value, 'kycLevel', value, t)
}

function getVerifyStatusLabel(value?: number) {
  return getOptionValueLabel(optionGroups.value, 'verifyStatus', value, t)
}

function getRiskLevelLabel(value?: number) {
  return getOptionValueLabel(optionGroups.value, 'riskLevel', value, t)
}

function getEnabledLabel(value?: number) {
  return Number(value) === 1 ? t('users.yes') : t('users.no')
}

function getOptionTagClass(groupKey: string, value?: number) {
  const normalizedValue = Number(value ?? 0)

  if (groupKey === 'registerType') {
    const registerTypeMap: Record<number, string> = {
      1: 'option-tag option-tag--blue',
      2: 'option-tag option-tag--green',
      3: 'option-tag option-tag--orange',
      4: 'option-tag option-tag--slate',
    }
    return registerTypeMap[normalizedValue] || 'option-tag'
  }

  if (groupKey === 'userStatus') {
    const userStatusMap: Record<number, string> = {
      1: 'option-tag option-tag--green',
      2: 'option-tag option-tag--slate',
      3: 'option-tag option-tag--orange',
      4: 'option-tag option-tag--red',
    }
    return userStatusMap[normalizedValue] || 'option-tag'
  }

  return 'option-tag'
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

function resetTenantCheck() {
  tenantChecked.value = false
  tenantExists.value = false
  tenantCheckName.value = ''
  tenantCheckCode.value = ''
  resetReferrerCheck()
}

function resetReferrerCheck() {
  referrerChecked.value = false
  referrerExists.value = false
  referrerDisplay.value = ''
}

async function fetchCreateOptions() {
  try {
    const res = await memberUserService.getOptions()
    if (!checkCode(res.code)) throw new Error(res.msg || t('users.loadOptionsFailed'))
    const groups = res.data || []
    optionGroups.value = groups
    const findGroupOptions = (key: string) =>
      (groups.find((item: OptionGroup) => item.key === key)?.options || []) as OptionItem[]

    Object.assign(createOptions, {
      registerTypes: findGroupOptions('registerType'),
      statuses: findGroupOptions('userStatus'),
    })
  } catch (error: unknown) {
    ElMessage.error(error instanceof Error ? error.message : t('users.loadOptionsFailed'))
  }
}

async function verifyTenant() {
  const tenantId = Number(editForm.tenantId || 0)
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
    tenantCheckCode.value = tenant?.tenantCode || ''
    resetReferrerCheck()

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

async function verifyReferrer() {
  const referrerInviteCode = String(editForm.referrerInviteCode || '').trim()
  if (!referrerInviteCode) {
    resetReferrerCheck()
    return true
  }

  referrerChecking.value = true
  try {
    const res = await memberUserService.checkReferrer(referrerInviteCode)
    if (!checkCode(res.code) || !res.exists) {
      throw new Error(res.msg || t('users.referrerNotFound'))
    }
    const info = res.data
    referrerChecked.value = true
    referrerExists.value = true
    referrerDisplay.value = info?.nickname
      ? `${info.username} (${info.nickname})`
      : info?.username || referrerInviteCode
    ElMessage.success(t('users.referrerFound', { name: referrerDisplay.value }))
    return true
  } catch (error: unknown) {
    resetReferrerCheck()
    ElMessage.error(error instanceof Error ? error.message : t('users.referrerNotFound'))
    return false
  } finally {
    referrerChecking.value = false
  }
}

async function fetchList() {
  loading.value = true
  try {
    const res = await memberUserService.getList({
      ...query,
      tenantId: query.tenantId || undefined,
      status: query.status,
      verifyStatus: query.verifyStatus,
    })
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
    keyword: '',
    username: '',
    phone: '',
    email: '',
    status: undefined,
    verifyStatus: undefined,
    limit: 100,
  })
  fetchList()
}

async function showDetail(row: UserItem) {
  loading.value = true
  try {
    const res = await memberUserService.getDetail(row.id, row.tenantId)
    if (!checkCode(res.code)) throw new Error(res.msg || t('users.loadDetailFailed'))
    detail.value = (res.detail || res.data) as UserDetail
    detailActiveTab.value = 'identity'
    detailVisible.value = true
  } catch (error: unknown) {
    ElMessage.error(error instanceof Error ? error.message : t('users.loadDetailFailed'))
  } finally {
    loading.value = false
  }
}

function openCreate() {
  Object.assign(editForm, {
    userId: 0,
    tenantId: Number(query.tenantId || 0),
    username: '',
    nickname: '',
    avatar: '',
    phone: '',
    email: '',
    password: '',
    registerType: 1,
    status: 1,
    memberLevel: 0,
    language: '',
    timezone: '',
    inviteCode: '',
    signature: '',
    source: '',
    referrerUserId: 0,
    referrerInviteCode: '',
    remark: '',
  })
  resetTenantCheck()
  resetReferrerCheck()
  editVisible.value = true
}

async function openEdit(row: UserItem) {
  const tenantId = Number(row.tenantId || query.tenantId || 0)
  if (!tenantId) {
    ElMessage.warning(t('users.queryTenantPrompt'))
    return
  }
  const res = await memberUserService.getDetail(row.id, tenantId)
  if (!checkCode(res.code)) {
    ElMessage.error(res.msg || t('users.loadDetailFailed'))
    return
  }
  const data = (res.detail || res.data) as UserDetail
  Object.assign(editForm, {
    userId: data.base.id,
    tenantId: data.base.tenantId,
    username: data.base.username,
    nickname: data.base.nickname,
    avatar: data.base.avatar,
    phone: data.identity.phone,
    email: data.identity.email,
    password: '',
    registerType: data.base.registerType,
    status: data.base.status,
    memberLevel: data.base.memberLevel,
    language: data.base.language,
    timezone: data.base.timezone,
    inviteCode: data.base.inviteCode,
    signature: data.base.signature,
    source: data.base.source,
    referrerUserId: data.base.referrerUserId,
    referrerInviteCode: '',
    remark: data.base.remark,
  })
  resetReferrerCheck()
  if (Number(data.base.referrerUserId || 0) > 0) {
    try {
      const referrerRes = await memberUserService.getDetail(
        Number(data.base.referrerUserId),
        data.base.tenantId,
      )
      if (checkCode(referrerRes.code)) {
        const referrerDetail = (referrerRes.detail || referrerRes.data) as UserDetail
        editForm.referrerInviteCode = referrerDetail.base.inviteCode || ''
        referrerChecked.value = true
        referrerExists.value = true
        referrerDisplay.value = referrerDetail.base.nickname
          ? `${referrerDetail.base.username} (${referrerDetail.base.nickname})`
          : referrerDetail.base.username
      }
    } catch {
      resetReferrerCheck()
    }
  }
  editVisible.value = true
}

async function submitEdit() {
  submitLoading.value = true
  try {
    if (isCreate.value) {
      if (!tenantChecked.value || !tenantExists.value) {
        const verified = await verifyTenant()
        if (!verified) return
      }
    }
    const referrerOk = await verifyReferrer()
    if (!referrerOk) return
    if (isCreate.value) {
      const payload: CreateMemberUserReq = {
        tenantCode: tenantCheckCode.value,
        username: editForm.username,
        nickname: editForm.nickname || undefined,
        avatar: editForm.avatar || undefined,
        phone: editForm.phone || undefined,
        email: editForm.email || undefined,
        password: editForm.password,
        registerType: Number(editForm.registerType),
        status: Number(editForm.status),
        memberLevel: Number(editForm.memberLevel || 0),
        language: editForm.language || undefined,
        timezone: editForm.timezone || undefined,
        inviteCode: editForm.inviteCode || undefined,
        signature: editForm.signature || undefined,
        source: editForm.source || undefined,
        referrerUserId: undefined,
        referrerInviteCode: editForm.referrerInviteCode || undefined,
        remark: editForm.remark || undefined,
      }
      const res = await memberUserService.create(payload)
      if (!checkCode(res.code)) throw new Error(res.msg || t('users.createFailed'))
    } else {
      const payload: UpdateMemberUserBaseReq = {
        tenantId: Number(editForm.tenantId),
        username: editForm.username || undefined,
        nickname: editForm.nickname || undefined,
        avatar: editForm.avatar || undefined,
        language: editForm.language || undefined,
        timezone: editForm.timezone || undefined,
        signature: editForm.signature || undefined,
        source: editForm.source || undefined,
        referrerUserId: undefined,
        referrerInviteCode: editForm.referrerInviteCode || undefined,
        remark: editForm.remark || undefined,
        phone: editForm.phone || undefined,
        email: editForm.email || undefined,
      }
      const res = await memberUserService.updateBase(Number(editForm.userId), payload)
      if (!checkCode(res.code)) throw new Error(res.msg || t('users.updateFailed'))
      await memberUserService.updateStatus(Number(editForm.userId), {
        tenantId: Number(editForm.tenantId),
        status: Number(editForm.status),
      })
      await memberUserService.updateLevel(Number(editForm.userId), {
        tenantId: Number(editForm.tenantId),
        memberLevel: Number(editForm.memberLevel || 0),
      })
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

function openPassword(row: UserItem, mode: 'login' | 'pay') {
  pwdMode.value = mode
  pwdForm.userId = row.id
  pwdForm.tenantId = Number(row.tenantId || query.tenantId || 0)
  pwdForm.password = ''
  pwdVisible.value = true
}

async function submitPassword() {
  submitLoading.value = true
  try {
    if (pwdMode.value === 'login') {
      const res = await memberUserService.resetLoginPassword(
        pwdForm.userId,
        pwdForm.tenantId,
        pwdForm.password,
      )
      if (!checkCode(res.code)) throw new Error(res.msg || t('users.resetFailed'))
    } else {
      const res = await memberUserService.resetPayPassword(
        pwdForm.userId,
        pwdForm.tenantId,
        pwdForm.password,
      )
      if (!checkCode(res.code)) throw new Error(res.msg || t('users.resetFailed'))
    }
    ElMessage.success(t('users.resetSuccess'))
    pwdVisible.value = false
  } catch (error: unknown) {
    ElMessage.error(error instanceof Error ? error.message : t('users.resetFailed'))
  } finally {
    submitLoading.value = false
  }
}

async function quickAction(row: UserItem, action: 'unlock' | 'reset2fa' | 'delete') {
  try {
    if (action === 'delete') {
      await ElMessageBox.confirm(
        t('users.deleteUserConfirm', { name: row.username }),
        t('common.warning'),
        { type: 'warning' },
      )
      const res = await memberUserService.delete(row.id, row.tenantId)
      if (!checkCode(res.code)) throw new Error(res.msg || t('users.deleteFailed'))
    }
    if (action === 'unlock') {
      const res = await memberUserService.unlock(row.id, row.tenantId)
      if (!checkCode(res.code)) throw new Error(res.msg || t('users.operationFailed'))
    }
    if (action === 'reset2fa') {
      const res = await memberUserService.reset2fa(row.id, row.tenantId)
      if (!checkCode(res.code)) throw new Error(res.msg || t('users.resetFailed'))
    }
    ElMessage.success(t('common.operationSuccess'))
    fetchList()
  } catch (error: unknown) {
    if (error === 'cancel') return
    ElMessage.error(error instanceof Error ? error.message : t('users.operationFailed'))
  }
}

async function updateSimpleValue(row: UserItem, field: 'status' | 'memberLevel' | 'riskLevel') {
  const tenantId = Number(row.tenantId || query.tenantId || 0)
  if (!tenantId) {
    ElMessage.warning(t('users.queryTenantPrompt'))
    return
  }
  const current = field === 'status' ? row.status : field === 'memberLevel' ? row.memberLevel : 0
  const titleMap = {
    status: t('users.modifyStatus'),
    memberLevel: t('users.modifyMemberLevel'),
    riskLevel: t('users.modifyRiskLevel'),
  }
  const { value } = await ElMessageBox.prompt(
    t('users.inputNewValue', { title: titleMap[field] }),
    titleMap[field],
    {
    inputValue: String(current),
    },
  )
  try {
    if (field === 'status') {
      const res = await memberUserService.updateStatus(row.id, { tenantId, status: Number(value) })
      if (!checkCode(res.code)) throw new Error(res.msg || t('users.updateFailed'))
    }
    if (field === 'memberLevel') {
      const res = await memberUserService.updateLevel(row.id, {
        tenantId,
        memberLevel: Number(value),
      })
      if (!checkCode(res.code)) throw new Error(res.msg || t('users.updateFailed'))
    }
    if (field === 'riskLevel') {
      const res = await memberUserService.updateRiskLevel(row.id, {
        tenantId,
        riskLevel: Number(value),
      })
      if (!checkCode(res.code)) throw new Error(res.msg || t('users.updateFailed'))
    }
    ElMessage.success(t('users.updateSuccess'))
    fetchList()
  } catch (error: unknown) {
    ElMessage.error(error instanceof Error ? error.message : t('users.updateFailed'))
  }
}

onMounted(fetchList)
onMounted(fetchCreateOptions)
</script>

<template>
  <div class="module-page">
    <div class="page-header">
      <h2>{{ t('users.memberUsers') }}</h2>
      <div class="header-actions">
        <el-button @click="fetchList"> {{ t('common.refresh') }} </el-button>
        <el-button type="primary" @click="openCreate"> {{ t('users.addUser') }} </el-button>
      </div>
    </div>

    <el-card shadow="never" class="query-card">
      <el-form :model="query" inline label-width="90px">
        <el-form-item :label="t('common.tenantId')">
          <el-input-number v-model="query.tenantId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('common.keyword')">
          <el-input v-model="query.keyword" clearable />
        </el-form-item>
        <el-form-item :label="t('users.username')">
          <el-input v-model="query.username" clearable />
        </el-form-item>
        <el-form-item :label="t('users.phone')">
          <el-input v-model="query.phone" clearable />
        </el-form-item>
        <el-form-item :label="t('users.email')">
          <el-input v-model="query.email" clearable />
        </el-form-item>
        <el-form-item :label="t('users.status')">
          <el-select v-model="query.status" clearable style="width: 140px">
            <el-option
              v-for="item in createOptions.statuses"
              :key="item.value"
              :label="getOptionLabel(t, item.code, item.value)"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('users.verifyStatus')">
          <el-select v-model="query.verifyStatus" clearable style="width: 140px">
            <el-option
              v-for="item in findOptionGroup(optionGroups, 'verifyStatus')"
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
        <el-table-column prop="id" :label="t('users.userId')" width="100" />
        <el-table-column prop="tenantId" :label="t('common.tenantId')" width="100" />
        <el-table-column prop="userNo" :label="t('users.userNo')" min-width="150" show-overflow-tooltip />
        <el-table-column prop="username" :label="t('users.username')" min-width="140" />
        <el-table-column prop="nickname" :label="t('users.nickname')" min-width="140" show-overflow-tooltip />
        <el-table-column :label="t('users.registerType')" width="120">
          <template #default="{ row }">
            <span :class="getOptionTagClass('registerType', row.registerType)">
              {{ getOptionValueLabel(optionGroups, 'registerType', row.registerType, t) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="memberLevel" :label="t('users.memberLevel')" width="100" />
        <el-table-column :label="t('users.status')" width="100">
          <template #default="{ row }">
            <span :class="getOptionTagClass('userStatus', row.status)">
              {{ getOptionValueLabel(optionGroups, 'userStatus', row.status, t) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column :label="t('users.isGuest')" width="80">
          <template #default="{ row }">
            <span :class="getBooleanTagClass(row.isGuest)">
              {{ getEnabledLabel(row.isGuest) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column :label="t('users.isRecharge')" width="90">
          <template #default="{ row }">
            <span :class="getBooleanTagClass(row.isRecharge)">
              {{ getEnabledLabel(row.isRecharge) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column
          prop="lastLoginIp"
          :label="t('users.lastLoginIp')"
          min-width="150"
          show-overflow-tooltip
        />
        <el-table-column :label="t('users.registerTime')" min-width="170">
          <template #default="{ row }">
            {{ formatDate(row.registerTime) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('common.actions')" width="120" fixed="right">
          <template #default="{ row }">
            <el-dropdown trigger="click">
              <el-button size="small">
                {{ t('common.actions') }}
                <el-icon class="el-icon--right">
                  <ArrowDown />
                </el-icon>
              </el-button>

              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item @click="showDetail(row)"> {{ t('common.detail') }} </el-dropdown-item>
                  <el-dropdown-item @click="openEdit(row)"> {{ t('common.edit') }} </el-dropdown-item>
                  <el-dropdown-item @click="updateSimpleValue(row, 'status')">
                    {{ t('users.modifyStatus') }}
                  </el-dropdown-item>
                  <el-dropdown-item @click="updateSimpleValue(row, 'memberLevel')">
                    {{ t('users.modifyMemberLevel') }}
                  </el-dropdown-item>
                  <el-dropdown-item @click="updateSimpleValue(row, 'riskLevel')">
                    {{ t('users.modifyRiskLevel') }}
                  </el-dropdown-item>
                  <el-dropdown-item @click="openPassword(row, 'login')">
                    {{ t('users.resetLoginPassword') }}
                  </el-dropdown-item>
                  <el-dropdown-item @click="openPassword(row, 'pay')">
                    {{ t('users.resetPayPassword') }}
                  </el-dropdown-item>
                  <el-dropdown-item @click="quickAction(row, 'unlock')"> {{ t('users.unlock') }} </el-dropdown-item>
                  <el-dropdown-item @click="quickAction(row, 'reset2fa')">
                    {{ t('users.reset2fa') }}
                  </el-dropdown-item>
                  <el-dropdown-item @click="quickAction(row, 'delete')">
                    <span style="color: var(--el-color-danger)">{{ t('common.delete') }}</span>
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="editVisible" :title="isCreate ? t('users.addUser') : t('users.editUser')" width="720px">
      <el-form label-width="110px" class="edit-form-grid">
        <el-form-item :label="t('common.tenantId')">
          <div class="tenant-check-row">
            <el-input-number
              v-model="editForm.tenantId"
              class="form-control"
              :min="0"
              :precision="0"
              @change="resetTenantCheck"
            />
            <el-button
              v-if="isCreate"
              type="primary"
              plain
              :loading="tenantChecking"
              @click="verifyTenant"
            >
              {{ t('users.queryTenant') }}
            </el-button>
          </div>
          <div v-if="isCreate" class="tenant-check-tip">
            <span
              v-if="tenantChecked && tenantExists"
              class="tenant-check-tip tenant-check-tip--success"
            >
              {{ t('users.tenantVerified', { name: tenantCheckName || editForm.tenantId }) }}
            </span>
            <span v-else-if="tenantChecked" class="tenant-check-tip tenant-check-tip--error">
              {{ t('users.tenantMissingRetry') }}
            </span>
            <span v-else class="tenant-check-tip tenant-check-tip--muted">
              {{ t('users.tenantQueryRequired') }}
            </span>
          </div>
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item :label="t('users.username')">
              <el-input v-model="editForm.username" />
            </el-form-item>
          </el-col>
          <el-col v-if="isCreate" :span="12">
            <el-form-item :label="t('users.password')">
              <el-input v-model="editForm.password" show-password />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('users.nickname')">
              <el-input v-model="editForm.nickname" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('users.phone')">
              <el-input v-model="editForm.phone" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('users.email')">
              <el-input v-model="editForm.email" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('users.avatar')">
              <el-input v-model="editForm.avatar" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('users.registerType')">
              <el-select v-model="editForm.registerType" style="width: 100%">
                <el-option
                  v-for="item in createOptions.registerTypes"
                  :key="item.value"
                  :label="getOptionLabel(t, item.code, item.value)"
                  :value="item.value"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('users.status')">
              <el-select v-model="editForm.status" style="width: 100%">
                <el-option
                  v-for="item in createOptions.statuses"
                  :key="item.value"
                  :label="getOptionLabel(t, item.code, item.value)"
                  :value="item.value"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('users.memberLevel')">
              <el-input-number
                v-model="editForm.memberLevel"
                class="form-control"
                :min="0"
                :precision="0"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('users.language')">
              <el-input v-model="editForm.language" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('users.timezone')">
              <el-input v-model="editForm.timezone" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('users.inviteCode')">
              <el-input v-model="editForm.inviteCode" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('users.signature')">
              <el-input v-model="editForm.signature" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('users.source')">
              <el-input v-model="editForm.source" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item :label="t('users.referrerInviteCode')">
          <div class="tenant-check-row">
            <el-input v-model="editForm.referrerInviteCode" @change="resetReferrerCheck" />
            <el-button plain :loading="referrerChecking" @click="verifyReferrer">
              {{ t('users.verifyReferrer') }}
            </el-button>
          </div>
          <div class="tenant-check-tip">
            <span
              v-if="referrerChecked && referrerExists"
              class="tenant-check-tip tenant-check-tip--success"
            >
              {{ t('users.referrerExists', { name: referrerDisplay }) }}
            </span>
            <span v-else-if="referrerChecked" class="tenant-check-tip tenant-check-tip--error">
              {{ t('users.referrerNotFoundRetry') }}
            </span>
            <span v-else class="tenant-check-tip tenant-check-tip--muted">
              {{ t('users.referrerOptionalTip') }}
            </span>
          </div>
        </el-form-item>
        <el-form-item :label="t('common.remark')">
          <el-input v-model="editForm.remark" type="textarea" :rows="3" />
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

    <el-dialog
      v-model="pwdVisible"
      :title="pwdMode === 'login' ? t('users.resetLoginPassword') : t('users.resetPayPassword')"
      width="520px"
    >
      <el-form label-width="100px">
        <el-form-item :label="t('users.newPassword')">
          <el-input v-model="pwdForm.password" show-password />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="pwdVisible = false"> {{ t('common.cancel') }} </el-button>
        <el-button type="primary" :loading="submitLoading" @click="submitPassword">
          {{ t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="detailVisible" :title="t('users.userDetail')" width="960px">
      <div v-if="detail" class="detail-sections">
        <el-tabs v-model="detailActiveTab">
          <el-tab-pane :label="t('users.identityInfo')" name="identity">
            <el-card shadow="never">
              <el-descriptions :column="2" border>
                <el-descriptions-item :label="t('users.phone')">
                  {{ displayText(detail.identity.phone) }}
                </el-descriptions-item>
                <el-descriptions-item :label="t('users.email')">
                  {{ displayText(detail.identity.email) }}
                </el-descriptions-item>
                <el-descriptions-item :label="t('users.actualName')">
                  {{ displayText(detail.identity.realName) }}
                </el-descriptions-item>
                <el-descriptions-item :label="t('users.gender')">
                  {{ getGenderLabel(detail.identity.gender) }}
                </el-descriptions-item>
                <el-descriptions-item :label="t('users.birthday')">
                  {{ formatDate(detail.identity.birthday) }}
                </el-descriptions-item>
                <el-descriptions-item :label="t('users.region')">
                  {{ displayText(detail.identity.countryCode) }}
                </el-descriptions-item>
                <el-descriptions-item :label="t('users.province')">
                  {{ displayText(detail.identity.province) }}
                </el-descriptions-item>
                <el-descriptions-item :label="t('users.city')">
                  {{ displayText(detail.identity.city) }}
                </el-descriptions-item>
                <el-descriptions-item :label="t('users.address')" :span="2">
                  {{ displayText(detail.identity.address) }}
                </el-descriptions-item>
                <el-descriptions-item :label="t('users.idType')">
                  {{ getIdTypeLabel(detail.identity.idType) }}
                </el-descriptions-item>
                <el-descriptions-item :label="t('users.idNo')">
                  {{ displayText(detail.identity.idNo) }}
                </el-descriptions-item>
                <el-descriptions-item :label="t('users.kycLevel')">
                  {{ getKycLevelLabel(detail.identity.kycLevel) }}
                </el-descriptions-item>
                <el-descriptions-item :label="t('users.identityStatus')">
                  {{ getVerifyStatusLabel(detail.identity.verifyStatus) }}
                </el-descriptions-item>
                <el-descriptions-item :label="t('users.submitTime')">
                  {{ formatDate(detail.identity.submitTime) }}
                </el-descriptions-item>
                <el-descriptions-item :label="t('users.reviewTime')">
                  {{ formatDate(detail.identity.verifyTime) }}
                </el-descriptions-item>
                <el-descriptions-item :label="t('users.reviewer')">
                  {{ displayText(detail.identity.verifyBy) }}
                </el-descriptions-item>
                <el-descriptions-item :label="t('users.rejectReason')">
                  {{ displayText(detail.identity.rejectReason) }}
                </el-descriptions-item>
                <el-descriptions-item :label="t('users.frontImage')">
                  <a
                    v-if="detail.identity.frontImage"
                    :href="detail.identity.frontImage"
                    target="_blank"
                    rel="noreferrer"
                  >
                    {{ t('users.viewImage') }}
                  </a>
                  <span v-else>-</span>
                </el-descriptions-item>
                <el-descriptions-item :label="t('users.backImage')">
                  <a
                    v-if="detail.identity.backImage"
                    :href="detail.identity.backImage"
                    target="_blank"
                    rel="noreferrer"
                  >
                    {{ t('users.viewImage') }}
                  </a>
                  <span v-else>-</span>
                </el-descriptions-item>
                <el-descriptions-item :label="t('users.handheldImage')" :span="2">
                  <a
                    v-if="detail.identity.handheldImage"
                    :href="detail.identity.handheldImage"
                    target="_blank"
                    rel="noreferrer"
                  >
                    {{ t('users.viewImage') }}
                  </a>
                  <span v-else>-</span>
                </el-descriptions-item>
              </el-descriptions>
            </el-card>
          </el-tab-pane>

          <el-tab-pane :label="t('users.securityInfo')" name="security">
            <el-card shadow="never">
              <el-descriptions :column="2" border>
                <el-descriptions-item :label="t('users.payPasswordHash')">
                  {{ displayText(detail.security.payPasswordHash) }}
                </el-descriptions-item>
                <el-descriptions-item :label="t('users.googleSecret')">
                  {{ displayText(detail.security.googleSecret) }}
                </el-descriptions-item>
                <el-descriptions-item :label="t('users.google2faEnabled')">
                  {{ getEnabledLabel(detail.security.googleEnabled) }}
                </el-descriptions-item>
                <el-descriptions-item :label="t('users.loginErrorCount')">
                  {{ displayText(detail.security.loginErrorCount) }}
                </el-descriptions-item>
                <el-descriptions-item :label="t('users.payErrorCount')">
                  {{ displayText(detail.security.payErrorCount) }}
                </el-descriptions-item>
                <el-descriptions-item :label="t('users.lockUntil')">
                  {{ formatDate(detail.security.lockUntil) }}
                </el-descriptions-item>
                <el-descriptions-item :label="t('users.riskLevel')">
                  {{ getRiskLevelLabel(detail.security.riskLevel) }}
                </el-descriptions-item>
                <el-descriptions-item :label="t('common.createTimes')">
                  {{ formatDate(detail.security.createTimes) }}
                </el-descriptions-item>
                <el-descriptions-item :label="t('common.updateTimes')">
                  {{ formatDate(detail.security.updateTimes) }}
                </el-descriptions-item>
              </el-descriptions>
            </el-card>
          </el-tab-pane>

          <el-tab-pane :label="t('users.bankCards')" name="banks">
            <el-card shadow="never">
              <el-table v-if="detail.banks?.length" :data="detail.banks" stripe>
                <el-table-column prop="bankName" :label="t('users.bankNameFull')" min-width="140" />
                <el-table-column prop="bankCode" :label="t('users.bankCode')" min-width="120" />
                <el-table-column prop="accountName" :label="t('users.accountName')" min-width="140" />
                <el-table-column :label="t('users.accountNo')" min-width="180" show-overflow-tooltip>
                  <template #default="{ row }">
                    {{ displayText(row.maskedAccountNo || row.accountNo) }}
                  </template>
                </el-table-column>
                <el-table-column
                  prop="branchName"
                  :label="t('users.branchNameFull')"
                  min-width="160"
                  show-overflow-tooltip
                />
                <el-table-column prop="countryCode" :label="t('users.region')" min-width="110" />
                <el-table-column :label="t('users.defaultBank')" width="90">
                  <template #default="{ row }">
                    <span :class="getBooleanTagClass(row.isDefault)">
                      {{ getEnabledLabel(row.isDefault) }}
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
              </el-table>
              <el-empty v-else :description="t('users.noBanks')" />
            </el-card>
          </el-tab-pane>
        </el-tabs>
      </div>
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

.option-tag--blue {
  background: #e0f2fe;
  color: #0369a1;
}

.option-tag--green {
  background: #dcfce7;
  color: #166534;
}

.option-tag--orange {
  background: #ffedd5;
  color: #c2410c;
}

.option-tag--red {
  background: #fee2e2;
  color: #b91c1c;
}

.option-tag--slate {
  background: #e5e7eb;
  color: #475467;
}

.edit-form-grid :deep(.el-form-item) {
  margin-bottom: 18px;
}

.form-control {
  width: 100%;
}

.tenant-check-row {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
}

.tenant-check-tip {
  margin-top: 8px;
  font-size: 12px;
  line-height: 1.4;
}

.tenant-check-tip--success {
  color: #16a34a;
}

.tenant-check-tip--error {
  color: #dc2626;
}

.tenant-check-tip--muted {
  color: #909399;
}

.detail-sections {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.detail-section-title {
  font-weight: 600;
}
</style>
