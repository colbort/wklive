<template>
  <div class="payment-page">
    <CrudQueryCard :model="query" @search="loadList" @reset="resetQuery">
      <el-form-item :label="t('common.tenantId')">
        <TenantSelect v-model="query.tenantId" class="tenant-select-filter" />
      </el-form-item>
      <el-form-item :label="t('payment.platformId')">
        <PayPlatformSelect
          v-model="query.platformId"
          :enabled-only="false"
          class="platform-select-filter"
        />
      </el-form-item>
      <el-form-item :label="t('common.keyword')">
        <el-input v-model="query.keyword" clearable />
      </el-form-item>

      <template #actions>
        <el-button v-perm="'payment:tenant-account:add'" type="primary" @click="openDialog()">
          {{ t('common.add') }}
        </el-button>
      </template>
    </CrudQueryCard>

    <el-card shadow="never" class="table-card">
      <el-table v-loading="loading" :data="list" stripe>
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
        <el-table-column prop="accountCode" :label="t('payment.accountCode')" min-width="140" />
        <el-table-column prop="accountName" :label="t('payment.accountName')" min-width="160" />
        <el-table-column prop="merchantId" :label="t('payment.merchantId')" min-width="140" />
        <el-table-column :label="t('common.enabled')" width="100">
          <template #default="{ row }">
            <el-tag :class="getEnabledTagClass(row.enabled)" disable-transitions>
              {{ getOptionValueLabel(optionGroups, 'enabled', row.enabled, t) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="t('common.default')" width="90">
          <template #default="{ row }">
            {{ getOptionValueLabel(optionGroups, 'yesNo', row.isDefault, t) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('common.actions')" align="center" width="160">
          <template #default="{ row }">
            <el-button
              v-perm="'payment:tenant-account:detail'"
              link
              type="primary"
              @click="showDetail(row)"
            >
              {{ t('common.detail') }}
            </el-button>
            <el-button
              v-perm="'payment:tenant-account:update'"
              link
              type="primary"
              @click="openDialog(row)"
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
      v-model="dialogVisible"
      :title="form.id ? t('payment.editAccount') : t('payment.addAccount')"
      width="1040px"
      class="account-form-dialog"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="formRules"
        label-width="120px"
        class="account-form-grid"
      >
        <el-form-item :label="t('common.tenantId')" prop="tenantId">
          <TenantSelect
            v-model="form.tenantId"
            :disabled="!!form.id"
            @change="handleTenantChange"
          />
        </el-form-item>

        <el-form-item v-if="!form.id" :label="t('payment.platformId')" prop="platformId">
          <PayPlatformSelect
            v-model="form.platformId"
            :clearable="false"
            @change="handlePlatformChange"
          />
        </el-form-item>

        <el-form-item v-if="!form.id" :label="t('payment.accountCode')" prop="accountCode">
          <el-input v-model="form.accountCode" />
        </el-form-item>
        <el-form-item :label="t('payment.accountName')">
          <el-input v-model="form.accountName" />
        </el-form-item>
        <el-form-item label="APP ID">
          <el-input v-model="form.appId" />
        </el-form-item>
        <el-form-item :label="t('payment.merchantId')" prop="merchantId">
          <el-input v-model="form.merchantId" />
        </el-form-item>
        <el-form-item :label="t('payment.merchantName')" prop="merchantName">
          <el-input v-model="form.merchantName" />
        </el-form-item>
        <el-form-item :label="t('common.enabled')">
          <el-select v-model="form.enabled" style="width: 100%">
            <el-option
              v-for="item in enabledFormOptions"
              :key="item.value"
              :label="getOptionLabel(t, item.code, item.value)"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('common.default')">
          <el-select v-model="form.isDefault" style="width: 100%">
            <el-option
              v-for="item in yesNoFormOptions"
              :key="item.value"
              :label="getOptionLabel(t, item.code, item.value)"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <div class="account-secret-grid account-form-grid__full">
          <el-form-item :label="t('payment.apiKeyCipher')">
            <el-input v-model="form.apiKeyCipher" type="textarea" :rows="2" />
          </el-form-item>
          <el-form-item :label="t('payment.apiSecretCipher')">
            <el-input v-model="form.apiSecretCipher" type="textarea" :rows="2" />
          </el-form-item>
          <el-form-item :label="t('payment.privateKeyCipher')">
            <el-input v-model="form.privateKeyCipher" type="textarea" :rows="3" />
          </el-form-item>
          <el-form-item :label="t('payment.publicKey')">
            <el-input v-model="form.publicKey" type="textarea" :rows="3" />
          </el-form-item>
        </div>
        <el-form-item :label="t('payment.certCipher')" class="account-form-grid__full">
          <el-input v-model="form.certCipher" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item
          :label="t('payment.extConfig')"
          prop="extConfig"
          class="account-form-grid__full"
        >
          <el-input v-model="form.extConfig" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item :label="t('common.remark')" class="account-form-grid__full">
          <el-input v-model="form.remark" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">
          {{ t('common.cancel') }}
        </el-button>
        <el-button
          v-perm="form.id ? 'payment:tenant-account:update' : 'payment:tenant-account:add'"
          type="primary"
          :disabled="submitDisabled"
          @click="submit"
        >
          {{ t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="detailVisible" :title="t('payment.detailTitle')" width="760px">
      <PaymentDetailDescriptions :data="detailData" :option-groups="optionGroups" />
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { usePagination } from '@/composables'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { tenantService, type OptionGroup, type TenantPayAccount } from '@/services'
import { findFormOptionGroup, getOptionLabel, getOptionValueLabel } from '@/utils/options'
import TenantSelect from '@/components/TenantSelect.vue'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'
import PaymentDetailDescriptions from '@/components/payment/PaymentDetailDescriptions.vue'
import PayPlatformSelect from '@/components/payment/PayPlatformSelect.vue'

const { t } = useI18n()
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)

const loading = ref(false)
const list = ref<TenantPayAccount[]>([])
const dialogVisible = ref(false)
const detailVisible = ref(false)
const detailData = ref<Record<string, unknown>>({})

const optionGroups = ref<OptionGroup[]>([])
const enabledFormOptions = computed(() => findFormOptionGroup(optionGroups.value, 'enabled'))
const yesNoFormOptions = computed(() => findFormOptionGroup(optionGroups.value, 'yesNo'))

const tenantVerified = ref(false)
const platformVerified = ref(false)
const verifiedTenantId = ref(0)
const verifiedPlatformId = ref(0)

const query = reactive({
  tenantId: undefined as number | undefined,
  platformId: undefined as number | undefined,
  keyword: '',
})

const createEmptyForm = () => ({
  id: 0,
  tenantId: 0,
  platformId: 0,
  accountCode: '',
  accountName: '',
  appId: '',
  merchantId: '',
  merchantName: '',
  apiKeyCipher: '',
  apiSecretCipher: '',
  privateKeyCipher: '',
  publicKey: '',
  certCipher: '',
  extConfig: '',
  enabled: 1,
  isDefault: 2,
  remark: '',
})

const form = reactive(createEmptyForm())
const formRef = ref<FormInstance>()
const formRules: FormRules = {
  tenantId: [
    {
      required: true,
      type: 'number',
      min: 1,
      message: t('payment.pleaseInputTenantId'),
      trigger: 'change',
    },
  ],
  platformId: [
    {
      required: true,
      type: 'number',
      min: 1,
      message: t('payment.pleaseInputPlatformId'),
      trigger: 'change',
    },
  ],
  accountCode: [
    {
      required: true,
      whitespace: true,
      message: t('payment.pleaseInputAccountCode'),
      trigger: 'blur',
    },
  ],
  merchantId: [
    {
      required: true,
      whitespace: true,
      message: t('payment.pleaseInputMerchantId'),
      trigger: 'blur',
    },
  ],
  merchantName: [
    {
      required: true,
      whitespace: true,
      message: t('payment.pleaseInputMerchantName'),
      trigger: 'blur',
    },
  ],
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

const submitDisabled = computed(
  () =>
    !form.id &&
    (!tenantVerified.value ||
      !platformVerified.value ||
      verifiedTenantId.value !== form.tenantId ||
      verifiedPlatformId.value !== form.platformId),
)

const loadOptions = async () => {
  const res = await tenantService.getOptions()
  optionGroups.value = res.data || []
}

const loadList = async () => {
  loading.value = true
  try {
    const res = await tenantService.getTenantAccountList({
      ...query,
      tenantId: query.tenantId || undefined,
      platformId: query.platformId || undefined,
      keyword: query.keyword || undefined,
      cursor: pagination.cursor,
      limit: pagination.limit,
    })
    list.value = res.data || []
    updateFromResponse(res)
  } finally {
    loading.value = false
  }
}

function resetQuery() {
  query.tenantId = undefined
  query.platformId = undefined
  query.keyword = ''
  resetAndLoad(loadList)
}

const resetVerifyState = () => {
  tenantVerified.value = false
  platformVerified.value = false
  verifiedTenantId.value = 0
  verifiedPlatformId.value = 0
}

const openDialog = (row?: TenantPayAccount) => {
  Object.assign(form, createEmptyForm(), row || {})
  if (row?.id) {
    tenantVerified.value = true
    platformVerified.value = true
    verifiedTenantId.value = row.tenantId
    verifiedPlatformId.value = row.platformId
  } else {
    resetVerifyState()
  }
  dialogVisible.value = true
}

const handleTenantChange = () => {
  tenantVerified.value = form.tenantId > 0
  verifiedTenantId.value = form.tenantId
}

const handlePlatformChange = () => {
  platformVerified.value = form.platformId > 0
  verifiedPlatformId.value = form.platformId
}

const submit = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  if (!form.id && submitDisabled.value) {
    ElMessage.warning(t('payment.pleaseCompleteAccountValidation'))
    return
  }

  if (form.id) {
    await tenantService.updateTenantAccount({ ...form })
  } else {
    await tenantService.createTenantAccount({ ...form })
  }
  ElMessage.success(t('common.operationSuccess'))
  dialogVisible.value = false
  loadList()
}

const showDetail = async (row: TenantPayAccount) => {
  const res = await tenantService.getTenantAccountDetail(row.id, row.tenantId)
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
.account-form-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  column-gap: 24px;
}

.account-form-grid__full {
  grid-column: 1 / -1;
}

.account-secret-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  column-gap: 24px;
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

@media (max-width: 1200px) {
  .account-form-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 900px) {
  .account-form-grid {
    grid-template-columns: minmax(0, 1fr);
  }

  .account-form-grid__full {
    grid-column: auto;
  }

  .account-secret-grid {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
