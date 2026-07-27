<template>
  <div class="payment-page">
    <CrudQueryCard :model="channelQuery" @search="loadList" @reset="resetQuery">
      <el-form-item :label="t('common.tenantId')">
        <TenantSelect v-model="channelQuery.tenantId" class="tenant-select-filter" />
      </el-form-item>
      <el-form-item :label="t('payment.platformId')">
        <PayPlatformSelect
          v-model="channelQuery.platformId"
          :enabled-only="false"
          class="platform-select-filter"
        />
      </el-form-item>
      <el-form-item :label="t('common.keyword')">
        <el-input v-model="channelQuery.keyword" clearable />
      </el-form-item>
      <template #actions>
        <el-button
          v-perm="'payment:tenant-channel:add'"
          type="primary"
          @click="openChannelDialog()"
        >
          {{ t('common.add') }}
        </el-button>
      </template>
    </CrudQueryCard>

    <el-card shadow="never" class="table-card">
      <el-table v-loading="channelLoading" :data="channels" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column :label="t('common.tenantId')" min-width="170">
          <template #default="{ row }">
            {{ formatRelationLabel(row.tenantName, row.tenantId) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('payment.platformId')" min-width="180">
          <template #default="{ row }">
            {{ formatRelationLabel(row.platformName, row.platformId) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('payment.productId')" min-width="180">
          <template #default="{ row }">
            {{ formatRelationLabel(row.productName, row.productId) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('payment.accountId')" min-width="180">
          <template #default="{ row }">
            {{ formatRelationLabel(row.accountName, row.accountId) }}
          </template>
        </el-table-column>
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
        <el-table-column :label="t('common.actions')" align="center" width="160">
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
      width="1040px"
      class="channel-form-dialog"
    >
      <el-form
        ref="channelFormRef"
        :model="channelForm"
        :rules="channelFormRules"
        label-width="110px"
        class="channel-form-grid"
      >
        <el-form-item :label="t('common.tenantId')">
          <TenantSelect
            v-model="channelForm.tenantId"
            :disabled="!!channelForm.id"
            @change="handleChannelTenantChange"
          />
        </el-form-item>

        <el-form-item v-if="!channelForm.id" :label="t('payment.platformId')">
          <PayPlatformSelect
            v-model="channelForm.platformId"
            :clearable="false"
            @change="handleChannelPlatformChange"
          />
        </el-form-item>

        <el-form-item v-if="!channelForm.id" :label="t('payment.productId')">
          <PayProductSelect
            v-model="channelForm.productId"
            :platform-id="channelForm.platformId"
            :clearable="false"
            @change="handleChannelProductChange"
          />
        </el-form-item>

        <el-form-item v-if="!channelForm.id" :label="t('payment.accountId')">
          <TenantPayAccountSelect
            v-model="channelForm.accountId"
            :tenant-id="channelForm.tenantId"
            :platform-id="channelForm.platformId"
            :clearable="false"
            @change="handleChannelAccountChange"
          />
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
              v-for="item in enabledFormOptions"
              :key="item.value"
              :label="getOptionLabel(t, item.code, item.value)"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('common.visible')">
          <el-select v-model="channelForm.visible" style="width: 100%">
            <el-option
              v-for="item in visibleFormOptions"
              :key="item.value"
              :label="getOptionLabel(t, item.code, item.value)"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('payment.feeType')">
          <el-select v-model="channelForm.feeType" style="width: 100%">
            <el-option
              v-for="item in feeTypeFormOptions"
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
        <el-form-item
          :label="t('payment.extConfig')"
          prop="extConfig"
          class="channel-form-grid__full"
        >
          <el-input v-model="channelForm.extConfig" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item :label="t('common.remark')" class="channel-form-grid__full">
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
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { tenantService, type OptionGroup, type TenantPayChannel } from '@/services'
import { findFormOptionGroup, getOptionLabel, getOptionValueLabel } from '@/utils/options'
import TenantSelect from '@/components/TenantSelect.vue'
import PaymentDetailDescriptions from '@/components/payment/PaymentDetailDescriptions.vue'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'
import PayPlatformSelect from '@/components/payment/PayPlatformSelect.vue'
import PayProductSelect from '@/components/payment/PayProductSelect.vue'
import TenantPayAccountSelect from '@/components/payment/TenantPayAccountSelect.vue'

const { t } = useI18n()
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)

const channelLoading = ref(false)
const channels = ref<TenantPayChannel[]>([])
const detailVisible = ref(false)
const detailData = ref<TenantPayChannel | null>(null)
const channelDialogVisible = ref(false)

const optionGroups = ref<OptionGroup[]>([])
const enabledFormOptions = computed(() => findFormOptionGroup(optionGroups.value, 'enabled'))
const visibleFormOptions = computed(() => findFormOptionGroup(optionGroups.value, 'visible'))
const feeTypeFormOptions = computed(() => findFormOptionGroup(optionGroups.value, 'feeType'))

const channelQuery = reactive({
  tenantId: undefined as number | undefined,
  platformId: undefined as number | undefined,
  keyword: '',
})

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
const channelFormRef = ref<FormInstance>()
const channelFormRules: FormRules = {
  extConfig: [
    {
      validator: (_rule, value: string, callback) => {
        if (!value.trim()) {
          callback()
          return
        }
        try {
          JSON.parse(value)
          callback()
        } catch {
          callback(new Error(t('payment.invalidExtConfigJson')))
        }
      },
      trigger: 'blur',
    },
  ],
}

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

const loadList = async () => {
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

function resetQuery() {
  channelQuery.tenantId = 0
  channelQuery.platformId = 0
  channelQuery.keyword = ''
  resetAndLoad(loadList)
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

const handleChannelTenantChange = () => {
  channelTenantVerified.value = channelForm.tenantId > 0
  verifiedChannelTenantId.value = channelForm.tenantId
  channelForm.accountId = 0
  channelAccountVerified.value = false
  verifiedChannelAccountId.value = 0
}

const handleChannelPlatformChange = () => {
  channelPlatformVerified.value = channelForm.platformId > 0
  verifiedChannelPlatformId.value = channelForm.platformId
  channelForm.productId = 0
  channelForm.accountId = 0
  channelProductVerified.value = false
  verifiedChannelProductId.value = 0
  channelAccountVerified.value = false
  verifiedChannelAccountId.value = 0
}

const handleChannelProductChange = () => {
  channelProductVerified.value = channelForm.productId > 0
  verifiedChannelProductId.value = channelForm.productId
}

const handleChannelAccountChange = () => {
  channelAccountVerified.value = channelForm.accountId > 0
  verifiedChannelAccountId.value = channelForm.accountId
}

const submitChannel = async () => {
  const valid = await channelFormRef.value?.validate().catch(() => false)
  if (!valid) return

  if (!channelForm.id && channelSubmitDisabled.value) {
    ElMessage.warning(t('payment.pleaseCompleteChannelValidation'))
    return
  }

  const payload = {
    ...channelForm,
    singleMinAmount: String(channelForm.singleMinAmount),
    singleMaxAmount: String(channelForm.singleMaxAmount),
    dailyMaxAmount: String(channelForm.dailyMaxAmount),
    feeFixedAmount: String(channelForm.feeFixedAmount),
  }
  if (channelForm.id) {
    await tenantService.updateTenantChannel(payload)
  } else {
    await tenantService.createTenantChannel(payload)
  }
  ElMessage.success(t('common.operationSuccess'))
  channelDialogVisible.value = false
  loadList()
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
  resetAndLoad(loadList)
}

function handlePrevPage() {
  prevAndLoad(loadList)
}

function handleNextPage() {
  nextAndLoad(loadList)
}

onMounted(async () => {
  await Promise.all([loadOptions(), loadList()])
})

function formatRelationLabel(name: string | undefined, id: number) {
  return name ? `${name} (${id})` : String(id)
}
</script>

<style scoped>
.channel-form-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  column-gap: 24px;
}

.channel-form-grid__full {
  grid-column: 1 / -1;
}

.verify-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.verify-row .el-input-number {
  flex: 1;
  min-width: 0;
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

@media (max-width: 900px) {
  .channel-form-grid {
    grid-template-columns: minmax(0, 1fr);
  }

  .channel-form-grid__full {
    grid-column: auto;
  }
}
</style>
