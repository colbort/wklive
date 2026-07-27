<template>
  <div class="payment-page">
    <CrudQueryCard :model="ruleQuery" @search="loadList" @reset="resetQuery">
      <el-form-item :label="t('common.tenantId')">
        <TenantSelect
          v-model="ruleQuery.tenantId"
          class="tenant-select-filter"
          @change="handleRuleQueryTenantChange"
        />
      </el-form-item>
      <el-form-item :label="t('payment.channelId')">
        <TenantPayChannelSelect
          v-model="ruleQuery.channelId"
          :tenant-id="ruleQuery.tenantId"
          :enabled-only="false"
        />
      </el-form-item>
      <template #actions>
        <el-button
          v-perm="'payment:tenant-channel-rule:add'"
          type="primary"
          @click="openRuleDialog()"
        >
          {{ t('common.add') }}
        </el-button>
      </template>
    </CrudQueryCard>

    <el-card shadow="never" class="table-card">
      <el-table v-loading="ruleLoading" :data="rules" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column :label="t('common.tenantId')" min-width="180">
          <template #default="{ row }">
            {{ formatRelationLabel(row.tenantName, row.tenantId) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('payment.channelId')" min-width="200">
          <template #default="{ row }">
            {{ formatRelationLabel(row.channelName, row.channelId) }}
          </template>
        </el-table-column>
        <el-table-column prop="ruleName" :label="t('payment.ruleName')" min-width="140" />
        <el-table-column prop="priority" :label="t('payment.priority')" width="90" />
        <el-table-column :label="t('common.enabled')" width="100">
          <template #default="{ row }">
            <el-tag :class="getEnabledTagClass(row.enabled)" disable-transitions>
              {{ getOptionValueLabel(optionGroups, 'enabled', row.enabled, t) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="t('common.actions')" align="center" width="160">
          <template #default="{ row }">
            <el-button
              v-perm="'payment:tenant-channel-rule:detail'"
              link
              type="primary"
              @click="showRuleDetail(row)"
            >
              {{ t('common.detail') }}
            </el-button>
            <el-button
              v-perm="'payment:tenant-channel-rule:update'"
              link
              type="primary"
              @click="openRuleDialog(row)"
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
      v-model="ruleDialogVisible"
      :title="ruleForm.id ? t('payment.editRule') : t('payment.addRule')"
      width="760px"
    >
      <el-form label-width="120px" class="rule-form-grid">
        <el-form-item :label="t('common.tenantId')">
          <TenantSelect
            v-model="ruleForm.tenantId"
            :disabled="!!ruleForm.id"
            @change="handleRuleTenantChange"
          />
        </el-form-item>

        <el-form-item :label="t('payment.channelId')">
          <TenantPayChannelSelect
            v-model="ruleForm.channelId"
            :tenant-id="ruleForm.tenantId"
            :disabled="!!ruleForm.id"
            :clearable="false"
            @change="handleRuleChannelChange"
          />
        </el-form-item>

        <el-form-item :label="t('payment.ruleName')">
          <el-input v-model="ruleForm.ruleName" />
        </el-form-item>
        <el-form-item :label="t('payment.priority')">
          <el-input-number v-model="ruleForm.priority" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item :label="t('common.enabled')">
          <el-select v-model="ruleForm.enabled" style="width: 100%">
            <el-option
              v-for="item in enabledFormOptions"
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
              v-for="item in yesNoFormOptions"
              :key="item.value"
              :label="getOptionLabel(t, item.code, item.value)"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('payment.allowOldUser')">
          <el-select v-model="ruleForm.allowOldUser" style="width: 100%">
            <el-option
              v-for="item in yesNoFormOptions"
              :key="item.value"
              :label="getOptionLabel(t, item.code, item.value)"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('common.remark')" class="rule-form-grid__full">
          <el-input v-model="ruleForm.remark" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="ruleDialogVisible = false">
          {{ t('common.cancel') }}
        </el-button>
        <el-button
          v-perm="
            ruleForm.id ? 'payment:tenant-channel-rule:update' : 'payment:tenant-channel-rule:add'
          "
          type="primary"
          :disabled="ruleSubmitDisabled"
          @click="submitRule"
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
import {
  tenantService,
  type OptionGroup,
  type TenantPayChannelRule,
} from '@/services'
import { findFormOptionGroup, getOptionLabel, getOptionValueLabel } from '@/utils/options'
import TenantSelect from '@/components/TenantSelect.vue'
import PaymentDetailDescriptions from '@/components/payment/PaymentDetailDescriptions.vue'
import CrudQueryCard from '@/components/common/CrudQueryCard.vue'
import TenantPayChannelSelect from '@/components/payment/TenantPayChannelSelect.vue'

const { t } = useI18n()
const { pagination, updateFromResponse, resetAndLoad, prevAndLoad, nextAndLoad } =
  usePagination<number>(20)

const ruleLoading = ref(false)
const rules = ref<TenantPayChannelRule[]>([])
const detailVisible = ref(false)
const detailData = ref<Record<string, unknown>>({})
const ruleDialogVisible = ref(false)

const optionGroups = ref<OptionGroup[]>([])
const enabledFormOptions = computed(() => findFormOptionGroup(optionGroups.value, 'enabled'))
const yesNoFormOptions = computed(() => findFormOptionGroup(optionGroups.value, 'yesNo'))

const ruleQuery = reactive({
  tenantId: undefined as number | undefined,
  channelId: undefined as number | undefined,
})

const ruleForm = reactive({
  id: 0,
  tenantId: 0,
  channelId: 0,
  ruleName: '',
  priority: 0,
  enabled: 1,
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
  allowTags: '[]',
  denyTags: '[]',
  remark: '',
})

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

const loadList = async () => {
  ruleLoading.value = true
  try {
    const res = await tenantService.getTenantChannelRuleList({
      ...ruleQuery,
      tenantId: ruleQuery.tenantId || undefined,
      channelId: ruleQuery.channelId || undefined,
      cursor: pagination.cursor,
      limit: pagination.limit,
    })
    rules.value = res.data || []
    updateFromResponse(res)
  } finally {
    ruleLoading.value = false
  }
}

function formatRelationLabel(name: string | undefined, id: number) {
  return name ? `${name} (${id})` : String(id)
}

function resetQuery() {
  ruleQuery.tenantId = 0
  ruleQuery.channelId = 0
  resetAndLoad(loadList)
}

function handleRuleQueryTenantChange() {
  ruleQuery.channelId = undefined
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
      enabled: 1,
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
      allowTags: '[]',
      denyTags: '[]',
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

const handleRuleTenantChange = () => {
  ruleTenantVerified.value = ruleForm.tenantId > 0
  verifiedRuleTenantId.value = ruleForm.tenantId
  ruleForm.channelId = 0
  ruleChannelVerified.value = false
  verifiedRuleChannelId.value = 0
}

const handleRuleChannelChange = () => {
  ruleChannelVerified.value = ruleForm.channelId > 0
  verifiedRuleChannelId.value = ruleForm.channelId
}

const submitRule = async () => {
  if (!ruleForm.id && ruleSubmitDisabled.value) {
    ElMessage.warning(t('payment.pleaseCompleteRuleValidation'))
    return
  }

  const payload = {
    ...ruleForm,
    singleAmountMin: String(ruleForm.singleAmountMin),
    singleAmountMax: String(ruleForm.singleAmountMax),
    userTotalRechargeMin: String(ruleForm.userTotalRechargeMin),
    userTotalRechargeMax: String(ruleForm.userTotalRechargeMax),
  }
  if (ruleForm.id) {
    await tenantService.updateTenantChannelRule(payload)
  } else {
    await tenantService.createTenantChannelRule(payload)
  }
  ElMessage.success(t('common.operationSuccess'))
  ruleDialogVisible.value = false
  loadList()
}

const showRuleDetail = async (row: TenantPayChannelRule) => {
  const res = await tenantService.getTenantChannelRuleDetail(row.id, row.tenantId)
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
</script>

<style scoped>
.rule-form-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  column-gap: 20px;
}

.rule-form-grid__full {
  grid-column: 1 / -1;
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
  .rule-form-grid {
    grid-template-columns: minmax(0, 1fr);
  }

  .rule-form-grid__full {
    grid-column: auto;
  }
}
</style>
