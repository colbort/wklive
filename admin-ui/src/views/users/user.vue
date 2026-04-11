<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { memberUserService, type CreateMemberUserReq, type MemberUserDetail, type MemberUserItem, type UpdateMemberUserBaseReq } from '@/services'
import { formatDate } from '@/utils'

const loading = ref(false)
const submitLoading = ref(false)
const list = ref<MemberUserItem[]>([])
const detailVisible = ref(false)
const detail = ref<MemberUserDetail | null>(null)
const editVisible = ref(false)
const pwdVisible = ref(false)
const pwdMode = ref<'login' | 'pay'>('login')

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

function checkCode(code: number) {
  return code === 0 || code === 200
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
  const tenantId = Number(query.tenantId || 0)
  if (!tenantId) {
    ElMessage.warning('请先输入租户ID')
    return
  }
  loading.value = true
  try {
    const res = await memberUserService.getDetail(row.userId, tenantId)
    if (!checkCode(res.code)) throw new Error(res.msg || '加载详情失败')
    detail.value = (res.detail || res.data) as MemberUserDetail
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
  editVisible.value = true
}

async function openEdit(row: MemberUserItem) {
  const tenantId = Number(query.tenantId || 0)
  if (!tenantId) {
    ElMessage.warning('请先输入租户ID')
    return
  }
  const res = await memberUserService.getDetail(row.userId, tenantId)
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
      const payload: CreateMemberUserReq = {
        tenantId: Number(editForm.tenantId),
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
  pwdForm.userId = row.userId
  pwdForm.tenantId = Number(query.tenantId || 0)
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
  const tenantId = Number(query.tenantId || 0)
  if (!tenantId) {
    ElMessage.warning('请先输入租户ID')
    return
  }
  try {
    if (action === 'delete') {
      await ElMessageBox.confirm(`确认删除用户 ${row.username} ?`, '提示', { type: 'warning' })
      const res = await memberUserService.delete(row.userId, tenantId)
      if (!checkCode(res.code)) throw new Error(res.msg || '删除失败')
    }
    if (action === 'unlock') {
      const res = await memberUserService.unlock(row.userId, tenantId)
      if (!checkCode(res.code)) throw new Error(res.msg || '解锁失败')
    }
    if (action === 'reset2fa') {
      const res = await memberUserService.reset2fa(row.userId, tenantId)
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
  const tenantId = Number(query.tenantId || 0)
  if (!tenantId) {
    ElMessage.warning('请先输入租户ID')
    return
  }
  const current = field === 'status' ? row.status : field === 'memberLevel' ? row.memberLevel : 0
  const titleMap = { status: '修改状态', memberLevel: '修改会员等级', riskLevel: '修改风控等级' }
  const { value } = await ElMessageBox.prompt(`请输入新的${titleMap[field]}`, titleMap[field], { inputValue: String(current) })
  try {
    if (field === 'status') {
      const res = await memberUserService.updateStatus(row.userId, { tenantId, status: Number(value) })
      if (!checkCode(res.code)) throw new Error(res.msg || '更新失败')
    }
    if (field === 'memberLevel') {
      const res = await memberUserService.updateLevel(row.userId, { tenantId, memberLevel: Number(value) })
      if (!checkCode(res.code)) throw new Error(res.msg || '更新失败')
    }
    if (field === 'riskLevel') {
      const res = await memberUserService.updateRiskLevel(row.userId, { tenantId, riskLevel: Number(value) })
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
        <el-table-column prop="userId" label="用户ID" width="100" />
        <el-table-column
          prop="userNo"
          label="用户编号"
          min-width="150"
          show-overflow-tooltip
        />
        <el-table-column prop="username" label="用户名" min-width="140" />
        <el-table-column prop="phone" label="手机号" min-width="140" />
        <el-table-column
          prop="email"
          label="邮箱"
          min-width="180"
          show-overflow-tooltip
        />
        <el-table-column prop="realName" label="实名" min-width="120" />
        <el-table-column prop="memberLevel" label="会员等级" width="100" />
        <el-table-column prop="status" label="状态" width="80" />
        <el-table-column prop="verifyStatus" label="认证状态" width="100" />
        <el-table-column label="注册时间" min-width="170">
          <template #default="{ row }">
            {{ formatDate(row.registerTime) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="420" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="showDetail(row)">
              详情
            </el-button>
            <el-button link type="primary" @click="openEdit(row)">
              编辑
            </el-button>
            <el-button link type="warning" @click="updateSimpleValue(row, 'status')">
              状态
            </el-button>
            <el-button link type="warning" @click="updateSimpleValue(row, 'memberLevel')">
              等级
            </el-button>
            <el-button link type="warning" @click="updateSimpleValue(row, 'riskLevel')">
              风控
            </el-button>
            <el-button link type="primary" @click="openPassword(row, 'login')">
              登录密码
            </el-button>
            <el-button link type="primary" @click="openPassword(row, 'pay')">
              支付密码
            </el-button>
            <el-button link type="success" @click="quickAction(row, 'unlock')">
              解锁
            </el-button>
            <el-button link type="warning" @click="quickAction(row, 'reset2fa')">
              重置2FA
            </el-button>
            <el-button link type="danger" @click="quickAction(row, 'delete')">
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="editVisible" :title="isCreate ? '新增用户' : '编辑用户'" width="720px">
      <el-form label-width="110px">
        <el-form-item label="租户ID">
          <el-input-number v-model="editForm.tenantId" :min="0" :precision="0" />
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
        <el-button type="primary" :loading="submitLoading" @click="submitEdit">
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

    <el-dialog v-model="detailVisible" title="用户详情" width="820px">
      <pre class="detail-pre">{{ JSON.stringify(detail, null, 2) }}</pre>
    </el-dialog>
  </div>
</template>

<style scoped></style>
