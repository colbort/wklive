<script setup lang="ts">
/**
 * 租户用户管理：固定当前租户，查询用户、查看详情、创建用户、重置密码和解锁。
 */
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { memberUserService, type MemberUserDetail, type MemberUserItem } from '@/services'
import { useTenantStore } from '@/stores/tenant'
import { formatDate } from '@/utils'

const tenant = useTenantStore()
const loading = ref(false)
const submitLoading = ref(false)
const list = ref<MemberUserItem[]>([])
const detailVisible = ref(false)
const detail = ref<MemberUserDetail | null>(null)
const formVisible = ref(false)
const editing = ref<MemberUserItem | null>(null)

const query = reactive({
  keyword: '',
  username: '',
  phone: '',
  email: '',
  status: undefined as number | undefined,
})

const form = reactive({
  username: '',
  nickname: '',
  phone: '',
  email: '',
  password: '',
  status: 1,
  registerType: 1,
  memberLevel: 0,
  language: 'zh-CN',
  timezone: 'Asia/Hong_Kong',
  remark: '',
})

async function loadList() {
  await tenant.ensureLoaded()
  loading.value = true
  try {
    const res = await memberUserService.getList({
      tenantId: tenant.tenantId,
      tenantCode: tenant.tenantCode,
      keyword: query.keyword || undefined,
      username: query.username || undefined,
      phone: query.phone || undefined,
      email: query.email || undefined,
      status: query.status,
      limit: 100,
    })
    list.value = res.data || []
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editing.value = null
  Object.assign(form, {
    username: '',
    nickname: '',
    phone: '',
    email: '',
    password: '',
    status: 1,
    registerType: 1,
    memberLevel: 0,
    language: 'zh-CN',
    timezone: 'Asia/Hong_Kong',
    remark: '',
  })
  formVisible.value = true
}

function openEdit(row: MemberUserItem) {
  editing.value = row
  Object.assign(form, {
    username: row.username,
    nickname: row.nickname || '',
    phone: '',
    email: '',
    password: '',
    status: row.status,
    registerType: row.registerType,
    memberLevel: row.memberLevel,
    language: row.language || 'zh-CN',
    timezone: row.timezone || 'Asia/Hong_Kong',
    remark: row.remark || '',
  })
  formVisible.value = true
}

async function showDetail(row: MemberUserItem) {
  const res = await memberUserService.getDetail(row.id, row.tenantId)
  detail.value = res.data || null
  detailVisible.value = true
}

async function submitForm() {
  await tenant.ensureLoaded()
  submitLoading.value = true
  try {
    if (editing.value) {
      await memberUserService.updateBase(editing.value.id, {
        tenantId: tenant.tenantId,
        username: form.username,
        nickname: form.nickname,
        language: form.language,
        timezone: form.timezone,
        remark: form.remark,
        phone: form.phone || undefined,
        email: form.email || undefined,
      })
      await memberUserService.updateStatus(editing.value.id, {
        tenantId: tenant.tenantId,
        status: form.status,
      })
      ElMessage.success('用户已更新')
    } else {
      if (!form.username || !form.password) {
        ElMessage.warning('请填写用户名和密码')
        return
      }
      await memberUserService.create({
        tenantCode: tenant.tenantCode,
        username: form.username,
        nickname: form.nickname || undefined,
        phone: form.phone || undefined,
        email: form.email || undefined,
        password: form.password,
        registerType: form.registerType,
        status: form.status,
        memberLevel: form.memberLevel,
        language: form.language,
        timezone: form.timezone,
        remark: form.remark || undefined,
      })
      ElMessage.success('用户已创建')
    }
    formVisible.value = false
    await loadList()
  } finally {
    submitLoading.value = false
  }
}

async function resetPassword(row: MemberUserItem) {
  const promptResult = (await ElMessageBox.prompt('请输入新的登录密码', '重置登录密码', {
    inputPattern: /^.{6,}$/,
    inputErrorMessage: '密码至少 6 位',
  })) as { value: string }
  await memberUserService.resetLoginPassword(row.id, row.tenantId, promptResult.value)
  ElMessage.success('登录密码已重置')
}

async function unlockUser(row: MemberUserItem) {
  await ElMessageBox.confirm(`确认解锁用户 ${row.username} 吗？`, '提示')
  await memberUserService.unlock(row.id, row.tenantId)
  ElMessage.success('用户已解锁')
}

onMounted(async () => {
  await tenant.ensureLoaded()
  await loadList()
})
</script>

<template>
  <div class="page">
    <div class="page-header">
      <div>
        <h2>用户管理</h2>
        <p>当前租户：{{ tenant.tenantName || tenant.tenantCode }}</p>
      </div>
      <el-button type="primary" @click="openCreate">创建用户</el-button>
    </div>

    <el-card shadow="never">
      <el-form :model="query" inline>
        <el-form-item label="关键字">
          <el-input v-model="query.keyword" clearable style="width: 180px" />
        </el-form-item>
        <el-form-item label="用户名">
          <el-input v-model="query.username" clearable style="width: 160px" />
        </el-form-item>
        <el-form-item label="手机号">
          <el-input v-model="query.phone" clearable style="width: 160px" />
        </el-form-item>
        <el-form-item label="邮箱">
          <el-input v-model="query.email" clearable style="width: 180px" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="query.status" clearable style="width: 140px">
            <el-option label="启用" :value="1" />
            <el-option label="停用" :value="2" />
            <el-option label="冻结" :value="3" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadList">查询</el-button>
          <el-button @click="Object.assign(query, { keyword: '', username: '', phone: '', email: '', status: undefined }); loadList()">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="never">
      <el-table v-loading="loading" :data="list" stripe>
        <el-table-column prop="id" label="用户ID" width="90" />
        <el-table-column prop="userNo" label="用户编号" min-width="140" />
        <el-table-column prop="username" label="用户名" min-width="140" />
        <el-table-column prop="nickname" label="昵称" min-width="120" />
        <el-table-column prop="memberLevel" label="会员等级" width="100" />
        <el-table-column prop="status" label="状态" width="90" />
        <el-table-column prop="isGuest" label="游客" width="80" />
        <el-table-column label="注册时间" min-width="170">
          <template #default="{ row }">{{ formatDate(row.registerTime) }}</template>
        </el-table-column>
        <el-table-column label="最近登录" min-width="170">
          <template #default="{ row }">{{ formatDate(row.lastLoginTime) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="260" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="showDetail(row)">详情</el-button>
            <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
            <el-button link type="warning" @click="resetPassword(row)">重置密码</el-button>
            <el-button link type="success" @click="unlockUser(row)">解锁</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="formVisible" :title="editing ? '编辑用户' : '创建用户'" width="560px">
      <el-form label-width="100px">
        <el-form-item label="用户名">
          <el-input v-model="form.username" :disabled="Boolean(editing)" />
        </el-form-item>
        <el-form-item v-if="!editing" label="密码">
          <el-input v-model="form.password" type="password" show-password />
        </el-form-item>
        <el-form-item label="昵称">
          <el-input v-model="form.nickname" />
        </el-form-item>
        <el-form-item label="手机号">
          <el-input v-model="form.phone" />
        </el-form-item>
        <el-form-item label="邮箱">
          <el-input v-model="form.email" />
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="form.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="2">停用</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="会员等级">
          <el-input-number v-model="form.memberLevel" :min="0" :precision="0" style="width: 100%" />
        </el-form-item>
        <el-form-item label="语言">
          <el-input v-model="form.language" />
        </el-form-item>
        <el-form-item label="时区">
          <el-input v-model="form.timezone" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remark" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="submitForm">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="detailVisible" title="用户详情" width="780px">
      <pre class="detail-pre">{{ JSON.stringify(detail, null, 2) }}</pre>
    </el-dialog>
  </div>
</template>

<style scoped>
.page { display: grid; gap: 16px; }
.page-header { display: flex; align-items: center; justify-content: space-between; }
.page-header h2 { margin: 0 0 6px; }
.page-header p { margin: 0; color: #909399; }
.detail-pre { margin: 0; white-space: pre-wrap; word-break: break-all; }
</style>
