<template>
  <div class="payment-page">
    <div class="page-header">
      <h2>{{ t('payment.tenantChannelRules') }}</h2>
      <el-button @click="loadRules">
        {{ t('common.refresh') }}
      </el-button>
    </div>

    <el-card shadow="never" class="query-card">
      <el-form :model="ruleQuery" inline label-width="90px">
        <el-form-item :label="t('common.tenantId')">
          <el-input-number v-model="ruleQuery.tenantId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('payment.channelId')">
          <el-input-number v-model="ruleQuery.channelId" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadRules">
            {{ t('common.search') }}
          </el-button>
          <el-button type="primary" @click="openRuleDialog()">
            {{ t('payment.addRule') }}
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="never">
      <el-table v-loading="ruleLoading" :data="rules" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="tenantId" :label="t('common.tenantId')" width="100" />
        <el-table-column prop="channelId" :label="t('payment.channelId')" width="100" />
        <el-table-column prop="ruleName" :label="t('payment.ruleName')" min-width="140" />
        <el-table-column prop="priority" :label="t('payment.priority')" width="90" />
        <el-table-column :label="t('common.status')" width="100">
          <template #default="{ row }">
            <el-tag :class="getStatusTagClass(row.status)" disable-transitions>
              {{ getOptionValueLabel(optionGroups, 'status', row.status, t) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="t('common.actions')" width="160">
          <template #default="{ row }">
            <el-button link type="primary" @click="showRuleDetail(row)">
              {{ t('common.detail') }}
            </el-button>
            <el-button link type="primary" @click="openRuleDialog(row)">
              {{ t('common.edit') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog
      v-model="ruleDialogVisible"
      :title="ruleForm.id ? t('payment.editRule') : t('payment.addRule')"
      width="760px"
    >
      <el-form label-width="120px">
        <el-form-item :label="t('common.tenantId')">
          <div class="verify-row">
            <el-input-number
              v-model="ruleForm.tenantId"
              :min="1"
              :precision="0"
              :disabled="!!ruleForm.id"
              @change="handleRuleTenantChange"
            />
            <el-button v-if="!ruleForm.id" :loading="ruleTenantChecking" @click="checkRuleTenant">
              {{ t('payment.verifyTenant') }}
            </el-button>
            <span v-if="!ruleForm.id && ruleTenantVerified" class="verified-text">
              {{ t('payment.verified') }}
            </span>
          </div>
        </el-form-item>

        <el-form-item :label="t('payment.channelId')">
          <div class="verify-row">
            <el-input-number
              v-model="ruleForm.channelId"
              :min="1"
              :precision="0"
              :disabled="!!ruleForm.id"
              @change="handleRuleChannelChange"
            />
            <el-button v-if="!ruleForm.id" :loading="ruleChannelChecking" @click="checkRuleChannel">
              {{ t('payment.verifyChannel') }}
            </el-button>
            <span v-if="!ruleForm.id && ruleChannelVerified" class="verified-text">
              {{ t('payment.verified') }}
            </span>
          </div>
        </el-form-item>

        <el-form-item :label="t('payment.ruleName')">
          <el-input v-model="ruleForm.ruleName" />
        </el-form-item>
        <el-form-item :label="t('payment.priority')">
          <el-input-number v-model="ruleForm.priority" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('common.status')">
          <el-select v-model="ruleForm.status" style="width: 100%">
            <el-option
              v-for="item in statusOptions"
              :key="item.value"
              :label="getOptionLabel(t, item.code, item.value)"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('payment.singleAmountMin')">
          <el-input-number v-model="ruleForm.singleAmountMin" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('payment.singleAmountMax')">
          <el-input-number v-model="ruleForm.singleAmountMax" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('payment.allowNewUser')">
          <el-select v-model="ruleForm.allowNewUser" style="width: 100%">
            <el-option
              v-for="item in yesNoOptions"
              :key="item.value"
              :label="getOptionLabel(t, item.code, item.value)"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('payment.allowOldUser')">
          <el-select v-model="ruleForm.allowOldUser" style="width: 100%">
            <el-option
              v-for="item in yesNoOptions"
              :key="item.value"
              :label="getOptionLabel(t, item.code, item.value)"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('common.remark')">
          <el-input v-model="ruleForm.remark" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="ruleDialogVisible = false">
          {{ t('common.cancel') }}
        </el-button>
        <el-button type="primary" :disabled="ruleSubmitDisabled" @click="submitRule">
          {{ t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="detailVisible" :title="t('payment.detailTitle')" width="700px">
      <pre class="detail-pre">{{ JSON.stringify(detailData, null, 2) }}</pre>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import {
  tenantService,
  tenantsService,
  type OptionGroup,
  type TenantPayChannelRule,
} from '@/services'
import { findOptionGroup, getOptionLabel, getOptionValueLabel } from '@/utils/options'

const { t } = useI18n()

const ruleLoading = ref(false)
const rules = ref<TenantPayChannelRule[]>([])
const detailVisible = ref(false)
const detailData = ref<Record<string, unknown>>({})
const ruleDialogVisible = ref(false)

const optionGroups = ref<OptionGroup[]>([])
const statusOptions = computed(() => findOptionGroup(optionGroups.value, 'status'))
const yesNoOptions = computed(() => findOptionGroup(optionGroups.value, 'yesNo'))

const ruleQuery = reactive({ tenantId: 0, channelId: 0 })

const ruleForm = reactive({
  id: 0,
  tenantId: 0,
  channelId: 0,
  ruleName: '',
  priority: 0,
  status: 1,
  singleAmountMin: 0,
  singleAmountMax: 0,
  userTotalRechargeMin: 0,
  userTotalRechargeMax: 0,
  memberLevelMin: 0,
  memberLevelMax: 0,
  kycLevelMin: 0,
  kycLevelMax: 0,
  allowNewUser: 1,
  allowOldUser: 1,
  allowTags: '{}',
  denyTags: '{}',
  remark: '',
})

const ruleTenantChecking = ref(false)
const ruleChannelChecking = ref(false)
const ruleTenantVerified = ref(false)
const ruleChannelVerified = ref(false)
const verifiedRuleTenantId = ref(0)
const verifiedRuleChannelId = ref(0)

const ruleSubmitDisabled = computed(
  () =>
    !ruleForm.id &&
    (!ruleTenantVerified.value ||
      !ruleChannelVerified.value ||
      verifiedRuleTenantId.value !== ruleForm.tenantId ||
      verifiedRuleChannelId.value !== ruleForm.channelId),
)

const loadOptions = async () => {
  const res = await tenantService.getOptions()
  optionGroups.value = res.data || []
}

const loadRules = async () => {
  ruleLoading.value = true
  try {
    const res = await tenantService.getTenantChannelRuleList({
      ...ruleQuery,
      tenantId: ruleQuery.tenantId || undefined,
      channelId: ruleQuery.channelId || undefined,
      limit: 100,
    })
    rules.value = res.data || []
  } finally {
    ruleLoading.value = false
  }
}

const resetRuleVerifyState = () => {
  ruleTenantVerified.value = false
  ruleChannelVerified.value = false
  verifiedRuleTenantId.value = 0
  verifiedRuleChannelId.value = 0
}

const openRuleDialog = (row?: TenantPayChannelRule) => {
  Object.assign(
    ruleForm,
    row || {
      id: 0,
      tenantId: 0,
      channelId: 0,
      ruleName: '',
      priority: 0,
      status: 1,
      singleAmountMin: 0,
      singleAmountMax: 0,
      userTotalRechargeMin: 0,
      userTotalRechargeMax: 0,
      memberLevelMin: 0,
      memberLevelMax: 0,
      kycLevelMin: 0,
      kycLevelMax: 0,
      allowNewUser: 1,
      allowOldUser: 1,
      allowTags: '{}',
      denyTags: '{}',
      remark: '',
    },
  )

  if (row?.id) {
    ruleTenantVerified.value = true
    ruleChannelVerified.value = true
    verifiedRuleTenantId.value = row.tenantId
    verifiedRuleChannelId.value = row.channelId
  } else {
    resetRuleVerifyState()
  }

  ruleDialogVisible.value = true
}

const validateTenantExists = async (tenantId: number) => {
  if (!tenantId) {
    ElMessage.warning(t('payment.pleaseInputTenantId'))
    return false
  }

  ruleTenantChecking.value = true
  try {
    const res = await tenantsService.detail({ tenantId })
    if (!res.data?.id) {
      ElMessage.error(t('payment.tenantNotFound'))
      return false
    }
    return true
  } catch {
    ElMessage.error(t('payment.tenantNotFound'))
    return false
  } finally {
    ruleTenantChecking.value = false
  }
}

const validateChannelExists = async (channelId: number, tenantId: number) => {
  if (!channelId) {
    ElMessage.warning(t('payment.pleaseInputChannelId'))
    return false
  }
  if (!tenantId) {
    ElMessage.warning(t('payment.pleaseInputTenantFirst'))
    return false
  }

  ruleChannelChecking.value = true
  try {
    const res = await tenantService.getTenantChannelDetail(channelId, tenantId)
    if (!res.data?.id) {
      ElMessage.error(t('payment.channelNotFound'))
      return false
    }
    return true
  } catch {
    ElMessage.error(t('payment.channelNotFound'))
    return false
  } finally {
    ruleChannelChecking.value = false
  }
}

const handleRuleTenantChange = () => {
  ruleTenantVerified.value = false
  verifiedRuleTenantId.value = 0
  ruleChannelVerified.value = false
  verifiedRuleChannelId.value = 0
}

const handleRuleChannelChange = () => {
  ruleChannelVerified.value = false
  verifiedRuleChannelId.value = 0
}

const checkRuleTenant = async () => {
  const exists = await validateTenantExists(ruleForm.tenantId)
  ruleTenantVerified.value = exists
  verifiedRuleTenantId.value = exists ? ruleForm.tenantId : 0
  if (!exists) {
    ruleChannelVerified.value = false
    verifiedRuleChannelId.value = 0
  }
  if (exists) ElMessage.success(t('payment.tenantVerifiedSuccess'))
}

const checkRuleChannel = async () => {
  const exists = await validateChannelExists(ruleForm.channelId, ruleForm.tenantId)
  ruleChannelVerified.value = exists
  verifiedRuleChannelId.value = exists ? ruleForm.channelId : 0
  if (exists) ElMessage.success(t('payment.channelVerifiedSuccess'))
}

const submitRule = async () => {
  if (!ruleForm.id && ruleSubmitDisabled.value) {
    ElMessage.warning(t('payment.pleaseCompleteRuleValidation'))
    return
  }

  if (ruleForm.id) {
    await tenantService.updateTenantChannelRule({ ...ruleForm })
  } else {
    await tenantService.createTenantChannelRule({ ...ruleForm })
  }
  ElMessage.success(t('common.operationSuccess'))
  ruleDialogVisible.value = false
  loadRules()
}

const showRuleDetail = async (row: TenantPayChannelRule) => {
  const res = await tenantService.getTenantChannelRuleDetail(row.id, row.tenantId)
  detailData.value = res.data || row
  detailVisible.value = true
}

function getStatusTagClass(value?: number) {
  const num = Number(value ?? 0)
  if (num === 1) return 'option-tag option-tag--green'
  if (num === 2) return 'option-tag option-tag--red'
  return 'option-tag option-tag--slate'
}

onMounted(async () => {
  await Promise.all([loadOptions(), loadRules()])
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
