<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { memberUserService, type AddUserBankReq, type MemberUserBankItem, type UpdateMemberUserBankReq } from '@/services'
import { formatDate } from '@/utils'

const loading = ref(false)
const submitLoading = ref(false)
const list = ref<MemberUserBankItem[]>([])
const editVisible = ref(false)
const detailVisible = ref(false)
const detailData = ref<any>(null)

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
  isDefault: false,
  status: 1,
})

function checkCode(code: number) {
  return code === 0 || code === 200
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
    isDefault: false,
    status: 1,
  })
  editVisible.value = true
}

async function openEdit(row: MemberUserBankItem) {
  const tenantId = Number(query.tenantId || 0)
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
  editVisible.value = true
}

async function showDetail(row: MemberUserBankItem) {
  const tenantId = Number(query.tenantId || 0)
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
        isDefault: Boolean(form.isDefault),
        status: Number(form.status),
      }
      const res = await memberUserService.updateBank(Number(form.id), payload)
      if (!checkCode(res.code)) throw new Error(res.msg || '更新失败')
    } else {
      const payload: AddUserBankReq = {
        tenantId: Number(form.tenantId),
        userId: Number(form.userId),
        bankName: form.bankName,
        bankCode: form.bankCode || undefined,
        accountName: form.accountName,
        accountNo: form.accountNo,
        branchName: form.branchName || undefined,
        countryCode: form.countryCode || undefined,
        isDefault: Boolean(form.isDefault),
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

async function updateStatus(row: MemberUserBankItem) {
  const tenantId = Number(query.tenantId || 0)
  const { value } = await ElMessageBox.prompt('请输入新的状态值', '修改状态', { inputValue: String(row.status) })
  try {
    const res = await memberUserService.updateBankStatus(row.id, { tenantId, status: Number(value) })
    if (!checkCode(res.code)) throw new Error(res.msg || '修改失败')
    ElMessage.success('修改成功')
    fetchList()
  } catch (error: any) {
    ElMessage.error(error?.message || '修改失败')
  }
}

async function setDefault(row: MemberUserBankItem) {
  const tenantId = Number(query.tenantId || 0)
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
  const tenantId = Number(query.tenantId || 0)
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
          <el-input-number v-model="query.status" :min="0" :precision="0" />
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
        <el-table-column prop="userId" label="用户ID" width="100" />
        <el-table-column prop="username" label="用户名" min-width="140" />
        <el-table-column prop="bankName" label="银行名" min-width="140" />
        <el-table-column prop="accountName" label="户名" min-width="120" />
        <el-table-column prop="maskedAccountNo" label="卡号" min-width="160" />
        <el-table-column prop="isDefault" label="默认" width="80">
          <template #default="{ row }">
            <el-tag :type="row.isDefault ? 'success' : 'info'">
              {{ row.isDefault ? '是' : '否' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="80" />
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
            <el-button link type="warning" @click="updateStatus(row)">
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
          <el-input-number v-model="form.tenantId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item label="用户ID">
          <el-input-number v-model="form.userId" :min="0" :precision="0" />
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
          <el-switch v-model="form.isDefault" />
        </el-form-item>
        <el-form-item label="状态">
          <el-input-number v-model="form.status" :min="0" :precision="0" />
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

    <el-dialog v-model="detailVisible" title="银行卡详情" width="760px">
      <pre class="detail-pre">{{ JSON.stringify(detailData, null, 2) }}</pre>
    </el-dialog>
  </div>
</template>

<style scoped></style>
