<template>
  <div class="sys-config module-page">
    <div class="page-header">
      <h2>{{ t('system.config') }}</h2>
      <el-button v-perm="'sys:config:add'" type="primary" @click="handleCreate">
        <el-icon><Plus /></el-icon>
        {{ t('common.add') }}
      </el-button>
    </div>

    <el-card class="query-card" shadow="never">
      <el-form :model="queryForm" inline>
        <el-form-item :label="t('common.tenantId')">
          <el-input-number
            v-model="queryForm.tenantId"
            :min="0"
            :precision="0"
            @change="fetchList"
          />
        </el-form-item>
        <el-form-item :label="t('system.configKey')">
          <el-select
            v-model="queryForm.keyword"
            :placeholder="t('system.pleaseSelect')"
            filterable
            clearable
            @change="fetchList"
          >
            <el-option
              v-for="key in keys"
              :key="key.code"
              :label="getOptionLabel(t, key.code, key.value)"
              :value="key.code"
            />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="fetchList">
            <el-icon><Search /></el-icon>
            {{ t('common.search') }}
          </el-button>
          <el-button @click="handleReset">
            <el-icon><Refresh /></el-icon>
            {{ t('common.reset') }}
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card class="table-card" shadow="never">
      <el-table
        v-loading="loading"
        :data="list"
        :empty-text="t('common.noData')"
        stripe
      >
        <el-table-column
          prop="id"
          :label="t('common.id')"
          width="80"
          align="center"
        />
        <el-table-column
          prop="tenantId"
          :label="t('common.tenantId')"
          width="100"
          align="center"
        />
        <el-table-column prop="configKey" :label="t('system.configKey')" min-width="150" />
        <el-table-column prop="configValue" :label="t('system.configValue')" min-width="200">
          <template #default="{ row }">
            <el-tooltip :content="row.configValue" placement="top" popper-class="config-tip">
              <span class="config-value">{{ row.configValue }}</span>
            </el-tooltip>
          </template>
        </el-table-column>
        <el-table-column prop="remark" :label="t('common.remark')" min-width="150" />
        <el-table-column
          prop="createTimes"
          :label="t('common.createTimes')"
          width="160"
          align="center"
        >
          <template #default="{ row }">
            {{ formatDate(row.createTimes) }}
          </template>
        </el-table-column>
        <el-table-column
          prop="updateTimes"
          :label="t('common.updateTimes')"
          width="160"
          align="center"
        >
          <template #default="{ row }">
            {{ formatDate(row.updateTimes) }}
          </template>
        </el-table-column>
        <el-table-column
          :label="t('common.actions')"
          width="150"
          align="center"
          fixed="right"
        >
          <template #default="{ row }">
            <el-button
              v-perm="'sys:config:update'"
              type="primary"
              size="small"
              @click="handleEdit(row)"
            >
              {{ t('common.edit') }}
            </el-button>
            <el-button
              v-perm="'sys:config:delete'"
              type="danger"
              size="small"
              @click="handleDelete(row)"
            >
              {{ t('common.delete') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <CursorPagination
        v-model:limit="pagination.limit"
        :total="pagination.total"
        :has-prev="pagination.hasPrev"
        :has-next="pagination.hasNext"
        @prev="prevPage"
        @next="nextPage"
        @limit-change="
          () => {
            resetAndLoad(fetchList)
          }
        "
      />
    </el-card>

    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? t('common.edit') : t('common.add')"
      width="860px"
      :close-on-click-modal="false"
      class="sys-config-dialog"
    >
      <el-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        label-width="160px"
      >
        <el-form-item :label="t('common.tenantId')" prop="tenantId">
          <TenantSelect
            v-model="formData.tenantId"
            include-system
            :disabled="isEdit"
            @change="handleTenantChange"
          />
          <div class="verify-row">
            <span v-if="tenantVerified" class="verified-text">
              {{ formData.tenantId === 0 ? t('system.systemConfigScope') : t('payment.verified') }}
            </span>
          </div>
        </el-form-item>
        <el-form-item :label="t('system.configKey')" prop="configKey">
          <el-select
            v-if="!isEdit"
            v-model="formData.configKey"
            :placeholder="t('system.pleaseSelect')"
            filterable
            clearable
            @change="handleConfigKeyChange"
          >
            <el-option
              v-for="key in keys"
              :key="key.code"
              :label="getOptionLabel(t, key.code, key.value)"
              :value="key.code"
            />
          </el-select>
          <el-input
            v-else
            :model-value="t('options.' + formData.configKey)"
            :placeholder="t('common.pleaseEnter')"
            disabled
          />
        </el-form-item>
        <template v-if="formData.configKey === 'SYSTEM_CORE'">
          <SystemCoreConfigComponent v-model="systemCoreForm" />
        </template>

        <template v-else-if="formData.configKey === 'OBJECT_STORAGE'">
          <ObjectStorageConfigComponent v-model="objectStorageForm" />
        </template>

        <template v-else-if="formData.configKey === 'ITICK_CONFIG'">
          <ItickConfigComponent v-model="itickConfigForm" />
        </template>

        <template v-else-if="formData.configKey === 'RECHARGE_CONFIG'">
          <RechargeConfigComponent v-model="rechargeConfigForm" />
        </template>

        <template v-else-if="formData.configKey === 'WITHDRAW_CONFIG'">
          <WithdrawConfigComponent v-model="withdrawConfigForm" />
        </template>

        <template v-else-if="formData.configKey === 'EMAIL_CONFIG'">
          <EmailConfigComponent v-model="emailConfigForm" />
        </template>

        <template v-else-if="formData.configKey === 'PHONE_CONFIG'">
          <PhoneConfigComponent v-model="phoneConfigForm" />
        </template>

        <template v-else>
          <el-form-item :label="t('system.configValue')" prop="configValue">
            <el-input
              v-model="formData.configValue"
              type="textarea"
              :rows="4"
              :placeholder="t('common.pleaseEnter')"
            />
          </el-form-item>
        </template>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">
          {{ t('common.cancel') }}
        </el-button>
        <el-button
          v-perm="isEdit ? 'sys:config:update' : 'sys:config:add'"
          type="primary"
          :loading="submitLoading"
          @click="handleSubmit"
        >
          {{ t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { Plus, Search, Refresh } from '@element-plus/icons-vue'
import { configService } from '@/services'
import type { SysConfigItem, SysConfigCreateReq, OptionItem } from '@/services'
import type {
  SystemCore,
  ObjectStorageConfig,
  ItickConfig,
  RechargeConfig,
  WithdrawConfig,
  EmailConfig,
  PhoneConfig,
} from '@/services/system/ConfigService'
import { usePagination } from '@/composables/usePagination'
import { useLoading } from '@/composables/useLoading'
import { useForm } from '@/composables/useForm'
import { formatDate } from '@/utils'
import SystemCoreConfigComponent from './components/SystemCoreConfig.vue'
import ObjectStorageConfigComponent from './components/ObjectStorageConfig.vue'
import ItickConfigComponent from './components/ItickConfig.vue'
import RechargeConfigComponent from './components/RechargeConfig.vue'
import WithdrawConfigComponent from './components/WithdrawConfig.vue'
import EmailConfigComponent from './components/EmailConfig.vue'
import PhoneConfigComponent from './components/PhoneConfig.vue'
import { getOptionLabel } from '@/utils/options'
import TenantSelect from '@/components/TenantSelect.vue'

const { t } = useI18n()

// Pagination and main list
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(10)
const list = ref<SysConfigItem[]>([])
const { loading, withLoading } = useLoading()

// Query form
const { form: queryForm } = useForm({
  initialData: {
    tenantId: 0,
    keyword: '',
  },
})

// Dialog and form
const dialogVisible = ref(false)
const isEdit = ref(false)
const submitLoading = ref(false)
const tenantVerified = ref(true)
const verifiedTenantId = ref(0)
const formRef = ref()

const { form: formData, reset: resetForm } = useForm({
  initialData: {
    id: 0,
    tenantId: 0,
    configKey: '',
    configValue: '',
    remark: '',
  },
})

// Keys for configKey selection
const keys = ref<OptionItem[]>([])

// Config type special forms
const systemCoreForm = ref<SystemCore>({
  site_name: '',
  site_logo: '',
  is_captcha_enabled: 2,
  is_register_enabled: 2,
  is_guest_enabled: 2,
  is_crypto_enabled: 2,
})

const activeTab = ref('aliyun')

const objectStorageForm = ref<ObjectStorageConfig>({
  aliyun_oss: {
    endpoint: '',
    access_key_id: '',
    access_key_secret: '',
    bucket_name: '',
    bucket_url: '',
  },
  tencent_cos: {
    market: '',
    secret_id: '',
    secret_key: '',
    bucket_name: '',
    bucket_url: '',
  },
  minio: {
    endpoint: '',
    access_key_id: '',
    access_key_secret: '',
    bucket_name: '',
    bucket_url: '',
  },
  oss_type: 1,
  oss_domain: '',
})

const itickConfigForm = ref<ItickConfig>({
  api_url: '',
  api_token: '',
  ws_url: '',
})

const rechargeConfigForm = ref<RechargeConfig>({
  minAmount: 0,
  maxAmount: 0,
  feeRate: 0,
})

const withdrawConfigForm = ref<WithdrawConfig>({
  minAmount: 0,
  maxAmount: 0,
  feeRate: 0,
  dailyLimitPerUser: 0,
  dailyAmountLimitPerUser: 0,
  allowedTimeRange: '',
  pendingWithdrawalLimitPerUser: 0,
  freeWithdrawTimesPerDay: 0,
})

const emailConfigForm = ref<EmailConfig>({
  enabled: 2,
  smtp_host: '',
  smtp_port: 587,
  username: '',
  password: '',
  from_email: '',
  from_name: '',
  subject_template: 'Verification code',
  body_template: 'Your verification code is {{code}}',
})

const phoneConfigForm = ref<PhoneConfig>({
  enabled: 2,
  provider: '',
  endpoint: '',
  method: 'POST',
  headers_json: '{}',
  body_template: '{"phone":"{{phone}}","code":"{{code}}","scene":"{{scene}}"}',
})

const configNumber = (data: Record<string, unknown>, camelKey: string, snakeKey: string) =>
  Number(data[camelKey] ?? data[snakeKey] ?? 0)

function enableValue(value: unknown) {
  if (value === true || value === 1 || value === '1') return 1
  if (value === false || value === 2 || value === '2') return 2
  return 2
}

// Form validation rules
const formRules = {
  configKey: [{ required: true, message: t('validation.required'), trigger: 'blur' }],
  configValue: [{ required: true, message: t('validation.required'), trigger: 'blur' }],
}

// Load available keys
async function loadKeys() {
  try {
    const res = await configService.getKeys()
    if (res.code !== 200) throw new Error(res.msg || 'Failed to load keys')
    keys.value = res.data || []
  } catch (error: unknown) {
    ElMessage.error(error instanceof Error ? error.message : 'Failed to load keys')
  }
}

// Fetch list
async function fetchList() {
  await withLoading(async () => {
    try {
      const res = await configService.getList({
        tenantId: queryForm.tenantId,
        keyword: queryForm.keyword || undefined,
        cursor: pagination.cursor,
        limit: pagination.limit,
      })
      if (res.code !== 200) throw new Error(res.msg || 'list failed')
      list.value = res.data || []
      updateFromResponse(res)
    } catch (error: unknown) {
      ElMessage.error(error instanceof Error ? error.message : t('common.loadFailed'))
    }
  })
}

// Handle reset
function handleReset() {
  queryForm.tenantId = 0
  queryForm.keyword = ''
  resetAndLoad(fetchList)
}

function resetTypeForms() {
  systemCoreForm.value = {
    site_name: '',
    site_logo: '',
    is_captcha_enabled: 2,
    is_register_enabled: 2,
    is_guest_enabled: 2,
    is_crypto_enabled: 2,
  }
  objectStorageForm.value = {
    aliyun_oss: {
      endpoint: '',
      access_key_id: '',
      access_key_secret: '',
      bucket_name: '',
      bucket_url: '',
    },
    tencent_cos: {
      market: '',
      secret_id: '',
      secret_key: '',
      bucket_name: '',
      bucket_url: '',
    },
    minio: {
      endpoint: '',
      access_key_id: '',
      access_key_secret: '',
      bucket_name: '',
      bucket_url: '',
    },
    oss_type: 1,
    oss_domain: '',
  }
  itickConfigForm.value = {
    api_url: '',
    api_token: '',
    ws_url: '',
  }
  rechargeConfigForm.value = {
    minAmount: 0,
    maxAmount: 0,
    feeRate: 0,
  }
  withdrawConfigForm.value = {
    minAmount: 0,
    maxAmount: 0,
    feeRate: 0,
    dailyLimitPerUser: 0,
    dailyAmountLimitPerUser: 0,
    allowedTimeRange: '',
    pendingWithdrawalLimitPerUser: 0,
    freeWithdrawTimesPerDay: 0,
  }
  emailConfigForm.value = {
    enabled: 2,
    smtp_host: '',
    smtp_port: 587,
    username: '',
    password: '',
    from_email: '',
    from_name: '',
    subject_template: 'Verification code',
    body_template: 'Your verification code is {{code}}',
  }
  phoneConfigForm.value = {
    enabled: 2,
    provider: '',
    endpoint: '',
    method: 'POST',
    headers_json: '{}',
    body_template: '{"phone":"{{phone}}","code":"{{code}}","scene":"{{scene}}"}',
  }
}

function handleConfigKeyChange(value: string) {
  if (value === 'SYSTEM_CORE') {
    systemCoreForm.value = {
      site_name: '',
      site_logo: '',
      is_captcha_enabled: 2,
      is_register_enabled: 2,
      is_guest_enabled: 2,
      is_crypto_enabled: 2,
    }
    formData.configValue = ''
  } else if (value === 'OBJECT_STORAGE') {
    objectStorageForm.value = {
      aliyun_oss: {
        endpoint: '',
        access_key_id: '',
        access_key_secret: '',
        bucket_name: '',
        bucket_url: '',
      },
      tencent_cos: {
        market: '',
        secret_id: '',
        secret_key: '',
        bucket_name: '',
        bucket_url: '',
      },
      minio: {
        endpoint: '',
        access_key_id: '',
        access_key_secret: '',
        bucket_name: '',
        bucket_url: '',
      },
      oss_type: 1,
      oss_domain: '',
    }
    formData.configValue = ''
  } else if (value === 'ITICK_CONFIG') {
    itickConfigForm.value = {
      api_url: '',
      api_token: '',
      ws_url: '',
    }
    formData.configValue = ''
  } else if (value === 'RECHARGE_CONFIG') {
    rechargeConfigForm.value = {
      minAmount: 0,
      maxAmount: 0,
      feeRate: 0,
    }
    formData.configValue = ''
  } else if (value === 'WITHDRAW_CONFIG') {
    withdrawConfigForm.value = {
      minAmount: 0,
      maxAmount: 0,
      feeRate: 0,
      dailyLimitPerUser: 0,
      dailyAmountLimitPerUser: 0,
      allowedTimeRange: '',
      pendingWithdrawalLimitPerUser: 0,
      freeWithdrawTimesPerDay: 0,
    }
    formData.configValue = ''
  } else if (value === 'EMAIL_CONFIG') {
    emailConfigForm.value = {
      enabled: 2,
      smtp_host: '',
      smtp_port: 587,
      username: '',
      password: '',
      from_email: '',
      from_name: '',
      subject_template: 'Verification code',
      body_template: 'Your verification code is {{code}}',
    }
    formData.configValue = ''
  } else if (value === 'PHONE_CONFIG') {
    phoneConfigForm.value = {
      enabled: 2,
      provider: '',
      endpoint: '',
      method: 'POST',
      headers_json: '{}',
      body_template: '{"phone":"{{phone}}","code":"{{code}}","scene":"{{scene}}"}',
    }
    formData.configValue = ''
  }
}

function nextPage() {
  nextAndLoad(fetchList)
}

function prevPage() {
  prevAndLoad(fetchList)
}

// Handle create
function handleCreate() {
  isEdit.value = false
  resetForm()
  resetTypeForms()
  handleTenantChange()
  loadKeys()
  dialogVisible.value = true
}

// Handle edit
function handleEdit(row: SysConfigItem) {
  isEdit.value = true
  resetForm()
  resetTypeForms()
  Object.assign(formData, {
    id: row.id,
    tenantId: row.tenantId || 0,
    configKey: row.configKey,
    remark: row.remark || '',
  })
  tenantVerified.value = true
  verifiedTenantId.value = row.tenantId || 0

  if (row.configKey === 'SYSTEM_CORE') {
    try {
      const parsed = JSON.parse(row.configValue || '{}')
      systemCoreForm.value = {
        site_name: parsed.site_name || '',
        site_logo: parsed.site_logo || '',
        is_captcha_enabled: enableValue(parsed.is_captcha_enabled),
        is_register_enabled: enableValue(parsed.is_register_enabled),
        is_guest_enabled: enableValue(parsed.is_guest_enabled),
        is_crypto_enabled: enableValue(parsed.is_crypto_enabled),
      }
    } catch {
      systemCoreForm.value = {
        site_name: '',
        site_logo: '',
        is_captcha_enabled: 2,
        is_register_enabled: 2,
        is_guest_enabled: 2,
        is_crypto_enabled: 2,
      }
    }
  } else if (row.configKey === 'OBJECT_STORAGE') {
    try {
      const parsed = JSON.parse(row.configValue || '{}')
      objectStorageForm.value = {
        aliyun_oss: parsed.aliyun_oss || objectStorageForm.value.aliyun_oss,
        tencent_cos: parsed.tencent_cos || objectStorageForm.value.tencent_cos,
        minio: parsed.minio || objectStorageForm.value.minio,
        oss_type: parsed.oss_type || 1,
        oss_domain: parsed.oss_domain || '',
      }
      if (parsed.oss_type === 1) activeTab.value = 'aliyun'
      else if (parsed.oss_type === 2) activeTab.value = 'tencent'
      else activeTab.value = 'minio'
    } catch {
      resetTypeForms()
    }
  } else if (row.configKey === 'ITICK_CONFIG') {
    try {
      const parsed = JSON.parse(row.configValue || '{}')
      itickConfigForm.value = {
        api_url: parsed.api_url || '',
        api_token: parsed.api_token || '',
        ws_url: parsed.ws_url || '',
      }
    } catch {
      itickConfigForm.value = {
        api_url: '',
        api_token: '',
        ws_url: '',
      }
    }
  } else if (row.configKey === 'RECHARGE_CONFIG') {
    try {
      const parsed = JSON.parse(row.configValue || '{}') as Record<string, unknown>
      rechargeConfigForm.value = {
        minAmount: configNumber(parsed, 'minAmount', 'min_amount'),
        maxAmount: configNumber(parsed, 'maxAmount', 'max_amount'),
        feeRate: configNumber(parsed, 'feeRate', 'fee_rate'),
      }
    } catch {
      rechargeConfigForm.value = {
        minAmount: 0,
        maxAmount: 0,
        feeRate: 0,
      }
    }
  } else if (row.configKey === 'WITHDRAW_CONFIG') {
    try {
      const parsed = JSON.parse(row.configValue || '{}') as Record<string, unknown>
      withdrawConfigForm.value = {
        minAmount: configNumber(parsed, 'minAmount', 'min_amount'),
        maxAmount: configNumber(parsed, 'maxAmount', 'max_amount'),
        feeRate: configNumber(parsed, 'feeRate', 'fee_rate'),
        dailyLimitPerUser: configNumber(parsed, 'dailyLimitPerUser', 'daily_limit_per_user'),
        dailyAmountLimitPerUser: configNumber(
          parsed,
          'dailyAmountLimitPerUser',
          'daily_amount_limit_per_user',
        ),
        allowedTimeRange: String(parsed.allowedTimeRange ?? parsed.allowed_time_range ?? ''),
        pendingWithdrawalLimitPerUser: configNumber(
          parsed,
          'pendingWithdrawalLimitPerUser',
          'pending_withdrawal_limit_per_user',
        ),
        freeWithdrawTimesPerDay: configNumber(
          parsed,
          'freeWithdrawTimesPerDay',
          'free_withdraw_times_per_day',
        ),
      }
    } catch {
      withdrawConfigForm.value = {
        minAmount: 0,
        maxAmount: 0,
        feeRate: 0,
        dailyLimitPerUser: 0,
        dailyAmountLimitPerUser: 0,
        allowedTimeRange: '',
        pendingWithdrawalLimitPerUser: 0,
        freeWithdrawTimesPerDay: 0,
      }
    }
  } else if (row.configKey === 'EMAIL_CONFIG') {
    try {
      const parsed = JSON.parse(row.configValue || '{}')
      emailConfigForm.value = {
        enabled: enableValue(parsed.enabled),
        smtp_host: parsed.smtp_host || '',
        smtp_port: Number(parsed.smtp_port || 587),
        username: parsed.username || '',
        password: parsed.password || '',
        from_email: parsed.from_email || '',
        from_name: parsed.from_name || '',
        subject_template: parsed.subject_template || 'Verification code',
        body_template: parsed.body_template || 'Your verification code is {{code}}',
      }
    } catch {
      resetTypeForms()
    }
  } else if (row.configKey === 'PHONE_CONFIG') {
    try {
      const parsed = JSON.parse(row.configValue || '{}')
      phoneConfigForm.value = {
        enabled: enableValue(parsed.enabled),
        provider: parsed.provider || '',
        endpoint: parsed.endpoint || '',
        method: parsed.method || 'POST',
        headers_json: parsed.headers_json || '{}',
        body_template:
          parsed.body_template || '{"phone":"{{phone}}","code":"{{code}}","scene":"{{scene}}"}',
      }
    } catch {
      resetTypeForms()
    }
  } else {
    formData.configValue = row.configValue
  }

  dialogVisible.value = true
}

// Handle delete
async function handleDelete(row: SysConfigItem) {
  try {
    await ElMessageBox.confirm(t('common.confirmDelete'), t('common.warning'), {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      type: 'warning',
    })

    const res = await configService.delete(row.id)
    if (res.code !== 200) throw new Error(res.msg || 'delete failed')

    ElMessage.success(t('common.deleteSuccess'))
    fetchList()
  } catch (error: unknown) {
    if ((error instanceof Error ? error.message : '') !== 'cancel') {
      ElMessage.error(error instanceof Error ? error.message : t('common.deleteFailed'))
    }
  }
}

function handleTenantChange() {
  if (formData.tenantId === 0) {
    tenantVerified.value = true
    verifiedTenantId.value = 0
    return
  }
  tenantVerified.value = formData.tenantId > 0
  verifiedTenantId.value = formData.tenantId
}

async function ensureTenantVerified() {
  if (formData.tenantId === 0) return true
  if (tenantVerified.value && verifiedTenantId.value === formData.tenantId) return true
  return false
}

// Handle submit
async function handleSubmit() {
  if (!formRef.value) return

  try {
    await formRef.value.validate()
    const tenantOk = await ensureTenantVerified()
    if (!tenantOk) return

    submitLoading.value = true

    if (formData.configKey === 'SYSTEM_CORE') {
      if (!systemCoreForm.value.site_name) {
        throw new Error(t('validation.required'))
      }
      formData.configValue = JSON.stringify(systemCoreForm.value)
    } else if (formData.configKey === 'OBJECT_STORAGE') {
      if (!objectStorageForm.value.oss_domain) {
        throw new Error(t('validation.required'))
      }
      formData.configValue = JSON.stringify(objectStorageForm.value)
    } else if (formData.configKey === 'ITICK_CONFIG') {
      if (
        !itickConfigForm.value.api_url ||
        !itickConfigForm.value.api_token ||
        !itickConfigForm.value.ws_url
      ) {
        throw new Error(t('validation.required'))
      }
      formData.configValue = JSON.stringify(itickConfigForm.value)
    } else if (formData.configKey === 'RECHARGE_CONFIG') {
      formData.configValue = JSON.stringify(rechargeConfigForm.value)
    } else if (formData.configKey === 'WITHDRAW_CONFIG') {
      formData.configValue = JSON.stringify(withdrawConfigForm.value)
    } else if (formData.configKey === 'EMAIL_CONFIG') {
      if (emailConfigForm.value.enabled) {
        if (
          !emailConfigForm.value.smtp_host ||
          !emailConfigForm.value.smtp_port ||
          !emailConfigForm.value.from_email
        ) {
          throw new Error(t('validation.required'))
        }
      }
      formData.configValue = JSON.stringify(emailConfigForm.value)
    } else if (formData.configKey === 'PHONE_CONFIG') {
      if (phoneConfigForm.value.enabled && !phoneConfigForm.value.endpoint) {
        throw new Error(t('validation.required'))
      }
      if (phoneConfigForm.value.headers_json) {
        JSON.parse(phoneConfigForm.value.headers_json)
      }
      formData.configValue = JSON.stringify(phoneConfigForm.value)
    }

    if (isEdit.value) {
      const { id, ...updateData } = formData
      const res = await configService.update(id, updateData)
      if (res.code !== 200) throw new Error(res.msg || t('common.updateFailed'))
      ElMessage.success(t('common.updateSuccess'))
    } else {
      const data: SysConfigCreateReq = {
        tenantId: formData.tenantId,
        configKey: formData.configKey,
        configValue: formData.configValue,
        remark: formData.remark || undefined,
      }
      const res = await configService.create(data)
      if (res.code !== 200) throw new Error(res.msg || t('common.createFailed'))
      ElMessage.success(t('common.createSuccess'))
    }

    dialogVisible.value = false
    fetchList()
  } catch (error: unknown) {
    ElMessage.error(error instanceof Error ? error.message : t('common.operationFailed'))
  } finally {
    submitLoading.value = false
  }
}

// Initialize
onMounted(() => {
  loadKeys()
  fetchList()
})
</script>

<style scoped>
.sys-config {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.page-header h2 {
  margin: 0;
  color: #333;
}

.query-card {
  margin-bottom: 20px;
}

.table-card {
  margin-bottom: 20px;
}

.config-value {
  display: inline-block;
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
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
</style>

<style>
.config-tip {
  max-width: 520px !important;
  white-space: normal !important;
  word-break: break-all !important;
  overflow-wrap: anywhere !important;
  line-height: 1.5;
}

.sys-config-dialog {
  max-width: calc(100vw - 48px);
}

.sys-config-dialog .el-dialog__body {
  overflow-x: hidden;
}
</style>
