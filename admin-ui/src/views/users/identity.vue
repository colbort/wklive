<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { memberUserService, type UserIdentityItem, type OptionGroup, UserDetail } from '@/services'
import { formatDate } from '@/utils'
import { findOptionGroup, getOptionLabel, getOptionValueLabel } from '@/utils/options'

const { t } = useI18n()
const loading = ref(false)
const submitLoading = ref(false)
const list = ref<UserIdentityItem[]>([])
const reviewVisible = ref(false)
const detailVisible = ref(false)
const detailData = ref<UserDetail>()
const optionGroups = ref<OptionGroup[]>([])
const verifyStatusOptions = computed(() => findOptionGroup(optionGroups.value, 'verifyStatus'))
const kycLevelOptions = computed(() => findOptionGroup(optionGroups.value, 'kycLevel'))

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

function getOptionTagClass(groupKey: string, value?: number) {
  const normalizedValue = Number(value ?? 0)

  if (groupKey === 'idType') {
    const idTypeMap: Record<number, string> = {
      0: 'option-tag option-tag--slate',
      1: 'option-tag option-tag--blue',
      2: 'option-tag option-tag--green',
      3: 'option-tag option-tag--orange',
    }
    return idTypeMap[normalizedValue] || 'option-tag'
  }

  if (groupKey === 'kycLevel') {
    const kycLevelMap: Record<number, string> = {
      0: 'option-tag option-tag--slate',
      1: 'option-tag option-tag--blue',
      2: 'option-tag option-tag--green',
    }
    return kycLevelMap[normalizedValue] || 'option-tag'
  }

  if (groupKey === 'verifyStatus') {
    const verifyStatusMap: Record<number, string> = {
      0: 'option-tag option-tag--slate',
      1: 'option-tag option-tag--orange',
      2: 'option-tag option-tag--green',
      3: 'option-tag option-tag--red',
    }
    return verifyStatusMap[normalizedValue] || 'option-tag'
  }

  return 'option-tag'
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

async function fetchList() {
  loading.value = true
  try {
    const res = await memberUserService.listIdentities(query)
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
    userId: undefined,
    username: '',
    realName: '',
    verifyStatus: undefined,
    kycLevel: undefined,
    limit: 100,
  })
  fetchList()
}

async function showDetail(row: UserIdentityItem) {
  const tenantId = Number(query.tenantId || 0)
  if (!tenantId) {
    ElMessage.warning(t('users.queryTenantPrompt'))
    return
  }
  const res = await memberUserService.getDetail(row.userId, tenantId)
  if (!checkCode(res.code)) {
    ElMessage.error(res.msg || t('users.loadDetailFailed'))
    return
  }
  detailData.value = res.detail || res.data
  detailVisible.value = true
}

function openReview(row: UserIdentityItem) {
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
    if (!checkCode(res.code)) throw new Error(res.msg || t('users.reviewFailed'))
    ElMessage.success(t('users.reviewSuccess'))
    reviewVisible.value = false
    fetchList()
  } catch (error: unknown) {
    ElMessage.error(error instanceof Error ? error.message : t('users.reviewFailed'))
  } finally {
    submitLoading.value = false
  }
}

onMounted(fetchList)
onMounted(fetchOptions)
</script>

<template>
  <div class="module-page">
    <div class="page-header">
      <h2>{{ t('users.identities') }}</h2>
      <el-button @click="fetchList"> {{ t('common.refresh') }} </el-button>
    </div>

    <el-card shadow="never" class="query-card">
      <el-form :model="query" inline label-width="90px">
        <el-form-item :label="t('common.tenantId')">
          <el-input-number v-model="query.tenantId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('common.keyword')">
          <el-input v-model="query.keyword" clearable />
        </el-form-item>
        <el-form-item :label="t('users.userId')">
          <el-input-number v-model="query.userId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('users.username')">
          <el-input v-model="query.username" clearable />
        </el-form-item>
        <el-form-item :label="t('users.realName')">
          <el-input v-model="query.realName" clearable />
        </el-form-item>
        <el-form-item :label="t('users.verifyStatus')">
          <el-select v-model="query.verifyStatus" clearable style="width: 140px">
            <el-option
              v-for="item in verifyStatusOptions"
              :key="item.value"
              :label="getOptionLabel(t, item.code, item.value)"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('users.kycLevel')">
          <el-select v-model="query.kycLevel" clearable style="width: 140px">
            <el-option
              v-for="item in kycLevelOptions"
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
        <el-table-column prop="userId" :label="t('users.userId')" width="100" />
        <el-table-column prop="userNo" :label="t('users.userNo')" min-width="150" />
        <el-table-column prop="username" :label="t('users.username')" min-width="140" />
        <el-table-column prop="phone" :label="t('users.phone')" min-width="140" />
        <el-table-column prop="email" :label="t('users.email')" min-width="180" />
        <el-table-column prop="realName" :label="t('users.realName')" min-width="120" />
        <el-table-column :label="t('users.idType')" width="110">
          <template #default="{ row }">
            <span :class="getOptionTagClass('idType', row.idType)">
              {{ getOptionValueLabel(optionGroups, 'idType', row.idType, t) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column :label="t('users.kycLevel')" width="100">
          <template #default="{ row }">
            <span :class="getOptionTagClass('kycLevel', row.kycLevel)">
              {{ getOptionValueLabel(optionGroups, 'kycLevel', row.kycLevel, t) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column :label="t('users.verifyStatus')" width="110">
          <template #default="{ row }">
            <span :class="getOptionTagClass('verifyStatus', row.verifyStatus)">
              {{ getOptionValueLabel(optionGroups, 'verifyStatus', row.verifyStatus, t) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column :label="t('users.submitTime')" min-width="170">
          <template #default="{ row }">
            {{ formatDate(row.submitTime) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('common.actions')" width="160" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="showDetail(row)"> {{ t('common.detail') }} </el-button>
            <el-button link type="success" @click="openReview(row)"> {{ t('users.review') }} </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="reviewVisible" :title="t('users.reviewIdentity')" width="560px">
      <el-form label-width="100px">
        <el-form-item :label="t('common.tenantId')">
          <el-input-number v-model="reviewForm.tenantId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('users.verifyStatus')">
          <el-select v-model="reviewForm.verifyStatus" style="width: 100%">
            <el-option
              v-for="item in verifyStatusOptions"
              :key="item.value"
              :label="getOptionLabel(t, item.code, item.value)"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('users.reviewerId')">
          <el-input-number v-model="reviewForm.verifyBy" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('users.rejectReason')">
          <el-input v-model="reviewForm.rejectReason" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="reviewVisible = false"> {{ t('common.cancel') }} </el-button>
        <el-button type="primary" :loading="submitLoading" @click="submitReview"> {{ t('common.confirm') }} </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="detailVisible" :title="t('users.identityDetail')" width="820px">
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
</style>
