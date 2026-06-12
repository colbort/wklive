<template>
  <div class="payment-page">
    <div class="page-header">
      <h2>{{ t('payment.tenantChannels') }}</h2>
      <el-button @click="loadChannels">
        {{ t('common.refresh') }}
      </el-button>
    </div>

    <CrudQueryCard :model="channelQuery" label-width="90px" :show-actions="false">
      <el-form-item :label="t('common.tenantId')">
        <TenantSelect v-model="channelQuery.tenantId" class="tenant-select-filter" />
      </el-form-item>
      <el-form-item :label="t('payment.platformId')">
        <el-input-number v-model="channelQuery.platformId" :min="0" :precision="0" />
      </el-form-item>
      <el-form-item :label="t('common.keyword')">
        <el-input v-model="channelQuery.keyword" clearable />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="loadChannels">
          {{ t('common.search') }}
        </el-button>
        <el-button
          v-perm="'payment:tenant-channel:add'"
          type="primary"
          @click="openChannelDialog()"
        >
          {{ t('payment.addChannel') }}
        </el-button>
      </el-form-item>
    </CrudQueryCard>

    <el-card shadow="never" class="table-card">
      <el-table v-loading="channelLoading" :data="channels" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="tenantId" :label="t('common.tenantId')" width="100" />
        <el-table-column prop="channelCode" :label="t('payment.channelCode')" min-width="120" />
        <el-table-column prop="channelName" :label="t('payment.channelName')" min-width="140" />
        <el-table-column prop="displayName" :label="t('payment.displayName')" min-width="140" />
        <el-table-column prop="currency" :label="t('payment.currency')" width="90" />
        <el-table-column :label="t('common.enabled')" width="100">
          <template #default="{ row }">
            <el-tag :class="getEnabledTagClass(row.enabled)" disable-transitions>
              {{ getOptionValueLabel(optionGroups, 'enabled', row.enabled, t) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="t('common.visible')" width="90">
          <template #default="{ row }">
            {{ getOptionValueLabel(optionGroups, 'visible', row.visible, t) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('common.actions')" width="160">
          <template #default="{ row }">
            <el-button
              v-perm="'payment:tenant-channel:detail'"
              link
              type="primary"
              @click="showChannelDetail(row)"
            >
              {{ t('common.detail') }}
            </el-button>
            <el-button
              v-perm="'payment:tenant-channel:update'"
              link
              type="primary"
              @click="openChannelDialog(row)"
            >
              {{ t('common.edit') }}
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

    <el-dialog
      v-model="channelDialogVisible"
      :title="channelForm.id ? t('payment.editChannel') : t('payment.addChannel')"
      width="760px"
    >
      <el-form label-width="110px">
        <el-form-item :label="t('common.tenantId')">
          <TenantSelect
            v-model="channelForm.tenantId"
            :disabled="!!channelForm.id"
            @change="handleChannelTenantChange"
          />
        </el-form-item>

        <el-form-item v-if="!channelForm.id" :label="t('payment.platformId')">
          <div class="verify-row">
            <el-input-number
              v-model="channelForm.platformId"
              :min="1"
              :precision="0"
              @change="handleChannelPlatformChange"
            />
            <el-button :loading="channelPlatformChecking" @click="checkChannelPlatform">
              {{ t('payment.verifyPlatform') }}
            </el-button>
            <span v-if="channelPlatformVerified" class="verified-text">
              {{ t('payment.verified') }}
            </span>
          </div>
        </el-form-item>

        <el-form-item v-if="!channelForm.id" :label="t('payment.productId')">
          <div class="verify-row">
            <el-input-number
              v-model="channelForm.productId"
              :min="1"
              :precision="0"
              @change="handleChannelProductChange"
            />
            <el-button :loading="channelProductChecking" @click="checkChannelProduct">
              {{ t('payment.verifyProduct') }}
            </el-button>
            <span v-if="channelProductVerified" class="verified-text">
              {{ t('payment.verified') }}
            </span>
          </div>
        </el-form-item>

        <el-form-item v-if="!channelForm.id" :label="t('payment.accountId')">
          <div class="verify-row">
            <el-input-number
              v-model="channelForm.accountId"
              :min="1"
              :precision="0"
              @change="handleChannelAccountChange"
            />
            <el-button :loading="channelAccountChecking" @click="checkChannelAccount">
              {{ t('payment.verifyAccount') }}
            </el-button>
            <span v-if="channelAccountVerified" class="verified-text">
              {{ t('payment.verified') }}
            </span>
          </div>
        </el-form-item>

        <el-form-item v-if="!channelForm.id" :label="t('payment.channelCode')">
          <el-input v-model="channelForm.channelCode" />
        </el-form-item>
        <el-form-item :label="t('payment.channelName')">
          <el-input v-model="channelForm.channelName" />
        </el-form-item>
        <el-form-item :label="t('payment.displayName')">
          <el-input v-model="channelForm.displayName" />
        </el-form-item>
        <el-form-item :label="t('payment.currency')">
          <el-input v-model="channelForm.currency" />
        </el-form-item>
        <el-form-item :label="t('common.enabled')">
          <el-select v-model="channelForm.enabled" style="width: 100%">
            <el-option
              v-for="item in enabledOptions"
              :key="item.value"
              :label="getOptionLabel(t, item.code, item.value)"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('common.visible')">
          <el-select v-model="channelForm.visible" style="width: 100%">
            <el-option
              v-for="item in visibleOptions"
              :key="item.value"
              :label="getOptionLabel(t, item.code, item.value)"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('payment.feeType')">
          <el-select v-model="channelForm.feeType" style="width: 100%">
            <el-option
              v-for="item in feeTypeOptions"
              :key="item.value"
              :label="getOptionLabel(t, item.code, item.value)"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('payment.feeRate')">
          <el-input v-model="channelForm.feeRate" />
        </el-form-item>
        <el-form-item :label="t('payment.feeFixedAmount')">
          <el-input-number v-model="channelForm.feeFixedAmount" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('payment.extConfig')">
          <el-input v-model="channelForm.extConfig" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item :label="t('common.remark')">
          <el-input v-model="channelForm.remark" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="channelDialogVisible = false">
          {{ t('common.cancel') }}
        </el-button>
        <el-button
          v-perm="channelForm.id ? 'payment:tenant-channel:update' : 'payment:tenant-channel:add'"
          type="primary"
          :disabled="channelSubmitDisabled"
          @click="submitChannel"
        >
          {{ t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="detailVisible" :title="t('payment.detailTitle')" width="700px">
      <PaymentDetailDescriptions :data="detailData" :option-groups="optionGroups" />
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { usePagination } from '@/composables'
import { ElMessage } from 'element-plus'
import { catalogService, tenantService, type OptionGroup, type TenantPayChannel } from '@/services'
import { findOptionGroup, getOptionLabel, getOptionValueLabel } from '@/utils/options'
import TenantSelect from '@/components/TenantSelect.vue'
import PaymentDetailDescriptions from '@/components/payment/PaymentDetailDescriptions.vue'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'

const { t } = useI18n()
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)

const channelLoading = ref(false)
const channels = ref<TenantPayChannel[]>([])
const detailVisible = ref(false)
const detailData = ref<TenantPayChannel | null>(null)
const channelDialogVisible = ref(false)

const optionGroups = ref<OptionGroup[]>([])
const enabledOptions = computed(() => findOptionGroup(optionGroups.value, 'enabled'))
const visibleOptions = computed(() => findOptionGroup(optionGroups.value, 'visible'))
const feeTypeOptions = computed(() => findOptionGroup(optionGroups.value, 'feeType'))

const channelQuery = reactive({ tenantId: 0, platformId: 0, keyword: '' })

const channelForm = reactive({
  id: 0,
  tenantId: 0,
  platformId: 0,
  productId: 0,
  accountId: 0,
  channelCode: '',
  channelName: '',
  displayName: '',
  icon: '',
  currency: '',
  sort: 0,
  visible: 1,
  enabled: 1,
  singleMinAmount: 0,
  singleMaxAmount: 0,
  dailyMaxAmount: 0,
  dailyMaxCount: 0,
  feeType: 1,
  feeRate: '',
  feeFixedAmount: 0,
  extConfig: '',
  remark: '',
})

const channelPlatformChecking = ref(false)
const channelProductChecking = ref(false)
const channelAccountChecking = ref(false)

const channelTenantVerified = ref(false)
const channelPlatformVerified = ref(false)
const channelProductVerified = ref(false)
const channelAccountVerified = ref(false)

const verifiedChannelTenantId = ref(0)
const verifiedChannelPlatformId = ref(0)
const verifiedChannelProductId = ref(0)
const verifiedChannelAccountId = ref(0)

const channelSubmitDisabled = computed(
  () =>
    !channelForm.id &&
    (!channelTenantVerified.value ||
      !channelPlatformVerified.value ||
      !channelProductVerified.value ||
      !channelAccountVerified.value ||
      verifiedChannelTenantId.value !== channelForm.tenantId ||
      verifiedChannelPlatformId.value !== channelForm.platformId ||
      verifiedChannelProductId.value !== channelForm.productId ||
      verifiedChannelAccountId.value !== channelForm.accountId),
)

const loadOptions = async () => {
  const res = await tenantService.getOptions()
  optionGroups.value = res.data || []
}

const loadChannels = async () => {
  channelLoading.value = true
  try {
    const res = await tenantService.getTenantChannelList({
      ...channelQuery,
      tenantId: channelQuery.tenantId || undefined,
      platformId: channelQuery.platformId || undefined,
      keyword: channelQuery.keyword || undefined,
      cursor: pagination.cursor,
      limit: pagination.limit,
    })
    channels.value = res.data || []
    updateFromResponse(res)
  } finally {
    channelLoading.value = false
  }
}

const resetChannelVerifyState = () => {
  channelTenantVerified.value = false
  channelPlatformVerified.value = false
  channelProductVerified.value = false
  channelAccountVerified.value = false
  verifiedChannelTenantId.value = 0
  verifiedChannelPlatformId.value = 0
  verifiedChannelProductId.value = 0
  verifiedChannelAccountId.value = 0
}

const openChannelDialog = (row?: TenantPayChannel) => {
  Object.assign(
    channelForm,
    row || {
      id: 0,
      tenantId: 0,
      platformId: 0,
      productId: 0,
      accountId: 0,
      channelCode: '',
      channelName: '',
      displayName: '',
      icon: '',
      currency: '',
      sort: 0,
      visible: 1,
      enabled: 1,
      singleMinAmount: 0,
      singleMaxAmount: 0,
      dailyMaxAmount: 0,
      dailyMaxCount: 0,
      feeType: 1,
      feeRate: '',
      feeFixedAmount: 0,
      extConfig: '',
      remark: '',
    },
  )

  if (row?.id) {
    channelTenantVerified.value = true
    channelPlatformVerified.value = true
    channelProductVerified.value = true
    channelAccountVerified.value = true
    verifiedChannelTenantId.value = row.tenantId
    verifiedChannelPlatformId.value = row.platformId
    verifiedChannelProductId.value = row.productId
    verifiedChannelAccountId.value = row.accountId
  } else {
    resetChannelVerifyState()
  }

  channelDialogVisible.value = true
}

const validatePlatformExists = async (platformId: number) => {
  if (!platformId) {
    ElMessage.warning(t('payment.pleaseInputPlatformId'))
    return false
  }

  channelPlatformChecking.value = true
  try {
    const res = await catalogService.getPlatformDetail(platformId)
    if (!res.data?.id) {
      ElMessage.error(t('payment.platformNotFound'))
      return false
    }
    return true
  } catch {
    ElMessage.error(t('payment.platformNotFound'))
    return false
  } finally {
    channelPlatformChecking.value = false
  }
}

const validateProductExists = async (productId: number, platformId: number) => {
  if (!productId) {
    ElMessage.warning(t('payment.pleaseInputProductId'))
    return false
  }
  if (!platformId) {
    ElMessage.warning(t('payment.pleaseInputPlatformFirst'))
    return false
  }

  channelProductChecking.value = true
  try {
    const res = await catalogService.getProductDetail(productId)
    if (!res.data?.id) {
      ElMessage.error(t('payment.productNotFound'))
      return false
    }
    if (res.data.platformId !== platformId) {
      ElMessage.error(t('payment.productPlatformMismatch'))
      return false
    }
    return true
  } catch {
    ElMessage.error(t('payment.productNotFound'))
    return false
  } finally {
    channelProductChecking.value = false
  }
}

const validateAccountExists = async (accountId: number, tenantId: number, platformId: number) => {
  if (!accountId) {
    ElMessage.warning(t('payment.pleaseInputAccountId'))
    return false
  }
  if (!tenantId) {
    ElMessage.warning(t('payment.pleaseInputTenantFirst'))
    return false
  }
  if (!platformId) {
    ElMessage.warning(t('payment.pleaseInputPlatformFirst'))
    return false
  }

  channelAccountChecking.value = true
  try {
    const res = await tenantService.getTenantAccountDetail(accountId, tenantId)
    if (!res.data?.id) {
      ElMessage.error(t('payment.accountNotFound'))
      return false
    }
    if (res.data.platformId !== platformId) {
      ElMessage.error(t('payment.accountPlatformMismatch'))
      return false
    }
    return true
  } catch {
    ElMessage.error(t('payment.accountNotFound'))
    return false
  } finally {
    channelAccountChecking.value = false
  }
}

const handleChannelTenantChange = () => {
  channelTenantVerified.value = channelForm.tenantId > 0
  verifiedChannelTenantId.value = channelForm.tenantId
  channelAccountVerified.value = false
  verifiedChannelAccountId.value = 0
}

const handleChannelPlatformChange = () => {
  channelPlatformVerified.value = false
  verifiedChannelPlatformId.value = 0
  channelProductVerified.value = false
  verifiedChannelProductId.value = 0
  channelAccountVerified.value = false
  verifiedChannelAccountId.value = 0
}

const handleChannelProductChange = () => {
  channelProductVerified.value = false
  verifiedChannelProductId.value = 0
}

const handleChannelAccountChange = () => {
  channelAccountVerified.value = false
  verifiedChannelAccountId.value = 0
}

const checkChannelPlatform = async () => {
  const exists = await validatePlatformExists(channelForm.platformId)
  channelPlatformVerified.value = exists
  verifiedChannelPlatformId.value = exists ? channelForm.platformId : 0
  if (exists) ElMessage.success(t('payment.platformVerifiedSuccess'))
}

const checkChannelProduct = async () => {
  const exists = await validateProductExists(channelForm.productId, channelForm.platformId)
  channelProductVerified.value = exists
  verifiedChannelProductId.value = exists ? channelForm.productId : 0
  if (exists) ElMessage.success(t('payment.productVerifiedSuccess'))
}

const checkChannelAccount = async () => {
  const exists = await validateAccountExists(
    channelForm.accountId,
    channelForm.tenantId,
    channelForm.platformId,
  )
  channelAccountVerified.value = exists
  verifiedChannelAccountId.value = exists ? channelForm.accountId : 0
  if (exists) ElMessage.success(t('payment.accountVerifiedSuccess'))
}

const submitChannel = async () => {
  if (!channelForm.id && channelSubmitDisabled.value) {
    ElMessage.warning(t('payment.pleaseCompleteChannelValidation'))
    return
  }

  if (channelForm.id) {
    await tenantService.updateTenantChannel({ ...channelForm })
  } else {
    await tenantService.createTenantChannel({ ...channelForm })
  }
  ElMessage.success(t('common.operationSuccess'))
  channelDialogVisible.value = false
  loadChannels()
}

const showChannelDetail = async (row: TenantPayChannel) => {
  const res = await tenantService.getTenantChannelDetail(row.id, row.tenantId)
  detailData.value = res.data || row
  detailVisible.value = true
}

function getEnabledTagClass(value?: number) {
  const num = Number(value ?? 0)
  if (num === 1) return 'option-tag option-tag--green'
  if (num === 2) return 'option-tag option-tag--red'
  return 'option-tag option-tag--slate'
}

function handleLimitChange() {
  resetAndLoad(loadChannels)
}

function handlePrevPage() {
  prevAndLoad(loadChannels)
}

function handleNextPage() {
  nextAndLoad(loadChannels)
}

onMounted(async () => {
  await Promise.all([loadOptions(), loadChannels()])
})
</script>

<style scoped>
.query-card :deep(.el-form-item) {
  margin-bottom: 12px;
}

.verify-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.verified-text {
  color: var(--el-color-success);
  font-size: 14px;
}

.option-tag {
  border: none;
}

.option-tag--green {
  color: var(--el-color-success);
  background: var(--el-color-success-light-9);
}

.option-tag--red {
  color: var(--el-color-danger);
  background: var(--el-color-danger-light-9);
}

.option-tag--slate {
  color: var(--el-text-color-regular);
  background: var(--el-fill-color-light);
}
</style>
