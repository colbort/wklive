<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowDown } from '@element-plus/icons-vue'
import {
  memberUserService,
  tenantsService,
  type CreateMemberUserReq,
  type MemberUserDetail,
  type MemberUserItem,
  type UpdateMemberUserBaseReq,
} from '@/services'
import { formatDate } from '@/utils'

const loading = ref(false)
const submitLoading = ref(false)
const list = ref<MemberUserItem[]>([])
const detailVisible = ref(false)
const detail = ref<MemberUserDetail | null>(null)
const detailActiveTab = ref('identity')
const editVisible = ref(false)
const pwdVisible = ref(false)
const pwdMode = ref<'login' | 'pay'>('login')
const tenantChecking = ref(false)
const tenantChecked = ref(false)
const tenantExists = ref(false)
const tenantCheckName = ref('')
const tenantCheckCode = ref('')

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
  remark: '',
})

const pwdForm = reactive({
  userId: 0,
  tenantId: 0,
  password: '',
})

const isCreate = computed(() => !editForm.userId)
const canSubmitCreate = computed(() => !isCreate.value || (tenantChecked.value && tenantExists.value))

function checkCode(code: number) {
  return code === 0 || code === 200
}

function displayText(value: unknown) {
  if (value === null || value === undefined || value === '') return '-'
  return String(value)
}

function formatTimeValue(value?: number | null) {
  if (!value) return '-'
  return formatDate(value)
}

function getGenderLabel(value?: number) {
  const map: Record<number, string> = {
    0: '未知',
    1: '男',
    2: '女',
  }
  return map[value || 0] || String(value)
}

function getIdTypeLabel(value?: number) {
  const map: Record<number, string> = {
    0: '未提交',
    1: '身份证',
    2: '护照',
    3: '驾驶证',
  }
  return map[value || 0] || String(value)
}

function getKycLevelLabel(value?: number) {
  const map: Record<number, string> = {
    0: '未认证',
    1: '初级',
    2: '高级',
  }
  return map[value || 0] || String(value)
}

function getVerifyStatusLabel(value?: number) {
  const map: Record<number, string> = {
    0: '未提交',
    1: '审核中',
    2: '通过',
    3: '拒绝',
  }
  return map[value || 0] || String(value)
}

function getRiskLevelLabel(value?: number) {
  const map: Record<number, string> = {
    0: '正常',
    1: '关注',
    2: '高风险',
  }
  return map[value || 0] || String(value)
}

function getEnabledLabel(value?: number) {
  return Number(value) === 1 ? '是' : '否'
}

function getBankStatusLabel(value?: number) {
  const map: Record<number, string> = {
    1: '正常',
    2: '禁用',
  }
  return map[value || 0] || String(value)
}

function resetTenantCheck() {
  tenantChecked.value = false
  tenantExists.value = false
  tenantCheckName.value = ''
  tenantCheckCode.value = ''
}

