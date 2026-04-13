<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  memberUserService,
  tenantsService,
  type AddUserBankReq,
  type MemberUserBankItem,
  type UpdateMemberUserBankReq,
} from '@/services'
import { formatDate } from '@/utils'

const loading = ref(false)
const submitLoading = ref(false)
const list = ref<MemberUserBankItem[]>([])
const editVisible = ref(false)
const statusVisible = ref(false)
const detailVisible = ref(false)
const detailData = ref<any>(null)
const tenantChecking = ref(false)
const tenantChecked = ref(false)
const tenantExists = ref(false)
const tenantCheckName = ref('')
const userChecking = ref(false)
const userChecked = ref(false)
const userExists = ref(false)
const userCheckName = ref('')
const userCheckUserNo = ref('')

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

const bankStatusOptions = [
  { label: '正常', value: 1 },
  { label: '禁用', value: 2 },
]

const isCreate = computed(() => !form.id)
const canSubmitCreate = computed(() => !isCreate.value || (tenantChecked.value && tenantExists.value && userChecked.value && userExists.value))

function checkCode(code: number) {
  return code === 0 || code === 200
}

function getBankStatusLabel(value?: number) {
  return bankStatusOptions.find((item) => item.value === Number(value))?.label || String(value ?? '-')
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

async function verifyUser() {
  const tenantId = Number(form.tenantId || 0)
  const userId = Number(form.userId || 0)

  if (!tenantId) {
    ElMessage.warning('请先输入并确认租户ID')
    return false
  }
  if (!tenantChecked.value || !tenantExists.value) {
    const verified = await verifyTenant()
    if (!verified) return false
  }
  if (!userId) {
    resetUserCheck()
    ElMessage.warning('请输入用户ID')
    return false
  }

  userChecking.value = true
  try {
    const res = await memberUserService.getDetail(userId, tenantId)
    if (!checkCode(res.code)) throw new Error(res.msg || '查询用户失败')

    const data = res.detail || res.data
    const base = data?.base
    userChecked.value = true
    userExists.value = Boolean(base?.id)
    userCheckName.value = base?.username || ''
    userCheckUserNo.value = base?.userNo || ''

    if (!base?.id) {
      ElMessage.warning('用户不存在，请确认用户ID')
      return false
    }
    ElMessage.success(`已找到用户：${base.username}`)
    return true
  } catch (error: any) {
    resetUserCheck()
    ElMessage.error(error?.message || '查询用户失败')
    return false
  } finally {
    userChecking.value = false
  }
}

async function fetchList() {
  loading.value = true
  try {
    const res = await memberUserService.listBanks(query)
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

async function openEdit(row: MemberUserBankItem) {
  const tenantId = Number(row.tenantId || query.tenantId || 0)
  if (!tenantId) {
    ElMessage.warning('请先输入租户ID')
    return
  }
  const res = await memberUserService.getBank(row.id, tenantId)
  if (!checkCode(res.code)) {
    ElMessage.error(res.msg || '加载详情失败')
    return
  }
  Object.assign(form, res.bank || res.data, { id: row.id })
  resetTenantCheck()
  resetUserCheck()
  editVisible.value = true
}

async function showDetail(row: MemberUserBankItem) {
  const tenantId = Number(row.tenantId || query.tenantId || 0)
  if (!tenantId) {
    ElMessage.warning('请先输入租户ID')
    return
  }
  const res = await memberUserService.getBank(row.id, tenantId)
  if (!checkCode(res.code)) {
    ElMessage.error(res.msg || '加载详情失败')
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
      if (!checkCode(res.code)) throw new Error(res.msg || '更新失败')
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
      if (!checkCode(res.code)) throw new Error(res.msg || '新增失败')
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

function openStatus(row: MemberUserBankItem) {
  const tenantId = Number(row.tenantId || query.tenantId || 0)
  if (!tenantId) {
    ElMessage.warning('请先输入租户ID')
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
    if (!checkCode(res.code)) throw new Error(res.msg || '修改失败')
    ElMessage.success('修改成功')
    statusVisible.value = false
    fetchList()
  } catch (error: any) {
    ElMessage.error(error?.message || '修改失败')
  }
}

async function setDefault(row: MemberUserBankItem) {
  const tenantId = Number(row.tenantId || query.tenantId || 0)
  try {
    const res = await memberUserService.setDefaultBank(row.id, { tenantId, userId: row.userId })
    if (!checkCode(res.code)) throw new Error(res.msg || '设置失败')
    ElMessage.success('设置成功')
    fetchList()
  } catch (error: any) {
    ElMessage.error(error?.message || '设置失败')
  }
}

async function remove(row: MemberUserBankItem) {
  const tenantId = Number(row.tenantId || query.tenantId || 0)
  try {
    await ElMessageBox.confirm(`确认删除银行卡 ${row.bankName} ?`, '提示', { type: 'warning' })
    const res = await memberUserService.deleteBank(row.id, tenantId)
    if (!checkCode(res.code)) throw new Error(res.msg || '删除失败')
    ElMessage.success('删除成功')
    fetchList()
  } catch (error: any) {
    if (error === 'cancel') return
    ElMessage.error(error?.message || '删除失败')
  }
}

onMounted(fetchList)
</script>

<template>
  <div class="module-page">
    <div class="page-header">
      <h2>银行卡管理</h2>
      <div class="header-actions">
        <el-button @click="fetchList">
          刷新
        </el-button>
        <el-button type="primary" @click="openCreate">
          新增银行卡
        </el-button>
      </div>
    </div>

    <el-card shadow="never" class="query-card">
      <el-form :model="query" inline label-width="90px">
        <el-form-item label="租户ID">
          <el-input-number v-model="query.tenantId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item label="用户ID">
          <el-input-number v-model="query.userId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item label="关键字">
          <el-input v-model="query.keyword" clearable />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="query.status" clearable style="width: 140px">
            <el-option
              v-for="item in bankStatusOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
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
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="tenantId" label="租户ID" width="100" />
        <el-table-column prop="userId" label="用户ID" width="100" />
        <el-table-column prop="bankName" label="银行名" min-width="140" />
        <el-table-column prop="accountName" label="户名" min-width="120" />
        <el-table-column prop="accountNo" label="卡号" min-width="160" show-overflow-tooltip />
        <el-table-column prop="isDefault" label="默认" width="80">
          <template #default="{ row }">
            <el-tag :type="Number(row.isDefault) === 1 ? 'success' : 'info'">
              {{ Number(row.isDefault) === 1 ? '是' : '否' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'warning'">
              {{ getBankStatusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" min-width="170">
          <template #default="{ row }">
            {{ formatDate(row.createTimes) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="260" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="showDetail(row)">
              详情
            </el-button>
            <el-button link type="primary" @click="openEdit(row)">
              编辑
            </el-button>
            <el-button link type="warning" @click="openStatus(row)">
              状态
            </el-button>
            <el-button link type="success" @click="setDefault(row)">
              设默认
            </el-button>
            <el-button link type="danger" @click="remove(row)">
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="editVisible" :title="form.id ? '编辑银行卡' : '新增银行卡'" width="620px">
      <el-form label-width="100px">
        <el-form-item label="租户ID">
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
              确认租户
            </el-button>
          </div>
          <div v-if="isCreate" class="verify-tip">
            <span v-if="tenantChecked && tenantExists" class="verify-tip verify-tip--success">
              已确认租户：{{ tenantCheckName || form.tenantId }}
            </span>
            <span v-else-if="tenantChecked" class="verify-tip verify-tip--error">
              租户不存在，请重新输入
            </span>
            <span v-else class="verify-tip verify-tip--muted">
              新增前请先确认租户
            </span>
          </div>
        </el-form-item>
        <el-form-item label="用户ID">
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
              确认用户
            </el-button>
          </div>
          <div v-if="isCreate" class="verify-tip">
            <span v-if="userChecked && userExists" class="verify-tip verify-tip--success">
              已确认用户：{{ userCheckName }}<template v-if="userCheckUserNo">（{{ userCheckUserNo }}）</template>
            </span>
            <span v-else-if="userChecked" class="verify-tip verify-tip--error">
              用户不存在，请重新输入
            </span>
            <span v-else class="verify-tip verify-tip--muted">
              新增前请确认用户归属
            </span>
          </div>
        </el-form-item>
        <el-form-item label="银行名">
          <el-input v-model="form.bankName" />
        </el-form-item>
        <el-form-item label="银行编码">
          <el-input v-model="form.bankCode" />
        </el-form-item>
        <el-form-item label="户名">
          <el-input v-model="form.accountName" />
        </el-form-item>
        <el-form-item label="卡号">
          <el-input v-model="form.accountNo" />
        </el-form-item>
        <el-form-item label="支行">
          <el-input v-model="form.branchName" />
        </el-form-item>
        <el-form-item label="国家码">
          <el-input v-model="form.countryCode" />
        </el-form-item>
        <el-form-item label="默认">
          <el-switch
            v-model="form.isDefault"
            :active-value="1"
            :inactive-value="0"
          />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="form.status" style="width: 100%">
            <el-option
              v-for="item in bankStatusOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
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

    <el-dialog v-model="statusVisible" title="修改银行卡状态" width="420px">
      <el-form label-width="90px">
        <el-form-item label="状态">
          <el-select v-model="statusForm.status" style="width: 100%">
            <el-option
              v-for="item in bankStatusOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="statusVisible = false">
          取消
        </el-button>
        <el-button type="primary" @click="submitStatus">
          确定
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="detailVisible" title="银行卡详情" width="760px">
      <pre class="detail-pre">{{ JSON.stringify(detailData, null, 2) }}</pre>
    </el-dialog>
  </div>
</template>

<style scoped>
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
