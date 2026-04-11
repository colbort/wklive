<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { memberUserService, type MemberUserIdentityItem } from '@/services'
import { formatDate } from '@/utils'

const loading = ref(false)
const submitLoading = ref(false)
const list = ref<MemberUserIdentityItem[]>([])
const reviewVisible = ref(false)
const detailVisible = ref(false)
const detailData = ref<any>(null)

const query = reactive({
  tenantId: undefined as number | undefined,
  keyword: '',
  userId: undefined as number | undefined,
  username: '',
  realName: '',
  verifyStatus: undefined as number | undefined,
  kycLevel: undefined as number | undefined,
  limit: 100,
})

const reviewForm = reactive({
  userId: 0,
  tenantId: 0,
  verifyStatus: 1,
  rejectReason: '',
  verifyBy: 0,
})

function checkCode(code: number) {
  return code === 0 || code === 200
}

async function fetchList() {
  loading.value = true
  try {
    const res = await memberUserService.listIdentities(query)
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
    userId: undefined,
    username: '',
    realName: '',
    verifyStatus: undefined,
    kycLevel: undefined,
    limit: 100,
  })
  fetchList()
}

async function showDetail(row: MemberUserIdentityItem) {
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
  detailData.value = res.detail || res.data
  detailVisible.value = true
}

function openReview(row: MemberUserIdentityItem) {
  Object.assign(reviewForm, {
    userId: row.userId,
    tenantId: Number(query.tenantId || 0),
    verifyStatus: row.verifyStatus || 1,
    rejectReason: row.rejectReason || '',
    verifyBy: row.verifyBy || 0,
  })
  reviewVisible.value = true
}

async function submitReview() {
  submitLoading.value = true
  try {
    const res = await memberUserService.reviewIdentity(reviewForm.userId, {
      tenantId: reviewForm.tenantId,
      verifyStatus: reviewForm.verifyStatus,
      rejectReason: reviewForm.rejectReason || undefined,
      verifyBy: reviewForm.verifyBy,
    })
    if (!checkCode(res.code)) throw new Error(res.msg || '审核失败')
    ElMessage.success('审核成功')
    reviewVisible.value = false
    fetchList()
  } catch (error: any) {
    ElMessage.error(error?.message || '审核失败')
  } finally {
    submitLoading.value = false
  }
}

onMounted(fetchList)
</script>

<template>
  <div class="module-page">
    <div class="page-header">
      <h2>实名认证</h2>
      <el-button @click="fetchList">
        刷新
      </el-button>
    </div>

    <el-card shadow="never" class="query-card">
      <el-form :model="query" inline label-width="90px">
        <el-form-item label="租户ID">
          <el-input-number v-model="query.tenantId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item label="关键字">
          <el-input v-model="query.keyword" clearable />
        </el-form-item>
        <el-form-item label="用户ID">
          <el-input-number v-model="query.userId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item label="用户名">
          <el-input v-model="query.username" clearable />
        </el-form-item>
        <el-form-item label="实名">
          <el-input v-model="query.realName" clearable />
        </el-form-item>
        <el-form-item label="认证状态">
          <el-input-number v-model="query.verifyStatus" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item label="KYC等级">
          <el-input-number v-model="query.kycLevel" :min="0" :precision="0" />
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
        <el-table-column prop="userNo" label="用户编号" min-width="150" />
        <el-table-column prop="username" label="用户名" min-width="140" />
        <el-table-column prop="phone" label="手机号" min-width="140" />
        <el-table-column prop="email" label="邮箱" min-width="180" />
        <el-table-column prop="realName" label="实名" min-width="120" />
        <el-table-column prop="idType" label="证件类型" width="100" />
        <el-table-column prop="kycLevel" label="KYC等级" width="100" />
        <el-table-column prop="verifyStatus" label="认证状态" width="100" />
        <el-table-column label="提交时间" min-width="170">
          <template #default="{ row }">
            {{ formatDate(row.submitTime) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="showDetail(row)">
              详情
            </el-button>
            <el-button link type="success" @click="openReview(row)">
              审核
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="reviewVisible" title="实名认证审核" width="560px">
      <el-form label-width="100px">
        <el-form-item label="租户ID">
          <el-input-number v-model="reviewForm.tenantId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item label="审核状态">
          <el-input-number v-model="reviewForm.verifyStatus" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item label="审核人ID">
          <el-input-number v-model="reviewForm.verifyBy" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item label="拒绝原因">
          <el-input v-model="reviewForm.rejectReason" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="reviewVisible = false">
          取消
        </el-button>
        <el-button type="primary" :loading="submitLoading" @click="submitReview">
          确定
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="detailVisible" title="认证详情" width="820px">
      <pre class="detail-pre">{{ JSON.stringify(detailData, null, 2) }}</pre>
    </el-dialog>
  </div>
</template>

<style scoped></style>