async function verifyTenant() {
  const tenantId = Number(editForm.tenantId || 0)
  if (!tenantId) {
    resetTenantCheck()
    ElMessage.warning('请输入租户ID')
    return false
  }

  tenantChecking.value = true
  try {
    const res = await tenantsService.detail({ tenantId })
    if (!checkCode(res.code)) throw new Error(res.msg || '查询租户失败')

    const tenant = res.data
    tenantChecked.value = true
    tenantExists.value = Boolean(tenant)
    tenantCheckName.value = tenant?.tenantName || ''
    tenantCheckCode.value = tenant?.tenantCode || ''

    if (!tenant) {
      ElMessage.warning('租户不存在，请确认租户ID')
      return false
    }
    ElMessage.success(`已找到租户：${tenant.tenantName}`)
    return true
  } catch (error: any) {
    resetTenantCheck()
    ElMessage.error(error?.message || '查询租户失败')
    return false
  } finally {
    tenantChecking.value = false
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
    if (!checkCode(res.code)) throw new Error(res.msg || '加载失败')
    list.value = res.data || []
  } catch (error: any) {
    ElMessage.error(error?.message || '加载失败')
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

async function showDetail(row: MemberUserItem) {
  loading.value = true
  try {
    const res = await memberUserService.getDetail(row.id, row.tenantId)
    if (!checkCode(res.code)) throw new Error(res.msg || '加载详情失败')
    detail.value = (res.detail || res.data) as MemberUserDetail
    detailActiveTab.value = 'identity'
    detailVisible.value = true
  } catch (error: any) {
    ElMessage.error(error?.message || '加载详情失败')
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
    remark: '',
  })
  resetTenantCheck()
  editVisible.value = true
}

async function openEdit(row: MemberUserItem) {
  const tenantId = Number(row.tenantId || query.tenantId || 0)
  if (!tenantId) {
    ElMessage.warning('请先输入租户ID')
    return
  }
  const res = await memberUserService.getDetail(row.id, tenantId)
  if (!checkCode(res.code)) {
    ElMessage.error(res.msg || '加载详情失败')
    return
  }
  const data = (res.detail || res.data) as MemberUserDetail
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
    remark: data.base.remark,
  })
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
        referrerUserId: Number(editForm.referrerUserId || 0) || undefined,
        remark: editForm.remark || undefined,
      }
      const res = await memberUserService.create(payload)
      if (!checkCode(res.code)) throw new Error(res.msg || '创建失败')
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
        referrerUserId: Number(editForm.referrerUserId || 0) || undefined,
        remark: editForm.remark || undefined,
        phone: editForm.phone || undefined,
        email: editForm.email || undefined,
      }
      const res = await memberUserService.updateBase(Number(editForm.userId), payload)
      if (!checkCode(res.code)) throw new Error(res.msg || '更新失败')
      await memberUserService.updateStatus(Number(editForm.userId), {
        tenantId: Number(editForm.tenantId),
        status: Number(editForm.status),
      })
      await memberUserService.updateLevel(Number(editForm.userId), {
        tenantId: Number(editForm.tenantId),
        memberLevel: Number(editForm.memberLevel || 0),
      })
    }
    ElMessage.success('保存成功')
    editVisible.value = false
    fetchList()
  } catch (error: any) {
    ElMessage.error(error?.message || '保存失败')
  } finally {
    submitLoading.value = false
  }
}

function openPassword(row: MemberUserItem, mode: 'login' | 'pay') {
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
      const res = await memberUserService.resetLoginPassword(pwdForm.userId, pwdForm.tenantId, pwdForm.password)
      if (!checkCode(res.code)) throw new Error(res.msg || '重置失败')
    } else {
      const res = await memberUserService.resetPayPassword(pwdForm.userId, pwdForm.tenantId, pwdForm.password)
      if (!checkCode(res.code)) throw new Error(res.msg || '重置失败')
    }
    ElMessage.success('重置成功')
    pwdVisible.value = false
  } catch (error: any) {
    ElMessage.error(error?.message || '重置失败')
  } finally {
    submitLoading.value = false
  }
}

async function quickAction(row: MemberUserItem, action: 'unlock' | 'reset2fa' | 'delete') {
  try {
    if (action === 'delete') {
      await ElMessageBox.confirm(`确认删除用户 ${row.username} ?`, '提示', { type: 'warning' })
      const res = await memberUserService.delete(row.id, row.tenantId)
      if (!checkCode(res.code)) throw new Error(res.msg || '删除失败')
    }
    if (action === 'unlock') {
      const res = await memberUserService.unlock(row.id, row.tenantId)
      if (!checkCode(res.code)) throw new Error(res.msg || '解锁失败')
    }
    if (action === 'reset2fa') {
      const res = await memberUserService.reset2fa(row.id, row.tenantId)
      if (!checkCode(res.code)) throw new Error(res.msg || '重置失败')
    }
    ElMessage.success('操作成功')
    fetchList()
  } catch (error: any) {
    if (error === 'cancel') return
    ElMessage.error(error?.message || '操作失败')
  }
}

async function updateSimpleValue(row: MemberUserItem, field: 'status' | 'memberLevel' | 'riskLevel') {
  const tenantId = Number(row.tenantId || query.tenantId || 0)
  if (!tenantId) {
    ElMessage.warning('请先输入租户ID')
    return
  }
  const current = field === 'status' ? row.status : field === 'memberLevel' ? row.memberLevel : 0
  const titleMap = { status: '修改状态', memberLevel: '修改会员等级', riskLevel: '修改风控等级' }
  const { value } = await ElMessageBox.prompt(`请输入新的${titleMap[field]}`, titleMap[field], { inputValue: String(current) })
  try {
    if (field === 'status') {
      const res = await memberUserService.updateStatus(row.id, { tenantId, status: Number(value) })
      if (!checkCode(res.code)) throw new Error(res.msg || '更新失败')
    }
    if (field === 'memberLevel') {
      const res = await memberUserService.updateLevel(row.id, { tenantId, memberLevel: Number(value) })
      if (!checkCode(res.code)) throw new Error(res.msg || '更新失败')
    }
    if (field === 'riskLevel') {
      const res = await memberUserService.updateRiskLevel(row.id, { tenantId, riskLevel: Number(value) })
      if (!checkCode(res.code)) throw new Error(res.msg || '更新失败')
    }
    ElMessage.success('更新成功')
    fetchList()
  } catch (error: any) {
    ElMessage.error(error?.message || '更新失败')
  }
}

