<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { usePagination } from '@/composables'
import { ElMessage } from 'element-plus'
import { memberUserService, type UserIdentityItem, type OptionGroup, UserDetail } from '@/services'
import { formatDate } from '@/utils'
import { findOptionGroup, getOptionLabel, getOptionValueLabel } from '@/utils/options'

const { t } = useI18n()
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)
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
  limit: 20,
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
    const res = await memberUserService.listIdentities({
      ...query,
      cursor: pagination.cursor,
      limit: pagination.limit,
    })
    if (!checkCode(res.code)) throw new Error(res.msg || t('users.loadFailed'))
    list.value = res.data || []
    updateFromResponse(res)
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
    cursor: pagination.cursor,
    limit: pagination.limit,
  })
  fetchList()
}

async function showDetail(row: UserIdentityItem) {
  const res = await memberUserService.getDetail(row.userId)
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

function handleLimitChange() {
  resetAndLoad(fetchList)
}

function handlePrevPage() {
  prevAndLoad(fetchList)
}

function handleNextPage() {
  nextAndLoad(fetchList)
}

onMounted(fetchList)
onMounted(fetchOptions)
</script>

<template>
  <div class="module-page">
    <div class="page-header">
      <h2>{{ t('users.identities') }}</h2>
      <el-button @click="fetchList">
        {{ t('common.refresh') }}
      </el-button>
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
          <el-button type="primary" @click="fetchList">
            {{ t('common.search') }}
          </el-button>
          <el-button @click="resetQuery">
            {{ t('common.reset') }}
          </el-button>
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
            <el-button link type="primary" @click="showDetail(row)">
              {{ t('common.detail') }}
            </el-button>
            <el-button
              v-perm="'users:user:identities:review'"
              link
              type="success"
              @click="openReview(row)"
            >
              {{ t('users.review') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <CursorPagination
        v-model:limit="pagination.limit"
        :total="pagination.total"
        :has-prev="pagination.hasPrev"
        :has-next="pagination.hasNext"
        @prev="handlePrevPage"
        @next="handleNextPage"
        @limit-change="handleLimitChange"
      />
    </el-card>

    <el-dialog v-model="reviewVisible" :title="t('users.reviewIdentity')" width="560px">
      <el-form label-width="100px">
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
        <el-button @click="reviewVisible = false">
          {{ t('common.cancel') }}
        </el-button>
        <el-button
          v-perm="'users:user:identities:review'"
          type="primary"
          :loading="submitLoading"
          @click="submitReview"
        >
          {{ t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="detailVisible" :title="t('users.identityDetail')" width="1000px">
      <el-tabs v-if="detailData" type="border-card">
        <!-- 用户基础信息标签页 -->
        <el-tab-pane :label="t('users.baseInfo')">
          <el-descriptions v-if="detailData.base" :column="2" border>
            <el-descriptions-item :label="t('common.id')" width="150px">
              {{ detailData.base.id }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('common.tenantId')">
              {{ detailData.base.tenantId }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('users.userNo')">
              {{ detailData.base.userNo }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('users.username')">
              {{ detailData.base.username }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('users.nickname')">
              {{ detailData.base.nickname || '--' }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('users.language')">
              {{ detailData.base.language || '--' }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('users.timezone')">
              {{ detailData.base.timezone || '--' }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('users.inviteCode')">
              {{ detailData.base.inviteCode || '--' }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('common.status')">
              {{ detailData.base.status }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('users.memberLevel')">
              {{ detailData.base.memberLevel }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('users.registerType')">
              {{ detailData.base.registerType }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('users.source')">
              {{ detailData.base.source || '--' }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('users.referrerUserId')">
              {{ detailData.base.referrerUserId || '--' }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('users.lastLoginIp')">
              {{ detailData.base.lastLoginIp || '--' }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('users.lastLoginTime')">
              {{ formatDate(detailData.base.lastLoginTime) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('users.registerIp')">
              {{ detailData.base.registerIp || '--' }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('users.registerTime')">
              {{ formatDate(detailData.base.registerTime) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('common.createTimes')">
              {{ formatDate(detailData.base.createTimes) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('common.updateTimes')">
              {{ formatDate(detailData.base.updateTimes) }}
            </el-descriptions-item>
          </el-descriptions>
        </el-tab-pane>

        <!-- 实名信息标签页 -->
        <el-tab-pane :label="t('users.identityInfo')">
          <el-descriptions v-if="detailData.identity" :column="2" border>
            <el-descriptions-item :label="t('common.id')" width="150px">
              {{ detailData.identity.id }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('common.tenantId')">
              {{ detailData.identity.tenantId }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('users.userId')">
              {{ detailData.identity.userId }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('users.realName')">
              {{ detailData.identity.realName }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('users.phone')">
              {{ detailData.identity.phone || '--' }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('users.email')">
              {{ detailData.identity.email || '--' }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('users.gender')">
              {{ detailData.identity.gender }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('users.birthday')">
              {{ formatDate(detailData.identity.birthday) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('users.country')">
              {{ detailData.identity.countryCode || '--' }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('users.province')">
              {{ detailData.identity.province || '--' }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('users.city')">
              {{ detailData.identity.city || '--' }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('users.address')">
              {{ detailData.identity.address || '--' }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('users.idType')">
              <span :class="getOptionTagClass('idType', detailData.identity.idType)">
                {{ getOptionValueLabel(optionGroups, 'idType', detailData.identity.idType, t) }}
              </span>
            </el-descriptions-item>
            <el-descriptions-item :label="t('users.idNo')">
              {{ detailData.identity.idNo || '--' }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('users.frontImage')">
              <el-image
                v-if="detailData.identity.frontImage"
                :src="detailData.identity.frontImage"
                style="width: 100px"
                preview-teleported
              />
            </el-descriptions-item>
            <el-descriptions-item :label="t('users.backImage')">
              <el-image
                v-if="detailData.identity.backImage"
                :src="detailData.identity.backImage"
                style="width: 100px"
                preview-teleported
              />
            </el-descriptions-item>
            <el-descriptions-item :label="t('users.handheldImage')">
              <el-image
                v-if="detailData.identity.handheldImage"
                :src="detailData.identity.handheldImage"
                style="width: 100px"
                preview-teleported
              />
            </el-descriptions-item>
            <el-descriptions-item :label="t('users.kycLevel')">
              <span :class="getOptionTagClass('kycLevel', detailData.identity.kycLevel)">
                {{ getOptionValueLabel(optionGroups, 'kycLevel', detailData.identity.kycLevel, t) }}
              </span>
            </el-descriptions-item>
            <el-descriptions-item :label="t('users.verifyStatus')">
              <span :class="getOptionTagClass('verifyStatus', detailData.identity.verifyStatus)">
                {{
                  getOptionValueLabel(
                    optionGroups,
                    'verifyStatus',
                    detailData.identity.verifyStatus,
                    t,
                  )
                }}
              </span>
            </el-descriptions-item>
            <el-descriptions-item :label="t('users.rejectReason')">
              {{ detailData.identity.rejectReason || '--' }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('users.submitTime')">
              {{ formatDate(detailData.identity.submitTime) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('users.verifyTime')">
              {{ formatDate(detailData.identity.verifyTime) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('users.verifyBy')">
              {{ detailData.identity.verifyBy || '--' }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('common.createTimes')">
              {{ formatDate(detailData.identity.createTimes) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('common.updateTimes')">
              {{ formatDate(detailData.identity.updateTimes) }}
            </el-descriptions-item>
          </el-descriptions>
        </el-tab-pane>

        <!-- 安全信息标签页 -->
        <el-tab-pane :label="t('users.securityInfo')">
          <el-descriptions v-if="detailData.security" :column="2" border>
            <el-descriptions-item :label="t('common.id')" width="150px">
              {{ detailData.security.id }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('common.tenantId')">
              {{ detailData.security.tenantId }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('users.userId')">
              {{ detailData.security.userId }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('users.googleEnabled')">
              {{ detailData.security.googleEnabled === 1 ? t('users.yes') : t('users.no') }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('users.loginErrorCount')">
              {{ detailData.security.loginErrorCount }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('users.payErrorCount')">
              {{ detailData.security.payErrorCount }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('users.lockUntil')">
              {{ formatDate(detailData.security.lockUntil) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('users.riskLevel')">
              {{ detailData.security.riskLevel }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('common.createTimes')">
              {{ formatDate(detailData.security.createTimes) }}
            </el-descriptions-item>
            <el-descriptions-item :label="t('common.updateTimes')">
              {{ formatDate(detailData.security.updateTimes) }}
            </el-descriptions-item>
          </el-descriptions>
        </el-tab-pane>

        <!-- 银行卡列表标签页 -->
        <el-tab-pane :label="t('users.bankList')">
          <el-table
            v-if="detailData.banks && detailData.banks.length"
            :data="detailData.banks"
            stripe
          >
            <el-table-column prop="id" :label="t('common.id')" width="80" />
            <el-table-column prop="bankName" :label="t('users.bankName')" min-width="120" />
            <el-table-column prop="accountName" :label="t('users.accountName')" min-width="120" />
            <el-table-column prop="accountNo" :label="t('users.accountNo')" min-width="140" />
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
          </el-table>
          <div v-else style="text-align: center; padding: 20px; color: #909399">
            {{ t('common.noData') }}
          </div>
        </el-tab-pane>
      </el-tabs>
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