onMounted(fetchList)
</script>

<template>
  <div class="module-page">
    <div class="page-header">
      <h2>会员用户</h2>
      <div class="header-actions">
        <el-button @click="fetchList">
          刷新
        </el-button>
        <el-button type="primary" @click="openCreate">
          新增用户
        </el-button>
      </div>
    </div>

    <el-card shadow="never" class="query-card">
      <el-form :model="query" inline label-width="90px">
        <el-form-item label="租户ID">
          <el-input-number v-model="query.tenantId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item label="关键字">
          <el-input v-model="query.keyword" clearable />
        </el-form-item>
        <el-form-item label="用户名">
          <el-input v-model="query.username" clearable />
        </el-form-item>
        <el-form-item label="手机号">
          <el-input v-model="query.phone" clearable />
        </el-form-item>
        <el-form-item label="邮箱">
          <el-input v-model="query.email" clearable />
        </el-form-item>
        <el-form-item label="状态">
          <el-input-number v-model="query.status" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item label="认证状态">
          <el-input-number v-model="query.verifyStatus" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="fetchList">
            查询
          </el-button>
          <el-button @click="resetQuery">
            重置
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="never" class="table-card">
      <el-table v-loading="loading" :data="list" stripe>
        <el-table-column prop="id" label="用户ID" width="100" />
        <el-table-column prop="tenantId" label="租户ID" width="100" />
        <el-table-column
          prop="userNo"
          label="用户编号"
          min-width="150"
          show-overflow-tooltip
        />
        <el-table-column prop="username" label="用户名" min-width="140" />
        <el-table-column
          prop="nickname"
          label="昵称"
          min-width="140"
          show-overflow-tooltip
        />
        <el-table-column prop="registerType" label="注册方式" width="100" />
        <el-table-column prop="memberLevel" label="会员等级" width="100" />
        <el-table-column prop="status" label="状态" width="80" />
        <el-table-column prop="isGuest" label="游客" width="80" />
        <el-table-column prop="isRecharge" label="已充值" width="90" />
        <el-table-column
          prop="lastLoginIp"
          label="最后登录IP"
          min-width="150"
          show-overflow-tooltip
        />
        <el-table-column label="注册时间" min-width="170">
          <template #default="{ row }">
            {{ formatDate(row.registerTime) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-dropdown trigger="click">
              <el-button size="small">
                操作
                <el-icon class="el-icon--right">
                  <ArrowDown />
                </el-icon>
              </el-button>

              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item @click="showDetail(row)">
                    详情
                  </el-dropdown-item>
                  <el-dropdown-item @click="openEdit(row)">
                    编辑
                  </el-dropdown-item>
                  <el-dropdown-item @click="updateSimpleValue(row, 'status')">
                    修改状态
                  </el-dropdown-item>
                  <el-dropdown-item @click="updateSimpleValue(row, 'memberLevel')">
                    修改等级
                  </el-dropdown-item>
                  <el-dropdown-item @click="updateSimpleValue(row, 'riskLevel')">
                    修改风控
                  </el-dropdown-item>
                  <el-dropdown-item @click="openPassword(row, 'login')">
                    重置登录密码
                  </el-dropdown-item>
                  <el-dropdown-item @click="openPassword(row, 'pay')">
                    重置支付密码
                  </el-dropdown-item>
                  <el-dropdown-item @click="quickAction(row, 'unlock')">
                    解锁
                  </el-dropdown-item>
                  <el-dropdown-item @click="quickAction(row, 'reset2fa')">
                    重置2FA
                  </el-dropdown-item>
                  <el-dropdown-item @click="quickAction(row, 'delete')">
                    <span style="color: var(--el-color-danger)">
                      删除
                    </span>
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="editVisible" :title="isCreate ? '新增用户' : '编辑用户'" width="720px">
      <el-form label-width="110px">
        <el-form-item label="租户ID">
          <div class="tenant-check-row">
            <el-input-number
              v-model="editForm.tenantId"
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
              查询租户
            </el-button>
          </div>
          <div v-if="isCreate" class="tenant-check-tip">
            <span v-if="tenantChecked && tenantExists" class="tenant-check-tip tenant-check-tip--success">
              已验证租户：{{ tenantCheckName || editForm.tenantId }}
            </span>
            <span v-else-if="tenantChecked" class="tenant-check-tip tenant-check-tip--error">
              租户不存在，请重新输入
            </span>
            <span v-else class="tenant-check-tip tenant-check-tip--muted">
              输入租户ID后请先查询，验证通过才可提交
            </span>
          </div>
        </el-form-item>
        <el-form-item label="用户名">
          <el-input v-model="editForm.username" />
        </el-form-item>
        <el-form-item v-if="isCreate" label="密码">
          <el-input v-model="editForm.password" show-password />
        </el-form-item>
        <el-form-item label="昵称">
          <el-input v-model="editForm.nickname" />
        </el-form-item>
        <el-form-item label="手机号">
          <el-input v-model="editForm.phone" />
        </el-form-item>
        <el-form-item label="邮箱">
          <el-input v-model="editForm.email" />
        </el-form-item>
        <el-form-item label="头像">
          <el-input v-model="editForm.avatar" />
        </el-form-item>
        <el-form-item label="注册类型">
          <el-input-number v-model="editForm.registerType" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item label="状态">
          <el-input-number v-model="editForm.status" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item label="会员等级">
          <el-input-number v-model="editForm.memberLevel" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item label="语言">
          <el-input v-model="editForm.language" />
        </el-form-item>
        <el-form-item label="时区">
          <el-input v-model="editForm.timezone" />
        </el-form-item>
        <el-form-item label="邀请码">
          <el-input v-model="editForm.inviteCode" />
        </el-form-item>
        <el-form-item label="签名">
          <el-input v-model="editForm.signature" />
        </el-form-item>
        <el-form-item label="来源">
          <el-input v-model="editForm.source" />
        </el-form-item>
        <el-form-item label="推荐人ID">
          <el-input-number v-model="editForm.referrerUserId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="editForm.remark" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editVisible = false">
          取消
        </el-button>
        <el-button type="primary" :loading="submitLoading" :disabled="!canSubmitCreate" @click="submitEdit">
          确定
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="pwdVisible" :title="pwdMode === 'login' ? '重置登录密码' : '重置支付密码'" width="520px">
      <el-form label-width="100px">
        <el-form-item label="新密码">
          <el-input v-model="pwdForm.password" show-password />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="pwdVisible = false">
          取消
        </el-button>
        <el-button type="primary" :loading="submitLoading" @click="submitPassword">
          确定
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="detailVisible" title="用户详情" width="960px">
      <div v-if="detail" class="detail-sections">
        <el-tabs v-model="detailActiveTab">
          <el-tab-pane label="实名信息" name="identity">
            <el-card shadow="never">
              <el-descriptions :column="2" border>
                <el-descriptions-item label="手机号">
                  {{ displayText(detail.identity.phone) }}
                </el-descriptions-item>
                <el-descriptions-item label="邮箱">
                  {{ displayText(detail.identity.email) }}
                </el-descriptions-item>
                <el-descriptions-item label="真实姓名">
                  {{ displayText(detail.identity.realName) }}
                </el-descriptions-item>
                <el-descriptions-item label="性别">
                  {{ getGenderLabel(detail.identity.gender) }}
                </el-descriptions-item>
                <el-descriptions-item label="生日">
                  {{ formatTimeValue(detail.identity.birthday) }}
                </el-descriptions-item>
                <el-descriptions-item label="国家/地区">
                  {{ displayText(detail.identity.countryCode) }}
                </el-descriptions-item>
                <el-descriptions-item label="省/州">
                  {{ displayText(detail.identity.province) }}
                </el-descriptions-item>
                <el-descriptions-item label="城市">
                  {{ displayText(detail.identity.city) }}
                </el-descriptions-item>
                <el-descriptions-item label="地址" :span="2">
                  {{ displayText(detail.identity.address) }}
                </el-descriptions-item>
                <el-descriptions-item label="证件类型">
                  {{ getIdTypeLabel(detail.identity.idType) }}
                </el-descriptions-item>
                <el-descriptions-item label="证件号码">
                  {{ displayText(detail.identity.idNo) }}
                </el-descriptions-item>
                <el-descriptions-item label="KYC等级">
                  {{ getKycLevelLabel(detail.identity.kycLevel) }}
                </el-descriptions-item>
                <el-descriptions-item label="实名状态">
                  {{ getVerifyStatusLabel(detail.identity.verifyStatus) }}
                </el-descriptions-item>
                <el-descriptions-item label="提交时间">
                  {{ formatTimeValue(detail.identity.submitTime) }}
                </el-descriptions-item>
                <el-descriptions-item label="审核时间">
                  {{ formatTimeValue(detail.identity.verifyTime) }}
                </el-descriptions-item>
                <el-descriptions-item label="审核人">
                  {{ displayText(detail.identity.verifyBy) }}
                </el-descriptions-item>
                <el-descriptions-item label="驳回原因">
                  {{ displayText(detail.identity.rejectReason) }}
                </el-descriptions-item>
                <el-descriptions-item label="证件正面">
                  <a v-if="detail.identity.frontImage" :href="detail.identity.frontImage" target="_blank" rel="noreferrer">
                    查看图片
                  </a>
                  <span v-else>-</span>
                </el-descriptions-item>
                <el-descriptions-item label="证件反面">
                  <a v-if="detail.identity.backImage" :href="detail.identity.backImage" target="_blank" rel="noreferrer">
                    查看图片
                  </a>
                  <span v-else>-</span>
                </el-descriptions-item>
                <el-descriptions-item label="手持证件照" :span="2">
                  <a v-if="detail.identity.handheldImage" :href="detail.identity.handheldImage" target="_blank" rel="noreferrer">
                    查看图片
                  </a>
                  <span v-else>-</span>
                </el-descriptions-item>
              </el-descriptions>
            </el-card>
          </el-tab-pane>

          <el-tab-pane label="安全信息" name="security">
            <el-card shadow="never">
              <el-descriptions :column="2" border>
                <el-descriptions-item label="支付密码哈希">
                  {{ displayText(detail.security.payPasswordHash) }}
                </el-descriptions-item>
                <el-descriptions-item label="Google密钥">
                  {{ displayText(detail.security.googleSecret) }}
                </el-descriptions-item>
                <el-descriptions-item label="Google 2FA已启用">
                  {{ getEnabledLabel(detail.security.googleEnabled) }}
                </el-descriptions-item>
                <el-descriptions-item label="登录错误次数">
                  {{ displayText(detail.security.loginErrorCount) }}
                </el-descriptions-item>
                <el-descriptions-item label="支付密码错误次数">
                  {{ displayText(detail.security.payErrorCount) }}
                </el-descriptions-item>
                <el-descriptions-item label="锁定到期时间">
                  {{ formatTimeValue(detail.security.lockUntil) }}
                </el-descriptions-item>
                <el-descriptions-item label="风控等级">
                  {{ getRiskLevelLabel(detail.security.riskLevel) }}
                </el-descriptions-item>
                <el-descriptions-item label="创建时间">
                  {{ formatTimeValue(detail.security.createTimes) }}
                </el-descriptions-item>
                <el-descriptions-item label="更新时间">
                  {{ formatTimeValue(detail.security.updateTimes) }}
                </el-descriptions-item>
              </el-descriptions>
            </el-card>
          </el-tab-pane>

          <el-tab-pane label="银行卡" name="banks">
            <el-card shadow="never">
              <el-table v-if="detail.banks?.length" :data="detail.banks" stripe>
                <el-table-column prop="bankName" label="银行名称" min-width="140" />
                <el-table-column prop="bankCode" label="银行编码" min-width="120" />
                <el-table-column prop="accountName" label="开户名" min-width="140" />
                <el-table-column label="银行卡号" min-width="180" show-overflow-tooltip>
                  <template #default="{ row }">
                    {{ displayText(row.maskedAccountNo || row.accountNo) }}
                  </template>
                </el-table-column>
                <el-table-column prop="branchName" label="支行名称" min-width="160" show-overflow-tooltip />
                <el-table-column prop="countryCode" label="国家/地区" min-width="110" />
                <el-table-column label="默认卡" width="90">
                  <template #default="{ row }">
                    {{ row.isDefault ? '是' : '否' }}
                  </template>
                </el-table-column>
                <el-table-column label="状态" width="90">
                  <template #default="{ row }">
                    {{ getBankStatusLabel(row.status) }}
                  </template>
                </el-table-column>
                <el-table-column label="创建时间" min-width="170">
                  <template #default="{ row }">
                    {{ formatTimeValue(row.createTimes) }}
                  </template>
                </el-table-column>
              </el-table>
              <el-empty v-else description="暂无银行卡" />
            </el-card>
          </el-tab-pane>
        </el-tabs>
      </div>
    </el-dialog>
  </div>
</template>

<style scoped>
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
